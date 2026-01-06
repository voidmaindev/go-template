package user

import (
	"context"

	"github.com/voidmaindev/go-template/internal/common/validation"
)

// Validator handles user domain validation
type Validator struct {
	repo      Repository
	validator *validation.Validator
}

// NewValidator creates a user validator
func NewValidator(repo Repository) *Validator {
	return &Validator{
		repo:      repo,
		validator: validation.New(),
	}
}

// ValidateRegister validates registration request
func (v *Validator) ValidateRegister(ctx context.Context, req *RegisterRequest) *validation.Result {
	return validation.NewBuilder(ctx).
		StructWith(v.validator, req).
		CustomWithCode("email", "ALREADY_EXISTS", func(ctx context.Context) error {
			exists, err := v.repo.ExistsByEmail(ctx, req.Email)
			if err != nil {
				return err
			}
			if exists {
				return ErrEmailExists
			}
			return nil
		}).
		Result()
}

// ValidateLogin validates login request
func (v *Validator) ValidateLogin(ctx context.Context, req *LoginRequest) *validation.Result {
	return validation.NewBuilder(ctx).
		StructWith(v.validator, req).
		Result()
}

// ValidateUpdate validates update request
func (v *Validator) ValidateUpdate(ctx context.Context, req *UpdateUserRequest) *validation.Result {
	return validation.NewBuilder(ctx).
		StructWith(v.validator, req).
		Result()
}

// ValidateChangePassword validates password change request
func (v *Validator) ValidateChangePassword(ctx context.Context, req *ChangePasswordRequest) *validation.Result {
	return validation.NewBuilder(ctx).
		StructWith(v.validator, req).
		Condition(
			req.NewPassword == req.CurrentPassword,
			"new_password",
			"SAME_PASSWORD",
			"new password must be different from current password",
		).
		Result()
}

// ValidateRefreshToken validates refresh token request
func (v *Validator) ValidateRefreshToken(ctx context.Context, req *RefreshTokenRequest) *validation.Result {
	return validation.NewBuilder(ctx).
		StructWith(v.validator, req).
		Result()
}
