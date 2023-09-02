package errs

import (
	"errors"
	"fmt"
	"regexp"
	"sonamusica-backend/logging"

	goMySQL "github.com/go-sql-driver/mysql"
)

const (
	MySQL_ErrorNumber_DuplicateEntry    = 1062
	MySQL_ErrorNumber_RowIsReferenced_1 = 1217
	MySQL_ErrorNumber_RowIsReferenced_2 = 1451
	MySQL_ErrorNumber_NoReferencedRow_1 = 1216
	MySQL_ErrorNumber_NoReferencedRow_2 = 1452
)

var regexExtractTextInsideQuotes = regexp.MustCompile(`'([^']+)'`)

// WrapMySQLError wrap (or we can say parse) errors returned from SQL operations into a more useful service-layer logic.
//
//	Examples:
//	1. DuplicateKey, RowIsReferenced, NoReferencedRow errors in most cases should be propagated to client as a validation error (e.g. user is trying to sign up using used email/username)
func WrapMySQLError(err error) error {
	if err == nil {
		return nil
	}

	wrappedErr := err

	var mySQLErr *goMySQL.MySQLError
	if errors.As(err, &mySQLErr) {
		if mySQLErr.Number == MySQL_ErrorNumber_DuplicateEntry {
			wrappedErr = convertDuplicateKeyErrToValidationErr(*mySQLErr)
		} else if mySQLErr.Number == MySQL_ErrorNumber_RowIsReferenced_1 || mySQLErr.Number == MySQL_ErrorNumber_RowIsReferenced_2 {
			wrappedErr = convertRowIsReferencedErrToValidationErr(*mySQLErr)
		} else if mySQLErr.Number == MySQL_ErrorNumber_NoReferencedRow_1 || mySQLErr.Number == MySQL_ErrorNumber_NoReferencedRow_2 {
			wrappedErr = convertNoReferencedRowErrToValidationErr(*mySQLErr)
		}
	} else {
		panic(fmt.Sprintf("WrapMySQLError received an unsupported error type: %v", err))
	}

	return wrappedErr
}

func convertDuplicateKeyErrToValidationErr(mySQLError goMySQL.MySQLError) ValidationError {
	detail := make(ValidationErrorDetail, 0)

	// Add the duplicated keys if the error is a MySQL duplicated key error
	if mySQLError.Number == MySQL_ErrorNumber_DuplicateEntry {
		value, key := extractSQLValueAndKeyNameFromErrorString(mySQLError.Message)
		existingValue, ok := detail[key]
		if ok {
			logging.AppLogger.Warn("Replacing existing validation error detail: key='%s', from value='%s' to '%s'", key, existingValue, value)
		}
		detail[key] = fmt.Sprintf("'%s' '%s' already exists", key, value)
	} else {
		panic("received invalid MySQLError, not a 'duplicate key' error")
	}

	return NewValidationError(&mySQLError, detail)
}

func convertRowIsReferencedErrToValidationErr(mySQLError goMySQL.MySQLError) ValidationError {
	detail := make(ValidationErrorDetail, 0)

	if mySQLError.Number == MySQL_ErrorNumber_RowIsReferenced_1 || mySQLError.Number == MySQL_ErrorNumber_RowIsReferenced_2 {
		detail[ClientMessageKey_NonField] = "unable to update or delete as the object(s) is being referenced by another entity. try deleting the referencing entity first."
	} else {
		panic("received invalid MySQLError, not a 'row is referenced' error")
	}

	return NewValidationError(&mySQLError, detail)
}

func convertNoReferencedRowErrToValidationErr(mySQLError goMySQL.MySQLError) ValidationError {
	detail := make(ValidationErrorDetail, 0)

	if mySQLError.Number == MySQL_ErrorNumber_NoReferencedRow_1 || mySQLError.Number == MySQL_ErrorNumber_NoReferencedRow_2 {
		detail[ClientMessageKey_NonField] = "unable to create or update as the referred object(s) doesn't exist. try creating the entity first."
	} else {
		panic("received invalid MySQLError, not a 'no referenced row' error")
	}

	return NewValidationError(&mySQLError, detail)
}

func extractSQLValueAndKeyNameFromErrorString(s string) (string, string) {
	regexResults := regexExtractTextInsideQuotes.FindAllStringSubmatch(s, -1)
	if len(regexResults) != 2 {
		return "", ""
	}
	return regexResults[0][1], regexResults[1][1] // we take the 2nd element to take the group value (excluding the single tick "'")
}
