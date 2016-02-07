package dberror

import (
	"database/sql"
	"os"
	"testing"

	"github.com/letsencrypt/boulder/test"
)

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
	err := db.QueryRow("INSERT INTO accounts (id) VALUES (null)").Scan()
	dberr := GetDBError(err)
	switch e := dberr.(type) {
	case *DBError:
		{
			test.AssertEquals(t, e.Error(), "No id was provided. Please provide a id")
			test.AssertEquals(t, e.Column, "id")
			test.AssertEquals(t, e.Table, "accounts")
		}
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
		{
			test.AssertEquals(t, e.Error(), "Invalid input syntax for type uuid: \"foo\"")
		}
	default:
		t.Fail()
	}
}
