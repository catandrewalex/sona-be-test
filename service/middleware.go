package service

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"sonamusica-backend/app-service/auth"
	"sonamusica-backend/app-service/identity"
	"sonamusica-backend/logging"
	"sonamusica-backend/network"
)

// AuthenticationMiddleware is a middleware to authenticate request using JWT
func (s *BackendService) AuthenticationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")
		if tokenString == "" {
			http.Error(w, "Empty authentication token", http.StatusUnauthorized)
			return
		}

		var prefix = "Bearer"
		if len(tokenString) < len(prefix) || tokenString[:len(prefix)] != prefix {
			http.Error(w, "Token must start with 'Bearer'", http.StatusUnauthorized)
			return
		}

		tokenStringWithoutPrefix := strings.Trim(tokenString[len(prefix):], " ")

		claims, err := s.jwtService.VerifyTokenStringAndReturnClaims(tokenStringWithoutPrefix)
		if err != nil {
			logging.HTTPServerLogger.Error(fmt.Sprintf("Unable to verify authentication token: %v", err))
			// we can directly send the error message to client, as the JWT's error message doesn't leak any internal information
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		mainClaims := claims.(*auth.MainJWTClaims)

		if mainClaims.PurposeType != auth.JWTTokenPurposeType_Auth {
			http.Error(w, "invalid JWT token purpose", http.StatusUnauthorized)
			return
		}

		ctx := network.NewContextWithAuthInfo(r.Context(), network.AuthInfo{
			UserID:        mainClaims.UserID,
			PrivilegeType: mainClaims.PrivilegeType,
		})

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// AuthorizationMiddleware is an intermediate function to return a middleware to check user authority.
func (s *BackendService) AuthorizationMiddleware(minimalPrivilegeType identity.UserPrivilegeType) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			isValid, err := s.identityService.VerifyUserAuthority(r.Context(), minimalPrivilegeType)
			if err != nil {
				logging.HTTPServerLogger.Error(fmt.Sprintf("Unable to verify user: %v", err))
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

			if !isValid {
				http.Error(w, "user is not authorized to access this resources", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func RequestContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestContext := network.CreateRequestContext(r)

		ctx := r.Context()
		ctx = network.NewContextWithRequestContext(ctx, requestContext)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// LoggingMiddleware is a middleware function for logging all request information
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		ww := responseWriterWithStatus{w, http.StatusOK} // initialize with status = 200 (OK), in case the status is not filled
		next.ServeHTTP(&ww, r)

		reqID := network.GetRequestID(r)
		scheme := "http"
		if r.TLS != nil {
			scheme = "https"
		}

		logging.HTTPServerLogger.Info(
			"[%s] \"%s %s://%s%s %s\" from %s - %d %0.3f ms",
			reqID,
			r.Method,
			scheme,
			r.Host,
			r.RequestURI,
			r.Proto,
			r.RemoteAddr,
			ww.status,
			float64(time.Since(start).Microseconds()/1000),
		)
	})
}

// responseWriterWithStatus is required to let us access the HTTP status code
// We can reuse go-chi's NewWrapResponseWriter if we need more information (i.e. bytes written, header, etc.)
type responseWriterWithStatus struct {
	http.ResponseWriter
	status int
}

func (rec *responseWriterWithStatus) WriteHeader(code int) {
	rec.status = code
	rec.ResponseWriter.WriteHeader(code)
}
