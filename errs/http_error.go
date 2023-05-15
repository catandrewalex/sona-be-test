package errs

const (
	ClientMessageKey_NonField = "non-field"
)

type HTTPError interface {
	Error() string

	GetClientMessages() map[string]string
	GetHTTPErrorCode() int
}

type httpError struct {
	err            error
	clientMessages map[string]string
	httpErrorCode  int
}

func NewHTTPError(httpErrorCode int, err error, clientMessages map[string]string) HTTPError {
	return httpError{
		err:            err,
		clientMessages: clientMessages,
		httpErrorCode:  httpErrorCode,
	}
}

func (e httpError) Error() string {
	return e.err.Error()
}

func (e httpError) GetClientMessages() map[string]string {
	return e.clientMessages
}

func (e httpError) GetHTTPErrorCode() int {
	return e.httpErrorCode
}
