package dberror

import (
	"database/sql"
	"os"
	"testing"

	"github.com/Shyp/go-dberror/Godeps/_workspace/src/github.com/letsencrypt/boulder/test"
	"github.com/Shyp/go-dberror/Godeps/_workspace/src/github.com/lib/pq"
)

const uuid = "3c7d2b4a-3fc8-4782-a518-4ce9efef51e7"
const email = "test@example.com"

var db *sql.DB

func TestNilError(t *testing.T) {
	test.AssertEquals(t, GetError(nil), nil)
}

func setUp(t *testing.T) {
	if db != nil {
		return
	}
	ci := os.Getenv("CI")
	var err error
	if ci == "" {
		db, err = sql.Open("postgres", "postgres://localhost/dberror?sslmode=disable")
	} else {
		db, err = sql.Open("postgres", "postgres://ubuntu@localhost/circle_test?sslmode=disable")
	}
	if err != nil {
		t.Fatal(err)
	}
}

func TestNotNull(t *testing.T) {
	setUp(t)
	_, err := db.Exec("INSERT INTO accounts (id) VALUES (null)")
	dberr := GetError(err)
	switch e := dberr.(type) {
	case *Error:
		test.AssertEquals(t, e.Error(), "No id was provided. Please provide a id")
		test.AssertEquals(t, e.Column, "id")
		test.AssertEquals(t, e.Table, "accounts")
	default:
		t.Fail()
	}
}

func TestDefaultConstraint(t *testing.T) {
	// this test needs to go before the Register() below... not great, add an
	// unregister or clear out the map or something
	setUp(t)
	_, err := db.Exec("INSERT INTO accounts (id, email, balance) VALUES ($1, $2, -1)", uuid, email)
	dberr := GetError(err)
	switch e := dberr.(type) {
	case *Error:
		test.AssertEquals(t, e.Error(), "new row for relation \"accounts\" violates check constraint \"accounts_balance_check\"")
		test.AssertEquals(t, e.Table, "accounts")
	default:
		t.Fail()
	}
}

func TestCustomConstraint(t *testing.T) {
	setUp(t)
	constraint := Constraint{
		Name: "accounts_balance_check",
		GetError: func(e *pq.Error) *Error {
			return &Error{
				Message:  "Cannot write a negative balance",
				Severity: e.Severity,
				Table:    e.Table,
				Detail:   e.Detail,
				Code:     string(e.Code),
			}
		},
	}
	RegisterConstraint(constraint)
	_, err := db.Exec("INSERT INTO accounts (id, email, balance) VALUES ($1, $2, -1)", uuid, email)
	dberr := GetError(err)
	switch e := dberr.(type) {
	case *Error:
		test.AssertEquals(t, e.Error(), "Cannot write a negative balance")
		test.AssertEquals(t, e.Table, "accounts")
	default:
		t.Fail()
	}
}

func TestInvalidUUID(t *testing.T) {
	setUp(t)
	_, err := db.Exec("INSERT INTO accounts (id) VALUES ('foo')")
	dberr := GetError(err)
	switch e := dberr.(type) {
	case *Error:
		test.AssertEquals(t, e.Error(), "Invalid input syntax for type uuid: \"foo\"")
	default:
		t.Fail()
	}
}

func TestInvalidEnum(t *testing.T) {
	setUp(t)
	_, err := db.Exec("INSERT INTO accounts (id, email, balance, status) VALUES ($1, $2, 1, 'blah')", uuid, email)
	dberr := GetError(err)
	switch e := dberr.(type) {
	case *Error:
		test.AssertEquals(t, e.Error(), "Invalid account_status: \"blah\"")
	default:
		t.Fail()
	}
}

func TestTooLargeInt(t *testing.T) {
	setUp(t)
	_, err := db.Exec("INSERT INTO accounts (id, email, balance) VALUES ($1, $2, 40000)", uuid, email)
	dberr := GetError(err)
	switch e := dberr.(type) {
	case *Error:
		test.AssertEquals(t, e.Error(), "Smallint too large or too small")
	default:
		t.Fail()
	}
}

func TestCapitalize(t *testing.T) {
	test.AssertEquals(t, capitalize("foo"), "Foo")
	test.AssertEquals(t, capitalize("foo bar baz"), "Foo bar baz")
}
