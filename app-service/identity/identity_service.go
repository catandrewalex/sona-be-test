package identity

import (
	"context"
	"time"

	"sonamusica-backend/app-service/util"
)

type User struct {
	ID            UserID            `json:"id"`
	Username      string            `json:"username"`
	Email         string            `json:"email"`
	UserDetail    UserDetail        `json:"userDetail"`
	PrivilegeType UserPrivilegeType `json:"privilegeType"`
	IsDeactivated bool              `json:"isDeactivated"`
	CreatedAt     time.Time         `json:"createdAt"`
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

type IdentityService interface {
	GetUsers(ctx context.Context, pagination util.PaginationSpec) (GetUsersResult, error)
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
	ID                UserID
	Username          string
	Email             string
	UserDetail        UserDetail
	UserPrivilegeType UserPrivilegeType
	IsDeactivated     bool
}

type UpdateUserPasswordSpec struct {
	ID       UserID
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
