package service

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"sonamusica-backend/app-service/auth"
	"sonamusica-backend/app-service/identity"
	"sonamusica-backend/app-service/user_action_log"
	"sonamusica-backend/logging"
	"sonamusica-backend/network"
)

// AuthenticationMiddleware is a middleware to authenticate request using JWT.
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

// AuthorizationMiddleware is an intermediate function to return a middleware for checking user authority.
func (s *BackendService) AuthorizationMiddleware(minimalPrivilegeType identity.UserPrivilegeType) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			isValid, err := s.identityService.VerifyUserAuthority(r.Context(), minimalPrivilegeType)
			if err != nil {
				logging.HTTPServerLogger.Error(fmt.Sprintf("unable to verify user: %v", err))
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

// LoggingMiddleware is a middleware function for logging all request information.
// This middlware also injects a custom ResponseWriter into the context, which can be used to fetch status code.
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		ww := network.ResponseWriterWithStatus{
			ResponseWriter: w,
			Status:         http.StatusOK,
		} // initialize with status = 200 (OK), in case the status is not filled
		ctx := r.Context()
		ctx = network.NewContextWithStatusCodeWriter(ctx, &ww)

		next.ServeHTTP(&ww, r.WithContext(ctx))

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
			ww.Status,
			float64(time.Since(start).Microseconds()/1000),
		)
	})
}

// UserActionLogMiddleware logs request information from non "GET" method.
//
// IMPORTANT NOTES:
//
// Request body is also logged, so make sure NOT to use this middleware on endpoints that handle SENSITIVE INFORMATION (e.g. credentialS).
func (s *BackendService) UserActionLogMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// ignore the error, as the error handling of the request body is not this middleware's responsibility
		requestBody, _ := io.ReadAll(r.Body)
		if !utf8.Valid(requestBody) {
			requestBody = bytes.ToValidUTF8(requestBody, []byte("?"))
			logging.HTTPServerLogger.Warn("UserActionLogMiddleware found invalid utf-8 characters on the request body. replacing invalid characters with '?'.")
		}
		requestBodyStr := string(requestBody)

		// Replace the body with a new reader after reading from the original. To allow the real handler to read the body.
		r.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		next.ServeHTTP(w, r)

		ctx := r.Context()
		authInfo := network.GetAuthInfo(ctx)
		writerWithStatus := network.GetStatusCodeWriter(ctx)
		if r.Method != "GET" {
			_, err := s.userActionLogService.InsertUserActionLogs(ctx, []user_action_log.InsertUserActionLogSpec{
				{
					Date:          time.Now(),
					UserID:        authInfo.UserID,
					PrivilegeType: authInfo.PrivilegeType,
					Endpoint:      r.URL.Path,
					Method:        r.Method,
					StatusCode:    uint16(writerWithStatus.Status),
					RequestBody:   requestBodyStr,
				},
			})
			if err != nil {
				logging.HTTPServerLogger.Error("unable to log user action: %v", err)
			}
		}
	})
}
