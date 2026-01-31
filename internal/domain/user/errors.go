package user

import "github.com/voidmaindev/go-template/internal/common/errors"

const domainName = "user"

// Domain-specific errors for user operations
var (
	// ErrUserNotFound is returned when a user cannot be found
	ErrUserNotFound = errors.NotFound(domainName, "user")

	// ErrEmailExists is returned when trying to register with an existing email
	ErrEmailExists = errors.AlreadyExists(domainName, "user", "email")

	// ErrInvalidPassword is returned when password verification fails
	ErrInvalidPassword = errors.Validation(domainName, "invalid password")

	// ErrSamePassword is returned when new password is same as current
	ErrSamePassword = errors.Validation(domainName, "new password must be different from current password")

	// ErrNoPassword is returned when user has no password set (OAuth-only user)
	ErrNoPassword = errors.Validation(domainName, "user does not have a password set")

	// ErrInvalidCredentials is returned for failed login attempts
	ErrInvalidCredentials = errors.New(domainName, errors.CodeUnauthorized).
				WithMessage("invalid credentials")

	// ErrTokenRefreshUnavailable is returned when token refresh fails
	ErrTokenRefreshUnavailable = errors.New(domainName, errors.CodeInternal).
					WithMessage("token refresh temporarily unavailable")

	// ErrTokenExpired is returned when a token has expired
	ErrTokenExpired = errors.New(domainName, errors.CodeUnauthorized).
			WithMessage("token expired")

	// ErrTokenInvalid is returned when a token is invalid
	ErrTokenInvalid = errors.New(domainName, errors.CodeUnauthorized).
			WithMessage("invalid token")

	// ErrTokenBlacklisted is returned when a token has been revoked
	ErrTokenBlacklisted = errors.New(domainName, errors.CodeUnauthorized).
				WithMessage("token has been revoked")

	// ErrEmailNotVerified is returned when a self-registered user tries to login without verifying email
	ErrEmailNotVerified = errors.New(domainName, errors.CodeForbidden).
				WithMessage("email address not verified")

	// ErrTooManyLoginAttempts is returned when login rate limit is exceeded
	ErrTooManyLoginAttempts = errors.New(domainName, errors.CodeTooManyRequests).
				WithMessage("too many login attempts, please try again later")
)
