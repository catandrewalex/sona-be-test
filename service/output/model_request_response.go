package output

import (
	"sonamusica-backend/app-service/identity"
	"sonamusica-backend/errs"
)

type UserProfileRequest struct {
	ID int `json:"id"`
}
type UserProfileResponse struct {
	identity.User
}

func (r UserProfileRequest) Validate() errs.ValidationError {
	errorDetail := make(errs.ValidationErrorDetail, 0)
	if r.ID == 0 {
		errorDetail["id"] = "cannot be empty"
	}

	if len(errorDetail) > 0 {
		return errs.NewValidationError(errs.ErrInvalidRequest, errorDetail)
	}
	return nil
}

type SignUpRequest struct {
	Email      string              `json:"email"`
	Password   string              `json:"password"`
	Username   string              `json:"username"`
	UserDetail identity.UserDetail `json:"userDetail"`
}
type SignUpResponse struct {
	Message string `json:"message"`
}

func (r SignUpRequest) Validate() errs.ValidationError {
	errorDetail := make(errs.ValidationErrorDetail, 0)
	if r.Email == "" {
		errorDetail["email"] = "cannot be empty"
	}
	if r.Username == "" {
		errorDetail["username"] = "cannot be empty"
	}
	if r.Password == "" {
		errorDetail["password"] = "cannot be empty"
	}

	if len(errorDetail) > 0 {
		return errs.NewValidationError(errs.ErrInvalidRequest, errorDetail)
	}
	return nil
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
type LoginResponse struct {
	Message string `json:"message"`
}

func (r LoginRequest) Validate() errs.ValidationError {
	errorDetail := make(errs.ValidationErrorDetail, 0)
	if r.Email == "" {
		errorDetail["email"] = "cannot be empty"
	}
	if r.Password == "" {
		errorDetail["password"] = "cannot be empty"
	}

	if len(errorDetail) > 0 {
		return errs.NewValidationError(errs.ErrInvalidRequest, errorDetail)
	}

	return nil
}

type ForgotPasswordRequest struct {
	Email string `json:"email"`
}
type ForgotPasswordResponse struct{}

func (r ForgotPasswordRequest) Validate() errs.ValidationError {
	errorDetail := make(errs.ValidationErrorDetail, 0)
	if r.Email == "" {
		errorDetail["email"] = "cannot be empty"
	}

	if len(errorDetail) > 0 {
		return errs.NewValidationError(errs.ErrInvalidRequest, errorDetail)
	}

	return nil
}

type ResetPasswordRequest struct {
	ResetToken  string `json:"resetToken"`
	NewPassword string `json:"newPassword"`
}
type ResetPasswordResponse struct{}

func (r ResetPasswordRequest) Validate() errs.ValidationError {
	errorDetail := make(errs.ValidationErrorDetail, 0)
	if r.NewPassword == "" {
		errorDetail["newPassword"] = "cannot be empty"
	}

	if len(errorDetail) > 0 {
		return errs.NewValidationError(errs.ErrInvalidRequest, errorDetail)
	}

	return nil
}
