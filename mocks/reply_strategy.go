package mocks

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"strings"
	"sync"
)

type replyStrategy interface {
	HandleRequest(w http.ResponseWriter, r *http.Request) []error
}

type contextAwareStrategy interface {
	ResetRunningContext()
	EndRunningContext() []error
}

type constantReply struct {
	replyStrategy

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

func newFileReplyWithCode(filename string, statusCode int, headers map[string]string) (replyStrategy, error) {
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

func newConstantReplyWithCode(content []byte, statusCode int, headers map[string]string) replyStrategy {
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

type failReply struct {
}

func (s *failReply) HandleRequest(w http.ResponseWriter, r *http.Request) []error {
	return unhandledRequestError(r)
}

type nopReply struct {
	replyStrategy
}

func (s *nopReply) HandleRequest(w http.ResponseWriter, r *http.Request) []error {
	w.WriteHeader(http.StatusNoContent)
	return nil
}

type uriVaryReply struct {
	replyStrategy
	contextAwareStrategy

	basePath string
	variants map[string]*definition
}

func newUriVaryReply(basePath string, variants map[string]*definition) replyStrategy {
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
	replyStrategy
	contextAwareStrategy

	variants map[string]*definition
}

func newMethodVaryReply(variants map[string]*definition) replyStrategy {
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

func newSequentialReply(strategies []*definition) replyStrategy {
	return &sequentialReply{
		sequence: strategies,
	}
}

type sequentialReply struct {
	sync.Mutex
	count    int
	sequence []*definition
}

func (s *sequentialReply) ResetRunningContext() {
	s.Lock()
	s.count = 0
	s.Unlock()
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
	s.Lock()
	defer s.Unlock()
	// out of bounds, url requested more times than sequence length
	if s.count >= len(s.sequence) {
		return unhandledRequestError(r)
	}
	def := s.sequence[s.count]
	s.count++
	return def.Execute(w, r)
}
