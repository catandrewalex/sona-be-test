package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"database/sql"

	_ "github.com/go-sql-driver/mysql"

	emailImpl "sonamusica-backend/accessor/email/impl"
	"sonamusica-backend/accessor/relational_db"
	"sonamusica-backend/app-service/auth"
	"sonamusica-backend/app-service/dashboard"
	dashboardImpl "sonamusica-backend/app-service/dashboard/impl"
	"sonamusica-backend/app-service/email_composer"
	"sonamusica-backend/app-service/entity"
	entityImpl "sonamusica-backend/app-service/entity/impl"
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

	identityService  identity.IdentityService
	entityService    entity.EntityService
	teachingService  teaching.TeachingService
	dashboardService dashboard.DashboardService
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

	entityService := entityImpl.NewEntityServiceImpl(mySqlQueries, identityService)

	teachingService := teachingImpl.NewTeachingServiceImpl(mySqlQueries, entityService)

	dashhboardService := dashboardImpl.NewDashboardServiceImpl(mySqlQueries, entityService)

	return &BackendService{
		jwtService:       jwtService,
		identityService:  identityService,
		entityService:    entityService,
		teachingService:  teachingService,
		dashboardService: dashhboardService,
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
		errContext := "identityService.SignUpUser()"
		var validationErr errs.ValidationError
		if errors.As(err, &validationErr) {
			return nil, errs.NewHTTPError(http.StatusConflict, fmt.Errorf("%s: %v", errContext, err), validationErr.GetErrorDetail(), "")
		}
		return nil, errs.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("%s: %v", errContext, err), nil, "Failed to create user")
	}
	mainLog.Info("User created: userID='%d', email='%s', username='%s'", userID, req.Email, req.Username)

	return &output.SignUpResponse{
		Message: "User registered successfully",
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
		errContext := fmt.Errorf("identityService.LoginUser(): %w", err)
		if errors.Is(err, errs.ErrUserDeactivated) {
			return nil, errs.NewHTTPError(http.StatusForbidden, errContext, nil, "This account is deactivated. Please contact system administrator fur further action")
		}
		return nil, errs.NewHTTPError(http.StatusUnauthorized, errContext, nil, "Authentication failed")
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
		if errors.Is(err, sql.ErrNoRows) {
			mainLog.Warn("A ForgotPassword requests provided a non-existing email='%s'", req.Email)
		} else {
			return nil, errs.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("identityService.ForgotPassword(): %w", err), nil, "Failed to send forgot password request")
		}
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
		Message: "Successfully reset password",
	}, nil
}

func (s *BackendService) GetUsersHandler(ctx context.Context, req *output.GetUsersRequest) (*output.GetUsersResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	getUsersResult, err := s.identityService.GetUsers(ctx, util.PaginationSpec(req.PaginationRequest), req.Filter, req.IncludeDeactivated)
	if err != nil {
		return nil, errs.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("identityService.GetUsers(): %w", err), nil, "Failed to get users")
	}

	paginationResponse := output.NewPaginationResponse(getUsersResult.PaginationResult)

	return &output.GetUsersResponse{
		Data: output.GetUsersResult{
			Results:            getUsersResult.Users,
			PaginationResponse: paginationResponse,
		},
	}, nil
}

func (s *BackendService) GetUserByIdHandler(ctx context.Context, req *output.GetUserRequest) (*output.GetUserResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	user, err := s.identityService.GetUserById(ctx, req.UserID)
	if err != nil {
		return nil, handleReadError(err, "identityService.GetUserById()", "user")
	}

	return &output.GetUserResponse{
		Data: user,
	}, nil
}

func (s *BackendService) InsertUsersHandler(ctx context.Context, req *output.InsertUsersRequest) (*output.InsertUsersResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	specs := make([]identity.InsertUserSpec, 0, len(req.Data))
	for _, param := range req.Data {
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
		return nil, handleUpsertionError(err, "identityService.InsertUsers()", "user")
	}
	mainLog.Info("Users created: userIDs='%v'", userIDs)

	users, err := s.identityService.GetUsersByIds(ctx, userIDs)
	if err != nil {
		return nil, errs.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("identityService.GetUsersByIds: %v", err), nil, "")
	}

	return &output.InsertUsersResponse{
		Data: output.InsertUserResult{
			Results: users,
		},
		Message: "Successfully created users",
	}, nil
}

func (s *BackendService) UpdateUsersHandler(ctx context.Context, req *output.UpdateUsersRequest) (*output.UpdateUsersResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	specs := make([]identity.UpdateUserInfoSpec, 0, len(req.Data))
	for _, param := range req.Data {
		specs = append(specs, identity.UpdateUserInfoSpec{
			UserID:            param.UserID,
			Username:          param.Username,
			Email:             param.Email,
			UserDetail:        param.UserDetail,
			UserPrivilegeType: param.UserPrivilegeType,
			IsDeactivated:     param.IsDeactivated,
		})
	}

	userIDs, err := s.identityService.UpdateUserInfos(ctx, specs)
	if err != nil {
		return nil, handleUpsertionError(err, "identityService.UpdateUserInfos()", "user")
	}
	mainLog.Info("Users updated: userIDs='%v'", userIDs)

	users, err := s.identityService.GetUsersByIds(ctx, userIDs)
	if err != nil {
		return nil, errs.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("identityService.GetUsersByIds: %v", err), nil, "")
	}

	return &output.UpdateUsersResponse{
		Data: output.UpdateUserResult{
			Results: users,
		},
		Message: "Successfully updated users",
	}, nil
}

func (s *BackendService) UpdateUsersByUsernamesHandler(ctx context.Context, req *output.UpdateUsersByUsernamesRequest) (*output.UpdateUsersByUsernamesResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	specs := make([]identity.UpdateUserInfoByUsernameSpec, 0, len(req.Data))
	for _, param := range req.Data {
		specs = append(specs, identity.UpdateUserInfoByUsernameSpec{
			Username:          param.Username,
			Email:             param.Email,
			UserDetail:        param.UserDetail,
			UserPrivilegeType: param.UserPrivilegeType,
			IsDeactivated:     param.IsDeactivated,
		})
	}

	totalUpdatedRows, err := s.identityService.UpdateUserInfosByUsernames(ctx, specs)
	if err != nil {
		return nil, handleUpsertionError(err, "identityService.UpdateUserInfosByUsernames()", "user")
	}
	mainLog.Info("Users updated: totalUpdatedRows='%v'", totalUpdatedRows)

	return &output.UpdateUsersByUsernamesResponse{
		Message: fmt.Sprintf("Successfully updated users (%d affected users)", totalUpdatedRows),
	}, nil
}

func (s *BackendService) UpdateUserPasswordHandler(ctx context.Context, req *output.UpdateUserPasswordRequest) (*output.UpdateUserPasswordResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	authInfo := network.GetAuthInfo(ctx)
	if authInfo.PrivilegeType < identity.UserPrivilegeType_Admin {
		if req.UserID != identity.UserID_None && req.UserID != authInfo.UserID {
			errContext := fmt.Errorf("unauthorized UpdateUserPassword(): userId='%d', requesteduserId='%d'", authInfo.UserID, req.UserID)
			return nil, errs.NewHTTPError(http.StatusForbidden, errContext, nil, "You're not authorized to change this user's password. Please contact the system administrator for further information.")
		}
	}

	err := s.identityService.UpdateUserPassword(ctx, identity.UpdateUserPasswordSpec{
		UserID:   req.UserID,
		Password: req.NewPassword,
	})
	if err != nil {
		return nil, errs.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("identityService.UpdateUserPassword(): %w", err), nil, "Failed to reset password")
	}

	return &output.UpdateUserPasswordResponse{
		Message: "Successfully updated password",
	}, nil
}

func (s *BackendService) GetTeachersHandler(ctx context.Context, req *output.GetTeachersRequest) (*output.GetTeachersResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	getTeachersResult, err := s.entityService.GetTeachers(ctx, util.PaginationSpec(req.PaginationRequest))
	if err != nil {
		return nil, errs.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("entityService.GetTeachers(): %w", err), nil, "Failed to get teachers")
	}

	paginationResponse := output.NewPaginationResponse(getTeachersResult.PaginationResult)

	return &output.GetTeachersResponse{
		Data: output.GetTeachersResult{
			Results:            getTeachersResult.Teachers,
			PaginationResponse: paginationResponse,
		},
	}, nil
}

func (s *BackendService) GetTeacherByIdHandler(ctx context.Context, req *output.GetTeacherRequest) (*output.GetTeacherResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	teacher, err := s.entityService.GetTeacherById(ctx, req.TeacherID)
	if err != nil {
		return nil, handleReadError(err, "identityService.GetTeacherById()", "teacher")
	}

	return &output.GetTeacherResponse{
		Data: teacher,
	}, nil
}

func (s *BackendService) InsertTeachersHandler(ctx context.Context, req *output.InsertTeachersRequest) (*output.InsertTeachersResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	userIDs := make([]identity.UserID, 0, len(req.Data))
	for _, param := range req.Data {
		userIDs = append(userIDs, param.UserID)
	}

	teacherIDs, err := s.entityService.InsertTeachers(ctx, userIDs)
	if err != nil {
		return nil, handleUpsertionError(err, "entityService.InsertTeachers()", "teacher")
	}
	mainLog.Info("Teachers created: teacherIDs='%v'", teacherIDs)

	teachers, err := s.entityService.GetTeachersByIds(ctx, teacherIDs)
	if err != nil {
		return nil, errs.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("entityService.GetTeachersByIds: %v", err), nil, "")
	}

	return &output.InsertTeachersResponse{
		Data: output.InsertTeacherResult{
			Results: teachers,
		},
		Message: "Successfully created teachers",
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

	teacherIDs, err := s.entityService.InsertTeachersWithNewUsers(ctx, insertUserSpecs)
	if err != nil {
		return nil, handleUpsertionError(err, "entityService.InsertTeachersWithNewUsers()", "teacher")
	}
	mainLog.Info("Teachers created: teacherIDs='%v'", teacherIDs)

	teachers, err := s.entityService.GetTeachersByIds(ctx, teacherIDs)
	if err != nil {
		return nil, errs.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("entityService.GetTeachersByIds: %v", err), nil, "")
	}

	return &output.InsertTeachersWithNewUsersResponse{
		InsertTeachersResponse: output.InsertTeachersResponse{
			Data: output.InsertTeacherResult{
				Results: teachers,
			},
			Message: "Successfully created teachers",
		},
	}, nil
}

func (s *BackendService) DeleteTeachersHandler(ctx context.Context, req *output.DeleteTeachersRequest) (*output.DeleteTeachersResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	teacherIDs := make([]entity.TeacherID, 0, len(req.Data))
	for _, param := range req.Data {
		teacherIDs = append(teacherIDs, param.TeacherID)
	}

	err := s.entityService.DeleteTeachers(ctx, teacherIDs)
	if err != nil {
		return nil, handleDeletionError(err, "identityService.DeleteTeachers()", "teacher")
	}

	return &output.DeleteTeachersResponse{
		Message: "Successfully deleted teachers",
	}, nil
}

func (s *BackendService) GetStudentsHandler(ctx context.Context, req *output.GetStudentsRequest) (*output.GetStudentsResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	getStudentsResult, err := s.entityService.GetStudents(ctx, util.PaginationSpec(req.PaginationRequest))
	if err != nil {
		return nil, errs.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("entityService.GetStudents(): %w", err), nil, "Failed to get students")
	}

	paginationResponse := output.NewPaginationResponse(getStudentsResult.PaginationResult)

	return &output.GetStudentsResponse{
		Data: output.GetStudentsResult{
			Results:            getStudentsResult.Students,
			PaginationResponse: paginationResponse,
		},
	}, nil
}

func (s *BackendService) GetStudentByIdHandler(ctx context.Context, req *output.GetStudentRequest) (*output.GetStudentResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	student, err := s.entityService.GetStudentById(ctx, req.StudentID)
	if err != nil {
		return nil, handleReadError(err, "identityService.GetStudentById()", "student")
	}

	return &output.GetStudentResponse{
		Data: student,
	}, nil
}

func (s *BackendService) InsertStudentsHandler(ctx context.Context, req *output.InsertStudentsRequest) (*output.InsertStudentsResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	userIDs := make([]identity.UserID, 0, len(req.Data))
	for _, param := range req.Data {
		userIDs = append(userIDs, param.UserID)
	}

	studentIDs, err := s.entityService.InsertStudents(ctx, userIDs)
	if err != nil {
		return nil, handleUpsertionError(err, "entityService.InsertStudents()", "student")
	}
	mainLog.Info("Students created: studentIDs='%v'", studentIDs)

	students, err := s.entityService.GetStudentsByIds(ctx, studentIDs)
	if err != nil {
		return nil, errs.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("entityService.GetStudentsByIds: %v", err), nil, "")
	}

	return &output.InsertStudentsResponse{
		Data: output.InsertStudentResult{
			Results: students,
		},
		Message: "Successfully created students",
	}, nil
}

func (s *BackendService) InsertStudentsWithNewUsersHandler(ctx context.Context, req *output.InsertStudentsWithNewUsersRequest) (*output.InsertStudentsWithNewUsersResponse, errs.HTTPError) {
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

	studentIDs, err := s.entityService.InsertStudentsWithNewUsers(ctx, insertUserSpecs)
	if err != nil {
		return nil, handleUpsertionError(err, "entityService.InsertStudentsWithNewUsers()", "student")
	}
	mainLog.Info("Students created: studentIDs='%v'", studentIDs)

	students, err := s.entityService.GetStudentsByIds(ctx, studentIDs)
	if err != nil {
		return nil, errs.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("entityService.GetStudentsByIds: %v", err), nil, "")
	}

	return &output.InsertStudentsWithNewUsersResponse{
		InsertStudentsResponse: output.InsertStudentsResponse{
			Data: output.InsertStudentResult{
				Results: students,
			},
			Message: "Successfully created students",
		},
	}, nil
}

func (s *BackendService) DeleteStudentsHandler(ctx context.Context, req *output.DeleteStudentsRequest) (*output.DeleteStudentsResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	studentIDs := make([]entity.StudentID, 0, len(req.Data))
	for _, param := range req.Data {
		studentIDs = append(studentIDs, param.StudentID)
	}

	err := s.entityService.DeleteStudents(ctx, studentIDs)
	if err != nil {
		return nil, handleDeletionError(err, "identityService.DeleteStudents()", "student")
	}

	return &output.DeleteStudentsResponse{
		Message: "Successfully deleted students",
	}, nil
}

func (s *BackendService) GetInstrumentsHandler(ctx context.Context, req *output.GetInstrumentsRequest) (*output.GetInstrumentsResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	getInstrumentsResult, err := s.entityService.GetInstruments(ctx, util.PaginationSpec(req.PaginationRequest))
	if err != nil {
		return nil, errs.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("entityService.GetInstruments(): %w", err), nil, "Failed to get instruments")
	}

	paginationResponse := output.NewPaginationResponse(getInstrumentsResult.PaginationResult)

	return &output.GetInstrumentsResponse{
		Data: output.GetInstrumentsResult{
			Results:            getInstrumentsResult.Instruments,
			PaginationResponse: paginationResponse,
		},
	}, nil
}

func (s *BackendService) GetInstrumentByIdHandler(ctx context.Context, req *output.GetInstrumentRequest) (*output.GetInstrumentResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	instrument, err := s.entityService.GetInstrumentById(ctx, req.InstrumentID)
	if err != nil {
		return nil, handleReadError(err, "identityService.GetInstrumentById()", "instrument")
	}

	return &output.GetInstrumentResponse{
		Data: instrument,
	}, nil
}

func (s *BackendService) InsertInstrumentsHandler(ctx context.Context, req *output.InsertInstrumentsRequest) (*output.InsertInstrumentsResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	specs := make([]entity.InsertInstrumentSpec, 0, len(req.Data))
	for _, param := range req.Data {
		specs = append(specs, entity.InsertInstrumentSpec{
			Name: param.Name,
		})
	}

	instrumentIDs, err := s.entityService.InsertInstruments(ctx, specs)
	if err != nil {
		return nil, handleUpsertionError(err, "entityService.InsertInstruments()", "instrument")
	}
	mainLog.Info("Instruments created: instrumentIDs='%v'", instrumentIDs)

	instruments, err := s.entityService.GetInstrumentsByIds(ctx, instrumentIDs)
	if err != nil {
		return nil, errs.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("entityService.GetInstrumentsByIds: %v", err), nil, "")
	}

	return &output.InsertInstrumentsResponse{
		Data: output.UpsertInstrumentResult{
			Results: instruments,
		},
		Message: "Successfully created instruments",
	}, nil
}

func (s *BackendService) UpdateInstrumentsHandler(ctx context.Context, req *output.UpdateInstrumentsRequest) (*output.UpdateInstrumentsResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	specs := make([]entity.UpdateInstrumentSpec, 0, len(req.Data))
	for _, param := range req.Data {
		specs = append(specs, entity.UpdateInstrumentSpec{
			InstrumentID: param.InstrumentID,
			Name:         param.Name,
		})
	}

	instrumentIDs, err := s.entityService.UpdateInstruments(ctx, specs)
	if err != nil {
		return nil, handleUpsertionError(err, "entityService.UpdateInstruments()", "instrument")
	}
	mainLog.Info("Instruments updated: instrumentIDs='%v'", instrumentIDs)

	instruments, err := s.entityService.GetInstrumentsByIds(ctx, instrumentIDs)
	if err != nil {
		return nil, errs.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("entityService.GetInstrumentsByIds: %v", err), nil, "")
	}

	return &output.UpdateInstrumentsResponse{
		Data: output.UpsertInstrumentResult{
			Results: instruments,
		},
		Message: "Successfully updated instruments",
	}, nil
}

func (s *BackendService) DeleteInstrumentsHandler(ctx context.Context, req *output.DeleteInstrumentsRequest) (*output.DeleteInstrumentsResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	ids := make([]entity.InstrumentID, 0, len(req.Data))
	for _, param := range req.Data {
		ids = append(ids, param.InstrumentID)
	}

	err := s.entityService.DeleteInstruments(ctx, ids)
	if err != nil {
		return nil, handleDeletionError(err, "identityService.DeleteInstruments()", "instrument")
	}

	return &output.DeleteInstrumentsResponse{
		Message: "Successfully deleted instruments",
	}, nil
}

func (s *BackendService) GetGradesHandler(ctx context.Context, req *output.GetGradesRequest) (*output.GetGradesResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	getGradesResult, err := s.entityService.GetGrades(ctx, util.PaginationSpec(req.PaginationRequest))
	if err != nil {
		return nil, errs.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("entityService.GetGrades(): %w", err), nil, "Failed to get grades")
	}

	paginationResponse := output.NewPaginationResponse(getGradesResult.PaginationResult)

	return &output.GetGradesResponse{
		Data: output.GetGradesResult{
			Results:            getGradesResult.Grades,
			PaginationResponse: paginationResponse,
		},
	}, nil
}

func (s *BackendService) GetGradeByIdHandler(ctx context.Context, req *output.GetGradeRequest) (*output.GetGradeResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	grade, err := s.entityService.GetGradeById(ctx, req.GradeID)
	if err != nil {
		return nil, handleReadError(err, "identityService.GetGradeById()", "grade")
	}

	return &output.GetGradeResponse{
		Data: grade,
	}, nil
}

func (s *BackendService) InsertGradesHandler(ctx context.Context, req *output.InsertGradesRequest) (*output.InsertGradesResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	specs := make([]entity.InsertGradeSpec, 0, len(req.Data))
	for _, param := range req.Data {
		specs = append(specs, entity.InsertGradeSpec{
			Name: param.Name,
		})
	}

	gradeIDs, err := s.entityService.InsertGrades(ctx, specs)
	if err != nil {
		return nil, handleUpsertionError(err, "entityService.InsertGrades()", "grade")
	}
	mainLog.Info("Grades created: gradeIDs='%v'", gradeIDs)

	grades, err := s.entityService.GetGradesByIds(ctx, gradeIDs)
	if err != nil {
		return nil, errs.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("entityService.GetGradesByIds: %v", err), nil, "")
	}

	return &output.InsertGradesResponse{
		Data: output.UpsertGradeResult{
			Results: grades,
		},
		Message: "Successfully created grades",
	}, nil
}

func (s *BackendService) UpdateGradesHandler(ctx context.Context, req *output.UpdateGradesRequest) (*output.UpdateGradesResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	specs := make([]entity.UpdateGradeSpec, 0, len(req.Data))
	for _, param := range req.Data {
		specs = append(specs, entity.UpdateGradeSpec{
			GradeID: param.GradeID,
			Name:    param.Name,
		})
	}

	gradeIDs, err := s.entityService.UpdateGrades(ctx, specs)
	if err != nil {
		return nil, handleUpsertionError(err, "entityService.UpdateGrades()", "grade")
	}
	mainLog.Info("Grades updated: gradeIDs='%v'", gradeIDs)

	grades, err := s.entityService.GetGradesByIds(ctx, gradeIDs)
	if err != nil {
		return nil, errs.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("entityService.GetGradesByIds: %v", err), nil, "")
	}

	return &output.UpdateGradesResponse{
		Data: output.UpsertGradeResult{
			Results: grades,
		},
		Message: "Successfully updated grades",
	}, nil
}

func (s *BackendService) DeleteGradesHandler(ctx context.Context, req *output.DeleteGradesRequest) (*output.DeleteGradesResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	ids := make([]entity.GradeID, 0, len(req.Data))
	for _, param := range req.Data {
		ids = append(ids, param.GradeID)
	}

	err := s.entityService.DeleteGrades(ctx, ids)
	if err != nil {
		return nil, handleDeletionError(err, "identityService.DeleteGrades()", "grade")
	}

	return &output.DeleteGradesResponse{
		Message: "Successfully deleted grades",
	}, nil
}

func (s *BackendService) GetCoursesHandler(ctx context.Context, req *output.GetCoursesRequest) (*output.GetCoursesResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	getCoursesResult, err := s.entityService.GetCourses(ctx, util.PaginationSpec(req.PaginationRequest))
	if err != nil {
		return nil, errs.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("entityService.GetCourses(): %w", err), nil, "Failed to get courses")
	}

	paginationResponse := output.NewPaginationResponse(getCoursesResult.PaginationResult)

	return &output.GetCoursesResponse{
		Data: output.GetCoursesResult{
			Results:            getCoursesResult.Courses,
			PaginationResponse: paginationResponse,
		},
	}, nil
}

func (s *BackendService) GetCourseByIdHandler(ctx context.Context, req *output.GetCourseRequest) (*output.GetCourseResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	course, err := s.entityService.GetCourseById(ctx, req.CourseID)
	if err != nil {
		return nil, handleReadError(err, "identityService.GetCourseById()", "course")
	}

	return &output.GetCourseResponse{
		Data: course,
	}, nil
}

func (s *BackendService) InsertCoursesHandler(ctx context.Context, req *output.InsertCoursesRequest) (*output.InsertCoursesResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	specs := make([]entity.InsertCourseSpec, 0, len(req.Data))
	for _, param := range req.Data {
		specs = append(specs, entity.InsertCourseSpec{
			InstrumentID:          param.InstrumentID,
			GradeID:               param.GradeID,
			DefaultFee:            param.DefaultFee,
			DefaultDurationMinute: param.DefaultDurationMinute,
		})
	}

	courseIDs, err := s.entityService.InsertCourses(ctx, specs)
	if err != nil {
		return nil, handleUpsertionError(err, "entityService.InsertCourses()", "course")
	}
	mainLog.Info("Courses created: courseIDs='%v'", courseIDs)

	courses, err := s.entityService.GetCoursesByIds(ctx, courseIDs)
	if err != nil {
		return nil, errs.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("entityService.GetCoursesByIds: %v", err), nil, "")
	}

	return &output.InsertCoursesResponse{
		Data: output.UpsertCourseResult{
			Results: courses,
		},
		Message: "Successfully created courses",
	}, nil
}

func (s *BackendService) UpdateCoursesHandler(ctx context.Context, req *output.UpdateCoursesRequest) (*output.UpdateCoursesResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	specs := make([]entity.UpdateCourseSpec, 0, len(req.Data))
	for _, param := range req.Data {
		specs = append(specs, entity.UpdateCourseSpec{
			CourseID:              param.CourseID,
			DefaultFee:            param.DefaultFee,
			DefaultDurationMinute: param.DefaultDurationMinute,
		})
	}

	courseIDs, err := s.entityService.UpdateCourses(ctx, specs)
	if err != nil {
		return nil, handleUpsertionError(err, "entityService.UpdateCourses()", "course")
	}
	mainLog.Info("Courses updated: courseIDs='%v'", courseIDs)

	courses, err := s.entityService.GetCoursesByIds(ctx, courseIDs)
	if err != nil {
		return nil, errs.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("entityService.GetCoursesByIds: %v", err), nil, "")
	}

	return &output.UpdateCoursesResponse{
		Data: output.UpsertCourseResult{
			Results: courses,
		},
		Message: "Successfully updated courses",
	}, nil
}

func (s *BackendService) DeleteCoursesHandler(ctx context.Context, req *output.DeleteCoursesRequest) (*output.DeleteCoursesResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	ids := make([]entity.CourseID, 0, len(req.Data))
	for _, param := range req.Data {
		ids = append(ids, param.CourseID)
	}

	err := s.entityService.DeleteCourses(ctx, ids)
	if err != nil {
		return nil, handleDeletionError(err, "identityService.DeleteCourses()", "course")
	}

	return &output.DeleteCoursesResponse{
		Message: "Successfully deleted courses",
	}, nil
}

func (s *BackendService) GetClassesHandler(ctx context.Context, req *output.GetClassesRequest) (*output.GetClassesResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	paginationSpec := util.PaginationSpec(req.PaginationRequest)
	getClassesSpec := entity.GetClassesSpec{
		IncludeDeactivated: req.IncludeDeactivated,
	}

	getClassesResult, err := s.entityService.GetClasses(ctx, paginationSpec, getClassesSpec)
	if err != nil {
		return nil, errs.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("entityService.GetClasses(): %w", err), nil, "Failed to get classes")
	}

	paginationResponse := output.NewPaginationResponse(getClassesResult.PaginationResult)

	return &output.GetClassesResponse{
		Data: output.GetClassesResult{
			Results:            getClassesResult.Classes,
			PaginationResponse: paginationResponse,
		},
	}, nil
}

func (s *BackendService) GetClassByIdHandler(ctx context.Context, req *output.GetClassRequest) (*output.GetClassResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	class, err := s.entityService.GetClassById(ctx, req.ClassID)
	if err != nil {
		return nil, handleReadError(err, "identityService.GetClassById()", "class")
	}

	return &output.GetClassResponse{
		Data: class,
	}, nil
}

func (s *BackendService) InsertClassesHandler(ctx context.Context, req *output.InsertClassesRequest) (*output.InsertClassesResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	specs := make([]entity.InsertClassSpec, 0, len(req.Data))
	for _, param := range req.Data {
		specs = append(specs, entity.InsertClassSpec{
			TeacherID:    param.TeacherID,
			StudentIDs:   param.StudentIDs,
			CourseID:     param.CourseID,
			TransportFee: param.TransportFee,
		})
	}

	classIDs, err := s.entityService.InsertClasses(ctx, specs)
	if err != nil {
		return nil, handleUpsertionError(err, "entityService.InsertClasses()", "class")
	}
	mainLog.Info("Classes created: classIDs='%v'", classIDs)

	classes, err := s.entityService.GetClassesByIds(ctx, classIDs)
	if err != nil {
		return nil, errs.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("entityService.GetClassesByIds: %v", err), nil, "")
	}

	return &output.InsertClassesResponse{
		Data: output.UpsertClassResult{
			Results: classes,
		},
		Message: "Successfully created classes",
	}, nil
}

func (s *BackendService) UpdateClassesHandler(ctx context.Context, req *output.UpdateClassesRequest) (*output.UpdateClassesResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	specs := make([]entity.UpdateClassSpec, 0, len(req.Data))
	for _, param := range req.Data {
		specs = append(specs, entity.UpdateClassSpec{
			ClassID:       param.ClassID,
			TeacherID:     param.TeacherID,
			StudentIDs:    param.StudentIDs,
			CourseID:      param.CourseID,
			TransportFee:  param.TransportFee,
			IsDeactivated: param.IsDeactivated,
		})
	}

	classIDs, err := s.entityService.UpdateClasses(ctx, specs)
	if err != nil {
		return nil, handleUpsertionError(err, "entityService.UpdateClasses()", "class")
	}
	mainLog.Info("Classes updated: classIDs='%v'", classIDs)

	classes, err := s.entityService.GetClassesByIds(ctx, classIDs)
	if err != nil {
		return nil, errs.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("entityService.GetClassesByIds: %v", err), nil, "")
	}

	return &output.UpdateClassesResponse{
		Data: output.UpsertClassResult{
			Results: classes,
		},
		Message: "Successfully updated classes",
	}, nil
}

func (s *BackendService) DeleteClassesHandler(ctx context.Context, req *output.DeleteClassesRequest) (*output.DeleteClassesResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	ids := make([]entity.ClassID, 0, len(req.Data))
	for _, param := range req.Data {
		ids = append(ids, param.ClassID)
	}

	err := s.entityService.DeleteClasses(ctx, ids)
	if err != nil {
		return nil, handleDeletionError(err, "identityService.DeleteClasses()", "class")
	}

	return &output.DeleteClassesResponse{
		Message: "Successfully deleted classes",
	}, nil
}

func (s *BackendService) GetStudentEnrollmentsHandler(ctx context.Context, req *output.GetStudentEnrollmentsRequest) (*output.GetStudentEnrollmentsResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	getStudentEnrollmentsResult, err := s.entityService.GetStudentEnrollments(ctx, util.PaginationSpec(req.PaginationRequest))
	if err != nil {
		return nil, errs.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("entityService.GetStudentEnrollments(): %w", err), nil, "Failed to get StudentEnrollments")
	}

	paginationResponse := output.NewPaginationResponse(getStudentEnrollmentsResult.PaginationResult)

	return &output.GetStudentEnrollmentsResponse{
		Data: output.GetStudentEnrollmentsResult{
			Results:            getStudentEnrollmentsResult.StudentEnrollments,
			PaginationResponse: paginationResponse,
		},
	}, nil
}

func (s *BackendService) GetTeacherSpecialFeesHandler(ctx context.Context, req *output.GetTeacherSpecialFeesRequest) (*output.GetTeacherSpecialFeesResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	getTeacherSpecialFeesResult, err := s.entityService.GetTeacherSpecialFees(ctx, util.PaginationSpec((req.PaginationRequest)))
	if err != nil {
		return nil, errs.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("entityService.GetTeacherSpecialFees(): %w", err), nil, "Failed to get teacherSpecialFees")
	}

	paginationResponse := output.NewPaginationResponse(getTeacherSpecialFeesResult.PaginationResult)

	return &output.GetTeacherSpecialFeesResponse{
		Data: output.GetTeacherSpecialFeesResult{
			Results:            getTeacherSpecialFeesResult.TeacherSpecialFees,
			PaginationResponse: paginationResponse,
		},
	}, nil
}

func (s *BackendService) GetTeacherSpecialFeeByIdHandler(ctx context.Context, req *output.GetTeacherSpecialFeeRequest) (*output.GetTeacherSpecialFeeResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	teacherSpecialFee, err := s.entityService.GetTeacherSpecialFeeById(ctx, req.TeacherSpecialFeeID)
	if err != nil {
		return nil, handleReadError(err, "identityService.GetTeacherSpecialFeeById()", "teacherSpecialFee")
	}

	return &output.GetTeacherSpecialFeeResponse{
		Data: teacherSpecialFee,
	}, nil
}

func (s *BackendService) InsertTeacherSpecialFeesHandler(ctx context.Context, req *output.InsertTeacherSpecialFeesRequest) (*output.InsertTeacherSpecialFeesResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	specs := make([]entity.InsertTeacherSpecialFeeSpec, 0, len(req.Data))
	for _, param := range req.Data {
		specs = append(specs, entity.InsertTeacherSpecialFeeSpec{
			TeacherID: param.TeacherID,
			CourseID:  param.CourseID,
			Fee:       param.Fee,
		})
	}

	teacherSpecialFeeIDs, err := s.entityService.InsertTeacherSpecialFees(ctx, specs)
	if err != nil {
		return nil, handleUpsertionError(err, "entityService.InsertTeacherSpecialFees()", "teacherSpecialFee")
	}
	mainLog.Info("TeacherSpecialFees created: teacherSpecialFeeIDs='%v'", teacherSpecialFeeIDs)

	teacherSpecialFees, err := s.entityService.GetTeacherSpecialFeesByIds(ctx, teacherSpecialFeeIDs)
	if err != nil {
		return nil, errs.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("entityService.GetTeacherSpecialFeesByIds: %v", err), nil, "")
	}

	return &output.InsertTeacherSpecialFeesResponse{
		Data: output.UpsertTeacherSpecialFeeResult{
			Results: teacherSpecialFees,
		},
		Message: "Successfully created teacherSpecialFees",
	}, nil
}

func (s *BackendService) UpdateTeacherSpecialFeesHandler(ctx context.Context, req *output.UpdateTeacherSpecialFeesRequest) (*output.UpdateTeacherSpecialFeesResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	specs := make([]entity.UpdateTeacherSpecialFeeSpec, 0, len(req.Data))
	for _, param := range req.Data {
		specs = append(specs, entity.UpdateTeacherSpecialFeeSpec{
			TeacherSpecialFeeID: param.TeacherSpecialFeeID,
			Fee:                 param.Fee,
		})
	}

	teacherSpecialFeeIDs, err := s.entityService.UpdateTeacherSpecialFees(ctx, specs)
	if err != nil {
		return nil, handleUpsertionError(err, "entityService.UpdateTeacherSpecialFees()", "teacherSpecialFee")
	}
	mainLog.Info("TeacherSpecialFees updated: teacherSpecialFeeIDs='%v'", teacherSpecialFeeIDs)

	teacherSpecialFees, err := s.entityService.GetTeacherSpecialFeesByIds(ctx, teacherSpecialFeeIDs)
	if err != nil {
		return nil, errs.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("entityService.GetTeacherSpecialFeesByIds: %v", err), nil, "")
	}

	return &output.UpdateTeacherSpecialFeesResponse{
		Data: output.UpsertTeacherSpecialFeeResult{
			Results: teacherSpecialFees,
		},
		Message: "Successfully updated teacherSpecialFees",
	}, nil
}

func (s *BackendService) DeleteTeacherSpecialFeesHandler(ctx context.Context, req *output.DeleteTeacherSpecialFeesRequest) (*output.DeleteTeacherSpecialFeesResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	ids := make([]entity.TeacherSpecialFeeID, 0, len(req.Data))
	for _, param := range req.Data {
		ids = append(ids, param.TeacherSpecialFeeID)
	}

	err := s.entityService.DeleteTeacherSpecialFees(ctx, ids)
	if err != nil {
		return nil, handleDeletionError(err, "identityService.DeleteTeacherSpecialFees()", "teacherSpecialFee")
	}

	return &output.DeleteTeacherSpecialFeesResponse{
		Message: "Successfully deleted teacherSpecialFees",
	}, nil
}

func (s *BackendService) GetEnrollmentPaymentsHandler(ctx context.Context, req *output.GetEnrollmentPaymentsRequest) (*output.GetEnrollmentPaymentsResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	paginationSpec := util.PaginationSpec(req.PaginationRequest)
	timeFilter := util.TimeSpec(req.TimeFilter)
	getEnrollmentPaymentsResult, err := s.entityService.GetEnrollmentPayments(ctx, paginationSpec, timeFilter, false)
	if err != nil {
		return nil, errs.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("entityService.GetEnrollmentPayments(): %w", err), nil, "Failed to get courses")
	}

	paginationResponse := output.NewPaginationResponse(getEnrollmentPaymentsResult.PaginationResult)

	return &output.GetEnrollmentPaymentsResponse{
		Data: output.GetEnrollmentPaymentsResult{
			Results:            getEnrollmentPaymentsResult.EnrollmentPayments,
			PaginationResponse: paginationResponse,
		},
	}, nil
}

func (s *BackendService) GetEnrollmentPaymentByIdHandler(ctx context.Context, req *output.GetEnrollmentPaymentRequest) (*output.GetEnrollmentPaymentResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	enrollmentPayment, err := s.entityService.GetEnrollmentPaymentById(ctx, req.EnrollmentPaymentID)
	if err != nil {
		return nil, handleReadError(err, "identityService.GetEnrollmentPaymentById()", "enrollmentPayment")
	}

	return &output.GetEnrollmentPaymentResponse{
		Data: enrollmentPayment,
	}, nil
}

func (s *BackendService) InsertEnrollmentPaymentsHandler(ctx context.Context, req *output.InsertEnrollmentPaymentsRequest) (*output.InsertEnrollmentPaymentsResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	specs := make([]entity.InsertEnrollmentPaymentSpec, 0, len(req.Data))
	for _, param := range req.Data {
		specs = append(specs, entity.InsertEnrollmentPaymentSpec{
			StudentEnrollmentID: param.StudentEnrollmentID,
			PaymentDate:         param.PaymentDate,
			BalanceTopUp:        param.BalanceTopUp,
			CourseFeeValue:      param.CourseFeeValue,
			TransportFeeValue:   param.TransportFeeValue,
			PenaltyFeeValue:     param.PenaltyFeeValue,
		})
	}

	enrollmentPaymentIDs, err := s.entityService.InsertEnrollmentPayments(ctx, specs)
	if err != nil {
		return nil, handleUpsertionError(err, "entityService.InsertEnrollmentPayments()", "enrollmentPayment")
	}
	mainLog.Info("EnrollmentPayments created: enrollmentPaymentIDs='%v'", enrollmentPaymentIDs)

	enrollmentPayments, err := s.entityService.GetEnrollmentPaymentsByIds(ctx, enrollmentPaymentIDs)
	if err != nil {
		return nil, errs.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("entityService.GetEnrollmentPaymentsByIds: %v", err), nil, "")
	}

	return &output.InsertEnrollmentPaymentsResponse{
		Data: output.UpsertEnrollmentPaymentResult{
			Results: enrollmentPayments,
		},
		Message: "Successfully created enrollmentPayments",
	}, nil
}

func (s *BackendService) UpdateEnrollmentPaymentsHandler(ctx context.Context, req *output.UpdateEnrollmentPaymentsRequest) (*output.UpdateEnrollmentPaymentsResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	specs := make([]entity.UpdateEnrollmentPaymentSpec, 0, len(req.Data))
	for _, param := range req.Data {
		specs = append(specs, entity.UpdateEnrollmentPaymentSpec{
			EnrollmentPaymentID: param.EnrollmentPaymentID,
			PaymentDate:         param.PaymentDate,
			BalanceTopUp:        param.BalanceTopUp,
			CourseFeeValue:      param.CourseFeeValue,
			TransportFeeValue:   param.TransportFeeValue,
			PenaltyFeeValue:     param.PenaltyFeeValue,
		})
	}

	enrollmentPaymentIDs, err := s.entityService.UpdateEnrollmentPayments(ctx, specs)
	if err != nil {
		return nil, handleUpsertionError(err, "entityService.UpdateEnrollmentPayments()", "enrollmentPayment")
	}
	mainLog.Info("EnrollmentPayments updated: enrollmentPaymentIDs='%v'", enrollmentPaymentIDs)

	enrollmentPayments, err := s.entityService.GetEnrollmentPaymentsByIds(ctx, enrollmentPaymentIDs)
	if err != nil {
		return nil, errs.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("entityService.GetEnrollmentPaymentsByIds: %v", err), nil, "")
	}

	return &output.UpdateEnrollmentPaymentsResponse{
		Data: output.UpsertEnrollmentPaymentResult{
			Results: enrollmentPayments,
		},
		Message: "Successfully updated enrollmentPayments",
	}, nil
}

func (s *BackendService) DeleteEnrollmentPaymentsHandler(ctx context.Context, req *output.DeleteEnrollmentPaymentsRequest) (*output.DeleteEnrollmentPaymentsResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	ids := make([]entity.EnrollmentPaymentID, 0, len(req.Data))
	for _, param := range req.Data {
		ids = append(ids, param.EnrollmentPaymentID)
	}

	err := s.entityService.DeleteEnrollmentPayments(ctx, ids)
	if err != nil {
		return nil, handleDeletionError(err, "identityService.DeleteEnrollmentPayments()", "enrollmentPayment")
	}

	return &output.DeleteEnrollmentPaymentsResponse{
		Message: "Successfully deleted enrollmentPayments",
	}, nil
}

func (s *BackendService) GetStudentLearningTokensHandler(ctx context.Context, req *output.GetStudentLearningTokensRequest) (*output.GetStudentLearningTokensResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	getStudentLearningTokensResult, err := s.entityService.GetStudentLearningTokens(ctx, util.PaginationSpec(req.PaginationRequest))
	if err != nil {
		return nil, errs.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("entityService.GetStudentLearningTokens(): %w", err), nil, "Failed to get courses")
	}

	paginationResponse := output.NewPaginationResponse(getStudentLearningTokensResult.PaginationResult)

	return &output.GetStudentLearningTokensResponse{
		Data: output.GetStudentLearningTokensResult{
			Results:            getStudentLearningTokensResult.StudentLearningTokens,
			PaginationResponse: paginationResponse,
		},
	}, nil
}

func (s *BackendService) GetStudentLearningTokenByIdHandler(ctx context.Context, req *output.GetStudentLearningTokenRequest) (*output.GetStudentLearningTokenResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	studentLearningToken, err := s.entityService.GetStudentLearningTokenById(ctx, req.StudentLearningTokenID)
	if err != nil {
		return nil, handleReadError(err, "identityService.GetStudentLearningTokenById()", "studentLearningToken")
	}

	return &output.GetStudentLearningTokenResponse{
		Data: studentLearningToken,
	}, nil
}

func (s *BackendService) InsertStudentLearningTokensHandler(ctx context.Context, req *output.InsertStudentLearningTokensRequest) (*output.InsertStudentLearningTokensResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	specs := make([]entity.InsertStudentLearningTokenSpec, 0, len(req.Data))
	for _, param := range req.Data {
		specs = append(specs, entity.InsertStudentLearningTokenSpec{
			StudentEnrollmentID: param.StudentEnrollmentID,
			Quota:               param.Quota,
			CourseFeeValue:      param.CourseFeeValue,
			TransportFeeValue:   param.TransportFeeValue,
		})
	}

	studentLearningTokenIDs, err := s.entityService.InsertStudentLearningTokens(ctx, specs)
	if err != nil {
		return nil, handleUpsertionError(err, "entityService.InsertStudentLearningTokens()", "studentLearningToken")
	}
	mainLog.Info("StudentLearningTokens created: studentLearningTokenIDs='%v'", studentLearningTokenIDs)

	studentLearningTokens, err := s.entityService.GetStudentLearningTokensByIds(ctx, studentLearningTokenIDs)
	if err != nil {
		return nil, errs.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("entityService.GetStudentLearningTokensByIds: %v", err), nil, "")
	}

	return &output.InsertStudentLearningTokensResponse{
		Data: output.UpsertStudentLearningTokenResult{
			Results: studentLearningTokens,
		},
		Message: "Successfully created studentLearningTokens",
	}, nil
}

func (s *BackendService) UpdateStudentLearningTokensHandler(ctx context.Context, req *output.UpdateStudentLearningTokensRequest) (*output.UpdateStudentLearningTokensResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	specs := make([]entity.UpdateStudentLearningTokenSpec, 0, len(req.Data))
	for _, param := range req.Data {
		specs = append(specs, entity.UpdateStudentLearningTokenSpec{
			StudentLearningTokenID: param.StudentLearningTokenID,
			Quota:                  param.Quota,
			CourseFeeValue:         param.CourseFeeValue,
			TransportFeeValue:      param.TransportFeeValue,
		})
	}

	studentLearningTokenIDs, err := s.entityService.UpdateStudentLearningTokens(ctx, specs)
	if err != nil {
		return nil, handleUpsertionError(err, "entityService.UpdateStudentLearningTokens()", "studentLearningToken")
	}
	mainLog.Info("StudentLearningTokens updated: studentLearningTokenIDs='%v'", studentLearningTokenIDs)

	studentLearningTokens, err := s.entityService.GetStudentLearningTokensByIds(ctx, studentLearningTokenIDs)
	if err != nil {
		return nil, errs.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("entityService.GetStudentLearningTokensByIds: %v", err), nil, "")
	}

	return &output.UpdateStudentLearningTokensResponse{
		Data: output.UpsertStudentLearningTokenResult{
			Results: studentLearningTokens,
		},
		Message: "Successfully updated studentLearningTokens",
	}, nil
}

func (s *BackendService) DeleteStudentLearningTokensHandler(ctx context.Context, req *output.DeleteStudentLearningTokensRequest) (*output.DeleteStudentLearningTokensResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	ids := make([]entity.StudentLearningTokenID, 0, len(req.Data))
	for _, param := range req.Data {
		ids = append(ids, param.StudentLearningTokenID)
	}

	err := s.entityService.DeleteStudentLearningTokens(ctx, ids)
	if err != nil {
		return nil, handleDeletionError(err, "identityService.DeleteStudentLearningTokens()", "studentLearningToken")
	}

	return &output.DeleteStudentLearningTokensResponse{
		Message: "Successfully deleted studentLearningTokens",
	}, nil
}

func (s *BackendService) GetAttendancesHandler(ctx context.Context, req *output.GetAttendancesRequest) (*output.GetAttendancesResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	paginationSpec := util.PaginationSpec(req.PaginationRequest)
	getAttendancesSpec := entity.GetAttendancesSpec{
		TimeSpec: util.TimeSpec(req.TimeFilter),
	}
	getAttendancesResult, err := s.entityService.GetAttendances(ctx, paginationSpec, getAttendancesSpec, false)
	if err != nil {
		return nil, errs.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("entityService.GetAttendances(): %w", err), nil, "Failed to get courses")
	}

	paginationResponse := output.NewPaginationResponse(getAttendancesResult.PaginationResult)

	return &output.GetAttendancesResponse{
		Data: output.GetAttendancesResult{
			Results:            getAttendancesResult.Attendances,
			PaginationResponse: paginationResponse,
		},
	}, nil
}

func (s *BackendService) GetAttendanceByIdHandler(ctx context.Context, req *output.GetAttendanceRequest) (*output.GetAttendanceResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	attendance, err := s.entityService.GetAttendanceById(ctx, req.AttendanceID)
	if err != nil {
		return nil, handleReadError(err, "identityService.GetAttendanceById()", "attendance")
	}

	return &output.GetAttendanceResponse{
		Data: attendance,
	}, nil
}

func (s *BackendService) InsertAttendancesHandler(ctx context.Context, req *output.InsertAttendancesRequest) (*output.InsertAttendancesResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	specs := make([]entity.InsertAttendanceSpec, 0, len(req.Data))
	for _, param := range req.Data {
		specs = append(specs, entity.InsertAttendanceSpec{
			ClassID:                param.ClassID,
			TeacherID:              param.TeacherID,
			StudentID:              param.StudentID,
			StudentLearningTokenID: param.StudentLearningTokenID,
			Date:                   param.Date,
			UsedStudentTokenQuota:  param.UsedStudentTokenQuota,
			Duration:               param.Duration,
			Note:                   param.Note,
		})
	}

	attendanceIDs, err := s.entityService.InsertAttendances(ctx, specs)
	if err != nil {
		return nil, handleUpsertionError(err, "entityService.InsertAttendances()", "attendance")
	}
	mainLog.Info("Attendances created: attendanceIDs='%v'", attendanceIDs)

	attendances, err := s.entityService.GetAttendancesByIds(ctx, attendanceIDs)
	if err != nil {
		return nil, errs.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("entityService.GetAttendancesByIds: %v", err), nil, "")
	}

	return &output.InsertAttendancesResponse{
		Data: output.UpsertAttendanceResult{
			Results: attendances,
		},
		Message: "Successfully created attendances",
	}, nil
}

func (s *BackendService) UpdateAttendancesHandler(ctx context.Context, req *output.UpdateAttendancesRequest) (*output.UpdateAttendancesResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	specs := make([]entity.UpdateAttendanceSpec, 0, len(req.Data))
	for _, param := range req.Data {
		specs = append(specs, entity.UpdateAttendanceSpec{
			AttendanceID:           param.AttendanceID,
			ClassID:                param.ClassID,
			TeacherID:              param.TeacherID,
			StudentID:              param.StudentID,
			StudentLearningTokenID: param.StudentLearningTokenID,
			Date:                   param.Date,
			UsedStudentTokenQuota:  param.UsedStudentTokenQuota,
			Duration:               param.Duration,
			Note:                   param.Note,
			IsPaid:                 param.IsPaid,
		})
	}

	attendanceIDs, err := s.entityService.UpdateAttendances(ctx, specs)
	if err != nil {
		return nil, handleUpsertionError(err, "entityService.UpdateAttendances()", "attendance")
	}
	mainLog.Info("Attendances updated: attendanceIDs='%v'", attendanceIDs)

	attendances, err := s.entityService.GetAttendancesByIds(ctx, attendanceIDs)
	if err != nil {
		return nil, errs.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("entityService.GetAttendancesByIds: %v", err), nil, "")
	}

	return &output.UpdateAttendancesResponse{
		Data: output.UpsertAttendanceResult{
			Results: attendances,
		},
		Message: "Successfully updated attendances",
	}, nil
}

func (s *BackendService) DeleteAttendancesHandler(ctx context.Context, req *output.DeleteAttendancesRequest) (*output.DeleteAttendancesResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	ids := make([]entity.AttendanceID, 0, len(req.Data))
	for _, param := range req.Data {
		ids = append(ids, param.AttendanceID)
	}

	err := s.entityService.DeleteAttendances(ctx, ids)
	if err != nil {
		return nil, handleDeletionError(err, "identityService.DeleteAttendances()", "attendance")
	}

	return &output.DeleteAttendancesResponse{
		Message: "Successfully deleted attendances",
	}, nil
}

func (s *BackendService) GetTeacherPaymentsHandler(ctx context.Context, req *output.GetTeacherPaymentsRequest) (*output.GetTeacherPaymentsResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	paginationSpec := util.PaginationSpec(req.PaginationRequest)
	getTeacherPaymentsSpec := entity.GetTeacherPaymentsSpec{
		TimeSpec: util.TimeSpec(req.TimeFilter),
	}
	getTeacherPaymentsResult, err := s.entityService.GetTeacherPayments(ctx, paginationSpec, getTeacherPaymentsSpec)
	if err != nil {
		return nil, errs.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("entityService.GetTeacherPayments(): %w", err), nil, "Failed to get teacherPayments")
	}

	paginationResponse := output.NewPaginationResponse(getTeacherPaymentsResult.PaginationResult)

	return &output.GetTeacherPaymentsResponse{
		Data: output.GetTeacherPaymentsResult{
			Results:            getTeacherPaymentsResult.TeacherPayments,
			PaginationResponse: paginationResponse,
		},
	}, nil
}

func (s *BackendService) SearchEnrollmentPaymentHandler(ctx context.Context, req *output.SearchEnrollmentPaymentsRequest) (*output.SearchEnrollmentPaymentsResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	timeFilter := util.TimeSpec(req.TimeFilter)
	enrollmentPayments, err := s.teachingService.SearchEnrollmentPayment(ctx, timeFilter)
	if err != nil {
		return nil, errs.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("teachingService.SearchEnrollmentPayments(): %w", err), nil, "Failed to get courses")
	}

	return &output.SearchEnrollmentPaymentsResponse{
		Data: output.SearchEnrollmentPaymentsResult{
			Results: enrollmentPayments,
		},
	}, nil
}

func (s *BackendService) GetEnrollmentPaymentInvoiceHandler(ctx context.Context, req *output.GetEnrollmentPaymentInvoiceRequest) (*output.GetEnrollmentPaymentInvoiceResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	paymentInvoice, err := s.teachingService.GetEnrollmentPaymentInvoice(ctx, req.StudentEnrollmentID)
	if err != nil {
		return nil, handleReadError(err, "teachingService.GetEnrollmentPaymentInvoice()", "studentEnrollment")
	}

	return &output.GetEnrollmentPaymentInvoiceResponse{
		Data: paymentInvoice,
	}, nil
}

func (s *BackendService) SubmitEnrollmentPaymentHandler(ctx context.Context, req *output.SubmitEnrollmentPaymentRequest) (*output.SubmitEnrollmentPaymentResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	err := s.teachingService.SubmitEnrollmentPayment(ctx, teaching.SubmitStudentEnrollmentPaymentSpec{
		StudentEnrollmentID: req.StudentEnrollmentID,
		PaymentDate:         req.PaymentDate,
		BalanceTopUp:        req.BalanceTopUp,
		PenaltyFeeValue:     req.PenaltyFeeValue,
		CourseFeeValue:      req.CourseFeeValue,
		TransportFeeValue:   req.TransportFeeValue,
	})
	if err != nil {
		return nil, handleUpsertionError(err, "teachingService.SubmitStudentEnrollmentPayment()", "enrollmentPayment")
	}

	return &output.SubmitEnrollmentPaymentResponse{
		Message: "Successfully submitted enrollmentPayment",
	}, nil
}

func (s *BackendService) EditEnrollmentPaymentHandler(ctx context.Context, req *output.EditEnrollmentPaymentRequest) (*output.EditEnrollmentPaymentResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	enrollmentPaymentID, err := s.teachingService.EditEnrollmentPayment(ctx, teaching.EditStudentEnrollmentPaymentSpec{
		EnrollmentPaymentID: req.EnrollmentPaymentID,
		PaymentDate:         req.PaymentDate,
		BalanceTopUp:        req.BalanceTopUp,
	})
	if err != nil {
		return nil, handleReadUpsertError(err, "teachingService.EditEnrollmentPayment()", "enrollmentPayment")
	}

	enrollmentPayment, err := s.entityService.GetEnrollmentPaymentById(ctx, enrollmentPaymentID)
	if err != nil {
		return nil, errs.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("entityService.GetEnrollmentPaymentsByIds: %v", err), nil, "")
	}

	return &output.EditEnrollmentPaymentResponse{
		Data:    enrollmentPayment,
		Message: "Successfully edited enrollmentPayment",
	}, nil
}

func (s *BackendService) RemoveEnrollmentPaymentHandler(ctx context.Context, req *output.RemoveEnrollmentPaymentRequest) (*output.RemoveEnrollmentPaymentResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	err := s.teachingService.RemoveEnrollmentPayment(ctx, req.EnrollmentPaymentID)
	if err != nil {
		return nil, handleDeletionError(err, "teachingService.RemoveStudentEnrollmentPayment()", "enrollmentPayment")
	}

	return &output.RemoveEnrollmentPaymentResponse{
		Message: "Successfully removed enrollmentPayment",
	}, nil
}

func (s *BackendService) SearchClass(ctx context.Context, req *output.SearchClassRequest) (*output.SearchClassResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	authInfo := network.GetAuthInfo(ctx)
	if authInfo.PrivilegeType < identity.UserPrivilegeType_Staff {
		userTeachingInfo, err := s.teachingService.GetUserTeachingInfo(ctx, authInfo.UserID)
		if err != nil {
			return nil, handleReadError(err, "teachingService.GetUserTeachingInfo()", "class")
		}
		if req.StudentID != entity.StudentID_None && req.StudentID != userTeachingInfo.StudentID {
			errContext := fmt.Errorf("unauthorized SearchClass(): userId='%d', studentId='%d', requestedStudentId='%d'", authInfo.UserID, userTeachingInfo.StudentID, req.StudentID)
			return nil, errs.NewHTTPError(http.StatusForbidden, errContext, nil, "You're not authorized to search this class. Please contact the system administrator for further information.")
		}
		if req.TeacherID != entity.TeacherID_None && req.TeacherID != userTeachingInfo.TeacherID {
			errContext := fmt.Errorf("unauthorized SearchClass(): userId='%d', teacherId='%d', requestedTeacherId='%d'", authInfo.UserID, userTeachingInfo.TeacherID, req.TeacherID)
			return nil, errs.NewHTTPError(http.StatusForbidden, errContext, nil, "You're not authorized to search this class. Please contact the system administrator for further information.")
		}
	}

	spec := teaching.SearchClassSpec{
		TeacherID: req.TeacherID,
		StudentID: req.StudentID,
		CourseID:  req.CourseID,
	}
	classes, err := s.teachingService.SearchClass(ctx, spec)
	if err != nil {
		return nil, handleReadError(err, "teachingService.SearchClass()", "class")
	}

	return &output.SearchClassResponse{
		Data: output.SearchClassResult{
			Results: classes,
		},
	}, nil
}

func (s *BackendService) GetStudentLearningTokensByClassIDHandler(ctx context.Context, req *output.GetStudentLearningTokensByClassIDRequest) (*output.GetStudentLearningTokensByClassIDResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	authInfo := network.GetAuthInfo(ctx)
	if authInfo.PrivilegeType < identity.UserPrivilegeType_Staff {
		isInvolved, err := s.teachingService.IsUserInvolvedInClass(ctx, authInfo.UserID, req.ClassID)
		if err != nil {
			return nil, handleReadError(err, "teachingService.IsUserInvolvedInClass()", "studentLearningToken")
		}
		if !isInvolved {
			errContext := fmt.Errorf("unauthorized GetStudentLearningTokensByClassIDHandler(): userId='%d', requestedClassId='%d'", authInfo.UserID, req.ClassID)
			return nil, errs.NewHTTPError(http.StatusForbidden, errContext, nil, "You're not involved in this class to read its student learning tokens. Please contact the system administrator for further information.")
		}
	}

	getStudentLearningTokensResults, err := s.teachingService.GetSLTsByClassID(ctx, req.ClassID)
	if err != nil {
		return nil, handleReadError(err, "teachingService.GetSLTsByClassID()", "studentLearningToken")
	}

	return &output.GetStudentLearningTokensByClassIDResponse{
		Data: output.GetStudentLearningTokensByClassIDResult{
			Results: getStudentLearningTokensResults,
		},
	}, nil
}

func (s *BackendService) GetAttendancesByClassIDHandler(ctx context.Context, req *output.GetAttendancesByClassIDRequest) (*output.GetAttendancesByClassIDResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	authInfo := network.GetAuthInfo(ctx)
	if authInfo.PrivilegeType < identity.UserPrivilegeType_Staff {
		isInvolved, err := s.teachingService.IsUserInvolvedInClass(ctx, authInfo.UserID, req.ClassID)
		if err != nil {
			return nil, handleReadError(err, "teachingService.IsUserInvolvedInClass()", "attendance")
		}
		if !isInvolved {
			errContext := fmt.Errorf("unauthorized GetAttendancesByClassIDHandler(): userId='%d', requestedClassId='%d'", authInfo.UserID, req.ClassID)
			return nil, errs.NewHTTPError(http.StatusForbidden, errContext, nil, "You're not involved in this class to read its attendances. Please contact the system administrator for further information.")
		}
	}

	timeFilter := req.YearMonthFilter.ToTimeFilter(output.YearMonthFilterType_Standard)

	spec := teaching.GetAttendancesByClassIDSpec{
		ClassID:        req.ClassID,
		StudentID:      req.StudentID,
		PaginationSpec: util.PaginationSpec(req.PaginationRequest),
		TimeSpec:       util.TimeSpec(timeFilter),
	}
	getAttendancesResult, err := s.teachingService.GetAttendancesByClassID(ctx, spec)
	if err != nil {
		return nil, handleReadError(err, "teachingService.GetAttendancesByClassID()", "class")
	}

	return &output.GetAttendancesByClassIDResponse{
		Data: output.GetAttendancesByClassIDResult{
			Results: getAttendancesResult.Attendances,
			PaginationResponse: output.PaginationResponse{
				TotalPages:   getAttendancesResult.PaginationResult.TotalPages,
				TotalResults: getAttendancesResult.PaginationResult.TotalResults,
				CurrentPage:  getAttendancesResult.PaginationResult.CurrentPage,
			},
		},
	}, nil
}

func (s *BackendService) AddAttendanceHandler(ctx context.Context, req *output.AddAttendanceRequest) (*output.AddAttendanceResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	authInfo := network.GetAuthInfo(ctx)
	if authInfo.PrivilegeType < identity.UserPrivilegeType_Staff {
		isInvolved, err := s.teachingService.IsUserInvolvedInClass(ctx, authInfo.UserID, req.ClassID)
		if err != nil {
			return nil, handleReadError(err, "teachingService.IsUserInvolvedInClass()", "userInvolvementInClass")
		}
		if !isInvolved {
			errContext := fmt.Errorf("unauthorized GetAttendancesByClassIDHandler(): userId='%d', requestedClassId='%d'", authInfo.UserID, req.ClassID)
			return nil, errs.NewHTTPError(http.StatusForbidden, errContext, nil, "You're not involved in this class as a teacher to add new attendances. Please contact the system administrator for further information.")
		}
	}

	allowAutoCreateSLT := configObject.AllowAutoCreateSLTOnAddAttendance

	attendanceIDs, err := s.teachingService.AddAttendance(ctx, teaching.AddAttendanceSpec{
		ClassID:               req.ClassID,
		TeacherID:             req.TeacherID,
		Date:                  req.Date,
		UsedStudentTokenQuota: req.UsedStudentTokenQuota,
		Duration:              req.Duration,
		Note:                  req.Note,
	}, allowAutoCreateSLT)
	if err != nil {
		errContext := fmt.Errorf("teachingService.AddAttendance(): %w", err)
		if errors.Is(err, errs.ErrClassHaveNoStudent) {
			return nil, errs.NewHTTPError(http.StatusUnprocessableEntity, errContext, nil, "This class doesn't have any student, try registering a student first")
		}
		if errors.Is(err, errs.ErrStudentEnrollmentHaveNoLearningToken) {
			return nil, errs.NewHTTPError(http.StatusUnprocessableEntity, errContext, nil, "One/more students of this class don't have learningToken, try adding students' enrollmentPayment first")
		}

		return nil, handleUpsertionError(err, errContext.Error(), "attendance")
	}
	mainLog.Info("Attendances added: attendanceIDs='%v'", attendanceIDs)

	attendances, err := s.entityService.GetAttendancesByIds(ctx, attendanceIDs)
	if err != nil {
		return nil, handleReadError(err, "entityService.GetAttendancesByIds()", "attendance")
	}

	return &output.AddAttendanceResponse{
		Data: output.UpsertAttendanceResult{
			Results: attendances,
		},
		Message: "Successfully added attendances",
	}, nil
}

func (s *BackendService) EditAttendanceHandler(ctx context.Context, req *output.EditAttendanceRequest) (*output.EditAttendanceResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	authInfo := network.GetAuthInfo(ctx)
	if authInfo.PrivilegeType < identity.UserPrivilegeType_Staff {
		isInvolved, err := s.teachingService.IsUserInvolvedInAttendance(ctx, authInfo.UserID, req.AttendanceID)
		if err != nil {
			return nil, handleReadError(err, "teachingService.IsUserInvolvedInAttendance()", "userInvolvementInAttendance")
		}
		if !isInvolved {
			errContext := fmt.Errorf("unauthorized GetAttendancesByAttendanceIDHandler(): userId='%d', requestedAttendanceId='%d'", authInfo.UserID, req.AttendanceID)
			return nil, errs.NewHTTPError(http.StatusForbidden, errContext, nil, "You're not involved in this attendance as a teacher to edit. Please contact the system administrator for further information.")
		}
	}

	attendanceIDs, err := s.teachingService.EditAttendance(ctx, teaching.EditAttendanceSpec{
		AttendanceID:          req.AttendanceID,
		TeacherID:             req.TeacherID,
		Date:                  req.Date,
		UsedStudentTokenQuota: req.UsedStudentTokenQuota,
		Duration:              req.Duration,
		Note:                  req.Note,
	})
	if err != nil {
		errContext := fmt.Errorf("teachingService.EditAttendance(): %w", err)
		if errors.Is(err, errs.ErrModifyingPaidAttendance) {
			return nil, errs.NewHTTPError(http.StatusUnprocessableEntity, errContext, nil, "You are editing a paid attendance, try de-registering the attendance from teacher payment first")
		}

		return nil, handleUpsertionError(err, errContext.Error(), "attendance")
	}
	mainLog.Info("Attendances edited: attendanceIDs='%v'", attendanceIDs)

	attendances, err := s.entityService.GetAttendancesByIds(ctx, attendanceIDs)
	if err != nil {
		return nil, handleReadError(err, "entityService.GetAttendancesByIds()", "attendance")
	}

	return &output.EditAttendanceResponse{
		Data: output.UpsertAttendanceResult{
			Results: attendances,
		},
		Message: "Successfully edited attendance",
	}, nil
}

func (s *BackendService) RemoveAttendanceHandler(ctx context.Context, req *output.RemoveAttendanceRequest) (*output.RemoveAttendanceResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	authInfo := network.GetAuthInfo(ctx)
	if authInfo.PrivilegeType < identity.UserPrivilegeType_Staff {
		isInvolved, err := s.teachingService.IsUserInvolvedInAttendance(ctx, authInfo.UserID, req.AttendanceID)
		if err != nil {
			return nil, handleReadError(err, "teachingService.IsUserInvolvedInAttendance()", "userInvolvementInAttendance")
		}
		if !isInvolved {
			errContext := fmt.Errorf("unauthorized GetAttendancesByAttendanceIDHandler(): userId='%d', requestedAttendanceId='%d'", authInfo.UserID, req.AttendanceID)
			return nil, errs.NewHTTPError(http.StatusForbidden, errContext, nil, "You're not involved in this attendance as a teacher to remove. Please contact the system administrator for further information.")
		}
	}

	attendanceIDs, err := s.teachingService.RemoveAttendance(ctx, req.AttendanceID)
	if err != nil {
		errContext := fmt.Errorf("teachingService.RemoveAttendance(): %w", err)
		if errors.Is(err, errs.ErrModifyingPaidAttendance) {
			return nil, errs.NewHTTPError(http.StatusUnprocessableEntity, errContext, nil, "You are removing a paid attendance, try de-registering the attendance from teacher payment first")
		}

		return nil, handleUpsertionError(err, errContext.Error(), "attendance")
	}
	mainLog.Info("Attendances removed: attendanceIDs='%v'", attendanceIDs)

	return &output.RemoveAttendanceResponse{
		Message: "Successfully removed attendance",
	}, nil
}

func (s *BackendService) GetUnpaidTeachersHandler(ctx context.Context, req *output.GetUnpaidTeachersRequest) (*output.GetUnpaidTeachersResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	timeFilter := req.YearMonthFilter.ToTimeFilter(output.YearMonthFilterType_CalculatingSalary)

	getUnpaidTeachersResult, err := s.teachingService.GetTeachersForPayment(ctx, teaching.GetTeachersForPaymentSpec{
		IsPaid:     false,
		Pagination: util.PaginationSpec(req.PaginationRequest),
		TimeSpec:   util.TimeSpec(timeFilter),
	})
	if err != nil {
		return nil, handleReadError(err, "teachingService.GetTeachersForPayment(IsPaid=false)", "teacherPayment")
	}

	paginationResponse := output.NewPaginationResponse(getUnpaidTeachersResult.PaginationResult)

	return &output.GetUnpaidTeachersResponse{
		Data: output.GetUnpaidTeachersResult{
			Results:            getUnpaidTeachersResult.TeachersForPayment,
			PaginationResponse: paginationResponse,
		},
	}, nil
}

func (s *BackendService) GetPaidTeachersHandler(ctx context.Context, req *output.GetPaidTeachersRequest) (*output.GetPaidTeachersResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	timeFilter := req.YearMonthFilter.ToTimeFilter(output.YearMonthFilterType_Standard)

	getPaidTeachersResult, err := s.teachingService.GetTeachersForPayment(ctx, teaching.GetTeachersForPaymentSpec{
		IsPaid:     true,
		Pagination: util.PaginationSpec(req.PaginationRequest),
		TimeSpec:   util.TimeSpec(timeFilter),
	})
	if err != nil {
		return nil, handleReadError(err, "teachingService.GetTeachersForPayment(IsPaid=true)", "teacherPayment")
	}

	paginationResponse := output.NewPaginationResponse(getPaidTeachersResult.PaginationResult)

	return &output.GetPaidTeachersResponse{
		Data: output.GetPaidTeachersResult{
			Results:            getPaidTeachersResult.TeachersForPayment,
			PaginationResponse: paginationResponse,
		},
	}, nil
}

// GetTeacherPaymentInvoiceItemsHandler gets all `Attendances`s within the range of specific rule (prev month's 28th to curr month's 27) for a selected TeacherID, then convert to `TeacherPaymentInvoiceItem`s.
func (s *BackendService) GetTeacherPaymentInvoiceItemsHandler(ctx context.Context, req *output.GetTeacherPaymentInvoiceItemsRequest) (*output.GetTeacherPaymentInvoiceItemsResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	timeFilter := req.YearMonthFilter.ToTimeFilter(output.YearMonthFilterType_CalculatingSalary)

	invoiceItems, err := s.teachingService.GetTeacherPaymentInvoiceItems(ctx, teaching.GetTeacherPaymentInvoiceItemsSpec{
		TeacherID: req.TeacherID,
		TimeSpec:  util.TimeSpec(timeFilter),
	})
	if err != nil {
		return nil, handleReadError(err, "teachingService.GetTeacherPaymentInvoiceItems()", "teacherPayment")
	}

	return &output.GetTeacherPaymentInvoiceItemsResponse{
		Data: output.GetTeacherPaymentInvoiceItemsResult{
			Results: invoiceItems,
		},
	}, nil
}

// GetTeacherPaymentsAsInvoiceItemsHandler gets all `TeacherPayment`s within the range of a whole month (curr month's 1th to last day) for a selected TeacherID, then convert to `TeacherPaymentInvoiceItem`s.
func (s *BackendService) GetTeacherPaymentsAsInvoiceItemsHandler(ctx context.Context, req *output.GetTeacherPaymentsAsInvoiceItemsRequest) (*output.GetTeacherPaymentsAsInvoiceItemsResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	attendanceTimeFilter := req.YearMonthFilter.ToTimeFilter(output.YearMonthFilterType_Standard)

	invoiceItems, err := s.teachingService.GetExistingTeacherPaymentInvoiceItems(ctx, teaching.GetExistingTeacherPaymentInvoiceItemsSpec{
		TeacherID: req.TeacherID,
		TimeSpec:  util.TimeSpec(attendanceTimeFilter),
	})
	if err != nil {
		return nil, handleReadError(err, "teachingService.GetExistingTeacherPaymentInvoiceItems()", "teacherPayment")
	}

	return &output.GetTeacherPaymentsAsInvoiceItemsResponse{
		Data: output.GetTeacherPaymentsAsInvoiceItemsResult{
			Results: invoiceItems,
		},
	}, nil
}

func (s *BackendService) SubmitTeacherPaymentsHandler(ctx context.Context, req *output.SubmitTeacherPaymentsRequest) (*output.SubmitTeacherPaymentsResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	specs := make([]teaching.SubmitTeacherPaymentsSpec, 0, len(req.Data))
	for _, param := range req.Data {
		specs = append(specs, teaching.SubmitTeacherPaymentsSpec{
			AttendanceID:          param.AttendanceID,
			PaidCourseFeeValue:    param.PaidCourseFeeValue,
			PaidTransportFeeValue: param.PaidTransportFeeValue,
		})
	}

	err := s.teachingService.SubmitTeacherPayments(ctx, specs)
	if err != nil {
		return nil, handleUpsertionError(err, "teachingService.SubmitTeacherPayments()", "TeacherPayments")
	}

	return &output.SubmitTeacherPaymentsResponse{
		Message: "Successfully submitted TeacherPayments",
	}, nil
}

func (s *BackendService) ModifyTeacherPaymentsHandler(ctx context.Context, req *output.ModifyTeacherPaymentsRequest) (*output.ModifyTeacherPaymentsResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	specs := make([]teaching.ModifyTeacherPaymentsSpec, 0, len(req.Data))
	for _, param := range req.Data {
		specs = append(specs, teaching.ModifyTeacherPaymentsSpec{
			TeacherPaymentID:      param.TeacherPaymentID,
			PaidCourseFeeValue:    param.PaidCourseFeeValue,
			PaidTransportFeeValue: param.PaidTransportFeeValue,
			IsDeleted:             param.IsDeleted,
		})
	}

	modifyTeacherPaymentsResult, err := s.teachingService.ModifyTeacherPayments(ctx, specs)
	if err != nil {
		return nil, handleUpsertionError(err, "teachingService.ModifyTeacherPayments()", "teacherPayment")
	}
	mainLog.Info("TeacherPayments edited: teacherPaymentIDs='%v'; removed: teacherPaymentIDs='%v'", modifyTeacherPaymentsResult.EditedTeacherPaymentIDs, modifyTeacherPaymentsResult.DeletedTeacherPaymentIDs)

	teacherPayments, err := s.entityService.GetTeacherPaymentsByIds(ctx, modifyTeacherPaymentsResult.EditedTeacherPaymentIDs)
	if err != nil {
		return nil, errs.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("entityService.GetTeacherPaymentsByIds: %v", err), nil, "")
	}

	return &output.ModifyTeacherPaymentsResponse{
		Data: output.ModifyTeacherPaymentsResult{
			Results: teacherPayments,
		},
		Message: "Successfully modified teacherPayment",
	}, nil
}

func (s *BackendService) EditTeacherPaymentsHandler(ctx context.Context, req *output.EditTeacherPaymentsRequest) (*output.EditTeacherPaymentsResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	specs := make([]teaching.EditTeacherPaymentsSpec, 0, len(req.Data))
	for _, param := range req.Data {
		specs = append(specs, teaching.EditTeacherPaymentsSpec{
			TeacherPaymentID:      param.TeacherPaymentID,
			PaidCourseFeeValue:    param.PaidCourseFeeValue,
			PaidTransportFeeValue: param.PaidTransportFeeValue,
		})
	}

	teacherPaymentIDs, err := s.teachingService.EditTeacherPayments(ctx, specs)
	if err != nil {
		return nil, handleUpsertionError(err, "teachingService.EditTeacherPayments()", "teacherPayment")
	}
	mainLog.Info("TeacherPayments edited: teacherPaymentIDs='%v'", teacherPaymentIDs)

	teacherPayments, err := s.entityService.GetTeacherPaymentsByIds(ctx, teacherPaymentIDs)
	if err != nil {
		return nil, errs.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("entityService.GetTeacherPaymentsByIds: %v", err), nil, "")
	}

	return &output.EditTeacherPaymentsResponse{
		Data: output.EditTeacherPaymentsResult{
			Results: teacherPayments,
		},
		Message: "Successfully edited teacherPayment",
	}, nil
}

func (s *BackendService) RemoveTeacherPaymentsHandler(ctx context.Context, req *output.RemoveTeacherPaymentsRequest) (*output.RemoveTeacherPaymentsResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	ids := make([]entity.TeacherPaymentID, 0, len(req.Data))
	for _, param := range req.Data {
		ids = append(ids, param.TeacherPaymentID)
	}

	err := s.teachingService.RemoveTeacherPayments(ctx, ids)
	if err != nil {
		return nil, handleUpsertionError(err, "teachingService.RemoveTeacherPayments():", "teacherPayment")
	}
	mainLog.Info("TeacherPaymentss removed: teacherPaymentIDs='%v'", ids)

	return &output.RemoveTeacherPaymentsResponse{
		Message: "Successfully removed teacherPayment",
	}, nil
}

func (s *BackendService) GetDashboardExpenseOverview(ctx context.Context, req *output.GetDashboardExpenseOverviewRequest) (*output.GetDashboardExpenseOverviewResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	overviewResult, err := s.dashboardService.GetExpenseOverview(ctx, dashboard.GetExpenseOverviewSpec{
		TimeSpec:      util.TimeSpec(req.DateRange.ToTimeFilter(output.YearMonthFilterType_Standard)),
		TeacherIDs:    req.TeacherIDs,
		InstrumentIDs: req.InstrumentIDs,
	})
	if err != nil {
		return &output.GetDashboardExpenseOverviewResponse{}, errs.NewHTTPError(http.StatusInternalServerError,
			fmt.Errorf("dashboardService.GetExpenseOverview(): %v", err), nil, "Failed to get dashboardExpenseOverview data")
	}

	return &output.GetDashboardExpenseOverviewResponse{
		Data: output.GetDashboardExpenseOverviewResult{
			Results: overviewResult.Data,
		},
	}, nil
}
func (s *BackendService) GetDashboardExpenseMonthlySummary(ctx context.Context, req *output.GetDashboardExpenseMonthlySummaryRequest) (*output.GetDashboardExpenseMonthlySummaryResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	monthlySummary, err := s.dashboardService.GetExpenseMonthlySummary(ctx, dashboard.GetExpenseMontlySummarySpec{
		TimeSpec:      util.TimeSpec(req.SelectedDate.ToTimeFilter(output.YearMonthFilterType_Standard)),
		GroupBy:       req.GroupBy,
		TeacherIDs:    req.TeacherIDs,
		InstrumentIDs: req.InstrumentIDs,
	})
	if err != nil {
		return &output.GetDashboardExpenseMonthlySummaryResponse{}, errs.NewHTTPError(http.StatusInternalServerError,
			fmt.Errorf("dashboardService.GetExpenseMonthlySummary(): %v", err), nil, "Failed to get dashboardExpenseMonthlySummary data")
	}

	return &output.GetDashboardExpenseMonthlySummaryResponse{
		Data: output.GetDashboardExpenseMonthlySummaryResult{
			Results: monthlySummary.Data,
		},
	}, nil
}
func (s *BackendService) GetDashboardIncomeOverview(ctx context.Context, req *output.GetDashboardIncomeOverviewRequest) (*output.GetDashboardIncomeOverviewResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	overviewResult, err := s.dashboardService.GetIncomeOverview(ctx, dashboard.GetIncomeOverviewSpec{
		TimeSpec:      util.TimeSpec(req.DateRange.ToTimeFilter(output.YearMonthFilterType_Standard)),
		StudentIDs:    req.StudentIDs,
		InstrumentIDs: req.InstrumentIDs,
	})
	if err != nil {
		return &output.GetDashboardIncomeOverviewResponse{}, errs.NewHTTPError(http.StatusInternalServerError,
			fmt.Errorf("dashboardService.GetIncomeOverview(): %v", err), nil, "Failed to get dashboardIncomeOverview data")
	}

	return &output.GetDashboardIncomeOverviewResponse{
		Data: output.GetDashboardIncomeOverviewResult{
			Results: overviewResult.Data,
		},
	}, nil
}
func (s *BackendService) GetDashboardIncomeMonthlySummary(ctx context.Context, req *output.GetDashboardIncomeMonthlySummaryRequest) (*output.GetDashboardIncomeMonthlySummaryResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	monthlySummary, err := s.dashboardService.GetIncomeMonthlySummary(ctx, dashboard.GetIncomeMontlySummarySpec{
		TimeSpec:      util.TimeSpec(req.SelectedDate.ToTimeFilter(output.YearMonthFilterType_Standard)),
		GroupBy:       req.GroupBy,
		StudentIDs:    req.StudentIDs,
		InstrumentIDs: req.InstrumentIDs,
	})
	if err != nil {
		return &output.GetDashboardIncomeMonthlySummaryResponse{}, errs.NewHTTPError(http.StatusInternalServerError,
			fmt.Errorf("dashboardService.GetIncomeMonthlySummary(): %v", err), nil, "Failed to get dashboardIncomeMonthlySummary data")
	}

	return &output.GetDashboardIncomeMonthlySummaryResponse{
		Data: output.GetDashboardIncomeMonthlySummaryResult{
			Results: monthlySummary.Data,
		},
	}, nil
}

func (s *BackendService) GetUserProfile(ctx context.Context, req *output.GetUserProfileRequest) (*output.GetUserProfileResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	authInfo := network.GetAuthInfo(ctx)

	user, err := s.identityService.GetUserById(ctx, authInfo.UserID)
	if err != nil {
		return nil, handleReadError(err, "identityService.GetUserProfileById()", "user")
	}

	return &output.GetUserProfileResponse{
		Data: user,
	}, nil
}

func (s *BackendService) GetUserTeachingInfo(ctx context.Context, req *output.GetUserTeachingInfoRequest) (*output.GetUserTeachingInfoResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	authInfo := network.GetAuthInfo(ctx)

	userTeachingInfo, err := s.teachingService.GetUserTeachingInfo(ctx, authInfo.UserID)
	if err != nil {
		return nil, handleReadError(err, "teachingService.GetUserTeachingInfo()", "user")
	}

	return &output.GetUserTeachingInfoResponse{
		Data: userTeachingInfo,
	}, nil
}

// handleReadUpsertError combines handleReadError & handleUpsertError.
func handleReadUpsertError(err error, methodName, entityName string) errs.HTTPError {
	if err == nil {
		return nil
	}

	var validationErr errs.ValidationError
	wrappedErr := fmt.Errorf("%s: %w", methodName, err)
	if errors.Is(err, sql.ErrNoRows) {
		return errs.NewHTTPError(http.StatusNotFound, wrappedErr, nil, fmt.Sprintf("%s is not found", strings.Title(entityName)))
	} else if errors.As(err, &validationErr) {
		return errs.NewHTTPError(http.StatusConflict, wrappedErr, validationErr.GetErrorDetail(), fmt.Sprintf("Invalid %s properties", entityName))
	}
	return errs.NewHTTPError(http.StatusInternalServerError, wrappedErr, nil, fmt.Sprintf("Failed to get %s", entityName))
}

// handleReadError detects non-existing result error (e.g. sql.ErrNoRows) and returns HTTP 404-NotFound. Else, returns HTTP 500.
func handleReadError(err error, methodName, entityName string) errs.HTTPError {
	if err == nil {
		return nil
	}

	wrappedErr := fmt.Errorf("%s: %w", methodName, err)
	if errors.Is(err, sql.ErrNoRows) {
		return errs.NewHTTPError(http.StatusNotFound, wrappedErr, nil, fmt.Sprintf("%s is not found", strings.Title(entityName)))
	}
	return errs.NewHTTPError(http.StatusInternalServerError, wrappedErr, nil, fmt.Sprintf("Failed to get %s", entityName))
}

// handleReadError detects update/insert error due to rule violation (e.g. duplicate entries) and returns HTTP 409-Conflict. Else, returns HTTP 500.
func handleUpsertionError(err error, methodName, entityName string) errs.HTTPError {
	if err == nil {
		return nil
	}

	var validationErr errs.ValidationError
	if errors.As(err, &validationErr) {
		return errs.NewHTTPError(http.StatusConflict, fmt.Errorf("%s: %v", methodName, err), validationErr.GetErrorDetail(), fmt.Sprintf("Invalid %s properties", entityName))
	}
	return errs.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("%s: %v", methodName, err), nil, fmt.Sprintf("Failed to create or update %s(s)", entityName))
}

// handleReadError detects update/insert error due to rule violation (e.g. referred row) and returns HTTP 409-Conflict. Else, returns HTTP 500.
func handleDeletionError(err error, methodName, entityName string) errs.HTTPError {
	if err == nil {
		return nil
	}

	var validationErr errs.ValidationError
	if errors.As(err, &validationErr) {
		return errs.NewHTTPError(
			http.StatusConflict,
			fmt.Errorf("%s: %v", methodName, err),
			validationErr.GetErrorDetail(),
			fmt.Sprintf("Unable to delete %s(s) as it is still required by another entity. You need to remove all other entities which still refer to this %s(s). If removing is not possible, you can deactivate the %s(s)", entityName, entityName, entityName),
		)
	}
	return errs.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("%s: %v", methodName, err), nil, fmt.Sprintf("Failed to delete %s(s)", entityName))
}
