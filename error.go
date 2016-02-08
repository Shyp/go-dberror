package dberror

import (
	"fmt"
	"strings"

	"github.com/Shyp/go-dberror/Godeps/_workspace/src/github.com/lib/pq"
)

const (
	CodeInvalidTextRepresentation = "22P02"
	CodeNotNullViolation          = "23502"
	CodeCheckViolation            = "23514"
)

type Error struct {
	Message    string
	Code       string
	Constraint string
	Severity   string
	Routine    string
	Table      string
	Detail     string
	Column     string
}

func (dbe *Error) Error() string {
	return dbe.Message
}

type Constraint struct {
	Name     string
	GetError func(*pq.Error) *Error
}

var constraintMap = map[string]Constraint{}

func RegisterConstraint(c Constraint) {
	constraintMap[c.Name] = c
}

// GetError parses a given database error and returns a human-readable
// version of that error. If the error is unknown, it's returned as is.
func GetError(err error) error {
	if err == nil {
		return nil
	}
	fmt.Printf("%#v\n", err)
	switch pqerr := err.(type) {
	case *pq.Error:
		switch pqerr.Code {
		case CodeInvalidTextRepresentation:
			msg := strings.Replace(pqerr.Message, "invalid input syntax for", "Invalid input syntax for type", -1)
			msg = strings.Replace(msg, "invalid input value for enum", "Invalid", -1)
			return &Error{
				Message:  msg,
				Code:     string(pqerr.Code),
				Severity: pqerr.Severity,
			}
		case CodeNotNullViolation:
			msg := fmt.Sprintf("No %[1]s was provided. Please provide a %[1]s", pqerr.Column)
			return &Error{
				Message:  msg,
				Code:     string(pqerr.Code),
				Column:   pqerr.Column,
				Table:    pqerr.Table,
				Severity: pqerr.Severity,
			}
		case CodeCheckViolation:
			c, ok := constraintMap[pqerr.Constraint]
			if ok {
				return c.GetError(pqerr)
			} else {
				return &Error{
					Message:    pqerr.Message,
					Code:       string(pqerr.Code),
					Column:     pqerr.Column,
					Table:      pqerr.Table,
					Severity:   pqerr.Severity,
					Constraint: pqerr.Constraint,
				}
			}
		default:
			fmt.Printf("%#v\n", pqerr)
			return &Error{
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
}
