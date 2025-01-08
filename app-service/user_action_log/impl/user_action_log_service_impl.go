package impl

import (
	"context"
	"database/sql"
	"fmt"
	"sonamusica-backend/accessor/relational_db"
	"sonamusica-backend/accessor/relational_db/mysql"
	"sonamusica-backend/app-service/identity"
	"sonamusica-backend/app-service/user_action_log"
	"sonamusica-backend/app-service/util"
	"sonamusica-backend/config"
	"sonamusica-backend/logging"
)

var (
	configObject = config.Get()
	mainLog      = logging.NewGoLogger("UserActionLog", logging.GetLevel(configObject.LogLevel))
)

type userActionLogImpl struct {
	mySQLQueries *relational_db.MySQLQueries
}

var _ user_action_log.UserActionLogService = (*userActionLogImpl)(nil)

func NewUserActionLogImpl(mySQLQueries *relational_db.MySQLQueries) *userActionLogImpl {
	return &userActionLogImpl{
		mySQLQueries: mySQLQueries,
	}
}

func (s *userActionLogImpl) GetUserActionLogs(ctx context.Context, pagination util.PaginationSpec, spec user_action_log.GetUserActionLogSpec) (user_action_log.GetUserActionLogsResult, error) {
	pagination.SetDefaultOnInvalidValues()
	limit, offset := pagination.GetLimitAndOffset()

	timeFilter := spec.TimeSpec
	timeFilter.SetDefaultForZeroValues()

	useUserIDFilter := spec.UserID != identity.UserID_None
	usePrivilegeTypeFilter := spec.PrivilegeType != identity.UserPrivilegeType_None
	useMethodFilter := spec.Method != ""
	useStatusCodeFilter := spec.StatusCode != 0

	logRows, err := s.mySQLQueries.GetUserActionLogs(ctx, mysql.GetUserActionLogsParams{
		StartDate:              timeFilter.StartDatetime,
		EndDate:                timeFilter.EndDatetime,
		UserID:                 sql.NullInt64{Int64: int64(spec.UserID), Valid: true},
		UseUserIDFilter:        useUserIDFilter,
		PrivilegeType:          int32(spec.PrivilegeType),
		UsePrivilegeTypeFilter: usePrivilegeTypeFilter,
		Method:                 spec.Method,
		UseMethodFilter:        useMethodFilter,
		StatusCode:             spec.StatusCode,
		UseStatusCodeFilter:    useStatusCodeFilter,
		Limit:                  int32(limit),
		Offset:                 int32(offset),
	})
	if err != nil {
		return user_action_log.GetUserActionLogsResult{}, fmt.Errorf("mySQLQueries.GetUserActionLogs(): %w", err)
	}

	userActionLogs := NewUserActionLogsFromMySQLUserActionLogs(logRows)

	totalResults, err := s.mySQLQueries.CountUserActionLogs(ctx, mysql.CountUserActionLogsParams{
		StartDate:              timeFilter.StartDatetime,
		EndDate:                timeFilter.EndDatetime,
		UserID:                 sql.NullInt64{Int64: int64(spec.UserID), Valid: true},
		UseUserIDFilter:        useUserIDFilter,
		PrivilegeType:          int32(spec.PrivilegeType),
		UsePrivilegeTypeFilter: usePrivilegeTypeFilter,
		Method:                 spec.Method,
		UseMethodFilter:        useMethodFilter,
		StatusCode:             spec.StatusCode,
		UseStatusCodeFilter:    useStatusCodeFilter,
	})
	if err != nil {
		return user_action_log.GetUserActionLogsResult{}, fmt.Errorf("mySQLQueries.CountUserActionLogs(): %w", err)
	}

	return user_action_log.GetUserActionLogsResult{
		UserActionLogs:   userActionLogs,
		PaginationResult: *util.NewPaginationResult(int(totalResults), pagination.ResultsPerPage, pagination.Page),
	}, nil

}
func (s *userActionLogImpl) InsertUserActionLogs(ctx context.Context, specs []user_action_log.InsertUserActionLogSpec) ([]user_action_log.UserActionLogID, error) {
	userActionLogs := make([]user_action_log.UserActionLogID, 0, len(specs))

	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		for _, spec := range specs {
			userActionLog, err := qtx.InsertUserActionLog(newCtx, mysql.InsertUserActionLogParams{
				Date:          spec.Date,
				UserID:        sql.NullInt64{Int64: int64(spec.UserID), Valid: true},
				PrivilegeType: int32(spec.PrivilegeType),
				Endpoint:      spec.Endpoint,
				Method:        spec.Method,
				StatusCode:    spec.StatusCode,
				RequestBody:   spec.RequestBody,
			})
			if err != nil {
				return fmt.Errorf("qtx.InsertCourse(): %w", err)
			}
			userActionLogs = append(userActionLogs, user_action_log.UserActionLogID(userActionLog))
		}
		return nil
	})
	if err != nil {
		return []user_action_log.UserActionLogID{}, fmt.Errorf("ExecuteInTransaction(): %w", err)
	}

	return userActionLogs, nil
}

func (s *userActionLogImpl) DeleteUserActionLogsByIds(ctx context.Context, ids []user_action_log.UserActionLogID) error {
	userActionLogIdsInt64 := make([]int64, 0, len(ids))
	for _, id := range ids {
		userActionLogIdsInt64 = append(userActionLogIdsInt64, int64(id))
	}

	err := s.mySQLQueries.DeleteUserActionLogsByIds(ctx, userActionLogIdsInt64)
	if err != nil {
		return fmt.Errorf("mySQLQueries.DeleteUserActionLogsByIds(): %w", err)
	}

	return nil
}

func (s *userActionLogImpl) DeleteUserActionLogs(ctx context.Context, spec user_action_log.DeleteUserActionLogSpec) (int64, error) {
	timeFilter := spec.TimeSpec
	timeFilter.SetDefaultForZeroValues()

	useUserIDFilter := spec.UserID != identity.UserID_None
	usePrivilegeTypeFilter := spec.PrivilegeType != identity.UserPrivilegeType_None
	useMethodFilter := spec.Method != ""
	useStatusCodeFilter := spec.StatusCode != 0

	totalRows, err := s.mySQLQueries.DeleteUserActionLogs(ctx, mysql.DeleteUserActionLogsParams{
		StartDate:              timeFilter.StartDatetime,
		EndDate:                timeFilter.EndDatetime,
		UserID:                 sql.NullInt64{Int64: int64(spec.UserID), Valid: true},
		UseUserIDFilter:        useUserIDFilter,
		PrivilegeType:          int32(spec.PrivilegeType),
		UsePrivilegeTypeFilter: usePrivilegeTypeFilter,
		Method:                 spec.Method,
		UseMethodFilter:        useMethodFilter,
		StatusCode:             spec.StatusCode,
		UseStatusCodeFilter:    useStatusCodeFilter,
	})
	if err != nil {
		return 0, fmt.Errorf("mySQLQueries.DeleteUserActionLogs(): %w", err)
	}

	return totalRows, nil
}
