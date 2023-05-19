package identity

import (
	"context"
	"time"
)

type User struct {
	ID            UserID            `json:"id"`
	Username      string            `json:"username"`
	Email         string            `json:"email"`
	UserDetail    UserDetail        `json:"userDetail"`
	PrivilegeType UserPrivilegeType `json:"privilegeType"`
	CreatedAt     time.Time         `json:"createdAt"`
}

type UserID int64

const (
	UserID_None UserID = iota
)

type AuthToken string

type UserPrivilegeType int32

const (
	UserPrivilegeType_None      UserPrivilegeType = iota
	UserPrivilegeType_Anonymous UserPrivilegeType = 100
	UserPrivilegeType_Member    UserPrivilegeType = 200
	UserPrivilegeType_Staff     UserPrivilegeType = 300
	UserPrivilegeType_Admin     UserPrivilegeType = 400
)

type IdentityService interface {
	GetUserById(ctx context.Context, id UserID) (User, error)

	SignUpUser(ctx context.Context, spec SignUpUserSpec) (UserID, error)
	LoginUser(ctx context.Context, spec LoginUserSpec) (LoginUserResult, error)
	ForgotPassword(ctx context.Context, spec ForgotPasswordSpec) error
	ResetPassword(ctx context.Context, spec ResetPasswordSpec) error
}

type SignUpUserSpec struct {
	Email      string
	Password   string
	Username   string
	UserDetail UserDetail
}

type UserDetail struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
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
