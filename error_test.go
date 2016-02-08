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
	test.AssertEquals(t, GetDBError(nil), nil)
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
	dberr := GetDBError(err)
	switch e := dberr.(type) {
	case *DBError:
		test.AssertEquals(t, e.Error(), "No id was provided. Please provide a id")
		test.AssertEquals(t, e.Column, "id")
		test.AssertEquals(t, e.Table, "accounts")
	default:
		t.Fail()
	}
}

func TestConstraint(t *testing.T) {
	setUp(t)
	constraint := Constraint{
		Name: "accounts_balance_check",
		GetError: func(e *pq.Error) *DBError {
			return &DBError{
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
	dberr := GetDBError(err)
	switch e := dberr.(type) {
	case *DBError:
		test.AssertEquals(t, e.Error(), "Cannot write a negative balance")
		test.AssertEquals(t, e.Table, "accounts")
	default:
		t.Fail()
	}
}

func TestInvalidUUID(t *testing.T) {
	setUp(t)
	var id string
	err := db.QueryRow("INSERT INTO accounts (id) VALUES ($1)", "foo").Scan(&id)
	dberr := GetDBError(err)
	switch e := dberr.(type) {
	case *DBError:
		test.AssertEquals(t, e.Error(), "Invalid input syntax for type uuid: \"foo\"")
	default:
		t.Fail()
	}
}
