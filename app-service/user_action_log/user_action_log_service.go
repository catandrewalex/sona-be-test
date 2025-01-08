package user_action_log

import (
	"context"
	"time"

	"sonamusica-backend/app-service/identity"
	"sonamusica-backend/app-service/util"
)

type UserActionLogID int64

const (
	UserActionLogID_None UserActionLogID = iota
)

type UserActionLog struct {
	ID            UserActionLogID            `json:"id"`
	Date          time.Time                  `json:"date"`
	UserID        identity.UserID            `json:"userId"`
	Username      string                     `json:"username"`
	PrivilegeType identity.UserPrivilegeType `json:"privilegeType"`
	Endpoint      string                     `json:"endpoint"`
	Method        string                     `json:"method"`
	StatusCode    uint16                     `json:"statusCode"`
	RequestBody   string                     `json:"requestBody"`
}

type UserActionLogService interface {
	GetUserActionLogs(ctx context.Context, pagination util.PaginationSpec, spec GetUserActionLogSpec) (GetUserActionLogsResult, error)
	InsertUserActionLogs(ctx context.Context, specs []InsertUserActionLogSpec) ([]UserActionLogID, error)
	DeleteUserActionLogsByIds(ctx context.Context, ids []UserActionLogID) error
	DeleteUserActionLogs(ctx context.Context, spec DeleteUserActionLogSpec) (int64, error)
}

type InsertUserActionLogSpec struct {
	Date          time.Time
	UserID        identity.UserID
	PrivilegeType identity.UserPrivilegeType
	Endpoint      string
	Method        string
	StatusCode    uint16
	RequestBody   string
}

type GetUserActionLogSpec struct {
	util.TimeSpec
	UserID        identity.UserID
	PrivilegeType identity.UserPrivilegeType
	Method        string
	StatusCode    uint16
}

type GetUserActionLogsResult struct {
	UserActionLogs   []UserActionLog
	PaginationResult util.PaginationResult
}

type DeleteUserActionLogSpec struct {
	util.TimeSpec
	UserID        identity.UserID
	PrivilegeType identity.UserPrivilegeType
	Method        string
	StatusCode    uint16
}
