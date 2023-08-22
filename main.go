package main

import (
	"fmt"
	"net/http"
	"os"
	"syscall"

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

func handleSignal(signalChan chan os.Signal, handlers []func()) {
	sig := <-signalChan
	switch sig {
	case syscall.SIGINT:
		fmt.Println("SIGINT signal is received")
		for _, fn := range handlers {
			fn()
		}
		break
	case syscall.SIGTERM:
		fmt.Println("SIGINT signal is received")
		for _, fn := range handlers {
			fn()
		}
	}
}

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

	// Router group for admin-only endpoints
	baseRouter.Route("/admin", func(authRouter chi.Router) {
		authRouter.Use(backendService.AuthenticationMiddleware)
		authRouter.Use(backendService.AuthorizationMiddleware(identity.UserPrivilegeType_Admin))

		authRouter.Get("/users", jsonSerdeWrapper.WrapFunc(backendService.GetUsersHandler))
		authRouter.Get("/user/{UserID}", jsonSerdeWrapper.WrapFunc(backendService.GetUserByIdHandler, "UserID"))
		authRouter.Post("/users", jsonSerdeWrapper.WrapFunc(backendService.InsertUsersHandler))
		authRouter.Put("/users", jsonSerdeWrapper.WrapFunc(backendService.UpdateUsersHandler))
		authRouter.Put("/user/{UserID}/password", jsonSerdeWrapper.WrapFunc(backendService.UpdateUserPasswordHandler, "UserID"))

		authRouter.Get("/teachers", jsonSerdeWrapper.WrapFunc(backendService.GetTeachersHandler))
		authRouter.Get("/teacher/{TeacherID}", jsonSerdeWrapper.WrapFunc(backendService.GetTeacherByIdHandler, "TeacherID"))
		authRouter.Post("/teachers", jsonSerdeWrapper.WrapFunc(backendService.InsertTeachersHandler))
		authRouter.Post("/teachers/new-users", jsonSerdeWrapper.WrapFunc(backendService.InsertTeachersWithNewUsersHandler))
		authRouter.Delete("/teachers", jsonSerdeWrapper.WrapFunc(backendService.DeleteTeachersHandler))

		authRouter.Get("/students", jsonSerdeWrapper.WrapFunc(backendService.GetStudentsHandler))
		authRouter.Get("/student/{StudentID}", jsonSerdeWrapper.WrapFunc(backendService.GetStudentByIdHandler, "StudentID"))
		authRouter.Post("/students", jsonSerdeWrapper.WrapFunc(backendService.InsertStudentsHandler))
		authRouter.Post("/students/new-users", jsonSerdeWrapper.WrapFunc(backendService.InsertStudentsWithNewUsersHandler))
		authRouter.Delete("/students", jsonSerdeWrapper.WrapFunc(backendService.DeleteStudentsHandler))

		authRouter.Get("/instruments", jsonSerdeWrapper.WrapFunc(backendService.GetInstrumentsHandler))
		authRouter.Get("/instrument/{InstrumentID}", jsonSerdeWrapper.WrapFunc(backendService.GetInstrumentByIdHandler, "InstrumentID"))
		authRouter.Post("/instruments", jsonSerdeWrapper.WrapFunc(backendService.InsertInstrumentsHandler))
		authRouter.Put("/instruments", jsonSerdeWrapper.WrapFunc(backendService.UpdateInstrumentsHandler))
		authRouter.Delete("/instruments", jsonSerdeWrapper.WrapFunc(backendService.DeleteInstrumentsHandler))

		authRouter.Get("/grades", jsonSerdeWrapper.WrapFunc(backendService.GetGradesHandler))
		authRouter.Get("/grade/{GradeID}", jsonSerdeWrapper.WrapFunc(backendService.GetGradeByIdHandler, "GradeID"))
		authRouter.Post("/grades", jsonSerdeWrapper.WrapFunc(backendService.InsertGradesHandler))
		authRouter.Put("/grades", jsonSerdeWrapper.WrapFunc(backendService.UpdateGradesHandler))
		authRouter.Delete("/grades", jsonSerdeWrapper.WrapFunc(backendService.DeleteGradesHandler))

		authRouter.Get("/courses", jsonSerdeWrapper.WrapFunc(backendService.GetCoursesHandler))
		authRouter.Get("/course/{CourseID}", jsonSerdeWrapper.WrapFunc(backendService.GetCourseByIdHandler, "CourseID"))
		authRouter.Post("/courses", jsonSerdeWrapper.WrapFunc(backendService.InsertCoursesHandler))
		authRouter.Put("/courses", jsonSerdeWrapper.WrapFunc(backendService.UpdateCoursesHandler))
		authRouter.Delete("/courses", jsonSerdeWrapper.WrapFunc(backendService.DeleteCoursesHandler))

		authRouter.Get("/classes", jsonSerdeWrapper.WrapFunc(backendService.GetClassesHandler))
		authRouter.Get("/class/{ClassID}", jsonSerdeWrapper.WrapFunc(backendService.GetClassByIdHandler, "ClassID"))
		authRouter.Post("/classes", jsonSerdeWrapper.WrapFunc(backendService.InsertClassesHandler))
		authRouter.Put("/classes", jsonSerdeWrapper.WrapFunc(backendService.UpdateClassesHandler))
		authRouter.Delete("/classes", jsonSerdeWrapper.WrapFunc(backendService.DeleteClassesHandler))

		authRouter.Get("/studentEnrollments", jsonSerdeWrapper.WrapFunc(backendService.GetStudentEnrollmentsHandler))

		authRouter.Get("/teacherSpecialFees", jsonSerdeWrapper.WrapFunc(backendService.GetTeacherSpecialFeesHandler))
		authRouter.Get("/teacherSpecialFee/{TeacherSpecialFeeID}", jsonSerdeWrapper.WrapFunc(backendService.GetTeacherSpecialFeeByIdHandler, "TeacherSpecialFeeID"))
		authRouter.Post("/teacherSpecialFees", jsonSerdeWrapper.WrapFunc(backendService.InsertTeacherSpecialFeesHandler))
		authRouter.Put("/teacherSpecialFees", jsonSerdeWrapper.WrapFunc(backendService.UpdateTeacherSpecialFeesHandler))
		authRouter.Delete("/teacherSpecialFees", jsonSerdeWrapper.WrapFunc(backendService.DeleteTeacherSpecialFeesHandler))

		authRouter.Get("/enrollmentPayments", jsonSerdeWrapper.WrapFunc(backendService.GetEnrollmentPaymentsHandler))
		authRouter.Get("/enrollmentPayment/{EnrollmentPaymentID}", jsonSerdeWrapper.WrapFunc(backendService.GetEnrollmentPaymentByIdHandler, "EnrollmentPaymentID"))
		authRouter.Post("/enrollmentPayments", jsonSerdeWrapper.WrapFunc(backendService.InsertEnrollmentPaymentsHandler))
		authRouter.Put("/enrollmentPayments", jsonSerdeWrapper.WrapFunc(backendService.UpdateEnrollmentPaymentsHandler))
		authRouter.Delete("/enrollmentPayments", jsonSerdeWrapper.WrapFunc(backendService.DeleteEnrollmentPaymentsHandler))

		authRouter.Get("/studentLearningTokens", jsonSerdeWrapper.WrapFunc(backendService.GetStudentLearningTokensHandler))
		authRouter.Get("/studentLearningToken/{StudentLearningTokenID}", jsonSerdeWrapper.WrapFunc(backendService.GetStudentLearningTokenByIdHandler, "StudentLearningTokenID"))
		authRouter.Post("/studentLearningTokens", jsonSerdeWrapper.WrapFunc(backendService.InsertStudentLearningTokensHandler))
		authRouter.Put("/studentLearningTokens", jsonSerdeWrapper.WrapFunc(backendService.UpdateStudentLearningTokensHandler))
		authRouter.Delete("/studentLearningTokens", jsonSerdeWrapper.WrapFunc(backendService.DeleteStudentLearningTokensHandler))

		authRouter.Get("/presences", jsonSerdeWrapper.WrapFunc(backendService.GetPresencesHandler))
		authRouter.Get("/presence/{PresenceID}", jsonSerdeWrapper.WrapFunc(backendService.GetPresenceByIdHandler, "PresenceID"))
		authRouter.Post("/presences", jsonSerdeWrapper.WrapFunc(backendService.InsertPresencesHandler))
		authRouter.Put("/presences", jsonSerdeWrapper.WrapFunc(backendService.UpdatePresencesHandler))
		authRouter.Delete("/presences", jsonSerdeWrapper.WrapFunc(backendService.DeletePresencesHandler))
	})

	// Router group for staff-only (and above) endpoints
	baseRouter.Group(func(authRouter chi.Router) {
		authRouter.Use(backendService.AuthenticationMiddleware)
		authRouter.Use(backendService.AuthorizationMiddleware(identity.UserPrivilegeType_Staff))

		authRouter.Get("/students", jsonSerdeWrapper.WrapFunc(backendService.GetStudentsHandler))
		authRouter.Get("/teachers", jsonSerdeWrapper.WrapFunc(backendService.GetTeachersHandler))
		authRouter.Get("/courses", jsonSerdeWrapper.WrapFunc(backendService.GetCoursesHandler))
		authRouter.Get("/studentEnrollments", jsonSerdeWrapper.WrapFunc(backendService.GetStudentEnrollmentsHandler))

		authRouter.Get("/enrollmentPayment/search", jsonSerdeWrapper.WrapFunc(backendService.SearchEnrollmentPayment))
		authRouter.Get("/enrollmentPayment/invoice/{StudentEnrollmentID}", jsonSerdeWrapper.WrapFunc(backendService.GetEnrollmentPaymentInvoice, "StudentEnrollmentID"))
		authRouter.Post("/enrollmentPayment/submit", jsonSerdeWrapper.WrapFunc(backendService.SubmitEnrollmentPayment))
		authRouter.Post("/enrollmentPayment/edit", jsonSerdeWrapper.WrapFunc(backendService.EditEnrollmentPayment))
		authRouter.Post("/enrollmentPayment/remove", jsonSerdeWrapper.WrapFunc(backendService.RemoveEnrollmentPayment))

		authRouter.Get("/class/search", jsonSerdeWrapper.WrapFunc(backendService.SearchClass))
		authRouter.Get("/presences/{ClassID}", jsonSerdeWrapper.WrapFunc(backendService.GetPresencesByClassID, "ClassID"))
	})

	serverAddr := fmt.Sprintf("%s:%s", configObject.Host, configObject.Port)
	server := &http.Server{
		Addr:           serverAddr,
		Handler:        baseRouter,
		ReadTimeout:    2 * configObject.ServerTimeout, // we use 2 times ServerTimeout just for extra layer of timeout assurance
		WriteTimeout:   2 * configObject.ServerTimeout,
		MaxHeaderBytes: 1 << 20,
	}

	logging.AppLogger.Info("Server is starting...")
	logging.AppLogger.Info("Serving on %s", serverAddr)
	logging.AppLogger.Error("err: %s", server.ListenAndServe())
}
