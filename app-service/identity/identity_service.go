package identity

import (
	"context"
	"encoding/json"
	"time"

	"sonamusica-backend/app-service/util"
	"sonamusica-backend/logging"
)

type User struct {
	UserID        UserID            `json:"userId"`
	Username      string            `json:"username"`
	Email         string            `json:"email"`
	UserDetail    UserDetail        `json:"userDetail"`
	PrivilegeType UserPrivilegeType `json:"privilegeType"`
	IsDeactivated bool              `json:"isDeactivated"`
	CreatedAt     time.Time         `json:"createdAt"`
}

// UserInfo_Minimal is a subset of struct User that must have the same schema.
type UserInfo_Minimal struct {
	Username   string     `json:"username"`
	UserDetail UserDetail `json:"userDetail"`
}

type UserID int64

const (
	UserID_None UserID = iota
)

type AuthToken string

type UserDetail struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName,omitempty"`
}

type UserPrivilegeType int32

const (
	UserPrivilegeType_None      UserPrivilegeType = iota
	UserPrivilegeType_Anonymous UserPrivilegeType = 100
	UserPrivilegeType_Member    UserPrivilegeType = 200
	UserPrivilegeType_Staff     UserPrivilegeType = 300
	UserPrivilegeType_Admin     UserPrivilegeType = 400
)

type GetUsersFilter string

const (
	GetUsersFilter_None       GetUsersFilter = ""
	GetUsersFilter_NotTeacher GetUsersFilter = "NOT_TEACHER"
	GetUsersFilter_NotStudent GetUsersFilter = "NOT_STUDENT"
)

var ValidGetUsersFilter = map[GetUsersFilter]struct{}{
	GetUsersFilter_None:       {},
	GetUsersFilter_NotTeacher: {},
	GetUsersFilter_NotStudent: {},
}

func (f GetUsersFilter) String() string {
	return string(f)
}

type IdentityService interface {
	GetUsers(ctx context.Context, pagination util.PaginationSpec, filter GetUsersFilter, includeDeactivated bool) (GetUsersResult, error)
	GetUserById(ctx context.Context, id UserID) (User, error)
	GetUsersByIds(ctx context.Context, ids []UserID) ([]User, error)
	InsertUsers(ctx context.Context, specs []InsertUserSpec) ([]UserID, error)
	UpdateUserInfos(ctx context.Context, specs []UpdateUserInfoSpec) ([]UserID, error)
	UpdateUserPassword(ctx context.Context, spec UpdateUserPasswordSpec) error

	SignUpUser(ctx context.Context, spec SignUpUserSpec) (UserID, error)
	LoginUser(ctx context.Context, spec LoginUserSpec) (LoginUserResult, error)
	ForgotPassword(ctx context.Context, spec ForgotPasswordSpec) error
	ResetPassword(ctx context.Context, spec ResetPasswordSpec) error
}

func UnmarshalUserDetail(jsonRaw json.RawMessage, logger logging.Logger) UserDetail {
	var userDetail UserDetail
	err := json.Unmarshal(jsonRaw, &userDetail)
	if err != nil {
		logger.Warn("Unable to unmarshal UserDetail=%q: %v", jsonRaw, err)
	}
	return userDetail
}

type GetUsersResult struct {
	Users            []User
	PaginationResult util.PaginationResult
}

type InsertUserSpec struct {
	Email             string
	Password          string
	Username          string
	UserDetail        UserDetail
	UserPrivilegeType UserPrivilegeType
}

type UpdateUserInfoSpec struct {
	UserID            UserID
	Username          string
	Email             string
	UserDetail        UserDetail
	UserPrivilegeType UserPrivilegeType
	IsDeactivated     bool
}

func (s UpdateUserInfoSpec) GetInt64ID() int64 {
	return int64(s.UserID)
}

type UpdateUserPasswordSpec struct {
	UserID   UserID
	Password string
}

type SignUpUserSpec struct {
	Email      string
	Password   string
	Username   string
	UserDetail UserDetail
}

type LoginUserSpec struct {
	UsernameOrEmail string
	Password        string
}
type LoginUserResult struct {
	User      User
	AuthToken AuthToken
}

type ForgotPasswordSpec struct {
	Email string
}

type ResetPasswordSpec struct {
	ResetToken  string
	NewPassword string
}
