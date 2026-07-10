package mocks

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"text/template"
)

const (
	fromCallLast  = "last"
	fromCallFirst = "first"
)

type variableExtractor struct {
	name     string
	tmpl     *template.Template
	fromCall string
}

func newVariableExtractor(name, tmplStr, fromCall string) (*variableExtractor, error) {
	if tmplStr == "" {
		return nil, fmt.Errorf("variable %s: template must not be empty", name)
	}
	switch fromCall {
	case "":
		fromCall = fromCallLast
	case fromCallLast, fromCallFirst:
	default:
		return nil, fmt.Errorf("variable %s: `fromCall` must be one of `%s`, `%s`, got `%s`",
			name, fromCallLast, fromCallFirst, fromCall)
	}
	tmpl, err := template.New(name).Option("missingkey=error").Parse(tmplStr)
	if err != nil {
		return nil, fmt.Errorf("variable %s: template syntax error: %w", name, err)
	}
	return &variableExtractor{
		name:     name,
		tmpl:     tmpl,
		fromCall: fromCall,
	}, nil
}

func (e *variableExtractor) extract(ctx map[string]*templateRequest) (string, error) {
	value := bytes.NewBuffer(nil)
	if err := e.tmpl.Execute(value, ctx); err != nil {
		return "", fmt.Errorf("failed to capture variable %s: %w", e.name, err)
	}
	return value.String(), nil
}

// captureVariables extracts values of the incoming request into the captured map.
// The request body is buffered and restored, so constraints and reply strategies
// can read it again.
func (d *Definition) captureVariables(r *http.Request) []error {
	if len(d.variablesToSet) == 0 {
		return nil
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return []error{err}
	}
	// write body for future reusing
	r.Body = io.NopCloser(bytes.NewReader(body))

	// templateRequest consumes the body, give it its own copy;
	// the context is shared between extractors, so the body is parsed only once
	requestCopy := *r
	requestCopy.Body = io.NopCloser(bytes.NewReader(body))
	ctx := map[string]*templateRequest{
		"request": {r: &requestCopy},
	}

	d.Lock()
	defer d.Unlock()

	var errs []error
	if d.captured == nil {
		d.captured = make(map[string]string, len(d.variablesToSet))
	}
	for _, e := range d.variablesToSet {
		if e.fromCall == fromCallFirst {
			if _, ok := d.captured[e.name]; ok {
				continue
			}
		}
		value, err := e.extract(ctx)
		if err != nil {
			errs = append(errs, fmt.Errorf("at path %s: %w", d.path, err))
			continue
		}
		d.captured[e.name] = value
	}
	return errs
}

// variablesCapturingStrategy is implemented by composite reply strategies
// that hold child Definitions capable of capturing variables.
type variablesCapturingStrategy interface {
	CapturedVariables() map[string]string
}

// CapturedVariables returns variables captured from requests
// since the last ResetRunningContext call, including variables
// captured by nested definitions of composite strategies.
func (d *Definition) CapturedVariables() map[string]string {
	d.Lock()
	res := make(map[string]string, len(d.captured))
	mergeVariables(res, d.captured)
	d.Unlock()

	if s, ok := d.replyStrategy.(variablesCapturingStrategy); ok {
		mergeVariables(res, s.CapturedVariables())
	}
	return res
}

func mergeVariables(dst, src map[string]string) {
	for name, value := range src {
		dst[name] = value
	}
}
