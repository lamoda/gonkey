package mocks

import (
	"fmt"
	"reflect"
)

type Error struct {
	error
	ServiceName string
}

func (e *Error) Error() string {
	return fmt.Sprintf("mock %s: %s", e.ServiceName, e.error.Error())
}

func (e *Error) Unwrap() error {
	return e.error
}

type RequestConstraintError struct {
	error
	Constraint  verifier
	RequestDump []byte
	Endpoint    string
}

func (e *RequestConstraintError) Error() string {
	kind := reflect.TypeOf(e.Constraint).String()
	return fmt.Sprintf("request constraint %s failed: %s, request was:\n %s", kind, e.error.Error(), e.RequestDump)
}

func (e *RequestConstraintError) Unwrap() error {
	return e.error
}

type CallsMismatchError struct {
	Path        string
	Expected    int
	Actual      int
	ServiceName string
}

func (e *CallsMismatchError) Error() string {
	return fmt.Sprintf("at path %s: number of calls does not match: expected %d, actual %d",
		e.Path, e.Expected, e.Actual)
}
