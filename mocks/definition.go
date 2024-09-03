package mocks

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"sync"
)

const CallsNoConstraint = -1

type Definition struct {
	path               string
	requestConstraints []verifier
	replyStrategy      ReplyStrategy
	mutex              sync.Mutex
	calls              int
	callsConstraint    int
}

func NewDefinition(path string, constraints []verifier, strategy ReplyStrategy, callsConstraint int) *Definition {
	return &Definition{
		path:               path,
		requestConstraints: constraints,
		replyStrategy:      strategy,
		callsConstraint:    callsConstraint,
	}
}

func (d *Definition) Execute(w http.ResponseWriter, r *http.Request) []error {
	d.mutex.Lock()
	d.calls++
	d.mutex.Unlock()
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

func (d *Definition) ResetRunningContext() {
	if s, ok := d.replyStrategy.(contextAwareStrategy); ok {
		s.ResetRunningContext()
	}
	d.mutex.Lock()
	d.calls = 0
	d.mutex.Unlock()
}

func (d *Definition) EndRunningContext() []error {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	var errs []error
	if s, ok := d.replyStrategy.(contextAwareStrategy); ok {
		errs = s.EndRunningContext()
	}
	if d.callsConstraint != CallsNoConstraint && d.calls != d.callsConstraint {
		err := fmt.Errorf("at path %s: number of calls does not match: expected %d, actual %d",
			d.path, d.callsConstraint, d.calls)
		errs = append(errs, err)
	}
	return errs
}

func verifyRequestConstraints(requestConstraints []verifier, r *http.Request) []error {
	var errors []error
	if len(requestConstraints) > 0 {
		requestDump, err := httputil.DumpRequest(r, true)
		if err != nil {
			requestDump = []byte(fmt.Sprintf("failed to dump request: %s", err))
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
	return errors
}
func (d *Definition) ExecuteWithoutVerifying(w http.ResponseWriter, r *http.Request) []error {
	d.mutex.Lock()
	d.calls++
	d.mutex.Unlock()
	if d.replyStrategy != nil {
		return d.replyStrategy.HandleRequest(w, r)
	}
	return []error{
		fmt.Errorf("reply strategy undefined"),
	}
}
