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

type RequestConstraintError struct {
	error
	Constraint  verifier
	RequestDump []byte
}

func (e *RequestConstraintError) Error() string {
	kind := reflect.TypeOf(e.Constraint).String()
	return fmt.Sprintf("request constraint %s failed: %s, request was:\n %s", kind, e.error.Error(), e.RequestDump)
}
