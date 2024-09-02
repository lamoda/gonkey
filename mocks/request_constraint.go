package mocks

import (
	"bytes"
	"encoding/json"
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

type nopConstraint struct{}

func (c *nopConstraint) Verify(r *http.Request) []error {
	return nil
}

type bodyMatchesXMLConstraint struct {
	expectedBody  interface{}
	compareParams compare.Params
}

func newBodyMatchesXMLConstraint(expected string, params compare.Params) (verifier, error) {
	expectedBody, err := xmlparsing.Parse(expected)
	if err != nil {
		return nil, err
	}

	res := &bodyMatchesXMLConstraint{
		expectedBody:  expectedBody,
		compareParams: params,
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
		return []error{fmt.Errorf("request is empty")}
	}

	actual, err := xmlparsing.Parse(string(body))
	if err != nil {
		return []error{err}
	}

	return compare.Compare(c.expectedBody, actual, c.compareParams)
}

type bodyMatchesJSONConstraint struct {
	expectedBody  interface{}
	compareParams compare.Params
}

func newBodyMatchesJSONConstraint(expected string, params compare.Params) (verifier, error) {
	var expectedBody interface{}
	err := json.Unmarshal([]byte(expected), &expectedBody)
	if err != nil {
		return nil, err
	}
	res := &bodyMatchesJSONConstraint{
		expectedBody:  expectedBody,
		compareParams: params,
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
		return []error{fmt.Errorf("request is empty")}
	}
	var actual interface{}
	err = json.Unmarshal(body, &actual)
	if err != nil {
		return []error{err}
	}
	return compare.Compare(c.expectedBody, actual, c.compareParams)
}

type bodyJSONFieldMatchesJSONConstraint struct {
	path          string
	expected      interface{}
	compareParams compare.Params
}

func newBodyJSONFieldMatchesJSONConstraint(path, expected string, params compare.Params) (verifier, error) {
	var v interface{}
	err := json.Unmarshal([]byte(expected), &v)
	if err != nil {
		return nil, err
	}
	res := &bodyJSONFieldMatchesJSONConstraint{
		path:          path,
		expected:      v,
		compareParams: params,
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
	if value.String() == "" {
		return []error{fmt.Errorf("json field %s is empty", c.path)}
	}

	var actual interface{}
	err = json.Unmarshal([]byte(value.String()), &actual)
	if err != nil {
		return []error{err}
	}
	return compare.Compare(c.expected, actual, c.compareParams)
}

type methodConstraint struct {
	method string
}

func (c *methodConstraint) Verify(r *http.Request) []error {
	if !strings.EqualFold(r.Method, c.method) {
		return []error{fmt.Errorf("method does not match: expected %s, actual %s", r.Method, c.method)}
	}
	return nil
}

type headerConstraint struct {
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

type queryRegexpConstraint struct {
	expectedQuery map[string][]string
}

func newQueryRegexpConstraint(query string) (*queryRegexpConstraint, error) {
	// user may begin his query with '?', just omit it in this case
	query = strings.TrimPrefix(query, "?")

	rawParams := strings.Split(query, "&")

	expectedQuery := map[string][]string{}
	for _, rawParam := range rawParams {
		parts := strings.Split(rawParam, "=")
		if len(parts) != 2 {
			return nil, fmt.Errorf("error parsing query: got %d parts, expected 2", len(parts))
		}

		_, ok := expectedQuery[parts[0]]
		if !ok {
			expectedQuery[parts[0]] = make([]string, 0)
		}
		expectedQuery[parts[0]] = append(expectedQuery[parts[0]], parts[1])
	}

	return &queryRegexpConstraint{expectedQuery}, nil
}

func (c *queryRegexpConstraint) Verify(r *http.Request) (errors []error) {
	gotQuery := r.URL.Query()
	for key, want := range c.expectedQuery {
		got, ok := gotQuery[key]
		if !ok {
			errors = append(errors, fmt.Errorf("'%s' parameter is missing in expQuery", key))
			continue
		}

		if ok, err := compare.Query(want, got); err != nil {
			errors = append(errors, fmt.Errorf(
				"'%s' parameters comparison failed. \n %s'", key, err.Error(),
			))
		} else if !ok {
			errors = append(errors, fmt.Errorf(
				"'%s' parameters are not equal.\n Got: %s \n Want: %s", key, got, want,
			))
		}
	}

	return errors
}

type pathConstraint struct {
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
