package network

import (
	"context"
	"net/http"
	"sonamusica-backend/app-service/identity"
	"sonamusica-backend/logging"

	"github.com/mileusna/useragent"
)

type requestContextKey struct{}

type RequestContext struct {
	UserAgent useragent.UserAgent
	Origin    string
	IPAddress string
	RequestID string
}

func NewContextWithRequestContext(ctx context.Context, reqCtx RequestContext) context.Context {
	return context.WithValue(ctx, requestContextKey{}, reqCtx)
}

func CreateRequestContext(request *http.Request) RequestContext {
	requestContext := RequestContext{
		UserAgent: useragent.Parse(request.UserAgent()),
		Origin:    GetOrigin(request),
		IPAddress: GetIPAddress(request),
		RequestID: GetRequestID(request),
	}

	return requestContext
}

func GetRequestContext(ctx context.Context) RequestContext {
	reqCtx, ok := ctx.Value(requestContextKey{}).(RequestContext)
	if !ok {
		logging.AppLogger.Warn("non-existing context: RequestContext")
		return RequestContext{}
	}
	return reqCtx
}

type authInfoKey struct{}

type AuthInfo struct {
	UserID        identity.UserID
	PrivilegeType identity.UserPrivilegeType
}

func NewContextWithAuthInfo(ctx context.Context, authInfo AuthInfo) context.Context {
	return context.WithValue(ctx, authInfoKey{}, authInfo)
}

func GetAuthInfo(ctx context.Context) AuthInfo {
	authInfo, ok := ctx.Value(authInfoKey{}).(AuthInfo)
	if !ok {
		return AuthInfo{}
	}
	return authInfo
}

type statusCodeWriterKey struct{}

// ResponseWriterWithStatus is required to let us access the HTTP status code
// We can reuse go-chi's NewWrapResponseWriter if we need more information (i.e. bytes written, header, etc.)
type ResponseWriterWithStatus struct {
	http.ResponseWriter
	Status int
}

func (rec *ResponseWriterWithStatus) WriteHeader(code int) {
	rec.Status = code
	rec.ResponseWriter.WriteHeader(code)
}

func NewContextWithStatusCodeWriter(ctx context.Context, w *ResponseWriterWithStatus) context.Context {
	return context.WithValue(ctx, statusCodeWriterKey{}, w)
}

func GetStatusCodeWriter(ctx context.Context) *ResponseWriterWithStatus {
	writer, ok := ctx.Value(statusCodeWriterKey{}).(*ResponseWriterWithStatus)
	if !ok {
		return &ResponseWriterWithStatus{}
	}
	return writer
}
