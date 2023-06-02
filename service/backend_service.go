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
		Message: "Successfully signed up user",
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
		Message: "Successfully reset password",
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
			ID:                param.ID,
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

func (s *BackendService) UpdateUserPasswordHandler(ctx context.Context, req *output.UpdateUserPasswordRequest) (*output.UpdateUserPasswordResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	err := s.identityService.UpdateUserPassword(ctx, identity.UpdateUserPasswordSpec{
		ID:       req.ID,
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

	teacherIDs, err := s.teachingService.InsertTeachers(ctx, userIDs)
	if err != nil {
		return nil, handleUpsertionError(err, "teachingService.InsertTeachers()", "teacher")
	}
	mainLog.Info("Teachers created: teacherIDs='%v'", teacherIDs)

	teachers, err := s.teachingService.GetTeachersByIds(ctx, teacherIDs)
	if err != nil {
		return nil, errs.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("teachingService.GetTeachersByIds: %v", err), nil, "")
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

	teacherIDs, err := s.teachingService.InsertTeachersWithNewUsers(ctx, insertUserSpecs)
	if err != nil {
		return nil, handleUpsertionError(err, "teachingService.InsertTeachersWithNewUsers()", "teacher")
	}
	mainLog.Info("Teachers created: teacherIDs='%v'", teacherIDs)

	teachers, err := s.teachingService.GetTeachersByIds(ctx, teacherIDs)
	if err != nil {
		return nil, errs.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("teachingService.GetTeachersByIds: %v", err), nil, "")
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

	studentIDs, err := s.teachingService.InsertStudents(ctx, userIDs)
	if err != nil {
		return nil, handleUpsertionError(err, "teachingService.InsertStudents()", "student")
	}
	mainLog.Info("Students created: studentIDs='%v'", studentIDs)

	students, err := s.teachingService.GetStudentsByIds(ctx, studentIDs)
	if err != nil {
		return nil, errs.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("teachingService.GetStudentsByIds: %v", err), nil, "")
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

	studentIDs, err := s.teachingService.InsertStudentsWithNewUsers(ctx, insertUserSpecs)
	if err != nil {
		return nil, handleUpsertionError(err, "teachingService.InsertStudentsWithNewUsers()", "student")
	}
	mainLog.Info("Students created: studentIDs='%v'", studentIDs)

	students, err := s.teachingService.GetStudentsByIds(ctx, studentIDs)
	if err != nil {
		return nil, errs.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("teachingService.GetStudentsByIds: %v", err), nil, "")
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

func (s *BackendService) GetInstrumentsHandler(ctx context.Context, req *output.GetInstrumentsRequest) (*output.GetInstrumentsResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	getInstrumentsResult, err := s.teachingService.GetInstruments(ctx, util.PaginationSpec{
		Page:           req.Page,
		ResultsPerPage: req.ResultsPerPage,
	})
	if err != nil {
		return nil, errs.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("teachingService.GetInstruments(): %w", err), nil, "Failed to get instruments")
	}

	paginationResult := getInstrumentsResult.PaginationResult

	return &output.GetInstrumentsResponse{
		Data: output.GetInstrumentsResult{
			Results: getInstrumentsResult.Instruments,
			PaginationResponse: output.PaginationResponse{
				TotalPages:   paginationResult.TotalPages,
				TotalResults: paginationResult.TotalResults,
				CurrentPage:  paginationResult.CurrentPage,
			},
		},
	}, nil
}

func (s *BackendService) GetInstrumentByIdHandler(ctx context.Context, req *output.GetInstrumentRequest) (*output.GetInstrumentResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	instrument, err := s.teachingService.GetInstrumentById(ctx, req.ID)
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

	specs := make([]teaching.InsertInstrumentSpec, 0, len(req.Data))
	for _, param := range req.Data {
		specs = append(specs, teaching.InsertInstrumentSpec{
			Name: param.Name,
		})
	}

	instrumentIDs, err := s.teachingService.InsertInstruments(ctx, specs)
	if err != nil {
		return nil, handleUpsertionError(err, "teachingService.InsertInstruments()", "instrument")
	}
	mainLog.Info("Instruments created: instrumentIDs='%v'", instrumentIDs)

	instruments, err := s.teachingService.GetInstrumentsByIds(ctx, instrumentIDs)
	if err != nil {
		return nil, errs.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("teachingService.GetInstrumentsByIds: %v", err), nil, "")
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

	specs := make([]teaching.UpdateInstrumentSpec, 0, len(req.Data))
	for _, param := range req.Data {
		specs = append(specs, teaching.UpdateInstrumentSpec{
			ID:   param.ID,
			Name: param.Name,
		})
	}

	instrumentIDs, err := s.teachingService.UpdateInstruments(ctx, specs)
	if err != nil {
		return nil, handleUpsertionError(err, "teachingService.UpdateInstruments()", "instrument")
	}
	mainLog.Info("Instruments updated: instrumentIDs='%v'", instrumentIDs)

	instruments, err := s.teachingService.GetInstrumentsByIds(ctx, instrumentIDs)
	if err != nil {
		return nil, errs.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("teachingService.GetInstrumentsByIds: %v", err), nil, "")
	}

	return &output.UpdateInstrumentsResponse{
		Data: output.UpsertInstrumentResult{
			Results: instruments,
		},
		Message: "Successfully updated instruments",
	}, nil
}

func (s *BackendService) GetGradesHandler(ctx context.Context, req *output.GetGradesRequest) (*output.GetGradesResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	getGradesResult, err := s.teachingService.GetGrades(ctx, util.PaginationSpec{
		Page:           req.Page,
		ResultsPerPage: req.ResultsPerPage,
	})
	if err != nil {
		return nil, errs.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("teachingService.GetGrades(): %w", err), nil, "Failed to get grades")
	}

	paginationResult := getGradesResult.PaginationResult

	return &output.GetGradesResponse{
		Data: output.GetGradesResult{
			Results: getGradesResult.Grades,
			PaginationResponse: output.PaginationResponse{
				TotalPages:   paginationResult.TotalPages,
				TotalResults: paginationResult.TotalResults,
				CurrentPage:  paginationResult.CurrentPage,
			},
		},
	}, nil
}

func (s *BackendService) GetGradeByIdHandler(ctx context.Context, req *output.GetGradeRequest) (*output.GetGradeResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	grade, err := s.teachingService.GetGradeById(ctx, req.ID)
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

	specs := make([]teaching.InsertGradeSpec, 0, len(req.Data))
	for _, param := range req.Data {
		specs = append(specs, teaching.InsertGradeSpec{
			Name: param.Name,
		})
	}

	gradeIDs, err := s.teachingService.InsertGrades(ctx, specs)
	if err != nil {
		return nil, handleUpsertionError(err, "teachingService.InsertGrades()", "grade")
	}
	mainLog.Info("Grades created: gradeIDs='%v'", gradeIDs)

	grades, err := s.teachingService.GetGradesByIds(ctx, gradeIDs)
	if err != nil {
		return nil, errs.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("teachingService.GetGradesByIds: %v", err), nil, "")
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

	specs := make([]teaching.UpdateGradeSpec, 0, len(req.Data))
	for _, param := range req.Data {
		specs = append(specs, teaching.UpdateGradeSpec{
			ID:   param.ID,
			Name: param.Name,
		})
	}

	gradeIDs, err := s.teachingService.UpdateGrades(ctx, specs)
	if err != nil {
		return nil, handleUpsertionError(err, "teachingService.UpdateGrades()", "grade")
	}
	mainLog.Info("Grades updated: gradeIDs='%v'", gradeIDs)

	grades, err := s.teachingService.GetGradesByIds(ctx, gradeIDs)
	if err != nil {
		return nil, errs.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("teachingService.GetGradesByIds: %v", err), nil, "")
	}

	return &output.UpdateGradesResponse{
		Data: output.UpsertGradeResult{
			Results: grades,
		},
		Message: "Successfully updated grades",
	}, nil
}

func (s *BackendService) GetCoursesHandler(ctx context.Context, req *output.GetCoursesRequest) (*output.GetCoursesResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	getCoursesResult, err := s.teachingService.GetCourses(ctx, util.PaginationSpec{
		Page:           req.Page,
		ResultsPerPage: req.ResultsPerPage,
	})
	if err != nil {
		return nil, errs.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("teachingService.GetCourses(): %w", err), nil, "Failed to get courses")
	}

	paginationResult := getCoursesResult.PaginationResult

	return &output.GetCoursesResponse{
		Data: output.GetCoursesResult{
			Results: getCoursesResult.Courses,
			PaginationResponse: output.PaginationResponse{
				TotalPages:   paginationResult.TotalPages,
				TotalResults: paginationResult.TotalResults,
				CurrentPage:  paginationResult.CurrentPage,
			},
		},
	}, nil
}

func (s *BackendService) GetCourseByIdHandler(ctx context.Context, req *output.GetCourseRequest) (*output.GetCourseResponse, errs.HTTPError) {
	if errV := errs.ValidateHTTPRequest(req, false); errV != nil {
		return nil, errV
	}

	course, err := s.teachingService.GetCourseById(ctx, req.ID)
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

	specs := make([]teaching.InsertCourseSpec, 0, len(req.Data))
	for _, param := range req.Data {
		specs = append(specs, teaching.InsertCourseSpec{
			InstrumentID:          param.InstrumentID,
			GradeID:               param.GradeID,
			DefaultFee:            param.DefaultFee,
			DefaultDurationMinute: param.DefaultDurationMinute,
		})
	}

	courseIDs, err := s.teachingService.InsertCourses(ctx, specs)
	if err != nil {
		return nil, handleUpsertionError(err, "teachingService.InsertCourses()", "course")
	}
	mainLog.Info("Courses created: courseIDs='%v'", courseIDs)

	courses, err := s.teachingService.GetCoursesByIds(ctx, courseIDs)
	if err != nil {
		return nil, errs.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("teachingService.GetCoursesByIds: %v", err), nil, "")
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

	specs := make([]teaching.UpdateCourseSpec, 0, len(req.Data))
	for _, param := range req.Data {
		specs = append(specs, teaching.UpdateCourseSpec{
			ID:                    param.ID,
			DefaultFee:            param.DefaultFee,
			DefaultDurationMinute: param.DefaultDurationMinute,
		})
	}

	courseIDs, err := s.teachingService.UpdateCourses(ctx, specs)
	if err != nil {
		return nil, handleUpsertionError(err, "teachingService.UpdateCourses()", "course")
	}
	mainLog.Info("Courses updated: courseIDs='%v'", courseIDs)

	courses, err := s.teachingService.GetCoursesByIds(ctx, courseIDs)
	if err != nil {
		return nil, errs.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("teachingService.GetCoursesByIds: %v", err), nil, "")
	}

	return &output.UpdateCoursesResponse{
		Data: output.UpsertCourseResult{
			Results: courses,
		},
		Message: "Successfully updated courses",
	}, nil
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
		return errs.NewHTTPError(http.StatusConflict, fmt.Errorf("%s: %v", methodName, validationErr), validationErr.GetErrorDetail(), fmt.Sprintf("Invalid %s properties", entityName))
	}
	return errs.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("%s: %v", methodName, err), nil, fmt.Sprintf("Failed to create %s", entityName))
}
