package dberror

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/Shyp/go-dberror/Godeps/_workspace/src/github.com/lib/pq"
)

// See http://www.postgresql.org/docs/9.3/static/errcodes-appendix.html for
// a full listing of the error codes present here.
const (
	CodeNumericValueOutOfRange    = "22003"
	CodeInvalidTextRepresentation = "22P02"
	CodeNotNullViolation          = "23502"
	CodeUniqueViolation           = "23505"
	CodeCheckViolation            = "23514"
)

// Error is a human-readable database error. Message should always be a
// non-empty, readable string, and is returned when you call err.Error(). The
// other fields may or may not be empty.
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

// Constraint is a custom database check constraint you've defined, like "CHECK
// balance > 0". Postgres doesn't define a very useful message for constraint
// failures (new row for relation "accounts" violates check constraint), so you
// can define your own. The Name should be the name of the constraint in the
// database. Define GetError to provide your own custom error handler for this
// constraint failure, with a custom message.
type Constraint struct {
	Name     string
	GetError func(*pq.Error) *Error
}

var constraintMap = map[string]*Constraint{}

// RegisterConstraint tells dberror about your custom constraint and its error
// handling.
func RegisterConstraint(c *Constraint) {
	constraintMap[c.Name] = c
}

// capitalize the first letter in the string
func capitalize(s string) string {
	r, size := utf8.DecodeRuneInString(s)
	return fmt.Sprintf("%c", unicode.ToTitle(r)) + s[size:]
}

var columnFinder *regexp.Regexp
var valueFinder *regexp.Regexp

func init() {
	columnFinder = regexp.MustCompile(`Key \((.+)\)=`)
	valueFinder = regexp.MustCompile(`Key \(.+\)=\((.+)\)`)
}

// findColumn finds the column in the given pq Detail error string. If the
// column does not exist, the empty string is returned.
//
// detail can look like this:
//    Key (id)=(3c7d2b4a-3fc8-4782-a518-4ce9efef51e7) already exists.
func findColumn(detail string) string {
	results := columnFinder.FindStringSubmatch(detail)
	if len(results) < 2 {
		return ""
	} else {
		return results[1]
	}
}

// findColumn finds the column in the given pq Detail error string. If the
// column does not exist, the empty string is returned.
//
// detail can look like this:
//    Key (id)=(3c7d2b4a-3fc8-4782-a518-4ce9efef51e7) already exists.
func findValue(detail string) string {
	results := valueFinder.FindStringSubmatch(detail)
	if len(results) < 2 {
		return ""
	} else {
		return results[1]
	}
}

// GetError parses a given database error and returns a human-readable
// version of that error. If the error is unknown, it's returned as is,
// however, all errors of type `pq.Error` are re-thrown as an Error, so it's
// impossible to get a `pq.Error` back from this function.
func GetError(err error) error {
	if err == nil {
		return nil
	}
	switch pqerr := err.(type) {
	case *pq.Error:
		switch pqerr.Code {
		case CodeUniqueViolation:
			columnName := findColumn(pqerr.Detail)
			if columnName == "" {
				columnName = "value"
			}
			valueName := findValue(pqerr.Detail)
			var msg string
			if valueName == "" {
				msg = fmt.Sprintf("A %s already exists with that value", columnName)
			} else {
				msg = fmt.Sprintf("A %s already exists with this value (%s)", columnName, valueName)
			}
			dbe := &Error{
				Message:    msg,
				Code:       string(pqerr.Code),
				Severity:   pqerr.Severity,
				Constraint: pqerr.Constraint,
				Table:      pqerr.Table,
				Detail:     pqerr.Detail,
			}
			if columnName != "value" {
				dbe.Column = columnName
			}
			return dbe
		case CodeNumericValueOutOfRange:
			msg := strings.Replace(pqerr.Message, "out of range", "too large or too small", 1)
			return &Error{
				Message:  capitalize(msg),
				Code:     string(pqerr.Code),
				Severity: pqerr.Severity,
			}
		case CodeInvalidTextRepresentation:
			msg := strings.Replace(pqerr.Message, "invalid input syntax for", "Invalid input syntax for type", 1)
			msg = strings.Replace(msg, "invalid input value for enum", "Invalid", 1)
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
