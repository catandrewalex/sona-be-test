package impl

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/matcornic/hermes/v2"
	"golang.org/x/crypto/bcrypt"

	"sonamusica-backend/accessor/email"
	"sonamusica-backend/accessor/relational_db"
	"sonamusica-backend/accessor/relational_db/mysql"
	"sonamusica-backend/app-service/auth"
	"sonamusica-backend/app-service/email_composer"
	"sonamusica-backend/app-service/identity"
	"sonamusica-backend/app-service/util"
	"sonamusica-backend/config"
	"sonamusica-backend/errs"
	"sonamusica-backend/logging"
	"sonamusica-backend/network"
)

var (
	configObject = config.Get()
	mainLog      = logging.NewGoLogger("IdentityService", logging.GetLevel(configObject.LogLevel))
)

type identityServiceImpl struct {
	mySQLQueries *relational_db.MySQLQueries

	smtpAccessor email.SMTPAccessor

	jwtService auth.JWTService

	emailComposer *hermes.Hermes
}

var _ identity.IdentityService = (*identityServiceImpl)(nil)

func NewIdentityServiceImpl(mySQLQueries *relational_db.MySQLQueries, smtpAccessor email.SMTPAccessor, jwtService auth.JWTService, emailComposer *hermes.Hermes) *identityServiceImpl {
	return &identityServiceImpl{
		mySQLQueries:  mySQLQueries,
		smtpAccessor:  smtpAccessor,
		jwtService:    jwtService,
		emailComposer: emailComposer,
	}
}

func (s identityServiceImpl) GetUsers(ctx context.Context, pagination util.PaginationSpec, filter identity.GetUsersFilter, includeDeactivated bool) (identity.GetUsersResult, error) {
	pagination.SetDefaultOnInvalidValues()
	limit, offset := pagination.GetLimitAndOffset()
	isDeactivatedFilters := []int32{0}
	if includeDeactivated {
		isDeactivatedFilters = append(isDeactivatedFilters, 1)
	}

	userRows, totalResults, err := s.getUsersWithFilter(ctx, limit, offset, filter, isDeactivatedFilters)
	if err != nil {
		return identity.GetUsersResult{}, err
	}

	users := NewUsersFromMySQLUsers(userRows)

	return identity.GetUsersResult{
		Users:            users,
		PaginationResult: *util.NewPaginationResult(int(totalResults), pagination.ResultsPerPage, pagination.Page),
	}, nil
}

func (s identityServiceImpl) getUsersWithFilter(ctx context.Context, limit, offset int, filter identity.GetUsersFilter, isDeactivatedFilters []int32) ([]mysql.User, int64, error) {
	var userRows []mysql.User
	var totalResults int64
	var err error
	var errCount error

	switch filter {
	case identity.GetUsersFilter_None:
		userRows, err = s.mySQLQueries.GetUsers(ctx, mysql.GetUsersParams{
			IsDeactivateds: isDeactivatedFilters,
			Limit:          int32(limit),
			Offset:         int32(offset),
		})
		totalResults, errCount = s.mySQLQueries.CountUsers(ctx, isDeactivatedFilters)
	case identity.GetUsersFilter_NotTeacher:
		userNotTeacherRows, err2 := s.mySQLQueries.GetUsersNotTeacher(ctx, mysql.GetUsersNotTeacherParams{
			IsDeactivateds: isDeactivatedFilters,
			Limit:          int32(limit),
			Offset:         int32(offset),
		})
		err = err2
		for _, row := range userNotTeacherRows {
			userRows = append(userRows, row.User)
		}
		totalResults, errCount = s.mySQLQueries.CountUsersNotTeacher(ctx, isDeactivatedFilters)
	case identity.GetUsersFilter_NotStudent:
		userNotStudentRows, err2 := s.mySQLQueries.GetUsersNotStudent(ctx, mysql.GetUsersNotStudentParams{
			IsDeactivateds: isDeactivatedFilters,
			Limit:          int32(limit),
			Offset:         int32(offset),
		})
		err = err2
		for _, row := range userNotStudentRows {
			userRows = append(userRows, row.User)
		}
		totalResults, errCount = s.mySQLQueries.CountUsersNotStudent(ctx, isDeactivatedFilters)
	default:
		panic(fmt.Sprintf("unhandled GetUsersFilter = '%s'", filter))
	}

	if err != nil {
		return []mysql.User{}, 0, fmt.Errorf("mySQLQueries.GetUsers() with filter = '%s': %w", filter, err)
	}
	if errCount != nil {
		return []mysql.User{}, 0, fmt.Errorf("mySQLQueries.CountUsers() with filter = '%s': %w", filter, err)
	}

	return userRows, totalResults, nil
}

func (s identityServiceImpl) GetUserById(ctx context.Context, id identity.UserID) (identity.User, error) {
	userRow, err := s.mySQLQueries.GetUserById(ctx, int64(id))
	if err != nil {
		return identity.User{}, fmt.Errorf("mySQLQueries.GetUserById(): %w", err)
	}

	user := NewUsersFromMySQLUsers([]mysql.User{userRow})[0]

	return user, nil
}

func (s identityServiceImpl) GetUsersByIds(ctx context.Context, ids []identity.UserID) ([]identity.User, error) {
	idsInt := make([]int64, 0, len(ids))
	for _, id := range ids {
		idsInt = append(idsInt, int64(id))
	}

	userRows, err := s.mySQLQueries.GetUsersByIds(ctx, idsInt)
	if err != nil {
		return []identity.User{}, fmt.Errorf("mySQLQueries.GetUsersByIds(): %w", err)
	}

	users := NewUsersFromMySQLUsers(userRows)

	return users, nil
}

func (s identityServiceImpl) InsertUsers(ctx context.Context, specs []identity.InsertUserSpec) ([]identity.UserID, error) {
	hashedPasswords := make([]string, 0, len(specs))
	userDetails := make([][]byte, 0, len(specs))

	for _, spec := range specs {
		hashedPassword, err := hashPassword(spec.Password)
		if err != nil {
			return []identity.UserID{}, fmt.Errorf("hashPassword(): %w", err)
		}
		hashedPasswords = append(hashedPasswords, hashedPassword)

		userDetail, err := json.Marshal(spec.UserDetail)
		if err != nil {
			return []identity.UserID{}, fmt.Errorf("marshal UserDetail: %w", err)
		}
		userDetails = append(userDetails, userDetail)
	}

	userIds := make([]identity.UserID, 0, len(specs))
	// We insert into 2 tables & also in bulk --> need to wrap inside SQL transaction
	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		for i, spec := range specs {
			email := strings.TrimSpace(spec.Email)
			sqlEmail := sql.NullString{String: email, Valid: email != ""}
			userID, err := qtx.InsertUser(newCtx, mysql.InsertUserParams{
				Email:         sqlEmail,
				Username:      spec.Username,
				UserDetail:    userDetails[i],
				PrivilegeType: int32(spec.UserPrivilegeType),
			})
			if err != nil {
				return fmt.Errorf("qtx.InsertUser(): %w", err)
			}

			_, err = qtx.InsertUserCredential(newCtx, mysql.InsertUserCredentialParams{
				UserID:   userID,
				Username: spec.Username,
				Password: hashedPasswords[i],
				Email:    sqlEmail,
			})
			if err != nil {
				return fmt.Errorf("qtx.InsertUserCredential(): %w", err)
			}

			userIds = append(userIds, identity.UserID(userID))
		}
		return nil
	})
	if err != nil {
		return []identity.UserID{}, fmt.Errorf("ExecuteInTransaction(): %w", err)
	}

	return userIds, nil
}

func (s identityServiceImpl) UpdateUserInfos(ctx context.Context, specs []identity.UpdateUserInfoSpec) ([]identity.UserID, error) {
	err := util.ValidateUpdateSpecs(ctx, specs, s.mySQLQueries.CountUsersByIds)
	if err != nil {
		return []identity.UserID{}, err
	}

	userDetails := make([][]byte, 0, len(specs))
	for _, spec := range specs {
		userDetail, err := json.Marshal(spec.UserDetail)
		if err != nil {
			return []identity.UserID{}, fmt.Errorf("marshal UserDetail: %w", err)
		}
		userDetails = append(userDetails, userDetail)
	}

	userIDs := make([]identity.UserID, 0, len(specs))

	err = s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		for i, spec := range specs {
			email := strings.TrimSpace(spec.Email)
			sqlEmail := sql.NullString{String: email, Valid: email != ""}
			err := qtx.UpdateUser(newCtx, mysql.UpdateUserParams{
				Username:      spec.Username,
				Email:         sqlEmail,
				UserDetail:    userDetails[i],
				PrivilegeType: int32(spec.UserPrivilegeType),
				IsDeactivated: util.BoolToInt32(spec.IsDeactivated),
				ID:            int64(spec.UserID),
			})
			if err != nil {
				return fmt.Errorf("qtx.UpdateUser(): %w", err)
			}

			err = qtx.UpdateUserCredentialInfoByUserId(newCtx, mysql.UpdateUserCredentialInfoByUserIdParams{
				Username: spec.Username,
				Email:    sqlEmail,
				UserID:   int64(spec.UserID),
			})
			if err != nil {
				return fmt.Errorf("qtx.UpdateUserCredentialInfoByUserId(): %w", err)
			}

			userIDs = append(userIDs, spec.UserID)
		}
		return nil
	})
	if err != nil {
		return []identity.UserID{}, fmt.Errorf("ExecuteInTransaction(): %w", err)
	}

	return userIDs, nil
}

func (s identityServiceImpl) UpdateUserInfosByUsernames(ctx context.Context, specs []identity.UpdateUserInfoByUsernameSpec) (int64, error) {
	// as a backend-only administrative tools, we can skip the `ValidateUpdateSpecs()`

	userDetails := make([][]byte, 0, len(specs))
	for _, spec := range specs {
		userDetail, err := json.Marshal(spec.UserDetail)
		if err != nil {
			return 0, fmt.Errorf("marshal UserDetail: %w", err)
		}
		userDetails = append(userDetails, userDetail)
	}

	var totalUpdatedRows int64

	err := s.mySQLQueries.ExecuteInTransaction(ctx, func(newCtx context.Context, qtx *mysql.Queries) error {
		for i, spec := range specs {
			email := strings.TrimSpace(spec.Email)
			sqlEmail := sql.NullString{String: email, Valid: email != ""}
			n1, err := qtx.UpdateUserByUsername(newCtx, mysql.UpdateUserByUsernameParams{
				Email:         sqlEmail,
				UserDetail:    userDetails[i],
				PrivilegeType: int32(spec.UserPrivilegeType),
				IsDeactivated: util.BoolToInt32(spec.IsDeactivated),
				Username:      spec.Username,
			})
			if err != nil {
				return fmt.Errorf("qtx.UpdateUserByUsername(): %w", err)
			}

			n2, err := qtx.UpdateUserCredentialInfoByUsername(newCtx, mysql.UpdateUserCredentialInfoByUsernameParams{
				Email:    sqlEmail,
				Username: spec.Username,
			})
			if err != nil {
				return fmt.Errorf("qtx.UpdateUserCredentialInfoByUsername(): %w", err)
			}

			if n1 != n2 {
				mainLog.Warn("UpdateUserByUsername()'s updated row (%d) != UpdateUserCredentialInfoByUsername()'s updated row (%d). There MUST be an inconsistency between `user` and `user_credential`.", n1, n2)
			}

			totalUpdatedRows += n1
		}
		return nil
	})
	if err != nil {
		return 0, fmt.Errorf("ExecuteInTransaction(): %w", err)
	}

	return totalUpdatedRows, nil
}

func (s identityServiceImpl) UpdateUserPassword(ctx context.Context, spec identity.UpdateUserPasswordSpec) error {
	hashedPassword, err := hashPassword(spec.Password)
	if err != nil {
		return fmt.Errorf("hashPassword(): %w", err)
	}

	err = s.mySQLQueries.UpdatePasswordByUserId(ctx, mysql.UpdatePasswordByUserIdParams{
		Password: hashedPassword,
		UserID:   int64(spec.UserID),
	})
	if err != nil {
		return fmt.Errorf("mySQLQueries.UpdatePasswordByUserId(): %v", err)
	}

	return nil
}

func (s identityServiceImpl) SignUpUser(ctx context.Context, spec identity.SignUpUserSpec) (identity.UserID, error) {
	userIds, err := s.InsertUsers(ctx, []identity.InsertUserSpec{
		{
			Email:             spec.Email,
			Password:          spec.Password,
			Username:          spec.Username,
			UserDetail:        spec.UserDetail,
			UserPrivilegeType: identity.UserPrivilegeType_Member,
		},
	})
	if err != nil {
		return identity.UserID_None, fmt.Errorf("InsertUser(): %w", err)
	}

	return userIds[0], nil
}

func (s identityServiceImpl) LoginUser(ctx context.Context, spec identity.LoginUserSpec) (identity.LoginUserResult, error) {
	userCredential, err := s.mySQLQueries.GetUserCredentialByEmail(ctx, sql.NullString{String: spec.UsernameOrEmail, Valid: true})
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return identity.LoginUserResult{}, fmt.Errorf("mySQLQueries.GetUserCredentialByEmail(): %w", err)
		}

		userCredential, err = s.mySQLQueries.GetUserCredentialByUsername(ctx, spec.UsernameOrEmail)
		if err != nil {
			return identity.LoginUserResult{}, fmt.Errorf("mySQLQueries.GetUserCredentialByUsername(): %w", err)
		}
	}

	// Compare the hashed password with the input password
	err = bcrypt.CompareHashAndPassword([]byte(userCredential.Password), []byte(spec.Password))
	if err != nil {
		return identity.LoginUserResult{}, fmt.Errorf("bcrypt.CompareHashAndPassword(): %w", err)
	}

	user, err := s.GetUserById(ctx, identity.UserID(userCredential.UserID))
	if err != nil {
		return identity.LoginUserResult{}, fmt.Errorf("GetUserById(): %w", err)
	}

	if user.IsDeactivated {
		return identity.LoginUserResult{}, fmt.Errorf("userId='%d': %w", user.UserID, errs.ErrUserDeactivated)
	}

	// Create a JWT token
	tokenString, err := s.jwtService.CreateJWTToken(
		identity.UserID(userCredential.UserID), identity.UserPrivilegeType(user.PrivilegeType),
		auth.JWTTokenPurposeType_Auth, auth.JWTToken_ExpiryTime_SetDefault,
	)
	if err != nil {
		return identity.LoginUserResult{}, fmt.Errorf("jwtService.CreateJWTToken(): %w", err)
	}

	return identity.LoginUserResult{
		User:      user,
		AuthToken: identity.AuthToken(tokenString),
	}, nil
}

func (s identityServiceImpl) ForgotPassword(ctx context.Context, spec identity.ForgotPasswordSpec) error {
	user, err := s.mySQLQueries.GetUserByEmail(ctx, sql.NullString{String: spec.Email, Valid: true})
	if err != nil {
		return fmt.Errorf("GetUserByEmail(): %w", err)
	}

	userID := identity.UserID(user.ID)
	recipientName := user.Username
	recipientEmail := user.Email.String

	if userID == identity.UserID_None {
		mainLog.Error("ForgotPassword invoked on non-existing user with email='%s", spec.Email)
		return nil // we return no error for security reason: avoid user knowing the existence of any email
	}

	tokenString, err := s.jwtService.CreateJWTToken(
		userID, identity.UserPrivilegeType_Anonymous,
		auth.JWTTokenPurposeType_ResetPassword, 2*time.Hour,
	)
	if err != nil {
		return fmt.Errorf("jwtService.CreateJWTToken(): %w", err)
	}

	mainLog.Info("%s", tokenString)

	resetPasswordTemplate := email_composer.NewPasswordReset(recipientName, fmt.Sprintf("%s%s?token=%s", configObject.Email_BaseAppURL, "/reset-password", tokenString))
	body, err := s.emailComposer.GenerateHTML(resetPasswordTemplate.Email())
	if err != nil {
		return fmt.Errorf("emailComposer.GenerateHTML(): %w", err)
	}
	err = s.smtpAccessor.SendEmail(
		true,
		"",
		[]string{recipientEmail},
		fmt.Sprintf("Reset Password Request on %s", s.emailComposer.Product.Name),
		body,
	)
	if err != nil {
		return fmt.Errorf("SendEmail(): %w", err)
	}
	return nil
}

func (s identityServiceImpl) ResetPassword(ctx context.Context, spec identity.ResetPasswordSpec) error {
	claims, err := s.jwtService.VerifyTokenStringAndReturnClaims(spec.ResetToken)
	if err != nil {
		return errs.NewHTTPError(http.StatusForbidden, fmt.Errorf("VerifyTokenStringAndReturnClaims(): %v", err), nil, "Invalid reset password token. Try requesting for a password reset again.")
	}

	mainClaims := claims.(*auth.MainJWTClaims)
	if mainClaims.PurposeType != auth.JWTTokenPurposeType_ResetPassword {
		return errs.NewHTTPError(http.StatusForbidden, fmt.Errorf("invalid JWT token purpose"), nil, "Invalid reset password token.")
	}

	err = s.UpdateUserPassword(ctx, identity.UpdateUserPasswordSpec{
		UserID:   mainClaims.UserID,
		Password: spec.NewPassword,
	})
	if err != nil {
		return fmt.Errorf("UpdateUserPassword(): %v", err)
	}

	return nil
}

func (s identityServiceImpl) VerifyUserAuthority(ctx context.Context, minimalPrivilegeType identity.UserPrivilegeType) (bool, error) {
	authInfo := network.GetAuthInfo(ctx)
	if authInfo.UserID == identity.UserID_None {
		return false, fmt.Errorf("request is unauthenticated")
	}

	if authInfo.PrivilegeType < minimalPrivilegeType {
		return false, nil
	}
	return true, nil
}

func hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("bcrypt.GenerateFromPassword: %w", err)
	}
	return string(hashedPassword), nil
}
