package output

import (
	"sonamusica-backend/app-service/identity"
	"sonamusica-backend/errs"
)

type UserDataRequest struct {
	ID int `json:"id"`
}
type UserDataResponse struct {
	Data    identity.User `json:"data"`
	Message string        `json:"message,omitempty"`
}

func (r UserDataRequest) Validate() errs.ValidationError {
	errorDetail := make(errs.ValidationErrorDetail, 0)
	if r.ID == 0 {
		errorDetail["id"] = "id cannot be empty"
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
	Message string `json:"message,omitempty"`
}

func (r SignUpRequest) Validate() errs.ValidationError {
	errorDetail := make(errs.ValidationErrorDetail, 0)
	if r.Email == "" {
		errorDetail["email"] = "email cannot be empty"
	}
	if r.Username == "" {
		errorDetail["username"] = "username cannot be empty"
	}
	if r.Password == "" {
		errorDetail["password"] = "password cannot be empty"
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
	Data    LoginResult `json:"data"`
	Message string      `json:"message,omitempty"`
}

type LoginResult struct {
	User      identity.User      `json:"user"`
	AuthToken identity.AuthToken `json:"authToken"`
}

func (r LoginRequest) Validate() errs.ValidationError {
	errorDetail := make(errs.ValidationErrorDetail, 0)
	if r.Email == "" {
		errorDetail["email"] = "email cannot be empty"
	}
	if r.Password == "" {
		errorDetail["password"] = "password cannot be empty"
	}

	if len(errorDetail) > 0 {
		return errs.NewValidationError(errs.ErrInvalidRequest, errorDetail)
	}

	return nil
}

type ForgotPasswordRequest struct {
	Email string `json:"email"`
}
type ForgotPasswordResponse struct {
	Message string `json:"message,omitempty"`
}

func (r ForgotPasswordRequest) Validate() errs.ValidationError {
	errorDetail := make(errs.ValidationErrorDetail, 0)
	if r.Email == "" {
		errorDetail["email"] = "email cannot be empty"
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
type ResetPasswordResponse struct {
	Message string `json:"message,omitempty"`
}

func (r ResetPasswordRequest) Validate() errs.ValidationError {
	errorDetail := make(errs.ValidationErrorDetail, 0)
	if r.ResetToken == "" {
		errorDetail["resetToken"] = "resetToken cannot be empty"
	}
	if r.NewPassword == "" {
		errorDetail["newPassword"] = "newPassword cannot be empty"
	}

	if len(errorDetail) > 0 {
		return errs.NewValidationError(errs.ErrInvalidRequest, errorDetail)
	}

	return nil
}
