package mocks

import (
	"errors"
	"fmt"
	"net/http"
)

type Loader struct {
	mocks *Mocks
}

func NewLoader(mocks *Mocks) *Loader {
	return &Loader{
		mocks: mocks,
	}
}

func (l *Loader) Load(mocksDefinition map[string]interface{}) error {
	// prevent deriving the definition from previous test
	l.mocks.ResetDefinitions()
	for serviceName, definition := range mocksDefinition {
		service := l.mocks.Service(serviceName)
		if service == nil {
			return fmt.Errorf("service mock not defined: %s", serviceName)
		}
		def, err := l.loadDefinition("$", definition)
		if err != nil {
			return fmt.Errorf("unable to load definition for %s: %v", serviceName, err)
		}
		// load the definition into the mock
		service.SetDefinition(def)
	}
	return nil
}

func (l *Loader) loadDefinition(path string, rawDef interface{}) (*definition, error) {
	def, ok := rawDef.(map[interface{}]interface{})
	if !ok {
		return nil, fmt.Errorf("at path %s: definition must be key-values", path)
	}

	// load request constraints
	var requestConstraints []verifier
	if constraints, ok := def["requestConstraints"]; ok {
		constraints, ok := constraints.([]interface{})
		if !ok || len(constraints) == 0 {
			return nil, fmt.Errorf("at path %s: `requestConstraints` requires array", path)
		}
		requestConstraints = make([]verifier, len(constraints))
		for i, c := range constraints {
			constraint, err := l.loadConstraint(c)
			if err != nil {
				return nil, fmt.Errorf("at path %s: unable to load constraint %d: %v", path, i+1, err)
			}
			requestConstraints[i] = constraint
		}
	}

	ak := []string{
		"requestConstraints",
		"strategy",
		"calls",
	}

	// load reply strategy
	var strategyName string
	s, ok := def["strategy"]
	if ok {
		strategyName, ok = s.(string)
	}
	if !ok {
		return nil, fmt.Errorf("at path %s: requires `strategy` key on root level", path)
	}
	replyStrategy, err := l.loadStrategy(path+"."+strategyName, strategyName, def, &ak)
	if err != nil {
		return nil, err
	}

	callsConstraint := callsNoConstraint
	if _, ok = def["calls"]; ok {
		if value, ok := def["calls"].(int); ok {
			callsConstraint = value
		}
	}

	if err := validateMapKeys(def, ak...); err != nil {
		return nil, err
	}

	return newDefinition(path, requestConstraints, replyStrategy, callsConstraint), nil
}

func (l *Loader) loadStrategy(path, strategyName string, definition map[interface{}]interface{}, ak *[]string) (replyStrategy, error) {
	switch strategyName {
	case "nop":
		return &nopReply{}, nil
	case "uriVary":
		*ak = append(*ak, "basePath", "uris")
		return l.loadUriVaryStrategy(path, definition)
	case "methodVary":
		*ak = append(*ak, "methods")
		return l.loadMethodVaryStrategy(path, definition)
	case "file":
		*ak = append(*ak, "filename", "statusCode", "headers")
		return l.loadFileStrategy(path, definition)
	case "constant":
		*ak = append(*ak, "body", "statusCode", "headers")
		return l.loadConstantStrategy(path, definition)
	default:
		return nil, fmt.Errorf("unknown strategy: %s", strategyName)
	}
}

func (l *Loader) loadUriVaryStrategy(path string, def map[interface{}]interface{}) (replyStrategy, error) {
	var basePath string
	if b, ok := def["basePath"]; ok {
		basePath = b.(string)
	}
	var uris map[string]*definition
	if u, ok := def["uris"]; ok {
		urisMap, ok := u.(map[interface{}]interface{})
		if !ok {
			return nil, errors.New("`uriVary` requires map under `uris` key")
		}
		uris = make(map[string]*definition, len(urisMap))
		for uri, v := range urisMap {
			def, err := l.loadDefinition(path+"."+uri.(string), v)
			if err != nil {
				return nil, err
			}
			uris[uri.(string)] = def
		}
	}
	return newUriVaryReply(basePath, uris), nil
}

func (l *Loader) loadMethodVaryStrategy(path string, def map[interface{}]interface{}) (replyStrategy, error) {
	var methods map[string]*definition
	if u, ok := def["methods"]; ok {
		methodsMap, ok := u.(map[interface{}]interface{})
		if !ok {
			return nil, errors.New("`methodVary` requires map under `methods` key")
		}
		methods = make(map[string]*definition, len(methodsMap))
		for method, v := range methodsMap {
			def, err := l.loadDefinition(path+"."+method.(string), v)
			if err != nil {
				return nil, err
			}
			methods[method.(string)] = def
		}
	}
	return newMethodVaryReply(methods), nil
}

func (l *Loader) loadFileStrategy(path string, def map[interface{}]interface{}) (replyStrategy, error) {
	f, ok := def["filename"]
	if !ok {
		return nil, errors.New("`file` requires `filename` key")
	}
	filename, ok := f.(string)
	if !ok {
		return nil, errors.New("`filename` must be string")
	}
	statusCode := http.StatusOK
	if c, ok := def["statusCode"]; ok {
		statusCode = c.(int)
	}
	headers, err := l.loadHeaders(def)
	if err != nil {
		return nil, err
	}
	return newFileReplyWithCode(filename, statusCode, headers), nil
}

func (l *Loader) loadConstantStrategy(path string, def map[interface{}]interface{}) (replyStrategy, error) {
	c, ok := def["body"]
	if !ok {
		return nil, errors.New("`constant` requires `body` key")
	}
	body, ok := c.(string)
	if !ok {
		return nil, errors.New("`body` must be string")
	}
	statusCode := http.StatusOK
	if c, ok := def["statusCode"]; ok {
		statusCode = c.(int)
	}
	headers, err := l.loadHeaders(def)
	if err != nil {
		return nil, err
	}
	return newConstantReplyWithCode([]byte(body), statusCode, headers), nil
}

func (l *Loader) loadHeaders(def map[interface{}]interface{}) (map[string]string, error) {
	var headers map[string]string
	if h, ok := def["headers"]; ok {
		hMap, ok := h.(map[interface{}]interface{})
		if !ok {
			return nil, errors.New("`headers` must be a map")
		}
		headers = make(map[string]string, len(hMap))
		for k, v := range hMap {
			key, ok := k.(string)
			if !ok {
				return nil, errors.New("`headers` requires string keys")
			}
			value, ok := v.(string)
			if !ok {
				return nil, errors.New("`headers` requires string values")
			}
			headers[key] = value
		}
	}
	return headers, nil
}

func (l *Loader) loadConstraint(definition interface{}) (verifier, error) {
	def, ok := definition.(map[interface{}]interface{})
	if !ok {
		return nil, errors.New("must be map")
	}
	if _, ok := def["kind"]; !ok {
		return nil, errors.New("requires `kind` key")
	}
	kind, ok := def["kind"].(string)
	if !ok {
		return nil, errors.New("`kind` must be string")
	}
	ak := []string{"kind"}
	c, err := l.loadConstraintOfKind(kind, def, &ak)
	if err != nil {
		return nil, err
	}
	if err := validateMapKeys(def, ak...); err != nil {
		return nil, err
	}
	return c, nil
}

func (l *Loader) loadConstraintOfKind(kind string, def map[interface{}]interface{}, ak *[]string) (verifier, error) {
	switch kind {
	case "nop":
		return &nopConstraint{}, nil
	case "bodyMatchesJSON":
		*ak = append(*ak, "body")
		return l.loadBodyMatchesJSONConstraint(def)
	case "queryMatches":
		*ak = append(*ak, "expectedQuery")
		return l.loadQueryMatchesConstraint(def)
	case "methodIsGET":
		return &methodConstraint{method: "GET"}, nil
	case "methodIsPOST":
		return &methodConstraint{method: "POST"}, nil
	case "methodIs":
		*ak = append(*ak, "method")
		return l.loadMethodIsConstraint(def)
	case "headerIs":
		*ak = append(*ak, "header", "value", "regexp")
		return l.loadHeaderIsConstraint(def)
	default:
		return nil, fmt.Errorf("unknown constraint: %s", kind)
	}
}

func (l *Loader) loadBodyMatchesJSONConstraint(def map[interface{}]interface{}) (verifier, error) {
	c, ok := def["body"]
	if !ok {
		return nil, errors.New("`bodyMatchesJSON` requires `body` key")
	}
	body, ok := c.(string)
	if !ok {
		return nil, errors.New("`body` must be string")
	}
	return newBodyMatchesJSONConstraint(body)
}

func (l *Loader) loadQueryMatchesConstraint(def map[interface{}]interface{}) (verifier, error) {
	c, ok := def["expectedQuery"]
	if !ok {
		return nil, errors.New("`queryMatches` requires `expectedQuery` key")
	}
	query, ok := c.(string)
	if !ok {
		return nil, errors.New("`expectedQuery` must be string")
	}
	return newQueryConstraint(query)
}

func (l *Loader) loadMethodIsConstraint(def map[interface{}]interface{}) (verifier, error) {
	c, ok := def["method"]
	if !ok {
		return nil, errors.New("`methodIs` requires `method` key")
	}
	method, ok := c.(string)
	if !ok || method == "" {
		return nil, errors.New("`method` must be string")
	}
	return &methodConstraint{method: method}, nil
}

func (l *Loader) loadHeaderIsConstraint(def map[interface{}]interface{}) (verifier, error) {
	c, ok := def["header"]
	if !ok {
		return nil, errors.New("`headerIs` requires `header` key")
	}
	header, ok := c.(string)
	if !ok || header == "" {
		return nil, errors.New("`header` must be string")
	}
	var valueStr, regexpStr string
	if value, ok := def["value"]; ok {
		valueStr, ok = value.(string)
		if !ok {
			return nil, errors.New("`value` must be string")
		}
	}
	if regexp, ok := def["regexp"]; ok {
		regexpStr, ok = regexp.(string)
		if !ok || regexp == "" {
			return nil, errors.New("`regexp` must be string")
		}
	}
	return newHeaderConstraint(header, valueStr, regexpStr)
}

func validateMapKeys(m map[interface{}]interface{}, allowedKeys ...string) error {
	for k, _ := range m {
		k := k.(string)
		found := false
		for _, ak := range allowedKeys {
			if ak == k {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("unexpected key %s (expecting %v)", k, allowedKeys)
		}
	}
	return nil
}
