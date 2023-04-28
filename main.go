package main

import (
	"fmt"
	"net/http"
	"os"
	"syscall"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

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

	// Router group for authenticated endpoints
	baseRouter.Group(func(authRouter chi.Router) {
		authRouter.Use(backendService.AuthenticationMiddleware)
		authRouter.Post("/user-profile", jsonSerdeWrapper.WrapFunc(backendService.UserProfileHandler))
	})

	baseRouter.Get("/", backendService.HomepageHandler)
	baseRouter.Get("/get-jwt", backendService.GetJWTHandler)
	baseRouter.Post("/sign-up", jsonSerdeWrapper.WrapFunc(backendService.SignUpHandler))
	baseRouter.Post("/login", jsonSerdeWrapper.WrapFunc(backendService.LoginHandler))
	baseRouter.Post("/forgot-password", jsonSerdeWrapper.WrapFunc(backendService.ForgotPasswordHandler))
	baseRouter.Post("/reset-password", jsonSerdeWrapper.WrapFunc(backendService.ResetPasswordHandler))

	server := &http.Server{
		Addr:           fmt.Sprintf(":%s", configObject.Port),
		Handler:        baseRouter,
		ReadTimeout:    2 * configObject.ServerTimeout, // we use 2 times ServerTimeout just for extra layer of timeout assurance
		WriteTimeout:   2 * configObject.ServerTimeout,
		MaxHeaderBytes: 1 << 20,
	}

	logging.AppLogger.Info("Server is starting...")
	logging.AppLogger.Error("err: %s", server.ListenAndServe())
}
