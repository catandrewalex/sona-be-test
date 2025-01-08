package impl

import (
	"sonamusica-backend/accessor/relational_db/mysql"
	"sonamusica-backend/app-service/identity"
	"sonamusica-backend/app-service/user_action_log"
)

func NewUserActionLogsFromMySQLUserActionLogs(logRows []mysql.GetUserActionLogsRow) []user_action_log.UserActionLog {
	logs := make([]user_action_log.UserActionLog, 0, len(logRows))
	for _, logRow := range logRows {
		logs = append(logs, user_action_log.UserActionLog{
			ID:            user_action_log.UserActionLogID(logRow.ID),
			Date:          logRow.Date,
			UserID:        identity.UserID(logRow.UserID.Int64),
			Username:      logRow.Username.String,
			Endpoint:      logRow.Endpoint,
			Method:        logRow.Method,
			StatusCode:    logRow.StatusCode,
			PrivilegeType: identity.UserPrivilegeType(logRow.PrivilegeType),
			RequestBody:   logRow.RequestBody,
		})
	}

	return logs
}
