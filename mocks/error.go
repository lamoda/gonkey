package mocks

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"reflect"
)

type Error struct {
	error
	ServiceName string
}

func (e *Error) Error() string {
	return fmt.Sprintf("mock %s: %s", e.ServiceName, e.error.Error())
}

type RequestConstraintError struct {
	error
	Constraint requestConstraint
	Request    *http.Request
}

func (e *RequestConstraintError) Error() string {
	kind := reflect.TypeOf(e.Constraint).String()
	req, err := httputil.DumpRequest(e.Request, true)
	if err != nil {
		return err.Error()
	}
	return fmt.Sprintf("request constraint %s failed: %s, request was:\n %s", kind, e.error.Error(), req)
}
