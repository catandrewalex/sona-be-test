package network

import (
	"context"
	"fmt"
	"net/http"
	"sonamusica-backend/app-service/identity"
	"sonamusica-backend/errs"

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

func GetRequestContext(ctx context.Context) (RequestContext, errs.HTTPError) {
	reqCtx, ok := ctx.Value(requestContextKey{}).(RequestContext)
	if !ok {
		return RequestContext{}, errs.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("non-existing context: RequestContext"), map[string]string{})
	}
	return reqCtx, nil
}

type userIDKey struct{}

func NewContextWithUserID(ctx context.Context, userID identity.UserID) context.Context {
	return context.WithValue(ctx, userIDKey{}, userID)
}

func GetUserID(ctx context.Context) identity.UserID {
	userID, ok := ctx.Value(userIDKey{}).(identity.UserID)
	if !ok {
		return identity.UserID_None
	}
	return userID
}
