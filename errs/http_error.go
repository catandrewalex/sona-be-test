package errs

const (
	ClientMessageKey_NonField = "non-field"
)

type HTTPError interface {
	Error() string

	GetProcessableErrors() map[string]string
	GetClientMessage() string
	GetHTTPErrorCode() int
}

type httpError struct {
	err error
	// processableErrors are errors that are expected to be processable by FE. E.g.: field-related errors to be rendered under every FE text boxes.
	processableErrors map[string]string
	// clientMessage stores a user-friendly message (i.e. an error) that is expected to be directly rendered by the FE.
	clientMessage string
	httpErrorCode int
}

func NewHTTPError(httpErrorCode int, err error, processableErrors map[string]string, clientMessage string) HTTPError {
	return httpError{
		err:               err,
		processableErrors: processableErrors,
		clientMessage:     clientMessage,
		httpErrorCode:     httpErrorCode,
	}
}

func (e httpError) Error() string {
	return e.err.Error()
}

func (e httpError) GetProcessableErrors() map[string]string {
	if e.processableErrors == nil {
		return map[string]string{}
	}
	return e.processableErrors
}

func (e httpError) GetClientMessage() string {
	return e.clientMessage
}

func (e httpError) GetHTTPErrorCode() int {
	return e.httpErrorCode
}
