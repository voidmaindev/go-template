package auth

import (
	"context"
	"time"

	"github.com/casbin/casbin/v3"
	"github.com/voidmaindev/go-template/internal/common/errors"
	"github.com/voidmaindev/go-template/internal/common/logging"
	"github.com/voidmaindev/go-template/internal/config"
	"github.com/voidmaindev/go-template/internal/domain/audit"
	"github.com/voidmaindev/go-template/internal/domain/email"
	"github.com/voidmaindev/go-template/internal/domain/rbac"
	"github.com/voidmaindev/go-template/internal/domain/user"
	"github.com/voidmaindev/go-template/internal/telemetry"
	"github.com/voidmaindev/go-template/pkg/utils"
)

// Rate limiting constants
const (
	verificationRateLimit    = 3  // per hour
	passwordResetRateLimit   = 5  // per hour
	rateLimitWindow          = time.Hour
	oauthStateExpiry         = 30 * time.Minute // Increased for real-world user flows
)

// service implements the Service interface
type service struct {
	userRepo       user.Repository
	tokenStore     *TokenStore
	userTokenStore *user.TokenStore
	emailSvc       email.Service
	rbacSvc        rbac.Service
	auditLogger    audit.Logger
	enforcer       *casbin.TransactionalEnforcer
	selfRegConfig  *config.SelfRegistrationConfig
	oauthConfig    *config.OAuthConfig
	oauthRegistry  *OAuthRegistry
	logger         *logging.Logger
}

// NewService creates a new auth service
func NewService(
	userRepo user.Repository,
	tokenStore *TokenStore,
	userTokenStore *user.TokenStore,
	emailSvc email.Service,
	rbacSvc rbac.Service,
	auditLogger audit.Logger,
	enforcer *casbin.TransactionalEnforcer,
	selfRegConfig *config.SelfRegistrationConfig,
	oauthConfig *config.OAuthConfig,
) Service {
	return &service{
		userRepo:       userRepo,
		tokenStore:     tokenStore,
		userTokenStore: userTokenStore,
		emailSvc:       emailSvc,
		rbacSvc:        rbacSvc,
		auditLogger:    auditLogger,
		enforcer:       enforcer,
		selfRegConfig:  selfRegConfig,
		oauthConfig:    oauthConfig,
		oauthRegistry:  NewOAuthRegistry(oauthConfig),
		logger:         logging.New(domainName),
	}
}

// SelfRegister handles email-based self-registration
func (s *service) SelfRegister(ctx context.Context, req *SelfRegisterRequest) (*SelfRegisterResponse, error) {
	// Check if self-registration is enabled
	if !s.selfRegConfig.Enabled {
		return nil, ErrSelfRegDisabled
	}

	// Check if email already exists
	exists, err := s.userRepo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.Internal(domainName, err).WithOperation("SelfRegister")
	}
	if exists {
		// Return generic success to prevent email enumeration
		return &SelfRegisterResponse{
			Message: "If this email is not already registered, you will receive a verification email",
		}, nil
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, errors.Internal(domainName, err).WithOperation("SelfRegister")
	}

	// Create user (not verified yet)
	newUser := &user.User{
		Email:            req.Email,
		Password:         &hashedPassword,
		Name:             req.Name,
		IsSelfRegistered: true,
	}

	// Create user and assign self-registered role atomically.
	// Begin GORM tx manually so it stays open until Casbin is also ready to commit.
	err = s.enforcer.WithTransaction(ctx, func(tx *casbin.Transaction) error {
		txRepo, gormTx, err := s.userRepo.BeginTx(ctx)
		if err != nil {
			return err
		}
		defer func() {
			if err != nil {
				gormTx.Rollback()
			}
		}()

		if err = txRepo.Create(ctx, newUser); err != nil {
			return err
		}

		// Assign self-registered role
		if err = s.rbacSvc.AssignRoleInTx(tx, ctx, newUser.ID, s.selfRegConfig.DefaultRole); err != nil {
			s.logger.Error(ctx, "failed to assign self-registered role", err,
				"userID", newUser.ID, "role", s.selfRegConfig.DefaultRole)
			return err
		}

		// Commit GORM only when Casbin is also ready to commit
		return gormTx.Commit().Error
	})
	if err != nil {
		return nil, errors.Internal(domainName, err).WithOperation("SelfRegister")
	}

	telemetry.IncrementUsersRegistered()

	// Send verification email if required
	if s.selfRegConfig.RequireEmailVerification {
		if err := s.sendVerificationEmail(ctx, newUser); err != nil {
			s.logger.Error(ctx, "failed to send verification email", err, "email", req.Email)
			// Don't fail registration, just log it
		}
	}

	return &SelfRegisterResponse{
		Message: "Registration successful. Please check your email to verify your account.",
		UserID:  newUser.ID,
	}, nil
}

// VerifyEmail verifies a user's email address
func (s *service) VerifyEmail(ctx context.Context, token string) error {
	// Get user ID from token
	userID, err := s.tokenStore.GetVerificationToken(ctx, token)
	if err != nil {
		return ErrInvalidToken
	}

	// Get user
	u, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return errors.Internal(domainName, err).WithOperation("VerifyEmail")
	}

	// Check if already verified
	if u.EmailVerifiedAt != nil {
		return ErrEmailAlreadyVerified
	}

	// Update verification status
	now := time.Now()
	if err := s.userRepo.UpdateFields(ctx, u.ID, map[string]any{"email_verified_at": now}); err != nil {
		return errors.Internal(domainName, err).WithOperation("VerifyEmail")
	}

	// Delete token only after successful DB update. If the delete fails the
	// token will still expire at TTL, but replay is possible until then — log
	// so the on-call can correlate with a Redis outage.
	if delErr := s.tokenStore.DeleteVerificationToken(ctx, token); delErr != nil {
		s.logger.Warn(ctx, "failed to delete verification token after use", "error", delErr.Error(), "userID", u.ID)
	}

	// Send welcome email (best effort)
	go func() {
		bgCtx := context.Background()
		defer func() {
			if r := recover(); r != nil {
				s.logger.Error(bgCtx, "panic sending welcome email", nil, "recovered", r, "email", u.Email)
			}
		}()
		if err := s.emailSvc.SendWelcomeEmail(bgCtx, u.Email, u.Name); err != nil {
			s.logger.Error(bgCtx, "failed to send welcome email", err, "email", u.Email)
		}
	}()

	return nil
}

// ResendVerification resends the verification email
func (s *service) ResendVerification(ctx context.Context, emailAddr string) error {
	// Check rate limit
	allowed, err := s.tokenStore.CheckRateLimit(ctx, emailAddr, "verification", verificationRateLimit, rateLimitWindow)
	if err != nil {
		s.logger.Error(ctx, "rate limit check failed", err)
	} else if !allowed {
		return ErrRateLimited
	}

	// Find user by email
	u, err := s.userRepo.FindByEmail(ctx, emailAddr)
	if err != nil {
		// Don't reveal if user exists
		return nil
	}

	// Check if already verified
	if u.EmailVerifiedAt != nil {
		// Don't reveal verification status
		return nil
	}

	// Send verification email
	if err := s.sendVerificationEmail(ctx, u); err != nil {
		s.logger.Error(ctx, "failed to resend verification email", err, "email", emailAddr)
	}

	return nil
}

// ForgotPassword initiates password reset flow
func (s *service) ForgotPassword(ctx context.Context, emailAddr string) error {
	// Check rate limit
	allowed, err := s.tokenStore.CheckRateLimit(ctx, emailAddr, "password_reset", passwordResetRateLimit, rateLimitWindow)
	if err != nil {
		s.logger.Error(ctx, "rate limit check failed", err)
	} else if !allowed {
		return ErrRateLimited
	}

	// Find user by email
	u, err := s.userRepo.FindByEmail(ctx, emailAddr)
	if err != nil {
		// Don't reveal if user exists - return success anyway
		return nil
	}

	// Generate reset token
	token, err := GenerateToken()
	if err != nil {
		return errors.Internal(domainName, err).WithOperation("ForgotPassword")
	}

	// Store token
	if err := s.tokenStore.StorePasswordResetToken(ctx, token, u.ID, s.selfRegConfig.PasswordResetTokenExpiry); err != nil {
		return errors.Internal(domainName, err).WithOperation("ForgotPassword")
	}

	// Send password reset email
	if err := s.emailSvc.SendPasswordResetEmail(ctx, u.Email, u.Name, token); err != nil {
		s.logger.Error(ctx, "failed to send password reset email", err, "email", emailAddr)
	}

	return nil
}

// ResetPassword resets a user's password
func (s *service) ResetPassword(ctx context.Context, req *ResetPasswordRequest) error {
	// Get user ID from token
	userID, err := s.tokenStore.GetPasswordResetToken(ctx, req.Token)
	if err != nil {
		return ErrInvalidToken
	}

	// Hash new password
	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return errors.Internal(domainName, err).WithOperation("ResetPassword")
	}

	// Update password
	if err := s.userRepo.UpdateFields(ctx, userID, map[string]any{"password": hashedPassword}); err != nil {
		return errors.Internal(domainName, err).WithOperation("ResetPassword")
	}

	// Delete token only after successful DB update. See note on verification
	// token above — TTL bounds exposure, but surface delete failures.
	if delErr := s.tokenStore.DeletePasswordResetToken(ctx, req.Token); delErr != nil {
		s.logger.Warn(ctx, "failed to delete password reset token after use", "error", delErr.Error(), "userID", userID)
	}

	// Invalidate all existing sessions (security: revoke tokens on password reset)
	if err := s.userTokenStore.InvalidateAllUserTokens(ctx, userID); err != nil {
		s.logger.Error(ctx, "failed to invalidate tokens after password reset", err, "userID", userID)
		// Don't fail the password reset, but log the error
	}

	return nil
}

// GetOAuthURL returns the OAuth authorization URL for a provider
func (s *service) GetOAuthURL(ctx context.Context, provider, redirectURL string) (string, string, error) {
	// Check if self-registration is enabled
	if !s.selfRegConfig.Enabled {
		return "", "", ErrSelfRegDisabled
	}

	p := s.oauthRegistry.Get(provider)
	if p == nil {
		return "", "", ErrOAuthDisabled
	}

	// Generate state
	state, err := GenerateToken()
	if err != nil {
		return "", "", errors.Internal(domainName, err).WithOperation("GetOAuthURL")
	}

	// Generate PKCE challenge if provider supports it
	var pkce *PKCEChallenge
	var verifier string
	if p.SupportsPKCE() {
		pkce, err = GeneratePKCE()
		if err != nil {
			return "", "", errors.Internal(domainName, err).WithOperation("GetOAuthURL")
		}
		verifier = pkce.Verifier
	}

	// Store state for verification (including PKCE verifier)
	stateData := &OAuthStateData{
		Provider:     provider,
		RedirectURL:  redirectURL,
		PKCEVerifier: verifier,
	}
	if err := s.tokenStore.StoreOAuthState(ctx, state, stateData, oauthStateExpiry); err != nil {
		return "", "", errors.Internal(domainName, err).WithOperation("GetOAuthURL")
	}

	return p.GetAuthURL(state, pkce), state, nil
}

// HandleOAuthCallback handles the OAuth callback
func (s *service) HandleOAuthCallback(ctx context.Context, provider, code, state string) (*OAuthResult, error) {
	// Check if self-registration is enabled
	if !s.selfRegConfig.Enabled {
		return nil, ErrSelfRegDisabled
	}

	// Verify state
	stateData, err := s.tokenStore.GetOAuthState(ctx, state)
	if err != nil {
		return nil, ErrOAuthStateMismatch
	}
	defer s.tokenStore.DeleteOAuthState(ctx, state)

	if stateData.Provider != provider {
		return nil, ErrOAuthStateMismatch
	}

	// Get provider
	p := s.oauthRegistry.Get(provider)
	if p == nil {
		return nil, ErrOAuthDisabled
	}

	// Exchange code for tokens (pass PKCE verifier from state)
	tokens, err := p.ExchangeCode(ctx, code, stateData.PKCEVerifier)
	if err != nil {
		s.logger.Error(ctx, "OAuth code exchange failed", err, "provider", provider)
		return nil, ErrOAuthCodeExchange
	}

	// Get user info
	userInfo, err := p.GetUserInfo(ctx, tokens.AccessToken)
	if err != nil {
		s.logger.Error(ctx, "OAuth get user info failed", err, "provider", provider)
		return nil, ErrOAuthUserInfo
	}

	// Find or create user
	result, err := s.findOrCreateOAuthUser(ctx, provider, userInfo)
	if err != nil {
		return nil, err
	}

	// Include redirect URL from state
	result.RedirectURL = stateData.RedirectURL
	return result, nil
}

// findOrCreateOAuthUser finds an existing user or creates a new one from OAuth info
func (s *service) findOrCreateOAuthUser(ctx context.Context, provider string, userInfo *OAuthUserInfo) (*OAuthResult, error) {
	// Check if identity already linked
	identity, err := s.userRepo.FindExternalIdentityByProvider(ctx, provider, userInfo.ID)
	if err == nil {
		// Identity exists, return the user
		return &OAuthResult{
			IsNewUser: false,
			UserID:    identity.UserID,
			Email:     userInfo.Email,
		}, nil
	}

	// Check if user with this email exists (auto-link if verified)
	if userInfo.Email != "" && userInfo.EmailVerified {
		existingUser, err := s.userRepo.FindByEmail(ctx, userInfo.Email)
		if err == nil {
			// User exists, link identity
			if err := s.createExternalIdentity(ctx, existingUser.ID, provider, userInfo); err != nil {
				return nil, err
			}

			// Log auto-linked identity
			s.auditLogger.LogAsync(ctx, &audit.AuditEntry{
				UserID:     &existingUser.ID,
				Action:     audit.ActionOAuthLinked,
				Resource:   "identity",
				ResourceID: &existingUser.ID,
				Success:    true,
				Details: map[string]any{
					"provider":  provider,
					"auto_link": true,
				},
			})

			return &OAuthResult{
				IsNewUser: false,
				UserID:    existingUser.ID,
				Email:     userInfo.Email,
			}, nil
		}
	}

	// Create new user
	var emailVerifiedAt *time.Time
	if userInfo.EmailVerified {
		now := time.Now()
		emailVerifiedAt = &now
	}

	newUser := &user.User{
		Email:            userInfo.Email,
		Name:             userInfo.Name,
		EmailVerifiedAt:  emailVerifiedAt,
		IsSelfRegistered: true,
	}

	// Create user and identity atomically.
	// Begin GORM tx manually so it stays open until Casbin is also ready to commit.
	err = s.enforcer.WithTransaction(ctx, func(tx *casbin.Transaction) error {
		txRepo, gormTx, err := s.userRepo.BeginTx(ctx)
		if err != nil {
			return err
		}
		defer func() {
			if err != nil {
				gormTx.Rollback()
			}
		}()

		if err = txRepo.Create(ctx, newUser); err != nil {
			return err
		}

		// Create external identity within the same GORM transaction
		extIdentity := &user.ExternalIdentity{
			UserID:     newUser.ID,
			Provider:   provider,
			ProviderID: userInfo.ID,
			Email:      userInfo.Email,
		}
		if err = txRepo.CreateExternalIdentity(ctx, extIdentity); err != nil {
			return err
		}

		// Assign self-registered role within the Casbin transaction
		if err = s.rbacSvc.AssignRoleInTx(tx, ctx, newUser.ID, s.selfRegConfig.DefaultRole); err != nil {
			return err
		}

		// Commit GORM only when Casbin is also ready to commit
		return gormTx.Commit().Error
	})
	if err != nil {
		return nil, errors.Internal(domainName, err).WithOperation("findOrCreateOAuthUser")
	}

	telemetry.IncrementUsersRegistered()

	// Log OAuth registration
	s.auditLogger.LogAsync(ctx, &audit.AuditEntry{
		UserID:   &newUser.ID,
		Action:   audit.ActionSelfRegistered,
		Resource: "user",
		Success:  true,
		Details: map[string]any{
			"provider": provider,
			"email":    userInfo.Email,
			"method":   "oauth",
		},
	})

	return &OAuthResult{
		IsNewUser: true,
		UserID:    newUser.ID,
		Email:     userInfo.Email,
	}, nil
}

// GetUserIdentities returns all OAuth identities for a user
func (s *service) GetUserIdentities(ctx context.Context, userID uint) ([]*user.ExternalIdentityResponse, error) {
	identities, err := s.userRepo.FindExternalIdentitiesByUserID(ctx, userID)
	if err != nil {
		return nil, errors.Internal(domainName, err).WithOperation("GetUserIdentities")
	}

	result := make([]*user.ExternalIdentityResponse, len(identities))
	for i, identity := range identities {
		result[i] = identity.ToResponse()
	}
	return result, nil
}

// LinkIdentity links an OAuth identity to an existing user
func (s *service) LinkIdentity(ctx context.Context, userID uint, provider, code, state string) error {
	// Verify state
	stateData, err := s.tokenStore.GetOAuthState(ctx, state)
	if err != nil {
		return ErrOAuthStateMismatch
	}
	defer s.tokenStore.DeleteOAuthState(ctx, state)

	if stateData.Provider != provider || stateData.UserID != userID {
		return ErrOAuthStateMismatch
	}

	// Get provider
	p := s.oauthRegistry.Get(provider)
	if p == nil {
		return ErrOAuthDisabled
	}

	// Exchange code for tokens (pass PKCE verifier from state)
	tokens, err := p.ExchangeCode(ctx, code, stateData.PKCEVerifier)
	if err != nil {
		return ErrOAuthCodeExchange
	}

	// Get user info
	userInfo, err := p.GetUserInfo(ctx, tokens.AccessToken)
	if err != nil {
		return ErrOAuthUserInfo
	}

	// Check if identity already linked to another user
	_, err = s.userRepo.FindExternalIdentityByProvider(ctx, provider, userInfo.ID)
	if err == nil {
		return ErrIdentityAlreadyLinked
	}

	if err := s.createExternalIdentity(ctx, userID, provider, userInfo); err != nil {
		return err
	}

	// Log identity linking
	s.auditLogger.LogAsync(ctx, &audit.AuditEntry{
		UserID:     &userID,
		Action:     audit.ActionOAuthLinked,
		Resource:   "identity",
		ResourceID: &userID,
		Success:    true,
		Details: map[string]any{
			"provider": provider,
		},
	})

	return nil
}

// UnlinkIdentity unlinks an OAuth identity from a user
func (s *service) UnlinkIdentity(ctx context.Context, userID uint, provider string) error {
	// Get user to check if they have a password
	u, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return errors.Internal(domainName, err).WithOperation("UnlinkIdentity")
	}

	// Count user's identities
	count, err := s.userRepo.CountExternalIdentitiesByUserID(ctx, userID)
	if err != nil {
		return errors.Internal(domainName, err).WithOperation("UnlinkIdentity")
	}

	// Prevent unlinking if it's the only login method
	hasPassword := u.Password != nil && *u.Password != ""
	if count <= 1 && !hasPassword {
		return ErrCannotUnlinkLastIdentity
	}

	// Delete the identity
	rowsAffected, err := s.userRepo.DeleteExternalIdentityByProvider(ctx, userID, provider)
	if err != nil {
		return errors.Internal(domainName, err).WithOperation("UnlinkIdentity")
	}
	if rowsAffected == 0 {
		return ErrIdentityNotFound
	}

	// Log identity unlinking
	s.auditLogger.LogAsync(ctx, &audit.AuditEntry{
		UserID:     &userID,
		Action:     audit.ActionOAuthUnlinked,
		Resource:   "identity",
		ResourceID: &userID,
		Success:    true,
		Details: map[string]any{
			"provider": provider,
		},
	})

	return nil
}

// SetPassword sets a password for OAuth-only users
func (s *service) SetPassword(ctx context.Context, userID uint, password string) error {
	// Hash password
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return errors.Internal(domainName, err).WithOperation("SetPassword")
	}

	// Update password
	if err := s.userRepo.UpdateFields(ctx, userID, map[string]any{"password": hashedPassword}); err != nil {
		return errors.Internal(domainName, err).WithOperation("SetPassword")
	}

	return nil
}

// sendVerificationEmail generates a token and sends verification email
func (s *service) sendVerificationEmail(ctx context.Context, u *user.User) error {
	token, err := GenerateToken()
	if err != nil {
		return err
	}

	if err := s.tokenStore.StoreVerificationToken(ctx, token, u.ID, s.selfRegConfig.VerificationTokenExpiry); err != nil {
		return err
	}

	return s.emailSvc.SendVerificationEmail(ctx, u.Email, u.Name, token)
}

// createExternalIdentity creates an external identity record
func (s *service) createExternalIdentity(ctx context.Context, userID uint, provider string, userInfo *OAuthUserInfo) error {
	identity := &user.ExternalIdentity{
		UserID:     userID,
		Provider:   provider,
		ProviderID: userInfo.ID,
		Email:      userInfo.Email,
	}
	return s.userRepo.CreateExternalIdentity(ctx, identity)
}
