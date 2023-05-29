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

	teachingService := teachingImpl.NewTeachingServiceImpl(mySqlQueries, identityService)

	return &BackendService{
		jwtService:      jwtService,
		identityService: identityService,
		teachingService: teachingService,
	}
}

func (s *BackendService) HomepageHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello, world!")
}

func (s *BackendService) GetJWTHandler(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		mainLog.Error("Failed to parse userID:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	userPrivilegeTypeStr := r.URL.Query().Get("privilegeType")
	privilegeType, err := strconv.Atoi(userPrivilegeTypeStr)
	if err != nil {
		mainLog.Warn("Failed to parse privilegeType: %v. Defaulting to Anonymous", err)
		privilegeType = int(identity.UserPrivilegeType_Anonymous)
	}

	tokenPurposeTypeStr := r.URL.Query().Get("tokenPurposeType")
	tokenPurposeType, err := strconv.Atoi(tokenPurposeTypeStr)
	if err != nil {
		mainLog.Warn("Failed to parse tokenPurposeType: %v. Defaulting to Auth", err)
		tokenPurposeType = int(auth.JWTTokenPurposeType_Auth)
	}

	tokenString, err := s.jwtService.CreateJWTToken(
		identity.UserID(userID), identity.UserPrivilegeType(privilegeType),
		auth.JWTTokenPurposeType(tokenPurposeType), auth.JWTToken_ExpiryTime_SetDefault,
	)
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
		Email:      req.Email,
		Password:   req.Password,
		Username:   req.Username,
		UserDetail: req.UserDetail,
	})
	if err != nil {
		errContext := fmt.Sprintf("identityService.SignUpUser()")
		var validationErr errs.ValidationError
		if errors.As(err, &validationErr) {
			return nil, errs.NewHTTPError(http.StatusConflict, fmt.Errorf("%s: %v", errContext, validationErr), validationErr.GetErrorDetail(), "")
		}
		return nil, errs.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("%s: %v", errContext, err), nil, "Failed to create user")
	}
	mainLog.Info("User created: userID='%d', email='%s', username='%s'", userID, req.Email, req.Username)

	return &output.SignUpResponse{
		Message: "User created successfully",
	}, nil
}

func (s *BackendService) LoginHandler(ctx context.Context, req *output.LoginRequest) (*output.LoginResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	requestContext := network.GetRequestContext(ctx)
	mainLog.Debug("context: %#v", ctx)
	mainLog.Debug("requestContext: %#v", requestContext)
	mainLog.Debug("request: %#v", req)

	loginResult, err := s.identityService.LoginUser(ctx, identity.LoginUserSpec{
		UsernameOrEmail: req.UsernameOrEmail,
		Password:        req.Password,
	})
	if err != nil {
		return nil, errs.NewHTTPError(http.StatusUnauthorized, fmt.Errorf("identityService.LoginUser(): %w", err), nil, "Authentication failed")
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
		return nil, errs.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("identityService.ForgotPassword(): %w", err), nil, "Failed to send forgot password request")
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
		return nil, errs.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("identityService.ResetPassword(): %w", err), nil, "Failed to reset password")
	}

	return &output.ResetPasswordResponse{
		Message: "Password reset successfully.",
	}, nil
}

func (s *BackendService) GetUsersHandler(ctx context.Context, req *output.GetUsersRequest) (*output.GetUsersResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	getUsersResult, err := s.identityService.GetUsers(ctx, util.PaginationSpec{
		Page:           req.Page,
		ResultsPerPage: req.ResultsPerPage,
	})
	if err != nil {
		return nil, errs.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("identityService.GetUsers(): %w", err), nil, "Failed to get users")
	}

	paginationResult := getUsersResult.PaginationResult

	return &output.GetUsersResponse{
		Data: output.GetUsersResult{
			Results: getUsersResult.Users,
			PaginationResponse: output.PaginationResponse{
				TotalPages:   paginationResult.TotalPages,
				TotalResults: paginationResult.TotalResults,
				CurrentPage:  paginationResult.CurrentPage,
			},
		},
	}, nil
}

func (s *BackendService) GetUserByIdHandler(ctx context.Context, req *output.GetUserRequest) (*output.GetUserResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	user, err := s.identityService.GetUserById(ctx, req.ID)
	if err != nil {
		wrappedErr := fmt.Errorf("identityService.GetUserById(): %w", err)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.NewHTTPError(http.StatusNotFound, wrappedErr, nil, "User is not found")
		}
		return nil, errs.NewHTTPError(http.StatusInternalServerError, wrappedErr, nil, "Failed to get user")
	}

	return &output.GetUserResponse{
		Data: user,
	}, nil
}

func (s *BackendService) InsertUsersHandler(ctx context.Context, req *output.InsertUsersRequest) (*output.InsertUsersResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	specs := make([]identity.InsertUserSpec, 0, len(req.Params))
	for _, param := range req.Params {
		specs = append(specs, identity.InsertUserSpec{
			Email:             param.Email,
			Password:          param.Password,
			Username:          param.Username,
			UserDetail:        param.UserDetail,
			UserPrivilegeType: param.UserPrivilegeType,
		})
	}

	userIDs, err := s.identityService.InsertUsers(ctx, specs)
	if err != nil {
		errContext := fmt.Sprintf("identityService.InsertUsers()")
		var validationErr errs.ValidationError
		if errors.As(err, &validationErr) {
			return nil, errs.NewHTTPError(http.StatusConflict, fmt.Errorf("%s: %v", errContext, validationErr), validationErr.GetErrorDetail(), "Invalid user properties")
		}
		return nil, errs.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("%s: %v", errContext, err), nil, "Failed to create user")
	}
	mainLog.Info("Users created: userIDs='%v'", userIDs)

	return &output.InsertUsersResponse{
		Data: userIDs,
	}, nil
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
		return nil, errs.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("teachingService.GetTeachers(): %w", err), nil, "Failed to get teachers")
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

func (s *BackendService) GetTeacherByIdHandler(ctx context.Context, req *output.GetTeacherRequest) (*output.GetTeacherResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	teacher, err := s.teachingService.GetTeacherById(ctx, req.ID)
	if err != nil {
		wrappedErr := fmt.Errorf("identityService.GetTeacherById(): %w", err)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.NewHTTPError(http.StatusNotFound, wrappedErr, nil, "Teacher is not found")
		}
		return nil, errs.NewHTTPError(http.StatusInternalServerError, wrappedErr, nil, "Failed to get teacher")
	}

	return &output.GetTeacherResponse{
		Data: teacher,
	}, nil
}

func (s *BackendService) InsertTeachersHandler(ctx context.Context, req *output.InsertTeachersRequest) (*output.InsertTeachersResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	teacherIDs, err := s.teachingService.InsertTeachers(ctx, req.UserIDs)
	if err != nil {
		errContext := fmt.Sprintf("teachingService.InsertTeachers()")
		var validationErr errs.ValidationError
		if errors.As(err, &validationErr) {
			return nil, errs.NewHTTPError(http.StatusConflict, fmt.Errorf("%s: %v", errContext, validationErr), validationErr.GetErrorDetail(), "Invalid teacher properties")
		}
		return nil, errs.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("%s: %v", errContext, err), nil, "Failed to create teachers")
	}
	mainLog.Info("Teachers created: teacherIDs='%v'", teacherIDs)

	return &output.InsertTeachersResponse{
		Data: teacherIDs,
	}, nil
}

func (s *BackendService) InsertTeachersWithNewUsersHandler(ctx context.Context, req *output.InsertTeachersWithNewUsersRequest) (*output.InsertTeachersWithNewUsersResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	insertUserSpecs := make([]identity.InsertUserSpec, 0, len(req.Data))
	for _, param := range req.Data {
		insertUserSpecs = append(insertUserSpecs, identity.InsertUserSpec{
			Email:             param.Email,
			Password:          param.Password,
			Username:          param.Username,
			UserDetail:        param.UserDetail,
			UserPrivilegeType: param.UserPrivilegeType,
		})
	}

	teacherIDs, err := s.teachingService.InsertTeachersWithNewUsers(ctx, insertUserSpecs)
	if err != nil {
		errContext := fmt.Sprintf("teachingService.InsertTeachersWithNewUsers()")
		var validationErr errs.ValidationError
		if errors.As(err, &validationErr) {
			return nil, errs.NewHTTPError(http.StatusConflict, fmt.Errorf("%s: %v", errContext, validationErr), validationErr.GetErrorDetail(), "Invalid teacher properties")
		}
		return nil, errs.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("%s: %v", errContext, err), nil, "Failed to create teachers")
	}
	mainLog.Info("Teachers created: teacherIDs='%v'", teacherIDs)

	return &output.InsertTeachersWithNewUsersResponse{
		InsertTeachersResponse: output.InsertTeachersResponse{
			Data: teacherIDs,
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
		return nil, errs.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("teachingService.GetStudents(): %w", err), nil, "Failed to get students")
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

func (s *BackendService) GetStudentByIdHandler(ctx context.Context, req *output.GetStudentRequest) (*output.GetStudentResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	student, err := s.teachingService.GetStudentById(ctx, req.ID)
	if err != nil {
		wrappedErr := fmt.Errorf("identityService.GetStudentById(): %w", err)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.NewHTTPError(http.StatusNotFound, wrappedErr, nil, "Student is not found")
		}
		return nil, errs.NewHTTPError(http.StatusInternalServerError, wrappedErr, nil, "Failed to get student")
	}

	return &output.GetStudentResponse{
		Data: student,
	}, nil
}
