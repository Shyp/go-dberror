package dberror_test

import (
	dberror "github.com/Shyp/go-dberror"
	"github.com/lib/pq"
)

func ExampleRegisterConstraint() {
	constraint := &dberror.Constraint{
		Name: "accounts_balance_check",
		GetError: func(e *pq.Error) *dberror.Error {
			return &dberror.Error{
				Message:  "Cannot write a negative balance",
				Severity: e.Severity,
				Table:    e.Table,
				Detail:   e.Detail,
				Code:     string(e.Code),
			}
		},
	}
	dberror.RegisterConstraint(constraint)
}
