package mocks

import (
	"io/ioutil"
	"net/http"
	"strings"
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

func newFileReplyWithCode(filename string, statusCode int, headers map[string]string) replyStrategy {
	content, _ := ioutil.ReadFile(filename)
	r := &constantReply{
		replyBody:  content,
		statusCode: statusCode,
		headers:    headers,
	}
	return r
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
	w.WriteHeader(http.StatusNotFound)
	return nil
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
	w.WriteHeader(http.StatusMethodNotAllowed)
	return nil
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
