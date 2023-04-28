package errs

import (
	"sonamusica-backend/logging"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	goMySQL "github.com/go-sql-driver/mysql"
)

var (
	ErrInvalidRequest = errors.New("invalid request")
)

type Validatable interface {
	Validate() ValidationError
}

// ValidateHTTPRequest is a helper to validate HTTP request, which must implement interface Validatable.
// This validator returns HTTPError, thus is expected to be used on controller layer.
func ValidateHTTPRequest(req Validatable, allowEmpty bool) HTTPError {
	if req == nil && allowEmpty {
		return NewHTTPError(http.StatusBadRequest, fmt.Errorf("nil request body"), "request body must not be empty")
	}
	if errV := req.Validate(); errV != nil {
		return NewHTTPError(http.StatusUnprocessableEntity, fmt.Errorf("req.Validate(): %v", errV), errV.GetErrorDetail())
	}

	return nil
}

type ValidationError interface {
	Error() string
	GetErrorDetail() string
	Unwrap() error
}

type validationError struct {
	Err    error
	Detail ValidationErrorDetail // in the format of: { errorKey: "<errorDetail>" }
}

type ValidationErrorDetail map[string]string

func NewValidationError(err error, detail ValidationErrorDetail) ValidationError {
	return &validationError{
		Err:    err,
		Detail: detail,
	}
}

func NewValidationErrorFromMySQLDuplicateKey(mySQLError goMySQL.MySQLError) ValidationError {
	detail := make(ValidationErrorDetail, 0)

	// Add the duplicated keys if the error is a MySQL duplicated key error
	if mySQLError.Number == MySQL_ErrorNumber_DuplicateKey {
		value, key := extractSQLValueAndKeyNameFromErrorString(mySQLError.Message)
		existingValue, ok := detail[key]
		if ok {
			logging.AppLogger.Warn("Replacing existing validation error detail: key='%s', from value='%s' to '%s'", key, existingValue, value)
		}
		detail[key] = fmt.Sprintf("'%s' already exists", value)
	} else {
		panic("received invalid MySQLError, not a 'duplicate key error'")
	}

	return &validationError{
		Err:    &mySQLError,
		Detail: detail,
	}
}

func (e *validationError) Error() string {
	return fmt.Sprintf("validation error (detail: `%s`): %v", e.GetErrorDetail(), e.Err)
}

func (e *validationError) GetErrorDetail() string {
	jsonString, err := json.MarshalIndent(e.Detail, "", " ")
	if err != nil {
		logging.AppLogger.Error("Unable to marshal ValidationErrorDetail, detail: %#v", e.Detail)
	}
	return fmt.Sprintf("%s", jsonString)
}

func (e *validationError) Unwrap() error {
	return e.Err
}
