package dberror

import (
	"strings"

	"github.com/lib/pq"
)

const (
	CodeInvalidTextRepresentation = "22P02"
)

type DBError struct {
	Message    string
	Code       string
	Constraint string
	Severity   string
	Routine    string
	Table      string
}

func (dbe *DBError) Error() string {
	return dbe.Message
}

// GetDBError parses a given database error and returns a human-readable
// version of that error. If the error is unknown, it's returned as is.
func GetDBError(err error) error {
	if err == nil {
		return nil
	}
	switch pqerr := err.(type) {
	case *pq.Error:
		{
			if pqerr.Code == CodeInvalidTextRepresentation {
				msg := strings.Replace(pqerr.Message, "invalid input syntax for", "Invalid input syntax for type", -1)
				return &DBError{
					Message: msg,
					Code:    string(pqerr.Code),
				}
			}
		}
	default:
		return pqerr
	}
	return nil
}
