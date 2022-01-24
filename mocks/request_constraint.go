package mocks

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"sort"
	"strings"

	"github.com/lamoda/gonkey/compare"
	"github.com/lamoda/gonkey/xmlparsing"
	"github.com/tidwall/gjson"
)

type verifier interface {
	Verify(r *http.Request) []error
}

type nopConstraint struct {
	verifier
}

func (c *nopConstraint) Verify(r *http.Request) []error {
	return nil
}

type bodyMatchesXMLConstraint struct {
	expectedBody interface{}
}

func newBodyMatchesXMLConstraint(expected string) (verifier, error) {
	expectedBody, err := xmlparsing.Parse(expected)
	if err != nil {
		return nil, err
	}

	res := &bodyMatchesXMLConstraint{
		expectedBody: expectedBody,
	}
	return res, nil
}

func (c *bodyMatchesXMLConstraint) Verify(r *http.Request) []error {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return []error{err}
	}
	// write body for future reusing
	r.Body = ioutil.NopCloser(bytes.NewReader(body))
	if len(body) == 0 {
		return []error{errors.New("request is empty")}
	}

	actual, err := xmlparsing.Parse(string(body))
	if err != nil {
		return []error{err}
	}

	params := compare.CompareParams{
		IgnoreArraysOrdering: true,
	}
	return compare.Compare(c.expectedBody, actual, params)
}

type bodyMatchesJSONConstraint struct {
	verifier

	expectedBody interface{}
}

func newBodyMatchesJSONConstraint(expected string) (verifier, error) {
	var expectedBody interface{}
	err := json.Unmarshal([]byte(expected), &expectedBody)
	if err != nil {
		return nil, err
	}
	res := &bodyMatchesJSONConstraint{
		expectedBody: expectedBody,
	}
	return res, nil
}

func (c *bodyMatchesJSONConstraint) Verify(r *http.Request) []error {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return []error{err}
	}
	// write body for future reusing
	r.Body = ioutil.NopCloser(bytes.NewReader(body))
	if len(body) == 0 {
		return []error{errors.New("request is empty")}
	}
	var actual interface{}
	err = json.Unmarshal(body, &actual)
	if err != nil {
		return []error{err}
	}
	params := compare.CompareParams{
		IgnoreArraysOrdering: true,
	}
	return compare.Compare(c.expectedBody, actual, params)
}

type bodyJSONFieldMatchesJSONConstraint struct {
	path     string
	expected interface{}
}

func newBodyJSONFieldMatchesJSONConstraint(path, expected string) (verifier, error) {
	var v interface{}
	err := json.Unmarshal([]byte(expected), &v)
	if err != nil {
		return nil, err
	}
	res := &bodyJSONFieldMatchesJSONConstraint{
		path:     path,
		expected: v,
	}
	return res, nil
}

func (c *bodyJSONFieldMatchesJSONConstraint) Verify(r *http.Request) []error {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return []error{err}
	}

	// write body for future reusing
	r.Body = ioutil.NopCloser(bytes.NewReader(body))

	value := gjson.Get(string(body), c.path)
	if !value.Exists() {
		return []error{fmt.Errorf("json field %s does not exist", c.path)}
	}
	if len(value.String()) == 0 {
		return []error{fmt.Errorf("json field %s is empty", c.path)}
	}

	var actual interface{}
	err = json.Unmarshal([]byte(value.String()), &actual)
	if err != nil {
		return []error{err}
	}
	params := compare.CompareParams{
		IgnoreArraysOrdering: true,
	}
	return compare.Compare(c.expected, actual, params)
}

type methodConstraint struct {
	verifier

	method string
}

func (c *methodConstraint) Verify(r *http.Request) []error {
	if !strings.EqualFold(r.Method, c.method) {
		return []error{fmt.Errorf("method does not match: expected %s, actual %s", r.Method, c.method)}
	}
	return nil
}

type headerConstraint struct {
	verifier

	header string
	value  string
	regexp *regexp.Regexp
}

func newHeaderConstraint(header, value, re string) (verifier, error) {
	var reCompiled *regexp.Regexp
	if re != "" {
		var err error
		reCompiled, err = regexp.Compile(re)
		if err != nil {
			return nil, err
		}
	}
	res := &headerConstraint{
		header: header,
		value:  value,
		regexp: reCompiled,
	}
	return res, nil
}

func (c *headerConstraint) Verify(r *http.Request) []error {
	value := r.Header.Get(c.header)
	if value == "" {
		return []error{fmt.Errorf("request doesn't have header %s", c.header)}
	}
	if c.value != "" && c.value != value {
		return []error{fmt.Errorf("%s header value %s doesn't match expected %s", c.header, value, c.value)}
	}
	if c.regexp != nil && !c.regexp.MatchString(value) {
		return []error{fmt.Errorf("%s header value %s doesn't match regexp %s", c.header, value, c.regexp)}
	}
	return nil
}

type queryConstraint struct {
	expectedQuery url.Values
}

func newQueryConstraint(query string) (*queryConstraint, error) {
	// user may begin his query with '?', just omit it in this case
	query = strings.TrimPrefix(query, "?")
	pq, err := url.ParseQuery(query)
	if err != nil {
		return nil, err
	}

	return &queryConstraint{expectedQuery: pq}, nil
}

func (c *queryConstraint) Verify(r *http.Request) (errors []error) {
	gotQuery := r.URL.Query()
	for key, want := range c.expectedQuery {
		got, ok := gotQuery[key]
		if !ok {
			errors = append(errors, fmt.Errorf("'%s' parameter is missing in expQuery", key))
			continue
		}

		sort.Strings(got)
		sort.Strings(want)
		if !reflect.DeepEqual(got, want) {
			errors = append(errors, fmt.Errorf(
				"'%s' parameters are not equal.\n Got: %s \n Want: %s", key, got, want,
			))
		}
	}

	return errors
}

type pathConstraint struct {
	verifier

	path   string
	regexp *regexp.Regexp
}

func newPathConstraint(path, re string) (verifier, error) {
	var reCompiled *regexp.Regexp
	if re != "" {
		var err error
		reCompiled, err = regexp.Compile(re)
		if err != nil {
			return nil, err
		}
	}
	res := &pathConstraint{
		path:   path,
		regexp: reCompiled,
	}
	return res, nil
}

func (c *pathConstraint) Verify(r *http.Request) []error {
	path := r.URL.Path
	if c.path != "" && c.path != path {
		return []error{fmt.Errorf("url path %s doesn't match expected %s", path, c.path)}
	}
	if c.regexp != nil && !c.regexp.MatchString(path) {
		return []error{fmt.Errorf("url path %s doesn't match regexp %s", path, c.regexp)}
	}
	return nil
}

type bodyMatchesTextConstraint struct {
	verifier

	body   string
	regexp *regexp.Regexp
}

func newBodyMatchesTextConstraint(body, re string) (verifier, error) {
	var reCompiled *regexp.Regexp
	if re != "" {
		var err error
		reCompiled, err = regexp.Compile(re)
		if err != nil {
			return nil, err
		}
	}
	res := &bodyMatchesTextConstraint{
		body:   body,
		regexp: reCompiled,
	}
	return res, nil
}

func (c *bodyMatchesTextConstraint) Verify(r *http.Request) []error {
	ioBody, err := ioutil.ReadAll(r.Body)

	if err != nil {
		return []error{err}
	}

	// write body for future reusing
	r.Body = ioutil.NopCloser(bytes.NewReader(ioBody))

	body := string(ioBody)

	if c.body != "" && c.body != body {
		return []error{fmt.Errorf("body value\n%s\ndoesn't match expected\n%s", body, c.body)}
	}
	if c.regexp != nil && !c.regexp.MatchString(body) {
		return []error{fmt.Errorf("body value\n%s\ndoesn't match regexp %s", body, c.regexp)}
	}
	return nil
}
