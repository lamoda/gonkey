package response_schema

import (
	"encoding/json"
	"strings"

	"github.com/keyclaim/gonkey/checker"
	"github.com/keyclaim/gonkey/models"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/loads"
	"github.com/go-openapi/spec"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/validate"
)

type ResponseSchemaChecker struct {
	checker.CheckerInterface

	swagger *spec.Swagger
}

func NewChecker(specLocation string) checker.CheckerInterface {
	document, err := loads.Spec(specLocation)
	if err != nil {
		return nil
	}
	document, err = document.Expanded()
	if err != nil {
		return nil
	}
	return &ResponseSchemaChecker{
		swagger: document.Spec(),
	}
}

func (c *ResponseSchemaChecker) Check(t models.TestInterface, result *models.Result) ([]error, error) {
	// decode actual body
	var actual interface{}
	if err := json.Unmarshal([]byte(result.ResponseBody), &actual); err != nil {
		return nil, err
	}

	errs := validateResponseAgainstSwagger(
		t.Path(),
		t.GetMethod(),
		result.ResponseStatusCode,
		actual,
		c.swagger,
	)
	return errs, nil
}

func validateResponseAgainstSwagger(path, method string, statusCode int, response interface{}, swagger *spec.Swagger) []error {
	var errs []error
	swaggerResponse := findResponse(swagger, path, method, statusCode)
	if swaggerResponse == nil {
		return errs
	}
	if len(swaggerResponse.Headers) > 0 {
		err := validateHeaders(response, swaggerResponse.Headers)
		if err != nil {
			errs = append(errs, err)
		}
	}
	err := validate.AgainstSchema(swaggerResponse.Schema, response, strfmt.Default)
	if err != nil {
		if compositeError, ok := err.(*errors.CompositeError); ok {
			errs = append(errs, compositeError.Errors...)
		} else {
			errs = append(errs, err)
		}
	}
	return errs
}

func findResponse(swagger *spec.Swagger, testPath, testMethod string, statusCode int) *spec.Response {
	if swagger.SwaggerProps.Paths.Paths == nil {
		return nil
	}
	var operation *spec.Operation
	for path, p := range swagger.SwaggerProps.Paths.Paths {
		if strings.ToLower(testPath) == strings.ToLower(swagger.BasePath+path) {
			if operation = methodToOperation(p, testMethod); operation != nil {
				break
			}
		}
	}
	if operation == nil {
		return nil
	}
	var swaggerResponse *spec.Response
	if resp, ok := operation.Responses.StatusCodeResponses[statusCode]; ok {
		swaggerResponse = &resp
	} else {
		swaggerResponse = operation.Responses.Default
	}
	return swaggerResponse
}

func validateHeaders(response interface{}, headers map[string]spec.Header) error {
	//for name, header := range headers {
	//	v := validate.NewHeaderValidator(name, &header, strfmt.Default)
	//	v.Validate(&spec.Header{})
	//}
	return nil
}

func methodToOperation(p spec.PathItem, method string) *spec.Operation {
	method = strings.ToUpper(method)
	switch method {
	case "GET":
		return p.Get
	case "PUT":
		return p.Put
	case "POST":
		return p.Post
	case "DELETE":
		return p.Delete
	case "OPTIONS":
		return p.Options
	case "HEAD":
		return p.Head
	case "PATCH":
		return p.Patch
	default:
		return nil
	}
}
