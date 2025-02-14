package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"sonamusica-backend/app-service/identity"
	"sonamusica-backend/config"
	"sonamusica-backend/logging"
	"sonamusica-backend/service"
	"sonamusica-backend/service/serde_wrapper"
)

var (
	configObject = config.Get()
)

func main() {
	backendService := service.NewBackendService()

	baseRouter := chi.NewRouter()
	jsonSerdeWrapper := serde_wrapper.NewJSONSerdeWrapper()

	baseRouter.Use(
		cors.Handler(cors.Options{
			AllowedOrigins:   []string{"*"},
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-Api-Version"},
			ExposedHeaders:   []string{"Link"},
			AllowCredentials: true,
			MaxAge:           300, // Maximum value not ignored by any of major browsers
		}),
		middleware.RequestID,
		middleware.RealIP,
		service.LoggingMiddleware,
		service.RequestContextMiddleware,
		middleware.Recoverer,

		// Set a timeout value on the request context (ctx), that will signal
		// through ctx.Done() that the request has timed out and further
		// processing should be stopped.
		middleware.Timeout(configObject.ServerTimeout),
	)

	baseRouter.Get("/", backendService.HomepageHandler)
	baseRouter.Get("/get-jwt", backendService.GetJWTHandler)
	baseRouter.Post("/sign-up", jsonSerdeWrapper.WrapFunc(backendService.SignUpHandler))
	baseRouter.Post("/login", jsonSerdeWrapper.WrapFunc(backendService.LoginHandler))
	baseRouter.Post("/forgot-password", jsonSerdeWrapper.WrapFunc(backendService.ForgotPasswordHandler))
	baseRouter.Post("/reset-password", jsonSerdeWrapper.WrapFunc(backendService.ResetPasswordHandler))

	baseRouter.Get("/maintenance/user-action-logs/page", backendService.UserActionLogsPage)
	// Router group for superAdmin-only endpoints
	baseRouter.Route("/maintenance", func(authRouter chi.Router) {
		authRouter.Use(backendService.AuthenticationMiddleware)
		authRouter.Use(backendService.AuthorizationMiddleware(identity.UserPrivilegeType_Super_Admin))
		authRouter.Use(backendService.UserActionLogMiddleware)

		authRouter.Get("/user-action-logs/fetch", jsonSerdeWrapper.WrapFunc(backendService.FetchUserActionLogs))
	})

	// Router group for admin-only endpoints
	baseRouter.Route("/admin", func(authRouter chi.Router) {
		authRouter.Use(backendService.AuthenticationMiddleware)
		authRouter.Use(backendService.AuthorizationMiddleware(identity.UserPrivilegeType_Admin))
		authRouter.Use(backendService.UserActionLogMiddleware)

		authRouter.Get("/users", jsonSerdeWrapper.WrapFunc(backendService.GetUsersHandler))
		authRouter.Get("/users/{UserID}", jsonSerdeWrapper.WrapFunc(backendService.GetUserByIdHandler, "UserID"))
		authRouter.Post("/users", jsonSerdeWrapper.WrapFunc(backendService.InsertUsersHandler))
		authRouter.Put("/users", jsonSerdeWrapper.WrapFunc(backendService.UpdateUsersHandler))
		// This endpoint is only used internally as an administrative tool (to update previously submitted data, without needing to know the UserID).
		// This endpoint won't be used in frontend.
		authRouter.Put("/users/by-usernames", jsonSerdeWrapper.WrapFunc(backendService.UpdateUsersByUsernamesHandler))

		authRouter.Get("/teachers", jsonSerdeWrapper.WrapFunc(backendService.GetTeachersHandler))
		authRouter.Get("/teachers/{TeacherID}", jsonSerdeWrapper.WrapFunc(backendService.GetTeacherByIdHandler, "TeacherID"))
		authRouter.Post("/teachers", jsonSerdeWrapper.WrapFunc(backendService.InsertTeachersHandler))
		authRouter.Post("/teachers/new-users", jsonSerdeWrapper.WrapFunc(backendService.InsertTeachersWithNewUsersHandler))
		authRouter.Delete("/teachers", jsonSerdeWrapper.WrapFunc(backendService.DeleteTeachersHandler))

		authRouter.Get("/students", jsonSerdeWrapper.WrapFunc(backendService.GetStudentsHandler))
		authRouter.Get("/students/{StudentID}", jsonSerdeWrapper.WrapFunc(backendService.GetStudentByIdHandler, "StudentID"))
		authRouter.Post("/students", jsonSerdeWrapper.WrapFunc(backendService.InsertStudentsHandler))
		authRouter.Post("/students/new-users", jsonSerdeWrapper.WrapFunc(backendService.InsertStudentsWithNewUsersHandler))
		authRouter.Delete("/students", jsonSerdeWrapper.WrapFunc(backendService.DeleteStudentsHandler))

		authRouter.Get("/instruments", jsonSerdeWrapper.WrapFunc(backendService.GetInstrumentsHandler))
		authRouter.Get("/instruments/{InstrumentID}", jsonSerdeWrapper.WrapFunc(backendService.GetInstrumentByIdHandler, "InstrumentID"))
		authRouter.Post("/instruments", jsonSerdeWrapper.WrapFunc(backendService.InsertInstrumentsHandler))
		authRouter.Put("/instruments", jsonSerdeWrapper.WrapFunc(backendService.UpdateInstrumentsHandler))
		authRouter.Delete("/instruments", jsonSerdeWrapper.WrapFunc(backendService.DeleteInstrumentsHandler))

		authRouter.Get("/grades", jsonSerdeWrapper.WrapFunc(backendService.GetGradesHandler))
		authRouter.Get("/grades/{GradeID}", jsonSerdeWrapper.WrapFunc(backendService.GetGradeByIdHandler, "GradeID"))
		authRouter.Post("/grades", jsonSerdeWrapper.WrapFunc(backendService.InsertGradesHandler))
		authRouter.Put("/grades", jsonSerdeWrapper.WrapFunc(backendService.UpdateGradesHandler))
		authRouter.Delete("/grades", jsonSerdeWrapper.WrapFunc(backendService.DeleteGradesHandler))

		authRouter.Get("/courses", jsonSerdeWrapper.WrapFunc(backendService.GetCoursesHandler))
		authRouter.Get("/courses/{CourseID}", jsonSerdeWrapper.WrapFunc(backendService.GetCourseByIdHandler, "CourseID"))
		authRouter.Post("/courses", jsonSerdeWrapper.WrapFunc(backendService.InsertCoursesHandler))
		authRouter.Put("/courses", jsonSerdeWrapper.WrapFunc(backendService.UpdateCoursesHandler))
		authRouter.Delete("/courses", jsonSerdeWrapper.WrapFunc(backendService.DeleteCoursesHandler))

		authRouter.Get("/classes", jsonSerdeWrapper.WrapFunc(backendService.GetClassesHandler))
		authRouter.Get("/classes/{ClassID}", jsonSerdeWrapper.WrapFunc(backendService.GetClassByIdHandler, "ClassID"))
		authRouter.Post("/classes", jsonSerdeWrapper.WrapFunc(backendService.InsertClassesHandler))
		authRouter.Put("/classes", jsonSerdeWrapper.WrapFunc(backendService.UpdateClassesHandler))
		authRouter.Delete("/classes", jsonSerdeWrapper.WrapFunc(backendService.DeleteClassesHandler))

		authRouter.Get("/studentEnrollments", jsonSerdeWrapper.WrapFunc(backendService.GetStudentEnrollmentsHandler))

		authRouter.Get("/teacherSpecialFees", jsonSerdeWrapper.WrapFunc(backendService.GetTeacherSpecialFeesHandler))
		authRouter.Get("/teacherSpecialFees/{TeacherSpecialFeeID}", jsonSerdeWrapper.WrapFunc(backendService.GetTeacherSpecialFeeByIdHandler, "TeacherSpecialFeeID"))
		authRouter.Post("/teacherSpecialFees", jsonSerdeWrapper.WrapFunc(backendService.InsertTeacherSpecialFeesHandler))
		authRouter.Put("/teacherSpecialFees", jsonSerdeWrapper.WrapFunc(backendService.UpdateTeacherSpecialFeesHandler))
		authRouter.Delete("/teacherSpecialFees", jsonSerdeWrapper.WrapFunc(backendService.DeleteTeacherSpecialFeesHandler))

		authRouter.Get("/enrollmentPayments", jsonSerdeWrapper.WrapFunc(backendService.GetEnrollmentPaymentsHandler))
		authRouter.Get("/enrollmentPayments/{EnrollmentPaymentID}", jsonSerdeWrapper.WrapFunc(backendService.GetEnrollmentPaymentByIdHandler, "EnrollmentPaymentID"))
		authRouter.Post("/enrollmentPayments", jsonSerdeWrapper.WrapFunc(backendService.InsertEnrollmentPaymentsHandler))
		authRouter.Put("/enrollmentPayments", jsonSerdeWrapper.WrapFunc(backendService.UpdateEnrollmentPaymentsHandler))
		authRouter.Delete("/enrollmentPayments", jsonSerdeWrapper.WrapFunc(backendService.DeleteEnrollmentPaymentsHandler))

		authRouter.Get("/studentLearningTokens", jsonSerdeWrapper.WrapFunc(backendService.GetStudentLearningTokensHandler))
		authRouter.Get("/studentLearningTokens/{StudentLearningTokenID}", jsonSerdeWrapper.WrapFunc(backendService.GetStudentLearningTokenByIdHandler, "StudentLearningTokenID"))
		authRouter.Post("/studentLearningTokens", jsonSerdeWrapper.WrapFunc(backendService.InsertStudentLearningTokensHandler))
		authRouter.Put("/studentLearningTokens", jsonSerdeWrapper.WrapFunc(backendService.UpdateStudentLearningTokensHandler))
		authRouter.Delete("/studentLearningTokens", jsonSerdeWrapper.WrapFunc(backendService.DeleteStudentLearningTokensHandler))

		authRouter.Get("/attendances", jsonSerdeWrapper.WrapFunc(backendService.GetAttendancesHandler))
		authRouter.Get("/attendances/{AttendanceID}", jsonSerdeWrapper.WrapFunc(backendService.GetAttendanceByIdHandler, "AttendanceID"))
		authRouter.Post("/attendances", jsonSerdeWrapper.WrapFunc(backendService.InsertAttendancesHandler))
		authRouter.Put("/attendances", jsonSerdeWrapper.WrapFunc(backendService.UpdateAttendancesHandler))
		authRouter.Delete("/attendances", jsonSerdeWrapper.WrapFunc(backendService.DeleteAttendancesHandler))

		authRouter.Get("/teacherPayments", jsonSerdeWrapper.WrapFunc(backendService.GetTeacherPaymentsHandler))
	})

	// Router group for staff-only (and above) endpoints
	baseRouter.Group(func(authRouter chi.Router) {
		authRouter.Use(backendService.AuthenticationMiddleware)
		authRouter.Use(backendService.AuthorizationMiddleware(identity.UserPrivilegeType_Staff))

		// non-GET requests (user actions) are logged
		authRouter.Group(func(loggedRouter chi.Router) {
			loggedRouter.Use(backendService.UserActionLogMiddleware)

			loggedRouter.Get("/students", jsonSerdeWrapper.WrapFunc(backendService.GetStudentsHandler))
			loggedRouter.Get("/teachers", jsonSerdeWrapper.WrapFunc(backendService.GetTeachersHandler))
			loggedRouter.Get("/courses", jsonSerdeWrapper.WrapFunc(backendService.GetCoursesHandler))
			loggedRouter.Get("/classes", jsonSerdeWrapper.WrapFunc(backendService.GetClassesHandler))
			loggedRouter.Get("/studentEnrollments", jsonSerdeWrapper.WrapFunc(backendService.GetStudentEnrollmentsHandler))
			loggedRouter.Get("/attendances", jsonSerdeWrapper.WrapFunc(backendService.GetAttendancesHandler))

			loggedRouter.Post("/classes/edit/config", jsonSerdeWrapper.WrapFunc(backendService.EditClassesConfigsHandler))
			loggedRouter.Post("/classes/edit/course", jsonSerdeWrapper.WrapFunc(backendService.EditClassesCoursesHandler))

			loggedRouter.Get("/enrollmentPayments/search", jsonSerdeWrapper.WrapFunc(backendService.SearchEnrollmentPaymentHandler))
			loggedRouter.Get("/enrollmentPayments/invoice/studentEnrollment/{StudentEnrollmentID}", jsonSerdeWrapper.WrapFunc(backendService.GetEnrollmentPaymentInvoiceHandler, "StudentEnrollmentID"))
			loggedRouter.Post("/enrollmentPayments/submit", jsonSerdeWrapper.WrapFunc(backendService.SubmitEnrollmentPaymentHandler))
			loggedRouter.Post("/enrollmentPayments/edit", jsonSerdeWrapper.WrapFunc(backendService.EditEnrollmentPaymentHandler))
			loggedRouter.Post("/enrollmentPayments/remove", jsonSerdeWrapper.WrapFunc(backendService.RemoveEnrollmentPaymentHandler))

			loggedRouter.Get("/teacherPayments/unpaidTeachers", jsonSerdeWrapper.WrapFunc(backendService.GetUnpaidTeachersHandler))
			loggedRouter.Get("/teacherPayments/paidTeachers", jsonSerdeWrapper.WrapFunc(backendService.GetPaidTeachersHandler))
			// TODO: improve naming for these 2 similar endpoints, as both return TeacherPaymentInvoiceItem.
			// But, for getting existing TeacherPayment as TeacherPaymentInvoiceItem, the URL doesn't seem to represent it.
			loggedRouter.Get("/teacherPayments/invoiceItems/teacher/{TeacherID}", jsonSerdeWrapper.WrapFunc(backendService.GetTeacherPaymentInvoiceItemsHandler, "TeacherID"))
			loggedRouter.Get("/teacherPayments/teacher/{TeacherID}", jsonSerdeWrapper.WrapFunc(backendService.GetTeacherPaymentsAsInvoiceItemsHandler, "TeacherID"))

			loggedRouter.Post("/teacherPayments/submit", jsonSerdeWrapper.WrapFunc(backendService.SubmitTeacherPaymentsHandler))
			loggedRouter.Post("/teacherPayments/modify", jsonSerdeWrapper.WrapFunc(backendService.ModifyTeacherPaymentsHandler))
			// TODO: delete these /edit & /remove. /modify already handles both of the operations
			loggedRouter.Post("/teacherPayments/edit", jsonSerdeWrapper.WrapFunc(backendService.EditTeacherPaymentsHandler))
			loggedRouter.Post("/teacherPayments/remove", jsonSerdeWrapper.WrapFunc(backendService.RemoveTeacherPaymentsHandler))

			loggedRouter.Post("/attendances/{AttendanceID}/assignToken", jsonSerdeWrapper.WrapFunc(backendService.AssignAttendanceTokenHandler, "AttendanceID"))
			// This endpoint is more similar with "/classes/{ClassID}/attendances/add", where (1) SLTs are automatically updated, (2) class with n students will get n attendances.
			// The goal is to simplify admin day-to-day work. Inputting in batch is simpler than navigating between pages and inserting the attendances one-by-one.
			loggedRouter.Post("/attendances/batch", jsonSerdeWrapper.WrapFunc(backendService.AddAttendancesBatchHandler))
		})

		authRouter.Post("/dashboard/expense/overview", jsonSerdeWrapper.WrapFunc(backendService.GetDashboardExpenseOverview))
		authRouter.Post("/dashboard/expense/monthlySummary", jsonSerdeWrapper.WrapFunc(backendService.GetDashboardExpenseMonthlySummary))
		authRouter.Post("/dashboard/income/overview", jsonSerdeWrapper.WrapFunc(backendService.GetDashboardIncomeOverview))
		authRouter.Post("/dashboard/income/monthlySummary", jsonSerdeWrapper.WrapFunc(backendService.GetDashboardIncomeMonthlySummary))
		// TODO: properly implement this, as we're reusing admin endpoint?
		authRouter.Get("/teachersForDashboard", jsonSerdeWrapper.WrapFunc(backendService.GetTeachersHandler))
		authRouter.Get("/instrumentsForDashboard", jsonSerdeWrapper.WrapFunc(backendService.GetInstrumentsHandler))
	})

	// Router group for member endpoints
	// without UserActionLog
	baseRouter.Group(func(authRouter chi.Router) {
		authRouter.Use(backendService.AuthenticationMiddleware)
		authRouter.Use(backendService.AuthorizationMiddleware(identity.UserPrivilegeType_Member))

		authRouter.Put("/users/{UserID}/password", jsonSerdeWrapper.WrapFunc(backendService.UpdateUserPasswordHandler, "UserID"))

		authRouter.Group(func(loggedRouter chi.Router) {
			loggedRouter.Use(backendService.UserActionLogMiddleware)

			loggedRouter.Get("/userProfile", jsonSerdeWrapper.WrapFunc(backendService.GetUserProfile))
			loggedRouter.Get("/userTeachingInfo", jsonSerdeWrapper.WrapFunc(backendService.GetUserTeachingInfo))

			// TODO: properly implement this, as we're reusing admin endpoint?
			loggedRouter.Get("/teachersForAttendance", jsonSerdeWrapper.WrapFunc(backendService.GetTeachersHandler))

			// these endpoints can be called by all type of members (except Anonymous). The privilege check is done inside the function.
			loggedRouter.Get("/classes/search", jsonSerdeWrapper.WrapFunc(backendService.SearchClass))
			loggedRouter.Get("/classes/{ClassID}", jsonSerdeWrapper.WrapFunc(backendService.GetClassByIdHandler, "ClassID"))
			loggedRouter.Get("/classes/{ClassID}/studentLearningTokensDisplay", jsonSerdeWrapper.WrapFunc(backendService.GetStudentLearningTokensByClassIDHandler, "ClassID"))
			loggedRouter.Get("/classes/{ClassID}/attendances", jsonSerdeWrapper.WrapFunc(backendService.GetAttendancesByClassIDHandler, "ClassID"))
			// for Member, only teachers are allowed to utilize these endpoints
			// TODO: implement the authorization.
			loggedRouter.Post("/classes/{ClassID}/attendances/add", jsonSerdeWrapper.WrapFunc(backendService.AddAttendanceHandler, "ClassID"))
			loggedRouter.Post("/attendances/{AttendanceID}/edit", jsonSerdeWrapper.WrapFunc(backendService.EditAttendanceHandler, "AttendanceID"))
			loggedRouter.Post("/attendances/{AttendanceID}/remove", jsonSerdeWrapper.WrapFunc(backendService.RemoveAttendanceHandler, "AttendanceID"))
		})
	})

	serverAddr := fmt.Sprintf("%s:%s", configObject.Host, configObject.Port)
	server := &http.Server{
		Addr:           serverAddr,
		Handler:        baseRouter,
		ReadTimeout:    2 * configObject.ServerTimeout, // we use 2 times ServerTimeout just for extra layer of timeout assurance
		WriteTimeout:   2 * configObject.ServerTimeout,
		MaxHeaderBytes: 1 << 20,
	}

	// Server run context
	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	// Listen for syscall signals for process to interrupt/quit
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sig

		// Shutdown signal with grace period of 30 seconds
		shutdownCtx, cancel := context.WithTimeout(serverCtx, 30*time.Second)
		defer cancel()

		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				logging.AppLogger.Fatal("graceful shutdown timed out.. forcing exit.")
			}
		}()

		// Trigger graceful shutdown
		err := server.Shutdown(shutdownCtx)
		if err != nil {
			logging.AppLogger.Fatal("graceful shutdown error: %v", err)
		}
		serverStopCtx()
	}()

	logging.AppLogger.Info("Server is starting...")
	logging.AppLogger.Info("Serving on %s", serverAddr)
	err := server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		logging.AppLogger.Fatal("Server error: %v", err)
	}

	<-serverCtx.Done()
}
