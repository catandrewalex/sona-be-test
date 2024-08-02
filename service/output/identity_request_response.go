package output

import (
	"fmt"
	"sonamusica-backend/app-service/identity"
	"sonamusica-backend/errs"
)

const (
	MaxPage_GetUsers           = Default_MaxPage
	MaxResultsPerPage_GetUsers = Default_MaxResultsPerPage
)

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
	return nil
}

type LoginRequest struct {
	UsernameOrEmail string `json:"usernameOrEmail"`
	Password        string `json:"password"`
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
	return nil
}

type ForgotPasswordRequest struct {
	Email string `json:"email"`
}
type ForgotPasswordResponse struct {
	Message string `json:"message,omitempty"`
}

func (r ForgotPasswordRequest) Validate() errs.ValidationError {
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
	return nil
}

type GetUserProfileRequest struct{}
type GetUserProfileResponse struct {
	Data    identity.User `json:"data"`
	Message string        `json:"message,omitempty"`
}

func (r GetUserProfileRequest) Validate() errs.ValidationError {
	return nil
}

type GetUsersRequest struct {
	PaginationRequest
	Filter             identity.GetUsersFilter `json:"filter,omitempty"`
	IncludeDeactivated bool                    `json:"includeDeactivated,omitempty"`
}
type GetUsersResponse struct {
	Data    GetUsersResult `json:"data"`
	Message string         `json:"message,omitempty"`
}
type GetUsersResult struct {
	Results []identity.User `json:"results"`
	PaginationResponse
}

func (r GetUsersRequest) Validate() errs.ValidationError {
	errorDetail := make(errs.ValidationErrorDetail, 0)
	if validationErr := r.PaginationRequest.Validate(MaxPage_GetUsers, MaxResultsPerPage_GetUsers); validationErr != nil {
		errorDetail = validationErr.GetErrorDetail()
	}

	if _, ok := identity.ValidGetUsersFilter[r.Filter]; !ok {
		errorDetail["filter"] = fmt.Sprintf("invalid filter with value '%s'", r.Filter)
	}

	if len(errorDetail) > 0 {
		return errs.NewValidationError(errs.ErrInvalidRequest, errorDetail)
	}
	return nil
}

type GetUserRequest struct {
	UserID identity.UserID `json:"-"` // we exclude the JSON tag as we'll populate the ID from URL param (not from JSON body or URL query param)
}
type GetUserResponse struct {
	Data    identity.User `json:"data"`
	Message string        `json:"message,omitempty"`
}

func (r GetUserRequest) Validate() errs.ValidationError {
	return nil
}

type InsertUsersRequest struct {
	Data []InsertUserRequestParam `json:"data"`
}
type InsertUserRequestParam struct {
	Username          string                     `json:"username"`
	Email             string                     `json:"email,omitempty"`    // admin may not know the user's email. empty is much preferable than invalid email
	Password          string                     `json:"password,omitempty"` // we allow password to be empty --> the user won't be able to login
	UserDetail        identity.UserDetail        `json:"userDetail"`
	UserPrivilegeType identity.UserPrivilegeType `json:"privilegeType"`
}
type InsertUsersResponse struct {
	Data    InsertUserResult `json:"data"`
	Message string           `json:"message,omitempty"`
}
type InsertUserResult struct {
	Results []identity.User `json:"results"`
}

func (r InsertUsersRequest) Validate() errs.ValidationError {
	return nil
}

type UpdateUsersRequest struct {
	Data []UpdateUserRequestParam `json:"data"`
}
type UpdateUserRequestParam struct {
	UserID            identity.UserID            `json:"userId"`
	Username          string                     `json:"username"`
	Email             string                     `json:"email,omitempty"`
	UserDetail        identity.UserDetail        `json:"userDetail"`
	UserPrivilegeType identity.UserPrivilegeType `json:"privilegeType"`
	IsDeactivated     bool                       `json:"isDeactivated,omitempty"` // false is a zero value, so we must allow this to be empty
}
type UpdateUsersResponse struct {
	Data    UpdateUserResult `json:"data"`
	Message string           `json:"message,omitempty"`
}
type UpdateUserResult struct {
	Results []identity.User `json:"results"`
}

func (r UpdateUsersRequest) Validate() errs.ValidationError {
	return nil
}

type UpdateUsersByUsernamesRequest struct {
	Data []UpdateUserByUsernameRequestParam `json:"data"`
}
type UpdateUserByUsernameRequestParam struct {
	Username          string                     `json:"username"`
	Email             string                     `json:"email,omitempty"`
	UserDetail        identity.UserDetail        `json:"userDetail"`
	UserPrivilegeType identity.UserPrivilegeType `json:"privilegeType"`
	IsDeactivated     bool                       `json:"isDeactivated,omitempty"` // false is a zero value, so we must allow this to be empty
}
type UpdateUsersByUsernamesResponse struct {
	Message string `json:"message,omitempty"`
}

func (r UpdateUsersByUsernamesRequest) Validate() errs.ValidationError {
	return nil
}

type UpdateUserPasswordRequest struct {
	UserID      identity.UserID `json:"-"` // we exclude the JSON tag as we'll populate the ID from URL param (not from JSON body or URL query param)
	NewPassword string          `json:"newPassword"`
}
type UpdateUserPasswordResponse struct {
	Message string `json:"message,omitempty"`
}

func (r UpdateUserPasswordRequest) Validate() errs.ValidationError {
	return nil
}
