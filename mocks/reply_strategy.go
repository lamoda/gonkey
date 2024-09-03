package mocks

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"strings"
	"sync"
)

type ReplyStrategy interface {
	HandleRequest(w http.ResponseWriter, r *http.Request) []error
}

type contextAwareStrategy interface {
	ResetRunningContext()
	EndRunningContext() []error
}

type constantReply struct {
	replyBody  []byte
	statusCode int
	headers    map[string]string
}

func unhandledRequestError(r *http.Request) []error {
	requestContent, err := httputil.DumpRequest(r, true)
	if err != nil {
		return []error{fmt.Errorf("Gonkey internal error during request dump: %s\n", err)}
	}
	return []error{fmt.Errorf("unhandled request to mock:\n%s", requestContent)}
}

func NewFileReplyWithCode(filename string, statusCode int, headers map[string]string) (ReplyStrategy, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	r := &constantReply{
		replyBody:  content,
		statusCode: statusCode,
		headers:    headers,
	}
	return r, nil
}

func NewConstantReplyWithCode(content []byte, statusCode int, headers map[string]string) ReplyStrategy {
	return &constantReply{
		replyBody:  content,
		statusCode: statusCode,
		headers:    headers,
	}
}

func (s *constantReply) HandleRequest(w http.ResponseWriter, r *http.Request) []error {
	for k, v := range s.headers {
		w.Header().Add(k, v)
	}
	w.WriteHeader(s.statusCode)
	w.Write(s.replyBody)
	return nil
}

type dropRequestReply struct {
}

func NewDropRequestReply() ReplyStrategy {
	return &dropRequestReply{}
}

func (s *dropRequestReply) HandleRequest(w http.ResponseWriter, r *http.Request) []error {
	hj, ok := w.(http.Hijacker)
	if !ok {
		return []error{fmt.Errorf("Gonkey internal error during drop request: webserver doesn't support hijacking\n")}
	}
	conn, _, err := hj.Hijack()
	if err != nil {
		return []error{fmt.Errorf("Gonkey internal error during connection hijacking: %s\n", err)}
	}
	conn.Close()
	return nil
}

type failReply struct{}

func (s *failReply) HandleRequest(w http.ResponseWriter, r *http.Request) []error {
	return unhandledRequestError(r)
}

type nopReply struct{}

func (s *nopReply) HandleRequest(w http.ResponseWriter, r *http.Request) []error {
	w.WriteHeader(http.StatusNoContent)
	return nil
}

type uriVaryReply struct {
	basePath string
	variants map[string]*Definition
}

func NewUriVaryReply(basePath string, variants map[string]*Definition) ReplyStrategy {
	return &uriVaryReply{
		basePath: strings.TrimRight(basePath, "/") + "/",
		variants: variants,
	}
}

func (s *uriVaryReply) HandleRequest(w http.ResponseWriter, r *http.Request) []error {
	for uri, def := range s.variants {
		uri = strings.TrimLeft(uri, "/")
		if s.basePath+uri == r.URL.Path {
			return def.Execute(w, r)
		}
	}
	return unhandledRequestError(r)
}

func (s *uriVaryReply) ResetRunningContext() {
	for _, def := range s.variants {
		def.ResetRunningContext()
	}
}

func (s *uriVaryReply) EndRunningContext() []error {
	var errs []error
	for _, def := range s.variants {
		errs = append(errs, def.EndRunningContext()...)
	}
	return errs
}

type methodVaryReply struct {
	variants map[string]*Definition
}

func NewMethodVaryReply(variants map[string]*Definition) ReplyStrategy {
	return &methodVaryReply{
		variants: variants,
	}
}

func (s *methodVaryReply) HandleRequest(w http.ResponseWriter, r *http.Request) []error {
	for method, def := range s.variants {
		if strings.EqualFold(r.Method, method) {
			return def.Execute(w, r)
		}
	}
	return unhandledRequestError(r)
}

func (s *methodVaryReply) ResetRunningContext() {
	for _, def := range s.variants {
		def.ResetRunningContext()
	}
}

func (s *methodVaryReply) EndRunningContext() []error {
	var errs []error
	for _, def := range s.variants {
		errs = append(errs, def.EndRunningContext()...)
	}
	return errs
}

func NewSequentialReply(strategies []*Definition) ReplyStrategy {
	return &sequentialReply{
		sequence: strategies,
	}
}

type sequentialReply struct {
	mutex    sync.Mutex
	count    int
	sequence []*Definition
}

func (s *sequentialReply) ResetRunningContext() {
	s.mutex.Lock()
	s.count = 0
	s.mutex.Unlock()
	for _, def := range s.sequence {
		def.ResetRunningContext()
	}
}

func (s *sequentialReply) EndRunningContext() []error {
	var errs []error
	for _, def := range s.sequence {
		errs = append(errs, def.EndRunningContext()...)
	}
	return errs
}

func (s *sequentialReply) HandleRequest(w http.ResponseWriter, r *http.Request) []error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	// out of bounds, url requested more times than sequence length
	if s.count >= len(s.sequence) {
		return unhandledRequestError(r)
	}
	def := s.sequence[s.count]
	s.count++
	return def.Execute(w, r)
}

type basedOnRequestReply struct {
	mutex    sync.Mutex
	variants []*Definition
}

func newBasedOnRequestReply(variants []*Definition) ReplyStrategy {
	return &basedOnRequestReply{
		variants: variants,
	}
}

func (s *basedOnRequestReply) HandleRequest(w http.ResponseWriter, r *http.Request) []error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	var errors []error
	for _, def := range s.variants {
		errs := verifyRequestConstraints(def.requestConstraints, r)
		if errs == nil {
			return def.ExecuteWithoutVerifying(w, r)
		}
		errors = append(errors, errs...)
	}
	return append(errors, unhandledRequestError(r)...)
}

func (s *basedOnRequestReply) ResetRunningContext() {
	for _, def := range s.variants {
		def.ResetRunningContext()
	}
}

func (s *basedOnRequestReply) EndRunningContext() []error {
	var errs []error
	for _, def := range s.variants {
		errs = append(errs, def.EndRunningContext()...)
	}
	return errs
}
