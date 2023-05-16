package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"database/sql"

	_ "github.com/go-sql-driver/mysql"

	emailImpl "sonamusica-backend/accessor/email/impl"
	"sonamusica-backend/accessor/relational_db"
	"sonamusica-backend/app-service/auth"
	"sonamusica-backend/app-service/email_composer"
	"sonamusica-backend/app-service/identity"
	identityImpl "sonamusica-backend/app-service/identity/impl"
	"sonamusica-backend/app-service/teaching"
	teachingImpl "sonamusica-backend/app-service/teaching/impl"
	"sonamusica-backend/app-service/util"
	"sonamusica-backend/config"
	"sonamusica-backend/errs"
	"sonamusica-backend/logging"
	"sonamusica-backend/network"
	"sonamusica-backend/service/output"
)

var (
	configObject = config.Get()
	mainLog      = logging.NewGoLogger("BackendService", logging.GetLevel(configObject.LogLevel))
)

type BackendService struct {
	jwtService auth.JWTService

	identityService identity.IdentityService
	teachingService teaching.TeachingService
}

func NewBackendService() *BackendService {
	db, err := sql.Open("mysql", configObject.GetMySQLURI())
	if err != nil {
		mainLog.Error("Failed to connect to database:", err)
	}
	mySqlQueries := relational_db.NewMySQLQueries(db)

	smtpAccessor := emailImpl.NewSMTPAccessor(
		configObject.SMTPHost,
		configObject.SMTPPort,
		configObject.SMTPLogin,
		configObject.SMTPPassword,
		configObject.Email_Sender,
	)

	jwtService := auth.NewJWTServiceImpl(auth.JWTServiceConfig{
		SecretKey:       configObject.JWTSecretKey,
		TokenExpiration: configObject.JWTTokenExpiration,
	})

	emailComposer := email_composer.NewComposer()

	identityService := identityImpl.NewIdentityServiceImpl(mySqlQueries, smtpAccessor, jwtService, emailComposer)

	teachingService := teachingImpl.NewTeachingServiceImpl(mySqlQueries)

	return &BackendService{
		jwtService:      jwtService,
		identityService: identityService,
		teachingService: teachingService,
	}
}

func (s *BackendService) HomepageHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello, world!")
}

func (s *BackendService) UserDataHandler(ctx context.Context, req *output.UserDataRequest) (*output.UserDataResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	user, err := s.identityService.GetUserById(ctx, identity.UserID(req.ID))
	if err != nil {
		mainLog.Error("User with ID='%d' is not found\n", req.ID)
		return nil, errs.NewHTTPError(http.StatusNotFound, fmt.Errorf("identityService.GetUserById(): %w", err), map[string]string{})
	}

	return &output.UserDataResponse{Data: user}, nil
}

func (s *BackendService) GetTeachersHandler(ctx context.Context, req *output.GetTeachersRequest) (*output.GetTeachersResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	getTeachersResult, err := s.teachingService.GetTeachers(ctx, util.PaginationSpec{
		Page:           req.Page,
		ResultsPerPage: req.ResultsPerPage,
	})
	if err != nil {
		return nil, errs.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("teachingService.GetTeachers(): %w", err), map[string]string{errs.ClientMessageKey_NonField: "Failed to get teachers"})
	}

	paginationResult := getTeachersResult.PaginationResult

	return &output.GetTeachersResponse{
		Data: output.GetTeachersResult{
			Results: getTeachersResult.Teachers,
			PaginationResponse: output.PaginationResponse{
				TotalPages:   paginationResult.TotalPages,
				TotalResults: paginationResult.TotalResults,
				CurrentPage:  paginationResult.CurrentPage,
			},
		},
	}, nil
}

func (s *BackendService) GetStudentsHandler(ctx context.Context, req *output.GetStudentsRequest) (*output.GetStudentsResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	getStudentsResult, err := s.teachingService.GetStudents(ctx, util.PaginationSpec{
		Page:           req.Page,
		ResultsPerPage: req.ResultsPerPage,
	})
	if err != nil {
		return nil, errs.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("teachingService.GetStudents(): %w", err), map[string]string{errs.ClientMessageKey_NonField: "Failed to get students"})
	}

	paginationResult := getStudentsResult.PaginationResult

	return &output.GetStudentsResponse{
		Data: output.GetStudentsResult{
			Results: getStudentsResult.Students,
			PaginationResponse: output.PaginationResponse{
				TotalPages:   paginationResult.TotalPages,
				TotalResults: paginationResult.TotalResults,
				CurrentPage:  paginationResult.CurrentPage,
			},
		},
	}, nil
}

func (s *BackendService) GetJWTHandler(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		mainLog.Error("Failed to parse userID:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	tokenString, err := s.jwtService.CreateJWTToken(identity.UserID(userID), auth.JWTTokenPurposeType_Auth, 0)
	if err != nil {
		mainLog.Error("Failed to create JWT token:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, tokenString)
}

func (s *BackendService) SignUpHandler(ctx context.Context, req *output.SignUpRequest) (*output.SignUpResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	userID, err := s.identityService.SignUpUser(ctx, identity.SignUpUserSpec{
		Email:    req.Email,
		Password: req.Password,
		Username: req.Username,
	})
	if err != nil {
		errContext := fmt.Sprintf("identityService.SignUpUser()")
		var validationErr errs.ValidationError
		if errors.As(err, &validationErr) {
			return nil, errs.NewHTTPError(http.StatusConflict, fmt.Errorf("%s: %v", errContext, validationErr), validationErr.GetErrorDetail())
		}
		return nil, errs.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("%s: %v", errContext, err), map[string]string{errs.ClientMessageKey_NonField: "Failed to create user"})
	}
	mainLog.Info("User created: userID='%d', email='%s', username='%s", userID, req.Email, req.Username)

	return &output.SignUpResponse{
		Message: "User created successfully",
	}, nil
}

func (s *BackendService) LoginHandler(ctx context.Context, req *output.LoginRequest) (*output.LoginResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	requestContext, errHTTP := network.GetRequestContext(ctx)
	if errHTTP != nil {
		return nil, errs.NewHTTPError(errHTTP.GetHTTPErrorCode(), fmt.Errorf("network.GetRequestContext(): %w", errHTTP), errHTTP.GetClientMessages())
	}
	mainLog.Debug("context: %#v", ctx)
	mainLog.Debug("requestContext: %#v", requestContext)
	mainLog.Debug("request: %#v", req)

	loginResult, err := s.identityService.LoginUser(ctx, identity.LoginUserSpec{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		return nil, errs.NewHTTPError(http.StatusUnauthorized, fmt.Errorf("identityService.LoginUser(): %w", err), map[string]string{errs.ClientMessageKey_NonField: "Authentication failed"})
	}

	return &output.LoginResponse{
		Data: output.LoginResult{
			User:      loginResult.User,
			AuthToken: loginResult.AuthToken,
		},
		Message: "Login successful!",
	}, nil
}

func (s *BackendService) ForgotPasswordHandler(ctx context.Context, req *output.ForgotPasswordRequest) (*output.ForgotPasswordResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	err := s.identityService.ForgotPassword(ctx, identity.ForgotPasswordSpec{
		Email: req.Email,
	})
	if err != nil {
		return nil, errs.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("identityService.ForgotPassword(): %w", err), map[string]string{})
	}

	return &output.ForgotPasswordResponse{
		Message: "A reset password link has been sent to your email.",
	}, nil
}

func (s *BackendService) ResetPasswordHandler(ctx context.Context, req *output.ResetPasswordRequest) (*output.ResetPasswordResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	err := s.identityService.ResetPassword(ctx, identity.ResetPasswordSpec{
		ResetToken:  req.ResetToken,
		NewPassword: req.NewPassword,
	})
	if err != nil {
		var httpErr errs.HTTPError
		if errors.As(err, &httpErr) {
			return nil, httpErr
		}
		return nil, errs.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("identityService.ResetPassword(): %w", err), map[string]string{})
	}

	return &output.ResetPasswordResponse{
		Message: "Password reset successfully.",
	}, nil
}
