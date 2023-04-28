package errs

import (
	"fmt"
	"net/http"
)

type HTTPError interface {
	Error() string

	GetClientMessage() string
	GetHTTPErrorCode() int
}

type httpError struct {
	err           error
	clientMessage string
	httpErrorCode int
}

func NewHTTPError(httpErrorCode int, err error, additionalClientMessage string) HTTPError {
	clientMessage := http.StatusText(httpErrorCode)
	if additionalClientMessage != "" {
		clientMessage = fmt.Sprintf("%s: %s", clientMessage, additionalClientMessage)
	}
	return httpError{
		err:           err,
		clientMessage: clientMessage,
		httpErrorCode: httpErrorCode,
	}
}

func (e httpError) Error() string {
	return e.err.Error()
}

func (e httpError) GetClientMessage() string {
	return e.clientMessage
}

func (e httpError) GetHTTPErrorCode() int {
	return e.httpErrorCode
}
