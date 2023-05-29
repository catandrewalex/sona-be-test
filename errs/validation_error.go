package errs

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"sonamusica-backend/logging"
	"strings"

	goMySQL "github.com/go-sql-driver/mysql"
)

var (
	ErrInvalidRequest = errors.New("invalid request")
)

type Validatable interface {
	Validate() ValidationError
}

// ValidateHTTPRequest is a helper to validate HTTP request, which must implement interface Validatable.
// This validator also marks zero value fields which doesn't have tag `json:"..,omitempty"` as error, unless allowZeroValues == true.
//
// This validator returns HTTPError, thus is expected to be used on controller layer.
func ValidateHTTPRequest(req Validatable, allowZeroValues bool) HTTPError {
	if req == nil {
		return NewHTTPError(http.StatusBadRequest, fmt.Errorf("nil request body"), map[string]string{ClientMessageKey_NonField: "request body must not be empty"}, "")
	}

	if !allowZeroValues {
		if errV := validateZeroValues(req); errV != nil {
			return NewHTTPError(http.StatusUnprocessableEntity, fmt.Errorf("validateZeroValues(): %v", errV), errV.GetErrorDetail(), "")
		}
	}

	if errV := req.Validate(); errV != nil {
		return NewHTTPError(http.StatusUnprocessableEntity, fmt.Errorf("req.Validate(): %v", errV), errV.GetErrorDetail(), "")
	}

	return nil
}

func validateZeroValues(struct_ interface{}) ValidationError {
	errorDetail := getStructZeroValuesErrorDetail(struct_, "")

	if len(errorDetail) > 0 {
		return NewValidationError(ErrInvalidRequest, errorDetail)
	}

	return nil
}

// getStructZeroValuesErrorDetail receives struct & pointer to struct,
//
//	and populate "empty field error" if the field json tag does not end with ",omitempty".
//	All fields & slices will be checked recursively if they are structs.
func getStructZeroValuesErrorDetail(struct_ interface{}, fieldNamePrefix string) ValidationErrorDetail {
	errorDetail := make(ValidationErrorDetail, 0)

	// we accept struct & pointer to struct
	val := reflect.ValueOf(struct_)
	logging.HTTPServerLogger.Debug("val type: %s", val.Type())
	if val.Kind() == reflect.Pointer {
		val = reflect.Indirect(val)
		logging.HTTPServerLogger.Debug("val type after conversion: %s", val.Type())
	}
	if val.Kind() != reflect.Struct {
		panic(fmt.Sprintf("invalid input parameter! expected raw/pointerTo %q but found %q", reflect.Struct.String(), val.Kind().String()))
	}

	for i := 0; i < val.NumField(); i++ {
		typeField := val.Type().Field(i)
		logging.HTTPServerLogger.Debug("%+v", typeField)
		jsonTag := typeField.Tag.Get("json")
		if strings.HasSuffix(jsonTag, ",omitempty") {
			continue
		}

		var fieldName string
		tags := strings.Split(jsonTag, ",")
		if len(tags) > 0 && tags[0] != "" {
			fieldName = tags[0]
		} else if !typeField.Anonymous { // anonymous == embedded type, which doesn't have its own field name
			fieldName = typeField.Name
		}
		if len(fieldNamePrefix) > 0 {
			fieldName = fmt.Sprintf("%s.%s", fieldNamePrefix, fieldName)
		}
		logging.HTTPServerLogger.Debug("\tfieldName: %s", fieldName)

		valueField := val.Field(i)
		logging.HTTPServerLogger.Debug("\tvalueField type: %s", valueField.Type())
		logging.HTTPServerLogger.Debug("\tvalueField kind: %s", valueField.Kind())

		if valueField.Kind() == reflect.Struct { // recursively check struct zero values
			childErrorDetail := getStructZeroValuesErrorDetail(valueField.Interface(), fieldName)
			for k, v := range childErrorDetail {
				errorDetail[k] = v
			}
		} else if valueField.IsZero() { // we put the recursive base here: to make sure all struct's field is properly validated
			errorDetail[fieldName] = fmt.Sprintf("%s cannot be empty", fieldName)
		} else if valueField.Kind() == reflect.Slice {
			elemType := valueField.Type().Elem()
			if elemType.Kind() == reflect.Struct { // iterate the slice of struct, & recursively check zero values
				for j := 0; j < valueField.Len(); j++ {
					elemFieldName := fmt.Sprintf("%s.%d", fieldName, j)
					childErrorDetail := getStructZeroValuesErrorDetail(valueField.Index(j).Interface(), elemFieldName)
					for k, v := range childErrorDetail {
						errorDetail[k] = v
					}
				}
			}
		}
	}

	return errorDetail
}

type ValidationError interface {
	Error() string
	GetErrorDetail() map[string]string
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
		detail[key] = fmt.Sprintf("'%s' '%s' already exists", key, value)
	} else {
		panic("received invalid MySQLError, not a 'duplicate key error'")
	}

	return &validationError{
		Err:    &mySQLError,
		Detail: detail,
	}
}

func (e *validationError) Error() string {
	return fmt.Sprintf("validation error (detail: `%#v`): %v", e.GetErrorDetail(), e.Err)
}

func (e *validationError) GetErrorDetail() map[string]string {
	return e.Detail
}

// func (e *validationError) GetErrorDetail() string {
// 	jsonString, err := json.MarshalIndent(e.Detail, "", " ")
// 	if err != nil {
// 		logging.AppLogger.Error("Unable to marshal ValidationErrorDetail, detail: %#v", e.Detail)
// 	}
// 	return fmt.Sprintf("%s", jsonString)
// }

func (e *validationError) Unwrap() error {
	return e.Err
}
