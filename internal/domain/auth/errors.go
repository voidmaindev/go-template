package auth

import "github.com/voidmaindev/go-template/internal/common/errors"

const domainName = "auth"

// Domain-specific errors for auth operations
var (
	// ErrSelfRegDisabled is returned when self-registration is disabled
	ErrSelfRegDisabled = errors.New(domainName, errors.CodeForbidden).
				WithMessage("self-registration is disabled")

	// ErrEmailNotVerified is returned when email verification is required but not completed
	ErrEmailNotVerified = errors.New(domainName, errors.CodeForbidden).
				WithMessage("email address not verified")

	// ErrInvalidToken is returned when a verification/reset token is invalid
	ErrInvalidToken = errors.New(domainName, errors.CodeBadRequest).
			WithMessage("invalid or expired token")

	// ErrTokenExpired is returned when a token has expired
	ErrTokenExpired = errors.New(domainName, errors.CodeBadRequest).
			WithMessage("token has expired")

	// ErrEmailAlreadyVerified is returned when trying to verify an already verified email
	ErrEmailAlreadyVerified = errors.New(domainName, errors.CodeBadRequest).
				WithMessage("email already verified")

	// ErrUserNotFound is returned when user cannot be found (deliberately vague)
	ErrUserNotFound = errors.New(domainName, errors.CodeBadRequest).
			WithMessage("if an account exists with this email, you will receive an email")

	// ErrRateLimited is returned when too many requests have been made
	ErrRateLimited = errors.New(domainName, errors.CodeTooManyRequests).
			WithMessage("too many requests, please try again later")

	// ErrOAuthDisabled is returned when OAuth is disabled for a provider
	ErrOAuthDisabled = errors.New(domainName, errors.CodeForbidden).
				WithMessage("OAuth provider is disabled")

	// ErrOAuthStateMismatch is returned when OAuth state doesn't match
	ErrOAuthStateMismatch = errors.New(domainName, errors.CodeBadRequest).
				WithMessage("OAuth state mismatch")

	// ErrOAuthCodeExchange is returned when OAuth code exchange fails
	ErrOAuthCodeExchange = errors.New(domainName, errors.CodeBadRequest).
				WithMessage("failed to exchange OAuth code")

	// ErrOAuthUserInfo is returned when getting OAuth user info fails
	ErrOAuthUserInfo = errors.New(domainName, errors.CodeBadRequest).
				WithMessage("failed to get user information from OAuth provider")

	// ErrIdentityAlreadyLinked is returned when trying to link an already linked identity
	ErrIdentityAlreadyLinked = errors.New(domainName, errors.CodeConflict).
					WithMessage("this OAuth identity is already linked to an account")

	// ErrIdentityNotFound is returned when trying to unlink a non-existent identity
	ErrIdentityNotFound = errors.NotFound(domainName, "identity")

	// ErrCannotUnlinkLastIdentity is returned when trying to unlink the last login method
	ErrCannotUnlinkLastIdentity = errors.New(domainName, errors.CodeBadRequest).
					WithMessage("cannot unlink the only login method")

	// ErrPasswordRequired is returned when password is required but not provided
	ErrPasswordRequired = errors.Validation(domainName, "password is required for email registration")

	// ErrTooManyLoginAttempts is returned when login rate limit is exceeded
	ErrTooManyLoginAttempts = errors.New(domainName, errors.CodeTooManyRequests).
				WithMessage("too many login attempts, please try again later")
)
