package impl

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
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

func (s identityServiceImpl) GetUsers(ctx context.Context, pagination util.PaginationSpec) (identity.GetUsersResult, error) {
	pagination.SetDefaultOnInvalidValues()
	limit, offset := pagination.GetLimitAndOffset()
	userRows, err := s.mySQLQueries.GetUsers(ctx, mysql.GetUsersParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return identity.GetUsersResult{}, fmt.Errorf("mySQLQueries.GetUsers(): %w", err)
	}

	users := make([]identity.User, 0, len(userRows))
	for _, userRow := range userRows {
		var userDetail identity.UserDetail
		err = json.Unmarshal(userRow.UserDetail, &userDetail)
		if err != nil {
			return identity.GetUsersResult{}, fmt.Errorf("json.Unmarshal(): %w", err)
		}

		users = append(users, identity.User{
			ID:            identity.UserID(userRow.ID),
			Username:      userRow.Username,
			Email:         userRow.Email,
			UserDetail:    userDetail,
			PrivilegeType: identity.UserPrivilegeType(userRow.PrivilegeType),
			CreatedAt:     userRow.CreatedAt.Time,
		})
	}

	totalResults, err := s.mySQLQueries.CountUsers(ctx)

	return identity.GetUsersResult{
		Users:            users,
		PaginationResult: *util.NewPaginationResult(int(totalResults), pagination.ResultsPerPage, pagination.Page),
	}, nil
}

func (s identityServiceImpl) GetUserById(ctx context.Context, id identity.UserID) (identity.User, error) {
	user, err := s.mySQLQueries.GetUserById(ctx, int64(id))
	if err != nil {
		return identity.User{}, fmt.Errorf("mySQLQueries.GetUserById(): %w", err)
	}

	var userDetail identity.UserDetail
	err = json.Unmarshal(user.UserDetail, &userDetail)
	if err != nil {
		return identity.User{}, fmt.Errorf("json.Unmarshal(): %w", err)
	}

	return identity.User{
		ID:            identity.UserID(user.ID),
		Username:      user.Username,
		Email:         user.Email,
		UserDetail:    userDetail,
		PrivilegeType: identity.UserPrivilegeType(user.PrivilegeType),
		CreatedAt:     user.CreatedAt.Time,
	}, nil
}

// InsertUsers insert users specified in spec in bulk, in single transaction (it's ALL successful or NONE successful).
//
// This methods check for existing *sql.Tx inside ctx, and will reuse the tx to execute the insertion.
// Else, will create a new *sql.Tx and will be committed immediately.
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
	var tx *sql.Tx
	var err error
	isReusingExistingTx := false
	if existingTx := network.GetSQLTx(ctx); existingTx != nil { // reuse existing pre-created SQL transaction (Tx) if exists
		tx = existingTx
		isReusingExistingTx = true
	} else {
		tx, err = s.mySQLQueries.DB.Begin()
		if err != nil {
			return []identity.UserID{}, fmt.Errorf("mySQLDB.Begin(): %w", err)
		}
		defer tx.Rollback()
	}

	qtx := s.mySQLQueries.WithTx(tx)

	for i, spec := range specs {
		userID, err := qtx.InsertUser(ctx, mysql.InsertUserParams{
			Email:         spec.Email,
			Username:      spec.Username,
			UserDetail:    userDetails[i],
			PrivilegeType: int32(spec.UserPrivilegeType),
		})
		wrappedErr := errs.WrapMySQLError(err)
		if wrappedErr != nil {
			return []identity.UserID{}, fmt.Errorf("qtx.InsertUser(): %w", wrappedErr)
		}

		_, err = qtx.InsertUserCredential(ctx, mysql.InsertUserCredentialParams{
			UserID:   userID,
			Username: spec.Username,
			Password: hashedPasswords[i],
			Email:    spec.Email,
		})
		wrappedErr = errs.WrapMySQLError(err)
		if wrappedErr != nil {
			return []identity.UserID{}, fmt.Errorf("qtx.InsertUserCredentia(): %w", wrappedErr)
		}

		userIds = append(userIds, identity.UserID(userID))
	}

	if !isReusingExistingTx {
		err = tx.Commit()
		if err != nil {
			return []identity.UserID{}, fmt.Errorf("tx.Commit(): %w", err)
		}
	}

	return userIds, nil
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
	userCredential, err := s.mySQLQueries.GetUserCredentialByEmail(ctx, spec.UsernameOrEmail)
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
	user, err := s.mySQLQueries.GetUserByEmail(ctx, spec.Email)
	if err != nil {
		return fmt.Errorf("GetUserByEmail(): %w", err)
	}

	userID := identity.UserID(user.ID)
	recipientName := user.Username
	recipientEmail := user.Email

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

	resetPasswordTemplate := email_composer.NewPasswordReset(recipientName, fmt.Sprintf("%s?token=%s", "http://localhost:8080/password_reset", tokenString))
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
		return errs.NewHTTPError(http.StatusForbidden, fmt.Errorf("invalid JWT token purpose"), nil, "Invalid reset password token")
	}

	userID := mainClaims.UserID
	hashedPassword, err := hashPassword(spec.NewPassword)
	if err != nil {
		return fmt.Errorf("hashPassword(): %w", err)
	}

	err = s.mySQLQueries.UpdatePasswordByUserId(ctx, mysql.UpdatePasswordByUserIdParams{
		Password: hashedPassword,
		UserID:   int64(userID),
	})
	if err != nil {
		return fmt.Errorf("UpdatePasswordByUserId(): %v", err)
	}

	return nil
}

func hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("bcrypt.GenerateFromPassword: %w", err)
	}
	return string(hashedPassword), nil
}
