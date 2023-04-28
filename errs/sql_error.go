package errs

import (
	"database/sql"
	"errors"
	"regexp"

	goMySQL "github.com/go-sql-driver/mysql"
)

const MySQL_ErrorNumber_DuplicateKey = 1062

var regexExtractTextInsideQuotes = regexp.MustCompile(`'([^']+)'`)

// WrapMySQLError wrap (or we can say parse) errors returned from SQL operations into a more useful service-layer logic.
//
//	Examples:
//	1. most of ErrNoRows are not errors, so we can ignore it
//	2. DuplicateKey error in most cases should be propagated to client as a validation error (e.g. user is trying to sign up using used email/username)
func WrapMySQLError(err error) error {
	if err == nil {
		return nil
	}

	wrappedErr := err

	var mySQLErr *goMySQL.MySQLError
	if errors.As(err, &mySQLErr) {
		if mySQLErr.Number == MySQL_ErrorNumber_DuplicateKey {
			wrappedErr = NewValidationErrorFromMySQLDuplicateKey(*mySQLErr)
		}
	} else if errors.Is(err, sql.ErrNoRows) { // we assume record not found = no error
		wrappedErr = nil
	} else {
		panic("WrapMySQLError received an unsupported error type")
	}

	return wrappedErr
}

func extractSQLValueAndKeyNameFromErrorString(s string) (string, string) {
	regexResults := regexExtractTextInsideQuotes.FindAllStringSubmatch(s, -1)
	if len(regexResults) != 2 {
		return "", ""
	}
	return regexResults[0][1], regexResults[1][1] // we take the 2nd element to take the group value (excluding the single tick "'")
}
