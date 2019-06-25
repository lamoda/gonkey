package mocks

import (
	"fmt"
	"net/http"
	"sync"
)

const callsNoConstraint = -1

type definition struct {
	path               string
	requestConstraints []verifier
	replyStrategy      replyStrategy
	sync.Mutex
	calls           int
	callsConstraint int
}

func newDefinition(path string, constraints []verifier, strategy replyStrategy, callsConstraint int) *definition {
	return &definition{
		path:               path,
		requestConstraints: constraints,
		replyStrategy:      strategy,
		callsConstraint:    callsConstraint,
	}
}

func (d *definition) Execute(w http.ResponseWriter, r *http.Request) []error {
	d.Lock()
	d.calls++
	d.Unlock()
	var errors []error
	if len(d.requestConstraints) > 0 {
		for _, c := range d.requestConstraints {
			errs := c.Verify(r)
			for _, e := range errs {
				errors = append(errors, &RequestConstraintError{
					error:      e,
					Constraint: c,
					Request:    r,
				})
			}
		}
	}
	if d.replyStrategy != nil {
		errors = append(errors, d.replyStrategy.HandleRequest(w, r)...)
	}
	return errors
}

func (d *definition) ResetRunningContext() {
	if s, ok := d.replyStrategy.(contextAwareStrategy); ok {
		s.ResetRunningContext()
	}
	d.Lock()
	d.calls = 0
	d.Unlock()
}

func (d *definition) EndRunningContext() []error {
	d.Lock()
	defer d.Unlock()
	var errs []error
	if s, ok := d.replyStrategy.(contextAwareStrategy); ok {
		errs = s.EndRunningContext()
	}
	if d.callsConstraint != callsNoConstraint && d.calls != d.callsConstraint {
		err := fmt.Errorf("at path %s: number of calls does not match: expected %d, actual %d",
			d.path, d.callsConstraint, d.calls)
		errs = append(errs, err)
	}
	return errs
}
