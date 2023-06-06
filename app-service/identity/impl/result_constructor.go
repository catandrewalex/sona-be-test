package impl

import (
	"sonamusica-backend/accessor/relational_db/mysql"
	"sonamusica-backend/app-service/identity"
	"sonamusica-backend/app-service/util"
)

func NewUsersFromMySQLUsers(userRows []mysql.User) []identity.User {
	users := make([]identity.User, 0, len(userRows))
	for _, userRow := range userRows {
		users = append(users, identity.User{
			ID:            identity.UserID(userRow.ID),
			Username:      userRow.Username,
			Email:         userRow.Email,
			UserDetail:    identity.UnmarshalUserDetail(userRow.UserDetail, mainLog),
			PrivilegeType: identity.UserPrivilegeType(userRow.PrivilegeType),
			IsDeactivated: util.Int32ToBool(userRow.IsDeactivated),
			CreatedAt:     userRow.CreatedAt.Time,
		})
	}

	return users
}
