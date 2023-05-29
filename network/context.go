package network

import (
	"context"
	"database/sql"
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

type sqlTxKey struct{}

// NewContextWithSQLTx copies a context, adds a Go's sql.Tx into it, and returns the new context.
//
// TODO: remove this and look for alternative? as we're utilizing this as optional parameter.
// Go' documentation officially doesn't recommend doing it.
func NewContextWithSQLTx(ctx context.Context, tx *sql.Tx) context.Context {
	return context.WithValue(ctx, sqlTxKey{}, tx)
}

func GetSQLTx(ctx context.Context) *sql.Tx {
	sqlTx, ok := ctx.Value(sqlTxKey{}).(*sql.Tx)
	if !ok {
		return nil
	}
	return sqlTx
}
