package user

import (
	"context"

	"github.com/casbin/casbin/v3"
	"github.com/voidmaindev/go-template/internal/common"
	"github.com/voidmaindev/go-template/internal/common/errors"
	"github.com/voidmaindev/go-template/internal/common/filter"
	"github.com/voidmaindev/go-template/internal/common/logging"
	"github.com/voidmaindev/go-template/internal/config"
	"github.com/voidmaindev/go-template/internal/domain/rbac"
	"github.com/voidmaindev/go-template/internal/telemetry"
	"github.com/voidmaindev/go-template/pkg/utils"
)

// blacklistMaxRetries is the number of attempts for atomic token blacklist operations.
// 3 retries provides resilience against transient Redis failures while preventing
// indefinite blocking on persistent errors.
const blacklistMaxRetries = 3

// LoginContext contains additional context for login (IP address, user agent)
type LoginContext struct {
	IP        string
	UserAgent string
}

// Service defines the user service interface
type Service interface {
	// Authentication
	Register(ctx context.Context, req *RegisterRequest) (*TokenResponse, error)
	Login(ctx context.Context, req *LoginRequest, loginCtx *LoginContext) (*TokenResponse, error)
	Logout(ctx context.Context, accessToken, refreshToken string) error
	RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error)

	// User management
	GetByID(ctx context.Context, id uint) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	Update(ctx context.Context, id uint, req *UpdateUserRequest) (*User, error)
	ChangePassword(ctx context.Context, id uint, req *ChangePasswordRequest) error
	Delete(ctx context.Context, id uint) error

	// Listing
	List(ctx context.Context, pagination *common.Pagination) (*common.PaginatedResult[User], error)
	ListFiltered(ctx context.Context, params *filter.Params) (*common.PaginatedResult[User], error)
}

// ServiceConfig holds all dependencies for the user service.
type ServiceConfig struct {
	Repo           Repository
	TokenStore     TokenStorer
	JWTConfig      *config.JWTConfig
	SelfRegConfig  *config.SelfRegistrationConfig
	SecurityConfig *config.SecurityConfig
	IsProduction   bool
	RBACService    rbac.Service
	Enforcer       *casbin.TransactionalEnforcer
}

// service implements the Service interface
type service struct {
	repo           Repository
	tokenStore     TokenStorer
	jwtConfig      *config.JWTConfig
	selfRegConfig  *config.SelfRegistrationConfig
	securityConfig *config.SecurityConfig
	isProduction   bool
	rbacSvc        rbac.Service
	enforcer       *casbin.TransactionalEnforcer
	logger         *logging.Logger
}

// NewService creates a new user service. Panics on a misconfigured dependency
// set, since every field here is load-bearing for user/auth/RBAC flows and a
// nil value would otherwise surface as an opaque NPE on the first request.
func NewService(cfg ServiceConfig) Service {
	switch {
	case cfg.Repo == nil:
		panic("user.NewService: Repo is required")
	case cfg.TokenStore == nil:
		panic("user.NewService: TokenStore is required")
	case cfg.JWTConfig == nil:
		panic("user.NewService: JWTConfig is required")
	case cfg.SelfRegConfig == nil:
		panic("user.NewService: SelfRegConfig is required")
	case cfg.SecurityConfig == nil:
		panic("user.NewService: SecurityConfig is required")
	case cfg.RBACService == nil:
		panic("user.NewService: RBACService is required (register rbac domain before user)")
	case cfg.Enforcer == nil:
		panic("user.NewService: Enforcer is required (register rbac domain before user)")
	}
	return &service{
		repo:           cfg.Repo,
		tokenStore:     cfg.TokenStore,
		jwtConfig:      cfg.JWTConfig,
		selfRegConfig:  cfg.SelfRegConfig,
		securityConfig: cfg.SecurityConfig,
		isProduction:   cfg.IsProduction,
		rbacSvc:        cfg.RBACService,
		enforcer:       cfg.Enforcer,
		logger:         logging.New(domainName),
	}
}

// Register creates a new user and returns tokens
func (s *service) Register(ctx context.Context, req *RegisterRequest) (*TokenResponse, error) {
	// Check if email already exists
	exists, err := s.repo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.Internal(domainName, err).WithOperation("Register")
	}
	if exists {
		return nil, ErrEmailExists
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, errors.Internal(domainName, err).WithOperation("Register")
	}

	// Prepare role codes (default to user if none specified)
	roleCodes := req.RoleCodes
	if len(roleCodes) == 0 {
		roleCodes = []string{rbac.RoleCodeUser}
	}

	// Create user and assign roles atomically
	user := &User{
		Email:    req.Email,
		Password: &hashedPassword,
		Name:     req.Name,
	}

	// Use Casbin's WithTransaction for atomic user creation and role assignment.
	// Begin GORM tx manually so it stays open until Casbin is also ready to commit.
	err = s.enforcer.WithTransaction(ctx, func(tx *casbin.Transaction) error {
		txRepo, gormTx, err := s.repo.BeginTx(ctx)
		if err != nil {
			return err
		}
		defer func() {
			if err != nil {
				gormTx.Rollback()
			}
		}()

		if err = txRepo.Create(ctx, user); err != nil {
			return err
		}

		// Assign RBAC roles using transaction - if any fails, everything rolls back
		for _, roleCode := range roleCodes {
			if err = s.rbacSvc.AssignRoleInTx(tx, ctx, user.ID, roleCode); err != nil {
				s.logger.Error(ctx, "failed to assign role during registration, rolling back", err,
					"role", roleCode, "userID", user.ID)
				return err
			}
		}

		// Commit GORM only when Casbin is also ready to commit
		return gormTx.Commit().Error
	})
	if err != nil {
		return nil, errors.Internal(domainName, err).WithOperation("Register")
	}

	// Record metric
	telemetry.IncrementUsersRegistered()

	// Generate tokens
	return s.generateTokenResponse(user)
}

// Login authenticates a user and returns tokens
func (s *service) Login(ctx context.Context, req *LoginRequest, loginCtx *LoginContext) (*TokenResponse, error) {
	// Check login rate limits if security config is set
	if s.securityConfig != nil && loginCtx != nil {
		if err := s.tokenStore.CheckLoginRateLimit(
			ctx,
			req.Email,
			loginCtx.IP,
			s.securityConfig.LoginRateLimitPerEmail,
			s.securityConfig.LoginRateLimitPerIP,
			s.securityConfig.LoginLockoutDuration,
		); err != nil {
			telemetry.IncrementAuthFailures("rate_limited")
			return nil, ErrTooManyLoginAttempts
		}
	}

	// Find user by email
	user, err := s.repo.FindByEmail(ctx, req.Email)
	if err != nil {
		if errors.IsNotFound(err) {
			telemetry.IncrementAuthFailures("user_not_found")
			s.recordFailedLogin(ctx, req.Email, loginCtx)
			return nil, ErrInvalidCredentials
		}
		return nil, errors.Internal(domainName, err).WithOperation("Login")
	}

	// Check if user has a password (OAuth-only users cannot login with password)
	if user.Password == nil || *user.Password == "" {
		telemetry.IncrementAuthFailures("no_password")
		s.recordFailedLogin(ctx, req.Email, loginCtx)
		return nil, ErrInvalidCredentials
	}

	// Verify password
	if !utils.CheckPassword(req.Password, *user.Password) {
		telemetry.IncrementAuthFailures("invalid_password")
		s.recordFailedLogin(ctx, req.Email, loginCtx)
		return nil, ErrInvalidCredentials
	}

	// Check email verification for self-registered users
	if user.IsSelfRegistered && user.EmailVerifiedAt == nil && s.selfRegConfig != nil && s.selfRegConfig.RequireEmailVerification {
		telemetry.IncrementAuthFailures("email_not_verified")
		return nil, ErrEmailNotVerified
	}

	// Clear rate limit counter on successful login
	if s.securityConfig != nil {
		if err := s.tokenStore.ClearLoginRateLimit(ctx, req.Email); err != nil {
			s.logger.Warn(ctx, "failed to clear login rate limit", "email", req.Email, "error", err)
		}
	}

	// Record metric
	telemetry.IncrementUsersLoggedIn()

	// Generate tokens
	return s.generateTokenResponse(user)
}

// recordFailedLogin records a failed login attempt for rate limiting
func (s *service) recordFailedLogin(ctx context.Context, email string, loginCtx *LoginContext) {
	if s.securityConfig == nil || loginCtx == nil {
		return
	}
	if err := s.tokenStore.RecordFailedLogin(ctx, email, loginCtx.IP, s.securityConfig.LoginLockoutDuration); err != nil {
		s.logger.Warn(ctx, "failed to record failed login attempt", "email", email, "error", err)
	}
}

// Logout invalidates both access and refresh tokens
func (s *service) Logout(ctx context.Context, accessToken, refreshToken string) error {
	// Blacklist access token
	accessExpiry, err := utils.GetTokenExpiry(accessToken, s.jwtConfig.SecretKey)
	if err != nil {
		// Token might already be expired, but we still try to blacklist it
		accessExpiry = s.jwtConfig.AccessTokenExpiry
	}
	if err := s.tokenStore.Blacklist(ctx, accessToken, accessExpiry); err != nil {
		return errors.Internal(domainName, err).WithOperation("Logout.BlacklistAccessToken")
	}

	// Blacklist refresh token if provided (prevents token reuse attack)
	if refreshToken != "" {
		refreshExpiry, err := utils.GetTokenExpiry(refreshToken, s.jwtConfig.SecretKey)
		if err != nil {
			refreshExpiry = s.jwtConfig.RefreshTokenExpiry
		}
		if err := s.tokenStore.Blacklist(ctx, refreshToken, refreshExpiry); err != nil {
			s.logger.Error(ctx, "failed to blacklist refresh token during logout", err)
			// Continue anyway - access token is already blacklisted
		}
	}

	return nil
}

// RefreshToken generates new tokens from a refresh token
func (s *service) RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error) {
	// Validate refresh token
	claims, err := utils.ValidateRefreshToken(refreshToken, s.jwtConfig.SecretKey)
	if err != nil {
		return nil, ErrTokenInvalid
	}

	// Atomically blacklist the refresh token to prevent race conditions.
	// This check-and-set operation ensures the token can only be used once,
	// even with concurrent requests.
	expiry, _ := utils.GetTokenExpiry(refreshToken, s.jwtConfig.SecretKey)
	if expiry > 0 {
		wasBlacklisted, err := s.tokenStore.BlacklistAtomicWithRetry(ctx, refreshToken, expiry, blacklistMaxRetries)
		if err != nil {
			s.logger.Error(ctx, "failed to blacklist refresh token after retries", err,
				"retries", blacklistMaxRetries,
				"userID", claims.UserID,
			)
			// In production, fail the request to prevent stolen token reuse
			if s.isProduction {
				return nil, ErrTokenRefreshUnavailable
			}
		} else if !wasBlacklisted {
			// Token was already blacklisted by another concurrent request
			return nil, ErrTokenBlacklisted
		}
	} else {
		// Token has no expiry or is already expired, check if blacklisted
		isBlacklisted, err := s.tokenStore.IsBlacklisted(ctx, refreshToken)
		if err != nil {
			return nil, errors.Internal(domainName, err).WithOperation("RefreshToken")
		}
		if isBlacklisted {
			return nil, ErrTokenBlacklisted
		}
	}

	// Get user
	user, err := s.repo.FindByID(ctx, claims.UserID)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, ErrUserNotFound
		}
		return nil, errors.Internal(domainName, err).WithOperation("RefreshToken")
	}

	// Generate new tokens
	return s.generateTokenResponse(user)
}

// GetByID retrieves a user by ID
func (s *service) GetByID(ctx context.Context, id uint) (*User, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, ErrUserNotFound
		}
		return nil, errors.Internal(domainName, err).WithOperation("GetByID")
	}
	return user, nil
}

// GetByEmail retrieves a user by email
func (s *service) GetByEmail(ctx context.Context, email string) (*User, error) {
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, ErrUserNotFound
		}
		return nil, errors.Internal(domainName, err).WithOperation("GetByEmail")
	}
	return user, nil
}

// Update updates a user
func (s *service) Update(ctx context.Context, id uint, req *UpdateUserRequest) (*User, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, ErrUserNotFound
		}
		return nil, errors.Internal(domainName, err).WithOperation("Update")
	}

	// Map non-nil/non-empty fields from request to model
	if err := utils.UpdateModel(user, req); err != nil {
		return nil, errors.Internal(domainName, err).WithOperation("Update")
	}

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, errors.Internal(domainName, err).WithOperation("Update")
	}

	// Record metric
	telemetry.IncrementUsersUpdated()

	return user, nil
}

// ChangePassword changes a user's password
func (s *service) ChangePassword(ctx context.Context, id uint, req *ChangePasswordRequest) error {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.IsNotFound(err) {
			return ErrUserNotFound
		}
		return errors.Internal(domainName, err).WithOperation("ChangePassword")
	}

	// Check if user has a password (OAuth-only users must use SetPassword instead)
	if user.Password == nil || *user.Password == "" {
		return ErrNoPassword
	}

	// Verify current password
	if !utils.CheckPassword(req.CurrentPassword, *user.Password) {
		return ErrInvalidPassword
	}

	// Check if new password is different
	if utils.CheckPassword(req.NewPassword, *user.Password) {
		return ErrSamePassword
	}

	// Hash new password
	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return errors.Internal(domainName, err).WithOperation("ChangePassword")
	}

	// Update password
	if err := s.repo.UpdateFields(ctx, id, map[string]any{
		"password": hashedPassword,
	}); err != nil {
		return errors.Internal(domainName, err).WithOperation("ChangePassword")
	}

	// Invalidate all existing tokens for this user (security: revoke sessions on password change)
	if err := s.tokenStore.InvalidateAllUserTokens(ctx, id); err != nil {
		s.logger.Error(ctx, "failed to invalidate tokens after password change", err, "userID", id)
		// Don't fail the password change, but log the error
	}

	return nil
}

// Delete soft-deletes a user and cascades to external identities within a transaction
func (s *service) Delete(ctx context.Context, id uint) error {
	_, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.IsNotFound(err) {
			return ErrUserNotFound
		}
		return errors.Internal(domainName, err).WithOperation("Delete")
	}

	// Delete external identities and user atomically
	txRepo, gormTx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return errors.Internal(domainName, err).WithOperation("Delete")
	}
	defer func() {
		if err != nil {
			gormTx.Rollback()
		}
	}()

	if err = txRepo.DeleteExternalIdentitiesByUserID(ctx, id); err != nil {
		return errors.Internal(domainName, err).WithOperation("Delete")
	}

	if err = txRepo.Delete(ctx, id); err != nil {
		return errors.Internal(domainName, err).WithOperation("Delete")
	}

	if err = gormTx.Commit().Error; err != nil {
		return errors.Internal(domainName, err).WithOperation("Delete")
	}

	// Invalidate all existing tokens for this user (security: revoke sessions on deletion)
	if err := s.tokenStore.InvalidateAllUserTokens(ctx, id); err != nil {
		s.logger.Error(ctx, "failed to invalidate tokens after user deletion", err, "userID", id)
	}

	// Remove RBAC role assignments (cleanup orphaned Casbin policies)
	if s.enforcer != nil {
		subject := rbac.FormatUserSubject(id)
		if _, err := s.enforcer.RemoveFilteredGroupingPolicy(0, subject); err != nil {
			s.logger.Error(ctx, "failed to remove RBAC roles after user deletion", err, "userID", id)
		} else if err := s.enforcer.SavePolicy(); err != nil {
			s.logger.Error(ctx, "failed to save RBAC policy after user deletion", err, "userID", id)
		}
	}

	// Record metric
	telemetry.IncrementUsersDeleted()

	return nil
}

// List retrieves all users with pagination
func (s *service) List(ctx context.Context, pagination *common.Pagination) (*common.PaginatedResult[User], error) {
	users, total, err := s.repo.FindAll(ctx, pagination)
	if err != nil {
		return nil, errors.Internal(domainName, err).WithOperation("List")
	}

	return common.NewPaginatedResult(users, total, pagination), nil
}

// ListFiltered retrieves users with dynamic filtering, sorting, and pagination
func (s *service) ListFiltered(ctx context.Context, params *filter.Params) (*common.PaginatedResult[User], error) {
	users, total, err := s.repo.FindAllFiltered(ctx, params)
	if err != nil {
		return nil, errors.Internal(domainName, err).WithOperation("ListFiltered")
	}

	return common.NewPaginatedResultFromFilter(users, total, params), nil
}

// toUtilsJWTConfig converts config.JWTConfig to utils.JWTConfig
func toUtilsJWTConfig(cfg *config.JWTConfig) *utils.JWTConfig {
	return &utils.JWTConfig{
		SecretKey:          cfg.SecretKey,
		AccessTokenExpiry:  cfg.AccessTokenExpiry,
		RefreshTokenExpiry: cfg.RefreshTokenExpiry,
		Issuer:             cfg.Issuer,
	}
}

// generateTokenResponse generates a token response for a user
func (s *service) generateTokenResponse(user *User) (*TokenResponse, error) {
	tokenPair, err := utils.GenerateTokenPair(user.ID, user.Email, toUtilsJWTConfig(s.jwtConfig))
	if err != nil {
		return nil, err
	}

	return &TokenResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresAt:    tokenPair.ExpiresAt,
		User:         user.ToResponse(),
	}, nil
}
