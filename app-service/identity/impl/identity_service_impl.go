package impl

import (
	"context"
	"encoding/json"
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
	"sonamusica-backend/config"
	"sonamusica-backend/errs"
	"sonamusica-backend/logging"
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

func (s identityServiceImpl) SignUpUser(ctx context.Context, spec identity.SignUpUserSpec) (identity.UserID, error) {
	hashedPassword, err := hashPassword(spec.Password)
	if err != nil {
		return identity.UserID_None, fmt.Errorf("hashPassword(): %w", err)
	}

	userDetail, err := json.Marshal(spec.UserDetail)
	if err != nil {
		return identity.UserID_None, fmt.Errorf("marshal UserDetail: %w", err)
	}

	// We insert into 2 tables, need to wrap inside SQL transaction
	tx, err := s.mySQLQueries.DB.Begin()
	if err != nil {
		return identity.UserID_None, fmt.Errorf("mySQLDB.Begin(): %w", err)
	}
	defer tx.Rollback()
	qtx := s.mySQLQueries.WithTx(tx)

	userID, err := qtx.InsertUser(ctx, mysql.InsertUserParams{
		Email:         spec.Email,
		Username:      spec.Username,
		UserDetail:    userDetail,
		PrivilegeType: int32(identity.UserPrivilegeType_Member),
	})
	dbErr := errs.WrapMySQLError(err)
	if dbErr != nil {
		return identity.UserID_None, fmt.Errorf("qtx.InsertUser(): %w", dbErr)
	}

	_, err = qtx.InsertUserCredential(ctx, mysql.InsertUserCredentialParams{
		UserID:   userID,
		Password: hashedPassword,
		Email:    spec.Email,
	})
	dbErr = errs.WrapMySQLError(err)
	if dbErr != nil {
		return identity.UserID_None, fmt.Errorf("qtx.InsertUserCredentia(): %w", dbErr)
	}

	err = tx.Commit()
	if err != nil {
		return identity.UserID_None, fmt.Errorf("tx.Commit(): %w", err)
	}

	return identity.UserID(userID), nil
}

func (s identityServiceImpl) LoginUser(ctx context.Context, spec identity.LoginUserSpec) (identity.LoginUserResult, error) {
	userCredential, err := s.mySQLQueries.GetUserCredentialByEmail(ctx, spec.Email)
	if err != nil {
		return identity.LoginUserResult{}, fmt.Errorf("mySQLQueries.GetUserCredentialByEmail(): %w", err)
	}

	// Compare the hashed password with the input password
	err = bcrypt.CompareHashAndPassword([]byte(userCredential.Password), []byte(spec.Password))
	if err != nil {
		return identity.LoginUserResult{}, fmt.Errorf("bcrypt.CompareHashAndPassword(): %w", err)
	}

	// Create a JWT token
	tokenString, err := s.jwtService.CreateJWTToken(identity.UserID(userCredential.UserID), auth.JWTTokenPurposeType_Auth, 0)
	if err != nil {
		return identity.LoginUserResult{}, fmt.Errorf("jwtService.CreateJWTToken(): %w", err)
	}

	user, err := s.GetUserById(ctx, identity.UserID(userCredential.UserID))
	if err != nil {
		return identity.LoginUserResult{}, fmt.Errorf("GetUserById(): %w", err)
	}

	return identity.LoginUserResult{
		User:      user,
		AuthToken: identity.AuthToken(tokenString),
	}, nil
}

func (s identityServiceImpl) ForgotPassword(ctx context.Context, spec identity.ForgotPasswordSpec) error {
	user, err := s.mySQLQueries.GetUserByEmail(ctx, spec.Email)
	dbErr := errs.WrapMySQLError(err)
	if dbErr != nil {
		return fmt.Errorf("GetUserByEmail(): %w", dbErr)
	}

	userID := identity.UserID(user.ID)
	recipientName := user.Username
	recipientEmail := user.Email

	if userID == identity.UserID_None {
		mainLog.Error("ForgotPassword invoked on non-existing user with email='%s", spec.Email)
		return nil // we return no error for security reason: avoid user knowing the existence of any email
	}

	tokenString, err := s.jwtService.CreateJWTToken(userID, auth.JWTTokenPurposeType_ResetPassword, 2*time.Hour)
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
		return errs.NewHTTPError(http.StatusForbidden, fmt.Errorf("VerifyTokenStringAndReturnClaims(): %v", err), map[string]string{errs.ClientMessageKey_NonField: "Invalid reset password token. Try requesting for a password reset again."})
	}

	mainClaims := claims.(*auth.MainJWTClaims)
	if mainClaims.PurposeType != auth.JWTTokenPurposeType_ResetPassword {
		return errs.NewHTTPError(http.StatusForbidden, fmt.Errorf("invalid JWT token purpose"), map[string]string{errs.ClientMessageKey_NonField: "Invalid reset password token"})
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
