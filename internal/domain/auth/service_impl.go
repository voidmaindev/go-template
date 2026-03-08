package auth

import (
	"context"
	"log/slog"
	"time"

	"github.com/casbin/casbin/v3"
	"github.com/voidmaindev/go-template/internal/common/errors"
	"github.com/voidmaindev/go-template/internal/config"
	"github.com/voidmaindev/go-template/internal/domain/audit"
	"github.com/voidmaindev/go-template/internal/domain/email"
	"github.com/voidmaindev/go-template/internal/domain/rbac"
	"github.com/voidmaindev/go-template/internal/domain/user"
	"github.com/voidmaindev/go-template/internal/telemetry"
	"github.com/voidmaindev/go-template/pkg/utils"
	"gorm.io/gorm"
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
	db             *gorm.DB
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
}

// NewService creates a new auth service
func NewService(
	db *gorm.DB,
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
		db:             db,
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

	// Create user and assign self-registered role atomically
	err = s.enforcer.WithTransaction(ctx, func(tx *casbin.Transaction) error {
		// Wrap GORM create in a database transaction so it rolls back
		// together with the Casbin transaction on failure
		if err := s.db.WithContext(ctx).Transaction(func(gormTx *gorm.DB) error {
			return gormTx.Create(newUser).Error
		}); err != nil {
			return err
		}

		// Assign self-registered role
		if err := s.rbacSvc.AssignRoleInTx(tx, ctx, newUser.ID, s.selfRegConfig.DefaultRole); err != nil {
			slog.Error("failed to assign self-registered role",
				"userID", newUser.ID, "role", s.selfRegConfig.DefaultRole, "error", err)
			return err
		}

		return nil
	})
	if err != nil {
		return nil, errors.Internal(domainName, err).WithOperation("SelfRegister")
	}

	telemetry.IncrementUsersRegistered()

	// Send verification email if required
	if s.selfRegConfig.RequireEmailVerification {
		if err := s.sendVerificationEmail(ctx, newUser); err != nil {
			slog.Error("failed to send verification email", "email", req.Email, "error", err)
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

	// Delete token (single-use)
	defer s.tokenStore.DeleteVerificationToken(ctx, token)

	// Get user
	var u user.User
	if err := s.db.WithContext(ctx).First(&u, userID).Error; err != nil {
		return errors.Internal(domainName, err).WithOperation("VerifyEmail")
	}

	// Check if already verified
	if u.EmailVerifiedAt != nil {
		return ErrEmailAlreadyVerified
	}

	// Update verification status
	now := time.Now()
	if err := s.db.WithContext(ctx).Model(&u).Update("email_verified_at", now).Error; err != nil {
		return errors.Internal(domainName, err).WithOperation("VerifyEmail")
	}

	// Send welcome email (best effort)
	go func() {
		if err := s.emailSvc.SendWelcomeEmail(context.Background(), u.Email, u.Name); err != nil {
			slog.Error("failed to send welcome email", "email", u.Email, "error", err)
		}
	}()

	return nil
}

// ResendVerification resends the verification email
func (s *service) ResendVerification(ctx context.Context, emailAddr string) error {
	// Check rate limit
	allowed, err := s.tokenStore.CheckRateLimit(ctx, emailAddr, "verification", verificationRateLimit, rateLimitWindow)
	if err != nil {
		slog.Error("rate limit check failed", "error", err)
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
		slog.Error("failed to resend verification email", "email", emailAddr, "error", err)
	}

	return nil
}

// ForgotPassword initiates password reset flow
func (s *service) ForgotPassword(ctx context.Context, emailAddr string) error {
	// Check rate limit
	allowed, err := s.tokenStore.CheckRateLimit(ctx, emailAddr, "password_reset", passwordResetRateLimit, rateLimitWindow)
	if err != nil {
		slog.Error("rate limit check failed", "error", err)
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
		slog.Error("failed to send password reset email", "email", emailAddr, "error", err)
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

	// Delete token (single-use)
	defer s.tokenStore.DeletePasswordResetToken(ctx, req.Token)

	// Hash new password
	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return errors.Internal(domainName, err).WithOperation("ResetPassword")
	}

	// Update password
	if err := s.db.WithContext(ctx).Model(&user.User{}).Where("id = ?", userID).
		Update("password", hashedPassword).Error; err != nil {
		return errors.Internal(domainName, err).WithOperation("ResetPassword")
	}

	// Invalidate all existing sessions (security: revoke tokens on password reset)
	if err := s.userTokenStore.InvalidateAllUserTokens(ctx, userID); err != nil {
		slog.Error("failed to invalidate tokens after password reset", "userID", userID, "error", err)
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
		slog.Error("OAuth code exchange failed", "provider", provider, "error", err)
		return nil, ErrOAuthCodeExchange
	}

	// Get user info
	userInfo, err := p.GetUserInfo(ctx, tokens.AccessToken)
	if err != nil {
		slog.Error("OAuth get user info failed", "provider", provider, "error", err)
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
	var identity user.ExternalIdentity
	err := s.db.WithContext(ctx).
		Where("provider = ? AND provider_id = ?", provider, userInfo.ID).
		First(&identity).Error

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

	// Create user and identity atomically
	err = s.enforcer.WithTransaction(ctx, func(tx *casbin.Transaction) error {
		// Wrap GORM operations in a database transaction for atomicity
		if err := s.db.WithContext(ctx).Transaction(func(gormTx *gorm.DB) error {
			if err := gormTx.Create(newUser).Error; err != nil {
				return err
			}

			// Create external identity
			identity := &user.ExternalIdentity{
				UserID:     newUser.ID,
				Provider:   provider,
				ProviderID: userInfo.ID,
				Email:      userInfo.Email,
			}
			if err := gormTx.Create(identity).Error; err != nil {
				return err
			}

			return nil
		}); err != nil {
			return err
		}

		// Assign self-registered role within the Casbin transaction
		if err := s.rbacSvc.AssignRoleInTx(tx, ctx, newUser.ID, s.selfRegConfig.DefaultRole); err != nil {
			return err
		}

		return nil
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
	var identities []user.ExternalIdentity
	if err := s.db.WithContext(ctx).Where("user_id = ?", userID).Find(&identities).Error; err != nil {
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
	var existing user.ExternalIdentity
	err = s.db.WithContext(ctx).
		Where("provider = ? AND provider_id = ?", provider, userInfo.ID).
		First(&existing).Error
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
	var u user.User
	if err := s.db.WithContext(ctx).First(&u, userID).Error; err != nil {
		return errors.Internal(domainName, err).WithOperation("UnlinkIdentity")
	}

	// Count user's identities
	var count int64
	if err := s.db.WithContext(ctx).Model(&user.ExternalIdentity{}).
		Where("user_id = ?", userID).Count(&count).Error; err != nil {
		return errors.Internal(domainName, err).WithOperation("UnlinkIdentity")
	}

	// Prevent unlinking if it's the only login method
	hasPassword := u.Password != nil && *u.Password != ""
	if count <= 1 && !hasPassword {
		return ErrCannotUnlinkLastIdentity
	}

	// Delete the identity
	result := s.db.WithContext(ctx).
		Where("user_id = ? AND provider = ?", userID, provider).
		Delete(&user.ExternalIdentity{})
	if result.Error != nil {
		return errors.Internal(domainName, result.Error).WithOperation("UnlinkIdentity")
	}
	if result.RowsAffected == 0 {
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
	if err := s.db.WithContext(ctx).Model(&user.User{}).
		Where("id = ?", userID).
		Update("password", hashedPassword).Error; err != nil {
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
	return s.db.WithContext(ctx).Create(identity).Error
}
