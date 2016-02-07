package dberror

import (
	"fmt"
	"strings"

	"github.com/lib/pq"
)

const (
	CodeInvalidTextRepresentation = "22P02"
	CodeNotNullViolation          = "23502"
)

type DBError struct {
	Message    string
	Code       string
	Constraint string
	Severity   string
	Routine    string
	Table      string
	Column     string
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
			fmt.Printf("%#v\n", pqerr)
			if pqerr.Code == CodeInvalidTextRepresentation {
				msg := strings.Replace(pqerr.Message, "invalid input syntax for", "Invalid input syntax for type", -1)
				return &DBError{
					Message:  msg,
					Code:     string(pqerr.Code),
					Severity: pqerr.Severity,
				}
			} else if pqerr.Code == CodeNotNullViolation {
				msg := fmt.Sprintf("No %[1]s was provided. Please provide a %[1]s", pqerr.Column)
				return &DBError{
					Message:  msg,
					Code:     string(pqerr.Code),
					Column:   pqerr.Column,
					Table:    pqerr.Table,
					Severity: pqerr.Severity,
				}
			}
			return &DBError{
				Message:    pqerr.Message,
				Code:       string(pqerr.Code),
				Column:     pqerr.Column,
				Constraint: pqerr.Constraint,
				Table:      pqerr.Table,
				Routine:    pqerr.Routine,
				Severity:   pqerr.Severity,
			}
		}
	default:
		return pqerr
	}
	return nil
}
