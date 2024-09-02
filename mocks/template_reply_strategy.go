package mocks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"text/template"

	"github.com/Masterminds/sprig/v3"
)

type templateReply struct {
	replyBodyTemplate *template.Template
	statusCode        int
	headers           map[string]string
}

type templateRequest struct {
	r *http.Request

	jsonOnce sync.Once
	jsonData map[string]interface{}
}

func (tr *templateRequest) Query(key string) string {
	return tr.r.URL.Query().Get(key)
}

func (tr *templateRequest) Json() (map[string]interface{}, error) {
	var err error
	tr.jsonOnce.Do(func() {
		err = json.NewDecoder(tr.r.Body).Decode(&tr.jsonData)
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse request as Json: %w", err)
	}

	return tr.jsonData, nil
}

func newTemplateReply(content string, statusCode int, headers map[string]string) (ReplyStrategy, error) {
	tmpl, err := template.New("").Funcs(sprig.GenericFuncMap()).Parse(content)
	if err != nil {
		return nil, fmt.Errorf("template syntax error: %w", err)
	}

	strategy := &templateReply{
		replyBodyTemplate: tmpl,
		statusCode:        statusCode,
		headers:           headers,
	}

	return strategy, nil
}

func (s *templateReply) executeResponseTemplate(r *http.Request) (string, error) {
	ctx := map[string]*templateRequest{
		"request": {
			r: r,
		},
	}

	reply := bytes.NewBuffer(nil)
	if err := s.replyBodyTemplate.Execute(reply, ctx); err != nil {
		return "", fmt.Errorf("template mock error: %w", err)
	}

	return reply.String(), nil
}

func (s *templateReply) HandleRequest(w http.ResponseWriter, r *http.Request) []error {
	for k, v := range s.headers {
		w.Header().Add(k, v)
	}

	responseBody, err := s.executeResponseTemplate(r)
	if err != nil {
		return []error{err}
	}

	w.WriteHeader(s.statusCode)
	w.Write([]byte(responseBody)) // nolint:errcheck

	return nil
}
