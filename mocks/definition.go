package mocks

import (
	"fmt"
	"net/http"
	"net/http/httputil"
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
		requestDump, err := httputil.DumpRequest(r, true)
		if err != nil {
			fmt.Printf("Gonkey internal error: %s\n", err)
		}

		for _, c := range d.requestConstraints {
			errs := c.Verify(r)
			for _, e := range errs {
				errors = append(errors, &RequestConstraintError{
					error:       e,
					Constraint:  c,
					RequestDump: requestDump,
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

func verifyRequestConstraints(requestConstraints []verifier, r *http.Request) bool {
	var errors []error
	if len(requestConstraints) > 0 {
		requestDump, err := httputil.DumpRequest(r, true)
		if err != nil {
			fmt.Printf("Gonkey internal error: %s\n", err)
		}

		for _, c := range requestConstraints {
			errs := c.Verify(r)
			for _, e := range errs {
				errors = append(errors, &RequestConstraintError{
					error:       e,
					Constraint:  c,
					RequestDump: requestDump,
				})
			}
		}
	}
	return errors == nil
}
func (d *definition) ExecuteWithoutVerifying(w http.ResponseWriter, r *http.Request) []error {
	d.Lock()
	d.calls++
	d.Unlock()
	if d.replyStrategy != nil {
		return d.replyStrategy.HandleRequest(w, r)
	}
	return []error{
		fmt.Errorf("reply strategy undefined"),
	}
}
