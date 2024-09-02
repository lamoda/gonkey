package compare

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
)

func makeErrorString(path, msg string, expected, actual interface{}) string {
	return fmt.Sprintf(
		"at path %s %s:\n     expected: %s\n       actual: %s",
		color.CyanString(path),
		msg,
		color.GreenString("%v", expected),
		color.RedString("%v", actual),
	)
}

func TestCompareNils(t *testing.T) {
	errors := Compare(nil, nil, Params{})
	if len(errors) != 0 {
		t.Error(
			"must return no errors",
			fmt.Sprintf("got result: %v", errors),
		)
		t.Fail()
	}
}

func TestCompareNilWithNonNil(t *testing.T) {
	errors := Compare("", nil, Params{})
	if errors[0].Error() != makeErrorString("$", "types do not match", "string", "nil") {
		t.Error(
			"must return one error",
			fmt.Sprintf("got result: %v", errors),
		)
		t.Fail()
	}
}

func TestCompareEqualStrings(t *testing.T) {
	errors := Compare("1", "1", Params{})
	if len(errors) != 0 {
		t.Error(
			"must return no errors",
			fmt.Sprintf("got result: %v", errors),
		)
		t.Fail()
	}
}

func TestCompareDifferStrings(t *testing.T) {
	errors := Compare("1", "2", Params{})
	if errors[0].Error() != makeErrorString("$", "values do not match", 1, 2) {
		t.Error(
			"must return one error",
			fmt.Sprintf("got result: %v", errors),
		)
		t.Fail()
	}
}

func TestCompareEqualIntegers(t *testing.T) {
	errors := Compare(1, 1, Params{})
	if len(errors) != 0 {
		t.Error(
			"must return no errors",
			fmt.Sprintf("got result: %v", errors),
		)
		t.Fail()
	}
}

func TestCompareDifferIntegers(t *testing.T) {
	errors := Compare(1, 2, Params{})
	if errors[0].Error() != makeErrorString("$", "values do not match", 1, 2) {
		t.Error(
			"must return one error",
			fmt.Sprintf("got result: %v", errors),
		)
		t.Fail()
	}
}

func TestCheckRegexMach(t *testing.T) {
	errors := Compare("$matchRegexp(x.+z)", "xyyyz", Params{})
	if len(errors) != 0 {
		t.Error(
			"must return no errors",
			fmt.Sprintf("got result: %v", errors),
		)
		t.Fail()
	}
}

func TestCheckRegexNotMach(t *testing.T) {
	errors := Compare("$matchRegexp(x.+z)", "ayyyb", Params{})
	if errors[0].Error() != makeErrorString("$",
		"value does not match regex", "$matchRegexp(x.+z)", "ayyyb") {
		t.Error(
			"must return one error",
			fmt.Sprintf("got result: %v", errors),
		)
		t.Fail()
	}
}

func TestCheckRegexCantCompile(t *testing.T) {
	errors := Compare("$matchRegexp((?x))", "2", Params{})
	if errors[0].Error() != makeErrorString("$", "can not compile regex", nil, "error") {
		t.Error(
			"must return one error",
			fmt.Sprintf("got result: %v", errors),
		)
		t.Fail()
	}
}

func TestCompareEqualArrays(t *testing.T) {
	array1 := []string{"1", "2"}
	array2 := []string{"1", "2"}
	errors := Compare(array1, array2, Params{})
	if len(errors) != 0 {
		t.Error(
			"must return no errors",
			fmt.Sprintf("got result: %v", errors),
		)
		t.Fail()
	}
}

func TestCompareEqualArraysWithDifferentElementsOrder(t *testing.T) {
	array1 := []string{"1", "2"}
	array2 := []string{"2", "1"}
	errors := Compare(array1, array2, Params{IgnoreArraysOrdering: true})
	if len(errors) != 0 {
		t.Error(
			"must return no errors",
			fmt.Sprintf("got result: %v", errors),
		)
		t.Fail()
	}
}

func TestCompareArraysDifferLengths(t *testing.T) {
	array1 := []string{"1", "2", "3"}
	array2 := []string{"1", "2"}
	errors := Compare(array1, array2, Params{})
	if errors[0].Error() != makeErrorString("$", "array lengths do not match", 3, 2) {
		t.Error(
			"must return one error",
			fmt.Sprintf("got result: %v", errors),
		)
		t.Fail()
	}
}

func TestCompareDifferArrays(t *testing.T) {
	array1 := []string{"1", "2"}
	array2 := []string{"1", "3"}
	errors := Compare(array1, array2, Params{})
	if errors[0].Error() != makeErrorString("$[1]", "values do not match", 2, 3) {
		t.Error(
			"must return one error",
			fmt.Sprintf("got result: %v", errors),
		)
		t.Fail()
	}
}

func TestCompareArraysFewErrors(t *testing.T) {
	array1 := []string{"1", "2", "3"}
	array2 := []string{"1", "3", "4"}
	errors := Compare(array1, array2, Params{})
	assert.Len(t, errors, 2)
}

func TestCompareNestedEqualArrays(t *testing.T) {
	array1 := [][]string{{"1", "2"}, {"3", "4"}}
	array2 := [][]string{{"1", "2"}, {"3", "4"}}
	errors := Compare(array1, array2, Params{})
	if len(errors) != 0 {
		t.Error(
			"must return no errors",
			fmt.Sprintf("got result: %v", errors),
		)
		t.Fail()
	}
}

func TestCompareNestedDifferArrays(t *testing.T) {
	array1 := [][]string{{"1", "2"}, {"3", "4"}}
	array2 := [][]string{{"1", "2"}, {"3", "5"}}
	errors := Compare(array1, array2, Params{})
	if errors[0].Error() != makeErrorString("$[1][1]", "values do not match", 4, 5) {
		t.Error(
			"must return one error",
			fmt.Sprintf("got result: %v", errors),
		)
		t.Fail()
	}
}

func TestCompareArraysWithRegex(t *testing.T) {
	arrayExpected := []string{"2", "$matchRegexp(x.+z)"}
	arrayActual := []string{"2", "xyyyz"}

	errors := Compare(arrayExpected, arrayActual, Params{})
	if len(errors) != 0 {
		t.Error(
			"must return no errors",
			fmt.Sprintf("got result: %v", errors),
		)
		t.Fail()
	}
}

func TestCompareArraysWithRegexMixedTypes(t *testing.T) {
	arrayExpected := []string{"2", "$matchRegexp([0-9]+)"}
	arrayActual := []interface{}{"2", 123}

	errors := Compare(arrayExpected, arrayActual, Params{})
	if len(errors) != 0 {
		t.Error(
			"must return no errors",
			fmt.Sprintf("got result: %v", errors),
		)
		t.Fail()
	}
}

func TestCompareArraysWithRegexNotMatch(t *testing.T) {
	arrayExpected := []string{"2", "$matchRegexp(x.+z)"}
	arrayActual := []string{"2", "ayyyb"}

	errors := Compare(arrayExpected, arrayActual, Params{})
	expectedErrors := makeErrorString("$[1]",
		"value does not match regex", "$matchRegexp(x.+z)", "ayyyb")
	if errors[0].Error() != expectedErrors {
		t.Error(
			"must return one error",
			fmt.Sprintf("got result: %v", errors),
		)
		t.Fail()
	}
}

func TestCompareEqualMaps(t *testing.T) {
	array1 := map[string]string{"a": "1", "b": "2"}
	array2 := map[string]string{"a": "1", "b": "2"}
	errors := Compare(array1, array2, Params{})
	if len(errors) != 0 {
		t.Error(
			"must return no errors",
			fmt.Sprintf("got result: %v", errors),
		)
		t.Fail()
	}
}

func TestCompareMapsWithRegex(t *testing.T) {
	mapExpected := map[string]string{"a": "1", "b": "$matchRegexp(x.+z)"}
	mapActual := map[string]string{"a": "1", "b": "xyyyz"}

	errors := Compare(mapExpected, mapActual, Params{})
	if len(errors) != 0 {
		t.Error(
			"must return no errors",
			fmt.Sprintf("got result: %v", errors),
		)
		t.Fail()
	}
}

func TestCompareMapsWithRegexNotMatch(t *testing.T) {
	mapExpected := map[string]string{"a": "1", "b": "$matchRegexp(x.+z)"}
	mapActual := map[string]string{"a": "1", "b": "ayyyb"}

	errors := Compare(mapExpected, mapActual, Params{})
	expectedErrors := makeErrorString("$.b", "value does not match regex", "$matchRegexp(x.+z)", "ayyyb")

	if errors[0].Error() != expectedErrors {
		t.Error(
			"must return one error",
			fmt.Sprintf("got result: %v", errors),
		)
		t.Fail()
	}
}

func TestCompareEqualMapsWithExtraFields(t *testing.T) {
	array1 := map[string]string{"a": "1", "b": "2"}
	array2 := map[string]string{"a": "1", "b": "2", "c": "3"}
	errors := Compare(array1, array2, Params{})
	if len(errors) != 0 {
		t.Error(
			"must return no errors",
			fmt.Sprintf("got result: %v", errors),
		)
		t.Fail()
	}
}

func TestCompareEqualMapsWithExtraFieldsCheckingEnabled(t *testing.T) {
	array1 := map[string]string{"a": "1", "b": "2"}
	array2 := map[string]string{"a": "1", "b": "2", "c": "3"}
	errors := Compare(array1, array2, Params{DisallowExtraFields: true})
	if errors[0].Error() != makeErrorString("$", "map lengths do not match", 2, 3) {
		t.Error(
			"must return one error",
			fmt.Sprintf("got result: %v", errors),
		)
		t.Fail()
	}
}

func TestCompareEqualMapsWithDifferentKeysOrder(t *testing.T) {
	array1 := map[string]string{"a": "1", "b": "2"}
	array2 := map[string]string{"b": "2", "a": "1"}
	errors := Compare(array1, array2, Params{})
	if len(errors) != 0 {
		t.Error(
			"must return no errors",
			fmt.Sprintf("got result: %v", errors),
		)
		t.Fail()
	}
}

func TestCompareMapsWithDifferentKeys(t *testing.T) {
	array1 := map[string]string{"a": "1", "b": "2"}
	array2 := map[string]string{"a": "1", "c": "2"}
	errors := Compare(array1, array2, Params{})
	expectedErr := makeErrorString("$", "key is missing", "b", "<missing>")
	if errors[0].Error() != expectedErr {
		t.Error(
			"must return one error",
			fmt.Sprintf("got result: %v", errors),
		)
		t.Fail()
	}
}

func TestCompareMapsWithDifferentValues(t *testing.T) {
	array1 := map[string]string{"a": "1", "b": "2"}
	array2 := map[string]string{"a": "1", "b": "3"}
	errors := Compare(array1, array2, Params{})
	if errors[0].Error() != makeErrorString("$.b", "values do not match", 2, 3) {
		t.Error(
			"must return one error",
			fmt.Sprintf("got result: %v", errors),
		)
		t.Fail()
	}
}

func TestCompareMapsWithFewErrors(t *testing.T) {
	array1 := map[string]string{"a": "1", "b": "2", "c": "5"}
	array2 := map[string]string{"a": "1", "b": "3", "d": "4"}
	errors := Compare(array1, array2, Params{})
	assert.Len(t, errors, 2)
}

func TestCompareEqualNestedMaps(t *testing.T) {
	array1 := map[string]map[string]string{"a": {"i": "3", "j": "4"}, "b": {"k": "5", "l": "6"}}
	array2 := map[string]map[string]string{"a": {"i": "3", "j": "4"}, "b": {"k": "5", "l": "6"}}
	errors := Compare(array1, array2, Params{})
	if len(errors) != 0 {
		t.Error(
			"must return no errors",
			fmt.Sprintf("got result: %v", errors),
		)
		t.Fail()
	}
}

func TestCompareNestedMapsWithDifferentKeys(t *testing.T) {
	array1 := map[string]map[string]string{"a": {"i": "3", "j": "4"}, "b": {"k": "5", "l": "6"}}
	array2 := map[string]map[string]string{"a": {"i": "3", "j": "4"}, "b": {"l": "6"}}
	errors := Compare(array1, array2, Params{})
	expectedErr := makeErrorString("$.b", "key is missing", "k", "<missing>")
	if errors[0].Error() != expectedErr {
		t.Error(
			"must return one error",
			fmt.Sprintf("got result: %v", errors),
		)
		t.Fail()
	}
}

func TestCompareNestedMapsWithDifferentValues(t *testing.T) {
	array1 := map[string]map[string]string{"a": {"i": "3", "j": "4"}, "b": {"k": "5", "l": "6"}}
	array2 := map[string]map[string]string{"a": {"i": "3", "j": "4"}, "b": {"k": "5", "l": "7"}}
	errors := Compare(array1, array2, Params{})
	if errors[0].Error() != makeErrorString("$.b.l", "values do not match", 6, 7) {
		t.Error(
			"must return one error",
			fmt.Sprintf("got result: %v", errors),
		)
		t.Fail()
	}
}

func TestCompareEqualJsonScalars(t *testing.T) {
	var json1, json2 interface{}
	_ = json.Unmarshal([]byte("1"), &json1)
	_ = json.Unmarshal([]byte("1"), &json2)
	errors := Compare(json1, json2, Params{})
	if len(errors) != 0 {
		t.Error(
			"must return no errors",
			fmt.Sprintf("got result: %v", errors),
		)
		t.Fail()
	}
}

func TestCompareDifferJsonScalars(t *testing.T) {
	var json1, json2 interface{}
	_ = json.Unmarshal([]byte("1"), &json1)
	_ = json.Unmarshal([]byte("2"), &json2)
	errors := Compare(json1, json2, Params{})
	if errors[0].Error() != makeErrorString("$", "values do not match", 1, 2) {
		t.Error(
			"must return one error",
			fmt.Sprintf("got result: %v", errors),
		)
		t.Fail()
	}
}

var expectedArrayJson = `
{
  "data":[
    {"name": "n111"},
    {"name": "n222"},
    {"name": "n333"}
  ]
}
`

var actualArrayJson = `
{
  "data": [
    {"message": "m555", "name": "n333"},
    {"message": "m777", "name": "n111"},
    {"message": "m999","name": "n222"}
  ]
}
`

func TestCompareEqualArraysWithIgnoreArraysOrdering(t *testing.T) {
	var json1, json2 interface{}
	_ = json.Unmarshal([]byte(expectedArrayJson), &json1)
	_ = json.Unmarshal([]byte(actualArrayJson), &json2)
	errors := Compare(json1, json2, Params{
		IgnoreArraysOrdering: true,
	})
	if len(errors) != 0 {
		t.Error(
			"must return no errors",
			fmt.Sprintf("got result: %v", errors),
		)
		t.Fail()
	}
}

func TestCompareEqualComplexJson(t *testing.T) {
	var json1, json2 interface{}
	_ = json.Unmarshal([]byte(complexJson1), &json1)
	_ = json.Unmarshal([]byte(complexJson1), &json2) // compare json with same json
	errors := Compare(json1, json2, Params{})
	if len(errors) != 0 {
		t.Error(
			"must return no errors",
			fmt.Sprintf("got result: %v", errors),
		)
		t.Fail()
	}
}

func TestCompareDifferComplexJson(t *testing.T) {
	var json1, json2 interface{}
	_ = json.Unmarshal([]byte(complexJson1), &json1)
	_ = json.Unmarshal([]byte(complexJson2), &json2)
	errors := Compare(json1, json2, Params{})
	expectedErr := makeErrorString(
		"$.paths./api/get-delivery-info.get.parameters[2].$ref",
		"values do not match",
		"#/parameters/profile_id",
		"#/parameters/profile_id2",
	)
	if len(errors) == 0 || errors[0].Error() != expectedErr {
		t.Error(
			"must return one error",
			fmt.Sprintf("got result: %v", errors),
		)
		t.Fail()
	}
}

var complexJson1 = `
{
    "swagger": "2.0",
    "info": {
        "title": "LEOS.Delivery API",
        "contact": {
            "url": "https://confluence.lamoda.ru/display/DELY/"
        },
        "version": "1.0.0"
    },
    "produces": [
        "application/json"
    ],
    "paths": {
        "/api/get-methods": {
            "get": {
                "tags": [
                    "rest"
                ],
                "summary": "Get methods",
                "operationId": "getMethods",
                "parameters": [
                    {
                        "$ref": "#/parameters/zipcode"
                    },
                    {
                        "$ref": "#/parameters/aoid"
                    },
                    {
                        "$ref": "#/parameters/latitude"
                    },
                    {
                        "$ref": "#/parameters/longitude"
                    },
                    {
                        "$ref": "#/parameters/item_count"
                    },
                    {
                        "$ref": "#/parameters/cart_amount"
                    },
                    {
                        "$ref": "#/parameters/payment_type"
                    },
                    {
                        "$ref": "#/parameters/no_airfreight"
                    },
                    {
                        "$ref": "#/parameters/liquids"
                    },
                    {
                        "$ref": "#/parameters/checkout"
                    },
                    {
                        "$ref": "#/parameters/method_code"
                    },
                    {
                        "$ref": "#/parameters/company"
                    },
                    {
                        "$ref": "#/parameters/origin_date"
                    },
                    {
                        "$ref": "#/parameters/profile_id"
                    },
                    {
                        "$ref": "#/parameters/ignore_capacity"
                    },
                    {
                        "$ref": "#/parameters/ignore_cutoff"
                    },
                    {
                        "$ref": "#/parameters/reserved_interval_id"
                    },
                    {
                        "$ref": "#/parameters/original_delivery_date"
                    },
                    {
                        "$ref": "#/parameters/tag"
                    }
                ],
                "responses": {
                    "default": {
                        "description": "Success/error response",
                        "schema": {
                            "$ref": "#/definitions/GetMethodsResponse"
                        }
                    }
                }
            }
        },
        "/api/get-checkout-methods": {
            "get": {
                "tags": [
                    "rest"
                ],
                "summary": "Get checkout methods",
                "operationId": "getCheckoutMethods",
                "parameters": [
                    {
                        "$ref": "#/parameters/zipcode"
                    },
                    {
                        "$ref": "#/parameters/aoid"
                    },
                    {
                        "$ref": "#/parameters/zone_id"
                    },
                    {
                        "$ref": "#/parameters/latitude"
                    },
                    {
                        "$ref": "#/parameters/longitude"
                    },
                    {
                        "$ref": "#/parameters/item_count"
                    },
                    {
                        "$ref": "#/parameters/cart_amount"
                    },
                    {
                        "$ref": "#/parameters/payment_type"
                    },
                    {
                        "$ref": "#/parameters/no_airfreight"
                    },
                    {
                        "$ref": "#/parameters/liquids"
                    },
                    {
                        "$ref": "#/parameters/checkout"
                    },
                    {
                        "$ref": "#/parameters/method_code"
                    },
                    {
                        "$ref": "#/parameters/company"
                    },
                    {
                        "$ref": "#/parameters/origin_date"
                    },
                    {
                        "$ref": "#/parameters/profile_id"
                    },
                    {
                        "$ref": "#/parameters/ignore_capacity"
                    },
                    {
                        "$ref": "#/parameters/ignore_cutoff"
                    },
                    {
                        "$ref": "#/parameters/reserved_interval_id"
                    },
                    {
                        "$ref": "#/parameters/original_delivery_date"
                    },
                    {
                        "$ref": "#/parameters/tag"
                    }
                ],
                "responses": {
                    "default": {
                        "description": "Success/error response",
                        "schema": {
                            "$ref": "#/definitions/GetCheckoutMethodsResponse"
                        }
                    }
                }
            }
        },
        "/api/get-checkout-methods-multiple": {
            "get": {
                "tags": [
                    "rest"
                ],
                "summary": "Get checkout methods for multiple orders",
                "description": "Two-dimensional orders array defines parameters for each individual order.\nAt least one order have to be specified with one or more parameters.",
                "operationId": "getCheckoutMethodsMultiple",
                "parameters": [
                    {
                        "$ref": "#/parameters/zipcode"
                    },
                    {
                        "$ref": "#/parameters/aoid"
                    },
                    {
                        "$ref": "#/parameters/zone_id"
                    },
                    {
                        "$ref": "#/parameters/latitude"
                    },
                    {
                        "$ref": "#/parameters/longitude"
                    },
                    {
                        "$ref": "#/parameters/checkout"
                    },
                    {
                        "$ref": "#/parameters/method_code"
                    },
                    {
                        "$ref": "#/parameters/origin_date"
                    },
                    {
                        "$ref": "#/parameters/ignore_capacity"
                    },
                    {
                        "$ref": "#/parameters/ignore_cutoff"
                    },
                    {
                        "$ref": "#/parameters/orders_profile_id"
                    },
                    {
                        "$ref": "#/parameters/orders_item_count"
                    },
                    {
                        "$ref": "#/parameters/orders_cart_amount"
                    },
                    {
                        "$ref": "#/parameters/orders_no_airfreight"
                    },
                    {
                        "$ref": "#/parameters/orders_payment_type"
                    },
                    {
                        "$ref": "#/parameters/tag"
                    }
                ],
                "responses": {
                    "default": {
                        "description": "Success/error response.\n<br />\nReturn list of orders in the same sequence as in request.",
                        "schema": {
                            "$ref": "#/definitions/GetCheckoutMethodsMultiple\\Response"
                        }
                    }
                }
            }
        },
        "/api/reserve": {
            "get": {
                "tags": [
                    "rest"
                ],
                "summary": "Reserve interval",
                "operationId": "reserve",
                "parameters": [
                    {
                        "$ref": "#/parameters/interval_id"
                    },
                    {
                        "$ref": "#/parameters/pallet_method_code"
                    },
                    {
                        "$ref": "#/parameters/profile_id"
                    },
                    {
                        "$ref": "#/parameters/force"
                    },
                    {
                        "$ref": "#/parameters/tag"
                    }
                ],
                "responses": {
                    "default": {
                        "description": "Success/error response",
                        "schema": {
                            "$ref": "#/definitions/ReserveResponse"
                        }
                    }
                }
            }
        },
        "/api/free": {
            "get": {
                "tags": [
                    "rest"
                ],
                "summary": "Free interval",
                "operationId": "free",
                "parameters": [
                    {
                        "$ref": "#/parameters/interval_id"
                    },
                    {
                        "$ref": "#/parameters/pallet_method_code"
                    },
                    {
                        "$ref": "#/parameters/profile_id"
                    },
                    {
                        "$ref": "#/parameters/tag"
                    }
                ],
                "responses": {
                    "default": {
                        "description": "Success/error response",
                        "schema": {
                            "$ref": "#/definitions/Response"
                        }
                    }
                }
            }
        },
        "/api/find-interval": {
            "get": {
                "tags": [
                    "rest"
                ],
                "summary": "Find interval within the same zone",
                "description": "Required either current_id or pair zone_id + service_level_type_code.",
                "operationId": "findInterval",
                "parameters": [
                    {
                        "$ref": "#/parameters/desired_start"
                    },
                    {
                        "$ref": "#/parameters/desired_end"
                    },
                    {
                        "$ref": "#/parameters/current_id"
                    },
                    {
                        "$ref": "#/parameters/zone_id"
                    },
                    {
                        "$ref": "#/parameters/service_level_type_code"
                    }
                ],
                "responses": {
                    "default": {
                        "description": "Success/error response",
                        "schema": {
                            "$ref": "#/definitions/FindIntervalResponse"
                        }
                    }
                }
            }
        },
        "/api/get-interval": {
            "get": {
                "tags": [
                    "rest"
                ],
                "summary": "Get interval",
                "operationId": "getInterval",
                "parameters": [
                    {
                        "$ref": "#/parameters/interval_id"
                    },
                    {
                        "$ref": "#/parameters/tag"
                    }
                ],
                "responses": {
                    "default": {
                        "description": "Success/error response",
                        "schema": {
                            "$ref": "#/definitions/GetIntervalResponse"
                        }
                    }
                }
            }
        },
        "/api/get-method-details": {
            "get": {
                "tags": [
                    "rest"
                ],
                "summary": "Get method details",
                "operationId": "getMethodDetails",
                "parameters": [
                    {
                        "$ref": "#/parameters/code"
                    },
                    {
                        "$ref": "#/parameters/join"
                    },
                    {
                        "$ref": "#/parameters/zone_id"
                    },
                    {
                        "$ref": "#/parameters/service_level"
                    },
                    {
                        "$ref": "#/parameters/zipcode"
                    },
                    {
                        "$ref": "#/parameters/aoid"
                    },
                    {
                        "$ref": "#/parameters/origin_date"
                    },
                    {
                        "$ref": "#/parameters/profile_id"
                    }
                ],
                "responses": {
                    "default": {
                        "description": "Success/error response",
                        "schema": {
                            "$ref": "#/definitions/GetMethodDetails\\Response"
                        }
                    }
                }
            }
        },
        "/api/get-method-details-multiple": {
            "get": {
                "tags": [
                    "rest"
                ],
                "summary": "Get method details multiple",
                "operationId": "getMethodDetailsMultiple",
                "parameters": [
                    {
                        "$ref": "#/parameters/get_method_details_multiple_code"
                    },
                    {
                        "$ref": "#/parameters/get_method_details_multiple_join"
                    },
                    {
                        "$ref": "#/parameters/get_method_details_multiple_zone_id"
                    },
                    {
                        "$ref": "#/parameters/get_method_details_multiple_service_level"
                    },
                    {
                        "$ref": "#/parameters/get_method_details_multiple_zipcode"
                    },
                    {
                        "$ref": "#/parameters/get_method_details_multiple_aoid"
                    },
                    {
                        "$ref": "#/parameters/get_method_details_multiple_origin_date"
                    },
                    {
                        "$ref": "#/parameters/profile_id"
                    }
                ],
                "responses": {
                    "default": {
                        "description": "Success/error response.\n<br />\nReturn list of methods in the same sequence as in request.",
                        "schema": {
                            "$ref": "#/definitions/GetMethodDetailsMultiple\\Response"
                        }
                    }
                }
            }
        },
        "/api/get-all-methods": {
            "get": {
                "tags": [
                    "rest"
                ],
                "summary": "Get all methods",
                "operationId": "getAllMethods",
                "responses": {
                    "default": {
                        "description": "Success/error response",
                        "schema": {
                            "$ref": "#/definitions/GetAllMethodsResponse"
                        }
                    }
                }
            }
        },
        "/api/get-pup-list": {
            "get": {
                "tags": [
                    "rest"
                ],
                "summary": "Get pup list",
                "operationId": "getPupList",
                "parameters": [
                    {
                        "$ref": "#/parameters/zipcode"
                    },
                    {
                        "$ref": "#/parameters/aoid"
                    },
                    {
                        "$ref": "#/parameters/latitude"
                    },
                    {
                        "$ref": "#/parameters/longitude"
                    },
                    {
                        "$ref": "#/parameters/item_count"
                    },
                    {
                        "$ref": "#/parameters/cart_amount"
                    },
                    {
                        "$ref": "#/parameters/payment_type"
                    },
                    {
                        "$ref": "#/parameters/no_airfreight"
                    },
                    {
                        "$ref": "#/parameters/liquids"
                    },
                    {
                        "$ref": "#/parameters/checkout"
                    },
                    {
                        "$ref": "#/parameters/method_code"
                    },
                    {
                        "$ref": "#/parameters/company"
                    },
                    {
                        "$ref": "#/parameters/origin_date"
                    },
                    {
                        "$ref": "#/parameters/profile_id"
                    },
                    {
                        "$ref": "#/parameters/ignore_capacity"
                    },
                    {
                        "$ref": "#/parameters/ignore_cutoff"
                    },
                    {
                        "$ref": "#/parameters/reserved_interval_id"
                    },
                    {
                        "$ref": "#/parameters/original_delivery_date"
                    },
                    {
                        "$ref": "#/parameters/tag"
                    }
                ],
                "responses": {
                    "default": {
                        "description": "Success/error response",
                        "schema": {
                            "$ref": "#/definitions/GetPupListResponse"
                        }
                    }
                }
            }
        },
        "/api/get-pup-details": {
            "get": {
                "tags": [
                    "rest"
                ],
                "summary": "Get pup details",
                "operationId": "getPupDetails",
                "parameters": [
                    {
                        "$ref": "#/parameters/pickup_id"
                    },
                    {
                        "$ref": "#/parameters/join"
                    },
                    {
                        "$ref": "#/parameters/zipcode"
                    },
                    {
                        "$ref": "#/parameters/aoid"
                    },
                    {
                        "$ref": "#/parameters/latitude"
                    },
                    {
                        "$ref": "#/parameters/longitude"
                    },
                    {
                        "$ref": "#/parameters/item_count"
                    },
                    {
                        "$ref": "#/parameters/cart_amount"
                    },
                    {
                        "$ref": "#/parameters/payment_type"
                    },
                    {
                        "$ref": "#/parameters/no_airfreight"
                    },
                    {
                        "$ref": "#/parameters/liquids"
                    },
                    {
                        "$ref": "#/parameters/checkout"
                    },
                    {
                        "$ref": "#/parameters/method_code"
                    },
                    {
                        "$ref": "#/parameters/company"
                    },
                    {
                        "$ref": "#/parameters/origin_date"
                    },
                    {
                        "$ref": "#/parameters/profile_id"
                    },
                    {
                        "$ref": "#/parameters/ignore_capacity"
                    },
                    {
                        "$ref": "#/parameters/ignore_cutoff"
                    },
                    {
                        "$ref": "#/parameters/reserved_interval_id"
                    },
                    {
                        "$ref": "#/parameters/original_delivery_date"
                    },
                    {
                        "$ref": "#/parameters/tag"
                    }
                ],
                "responses": {
                    "default": {
                        "description": "Success/error response",
                        "schema": {
                            "$ref": "#/definitions/GetPupDetailsResponse"
                        }
                    }
                }
            }
        },
        "/api/get-pickup": {
            "get": {
                "tags": [
                    "rest"
                ],
                "summary": "Get pickup",
                "operationId": "getPickup",
                "parameters": [
                    {
                        "$ref": "#/parameters/pickup_id"
                    },
                    {
                        "$ref": "#/parameters/join"
                    },
                    {
                        "$ref": "#/parameters/zipcode"
                    },
                    {
                        "$ref": "#/parameters/aoid"
                    },
                    {
                        "$ref": "#/parameters/profile_id"
                    }
                ],
                "responses": {
                    "default": {
                        "description": "Success/error response",
                        "schema": {
                            "$ref": "#/definitions/GetPickupResponse"
                        }
                    }
                }
            }
        },
        "/api/get-return-methods": {
            "get": {
                "tags": [
                    "rest"
                ],
                "summary": "Get return methods",
                "operationId": "getReturnMethods",
                "parameters": [
                    {
                        "$ref": "#/parameters/zipcode_required"
                    }
                ],
                "responses": {
                    "default": {
                        "description": "Success/error response",
                        "schema": {
                            "$ref": "#/definitions/GetReturnMethodsResponse"
                        }
                    }
                }
            }
        },
        "/api/get-zone-rate": {
            "get": {
                "tags": [
                    "rest"
                ],
                "summary": "Get zone rate",
                "description": "Either zipcode or zone_id parameter is required.\nIf zone_id is specified then it is used instead of zipcode to find required zone.",
                "operationId": "getZoneRate",
                "parameters": [
                    {
                        "$ref": "#/parameters/method_code_required"
                    },
                    {
                        "$ref": "#/parameters/zipcode"
                    },
                    {
                        "$ref": "#/parameters/zone_id"
                    },
                    {
                        "$ref": "#/parameters/weight"
                    },
                    {
                        "$ref": "#/parameters/insurance_sum"
                    }
                ],
                "responses": {
                    "default": {
                        "description": "Success/error response",
                        "schema": {
                            "$ref": "#/definitions/GetZoneRateResponse"
                        }
                    }
                }
            }
        },
        "/api/get-delivery-info": {
            "get": {
                "tags": [
                    "rest"
                ],
                "summary": "Get delivery info",
                "operationId": "getDeliveryInfo",
                "parameters": [
                    {
                        "$ref": "#/parameters/zipcode"
                    },
                    {
                        "$ref": "#/parameters/aoid"
                    },
                    {
                        "$ref": "#/parameters/profile_id"
                    },
                    {
                        "$ref": "#/parameters/no_airfreight"
                    }
                ],
                "responses": {
                    "default": {
                        "description": "Success/error response",
                        "schema": {
                            "$ref": "#/definitions/GetDeliveryInfoResponse"
                        }
                    }
                }
            }
        },
        "/api/get-delivery-info-short": {
            "get": {
                "tags": [
                    "rest"
                ],
                "summary": "Get short terms of delivery",
                "operationId": "getDeliveryInfoShort",
                "parameters": [
                    {
                        "$ref": "#/parameters/zipcode"
                    },
                    {
                        "$ref": "#/parameters/aoid"
                    },
                    {
                        "$ref": "#/parameters/profile_id"
                    },
                    {
                        "$ref": "#/parameters/no_airfreight"
                    }
                ],
                "responses": {
                    "default": {
                        "description": "Success/error response",
                        "schema": {
                            "$ref": "#/definitions/GetDeliveryInfoShortResponse"
                        }
                    }
                }
            }
        },
        "/api/get-zone-attributes": {
            "get": {
                "tags": [
                    "rest"
                ],
                "summary": "Get zone attributes",
                "operationId": "getZoneAttributes",
                "parameters": [
                    {
                        "$ref": "#/parameters/zone_id_required"
                    }
                ],
                "responses": {
                    "default": {
                        "description": "Success/error response",
                        "schema": {
                            "$ref": "#/definitions/GetZoneAttributesResponse"
                        }
                    }
                }
            }
        },
        "/api/create-pickups": {
            "post": {
                "tags": [
                    "rest"
                ],
                "summary": "Create pickups",
                "operationId": "createPickups",
                "parameters": [
                    {
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/CreatePickups\\Request"
                        }
                    }
                ],
                "responses": {
                    "default": {
                        "description": "Success/error response",
                        "schema": {
                            "$ref": "#/definitions/CreatePickups\\Response"
                        }
                    }
                }
            }
        },
        "/api/update-zone-rate": {
            "post": {
                "tags": [
                    "rest"
                ],
                "summary": "Update zone rate",
                "description": "Zone should have at least one search criteria: zone_id|pickup_id|pickup_external_id",
                "operationId": "updateZoneRate",
                "parameters": [
                    {
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/UpdateZoneRate\\Request"
                        }
                    }
                ],
                "responses": {
                    "default": {
                        "description": "Success/error response",
                        "schema": {
                            "$ref": "#/definitions/UpdateZoneRate\\Response"
                        }
                    }
                }
            }
        },
        "/api/import-profiles": {
            "post": {
                "tags": [
                    "rest"
                ],
                "summary": "Import profiles",
                "operationId": "importProfiles",
                "parameters": [
                    {
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "type": "object"
                        }
                    }
                ],
                "responses": {
                    "default": {
                        "description": "Success/error response",
                        "schema": {
                            "$ref": "#/definitions/Response"
                        }
                    }
                }
            }
        },
        "/jsonrpc/v1/delivery.get-info": {
            "get": {
                "tags": [
                    "json-rpc",
                    "delivery"
                ],
                "summary": "Get delivery info",
                "operationId": "delivery.getInfo",
                "parameters": [
                    {
                        "$ref": "#/parameters/jsonrpc_id"
                    },
                    {
                        "$ref": "#/parameters/zipcode"
                    },
                    {
                        "$ref": "#/parameters/aoid"
                    },
                    {
                        "$ref": "#/parameters/profile_id"
                    },
                    {
                        "$ref": "#/parameters/no_airfreight"
                    }
                ],
                "responses": {
                    "default": {
                        "description": "Success/error response",
                        "schema": {
                            "$ref": "#/definitions/JsonRpcDeliveryGetInfoResponse"
                        }
                    }
                }
            }
        },
        "/jsonrpc/v1/delivery.get-info-short": {
            "get": {
                "tags": [
                    "json-rpc",
                    "delivery"
                ],
                "summary": "Get short terms of delivery",
                "operationId": "delivery.getInfoShort",
                "parameters": [
                    {
                        "$ref": "#/parameters/jsonrpc_id"
                    },
                    {
                        "$ref": "#/parameters/zipcode"
                    },
                    {
                        "$ref": "#/parameters/aoid"
                    },
                    {
                        "$ref": "#/parameters/profile_id"
                    },
                    {
                        "$ref": "#/parameters/no_airfreight"
                    }
                ],
                "responses": {
                    "default": {
                        "description": "Success/error response",
                        "schema": {
                            "$ref": "#/definitions/JsonRpcDeliveryGetInfoShortResponse"
                        }
                    }
                }
            }
        },
        "/jsonrpc/v1/intervals.reserve": {
            "get": {
                "tags": [
                    "json-rpc",
                    "intervals"
                ],
                "summary": "Reserve interval",
                "operationId": "intervals.reserve",
                "parameters": [
                    {
                        "$ref": "#/parameters/jsonrpc_id"
                    },
                    {
                        "$ref": "#/parameters/interval_id"
                    },
                    {
                        "$ref": "#/parameters/pallet_method_code"
                    },
                    {
                        "$ref": "#/parameters/force"
                    },
                    {
                        "$ref": "#/parameters/tag"
                    }
                ],
                "responses": {
                    "default": {
                        "description": "Success/error response",
                        "schema": {
                            "$ref": "#/definitions/JsonRpcIntervalsReserveResponse"
                        }
                    }
                }
            }
        },
        "/jsonrpc/v1/intervals.free": {
            "get": {
                "tags": [
                    "json-rpc",
                    "intervals"
                ],
                "summary": "Free interval",
                "operationId": "intervals.free",
                "parameters": [
                    {
                        "$ref": "#/parameters/jsonrpc_id"
                    },
                    {
                        "$ref": "#/parameters/interval_id"
                    },
                    {
                        "$ref": "#/parameters/pallet_method_code"
                    },
                    {
                        "$ref": "#/parameters/tag"
                    }
                ],
                "responses": {
                    "default": {
                        "description": "Success/error response",
                        "schema": {
                            "$ref": "#/definitions/JsonRpcOkResponse"
                        }
                    }
                }
            }
        },
        "/jsonrpc/v1/intervals.get": {
            "get": {
                "tags": [
                    "json-rpc",
                    "intervals"
                ],
                "summary": "Get interval",
                "operationId": "intervals.get",
                "parameters": [
                    {
                        "$ref": "#/parameters/jsonrpc_id"
                    },
                    {
                        "$ref": "#/parameters/interval_id"
                    },
                    {
                        "$ref": "#/parameters/tag"
                    }
                ],
                "responses": {
                    "default": {
                        "description": "Success/error response",
                        "schema": {
                            "$ref": "#/definitions/JsonRpcIntervalsGetResponse"
                        }
                    }
                }
            }
        },
        "/jsonrpc/v1/intervals.find": {
            "get": {
                "tags": [
                    "json-rpc",
                    "intervals"
                ],
                "summary": "Find interval within the same zone",
                "description": "Required either current_id or pair zone_id + service_level_type_code.",
                "operationId": "intervals.find",
                "parameters": [
                    {
                        "$ref": "#/parameters/jsonrpc_id"
                    },
                    {
                        "$ref": "#/parameters/desired_start"
                    },
                    {
                        "$ref": "#/parameters/desired_end"
                    },
                    {
                        "$ref": "#/parameters/current_id"
                    },
                    {
                        "$ref": "#/parameters/zone_id"
                    },
                    {
                        "$ref": "#/parameters/service_level_type_code"
                    }
                ],
                "responses": {
                    "default": {
                        "description": "Success/error response",
                        "schema": {
                            "$ref": "#/definitions/JsonRpcIntervalsFindResponse"
                        }
                    }
                }
            }
        },
        "/jsonrpc/v1/methods.get": {
            "get": {
                "tags": [
                    "json-rpc",
                    "methods"
                ],
                "summary": "Get methods",
                "operationId": "methods.get",
                "parameters": [
                    {
                        "$ref": "#/parameters/jsonrpc_id"
                    },
                    {
                        "$ref": "#/parameters/zipcode"
                    },
                    {
                        "$ref": "#/parameters/aoid"
                    },
                    {
                        "$ref": "#/parameters/latitude"
                    },
                    {
                        "$ref": "#/parameters/longitude"
                    },
                    {
                        "$ref": "#/parameters/item_count"
                    },
                    {
                        "$ref": "#/parameters/cart_amount"
                    },
                    {
                        "$ref": "#/parameters/payment_type"
                    },
                    {
                        "$ref": "#/parameters/no_airfreight"
                    },
                    {
                        "$ref": "#/parameters/liquids"
                    },
                    {
                        "$ref": "#/parameters/checkout"
                    },
                    {
                        "$ref": "#/parameters/method_code"
                    },
                    {
                        "$ref": "#/parameters/company"
                    },
                    {
                        "$ref": "#/parameters/origin_date"
                    },
                    {
                        "$ref": "#/parameters/profile_id"
                    },
                    {
                        "$ref": "#/parameters/ignore_capacity"
                    },
                    {
                        "$ref": "#/parameters/ignore_cutoff"
                    },
                    {
                        "$ref": "#/parameters/reserved_interval_id"
                    },
                    {
                        "$ref": "#/parameters/original_delivery_date"
                    },
                    {
                        "$ref": "#/parameters/tag"
                    }
                ],
                "responses": {
                    "default": {
                        "description": "Success/error response",
                        "schema": {
                            "$ref": "#/definitions/JsonRpcMethodsGetResponse"
                        }
                    }
                }
            }
        },
        "/jsonrpc/v1/methods.get-checkout": {
            "get": {
                "tags": [
                    "json-rpc",
                    "methods"
                ],
                "summary": "Get checkout methods",
                "operationId": "methods.getCheckout",
                "parameters": [
                    {
                        "$ref": "#/parameters/jsonrpc_id"
                    },
                    {
                        "$ref": "#/parameters/zipcode"
                    },
                    {
                        "$ref": "#/parameters/aoid"
                    },
                    {
                        "$ref": "#/parameters/zone_id"
                    },
                    {
                        "$ref": "#/parameters/latitude"
                    },
                    {
                        "$ref": "#/parameters/longitude"
                    },
                    {
                        "$ref": "#/parameters/item_count"
                    },
                    {
                        "$ref": "#/parameters/cart_amount"
                    },
                    {
                        "$ref": "#/parameters/payment_type"
                    },
                    {
                        "$ref": "#/parameters/no_airfreight"
                    },
                    {
                        "$ref": "#/parameters/liquids"
                    },
                    {
                        "$ref": "#/parameters/checkout"
                    },
                    {
                        "$ref": "#/parameters/method_code"
                    },
                    {
                        "$ref": "#/parameters/company"
                    },
                    {
                        "$ref": "#/parameters/origin_date"
                    },
                    {
                        "$ref": "#/parameters/profile_id"
                    },
                    {
                        "$ref": "#/parameters/ignore_capacity"
                    },
                    {
                        "$ref": "#/parameters/ignore_cutoff"
                    },
                    {
                        "$ref": "#/parameters/reserved_interval_id"
                    },
                    {
                        "$ref": "#/parameters/original_delivery_date"
                    },
                    {
                        "$ref": "#/parameters/tag"
                    }
                ],
                "responses": {
                    "default": {
                        "description": "Success/error response",
                        "schema": {
                            "$ref": "#/definitions/JsonRpcMethodsGetCheckoutResponse"
                        }
                    }
                }
            }
        },
        "/jsonrpc/v1/methods.get-checkout-multiple": {
            "get": {
                "tags": [
                    "json-rpc",
                    "methods"
                ],
                "summary": "Get checkout methods for multiple orders",
                "description": "Two-dimensional orders array defines parameters for each individual order.\n    At least one order have to be specified with one or more parameters.",
                "operationId": "methods.getCheckoutMultiple",
                "parameters": [
                    {
                        "$ref": "#/parameters/jsonrpc_id"
                    },
                    {
                        "$ref": "#/parameters/zipcode"
                    },
                    {
                        "$ref": "#/parameters/aoid"
                    },
                    {
                        "$ref": "#/parameters/zone_id"
                    },
                    {
                        "$ref": "#/parameters/latitude"
                    },
                    {
                        "$ref": "#/parameters/longitude"
                    },
                    {
                        "$ref": "#/parameters/checkout"
                    },
                    {
                        "$ref": "#/parameters/method_code"
                    },
                    {
                        "$ref": "#/parameters/origin_date"
                    },
                    {
                        "$ref": "#/parameters/ignore_capacity"
                    },
                    {
                        "$ref": "#/parameters/ignore_cutoff"
                    },
                    {
                        "$ref": "#/parameters/orders_profile_id"
                    },
                    {
                        "$ref": "#/parameters/orders_item_count"
                    },
                    {
                        "$ref": "#/parameters/orders_cart_amount"
                    },
                    {
                        "$ref": "#/parameters/orders_no_airfreight"
                    },
                    {
                        "$ref": "#/parameters/orders_payment_type"
                    },
                    {
                        "$ref": "#/parameters/tag"
                    }
                ],
                "responses": {
                    "default": {
                        "description": "Success/error response.\n    <br />\n    Return list of orders in the same sequence as in request.",
                        "schema": {
                            "$ref": "#/definitions/JsonRpcMethodsGetCheckoutMultipleResponse"
                        }
                    }
                }
            }
        },
        "/jsonrpc/v1/methods.get-details": {
            "get": {
                "tags": [
                    "json-rpc",
                    "methods"
                ],
                "summary": "Get method details",
                "operationId": "methods.getDetails",
                "parameters": [
                    {
                        "$ref": "#/parameters/jsonrpc_id"
                    },
                    {
                        "$ref": "#/parameters/code"
                    },
                    {
                        "$ref": "#/parameters/join"
                    },
                    {
                        "$ref": "#/parameters/zone_id"
                    },
                    {
                        "$ref": "#/parameters/service_level"
                    },
                    {
                        "$ref": "#/parameters/zipcode"
                    },
                    {
                        "$ref": "#/parameters/aoid"
                    },
                    {
                        "$ref": "#/parameters/origin_date"
                    }
                ],
                "responses": {
                    "default": {
                        "description": "Success/error response",
                        "schema": {
                            "$ref": "#/definitions/JsonRpcMethodsGetDetailsResponse"
                        }
                    }
                }
            }
        },
        "/jsonrpc/v1/methods.get-details-multiple": {
            "get": {
                "tags": [
                    "json-rpc",
                    "methods"
                ],
                "summary": "Get method details multiple",
                "operationId": "methods.getDetailsMultiple",
                "parameters": [
                    {
                        "$ref": "#/parameters/jsonrpc_id"
                    },
                    {
                        "$ref": "#/parameters/get_method_details_multiple_code"
                    },
                    {
                        "$ref": "#/parameters/get_method_details_multiple_join"
                    },
                    {
                        "$ref": "#/parameters/get_method_details_multiple_zone_id"
                    },
                    {
                        "$ref": "#/parameters/get_method_details_multiple_service_level"
                    },
                    {
                        "$ref": "#/parameters/get_method_details_multiple_zipcode"
                    },
                    {
                        "$ref": "#/parameters/get_method_details_multiple_aoid"
                    },
                    {
                        "$ref": "#/parameters/get_method_details_multiple_origin_date"
                    }
                ],
                "responses": {
                    "default": {
                        "description": "Success/error response.\n    <br />\n    Return list of methods in the same sequence as in request.",
                        "schema": {
                            "$ref": "#/definitions/JsonRpcMethodsGetDetailsMultipleResponse"
                        }
                    }
                }
            }
        },
        "/jsonrpc/v1/methods.get-all": {
            "get": {
                "tags": [
                    "json-rpc",
                    "methods"
                ],
                "summary": "Get all methods",
                "operationId": "getAll",
                "parameters": [
                    {
                        "$ref": "#/parameters/jsonrpc_id"
                    }
                ],
                "responses": {
                    "default": {
                        "description": "Success/error response",
                        "schema": {
                            "$ref": "#/definitions/JsonRpcMethodsGetAllResponse"
                        }
                    }
                }
            }
        },
        "/jsonrpc/v1/pickups.get-list": {
            "get": {
                "tags": [
                    "json-rpc",
                    "pickups"
                ],
                "summary": "Get pup list",
                "operationId": "pickups.getList",
                "parameters": [
                    {
                        "$ref": "#/parameters/jsonrpc_id"
                    },
                    {
                        "$ref": "#/parameters/zipcode"
                    },
                    {
                        "$ref": "#/parameters/aoid"
                    },
                    {
                        "$ref": "#/parameters/latitude"
                    },
                    {
                        "$ref": "#/parameters/longitude"
                    },
                    {
                        "$ref": "#/parameters/item_count"
                    },
                    {
                        "$ref": "#/parameters/cart_amount"
                    },
                    {
                        "$ref": "#/parameters/payment_type"
                    },
                    {
                        "$ref": "#/parameters/no_airfreight"
                    },
                    {
                        "$ref": "#/parameters/liquids"
                    },
                    {
                        "$ref": "#/parameters/checkout"
                    },
                    {
                        "$ref": "#/parameters/method_code"
                    },
                    {
                        "$ref": "#/parameters/company"
                    },
                    {
                        "$ref": "#/parameters/origin_date"
                    },
                    {
                        "$ref": "#/parameters/profile_id"
                    },
                    {
                        "$ref": "#/parameters/ignore_capacity"
                    },
                    {
                        "$ref": "#/parameters/ignore_cutoff"
                    },
                    {
                        "$ref": "#/parameters/reserved_interval_id"
                    },
                    {
                        "$ref": "#/parameters/original_delivery_date"
                    },
                    {
                        "$ref": "#/parameters/tag"
                    }
                ],
                "responses": {
                    "default": {
                        "description": "Success/error response",
                        "schema": {
                            "$ref": "#/definitions/JsonRpcPickupsGetListResponse"
                        }
                    }
                }
            }
        },
        "/jsonrpc/v1/pickups.get-details": {
            "get": {
                "tags": [
                    "json-rpc",
                    "pickups"
                ],
                "summary": "Get pup details",
                "operationId": "pickups.getDetails",
                "parameters": [
                    {
                        "$ref": "#/parameters/jsonrpc_id"
                    },
                    {
                        "$ref": "#/parameters/pickup_id"
                    },
                    {
                        "$ref": "#/parameters/join"
                    },
                    {
                        "$ref": "#/parameters/zipcode"
                    },
                    {
                        "$ref": "#/parameters/aoid"
                    },
                    {
                        "$ref": "#/parameters/latitude"
                    },
                    {
                        "$ref": "#/parameters/longitude"
                    },
                    {
                        "$ref": "#/parameters/item_count"
                    },
                    {
                        "$ref": "#/parameters/cart_amount"
                    },
                    {
                        "$ref": "#/parameters/payment_type"
                    },
                    {
                        "$ref": "#/parameters/no_airfreight"
                    },
                    {
                        "$ref": "#/parameters/liquids"
                    },
                    {
                        "$ref": "#/parameters/checkout"
                    },
                    {
                        "$ref": "#/parameters/method_code"
                    },
                    {
                        "$ref": "#/parameters/company"
                    },
                    {
                        "$ref": "#/parameters/origin_date"
                    },
                    {
                        "$ref": "#/parameters/profile_id"
                    },
                    {
                        "$ref": "#/parameters/ignore_capacity"
                    },
                    {
                        "$ref": "#/parameters/ignore_cutoff"
                    },
                    {
                        "$ref": "#/parameters/reserved_interval_id"
                    },
                    {
                        "$ref": "#/parameters/original_delivery_date"
                    },
                    {
                        "$ref": "#/parameters/tag"
                    }
                ],
                "responses": {
                    "default": {
                        "description": "Success/error response",
                        "schema": {
                            "$ref": "#/definitions/JsonRpcPickupsGetDetailsResponse"
                        }
                    }
                }
            }
        },
        "/jsonrpc/v1/pickups.get": {
            "get": {
                "tags": [
                    "json-rpc",
                    "pickups"
                ],
                "summary": "Get pickup",
                "operationId": "pickups.get",
                "parameters": [
                    {
                        "$ref": "#/parameters/jsonrpc_id"
                    },
                    {
                        "$ref": "#/parameters/pickup_id"
                    },
                    {
                        "$ref": "#/parameters/join"
                    },
                    {
                        "$ref": "#/parameters/zipcode"
                    },
                    {
                        "$ref": "#/parameters/aoid"
                    }
                ],
                "responses": {
                    "default": {
                        "description": "Success/error response",
                        "schema": {
                            "$ref": "#/definitions/JsonRpcPickupsGetResponse"
                        }
                    }
                }
            }
        },
        "/jsonrpc/v1/pickups.create": {
            "post": {
                "tags": [
                    "json-rpc",
                    "pickups"
                ],
                "summary": "Create pickups",
                "operationId": "pickups.create",
                "parameters": [
                    {
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/JsonRpcPickupsCreateRequest"
                        }
                    }
                ],
                "responses": {
                    "default": {
                        "description": "Success/error response",
                        "schema": {
                            "$ref": "#/definitions/JsonRpcPickupsCreateResponse"
                        }
                    }
                }
            }
        },
        "/jsonrpc/v1/profiles.import": {
            "post": {
                "tags": [
                    "profiles",
                    "json-rpc"
                ],
                "summary": "Import profiles",
                "operationId": "profiles.import",
                "parameters": [
                    {
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/JsonRpcProfilesImportRequest"
                        }
                    }
                ],
                "responses": {
                    "default": {
                        "description": "Success/error response",
                        "schema": {
                            "$ref": "#/definitions/JsonRpcOkResponse"
                        }
                    }
                }
            }
        },
        "/jsonrpc/v1/zones.get-rate": {
            "get": {
                "tags": [
                    "json-rpc",
                    "zones"
                ],
                "summary": "Get zone rate",
                "description": "Either zipcode or zone_id parameter is required.\n    If zone_id is specified then it is used instead of zipcode to find required zone.",
                "operationId": "zones.getRate",
                "parameters": [
                    {
                        "$ref": "#/parameters/jsonrpc_id"
                    },
                    {
                        "$ref": "#/parameters/method_code_required"
                    },
                    {
                        "$ref": "#/parameters/zipcode"
                    },
                    {
                        "$ref": "#/parameters/zone_id"
                    },
                    {
                        "$ref": "#/parameters/weight"
                    },
                    {
                        "$ref": "#/parameters/insurance_sum"
                    }
                ],
                "responses": {
                    "default": {
                        "description": "Success/error response",
                        "schema": {
                            "$ref": "#/definitions/JsonRpcZonesGetRateResponse"
                        }
                    }
                }
            }
        },
        "/jsonrpc/v1/zones.update-rate": {
            "post": {
                "tags": [
                    "zones",
                    "json-rpc"
                ],
                "summary": "Update zone rate",
                "description": "Zone should have at least one search criteria: zone_id|pickup_id|pickup_external_id",
                "operationId": "zones.updateRate",
                "parameters": [
                    {
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/JsonRpcZonesUpdateRateRequest"
                        }
                    }
                ],
                "responses": {
                    "default": {
                        "description": "Success/error response",
                        "schema": {
                            "$ref": "#/definitions/JsonRpcZonesUpdateRateResponse"
                        }
                    }
                }
            }
        },
        "/jsonrpc/v1/zones.get-attributes": {
            "get": {
                "tags": [
                    "json-rpc",
                    "zones"
                ],
                "summary": "Get zone attributes",
                "operationId": "zones.getAttributes",
                "parameters": [
                    {
                        "$ref": "#/parameters/jsonrpc_id"
                    },
                    {
                        "$ref": "#/parameters/zone_id_required"
                    }
                ],
                "responses": {
                    "default": {
                        "description": "Success/error response",
                        "schema": {
                            "$ref": "#/definitions/JsonRpcZonesGetAttributesResponse"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "Address": {
            "required": [
                "pickup_id",
                "city",
                "street",
                "house"
            ],
            "properties": {
                "pickup_id": {
                    "type": "integer",
                    "example": 16213
                },
                "latitude": {
                    "type": "number",
                    "format": "float",
                    "example": 55.738176
                },
                "longitude": {
                    "type": "number",
                    "format": "float",
                    "example": 37.631595
                },
                "city": {
                    "type": "string",
                    "example": "\u041c\u043e\u0441\u043a\u0432\u0430"
                },
                "street": {
                    "type": "string",
                    "example": "\u041d\u043e\u0432\u043e\u043a\u0443\u0437\u043d\u0435\u0446\u043a\u0430\u044f"
                },
                "house": {
                    "type": "string",
                    "example": "11/13"
                },
                "office": {
                    "type": "string"
                },
                "underground_station": {
                    "type": "string",
                    "example": "\u043c. \u041d\u043e\u0432\u043e\u043a\u0443\u0437\u043d\u0435\u0446\u043a\u0430\u044f/\u0422\u0440\u0435\u0442\u044c\u044f\u043a\u043e\u0432\u0441\u043a\u0430\u044f/\u041f\u0430\u0432\u0435\u043b\u0435\u0446\u043a\u0430\u044f"
                }
            },
            "type": "object"
        },
        "ArrayItem": {
            "required": [
                "key",
                "value"
            ],
            "properties": {
                "key": {
                    "type": "string",
                    "example": "is_call_needed"
                },
                "value": {
                    "example": "1"
                }
            },
            "type": "object"
        },
        "CheckoutMethod": {
            "required": [
                "code",
                "name",
                "category_name",
                "has_horizon",
                "has_intervals",
                "is_client_name_required",
                "delivery_price",
                "checkout_name",
                "days",
                "payment_type"
            ],
            "properties": {
                "code": {
                    "type": "string",
                    "example": "lamoda_showroom_novokuznetskaya"
                },
                "name": {
                    "type": "string",
                    "example": "\u041d\u043e\u0432\u043e\u043a\u0443\u0437\u043d\u0435\u0446\u043a\u0430\u044f, \u043c. \u041d\u043e\u0432\u043e\u043a\u0443\u0437\u043d\u0435\u0446\u043a\u0430\u044f, \u043c.\u0422\u0440\u0435\u0442\u044c\u044f\u043a\u043e\u0432\u0441\u043a\u0430\u044f, \u043c. \u041f\u0430\u0432\u0435\u043b\u0435\u0446\u043a\u0430\u044f"
                },
                "category_name": {
                    "type": "string",
                    "example": "\u0421\u0430\u043c\u043e\u0432\u044b\u0432\u043e\u0437"
                },
                "checkout_description": {
                    "type": "string",
                    "example": "\u0414\u043e\u0441\u0442\u0430\u0432\u043a\u0430 \u0437\u0430\u043a\u0430\u0437\u0430 \u0432 \u043f\u0443\u043d\u043a\u0442 \u0432\u044b\u0434\u0430\u0447\u0438 \u0437\u0430\u043a\u0430\u0437\u043e\u0432 \u0431\u0435\u0441\u043f\u043b\u0430\u0442\u043d\u0430 \u0432 \u0441\u043b\u0443\u0447\u0430\u0435 ..."
                },
                "callcenter_description": {
                    "type": "string",
                    "example": "\u041f\u0440\u0438 \u0432\u044b\u043a\u0443\u043f\u0435 \u043d\u0430 \u0441\u0443\u043c\u043c\u0443 \u0431\u043e\u043b\u0435\u0435 2500 \u0440\u0443\u0431\u043b\u0435\u0439 \u0434\u043e\u0441\u0442\u0430\u0432\u043a\u0430 \u0431\u0435\u0441\u043f\u043b\u0430\u0442\u043d\u0430\u044f ..."
                },
                "has_horizon": {
                    "type": "boolean"
                },
                "has_intervals": {
                    "type": "boolean"
                },
                "is_client_name_required": {
                    "type": "boolean",
                    "example": false
                },
                "delivery_price": {
                    "type": "number",
                    "format": "float",
                    "example": 250
                },
                "free_delivery_net_threshold": {
                    "type": "number",
                    "format": "float",
                    "example": 2500
                },
                "checkout_name": {
                    "type": "string",
                    "example": "\u0414\u043e\u0441\u0442\u0430\u0432\u043a\u0430 \u0432 \u043f\u0443\u043d\u043a\u0442 \u0432\u044b\u0434\u0430\u0447\u0438 \u0437\u0430\u043a\u0430\u0437\u043e\u0432 Lamoda"
                },
                "delivery_date_min": {
                    "type": "string",
                    "example": "2017-03-04"
                },
                "delivery_date_max": {
                    "type": "string",
                    "example": "2017-03-08"
                },
                "days": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/Day"
                    }
                },
                "address": {
                    "$ref": "#/definitions/Address"
                },
                "payment_type": {
                    "description": "0 - all types, 1 - prepayment, 2 - postpayment",
                    "type": "integer",
                    "example": 0
                },
                "customs_threshold": {
                    "type": "number",
                    "format": "float"
                },
                "customs_threshold_description": {
                    "type": "string"
                },
                "communication_description": {
                    "type": "string",
                    "example": "{\u041a\u043e\u0433\u0434\u0430 \u0437\u0430\u043a\u0430\u0437 \u043f\u043e\u0441\u0442\u0443\u043f\u0438\u0442 \u0432 \u043f\u0443\u043d\u043a\u0442 \u0441\u0430\u043c\u043e\u0432\u044b\u0432\u043e\u0437\u0430, \u0432\u0430\u043c \u043f\u0440\u0438\u0434\u0435\u0442 SMS..."
                },
                "is_own": {
                    "type": "boolean",
                    "example": false
                }
            },
            "type": "object"
        },
        "Phone": {
            "required": [
                "code",
                "number"
            ],
            "properties": {
                "code": {
                    "type": "string",
                    "example": "+7"
                },
                "number": {
                    "type": "string",
                    "example": "(495) 363-63-93"
                }
            },
            "type": "object"
        },
        "CreatePickups\\Pickup": {
            "required": [
                "external_id",
                "name",
                "city",
                "street",
                "house"
            ],
            "properties": {
                "external_id": {
                    "type": "string"
                },
                "name": {
                    "type": "string",
                    "example": "\u041d\u043e\u0432\u043e\u043a\u0443\u0437\u043d\u0435\u0446\u043a\u0430\u044f"
                },
                "zipcode": {
                    "type": "string"
                },
                "region": {
                    "type": "string"
                },
                "city": {
                    "type": "string",
                    "example": "\u041c\u043e\u0441\u043a\u0432\u0430"
                },
                "street": {
                    "type": "string",
                    "example": "\u041d\u043e\u0432\u043e\u043a\u0443\u0437\u043d\u0435\u0446\u043a\u0430\u044f"
                },
                "house": {
                    "type": "string",
                    "example": "11/13"
                },
                "latitude": {
                    "type": "number",
                    "format": "float",
                    "example": 55.738176
                },
                "longitude": {
                    "type": "number",
                    "format": "float",
                    "example": 37.631595
                },
                "underground_station": {
                    "type": "string",
                    "example": "\u043c. \u041d\u043e\u0432\u043e\u043a\u0443\u0437\u043d\u0435\u0446\u043a\u0430\u044f/\u0422\u0440\u0435\u0442\u044c\u044f\u043a\u043e\u0432\u0441\u043a\u0430\u044f/\u041f\u0430\u0432\u0435\u043b\u0435\u0446\u043a\u0430\u044f"
                },
                "map_description": {
                    "type": "string",
                    "example": "\u0412\u044b\u0445\u043e\u0434\u0438\u0442\u0435 \u0438\u0437 \u0441\u0442\u0430\u043d\u0446\u0438\u0438 \u043c\u0435\u0442\u0440\u043e <b>\u041d\u043e\u0432\u043e\u043a\u0443\u0437\u043d\u0435\u0446\u043a\u0430\u044f </b>\u043d\u0430 \u0443\u043b\u0438\u0446\u0443 \u041d\u043e\u0432\u043e\u043a\u0443\u0437\u043d\u0435\u0446\u043a\u0430\u044f..."
                },
                "is_bankcard_accepted": {
                    "type": "boolean",
                    "example": true
                },
                "is_inactive": {
                    "type": "boolean",
                    "example": false
                },
                "is_24hours": {
                    "type": "boolean",
                    "example": false
                },
                "work_time": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/WorkTime"
                    }
                },
                "phones": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/Phone"
                    }
                },
                "group_code": {
                    "type": "string",
                    "example": "4"
                },
                "attributes": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/ArrayItem"
                    }
                },
                "photo_url": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "photo_base64": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                }
            },
            "type": "object"
        },
        "CreatePickups\\Request": {
            "required": [
                "method_code"
            ],
            "properties": {
                "method_code": {
                    "type": "string",
                    "example": "lamoda_showroom_novokuznetskaya"
                },
                "pickups": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/CreatePickups\\Pickup"
                    }
                }
            },
            "type": "object"
        },
        "CreatePickups\\Response": {
            "properties": {
                "success": {
                    "type": "boolean"
                },
                "error": {
                    "$ref": "#/definitions/ResponseError"
                }
            },
            "type": "object"
        },
        "DayWithPalletMethod": {
            "required": [
                "date",
                "available_intervals"
            ],
            "properties": {
                "date": {
                    "type": "string",
                    "example": "2017-03-04"
                },
                "available_intervals": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/IntervalWithPalletMethod"
                    }
                },
                "is_available": {
                    "type": "boolean",
                    "example": true
                }
            },
            "type": "object"
        },
        "FindIntervalResponse": {
            "properties": {
                "date": {
                    "description": "Date, YYYY-MM-DD",
                    "type": "string",
                    "example": "2017-03-04"
                },
                "interval": {
                    "$ref": "#/definitions/Interval"
                },
                "success": {
                    "type": "boolean"
                },
                "error": {
                    "$ref": "#/definitions/ResponseError"
                }
            },
            "type": "object"
        },
        "GetAllMethodsResponse": {
            "properties": {
                "methods": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/Method"
                    }
                },
                "success": {
                    "type": "boolean"
                },
                "error": {
                    "$ref": "#/definitions/ResponseError"
                }
            },
            "type": "object"
        },
        "Day": {
            "required": [
                "date",
                "intervals"
            ],
            "properties": {
                "date": {
                    "type": "string",
                    "example": "2017-03-04"
                },
                "intervals": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/IntervalWithPalletMethod"
                    }
                },
                "is_available": {
                    "type": "boolean",
                    "example": true
                }
            },
            "type": "object"
        },
        "GroupDay": {
            "required": [
                "date",
                "intervals"
            ],
            "properties": {
                "date": {
                    "type": "string",
                    "example": "2017-03-04"
                },
                "intervals": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/GroupInterval"
                    }
                }
            },
            "type": "object"
        },
        "GroupInterval": {
            "required": [
                "start",
                "end"
            ],
            "properties": {
                "start": {
                    "type": "string",
                    "example": "2017-03-04 10:00:00"
                },
                "end": {
                    "type": "string",
                    "example": "2017-03-04 22:00:00"
                }
            },
            "type": "object"
        },
        "GroupMethod": {
            "required": [
                "method_type_code",
                "method_type_name",
                "code",
                "name",
                "category_name",
                "checkout_name",
                "days"
            ],
            "properties": {
                "method_type_code": {
                    "type": "string",
                    "example": "pickup"
                },
                "method_type_name": {
                    "type": "string",
                    "example": "\u0421\u0430\u043c\u043e\u0432\u044b\u0432\u043e\u0437"
                },
                "code": {
                    "type": "string",
                    "example": "lamoda_showroom_novokuznetskaya"
                },
                "name": {
                    "type": "string",
                    "example": "\u041d\u043e\u0432\u043e\u043a\u0443\u0437\u043d\u0435\u0446\u043a\u0430\u044f, \u043c. \u041d\u043e\u0432\u043e\u043a\u0443\u0437\u043d\u0435\u0446\u043a\u0430\u044f, \u043c.\u0422\u0440\u0435\u0442\u044c\u044f\u043a\u043e\u0432\u0441\u043a\u0430\u044f, \u043c. \u041f\u0430\u0432\u0435\u043b\u0435\u0446\u043a\u0430\u044f"
                },
                "category_name": {
                    "type": "string",
                    "example": "\u0421\u0430\u043c\u043e\u0432\u044b\u0432\u043e\u0437"
                },
                "checkout_name": {
                    "type": "string",
                    "example": "\u0414\u043e\u0441\u0442\u0430\u0432\u043a\u0430 \u0432 \u043f\u0443\u043d\u043a\u0442 \u0432\u044b\u0434\u0430\u0447\u0438 \u0437\u0430\u043a\u0430\u0437\u043e\u0432 Lamoda"
                },
                "days": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/GroupDay"
                    }
                }
            },
            "type": "object"
        },
        "GetCheckoutMethodsMultiple\\Response": {
            "properties": {
                "orders": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/ResponseOrder"
                    }
                },
                "groups": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/ResponseGroup"
                    }
                },
                "success": {
                    "type": "boolean"
                },
                "error": {
                    "$ref": "#/definitions/ResponseError"
                }
            },
            "type": "object"
        },
        "ResponseGroup": {
            "required": [
                "orders",
                "methods"
            ],
            "properties": {
                "orders": {
                    "type": "array",
                    "items": {
                        "type": "integer"
                    },
                    "example": [
                        0,
                        1
                    ]
                },
                "methods": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/GroupMethod"
                    }
                }
            },
            "type": "object"
        },
        "ResponseOrder": {
            "required": [
                "method_types"
            ],
            "properties": {
                "method_types": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/MethodType"
                    }
                }
            },
            "type": "object"
        },
        "GetCheckoutMethodsResponse": {
            "properties": {
                "method_types": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/MethodType"
                    }
                },
                "success": {
                    "type": "boolean"
                },
                "error": {
                    "$ref": "#/definitions/ResponseError"
                }
            },
            "type": "object"
        },
        "GetDeliveryInfoMethodType": {
            "required": [
                "code",
                "name",
                "is_lme",
                "service_level_types"
            ],
            "properties": {
                "code": {
                    "type": "string",
                    "example": "pickup"
                },
                "name": {
                    "type": "string",
                    "example": "\u0421\u0430\u043c\u043e\u0432\u044b\u0432\u043e\u0437"
                },
                "is_lme": {
                    "type": "boolean",
                    "example": true
                },
                "service_level_types": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/GetDeliveryInfoServiceLevelType"
                    }
                }
            },
            "type": "object"
        },
        "GetDeliveryInfoResponse": {
            "properties": {
                "method_types": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/GetDeliveryInfoMethodType"
                    }
                },
                "success": {
                    "type": "boolean"
                },
                "error": {
                    "$ref": "#/definitions/ResponseError"
                }
            },
            "type": "object"
        },
        "GetDeliveryInfoServiceLevelType": {
            "required": [
                "code",
                "name",
                "checkout_name",
                "is_tryon_allowed",
                "is_rejection_allowed",
                "is_bankcard_accepted",
                "cutoff_time"
            ],
            "properties": {
                "code": {
                    "type": "string",
                    "example": "plus"
                },
                "name": {
                    "type": "string",
                    "example": "\u041f\u043b\u044e\u0441"
                },
                "checkout_name": {
                    "type": "string",
                    "example": "\u041f\u043b\u044e\u0441"
                },
                "description": {
                    "type": "string"
                },
                "free_delivery_net_threshold_from": {
                    "type": "number",
                    "format": "float",
                    "example": "2500.00"
                },
                "free_delivery_net_threshold_to": {
                    "type": "number",
                    "format": "float",
                    "example": "2500.00"
                },
                "free_delivery_net_threshold_percent_from": {
                    "type": "integer",
                    "example": 0
                },
                "free_delivery_net_threshold_percent_to": {
                    "type": "integer",
                    "example": 0
                },
                "free_delivery_threshold_share": {
                    "type": "integer",
                    "example": 0
                },
                "free_delivery_gross_threshold_from": {
                    "type": "number",
                    "format": "float"
                },
                "free_delivery_gross_threshold_to": {
                    "type": "number",
                    "format": "float"
                },
                "delivery_price_from": {
                    "type": "number",
                    "format": "float",
                    "example": "250.00"
                },
                "delivery_price_to": {
                    "type": "number",
                    "format": "float",
                    "example": "250.00"
                },
                "is_tryon_allowed": {
                    "type": "boolean",
                    "example": true
                },
                "is_rejection_allowed": {
                    "type": "boolean",
                    "example": true
                },
                "is_bankcard_accepted": {
                    "type": "boolean",
                    "example": true
                },
                "delivery_date_from": {
                    "type": "integer",
                    "example": 1
                },
                "delivery_date_to": {
                    "type": "integer",
                    "example": 5
                },
                "storage_days_from": {
                    "type": "integer",
                    "example": 2
                },
                "storage_days_to": {
                    "type": "integer",
                    "example": 2
                },
                "tryon_limit_from": {
                    "type": "integer",
                    "example": 15
                },
                "tryon_limit_to": {
                    "type": "integer",
                    "example": 15
                },
                "cutoff_time": {
                    "type": "string",
                    "example": "2017-03-03 23:59:00"
                },
                "has_horizon": {
                    "type": "boolean",
                    "example": true
                },
                "horizon_from": {
                    "type": "integer",
                    "example": 1
                },
                "horizon_till": {
                    "type": "integer",
                    "example": 7
                },
                "day_min": {
                    "type": "integer",
                    "example": 3
                },
                "day_max": {
                    "type": "integer",
                    "example": 5
                },
                "is_business_days": {
                    "type": "boolean",
                    "example": false
                },
                "payment_type": {
                    "type": "integer",
                    "example": 0
                }
            },
            "type": "object"
        },
        "GetDeliveryInfoShortResponse": {
            "properties": {
                "terms": {
                    "$ref": "#/definitions/GetDeliveryInfoServiceLevelType"
                },
                "success": {
                    "type": "boolean"
                },
                "error": {
                    "$ref": "#/definitions/ResponseError"
                }
            },
            "type": "object"
        },
        "GetIntervalResponse": {
            "properties": {
                "date": {
                    "type": "string",
                    "example": "2017-03-04"
                },
                "interval": {
                    "$ref": "#/definitions/Interval"
                },
                "success": {
                    "type": "boolean"
                },
                "error": {
                    "$ref": "#/definitions/ResponseError"
                }
            },
            "type": "object"
        },
        "GetMethodDetails\\Company": {
            "required": [
                "code",
                "name",
                "is_deleted"
            ],
            "properties": {
                "code": {
                    "type": "string",
                    "example": "lamoda"
                },
                "name": {
                    "type": "string",
                    "example": "Lamoda Express"
                },
                "is_deleted": {
                    "type": "boolean",
                    "example": false
                }
            },
            "type": "object"
        },
        "LeadTime": {
            "required": [
                "day_min",
                "day_max",
                "is_business_days",
                "day_min_calculated",
                "day_max_calculated"
            ],
            "properties": {
                "day_min": {
                    "type": "integer",
                    "example": 0
                },
                "day_max": {
                    "type": "integer",
                    "example": 5
                },
                "is_business_days": {
                    "type": "boolean",
                    "example": false
                },
                "day_min_calculated": {
                    "type": "string",
                    "example": "2017-03-04"
                },
                "day_max_calculated": {
                    "type": "string",
                    "example": "2017-03-08"
                }
            },
            "type": "object"
        },
        "GetMethodDetails\\Method": {
            "required": [
                "code",
                "name",
                "type",
                "group_name",
                "is_active",
                "is_deleted",
                "is_client_name_required",
                "checkout_name",
                "max_postpones"
            ],
            "properties": {
                "code": {
                    "type": "string",
                    "example": "lamoda_showroom_novokuznetskaya"
                },
                "name": {
                    "type": "string",
                    "example": "\u041d\u043e\u0432\u043e\u043a\u0443\u0437\u043d\u0435\u0446\u043a\u0430\u044f, \u043c. \u041d\u043e\u0432\u043e\u043a\u0443\u0437\u043d\u0435\u0446\u043a\u0430\u044f, \u043c.\u0422\u0440\u0435\u0442\u044c\u044f\u043a\u043e\u0432\u0441\u043a\u0430\u044f, \u043c. \u041f\u0430\u0432\u0435\u043b\u0435\u0446\u043a\u0430\u044f"
                },
                "type": {
                    "description": "1 - \u043a\u0443\u0440\u044c\u0435\u0440\u0441\u043a\u0430\u044f \u0434\u043e\u0441\u0442\u0430\u0432\u043a\u0430, 2 - \u0441\u0430\u043c\u043e\u0432\u044b\u0432\u043e\u0437, 3 - \u043f\u043e\u0447\u0442\u043e\u0432\u0430\u044f \u0434\u043e\u0441\u0442\u0430\u0432\u043a\u0430",
                    "type": "integer",
                    "example": 2
                },
                "group_name": {
                    "type": "string",
                    "example": "\u0421\u0430\u043c\u043e\u0432\u044b\u0432\u043e\u0437"
                },
                "description": {
                    "type": "string",
                    "example": "- \u041f\u0440\u0438 \u0437\u0430\u043a\u0430\u0437\u0435 \u0434\u043e 24:00, \u0434\u043e\u0441\u0442\u0430\u0432\u043a\u0430 \u043d\u0430 \u0441\u043b\u0435\u0434\u0443\u044e\u0449\u0438\u0439 \u0434\u0435\u043d\u044c..."
                },
                "is_active": {
                    "type": "boolean"
                },
                "is_deleted": {
                    "type": "boolean"
                },
                "carrier": {
                    "type": "string",
                    "example": "leos-express"
                },
                "tracking_url": {
                    "type": "string"
                },
                "is_own": {
                    "type": "boolean",
                    "example": true
                },
                "is_client_name_required": {
                    "type": "boolean",
                    "example": false
                },
                "company": {
                    "$ref": "#/definitions/GetMethodDetails\\Company"
                },
                "lead_time": {
                    "$ref": "#/definitions/LeadTime"
                },
                "zones": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/GetMethodDetails\\Zone"
                    }
                },
                "checkout_name": {
                    "type": "string",
                    "example": "\u0414\u043e\u0441\u0442\u0430\u0432\u043a\u0430 \u0432 \u043f\u0443\u043d\u043a\u0442 \u0432\u044b\u0434\u0430\u0447\u0438 \u0437\u0430\u043a\u0430\u0437\u043e\u0432 Lamoda"
                },
                "communication_description": {
                    "type": "string",
                    "example": "{\u041a\u043e\u0433\u0434\u0430 \u0437\u0430\u043a\u0430\u0437 \u043f\u043e\u0441\u0442\u0443\u043f\u0438\u0442 \u0432 \u043f\u0443\u043d\u043a\u0442 \u0441\u0430\u043c\u043e\u0432\u044b\u0432\u043e\u0437\u0430, \u0432\u0430\u043c \u043f\u0440\u0438\u0434\u0435\u0442 SMS..."
                },
                "max_postpones": {
                    "type": "integer",
                    "example": 3
                }
            },
            "type": "object"
        },
        "GetMethodDetails\\Pickup": {
            "required": [
                "pickup_id",
                "city",
                "street",
                "house",
                "is_24hours",
                "is_returns_accepted",
                "storage_days"
            ],
            "properties": {
                "pickup_id": {
                    "type": "integer",
                    "example": 16213
                },
                "external_id": {
                    "type": "string"
                },
                "group_id": {
                    "type": "integer",
                    "example": 4
                },
                "city": {
                    "type": "string",
                    "example": "\u041c\u043e\u0441\u043a\u0432\u0430"
                },
                "street": {
                    "type": "string",
                    "example": "\u041d\u043e\u0432\u043e\u043a\u0443\u0437\u043d\u0435\u0446\u043a\u0430\u044f"
                },
                "house": {
                    "type": "string",
                    "example": "11/13"
                },
                "office": {
                    "type": "string"
                },
                "underground_station": {
                    "type": "string",
                    "example": "\u043c. \u041d\u043e\u0432\u043e\u043a\u0443\u0437\u043d\u0435\u0446\u043a\u0430\u044f/\u0422\u0440\u0435\u0442\u044c\u044f\u043a\u043e\u0432\u0441\u043a\u0430\u044f/\u041f\u0430\u0432\u0435\u043b\u0435\u0446\u043a\u0430\u044f"
                },
                "latitude": {
                    "type": "string",
                    "example": "55.738176"
                },
                "longitude": {
                    "type": "string",
                    "example": "37.631595"
                },
                "is_24hours": {
                    "type": "boolean",
                    "example": false
                },
                "work_time": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/WorkTime"
                    }
                },
                "is_returns_accepted": {
                    "type": "boolean",
                    "example": true
                },
                "widget_description": {
                    "type": "string",
                    "example": "\u0412\u044b\u0445\u043e\u0434\u0438\u0442\u0435 \u0438\u0437 \u0441\u0442\u0430\u043d\u0446\u0438\u0438 \u043c\u0435\u0442\u0440\u043e <b>\u041d\u043e\u0432\u043e\u043a\u0443\u0437\u043d\u0435\u0446\u043a\u0430\u044f </b>\u043d\u0430 \u0443\u043b\u0438\u0446\u0443 \u041d\u043e\u0432\u043e\u043a\u0443\u0437\u043d\u0435\u0446\u043a\u0430\u044f..."
                },
                "storage_days": {
                    "type": "integer",
                    "example": 2
                },
                "phones": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    },
                    "example": [
                        "+7(495) 363-63-93"
                    ]
                },
                "photos": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    },
                    "example": [
                        "//files.lmcdn.ru/delivery/ru/75/6978cffab68e1e0ce126fa0d42368e75.png"
                    ]
                }
            },
            "type": "object"
        },
        "GetMethodDetails\\Response": {
            "properties": {
                "method": {
                    "$ref": "#/definitions/GetMethodDetails\\Method"
                },
                "success": {
                    "type": "boolean"
                },
                "error": {
                    "$ref": "#/definitions/ResponseError"
                }
            },
            "type": "object"
        },
        "GetMethodDetails\\ServiceLevel": {
            "required": [
                "delivery_price",
                "is_prepayment_free_delivery",
                "rejection_fee",
                "is_active",
                "is_deleted",
                "is_callcenter_confirm",
                "is_autoreserve_allowed",
                "payment_type",
                "has_horizon",
                "has_intervals",
                "is_weekend_included"
            ],
            "properties": {
                "delivery_price": {
                    "type": "number",
                    "format": "float",
                    "example": 250
                },
                "is_prepayment_free_delivery": {
                    "type": "boolean",
                    "example": true
                },
                "rejection_fee": {
                    "type": "number",
                    "format": "float",
                    "example": 250
                },
                "free_delivery_threshold_net": {
                    "type": "number",
                    "format": "float",
                    "example": 2500
                },
                "free_delivery_threshold_gross": {
                    "type": "number",
                    "format": "float"
                },
                "free_delivery_threshold_share": {
                    "type": "integer"
                },
                "max_items": {
                    "type": "integer",
                    "example": 15
                },
                "checkout_description": {
                    "type": "string",
                    "example": "\u0414\u043e\u0441\u0442\u0430\u0432\u043a\u0430 \u0437\u0430\u043a\u0430\u0437\u0430 \u0432 \u043f\u0443\u043d\u043a\u0442 \u0432\u044b\u0434\u0430\u0447\u0438 \u0437\u0430\u043a\u0430\u0437\u043e\u0432 \u0431\u0435\u0441\u043f\u043b\u0430\u0442\u043d\u0430 \u0432 \u0441\u043b\u0443\u0447\u0430\u0435 ..."
                },
                "callcenter_description": {
                    "type": "string",
                    "example": "\u041f\u0440\u0438 \u0432\u044b\u043a\u0443\u043f\u0435 \u043d\u0430 \u0441\u0443\u043c\u043c\u0443 \u0431\u043e\u043b\u0435\u0435 2500 \u0440\u0443\u0431\u043b\u0435\u0439 \u0434\u043e\u0441\u0442\u0430\u0432\u043a\u0430 \u0431\u0435\u0441\u043f\u043b\u0430\u0442\u043d\u0430\u044f ..."
                },
                "is_active": {
                    "type": "boolean",
                    "example": true
                },
                "is_deleted": {
                    "type": "boolean",
                    "example": false
                },
                "is_callcenter_confirm": {
                    "type": "boolean",
                    "example": false
                },
                "is_autoreserve_allowed": {
                    "type": "boolean",
                    "example": true
                },
                "max_cart_amount": {
                    "type": "number",
                    "format": "float"
                },
                "payment_type": {
                    "description": "0 - all types, 1 - prepayment, 2 - postpayment",
                    "type": "integer",
                    "example": 0
                },
                "has_horizon": {
                    "type": "boolean",
                    "example": true
                },
                "has_intervals": {
                    "type": "boolean",
                    "example": true
                },
                "is_weekend_included": {
                    "type": "boolean",
                    "example": false
                },
                "horizon_from": {
                    "type": "integer",
                    "example": 1
                },
                "horizon_till": {
                    "type": "integer",
                    "example": 5
                },
                "service_level_type": {
                    "$ref": "#/definitions/GetMethodDetails\\ServiceLevelType"
                }
            },
            "type": "object"
        },
        "GetMethodDetails\\ServiceLevelType": {
            "required": [
                "code",
                "name",
                "is_tryon_allowed",
                "is_rejection_allowed",
                "is_active",
                "is_deleted",
                "priority"
            ],
            "properties": {
                "code": {
                    "type": "string",
                    "example": "plus"
                },
                "name": {
                    "type": "string",
                    "example": "\u041f\u043b\u044e\u0441"
                },
                "description": {
                    "type": "string"
                },
                "is_tryon_allowed": {
                    "type": "boolean",
                    "example": true
                },
                "is_rejection_allowed": {
                    "type": "boolean",
                    "example": true
                },
                "is_active": {
                    "type": "boolean",
                    "example": true
                },
                "is_deleted": {
                    "type": "boolean",
                    "example": false
                },
                "priority": {
                    "type": "integer",
                    "example": 100
                }
            },
            "type": "object"
        },
        "GetMethodDetails\\Zone": {
            "required": [
                "zone_id",
                "name",
                "is_airfreight",
                "is_no_liquids",
                "operating_area",
                "is_active",
                "is_deleted",
                "is_bankcard_accepted"
            ],
            "properties": {
                "zone_id": {
                    "type": "integer",
                    "example": 104185
                },
                "name": {
                    "type": "string",
                    "example": "\u041d\u043e\u0432\u043e\u043a\u0443\u0437\u043d\u0435\u0446\u043a\u0430\u044f"
                },
                "is_airfreight": {
                    "type": "boolean",
                    "example": false
                },
                "is_no_liquids": {
                    "type": "boolean",
                    "example": false
                },
                "operating_area": {
                    "description": "0 - area specified, 1 - entire country",
                    "type": "integer",
                    "example": 0
                },
                "is_active": {
                    "type": "boolean",
                    "example": true
                },
                "is_deleted": {
                    "type": "boolean",
                    "example": false
                },
                "is_bankcard_accepted": {
                    "type": "boolean",
                    "example": true
                },
                "pickup": {
                    "$ref": "#/definitions/GetMethodDetails\\Pickup"
                },
                "service_levels": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/GetMethodDetails\\ServiceLevel"
                    }
                }
            },
            "type": "object"
        },
        "GetMethodDetailsMultiple\\Response": {
            "properties": {
                "method": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/GetMethodDetails\\Method"
                    }
                },
                "success": {
                    "type": "boolean"
                },
                "error": {
                    "$ref": "#/definitions/ResponseError"
                }
            },
            "type": "object"
        },
        "GetMethodsResponse": {
            "properties": {
                "methods": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/MethodWithIntervals"
                    }
                },
                "success": {
                    "type": "boolean"
                },
                "error": {
                    "$ref": "#/definitions/ResponseError"
                }
            },
            "type": "object"
        },
        "GetPickupResponse": {
            "properties": {
                "pickup_point": {
                    "$ref": "#/definitions/PickupResponse"
                },
                "success": {
                    "type": "boolean"
                },
                "error": {
                    "$ref": "#/definitions/ResponseError"
                }
            },
            "type": "object"
        },
        "GetPupDetailsResponse": {
            "properties": {
                "pickup_point": {
                    "$ref": "#/definitions/PupDetailsResponse"
                },
                "success": {
                    "type": "boolean"
                },
                "error": {
                    "$ref": "#/definitions/ResponseError"
                }
            },
            "type": "object"
        },
        "GetPupListResponse": {
            "properties": {
                "pickups": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/PupListResponse"
                    }
                },
                "success": {
                    "type": "boolean"
                },
                "error": {
                    "$ref": "#/definitions/ResponseError"
                }
            },
            "type": "object"
        },
        "GetReturnMethodsResponse": {
            "properties": {
                "methods": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/ReturnMethod"
                    }
                },
                "success": {
                    "type": "boolean"
                },
                "error": {
                    "$ref": "#/definitions/ResponseError"
                }
            },
            "type": "object"
        },
        "GetZoneAttributesResponse": {
            "properties": {
                "attributes": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/ArrayItem"
                    }
                },
                "success": {
                    "type": "boolean"
                },
                "error": {
                    "$ref": "#/definitions/ResponseError"
                }
            },
            "type": "object"
        },
        "GetZoneRateResponse": {
            "properties": {
                "price": {
                    "description": "Total",
                    "type": "number",
                    "format": "float",
                    "example": 189
                },
                "price_tax": {
                    "type": "number",
                    "format": "float",
                    "example": 28.83
                },
                "mass_rate": {
                    "type": "number",
                    "format": "float",
                    "example": 189
                },
                "mass_rate_tax": {
                    "type": "number",
                    "format": "float",
                    "example": 28.83
                },
                "air_rate": {
                    "type": "number",
                    "format": "float",
                    "example": 0
                },
                "air_rate_tax": {
                    "type": "number",
                    "format": "float",
                    "example": 0
                },
                "insurance_rate": {
                    "type": "number",
                    "format": "float",
                    "example": 0
                },
                "insurance_rate_tax": {
                    "type": "number",
                    "format": "float",
                    "example": 0
                },
                "transportation_type": {
                    "type": "string",
                    "example": "ground"
                },
                "parcel_online": {
                    "type": "boolean",
                    "example": true
                },
                "success": {
                    "type": "boolean"
                },
                "error": {
                    "$ref": "#/definitions/ResponseError"
                }
            },
            "type": "object"
        },
        "Interval": {
            "required": [
                "id",
                "start",
                "end"
            ],
            "properties": {
                "id": {
                    "type": "integer",
                    "example": 40704062
                },
                "start": {
                    "type": "string",
                    "example": "2017-03-04 10:00:00"
                },
                "end": {
                    "type": "string",
                    "example": "2017-03-04 22:00:00"
                }
            },
            "type": "object"
        },
        "IntervalWithPalletMethod": {
            "required": [
                "pallet_method_code"
            ],
            "properties": {
                "pallet_method_code": {
                    "type": "string",
                    "example": "lamoda_showroom_novokuznetskaya"
                },
                "is_available": {
                    "type": "boolean",
                    "example": true
                },
                "id": {
                    "type": "integer",
                    "example": 40704062
                },
                "start": {
                    "type": "string",
                    "example": "2017-03-04 10:00:00"
                },
                "end": {
                    "type": "string",
                    "example": "2017-03-04 22:00:00"
                }
            },
            "type": "object"
        },
        "Method": {
            "required": [
                "code",
                "name",
                "type",
                "group_name",
                "checkout_name"
            ],
            "properties": {
                "code": {
                    "type": "string",
                    "example": "lamoda_showroom_novokuznetskaya"
                },
                "name": {
                    "type": "string",
                    "example": "\u041d\u043e\u0432\u043e\u043a\u0443\u0437\u043d\u0435\u0446\u043a\u0430\u044f, \u043c. \u041d\u043e\u0432\u043e\u043a\u0443\u0437\u043d\u0435\u0446\u043a\u0430\u044f, \u043c.\u0422\u0440\u0435\u0442\u044c\u044f\u043a\u043e\u0432\u0441\u043a\u0430\u044f, \u043c. \u041f\u0430\u0432\u0435\u043b\u0435\u0446\u043a\u0430\u044f"
                },
                "type": {
                    "description": "1 - \u043a\u0443\u0440\u044c\u0435\u0440\u0441\u043a\u0430\u044f \u0434\u043e\u0441\u0442\u0430\u0432\u043a\u0430, 2 - \u0441\u0430\u043c\u043e\u0432\u044b\u0432\u043e\u0437, 3 - \u043f\u043e\u0447\u0442\u043e\u0432\u0430\u044f \u0434\u043e\u0441\u0442\u0430\u0432\u043a\u0430",
                    "type": "integer",
                    "example": 2
                },
                "description": {
                    "type": "string",
                    "example": "- \u041f\u0440\u0438 \u0437\u0430\u043a\u0430\u0437\u0435 \u0434\u043e 24:00, \u0434\u043e\u0441\u0442\u0430\u0432\u043a\u0430 \u043d\u0430 \u0441\u043b\u0435\u0434\u0443\u044e\u0449\u0438\u0439 \u0434\u0435\u043d\u044c..."
                },
                "is_active": {
                    "type": "boolean"
                },
                "is_deleted": {
                    "type": "boolean"
                },
                "group_name": {
                    "type": "string",
                    "example": "\u0421\u0430\u043c\u043e\u0432\u044b\u0432\u043e\u0437"
                },
                "carrier": {
                    "type": "string",
                    "example": "leos-express"
                },
                "is_own": {
                    "type": "boolean",
                    "example": true
                },
                "is_client_name_required": {
                    "type": "boolean",
                    "example": false
                },
                "checkout_name": {
                    "type": "string",
                    "example": "\u0414\u043e\u0441\u0442\u0430\u0432\u043a\u0430 \u0432 \u043f\u0443\u043d\u043a\u0442 \u0432\u044b\u0434\u0430\u0447\u0438 \u0437\u0430\u043a\u0430\u0437\u043e\u0432 Lamoda"
                },
                "max_postpones": {
                    "description": "available in get-all-methods only",
                    "type": "integer",
                    "example": 3
                }
            },
            "type": "object"
        },
        "MethodType": {
            "required": [
                "code",
                "name",
                "service_level_types"
            ],
            "properties": {
                "code": {
                    "type": "string",
                    "example": "pickup"
                },
                "name": {
                    "type": "string",
                    "example": "\u0421\u0430\u043c\u043e\u0432\u044b\u0432\u043e\u0437"
                },
                "service_level_types": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/ServiceLevelType"
                    }
                }
            },
            "type": "object"
        },
        "MethodWithIntervals": {
            "required": [
                "checkout_description",
                "is_callcenter_confirm",
                "is_autoreserve_allowed",
                "has_horizon",
                "has_intervals",
                "delivery_price",
                "zone_id",
                "available_days",
                "is_bankcard_accepted",
                "is_tryon_allowed",
                "is_rejection_allowed",
                "rejection_fee",
                "service_level_code",
                "service_level_name",
                "service_level_description"
            ],
            "properties": {
                "checkout_description": {
                    "type": "string",
                    "example": "\u0414\u043e\u0441\u0442\u0430\u0432\u043a\u0430 \u0437\u0430\u043a\u0430\u0437\u0430 \u0432 \u043f\u0443\u043d\u043a\u0442 \u0432\u044b\u0434\u0430\u0447\u0438 \u0437\u0430\u043a\u0430\u0437\u043e\u0432 \u0431\u0435\u0441\u043f\u043b\u0430\u0442\u043d\u0430 \u0432 \u0441\u043b\u0443\u0447\u0430\u0435 ..."
                },
                "is_callcenter_confirm": {
                    "type": "boolean",
                    "example": false
                },
                "is_autoreserve_allowed": {
                    "description": "If True, the chosen interval gets to be reserved right after the order is created",
                    "type": "boolean"
                },
                "has_horizon": {
                    "type": "boolean"
                },
                "has_intervals": {
                    "type": "boolean"
                },
                "delivery_price": {
                    "type": "number",
                    "format": "float",
                    "example": 250
                },
                "free_delivery_net_threshold": {
                    "type": "number",
                    "format": "float",
                    "example": 2500
                },
                "zone_id": {
                    "type": "string",
                    "example": "104185"
                },
                "macrozone_code": {
                    "type": "string",
                    "example": "e1"
                },
                "pickup_point": {
                    "$ref": "#/definitions/PupResponse"
                },
                "available_days": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/DayWithPalletMethod"
                    }
                },
                "is_bankcard_accepted": {
                    "type": "boolean"
                },
                "is_tryon_allowed": {
                    "type": "boolean"
                },
                "is_rejection_allowed": {
                    "type": "boolean"
                },
                "rejection_fee": {
                    "type": "number",
                    "format": "float",
                    "example": 250
                },
                "delivery_date_min": {
                    "type": "string",
                    "example": "2017-03-04"
                },
                "delivery_date_max": {
                    "type": "string",
                    "example": "2017-03-08"
                },
                "service_level_code": {
                    "type": "string",
                    "example": "plus"
                },
                "service_level_name": {
                    "type": "string",
                    "example": "\u041f\u043b\u044e\u0441"
                },
                "service_level_description": {
                    "type": "string",
                    "example": "\u0414\u043e\u0441\u0442\u0430\u0432\u043a\u0430 \u0431\u0435\u0441\u043f\u043b\u0430\u0442\u043d\u0430\u044f. \u041f\u0440\u0438\u043c\u0435\u0440\u043a\u0430 \u043f\u0435\u0440\u0435\u0434 \u043f\u043e\u043a\u0443\u043f\u043a\u043e\u0439, \u0432\u043e\u0437\u043c\u043e\u0436\u043d\u043e\u0441\u0442\u044c \u0447\u0430\u0441\u0442\u0438\u0447\u043d\u043e\u0433\u043e \u0432\u044b\u043a\u0443\u043f\u0430..."
                },
                "communication_description": {
                    "type": "string",
                    "example": "{\u041a\u043e\u0433\u0434\u0430 \u0437\u0430\u043a\u0430\u0437 \u043f\u043e\u0441\u0442\u0443\u043f\u0438\u0442 \u0432 \u043f\u0443\u043d\u043a\u0442 \u0441\u0430\u043c\u043e\u0432\u044b\u0432\u043e\u0437\u0430, \u0432\u0430\u043c \u043f\u0440\u0438\u0434\u0435\u0442 SMS..."
                },
                "code": {
                    "type": "string",
                    "example": "lamoda_showroom_novokuznetskaya"
                },
                "name": {
                    "type": "string",
                    "example": "\u041d\u043e\u0432\u043e\u043a\u0443\u0437\u043d\u0435\u0446\u043a\u0430\u044f, \u043c. \u041d\u043e\u0432\u043e\u043a\u0443\u0437\u043d\u0435\u0446\u043a\u0430\u044f, \u043c.\u0422\u0440\u0435\u0442\u044c\u044f\u043a\u043e\u0432\u0441\u043a\u0430\u044f, \u043c. \u041f\u0430\u0432\u0435\u043b\u0435\u0446\u043a\u0430\u044f"
                },
                "type": {
                    "description": "1 - \u043a\u0443\u0440\u044c\u0435\u0440\u0441\u043a\u0430\u044f \u0434\u043e\u0441\u0442\u0430\u0432\u043a\u0430, 2 - \u0441\u0430\u043c\u043e\u0432\u044b\u0432\u043e\u0437, 3 - \u043f\u043e\u0447\u0442\u043e\u0432\u0430\u044f \u0434\u043e\u0441\u0442\u0430\u0432\u043a\u0430",
                    "type": "integer",
                    "example": 2
                },
                "description": {
                    "type": "string",
                    "example": "- \u041f\u0440\u0438 \u0437\u0430\u043a\u0430\u0437\u0435 \u0434\u043e 24:00, \u0434\u043e\u0441\u0442\u0430\u0432\u043a\u0430 \u043d\u0430 \u0441\u043b\u0435\u0434\u0443\u044e\u0449\u0438\u0439 \u0434\u0435\u043d\u044c..."
                },
                "is_active": {
                    "type": "boolean"
                },
                "is_deleted": {
                    "type": "boolean"
                },
                "group_name": {
                    "type": "string",
                    "example": "\u0421\u0430\u043c\u043e\u0432\u044b\u0432\u043e\u0437"
                },
                "carrier": {
                    "type": "string",
                    "example": "leos-express"
                },
                "is_own": {
                    "type": "boolean",
                    "example": true
                },
                "is_client_name_required": {
                    "type": "boolean",
                    "example": false
                },
                "checkout_name": {
                    "type": "string",
                    "example": "\u0414\u043e\u0441\u0442\u0430\u0432\u043a\u0430 \u0432 \u043f\u0443\u043d\u043a\u0442 \u0432\u044b\u0434\u0430\u0447\u0438 \u0437\u0430\u043a\u0430\u0437\u043e\u0432 Lamoda"
                },
                "max_postpones": {
                    "description": "available in get-all-methods only",
                    "type": "integer",
                    "example": 3
                }
            },
            "type": "object"
        },
        "PickupResponse": {
            "required": [
                "storage_days",
                "is_returns_accepted",
                "method_code",
                "method_name",
                "method_type",
                "is_bankcard_accepted",
                "is_24hours",
                "checkout_name"
            ],
            "properties": {
                "work_time": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/WorkTime"
                    }
                },
                "storage_days": {
                    "type": "string",
                    "example": 2
                },
                "is_returns_accepted": {
                    "type": "boolean",
                    "example": true
                },
                "widget_description": {
                    "type": "string",
                    "example": "\u0412\u044b\u0445\u043e\u0434\u0438\u0442\u0435 \u0438\u0437 \u0441\u0442\u0430\u043d\u0446\u0438\u0438 \u043c\u0435\u0442\u0440\u043e <b>\u041d\u043e\u0432\u043e\u043a\u0443\u0437\u043d\u0435\u0446\u043a\u0430\u044f </b>\u043d\u0430 \u0443\u043b\u0438\u0446\u0443 \u041d\u043e\u0432\u043e\u043a\u0443\u0437\u043d\u0435\u0446\u043a\u0430\u044f..."
                },
                "phones": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    },
                    "example": [
                        "+7(495) 363-63-93"
                    ]
                },
                "photos": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    },
                    "example": [
                        "//files.lmcdn.ru/delivery/ru/75/6978cffab68e1e0ce126fa0d42368e75.png"
                    ]
                },
                "method_code": {
                    "type": "string",
                    "example": "lamoda_showroom_novokuznetskaya"
                },
                "method_name": {
                    "type": "string",
                    "example": "\u041d\u043e\u0432\u043e\u043a\u0443\u0437\u043d\u0435\u0446\u043a\u0430\u044f, \u043c. \u041d\u043e\u0432\u043e\u043a\u0443\u0437\u043d\u0435\u0446\u043a\u0430\u044f, \u043c.\u0422\u0440\u0435\u0442\u044c\u044f\u043a\u043e\u0432\u0441\u043a\u0430\u044f, \u043c. \u041f\u0430\u0432\u0435\u043b\u0435\u0446\u043a\u0430\u044f"
                },
                "method_type": {
                    "description": "1 - \u043a\u0443\u0440\u044c\u0435\u0440\u0441\u043a\u0430\u044f \u0434\u043e\u0441\u0442\u0430\u0432\u043a\u0430, 2 - \u0441\u0430\u043c\u043e\u0432\u044b\u0432\u043e\u0437, 3 - \u043f\u043e\u0447\u0442\u043e\u0432\u0430\u044f \u0434\u043e\u0441\u0442\u0430\u0432\u043a\u0430",
                    "type": "string",
                    "example": 2
                },
                "latitude": {
                    "type": "number",
                    "format": "float",
                    "example": 55.738176
                },
                "longitude": {
                    "type": "number",
                    "format": "float",
                    "example": 37.631595
                },
                "is_bankcard_accepted": {
                    "type": "boolean",
                    "example": true
                },
                "is_24hours": {
                    "type": "boolean",
                    "example": false
                },
                "group_id": {
                    "type": "integer",
                    "example": 4
                },
                "checkout_name": {
                    "type": "string",
                    "example": "\u0414\u043e\u0441\u0442\u0430\u0432\u043a\u0430 \u0432 \u043f\u0443\u043d\u043a\u0442 \u0432\u044b\u0434\u0430\u0447\u0438 \u0437\u0430\u043a\u0430\u0437\u043e\u0432 Lamoda"
                },
                "communication_description": {
                    "type": "string",
                    "example": "{\u041a\u043e\u0433\u0434\u0430 \u0437\u0430\u043a\u0430\u0437 \u043f\u043e\u0441\u0442\u0443\u043f\u0438\u0442 \u0432 \u043f\u0443\u043d\u043a\u0442 \u0441\u0430\u043c\u043e\u0432\u044b\u0432\u043e\u0437\u0430, \u0432\u0430\u043c \u043f\u0440\u0438\u0434\u0435\u0442 SMS..."
                },
                "lead_time": {
                    "$ref": "#/definitions/LeadTime"
                },
                "service_levels": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/GetMethodDetails\\ServiceLevel"
                    }
                },
                "is_own": {
                    "type": "boolean"
                },
                "tracking_url": {
                    "type": "string"
                },
                "id": {
                    "type": "integer",
                    "example": 16213
                },
                "external_id": {
                    "type": "string"
                },
                "name": {
                    "type": "string",
                    "example": "\u041d\u043e\u0432\u043e\u043a\u0443\u0437\u043d\u0435\u0446\u043a\u0430\u044f"
                },
                "city": {
                    "type": "string",
                    "example": "\u041c\u043e\u0441\u043a\u0432\u0430"
                },
                "street": {
                    "type": "string",
                    "example": "\u041d\u043e\u0432\u043e\u043a\u0443\u0437\u043d\u0435\u0446\u043a\u0430\u044f"
                },
                "house": {
                    "type": "string",
                    "example": "11/13"
                },
                "office": {
                    "type": "string"
                },
                "underground_station": {
                    "type": "string",
                    "example": "\u043c. \u041d\u043e\u0432\u043e\u043a\u0443\u0437\u043d\u0435\u0446\u043a\u0430\u044f/\u0422\u0440\u0435\u0442\u044c\u044f\u043a\u043e\u0432\u0441\u043a\u0430\u044f/\u041f\u0430\u0432\u0435\u043b\u0435\u0446\u043a\u0430\u044f"
                }
            },
            "type": "object"
        },
        "PupDetailsResponse": {
            "required": [
                "storage_days",
                "is_returns_accepted",
                "method_code",
                "method_name",
                "method_type",
                "delivery_price",
                "available_days",
                "rejection_fee",
                "is_client_name_required",
                "service_level_types"
            ],
            "properties": {
                "storage_days": {
                    "type": "string",
                    "example": 2
                },
                "is_returns_accepted": {
                    "type": "boolean",
                    "example": true
                },
                "widget_description": {
                    "type": "string",
                    "example": "\u0412\u044b\u0445\u043e\u0434\u0438\u0442\u0435 \u0438\u0437 \u0441\u0442\u0430\u043d\u0446\u0438\u0438 \u043c\u0435\u0442\u0440\u043e <b>\u041d\u043e\u0432\u043e\u043a\u0443\u0437\u043d\u0435\u0446\u043a\u0430\u044f </b>\u043d\u0430 \u0443\u043b\u0438\u0446\u0443 \u041d\u043e\u0432\u043e\u043a\u0443\u0437\u043d\u0435\u0446\u043a\u0430\u044f..."
                },
                "phones": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    },
                    "example": [
                        "+7(495) 363-63-93"
                    ]
                },
                "photos": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    },
                    "example": [
                        "//files.lmcdn.ru/delivery/ru/75/6978cffab68e1e0ce126fa0d42368e75.png"
                    ]
                },
                "method_code": {
                    "type": "string",
                    "example": "lamoda_showroom_novokuznetskaya"
                },
                "method_name": {
                    "type": "string",
                    "example": "\u041d\u043e\u0432\u043e\u043a\u0443\u0437\u043d\u0435\u0446\u043a\u0430\u044f, \u043c. \u041d\u043e\u0432\u043e\u043a\u0443\u0437\u043d\u0435\u0446\u043a\u0430\u044f, \u043c.\u0422\u0440\u0435\u0442\u044c\u044f\u043a\u043e\u0432\u0441\u043a\u0430\u044f, \u043c. \u041f\u0430\u0432\u0435\u043b\u0435\u0446\u043a\u0430\u044f"
                },
                "method_type": {
                    "description": "1 - \u043a\u0443\u0440\u044c\u0435\u0440\u0441\u043a\u0430\u044f \u0434\u043e\u0441\u0442\u0430\u0432\u043a\u0430, 2 - \u0441\u0430\u043c\u043e\u0432\u044b\u0432\u043e\u0437, 3 - \u043f\u043e\u0447\u0442\u043e\u0432\u0430\u044f \u0434\u043e\u0441\u0442\u0430\u0432\u043a\u0430",
                    "type": "string",
                    "example": 2
                },
                "communication_description": {
                    "type": "string",
                    "example": "{\u041a\u043e\u0433\u0434\u0430 \u0437\u0430\u043a\u0430\u0437 \u043f\u043e\u0441\u0442\u0443\u043f\u0438\u0442 \u0432 \u043f\u0443\u043d\u043a\u0442 \u0441\u0430\u043c\u043e\u0432\u044b\u0432\u043e\u0437\u0430, \u0432\u0430\u043c \u043f\u0440\u0438\u0434\u0435\u0442 SMS..."
                },
                "delivery_price": {
                    "type": "number",
                    "format": "float",
                    "example": 250
                },
                "free_delivery_net_threshold": {
                    "type": "number",
                    "format": "float",
                    "example": 2500
                },
                "available_days": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/DayWithPalletMethod"
                    }
                },
                "checkout_description": {
                    "type": "string",
                    "example": "\u0414\u043e\u0441\u0442\u0430\u0432\u043a\u0430 \u0437\u0430\u043a\u0430\u0437\u0430 \u0432 \u043f\u0443\u043d\u043a\u0442 \u0432\u044b\u0434\u0430\u0447\u0438 \u0437\u0430\u043a\u0430\u0437\u043e\u0432 \u0431\u0435\u0441\u043f\u043b\u0430\u0442\u043d\u0430 \u0432 \u0441\u043b\u0443\u0447\u0430\u0435 ..."
                },
                "callcenter_description": {
                    "type": "string",
                    "example": "\u041f\u0440\u0438 \u0432\u044b\u043a\u0443\u043f\u0435 \u043d\u0430 \u0441\u0443\u043c\u043c\u0443 \u0431\u043e\u043b\u0435\u0435 2500 \u0440\u0443\u0431\u043b\u0435\u0439 \u0434\u043e\u0441\u0442\u0430\u0432\u043a\u0430 \u0431\u0435\u0441\u043f\u043b\u0430\u0442\u043d\u0430\u044f ..."
                },
                "rejection_fee": {
                    "type": "number",
                    "format": "float",
                    "example": 250
                },
                "is_client_name_required": {
                    "type": "boolean",
                    "example": false
                },
                "service_level_types": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/PupDetailsServiceLevelType"
                    }
                },
                "latitude": {
                    "type": "number",
                    "format": "float",
                    "example": 55.738176
                },
                "longitude": {
                    "type": "number",
                    "format": "float",
                    "example": 37.631595
                },
                "is_bankcard_accepted": {
                    "type": "boolean",
                    "example": true
                },
                "is_24hours": {
                    "type": "boolean",
                    "example": false
                },
                "group_id": {
                    "type": "integer",
                    "example": 4
                },
                "work_time": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/WorkTime"
                    }
                },
                "delivery_date_min": {
                    "type": "string",
                    "example": "2017-03-04"
                },
                "delivery_date_max": {
                    "type": "string",
                    "example": "2017-03-08"
                },
                "lead_time": {
                    "$ref": "#/definitions/LeadTime"
                },
                "has_intervals": {
                    "type": "boolean"
                },
                "has_horizon": {
                    "type": "boolean"
                },
                "zone_id": {
                    "type": "integer"
                },
                "id": {
                    "type": "integer",
                    "example": 16213
                },
                "external_id": {
                    "type": "string"
                },
                "name": {
                    "type": "string",
                    "example": "\u041d\u043e\u0432\u043e\u043a\u0443\u0437\u043d\u0435\u0446\u043a\u0430\u044f"
                },
                "city": {
                    "type": "string",
                    "example": "\u041c\u043e\u0441\u043a\u0432\u0430"
                },
                "street": {
                    "type": "string",
                    "example": "\u041d\u043e\u0432\u043e\u043a\u0443\u0437\u043d\u0435\u0446\u043a\u0430\u044f"
                },
                "house": {
                    "type": "string",
                    "example": "11/13"
                },
                "office": {
                    "type": "string"
                },
                "underground_station": {
                    "type": "string",
                    "example": "\u043c. \u041d\u043e\u0432\u043e\u043a\u0443\u0437\u043d\u0435\u0446\u043a\u0430\u044f/\u0422\u0440\u0435\u0442\u044c\u044f\u043a\u043e\u0432\u0441\u043a\u0430\u044f/\u041f\u0430\u0432\u0435\u043b\u0435\u0446\u043a\u0430\u044f"
                }
            },
            "type": "object"
        },
        "PupDetailsServiceLevelType": {
            "required": [
                "code",
                "name",
                "is_tryon_allowed",
                "is_rejection_allowed",
                "has_horizon",
                "has_intervals",
                "available_days",
                "delivery_price",
                "rejection_fee"
            ],
            "properties": {
                "code": {
                    "type": "string",
                    "example": "plus"
                },
                "name": {
                    "type": "string",
                    "example": "\u041f\u043b\u044e\u0441"
                },
                "description": {
                    "type": "string"
                },
                "is_tryon_allowed": {
                    "type": "boolean",
                    "example": true
                },
                "is_rejection_allowed": {
                    "type": "boolean",
                    "example": true
                },
                "free_delivery_net_threshold": {
                    "type": "number",
                    "format": "float",
                    "example": 2500
                },
                "checkout_description": {
                    "type": "string",
                    "example": "\u0414\u043e\u0441\u0442\u0430\u0432\u043a\u0430 \u0437\u0430\u043a\u0430\u0437\u0430 \u0432 \u043f\u0443\u043d\u043a\u0442 \u0432\u044b\u0434\u0430\u0447\u0438 \u0437\u0430\u043a\u0430\u0437\u043e\u0432 \u0431\u0435\u0441\u043f\u043b\u0430\u0442\u043d\u0430 \u0432 \u0441\u043b\u0443\u0447\u0430\u0435 ..."
                },
                "callcenter_description": {
                    "type": "string",
                    "example": "\u041f\u0440\u0438 \u0432\u044b\u043a\u0443\u043f\u0435 \u043d\u0430 \u0441\u0443\u043c\u043c\u0443 \u0431\u043e\u043b\u0435\u0435 2500 \u0440\u0443\u0431\u043b\u0435\u0439 \u0434\u043e\u0441\u0442\u0430\u0432\u043a\u0430 \u0431\u0435\u0441\u043f\u043b\u0430\u0442\u043d\u0430\u044f ..."
                },
                "has_horizon": {
                    "type": "boolean"
                },
                "has_intervals": {
                    "type": "boolean"
                },
                "available_days": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/DayWithPalletMethod"
                    }
                },
                "delivery_price": {
                    "type": "number",
                    "format": "float",
                    "example": 250
                },
                "rejection_fee": {
                    "type": "number",
                    "format": "float",
                    "example": 250
                },
                "day_min": {
                    "type": "integer",
                    "example": 1
                },
                "day_max": {
                    "type": "integer",
                    "example": 5
                },
                "is_callcenter_confirm": {
                    "type": "boolean",
                    "example": false
                }
            },
            "type": "object"
        },
        "PupListResponse": {
            "required": [
                "is_bankcard_accepted",
                "is_24hours",
                "has_intervals",
                "has_horizon"
            ],
            "properties": {
                "latitude": {
                    "type": "number",
                    "format": "float",
                    "example": 55.738176
                },
                "longitude": {
                    "type": "number",
                    "format": "float",
                    "example": 37.631595
                },
                "is_bankcard_accepted": {
                    "type": "boolean",
                    "example": true
                },
                "is_24hours": {
                    "type": "boolean",
                    "example": false
                },
                "group_id": {
                    "type": "integer",
                    "example": 4
                },
                "work_time": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/WorkTime"
                    }
                },
                "delivery_date_min": {
                    "type": "string",
                    "example": "2017-03-04"
                },
                "delivery_date_max": {
                    "type": "string",
                    "example": "2017-03-08"
                },
                "lead_time": {
                    "$ref": "#/definitions/LeadTime"
                },
                "has_intervals": {
                    "type": "boolean"
                },
                "has_horizon": {
                    "type": "boolean"
                },
                "zone_id": {
                    "type": "integer"
                },
                "id": {
                    "type": "integer",
                    "example": 16213
                },
                "external_id": {
                    "type": "string"
                },
                "name": {
                    "type": "string",
                    "example": "\u041d\u043e\u0432\u043e\u043a\u0443\u0437\u043d\u0435\u0446\u043a\u0430\u044f"
                },
                "city": {
                    "type": "string",
                    "example": "\u041c\u043e\u0441\u043a\u0432\u0430"
                },
                "street": {
                    "type": "string",
                    "example": "\u041d\u043e\u0432\u043e\u043a\u0443\u0437\u043d\u0435\u0446\u043a\u0430\u044f"
                },
                "house": {
                    "type": "string",
                    "example": "11/13"
                },
                "office": {
                    "type": "string"
                },
                "underground_station": {
                    "type": "string",
                    "example": "\u043c. \u041d\u043e\u0432\u043e\u043a\u0443\u0437\u043d\u0435\u0446\u043a\u0430\u044f/\u0422\u0440\u0435\u0442\u044c\u044f\u043a\u043e\u0432\u0441\u043a\u0430\u044f/\u041f\u0430\u0432\u0435\u043b\u0435\u0446\u043a\u0430\u044f"
                }
            },
            "type": "object"
        },
        "PupResponse": {
            "required": [
                "id",
                "name",
                "city",
                "street",
                "house"
            ],
            "properties": {
                "id": {
                    "type": "integer",
                    "example": 16213
                },
                "external_id": {
                    "type": "string"
                },
                "name": {
                    "type": "string",
                    "example": "\u041d\u043e\u0432\u043e\u043a\u0443\u0437\u043d\u0435\u0446\u043a\u0430\u044f"
                },
                "city": {
                    "type": "string",
                    "example": "\u041c\u043e\u0441\u043a\u0432\u0430"
                },
                "street": {
                    "type": "string",
                    "example": "\u041d\u043e\u0432\u043e\u043a\u0443\u0437\u043d\u0435\u0446\u043a\u0430\u044f"
                },
                "house": {
                    "type": "string",
                    "example": "11/13"
                },
                "office": {
                    "type": "string"
                },
                "underground_station": {
                    "type": "string",
                    "example": "\u043c. \u041d\u043e\u0432\u043e\u043a\u0443\u0437\u043d\u0435\u0446\u043a\u0430\u044f/\u0422\u0440\u0435\u0442\u044c\u044f\u043a\u043e\u0432\u0441\u043a\u0430\u044f/\u041f\u0430\u0432\u0435\u043b\u0435\u0446\u043a\u0430\u044f"
                }
            },
            "type": "object"
        },
        "WorkTime": {
            "required": [
                "day",
                "time_from",
                "time_to"
            ],
            "properties": {
                "day": {
                    "description": "day of week from 1 to 7",
                    "type": "integer",
                    "example": 1
                },
                "time_from": {
                    "type": "string",
                    "example": "10:00"
                },
                "time_to": {
                    "type": "string",
                    "example": "22:00"
                }
            },
            "type": "object"
        },
        "ReserveResponse": {
            "properties": {
                "interval_id": {
                    "type": "integer",
                    "example": 40704062
                },
                "success": {
                    "type": "boolean"
                },
                "error": {
                    "$ref": "#/definitions/ResponseError"
                }
            },
            "type": "object"
        },
        "Response": {
            "required": [
                "success"
            ],
            "properties": {
                "success": {
                    "type": "boolean"
                },
                "error": {
                    "$ref": "#/definitions/ResponseError"
                }
            },
            "type": "object"
        },
        "ResponseError": {
            "required": [
                "code",
                "message"
            ],
            "properties": {
                "code": {
                    "type": "integer"
                },
                "message": {
                    "type": "string"
                }
            },
            "type": "object"
        },
        "ReturnMethod": {
            "required": [
                "code",
                "name",
                "zones"
            ],
            "properties": {
                "code": {
                    "type": "string",
                    "example": "pups_return"
                },
                "name": {
                    "type": "string",
                    "example": "\u041f\u0443\u043d\u043a\u0442\u044b \u0441\u0430\u043c\u043e\u0432\u044b\u0432\u043e\u0437\u0430"
                },
                "description": {
                    "type": "string"
                },
                "zones": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/ReturnMethodZone"
                    }
                }
            },
            "type": "object"
        },
        "ReturnMethodZone": {
            "required": [
                "id",
                "name"
            ],
            "properties": {
                "id": {
                    "type": "integer",
                    "example": 1
                },
                "name": {
                    "type": "string",
                    "example": "\u041f\u0412\u0417 \u041e\u0440\u0434\u0436\u043e\u043d\u0438\u043a\u0438\u0434\u0436\u0435"
                },
                "description": {
                    "type": "string"
                }
            },
            "type": "object"
        },
        "ServiceLevelType": {
            "required": [
                "code",
                "name",
                "methods"
            ],
            "properties": {
                "code": {
                    "type": "string",
                    "example": "plus"
                },
                "name": {
                    "type": "string",
                    "example": "\u041f\u043b\u044e\u0441"
                },
                "description": {
                    "type": "string"
                },
                "is_tryon_allowed": {
                    "type": "boolean"
                },
                "is_rejection_allowed": {
                    "type": "boolean"
                },
                "methods": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/CheckoutMethod"
                    }
                }
            },
            "type": "object"
        },
        "UpdateZoneRate\\Rate": {
            "required": [
                "weight_min",
                "weight_max",
                "price"
            ],
            "properties": {
                "weight_min": {
                    "type": "number",
                    "format": "float",
                    "example": 0
                },
                "weight_max": {
                    "type": "number",
                    "format": "float",
                    "example": 1
                },
                "price": {
                    "type": "number",
                    "format": "float",
                    "example": 150
                }
            },
            "type": "object"
        },
        "UpdateZoneRate\\Request": {
            "required": [
                "method_code",
                "zones"
            ],
            "properties": {
                "method_code": {
                    "type": "string",
                    "example": "lamoda_showroom_novokuznetskaya"
                },
                "is_update": {
                    "type": "boolean",
                    "example": false
                },
                "zones": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/UpdateZoneRate\\Zone"
                    }
                }
            },
            "type": "object"
        },
        "UpdateZoneRate\\Response": {
            "properties": {
                "success": {
                    "type": "boolean"
                },
                "error": {
                    "$ref": "#/definitions/ResponseError"
                }
            },
            "type": "object"
        },
        "UpdateZoneRate\\Zone": {
            "properties": {
                "zone_id": {
                    "type": "integer",
                    "example": 104185
                },
                "pickup_id": {
                    "type": "integer",
                    "example": 16213
                },
                "pickup_external_id": {
                    "type": "string"
                },
                "rates": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/UpdateZoneRate\\Rate"
                    }
                }
            },
            "type": "object"
        },
        "JsonRpcError": {
            "required": [
                "code",
                "message"
            ],
            "properties": {
                "code": {
                    "type": "integer"
                },
                "message": {
                    "type": "string"
                }
            },
            "type": "object"
        },
        "JsonRpcResponse": {
            "required": [
                "jsonrpc",
                "id"
            ],
            "properties": {
                "jsonrpc": {
                    "type": "string"
                },
                "id": {
                    "type": "string"
                },
                "error": {
                    "$ref": "#/definitions/JsonRpcError"
                }
            },
            "type": "object"
        },
        "JsonRpcOkResponse": {
            "allOf": [
                {
                    "$ref": "#/definitions/JsonRpcResponse"
                },
                {
                    "properties": {
                        "result": {
                            "$ref": "#/definitions/Response"
                        }
                    },
                    "type": "object"
                }
            ]
        },
        "JsonRpcRequest": {
            "required": [
                "jsonrpc",
                "id"
            ],
            "properties": {
                "jsonrpc": {
                    "type": "string"
                },
                "id": {
                    "type": "string"
                }
            },
            "type": "object"
        },
        "JsonRpcMethodsGetResponse": {
            "allOf": [
                {
                    "$ref": "#/definitions/JsonRpcResponse"
                },
                {
                    "properties": {
                        "result": {
                            "$ref": "#/definitions/GetMethodsResponse"
                        }
                    },
                    "type": "object"
                }
            ]
        },
        "JsonRpcMethodsGetCheckoutResponse": {
            "allOf": [
                {
                    "$ref": "#/definitions/JsonRpcResponse"
                },
                {
                    "properties": {
                        "result": {
                            "$ref": "#/definitions/GetCheckoutMethodsResponse"
                        }
                    },
                    "type": "object"
                }
            ]
        },
        "JsonRpcMethodsGetCheckoutMultipleResponse": {
            "allOf": [
                {
                    "$ref": "#/definitions/JsonRpcResponse"
                },
                {
                    "properties": {
                        "result": {
                            "$ref": "#/definitions/GetCheckoutMethodsMultiple\\Response"
                        }
                    },
                    "type": "object"
                }
            ]
        },
        "JsonRpcMethodsGetDetailsResponse": {
            "allOf": [
                {
                    "$ref": "#/definitions/JsonRpcResponse"
                },
                {
                    "properties": {
                        "result": {
                            "$ref": "#/definitions/GetMethodDetails\\Response"
                        }
                    },
                    "type": "object"
                }
            ]
        },
        "JsonRpcMethodsGetDetailsMultipleResponse": {
            "allOf": [
                {
                    "$ref": "#/definitions/JsonRpcResponse"
                },
                {
                    "properties": {
                        "result": {
                            "$ref": "#/definitions/GetMethodDetailsMultiple\\Response"
                        }
                    },
                    "type": "object"
                }
            ]
        },
        "JsonRpcMethodsGetAllResponse": {
            "allOf": [
                {
                    "$ref": "#/definitions/JsonRpcResponse"
                },
                {
                    "properties": {
                        "result": {
                            "$ref": "#/definitions/GetAllMethodsResponse"
                        }
                    },
                    "type": "object"
                }
            ]
        },
        "JsonRpcDeliveryGetInfoResponse": {
            "allOf": [
                {
                    "$ref": "#/definitions/JsonRpcResponse"
                },
                {
                    "properties": {
                        "result": {
                            "$ref": "#/definitions/GetDeliveryInfoResponse"
                        }
                    },
                    "type": "object"
                }
            ]
        },
        "JsonRpcDeliveryGetInfoShortResponse": {
            "allOf": [
                {
                    "$ref": "#/definitions/JsonRpcResponse"
                },
                {
                    "properties": {
                        "result": {
                            "$ref": "#/definitions/GetDeliveryInfoShortResponse"
                        }
                    },
                    "type": "object"
                }
            ]
        },
        "JsonRpcIntervalsReserveResponse": {
            "allOf": [
                {
                    "$ref": "#/definitions/JsonRpcResponse"
                },
                {
                    "properties": {
                        "result": {
                            "$ref": "#/definitions/ReserveResponse"
                        }
                    },
                    "type": "object"
                }
            ]
        },
        "JsonRpcIntervalsGetResponse": {
            "allOf": [
                {
                    "$ref": "#/definitions/JsonRpcResponse"
                },
                {
                    "properties": {
                        "result": {
                            "$ref": "#/definitions/GetIntervalResponse"
                        }
                    },
                    "type": "object"
                }
            ]
        },
        "JsonRpcIntervalsFindResponse": {
            "allOf": [
                {
                    "$ref": "#/definitions/JsonRpcResponse"
                },
                {
                    "properties": {
                        "result": {
                            "$ref": "#/definitions/FindIntervalResponse"
                        }
                    },
                    "type": "object"
                }
            ]
        },
        "JsonRpcPickupsGetListResponse": {
            "allOf": [
                {
                    "$ref": "#/definitions/JsonRpcResponse"
                },
                {
                    "properties": {
                        "result": {
                            "$ref": "#/definitions/GetPupListResponse"
                        }
                    },
                    "type": "object"
                }
            ]
        },
        "JsonRpcPickupsGetResponse": {
            "allOf": [
                {
                    "$ref": "#/definitions/JsonRpcResponse"
                },
                {
                    "properties": {
                        "result": {
                            "$ref": "#/definitions/GetPickupResponse"
                        }
                    },
                    "type": "object"
                }
            ]
        },
        "JsonRpcPickupsGetDetailsResponse": {
            "allOf": [
                {
                    "$ref": "#/definitions/JsonRpcResponse"
                },
                {
                    "properties": {
                        "result": {
                            "$ref": "#/definitions/GetPupDetailsResponse"
                        }
                    },
                    "type": "object"
                }
            ]
        },
        "JsonRpcPickupsCreateRequest": {
            "allOf": [
                {
                    "$ref": "#/definitions/JsonRpcRequest"
                },
                {
                    "properties": {
                        "params": {
                            "$ref": "#/definitions/CreatePickups\\Request"
                        }
                    },
                    "type": "object"
                }
            ]
        },
        "JsonRpcPickupsCreateResponse": {
            "allOf": [
                {
                    "$ref": "#/definitions/JsonRpcResponse"
                },
                {
                    "properties": {
                        "result": {
                            "$ref": "#/definitions/CreatePickups\\Response"
                        }
                    },
                    "type": "object"
                }
            ]
        },
        "JsonRpcProfilesImportRequest": {
            "allOf": [
                {
                    "$ref": "#/definitions/JsonRpcRequest"
                },
                {
                    "properties": {
                        "params": {
                            "type": "object"
                        }
                    },
                    "type": "object"
                }
            ]
        },
        "JsonRpcZonesUpdateRateRequest": {
            "allOf": [
                {
                    "$ref": "#/definitions/JsonRpcRequest"
                },
                {
                    "properties": {
                        "params": {
                            "$ref": "#/definitions/UpdateZoneRate\\Request"
                        }
                    },
                    "type": "object"
                }
            ]
        },
        "JsonRpcZonesUpdateRateResponse": {
            "allOf": [
                {
                    "$ref": "#/definitions/JsonRpcResponse"
                },
                {
                    "properties": {
                        "result": {
                            "$ref": "#/definitions/UpdateZoneRate\\Response"
                        }
                    },
                    "type": "object"
                }
            ]
        },
        "JsonRpcZonesGetAttributesResponse": {
            "allOf": [
                {
                    "$ref": "#/definitions/JsonRpcResponse"
                },
                {
                    "properties": {
                        "result": {
                            "$ref": "#/definitions/GetZoneAttributesResponse"
                        }
                    },
                    "type": "object"
                }
            ]
        },
        "JsonRpcZonesGetRateResponse": {
            "allOf": [
                {
                    "$ref": "#/definitions/JsonRpcResponse"
                },
                {
                    "properties": {
                        "result": {
                            "$ref": "#/definitions/GetCheckoutMethodsMultiple\\Response"
                        }
                    },
                    "type": "object"
                }
            ]
        }
    },
    "parameters": {
        "aoid": {
            "name": "aoid",
            "in": "query",
            "type": "string"
        },
        "cart_amount": {
            "name": "cart_amount",
            "in": "query",
            "type": "integer"
        },
        "checkout": {
            "name": "checkout",
            "in": "query",
            "type": "boolean",
            "default": false
        },
        "code": {
            "name": "code",
            "in": "query",
            "required": true,
            "type": "string"
        },
        "company": {
            "name": "company",
            "in": "query",
            "type": "string"
        },
        "current_id": {
            "name": "current_id",
            "in": "query",
            "type": "integer"
        },
        "desired_end": {
            "name": "desired_end",
            "in": "query",
            "description": "YYYY-MM-DD HH:MM:SS",
            "required": true,
            "type": "string"
        },
        "desired_start": {
            "name": "desired_start",
            "in": "query",
            "description": "YYYY-MM-DD HH:MM:SS",
            "required": true,
            "type": "string"
        },
        "force": {
            "name": "force",
            "in": "query",
            "type": "boolean",
            "default": false
        },
        "get_method_details_multiple_aoid": {
            "name": "getMethodDetails[0][aoid]",
            "in": "query",
            "type": "string"
        },
        "get_method_details_multiple_code": {
            "name": "getMethodDetails[0][code]",
            "in": "query",
            "required": true,
            "type": "string"
        },
        "get_method_details_multiple_join": {
            "name": "getMethodDetails[0][join]",
            "in": "query",
            "description": "Comma-separated list of values. Supported values may vary in methods. Possible values: company, zone, pickup, service_level, lead_time.",
            "type": "string"
        },
        "get_method_details_multiple_origin_date": {
            "name": "getMethodDetails[0][origin_date]",
            "in": "query",
            "description": "The origin of horizon, default is current date",
            "type": "string",
            "format": "date"
        },
        "get_method_details_multiple_service_level": {
            "name": "getMethodDetails[0][service_level]",
            "in": "query",
            "type": "string"
        },
        "get_method_details_multiple_zipcode": {
            "name": "getMethodDetails[0][zipcode]",
            "in": "query",
            "type": "string"
        },
        "get_method_details_multiple_zone_id": {
            "name": "getMethodDetails[0][zone_id]",
            "in": "query",
            "type": "integer"
        },
        "ignore_capacity": {
            "name": "ignore_capacity",
            "in": "query",
            "type": "boolean",
            "default": false
        },
        "ignore_cutoff": {
            "name": "ignore_cutoff",
            "in": "query",
            "type": "boolean",
            "default": false
        },
        "insurance_sum": {
            "name": "insurance_sum",
            "in": "query",
            "type": "number",
            "format": "float"
        },
        "interval_id": {
            "name": "interval_id",
            "in": "query",
            "required": true,
            "type": "integer"
        },
        "item_count": {
            "name": "item_count",
            "in": "query",
            "type": "integer"
        },
        "join": {
            "name": "join",
            "in": "query",
            "description": "Comma-separated list of values. Supported values may vary in methods. Possible values: company, zone, pickup, service_level, lead_time.",
            "type": "string"
        },
        "latitude": {
            "name": "latitude",
            "in": "query",
            "type": "number"
        },
        "liquids": {
            "name": "liquids",
            "in": "query",
            "type": "boolean",
            "default": false
        },
        "longitude": {
            "name": "longitude",
            "in": "query",
            "type": "number"
        },
        "method_code": {
            "name": "method_code",
            "in": "query",
            "type": "string"
        },
        "method_code_required": {
            "name": "method_code",
            "in": "query",
            "required": true,
            "type": "string"
        },
        "no_airfreight": {
            "name": "no_airfreight",
            "in": "query",
            "type": "boolean",
            "default": false
        },
        "orders_cart_amount": {
            "name": "orders[0][cart_amount]",
            "in": "query",
            "type": "integer"
        },
        "orders_item_count": {
            "name": "orders[0][item_count]",
            "in": "query",
            "type": "integer"
        },
        "orders_no_airfreight": {
            "name": "orders[0][no_airfreight]",
            "in": "query",
            "type": "boolean",
            "default": false
        },
        "orders_payment_type": {
            "name": "orders[0][payment_type]",
            "in": "query",
            "description": "0 - all types<br />1 - prepayment<br />2 - postpayment",
            "type": "integer",
            "default": 2
        },
        "orders_profile_id": {
            "name": "orders[0][profile_id]",
            "in": "query",
            "type": "string",
            "default": "LM"
        },
        "origin_date": {
            "name": "origin_date",
            "in": "query",
            "description": "The origin of horizon, default is current date",
            "type": "string",
            "format": "date"
        },
        "original_delivery_date": {
            "name": "original_delivery_date",
            "in": "query",
            "description": "If provided, results will be filtered to content only dates equal to or later than this",
            "type": "string",
            "format": "date"
        },
        "pallet_method_code": {
            "name": "pallet_method_code",
            "in": "query",
            "type": "string"
        },
        "payment_type": {
            "name": "payment_type",
            "in": "query",
            "description": "0 - all types<br />1 - prepayment<br />2 - postpayment",
            "type": "integer",
            "default": 2
        },
        "pickup_id": {
            "name": "pickup_id",
            "in": "query",
            "required": true,
            "type": "integer"
        },
        "profile_id": {
            "name": "profile_id",
            "in": "query",
            "type": "string",
            "default": "LM"
        },
        "reserved_interval_id": {
            "name": "reserved_interval_id",
            "in": "query",
            "type": "integer"
        },
        "service_level": {
            "name": "service_level",
            "in": "query",
            "type": "string"
        },
        "service_level_type_code": {
            "name": "service_level_type_code",
            "in": "query",
            "type": "string"
        },
        "tag": {
            "name": "tag",
            "in": "query",
            "type": "string"
        },
        "weight": {
            "name": "weight",
            "in": "query",
            "required": true,
            "type": "number",
            "format": "float"
        },
        "zipcode": {
            "name": "zipcode",
            "in": "query",
            "type": "string"
        },
        "zipcode_required": {
            "name": "zipcode",
            "in": "query",
            "required": true,
            "type": "string"
        },
        "zone_id": {
            "name": "zone_id",
            "in": "query",
            "type": "integer"
        },
        "zone_id_required": {
            "name": "zone_id",
            "in": "query",
            "required": true,
            "type": "integer"
        },
        "jsonrpc_id": {
            "name": "id",
            "in": "query",
            "required": true,
            "type": "string"
        }
    },
    "tags": [
        {
            "name": "rest",
            "description": "Rest API"
        },
        {
            "name": "json-rpc",
            "description": "JSON-RPC API"
        },
        {
            "name": "delivery",
            "description": "Delivery info API"
        },
        {
            "name": "intervals",
            "description": "Intervals API"
        },
        {
            "name": "methods",
            "description": "Methods API"
        },
        {
            "name": "pickups",
            "description": "Pickups API"
        },
        {
            "name": "profiles",
            "description": "Profiles API"
        },
        {
            "name": "zones",
            "description": "Zones API"
        }
    ]
}
`

var complexJson2 = `
{
    "swagger": "2.0",
    "info": {
        "title": "LEOS.Delivery API",
        "contact": {
            "url": "https://confluence.lamoda.ru/display/DELY/"
        },
        "version": "1.0.0"
    },
    "produces": [
        "application/json"
    ],
    "paths": {
        "/api/get-methods": {
            "get": {
                "tags": [
                    "rest"
                ],
                "summary": "Get methods",
                "operationId": "getMethods",
                "parameters": [
                    {
                        "$ref": "#/parameters/zipcode"
                    },
                    {
                        "$ref": "#/parameters/aoid"
                    },
                    {
                        "$ref": "#/parameters/latitude"
                    },
                    {
                        "$ref": "#/parameters/longitude"
                    },
                    {
                        "$ref": "#/parameters/item_count"
                    },
                    {
                        "$ref": "#/parameters/cart_amount"
                    },
                    {
                        "$ref": "#/parameters/payment_type"
                    },
                    {
                        "$ref": "#/parameters/no_airfreight"
                    },
                    {
                        "$ref": "#/parameters/liquids"
                    },
                    {
                        "$ref": "#/parameters/checkout"
                    },
                    {
                        "$ref": "#/parameters/method_code"
                    },
                    {
                        "$ref": "#/parameters/company"
                    },
                    {
                        "$ref": "#/parameters/origin_date"
                    },
                    {
                        "$ref": "#/parameters/profile_id"
                    },
                    {
                        "$ref": "#/parameters/ignore_capacity"
                    },
                    {
                        "$ref": "#/parameters/ignore_cutoff"
                    },
                    {
                        "$ref": "#/parameters/reserved_interval_id"
                    },
                    {
                        "$ref": "#/parameters/original_delivery_date"
                    },
                    {
                        "$ref": "#/parameters/tag"
                    }
                ],
                "responses": {
                    "default": {
                        "description": "Success/error response",
                        "schema": {
                            "$ref": "#/definitions/GetMethodsResponse"
                        }
                    }
                }
            }
        },
        "/api/get-checkout-methods": {
            "get": {
                "tags": [
                    "rest"
                ],
                "summary": "Get checkout methods",
                "operationId": "getCheckoutMethods",
                "parameters": [
                    {
                        "$ref": "#/parameters/zipcode"
                    },
                    {
                        "$ref": "#/parameters/aoid"
                    },
                    {
                        "$ref": "#/parameters/zone_id"
                    },
                    {
                        "$ref": "#/parameters/latitude"
                    },
                    {
                        "$ref": "#/parameters/longitude"
                    },
                    {
                        "$ref": "#/parameters/item_count"
                    },
                    {
                        "$ref": "#/parameters/cart_amount"
                    },
                    {
                        "$ref": "#/parameters/payment_type"
                    },
                    {
                        "$ref": "#/parameters/no_airfreight"
                    },
                    {
                        "$ref": "#/parameters/liquids"
                    },
                    {
                        "$ref": "#/parameters/checkout"
                    },
                    {
                        "$ref": "#/parameters/method_code"
                    },
                    {
                        "$ref": "#/parameters/company"
                    },
                    {
                        "$ref": "#/parameters/origin_date"
                    },
                    {
                        "$ref": "#/parameters/profile_id"
                    },
                    {
                        "$ref": "#/parameters/ignore_capacity"
                    },
                    {
                        "$ref": "#/parameters/ignore_cutoff"
                    },
                    {
                        "$ref": "#/parameters/reserved_interval_id"
                    },
                    {
                        "$ref": "#/parameters/original_delivery_date"
                    },
                    {
                        "$ref": "#/parameters/tag"
                    }
                ],
                "responses": {
                    "default": {
                        "description": "Success/error response",
                        "schema": {
                            "$ref": "#/definitions/GetCheckoutMethodsResponse"
                        }
                    }
                }
            }
        },
        "/api/get-checkout-methods-multiple": {
            "get": {
                "tags": [
                    "rest"
                ],
                "summary": "Get checkout methods for multiple orders",
                "description": "Two-dimensional orders array defines parameters for each individual order.\nAt least one order have to be specified with one or more parameters.",
                "operationId": "getCheckoutMethodsMultiple",
                "parameters": [
                    {
                        "$ref": "#/parameters/zipcode"
                    },
                    {
                        "$ref": "#/parameters/aoid"
                    },
                    {
                        "$ref": "#/parameters/zone_id"
                    },
                    {
                        "$ref": "#/parameters/latitude"
                    },
                    {
                        "$ref": "#/parameters/longitude"
                    },
                    {
                        "$ref": "#/parameters/checkout"
                    },
                    {
                        "$ref": "#/parameters/method_code"
                    },
                    {
                        "$ref": "#/parameters/origin_date"
                    },
                    {
                        "$ref": "#/parameters/ignore_capacity"
                    },
                    {
                        "$ref": "#/parameters/ignore_cutoff"
                    },
                    {
                        "$ref": "#/parameters/orders_profile_id"
                    },
                    {
                        "$ref": "#/parameters/orders_item_count"
                    },
                    {
                        "$ref": "#/parameters/orders_cart_amount"
                    },
                    {
                        "$ref": "#/parameters/orders_no_airfreight"
                    },
                    {
                        "$ref": "#/parameters/orders_payment_type"
                    },
                    {
                        "$ref": "#/parameters/tag"
                    }
                ],
                "responses": {
                    "default": {
                        "description": "Success/error response.\n<br />\nReturn list of orders in the same sequence as in request.",
                        "schema": {
                            "$ref": "#/definitions/GetCheckoutMethodsMultiple\\Response"
                        }
                    }
                }
            }
        },
        "/api/reserve": {
            "get": {
                "tags": [
                    "rest"
                ],
                "summary": "Reserve interval",
                "operationId": "reserve",
                "parameters": [
                    {
                        "$ref": "#/parameters/interval_id"
                    },
                    {
                        "$ref": "#/parameters/pallet_method_code"
                    },
                    {
                        "$ref": "#/parameters/profile_id"
                    },
                    {
                        "$ref": "#/parameters/force"
                    },
                    {
                        "$ref": "#/parameters/tag"
                    }
                ],
                "responses": {
                    "default": {
                        "description": "Success/error response",
                        "schema": {
                            "$ref": "#/definitions/ReserveResponse"
                        }
                    }
                }
            }
        },
        "/api/free": {
            "get": {
                "tags": [
                    "rest"
                ],
                "summary": "Free interval",
                "operationId": "free",
                "parameters": [
                    {
                        "$ref": "#/parameters/interval_id"
                    },
                    {
                        "$ref": "#/parameters/pallet_method_code"
                    },
                    {
                        "$ref": "#/parameters/profile_id"
                    },
                    {
                        "$ref": "#/parameters/tag"
                    }
                ],
                "responses": {
                    "default": {
                        "description": "Success/error response",
                        "schema": {
                            "$ref": "#/definitions/Response"
                        }
                    }
                }
            }
        },
        "/api/find-interval": {
            "get": {
                "tags": [
                    "rest"
                ],
                "summary": "Find interval within the same zone",
                "description": "Required either current_id or pair zone_id + service_level_type_code.",
                "operationId": "findInterval",
                "parameters": [
                    {
                        "$ref": "#/parameters/desired_start"
                    },
                    {
                        "$ref": "#/parameters/desired_end"
                    },
                    {
                        "$ref": "#/parameters/current_id"
                    },
                    {
                        "$ref": "#/parameters/zone_id"
                    },
                    {
                        "$ref": "#/parameters/service_level_type_code"
                    }
                ],
                "responses": {
                    "default": {
                        "description": "Success/error response",
                        "schema": {
                            "$ref": "#/definitions/FindIntervalResponse"
                        }
                    }
                }
            }
        },
        "/api/get-interval": {
            "get": {
                "tags": [
                    "rest"
                ],
                "summary": "Get interval",
                "operationId": "getInterval",
                "parameters": [
                    {
                        "$ref": "#/parameters/interval_id"
                    },
                    {
                        "$ref": "#/parameters/tag"
                    }
                ],
                "responses": {
                    "default": {
                        "description": "Success/error response",
                        "schema": {
                            "$ref": "#/definitions/GetIntervalResponse"
                        }
                    }
                }
            }
        },
        "/api/get-method-details": {
            "get": {
                "tags": [
                    "rest"
                ],
                "summary": "Get method details",
                "operationId": "getMethodDetails",
                "parameters": [
                    {
                        "$ref": "#/parameters/code"
                    },
                    {
                        "$ref": "#/parameters/join"
                    },
                    {
                        "$ref": "#/parameters/zone_id"
                    },
                    {
                        "$ref": "#/parameters/service_level"
                    },
                    {
                        "$ref": "#/parameters/zipcode"
                    },
                    {
                        "$ref": "#/parameters/aoid"
                    },
                    {
                        "$ref": "#/parameters/origin_date"
                    },
                    {
                        "$ref": "#/parameters/profile_id"
                    }
                ],
                "responses": {
                    "default": {
                        "description": "Success/error response",
                        "schema": {
                            "$ref": "#/definitions/GetMethodDetails\\Response"
                        }
                    }
                }
            }
        },
        "/api/get-method-details-multiple": {
            "get": {
                "tags": [
                    "rest"
                ],
                "summary": "Get method details multiple",
                "operationId": "getMethodDetailsMultiple",
                "parameters": [
                    {
                        "$ref": "#/parameters/get_method_details_multiple_code"
                    },
                    {
                        "$ref": "#/parameters/get_method_details_multiple_join"
                    },
                    {
                        "$ref": "#/parameters/get_method_details_multiple_zone_id"
                    },
                    {
                        "$ref": "#/parameters/get_method_details_multiple_service_level"
                    },
                    {
                        "$ref": "#/parameters/get_method_details_multiple_zipcode"
                    },
                    {
                        "$ref": "#/parameters/get_method_details_multiple_aoid"
                    },
                    {
                        "$ref": "#/parameters/get_method_details_multiple_origin_date"
                    },
                    {
                        "$ref": "#/parameters/profile_id"
                    }
                ],
                "responses": {
                    "default": {
                        "description": "Success/error response.\n<br />\nReturn list of methods in the same sequence as in request.",
                        "schema": {
                            "$ref": "#/definitions/GetMethodDetailsMultiple\\Response"
                        }
                    }
                }
            }
        },
        "/api/get-all-methods": {
            "get": {
                "tags": [
                    "rest"
                ],
                "summary": "Get all methods",
                "operationId": "getAllMethods",
                "responses": {
                    "default": {
                        "description": "Success/error response",
                        "schema": {
                            "$ref": "#/definitions/GetAllMethodsResponse"
                        }
                    }
                }
            }
        },
        "/api/get-pup-list": {
            "get": {
                "tags": [
                    "rest"
                ],
                "summary": "Get pup list",
                "operationId": "getPupList",
                "parameters": [
                    {
                        "$ref": "#/parameters/zipcode"
                    },
                    {
                        "$ref": "#/parameters/aoid"
                    },
                    {
                        "$ref": "#/parameters/latitude"
                    },
                    {
                        "$ref": "#/parameters/longitude"
                    },
                    {
                        "$ref": "#/parameters/item_count"
                    },
                    {
                        "$ref": "#/parameters/cart_amount"
                    },
                    {
                        "$ref": "#/parameters/payment_type"
                    },
                    {
                        "$ref": "#/parameters/no_airfreight"
                    },
                    {
                        "$ref": "#/parameters/liquids"
                    },
                    {
                        "$ref": "#/parameters/checkout"
                    },
                    {
                        "$ref": "#/parameters/method_code"
                    },
                    {
                        "$ref": "#/parameters/company"
                    },
                    {
                        "$ref": "#/parameters/origin_date"
                    },
                    {
                        "$ref": "#/parameters/profile_id"
                    },
                    {
                        "$ref": "#/parameters/ignore_capacity"
                    },
                    {
                        "$ref": "#/parameters/ignore_cutoff"
                    },
                    {
                        "$ref": "#/parameters/reserved_interval_id"
                    },
                    {
                        "$ref": "#/parameters/original_delivery_date"
                    },
                    {
                        "$ref": "#/parameters/tag"
                    }
                ],
                "responses": {
                    "default": {
                        "description": "Success/error response",
                        "schema": {
                            "$ref": "#/definitions/GetPupListResponse"
                        }
                    }
                }
            }
        },
        "/api/get-pup-details": {
            "get": {
                "tags": [
                    "rest"
                ],
                "summary": "Get pup details",
                "operationId": "getPupDetails",
                "parameters": [
                    {
                        "$ref": "#/parameters/pickup_id"
                    },
                    {
                        "$ref": "#/parameters/join"
                    },
                    {
                        "$ref": "#/parameters/zipcode"
                    },
                    {
                        "$ref": "#/parameters/aoid"
                    },
                    {
                        "$ref": "#/parameters/latitude"
                    },
                    {
                        "$ref": "#/parameters/longitude"
                    },
                    {
                        "$ref": "#/parameters/item_count"
                    },
                    {
                        "$ref": "#/parameters/cart_amount"
                    },
                    {
                        "$ref": "#/parameters/payment_type"
                    },
                    {
                        "$ref": "#/parameters/no_airfreight"
                    },
                    {
                        "$ref": "#/parameters/liquids"
                    },
                    {
                        "$ref": "#/parameters/checkout"
                    },
                    {
                        "$ref": "#/parameters/method_code"
                    },
                    {
                        "$ref": "#/parameters/company"
                    },
                    {
                        "$ref": "#/parameters/origin_date"
                    },
                    {
                        "$ref": "#/parameters/profile_id"
                    },
                    {
                        "$ref": "#/parameters/ignore_capacity"
                    },
                    {
                        "$ref": "#/parameters/ignore_cutoff"
                    },
                    {
                        "$ref": "#/parameters/reserved_interval_id"
                    },
                    {
                        "$ref": "#/parameters/original_delivery_date"
                    },
                    {
                        "$ref": "#/parameters/tag"
                    }
                ],
                "responses": {
                    "default": {
                        "description": "Success/error response",
                        "schema": {
                            "$ref": "#/definitions/GetPupDetailsResponse"
                        }
                    }
                }
            }
        },
        "/api/get-pickup": {
            "get": {
                "tags": [
                    "rest"
                ],
                "summary": "Get pickup",
                "operationId": "getPickup",
                "parameters": [
                    {
                        "$ref": "#/parameters/pickup_id"
                    },
                    {
                        "$ref": "#/parameters/join"
                    },
                    {
                        "$ref": "#/parameters/zipcode"
                    },
                    {
                        "$ref": "#/parameters/aoid"
                    },
                    {
                        "$ref": "#/parameters/profile_id"
                    }
                ],
                "responses": {
                    "default": {
                        "description": "Success/error response",
                        "schema": {
                            "$ref": "#/definitions/GetPickupResponse"
                        }
                    }
                }
            }
        },
        "/api/get-return-methods": {
            "get": {
                "tags": [
                    "rest"
                ],
                "summary": "Get return methods",
                "operationId": "getReturnMethods",
                "parameters": [
                    {
                        "$ref": "#/parameters/zipcode_required"
                    }
                ],
                "responses": {
                    "default": {
                        "description": "Success/error response",
                        "schema": {
                            "$ref": "#/definitions/GetReturnMethodsResponse"
                        }
                    }
                }
            }
        },
        "/api/get-zone-rate": {
            "get": {
                "tags": [
                    "rest"
                ],
                "summary": "Get zone rate",
                "description": "Either zipcode or zone_id parameter is required.\nIf zone_id is specified then it is used instead of zipcode to find required zone.",
                "operationId": "getZoneRate",
                "parameters": [
                    {
                        "$ref": "#/parameters/method_code_required"
                    },
                    {
                        "$ref": "#/parameters/zipcode"
                    },
                    {
                        "$ref": "#/parameters/zone_id"
                    },
                    {
                        "$ref": "#/parameters/weight"
                    },
                    {
                        "$ref": "#/parameters/insurance_sum"
                    }
                ],
                "responses": {
                    "default": {
                        "description": "Success/error response",
                        "schema": {
                            "$ref": "#/definitions/GetZoneRateResponse"
                        }
                    }
                }
            }
        },
        "/api/get-delivery-info": {
            "get": {
                "tags": [
                    "rest"
                ],
                "summary": "Get delivery info",
                "operationId": "getDeliveryInfo",
                "parameters": [
                    {
                        "$ref": "#/parameters/zipcode"
                    },
                    {
                        "$ref": "#/parameters/aoid"
                    },
                    {
                        "$ref": "#/parameters/profile_id2"
                    },
                    {
                        "$ref": "#/parameters/no_airfreight"
                    }
                ],
                "responses": {
                    "default": {
                        "description": "Success/error response",
                        "schema": {
                            "$ref": "#/definitions/GetDeliveryInfoResponse"
                        }
                    }
                }
            }
        },
        "/api/get-delivery-info-short": {
            "get": {
                "tags": [
                    "rest"
                ],
                "summary": "Get short terms of delivery",
                "operationId": "getDeliveryInfoShort",
                "parameters": [
                    {
                        "$ref": "#/parameters/zipcode"
                    },
                    {
                        "$ref": "#/parameters/aoid"
                    },
                    {
                        "$ref": "#/parameters/profile_id"
                    },
                    {
                        "$ref": "#/parameters/no_airfreight"
                    }
                ],
                "responses": {
                    "default": {
                        "description": "Success/error response",
                        "schema": {
                            "$ref": "#/definitions/GetDeliveryInfoShortResponse"
                        }
                    }
                }
            }
        },
        "/api/get-zone-attributes": {
            "get": {
                "tags": [
                    "rest"
                ],
                "summary": "Get zone attributes",
                "operationId": "getZoneAttributes",
                "parameters": [
                    {
                        "$ref": "#/parameters/zone_id_required"
                    }
                ],
                "responses": {
                    "default": {
                        "description": "Success/error response",
                        "schema": {
                            "$ref": "#/definitions/GetZoneAttributesResponse"
                        }
                    }
                }
            }
        },
        "/api/create-pickups": {
            "post": {
                "tags": [
                    "rest"
                ],
                "summary": "Create pickups",
                "operationId": "createPickups",
                "parameters": [
                    {
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/CreatePickups\\Request"
                        }
                    }
                ],
                "responses": {
                    "default": {
                        "description": "Success/error response",
                        "schema": {
                            "$ref": "#/definitions/CreatePickups\\Response"
                        }
                    }
                }
            }
        },
        "/api/update-zone-rate": {
            "post": {
                "tags": [
                    "rest"
                ],
                "summary": "Update zone rate",
                "description": "Zone should have at least one search criteria: zone_id|pickup_id|pickup_external_id",
                "operationId": "updateZoneRate",
                "parameters": [
                    {
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/UpdateZoneRate\\Request"
                        }
                    }
                ],
                "responses": {
                    "default": {
                        "description": "Success/error response",
                        "schema": {
                            "$ref": "#/definitions/UpdateZoneRate\\Response"
                        }
                    }
                }
            }
        },
        "/api/import-profiles": {
            "post": {
                "tags": [
                    "rest"
                ],
                "summary": "Import profiles",
                "operationId": "importProfiles",
                "parameters": [
                    {
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "type": "object"
                        }
                    }
                ],
                "responses": {
                    "default": {
                        "description": "Success/error response",
                        "schema": {
                            "$ref": "#/definitions/Response"
                        }
                    }
                }
            }
        },
        "/jsonrpc/v1/delivery.get-info": {
            "get": {
                "tags": [
                    "json-rpc",
                    "delivery"
                ],
                "summary": "Get delivery info",
                "operationId": "delivery.getInfo",
                "parameters": [
                    {
                        "$ref": "#/parameters/jsonrpc_id"
                    },
                    {
                        "$ref": "#/parameters/zipcode"
                    },
                    {
                        "$ref": "#/parameters/aoid"
                    },
                    {
                        "$ref": "#/parameters/profile_id"
                    },
                    {
                        "$ref": "#/parameters/no_airfreight"
                    }
                ],
                "responses": {
                    "default": {
                        "description": "Success/error response",
                        "schema": {
                            "$ref": "#/definitions/JsonRpcDeliveryGetInfoResponse"
                        }
                    }
                }
            }
        },
        "/jsonrpc/v1/delivery.get-info-short": {
            "get": {
                "tags": [
                    "json-rpc",
                    "delivery"
                ],
                "summary": "Get short terms of delivery",
                "operationId": "delivery.getInfoShort",
                "parameters": [
                    {
                        "$ref": "#/parameters/jsonrpc_id"
                    },
                    {
                        "$ref": "#/parameters/zipcode"
                    },
                    {
                        "$ref": "#/parameters/aoid"
                    },
                    {
                        "$ref": "#/parameters/profile_id"
                    },
                    {
                        "$ref": "#/parameters/no_airfreight"
                    }
                ],
                "responses": {
                    "default": {
                        "description": "Success/error response",
                        "schema": {
                            "$ref": "#/definitions/JsonRpcDeliveryGetInfoShortResponse"
                        }
                    }
                }
            }
        },
        "/jsonrpc/v1/intervals.reserve": {
            "get": {
                "tags": [
                    "json-rpc",
                    "intervals"
                ],
                "summary": "Reserve interval",
                "operationId": "intervals.reserve",
                "parameters": [
                    {
                        "$ref": "#/parameters/jsonrpc_id"
                    },
                    {
                        "$ref": "#/parameters/interval_id"
                    },
                    {
                        "$ref": "#/parameters/pallet_method_code"
                    },
                    {
                        "$ref": "#/parameters/force"
                    },
                    {
                        "$ref": "#/parameters/tag"
                    }
                ],
                "responses": {
                    "default": {
                        "description": "Success/error response",
                        "schema": {
                            "$ref": "#/definitions/JsonRpcIntervalsReserveResponse"
                        }
                    }
                }
            }
        },
        "/jsonrpc/v1/intervals.free": {
            "get": {
                "tags": [
                    "json-rpc",
                    "intervals"
                ],
                "summary": "Free interval",
                "operationId": "intervals.free",
                "parameters": [
                    {
                        "$ref": "#/parameters/jsonrpc_id"
                    },
                    {
                        "$ref": "#/parameters/interval_id"
                    },
                    {
                        "$ref": "#/parameters/pallet_method_code"
                    },
                    {
                        "$ref": "#/parameters/tag"
                    }
                ],
                "responses": {
                    "default": {
                        "description": "Success/error response",
                        "schema": {
                            "$ref": "#/definitions/JsonRpcOkResponse"
                        }
                    }
                }
            }
        },
        "/jsonrpc/v1/intervals.get": {
            "get": {
                "tags": [
                    "json-rpc",
                    "intervals"
                ],
                "summary": "Get interval",
                "operationId": "intervals.get",
                "parameters": [
                    {
                        "$ref": "#/parameters/jsonrpc_id"
                    },
                    {
                        "$ref": "#/parameters/interval_id"
                    },
                    {
                        "$ref": "#/parameters/tag"
                    }
                ],
                "responses": {
                    "default": {
                        "description": "Success/error response",
                        "schema": {
                            "$ref": "#/definitions/JsonRpcIntervalsGetResponse"
                        }
                    }
                }
            }
        },
        "/jsonrpc/v1/intervals.find": {
            "get": {
                "tags": [
                    "json-rpc",
                    "intervals"
                ],
                "summary": "Find interval within the same zone",
                "description": "Required either current_id or pair zone_id + service_level_type_code.",
                "operationId": "intervals.find",
                "parameters": [
                    {
                        "$ref": "#/parameters/jsonrpc_id"
                    },
                    {
                        "$ref": "#/parameters/desired_start"
                    },
                    {
                        "$ref": "#/parameters/desired_end"
                    },
                    {
                        "$ref": "#/parameters/current_id"
                    },
                    {
                        "$ref": "#/parameters/zone_id"
                    },
                    {
                        "$ref": "#/parameters/service_level_type_code"
                    }
                ],
                "responses": {
                    "default": {
                        "description": "Success/error response",
                        "schema": {
                            "$ref": "#/definitions/JsonRpcIntervalsFindResponse"
                        }
                    }
                }
            }
        },
        "/jsonrpc/v1/methods.get": {
            "get": {
                "tags": [
                    "json-rpc",
                    "methods"
                ],
                "summary": "Get methods",
                "operationId": "methods.get",
                "parameters": [
                    {
                        "$ref": "#/parameters/jsonrpc_id"
                    },
                    {
                        "$ref": "#/parameters/zipcode"
                    },
                    {
                        "$ref": "#/parameters/aoid"
                    },
                    {
                        "$ref": "#/parameters/latitude"
                    },
                    {
                        "$ref": "#/parameters/longitude"
                    },
                    {
                        "$ref": "#/parameters/item_count"
                    },
                    {
                        "$ref": "#/parameters/cart_amount"
                    },
                    {
                        "$ref": "#/parameters/payment_type"
                    },
                    {
                        "$ref": "#/parameters/no_airfreight"
                    },
                    {
                        "$ref": "#/parameters/liquids"
                    },
                    {
                        "$ref": "#/parameters/checkout"
                    },
                    {
                        "$ref": "#/parameters/method_code"
                    },
                    {
                        "$ref": "#/parameters/company"
                    },
                    {
                        "$ref": "#/parameters/origin_date"
                    },
                    {
                        "$ref": "#/parameters/profile_id"
                    },
                    {
                        "$ref": "#/parameters/ignore_capacity"
                    },
                    {
                        "$ref": "#/parameters/ignore_cutoff"
                    },
                    {
                        "$ref": "#/parameters/reserved_interval_id"
                    },
                    {
                        "$ref": "#/parameters/original_delivery_date"
                    },
                    {
                        "$ref": "#/parameters/tag"
                    }
                ],
                "responses": {
                    "default": {
                        "description": "Success/error response",
                        "schema": {
                            "$ref": "#/definitions/JsonRpcMethodsGetResponse"
                        }
                    }
                }
            }
        },
        "/jsonrpc/v1/methods.get-checkout": {
            "get": {
                "tags": [
                    "json-rpc",
                    "methods"
                ],
                "summary": "Get checkout methods",
                "operationId": "methods.getCheckout",
                "parameters": [
                    {
                        "$ref": "#/parameters/jsonrpc_id"
                    },
                    {
                        "$ref": "#/parameters/zipcode"
                    },
                    {
                        "$ref": "#/parameters/aoid"
                    },
                    {
                        "$ref": "#/parameters/zone_id"
                    },
                    {
                        "$ref": "#/parameters/latitude"
                    },
                    {
                        "$ref": "#/parameters/longitude"
                    },
                    {
                        "$ref": "#/parameters/item_count"
                    },
                    {
                        "$ref": "#/parameters/cart_amount"
                    },
                    {
                        "$ref": "#/parameters/payment_type"
                    },
                    {
                        "$ref": "#/parameters/no_airfreight"
                    },
                    {
                        "$ref": "#/parameters/liquids"
                    },
                    {
                        "$ref": "#/parameters/checkout"
                    },
                    {
                        "$ref": "#/parameters/method_code"
                    },
                    {
                        "$ref": "#/parameters/company"
                    },
                    {
                        "$ref": "#/parameters/origin_date"
                    },
                    {
                        "$ref": "#/parameters/profile_id"
                    },
                    {
                        "$ref": "#/parameters/ignore_capacity"
                    },
                    {
                        "$ref": "#/parameters/ignore_cutoff"
                    },
                    {
                        "$ref": "#/parameters/reserved_interval_id"
                    },
                    {
                        "$ref": "#/parameters/original_delivery_date"
                    },
                    {
                        "$ref": "#/parameters/tag"
                    }
                ],
                "responses": {
                    "default": {
                        "description": "Success/error response",
                        "schema": {
                            "$ref": "#/definitions/JsonRpcMethodsGetCheckoutResponse"
                        }
                    }
                }
            }
        },
        "/jsonrpc/v1/methods.get-checkout-multiple": {
            "get": {
                "tags": [
                    "json-rpc",
                    "methods"
                ],
                "summary": "Get checkout methods for multiple orders",
                "description": "Two-dimensional orders array defines parameters for each individual order.\n    At least one order have to be specified with one or more parameters.",
                "operationId": "methods.getCheckoutMultiple",
                "parameters": [
                    {
                        "$ref": "#/parameters/jsonrpc_id"
                    },
                    {
                        "$ref": "#/parameters/zipcode"
                    },
                    {
                        "$ref": "#/parameters/aoid"
                    },
                    {
                        "$ref": "#/parameters/zone_id"
                    },
                    {
                        "$ref": "#/parameters/latitude"
                    },
                    {
                        "$ref": "#/parameters/longitude"
                    },
                    {
                        "$ref": "#/parameters/checkout"
                    },
                    {
                        "$ref": "#/parameters/method_code"
                    },
                    {
                        "$ref": "#/parameters/origin_date"
                    },
                    {
                        "$ref": "#/parameters/ignore_capacity"
                    },
                    {
                        "$ref": "#/parameters/ignore_cutoff"
                    },
                    {
                        "$ref": "#/parameters/orders_profile_id"
                    },
                    {
                        "$ref": "#/parameters/orders_item_count"
                    },
                    {
                        "$ref": "#/parameters/orders_cart_amount"
                    },
                    {
                        "$ref": "#/parameters/orders_no_airfreight"
                    },
                    {
                        "$ref": "#/parameters/orders_payment_type"
                    },
                    {
                        "$ref": "#/parameters/tag"
                    }
                ],
                "responses": {
                    "default": {
                        "description": "Success/error response.\n    <br />\n    Return list of orders in the same sequence as in request.",
                        "schema": {
                            "$ref": "#/definitions/JsonRpcMethodsGetCheckoutMultipleResponse"
                        }
                    }
                }
            }
        },
        "/jsonrpc/v1/methods.get-details": {
            "get": {
                "tags": [
                    "json-rpc",
                    "methods"
                ],
                "summary": "Get method details",
                "operationId": "methods.getDetails",
                "parameters": [
                    {
                        "$ref": "#/parameters/jsonrpc_id"
                    },
                    {
                        "$ref": "#/parameters/code"
                    },
                    {
                        "$ref": "#/parameters/join"
                    },
                    {
                        "$ref": "#/parameters/zone_id"
                    },
                    {
                        "$ref": "#/parameters/service_level"
                    },
                    {
                        "$ref": "#/parameters/zipcode"
                    },
                    {
                        "$ref": "#/parameters/aoid"
                    },
                    {
                        "$ref": "#/parameters/origin_date"
                    }
                ],
                "responses": {
                    "default": {
                        "description": "Success/error response",
                        "schema": {
                            "$ref": "#/definitions/JsonRpcMethodsGetDetailsResponse"
                        }
                    }
                }
            }
        },
        "/jsonrpc/v1/methods.get-details-multiple": {
            "get": {
                "tags": [
                    "json-rpc",
                    "methods"
                ],
                "summary": "Get method details multiple",
                "operationId": "methods.getDetailsMultiple",
                "parameters": [
                    {
                        "$ref": "#/parameters/jsonrpc_id"
                    },
                    {
                        "$ref": "#/parameters/get_method_details_multiple_code"
                    },
                    {
                        "$ref": "#/parameters/get_method_details_multiple_join"
                    },
                    {
                        "$ref": "#/parameters/get_method_details_multiple_zone_id"
                    },
                    {
                        "$ref": "#/parameters/get_method_details_multiple_service_level"
                    },
                    {
                        "$ref": "#/parameters/get_method_details_multiple_zipcode"
                    },
                    {
                        "$ref": "#/parameters/get_method_details_multiple_aoid"
                    },
                    {
                        "$ref": "#/parameters/get_method_details_multiple_origin_date"
                    }
                ],
                "responses": {
                    "default": {
                        "description": "Success/error response.\n    <br />\n    Return list of methods in the same sequence as in request.",
                        "schema": {
                            "$ref": "#/definitions/JsonRpcMethodsGetDetailsMultipleResponse"
                        }
                    }
                }
            }
        },
        "/jsonrpc/v1/methods.get-all": {
            "get": {
                "tags": [
                    "json-rpc",
                    "methods"
                ],
                "summary": "Get all methods",
                "operationId": "getAll",
                "parameters": [
                    {
                        "$ref": "#/parameters/jsonrpc_id"
                    }
                ],
                "responses": {
                    "default": {
                        "description": "Success/error response",
                        "schema": {
                            "$ref": "#/definitions/JsonRpcMethodsGetAllResponse"
                        }
                    }
                }
            }
        },
        "/jsonrpc/v1/pickups.get-list": {
            "get": {
                "tags": [
                    "json-rpc",
                    "pickups"
                ],
                "summary": "Get pup list",
                "operationId": "pickups.getList",
                "parameters": [
                    {
                        "$ref": "#/parameters/jsonrpc_id"
                    },
                    {
                        "$ref": "#/parameters/zipcode"
                    },
                    {
                        "$ref": "#/parameters/aoid"
                    },
                    {
                        "$ref": "#/parameters/latitude"
                    },
                    {
                        "$ref": "#/parameters/longitude"
                    },
                    {
                        "$ref": "#/parameters/item_count"
                    },
                    {
                        "$ref": "#/parameters/cart_amount"
                    },
                    {
                        "$ref": "#/parameters/payment_type"
                    },
                    {
                        "$ref": "#/parameters/no_airfreight"
                    },
                    {
                        "$ref": "#/parameters/liquids"
                    },
                    {
                        "$ref": "#/parameters/checkout"
                    },
                    {
                        "$ref": "#/parameters/method_code"
                    },
                    {
                        "$ref": "#/parameters/company"
                    },
                    {
                        "$ref": "#/parameters/origin_date"
                    },
                    {
                        "$ref": "#/parameters/profile_id"
                    },
                    {
                        "$ref": "#/parameters/ignore_capacity"
                    },
                    {
                        "$ref": "#/parameters/ignore_cutoff"
                    },
                    {
                        "$ref": "#/parameters/reserved_interval_id"
                    },
                    {
                        "$ref": "#/parameters/original_delivery_date"
                    },
                    {
                        "$ref": "#/parameters/tag"
                    }
                ],
                "responses": {
                    "default": {
                        "description": "Success/error response",
                        "schema": {
                            "$ref": "#/definitions/JsonRpcPickupsGetListResponse"
                        }
                    }
                }
            }
        },
        "/jsonrpc/v1/pickups.get-details": {
            "get": {
                "tags": [
                    "json-rpc",
                    "pickups"
                ],
                "summary": "Get pup details",
                "operationId": "pickups.getDetails",
                "parameters": [
                    {
                        "$ref": "#/parameters/jsonrpc_id"
                    },
                    {
                        "$ref": "#/parameters/pickup_id"
                    },
                    {
                        "$ref": "#/parameters/join"
                    },
                    {
                        "$ref": "#/parameters/zipcode"
                    },
                    {
                        "$ref": "#/parameters/aoid"
                    },
                    {
                        "$ref": "#/parameters/latitude"
                    },
                    {
                        "$ref": "#/parameters/longitude"
                    },
                    {
                        "$ref": "#/parameters/item_count"
                    },
                    {
                        "$ref": "#/parameters/cart_amount"
                    },
                    {
                        "$ref": "#/parameters/payment_type"
                    },
                    {
                        "$ref": "#/parameters/no_airfreight"
                    },
                    {
                        "$ref": "#/parameters/liquids"
                    },
                    {
                        "$ref": "#/parameters/checkout"
                    },
                    {
                        "$ref": "#/parameters/method_code"
                    },
                    {
                        "$ref": "#/parameters/company"
                    },
                    {
                        "$ref": "#/parameters/origin_date"
                    },
                    {
                        "$ref": "#/parameters/profile_id"
                    },
                    {
                        "$ref": "#/parameters/ignore_capacity"
                    },
                    {
                        "$ref": "#/parameters/ignore_cutoff"
                    },
                    {
                        "$ref": "#/parameters/reserved_interval_id"
                    },
                    {
                        "$ref": "#/parameters/original_delivery_date"
                    },
                    {
                        "$ref": "#/parameters/tag"
                    }
                ],
                "responses": {
                    "default": {
                        "description": "Success/error response",
                        "schema": {
                            "$ref": "#/definitions/JsonRpcPickupsGetDetailsResponse"
                        }
                    }
                }
            }
        },
        "/jsonrpc/v1/pickups.get": {
            "get": {
                "tags": [
                    "json-rpc",
                    "pickups"
                ],
                "summary": "Get pickup",
                "operationId": "pickups.get",
                "parameters": [
                    {
                        "$ref": "#/parameters/jsonrpc_id"
                    },
                    {
                        "$ref": "#/parameters/pickup_id"
                    },
                    {
                        "$ref": "#/parameters/join"
                    },
                    {
                        "$ref": "#/parameters/zipcode"
                    },
                    {
                        "$ref": "#/parameters/aoid"
                    }
                ],
                "responses": {
                    "default": {
                        "description": "Success/error response",
                        "schema": {
                            "$ref": "#/definitions/JsonRpcPickupsGetResponse"
                        }
                    }
                }
            }
        },
        "/jsonrpc/v1/pickups.create": {
            "post": {
                "tags": [
                    "json-rpc",
                    "pickups"
                ],
                "summary": "Create pickups",
                "operationId": "pickups.create",
                "parameters": [
                    {
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/JsonRpcPickupsCreateRequest"
                        }
                    }
                ],
                "responses": {
                    "default": {
                        "description": "Success/error response",
                        "schema": {
                            "$ref": "#/definitions/JsonRpcPickupsCreateResponse"
                        }
                    }
                }
            }
        },
        "/jsonrpc/v1/profiles.import": {
            "post": {
                "tags": [
                    "profiles",
                    "json-rpc"
                ],
                "summary": "Import profiles",
                "operationId": "profiles.import",
                "parameters": [
                    {
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/JsonRpcProfilesImportRequest"
                        }
                    }
                ],
                "responses": {
                    "default": {
                        "description": "Success/error response",
                        "schema": {
                            "$ref": "#/definitions/JsonRpcOkResponse"
                        }
                    }
                }
            }
        },
        "/jsonrpc/v1/zones.get-rate": {
            "get": {
                "tags": [
                    "json-rpc",
                    "zones"
                ],
                "summary": "Get zone rate",
                "description": "Either zipcode or zone_id parameter is required.\n    If zone_id is specified then it is used instead of zipcode to find required zone.",
                "operationId": "zones.getRate",
                "parameters": [
                    {
                        "$ref": "#/parameters/jsonrpc_id"
                    },
                    {
                        "$ref": "#/parameters/method_code_required"
                    },
                    {
                        "$ref": "#/parameters/zipcode"
                    },
                    {
                        "$ref": "#/parameters/zone_id"
                    },
                    {
                        "$ref": "#/parameters/weight"
                    },
                    {
                        "$ref": "#/parameters/insurance_sum"
                    }
                ],
                "responses": {
                    "default": {
                        "description": "Success/error response",
                        "schema": {
                            "$ref": "#/definitions/JsonRpcZonesGetRateResponse"
                        }
                    }
                }
            }
        },
        "/jsonrpc/v1/zones.update-rate": {
            "post": {
                "tags": [
                    "zones",
                    "json-rpc"
                ],
                "summary": "Update zone rate",
                "description": "Zone should have at least one search criteria: zone_id|pickup_id|pickup_external_id",
                "operationId": "zones.updateRate",
                "parameters": [
                    {
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/JsonRpcZonesUpdateRateRequest"
                        }
                    }
                ],
                "responses": {
                    "default": {
                        "description": "Success/error response",
                        "schema": {
                            "$ref": "#/definitions/JsonRpcZonesUpdateRateResponse"
                        }
                    }
                }
            }
        },
        "/jsonrpc/v1/zones.get-attributes": {
            "get": {
                "tags": [
                    "json-rpc",
                    "zones"
                ],
                "summary": "Get zone attributes",
                "operationId": "zones.getAttributes",
                "parameters": [
                    {
                        "$ref": "#/parameters/jsonrpc_id"
                    },
                    {
                        "$ref": "#/parameters/zone_id_required"
                    }
                ],
                "responses": {
                    "default": {
                        "description": "Success/error response",
                        "schema": {
                            "$ref": "#/definitions/JsonRpcZonesGetAttributesResponse"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "Address": {
            "required": [
                "pickup_id",
                "city",
                "street",
                "house"
            ],
            "properties": {
                "pickup_id": {
                    "type": "integer",
                    "example": 16213
                },
                "latitude": {
                    "type": "number",
                    "format": "float",
                    "example": 55.738176
                },
                "longitude": {
                    "type": "number",
                    "format": "float",
                    "example": 37.631595
                },
                "city": {
                    "type": "string",
                    "example": "\u041c\u043e\u0441\u043a\u0432\u0430"
                },
                "street": {
                    "type": "string",
                    "example": "\u041d\u043e\u0432\u043e\u043a\u0443\u0437\u043d\u0435\u0446\u043a\u0430\u044f"
                },
                "house": {
                    "type": "string",
                    "example": "11/13"
                },
                "office": {
                    "type": "string"
                },
                "underground_station": {
                    "type": "string",
                    "example": "\u043c. \u041d\u043e\u0432\u043e\u043a\u0443\u0437\u043d\u0435\u0446\u043a\u0430\u044f/\u0422\u0440\u0435\u0442\u044c\u044f\u043a\u043e\u0432\u0441\u043a\u0430\u044f/\u041f\u0430\u0432\u0435\u043b\u0435\u0446\u043a\u0430\u044f"
                }
            },
            "type": "object"
        },
        "ArrayItem": {
            "required": [
                "key",
                "value"
            ],
            "properties": {
                "key": {
                    "type": "string",
                    "example": "is_call_needed"
                },
                "value": {
                    "example": "1"
                }
            },
            "type": "object"
        },
        "CheckoutMethod": {
            "required": [
                "code",
                "name",
                "category_name",
                "has_horizon",
                "has_intervals",
                "is_client_name_required",
                "delivery_price",
                "checkout_name",
                "days",
                "payment_type"
            ],
            "properties": {
                "code": {
                    "type": "string",
                    "example": "lamoda_showroom_novokuznetskaya"
                },
                "name": {
                    "type": "string",
                    "example": "\u041d\u043e\u0432\u043e\u043a\u0443\u0437\u043d\u0435\u0446\u043a\u0430\u044f, \u043c. \u041d\u043e\u0432\u043e\u043a\u0443\u0437\u043d\u0435\u0446\u043a\u0430\u044f, \u043c.\u0422\u0440\u0435\u0442\u044c\u044f\u043a\u043e\u0432\u0441\u043a\u0430\u044f, \u043c. \u041f\u0430\u0432\u0435\u043b\u0435\u0446\u043a\u0430\u044f"
                },
                "category_name": {
                    "type": "string",
                    "example": "\u0421\u0430\u043c\u043e\u0432\u044b\u0432\u043e\u0437"
                },
                "checkout_description": {
                    "type": "string",
                    "example": "\u0414\u043e\u0441\u0442\u0430\u0432\u043a\u0430 \u0437\u0430\u043a\u0430\u0437\u0430 \u0432 \u043f\u0443\u043d\u043a\u0442 \u0432\u044b\u0434\u0430\u0447\u0438 \u0437\u0430\u043a\u0430\u0437\u043e\u0432 \u0431\u0435\u0441\u043f\u043b\u0430\u0442\u043d\u0430 \u0432 \u0441\u043b\u0443\u0447\u0430\u0435 ..."
                },
                "callcenter_description": {
                    "type": "string",
                    "example": "\u041f\u0440\u0438 \u0432\u044b\u043a\u0443\u043f\u0435 \u043d\u0430 \u0441\u0443\u043c\u043c\u0443 \u0431\u043e\u043b\u0435\u0435 2500 \u0440\u0443\u0431\u043b\u0435\u0439 \u0434\u043e\u0441\u0442\u0430\u0432\u043a\u0430 \u0431\u0435\u0441\u043f\u043b\u0430\u0442\u043d\u0430\u044f ..."
                },
                "has_horizon": {
                    "type": "boolean"
                },
                "has_intervals": {
                    "type": "boolean"
                },
                "is_client_name_required": {
                    "type": "boolean",
                    "example": false
                },
                "delivery_price": {
                    "type": "number",
                    "format": "float",
                    "example": 250
                },
                "free_delivery_net_threshold": {
                    "type": "number",
                    "format": "float",
                    "example": 2500
                },
                "checkout_name": {
                    "type": "string",
                    "example": "\u0414\u043e\u0441\u0442\u0430\u0432\u043a\u0430 \u0432 \u043f\u0443\u043d\u043a\u0442 \u0432\u044b\u0434\u0430\u0447\u0438 \u0437\u0430\u043a\u0430\u0437\u043e\u0432 Lamoda"
                },
                "delivery_date_min": {
                    "type": "string",
                    "example": "2017-03-04"
                },
                "delivery_date_max": {
                    "type": "string",
                    "example": "2017-03-08"
                },
                "days": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/Day"
                    }
                },
                "address": {
                    "$ref": "#/definitions/Address"
                },
                "payment_type": {
                    "description": "0 - all types, 1 - prepayment, 2 - postpayment",
                    "type": "integer",
                    "example": 0
                },
                "customs_threshold": {
                    "type": "number",
                    "format": "float"
                },
                "customs_threshold_description": {
                    "type": "string"
                },
                "communication_description": {
                    "type": "string",
                    "example": "{\u041a\u043e\u0433\u0434\u0430 \u0437\u0430\u043a\u0430\u0437 \u043f\u043e\u0441\u0442\u0443\u043f\u0438\u0442 \u0432 \u043f\u0443\u043d\u043a\u0442 \u0441\u0430\u043c\u043e\u0432\u044b\u0432\u043e\u0437\u0430, \u0432\u0430\u043c \u043f\u0440\u0438\u0434\u0435\u0442 SMS..."
                },
                "is_own": {
                    "type": "boolean",
                    "example": false
                }
            },
            "type": "object"
        },
        "Phone": {
            "required": [
                "code",
                "number"
            ],
            "properties": {
                "code": {
                    "type": "string",
                    "example": "+7"
                },
                "number": {
                    "type": "string",
                    "example": "(495) 363-63-93"
                }
            },
            "type": "object"
        },
        "CreatePickups\\Pickup": {
            "required": [
                "external_id",
                "name",
                "city",
                "street",
                "house"
            ],
            "properties": {
                "external_id": {
                    "type": "string"
                },
                "name": {
                    "type": "string",
                    "example": "\u041d\u043e\u0432\u043e\u043a\u0443\u0437\u043d\u0435\u0446\u043a\u0430\u044f"
                },
                "zipcode": {
                    "type": "string"
                },
                "region": {
                    "type": "string"
                },
                "city": {
                    "type": "string",
                    "example": "\u041c\u043e\u0441\u043a\u0432\u0430"
                },
                "street": {
                    "type": "string",
                    "example": "\u041d\u043e\u0432\u043e\u043a\u0443\u0437\u043d\u0435\u0446\u043a\u0430\u044f"
                },
                "house": {
                    "type": "string",
                    "example": "11/13"
                },
                "latitude": {
                    "type": "number",
                    "format": "float",
                    "example": 55.738176
                },
                "longitude": {
                    "type": "number",
                    "format": "float",
                    "example": 37.631595
                },
                "underground_station": {
                    "type": "string",
                    "example": "\u043c. \u041d\u043e\u0432\u043e\u043a\u0443\u0437\u043d\u0435\u0446\u043a\u0430\u044f/\u0422\u0440\u0435\u0442\u044c\u044f\u043a\u043e\u0432\u0441\u043a\u0430\u044f/\u041f\u0430\u0432\u0435\u043b\u0435\u0446\u043a\u0430\u044f"
                },
                "map_description": {
                    "type": "string",
                    "example": "\u0412\u044b\u0445\u043e\u0434\u0438\u0442\u0435 \u0438\u0437 \u0441\u0442\u0430\u043d\u0446\u0438\u0438 \u043c\u0435\u0442\u0440\u043e <b>\u041d\u043e\u0432\u043e\u043a\u0443\u0437\u043d\u0435\u0446\u043a\u0430\u044f </b>\u043d\u0430 \u0443\u043b\u0438\u0446\u0443 \u041d\u043e\u0432\u043e\u043a\u0443\u0437\u043d\u0435\u0446\u043a\u0430\u044f..."
                },
                "is_bankcard_accepted": {
                    "type": "boolean",
                    "example": true
                },
                "is_inactive": {
                    "type": "boolean",
                    "example": false
                },
                "is_24hours": {
                    "type": "boolean",
                    "example": false
                },
                "work_time": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/WorkTime"
                    }
                },
                "phones": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/Phone"
                    }
                },
                "group_code": {
                    "type": "string",
                    "example": "4"
                },
                "attributes": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/ArrayItem"
                    }
                },
                "photo_url": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "photo_base64": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                }
            },
            "type": "object"
        },
        "CreatePickups\\Request": {
            "required": [
                "method_code"
            ],
            "properties": {
                "method_code": {
                    "type": "string",
                    "example": "lamoda_showroom_novokuznetskaya"
                },
                "pickups": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/CreatePickups\\Pickup"
                    }
                }
            },
            "type": "object"
        },
        "CreatePickups\\Response": {
            "properties": {
                "success": {
                    "type": "boolean"
                },
                "error": {
                    "$ref": "#/definitions/ResponseError"
                }
            },
            "type": "object"
        },
        "DayWithPalletMethod": {
            "required": [
                "date",
                "available_intervals"
            ],
            "properties": {
                "date": {
                    "type": "string",
                    "example": "2017-03-04"
                },
                "available_intervals": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/IntervalWithPalletMethod"
                    }
                },
                "is_available": {
                    "type": "boolean",
                    "example": true
                }
            },
            "type": "object"
        },
        "FindIntervalResponse": {
            "properties": {
                "date": {
                    "description": "Date, YYYY-MM-DD",
                    "type": "string",
                    "example": "2017-03-04"
                },
                "interval": {
                    "$ref": "#/definitions/Interval"
                },
                "success": {
                    "type": "boolean"
                },
                "error": {
                    "$ref": "#/definitions/ResponseError"
                }
            },
            "type": "object"
        },
        "GetAllMethodsResponse": {
            "properties": {
                "methods": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/Method"
                    }
                },
                "success": {
                    "type": "boolean"
                },
                "error": {
                    "$ref": "#/definitions/ResponseError"
                }
            },
            "type": "object"
        },
        "Day": {
            "required": [
                "date",
                "intervals"
            ],
            "properties": {
                "date": {
                    "type": "string",
                    "example": "2017-03-04"
                },
                "intervals": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/IntervalWithPalletMethod"
                    }
                },
                "is_available": {
                    "type": "boolean",
                    "example": true
                }
            },
            "type": "object"
        },
        "GroupDay": {
            "required": [
                "date",
                "intervals"
            ],
            "properties": {
                "date": {
                    "type": "string",
                    "example": "2017-03-04"
                },
                "intervals": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/GroupInterval"
                    }
                }
            },
            "type": "object"
        },
        "GroupInterval": {
            "required": [
                "start",
                "end"
            ],
            "properties": {
                "start": {
                    "type": "string",
                    "example": "2017-03-04 10:00:00"
                },
                "end": {
                    "type": "string",
                    "example": "2017-03-04 22:00:00"
                }
            },
            "type": "object"
        },
        "GroupMethod": {
            "required": [
                "method_type_code",
                "method_type_name",
                "code",
                "name",
                "category_name",
                "checkout_name",
                "days"
            ],
            "properties": {
                "method_type_code": {
                    "type": "string",
                    "example": "pickup"
                },
                "method_type_name": {
                    "type": "string",
                    "example": "\u0421\u0430\u043c\u043e\u0432\u044b\u0432\u043e\u0437"
                },
                "code": {
                    "type": "string",
                    "example": "lamoda_showroom_novokuznetskaya"
                },
                "name": {
                    "type": "string",
                    "example": "\u041d\u043e\u0432\u043e\u043a\u0443\u0437\u043d\u0435\u0446\u043a\u0430\u044f, \u043c. \u041d\u043e\u0432\u043e\u043a\u0443\u0437\u043d\u0435\u0446\u043a\u0430\u044f, \u043c.\u0422\u0440\u0435\u0442\u044c\u044f\u043a\u043e\u0432\u0441\u043a\u0430\u044f, \u043c. \u041f\u0430\u0432\u0435\u043b\u0435\u0446\u043a\u0430\u044f"
                },
                "category_name": {
                    "type": "string",
                    "example": "\u0421\u0430\u043c\u043e\u0432\u044b\u0432\u043e\u0437"
                },
                "checkout_name": {
                    "type": "string",
                    "example": "\u0414\u043e\u0441\u0442\u0430\u0432\u043a\u0430 \u0432 \u043f\u0443\u043d\u043a\u0442 \u0432\u044b\u0434\u0430\u0447\u0438 \u0437\u0430\u043a\u0430\u0437\u043e\u0432 Lamoda"
                },
                "days": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/GroupDay"
                    }
                }
            },
            "type": "object"
        },
        "GetCheckoutMethodsMultiple\\Response": {
            "properties": {
                "orders": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/ResponseOrder"
                    }
                },
                "groups": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/ResponseGroup"
                    }
                },
                "success": {
                    "type": "boolean"
                },
                "error": {
                    "$ref": "#/definitions/ResponseError"
                }
            },
            "type": "object"
        },
        "ResponseGroup": {
            "required": [
                "orders",
                "methods"
            ],
            "properties": {
                "orders": {
                    "type": "array",
                    "items": {
                        "type": "integer"
                    },
                    "example": [
                        0,
                        1
                    ]
                },
                "methods": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/GroupMethod"
                    }
                }
            },
            "type": "object"
        },
        "ResponseOrder": {
            "required": [
                "method_types"
            ],
            "properties": {
                "method_types": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/MethodType"
                    }
                }
            },
            "type": "object"
        },
        "GetCheckoutMethodsResponse": {
            "properties": {
                "method_types": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/MethodType"
                    }
                },
                "success": {
                    "type": "boolean"
                },
                "error": {
                    "$ref": "#/definitions/ResponseError"
                }
            },
            "type": "object"
        },
        "GetDeliveryInfoMethodType": {
            "required": [
                "code",
                "name",
                "is_lme",
                "service_level_types"
            ],
            "properties": {
                "code": {
                    "type": "string",
                    "example": "pickup"
                },
                "name": {
                    "type": "string",
                    "example": "\u0421\u0430\u043c\u043e\u0432\u044b\u0432\u043e\u0437"
                },
                "is_lme": {
                    "type": "boolean",
                    "example": true
                },
                "service_level_types": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/GetDeliveryInfoServiceLevelType"
                    }
                }
            },
            "type": "object"
        },
        "GetDeliveryInfoResponse": {
            "properties": {
                "method_types": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/GetDeliveryInfoMethodType"
                    }
                },
                "success": {
                    "type": "boolean"
                },
                "error": {
                    "$ref": "#/definitions/ResponseError"
                }
            },
            "type": "object"
        },
        "GetDeliveryInfoServiceLevelType": {
            "required": [
                "code",
                "name",
                "checkout_name",
                "is_tryon_allowed",
                "is_rejection_allowed",
                "is_bankcard_accepted",
                "cutoff_time"
            ],
            "properties": {
                "code": {
                    "type": "string",
                    "example": "plus"
                },
                "name": {
                    "type": "string",
                    "example": "\u041f\u043b\u044e\u0441"
                },
                "checkout_name": {
                    "type": "string",
                    "example": "\u041f\u043b\u044e\u0441"
                },
                "description": {
                    "type": "string"
                },
                "free_delivery_net_threshold_from": {
                    "type": "number",
                    "format": "float",
                    "example": "2500.00"
                },
                "free_delivery_net_threshold_to": {
                    "type": "number",
                    "format": "float",
                    "example": "2500.00"
                },
                "free_delivery_net_threshold_percent_from": {
                    "type": "integer",
                    "example": 0
                },
                "free_delivery_net_threshold_percent_to": {
                    "type": "integer",
                    "example": 0
                },
                "free_delivery_threshold_share": {
                    "type": "integer",
                    "example": 0
                },
                "free_delivery_gross_threshold_from": {
                    "type": "number",
                    "format": "float"
                },
                "free_delivery_gross_threshold_to": {
                    "type": "number",
                    "format": "float"
                },
                "delivery_price_from": {
                    "type": "number",
                    "format": "float",
                    "example": "250.00"
                },
                "delivery_price_to": {
                    "type": "number",
                    "format": "float",
                    "example": "250.00"
                },
                "is_tryon_allowed": {
                    "type": "boolean",
                    "example": true
                },
                "is_rejection_allowed": {
                    "type": "boolean",
                    "example": true
                },
                "is_bankcard_accepted": {
                    "type": "boolean",
                    "example": true
                },
                "delivery_date_from": {
                    "type": "integer",
                    "example": 1
                },
                "delivery_date_to": {
                    "type": "integer",
                    "example": 5
                },
                "storage_days_from": {
                    "type": "integer",
                    "example": 2
                },
                "storage_days_to": {
                    "type": "integer",
                    "example": 2
                },
                "tryon_limit_from": {
                    "type": "integer",
                    "example": 15
                },
                "tryon_limit_to": {
                    "type": "integer",
                    "example": 15
                },
                "cutoff_time": {
                    "type": "string",
                    "example": "2017-03-03 23:59:00"
                },
                "has_horizon": {
                    "type": "boolean",
                    "example": true
                },
                "horizon_from": {
                    "type": "integer",
                    "example": 1
                },
                "horizon_till": {
                    "type": "integer",
                    "example": 7
                },
                "day_min": {
                    "type": "integer",
                    "example": 3
                },
                "day_max": {
                    "type": "integer",
                    "example": 5
                },
                "is_business_days": {
                    "type": "boolean",
                    "example": false
                },
                "payment_type": {
                    "type": "integer",
                    "example": 0
                }
            },
            "type": "object"
        },
        "GetDeliveryInfoShortResponse": {
            "properties": {
                "terms": {
                    "$ref": "#/definitions/GetDeliveryInfoServiceLevelType"
                },
                "success": {
                    "type": "boolean"
                },
                "error": {
                    "$ref": "#/definitions/ResponseError"
                }
            },
            "type": "object"
        },
        "GetIntervalResponse": {
            "properties": {
                "date": {
                    "type": "string",
                    "example": "2017-03-04"
                },
                "interval": {
                    "$ref": "#/definitions/Interval"
                },
                "success": {
                    "type": "boolean"
                },
                "error": {
                    "$ref": "#/definitions/ResponseError"
                }
            },
            "type": "object"
        },
        "GetMethodDetails\\Company": {
            "required": [
                "code",
                "name",
                "is_deleted"
            ],
            "properties": {
                "code": {
                    "type": "string",
                    "example": "lamoda"
                },
                "name": {
                    "type": "string",
                    "example": "Lamoda Express"
                },
                "is_deleted": {
                    "type": "boolean",
                    "example": false
                }
            },
            "type": "object"
        },
        "LeadTime": {
            "required": [
                "day_min",
                "day_max",
                "is_business_days",
                "day_min_calculated",
                "day_max_calculated"
            ],
            "properties": {
                "day_min": {
                    "type": "integer",
                    "example": 0
                },
                "day_max": {
                    "type": "integer",
                    "example": 5
                },
                "is_business_days": {
                    "type": "boolean",
                    "example": false
                },
                "day_min_calculated": {
                    "type": "string",
                    "example": "2017-03-04"
                },
                "day_max_calculated": {
                    "type": "string",
                    "example": "2017-03-08"
                }
            },
            "type": "object"
        },
        "GetMethodDetails\\Method": {
            "required": [
                "code",
                "name",
                "type",
                "group_name",
                "is_active",
                "is_deleted",
                "is_client_name_required",
                "checkout_name",
                "max_postpones"
            ],
            "properties": {
                "code": {
                    "type": "string",
                    "example": "lamoda_showroom_novokuznetskaya"
                },
                "name": {
                    "type": "string",
                    "example": "\u041d\u043e\u0432\u043e\u043a\u0443\u0437\u043d\u0435\u0446\u043a\u0430\u044f, \u043c. \u041d\u043e\u0432\u043e\u043a\u0443\u0437\u043d\u0435\u0446\u043a\u0430\u044f, \u043c.\u0422\u0440\u0435\u0442\u044c\u044f\u043a\u043e\u0432\u0441\u043a\u0430\u044f, \u043c. \u041f\u0430\u0432\u0435\u043b\u0435\u0446\u043a\u0430\u044f"
                },
                "type": {
                    "description": "1 - \u043a\u0443\u0440\u044c\u0435\u0440\u0441\u043a\u0430\u044f \u0434\u043e\u0441\u0442\u0430\u0432\u043a\u0430, 2 - \u0441\u0430\u043c\u043e\u0432\u044b\u0432\u043e\u0437, 3 - \u043f\u043e\u0447\u0442\u043e\u0432\u0430\u044f \u0434\u043e\u0441\u0442\u0430\u0432\u043a\u0430",
                    "type": "integer",
                    "example": 2
                },
                "group_name": {
                    "type": "string",
                    "example": "\u0421\u0430\u043c\u043e\u0432\u044b\u0432\u043e\u0437"
                },
                "description": {
                    "type": "string",
                    "example": "- \u041f\u0440\u0438 \u0437\u0430\u043a\u0430\u0437\u0435 \u0434\u043e 24:00, \u0434\u043e\u0441\u0442\u0430\u0432\u043a\u0430 \u043d\u0430 \u0441\u043b\u0435\u0434\u0443\u044e\u0449\u0438\u0439 \u0434\u0435\u043d\u044c..."
                },
                "is_active": {
                    "type": "boolean"
                },
                "is_deleted": {
                    "type": "boolean"
                },
                "carrier": {
                    "type": "string",
                    "example": "leos-express"
                },
                "tracking_url": {
                    "type": "string"
                },
                "is_own": {
                    "type": "boolean",
                    "example": true
                },
                "is_client_name_required": {
                    "type": "boolean",
                    "example": false
                },
                "company": {
                    "$ref": "#/definitions/GetMethodDetails\\Company"
                },
                "lead_time": {
                    "$ref": "#/definitions/LeadTime"
                },
                "zones": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/GetMethodDetails\\Zone"
                    }
                },
                "checkout_name": {
                    "type": "string",
                    "example": "\u0414\u043e\u0441\u0442\u0430\u0432\u043a\u0430 \u0432 \u043f\u0443\u043d\u043a\u0442 \u0432\u044b\u0434\u0430\u0447\u0438 \u0437\u0430\u043a\u0430\u0437\u043e\u0432 Lamoda"
                },
                "communication_description": {
                    "type": "string",
                    "example": "{\u041a\u043e\u0433\u0434\u0430 \u0437\u0430\u043a\u0430\u0437 \u043f\u043e\u0441\u0442\u0443\u043f\u0438\u0442 \u0432 \u043f\u0443\u043d\u043a\u0442 \u0441\u0430\u043c\u043e\u0432\u044b\u0432\u043e\u0437\u0430, \u0432\u0430\u043c \u043f\u0440\u0438\u0434\u0435\u0442 SMS..."
                },
                "max_postpones": {
                    "type": "integer",
                    "example": 3
                }
            },
            "type": "object"
        },
        "GetMethodDetails\\Pickup": {
            "required": [
                "pickup_id",
                "city",
                "street",
                "house",
                "is_24hours",
                "is_returns_accepted",
                "storage_days"
            ],
            "properties": {
                "pickup_id": {
                    "type": "integer",
                    "example": 16213
                },
                "external_id": {
                    "type": "string"
                },
                "group_id": {
                    "type": "integer",
                    "example": 4
                },
                "city": {
                    "type": "string",
                    "example": "\u041c\u043e\u0441\u043a\u0432\u0430"
                },
                "street": {
                    "type": "string",
                    "example": "\u041d\u043e\u0432\u043e\u043a\u0443\u0437\u043d\u0435\u0446\u043a\u0430\u044f"
                },
                "house": {
                    "type": "string",
                    "example": "11/13"
                },
                "office": {
                    "type": "string"
                },
                "underground_station": {
                    "type": "string",
                    "example": "\u043c. \u041d\u043e\u0432\u043e\u043a\u0443\u0437\u043d\u0435\u0446\u043a\u0430\u044f/\u0422\u0440\u0435\u0442\u044c\u044f\u043a\u043e\u0432\u0441\u043a\u0430\u044f/\u041f\u0430\u0432\u0435\u043b\u0435\u0446\u043a\u0430\u044f"
                },
                "latitude": {
                    "type": "string",
                    "example": "55.738176"
                },
                "longitude": {
                    "type": "string",
                    "example": "37.631595"
                },
                "is_24hours": {
                    "type": "boolean",
                    "example": false
                },
                "work_time": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/WorkTime"
                    }
                },
                "is_returns_accepted": {
                    "type": "boolean",
                    "example": true
                },
                "widget_description": {
                    "type": "string",
                    "example": "\u0412\u044b\u0445\u043e\u0434\u0438\u0442\u0435 \u0438\u0437 \u0441\u0442\u0430\u043d\u0446\u0438\u0438 \u043c\u0435\u0442\u0440\u043e <b>\u041d\u043e\u0432\u043e\u043a\u0443\u0437\u043d\u0435\u0446\u043a\u0430\u044f </b>\u043d\u0430 \u0443\u043b\u0438\u0446\u0443 \u041d\u043e\u0432\u043e\u043a\u0443\u0437\u043d\u0435\u0446\u043a\u0430\u044f..."
                },
                "storage_days": {
                    "type": "integer",
                    "example": 2
                },
                "phones": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    },
                    "example": [
                        "+7(495) 363-63-93"
                    ]
                },
                "photos": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    },
                    "example": [
                        "//files.lmcdn.ru/delivery/ru/75/6978cffab68e1e0ce126fa0d42368e75.png"
                    ]
                }
            },
            "type": "object"
        },
        "GetMethodDetails\\Response": {
            "properties": {
                "method": {
                    "$ref": "#/definitions/GetMethodDetails\\Method"
                },
                "success": {
                    "type": "boolean"
                },
                "error": {
                    "$ref": "#/definitions/ResponseError"
                }
            },
            "type": "object"
        },
        "GetMethodDetails\\ServiceLevel": {
            "required": [
                "delivery_price",
                "is_prepayment_free_delivery",
                "rejection_fee",
                "is_active",
                "is_deleted",
                "is_callcenter_confirm",
                "is_autoreserve_allowed",
                "payment_type",
                "has_horizon",
                "has_intervals",
                "is_weekend_included"
            ],
            "properties": {
                "delivery_price": {
                    "type": "number",
                    "format": "float",
                    "example": 250
                },
                "is_prepayment_free_delivery": {
                    "type": "boolean",
                    "example": true
                },
                "rejection_fee": {
                    "type": "number",
                    "format": "float",
                    "example": 250
                },
                "free_delivery_threshold_net": {
                    "type": "number",
                    "format": "float",
                    "example": 2500
                },
                "free_delivery_threshold_gross": {
                    "type": "number",
                    "format": "float"
                },
                "free_delivery_threshold_share": {
                    "type": "integer"
                },
                "max_items": {
                    "type": "integer",
                    "example": 15
                },
                "checkout_description": {
                    "type": "string",
                    "example": "\u0414\u043e\u0441\u0442\u0430\u0432\u043a\u0430 \u0437\u0430\u043a\u0430\u0437\u0430 \u0432 \u043f\u0443\u043d\u043a\u0442 \u0432\u044b\u0434\u0430\u0447\u0438 \u0437\u0430\u043a\u0430\u0437\u043e\u0432 \u0431\u0435\u0441\u043f\u043b\u0430\u0442\u043d\u0430 \u0432 \u0441\u043b\u0443\u0447\u0430\u0435 ..."
                },
                "callcenter_description": {
                    "type": "string",
                    "example": "\u041f\u0440\u0438 \u0432\u044b\u043a\u0443\u043f\u0435 \u043d\u0430 \u0441\u0443\u043c\u043c\u0443 \u0431\u043e\u043b\u0435\u0435 2500 \u0440\u0443\u0431\u043b\u0435\u0439 \u0434\u043e\u0441\u0442\u0430\u0432\u043a\u0430 \u0431\u0435\u0441\u043f\u043b\u0430\u0442\u043d\u0430\u044f ..."
                },
                "is_active": {
                    "type": "boolean",
                    "example": true
                },
                "is_deleted": {
                    "type": "boolean",
                    "example": false
                },
                "is_callcenter_confirm": {
                    "type": "boolean",
                    "example": false
                },
                "is_autoreserve_allowed": {
                    "type": "boolean",
                    "example": true
                },
                "max_cart_amount": {
                    "type": "number",
                    "format": "float"
                },
                "payment_type": {
                    "description": "0 - all types, 1 - prepayment, 2 - postpayment",
                    "type": "integer",
                    "example": 0
                },
                "has_horizon": {
                    "type": "boolean",
                    "example": true
                },
                "has_intervals": {
                    "type": "boolean",
                    "example": true
                },
                "is_weekend_included": {
                    "type": "boolean",
                    "example": false
                },
                "horizon_from": {
                    "type": "integer",
                    "example": 1
                },
                "horizon_till": {
                    "type": "integer",
                    "example": 5
                },
                "service_level_type": {
                    "$ref": "#/definitions/GetMethodDetails\\ServiceLevelType"
                }
            },
            "type": "object"
        },
        "GetMethodDetails\\ServiceLevelType": {
            "required": [
                "code",
                "name",
                "is_tryon_allowed",
                "is_rejection_allowed",
                "is_active",
                "is_deleted",
                "priority"
            ],
            "properties": {
                "code": {
                    "type": "string",
                    "example": "plus"
                },
                "name": {
                    "type": "string",
                    "example": "\u041f\u043b\u044e\u0441"
                },
                "description": {
                    "type": "string"
                },
                "is_tryon_allowed": {
                    "type": "boolean",
                    "example": true
                },
                "is_rejection_allowed": {
                    "type": "boolean",
                    "example": true
                },
                "is_active": {
                    "type": "boolean",
                    "example": true
                },
                "is_deleted": {
                    "type": "boolean",
                    "example": false
                },
                "priority": {
                    "type": "integer",
                    "example": 100
                }
            },
            "type": "object"
        },
        "GetMethodDetails\\Zone": {
            "required": [
                "zone_id",
                "name",
                "is_airfreight",
                "is_no_liquids",
                "operating_area",
                "is_active",
                "is_deleted",
                "is_bankcard_accepted"
            ],
            "properties": {
                "zone_id": {
                    "type": "integer",
                    "example": 104185
                },
                "name": {
                    "type": "string",
                    "example": "\u041d\u043e\u0432\u043e\u043a\u0443\u0437\u043d\u0435\u0446\u043a\u0430\u044f"
                },
                "is_airfreight": {
                    "type": "boolean",
                    "example": false
                },
                "is_no_liquids": {
                    "type": "boolean",
                    "example": false
                },
                "operating_area": {
                    "description": "0 - area specified, 1 - entire country",
                    "type": "integer",
                    "example": 0
                },
                "is_active": {
                    "type": "boolean",
                    "example": true
                },
                "is_deleted": {
                    "type": "boolean",
                    "example": false
                },
                "is_bankcard_accepted": {
                    "type": "boolean",
                    "example": true
                },
                "pickup": {
                    "$ref": "#/definitions/GetMethodDetails\\Pickup"
                },
                "service_levels": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/GetMethodDetails\\ServiceLevel"
                    }
                }
            },
            "type": "object"
        },
        "GetMethodDetailsMultiple\\Response": {
            "properties": {
                "method": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/GetMethodDetails\\Method"
                    }
                },
                "success": {
                    "type": "boolean"
                },
                "error": {
                    "$ref": "#/definitions/ResponseError"
                }
            },
            "type": "object"
        },
        "GetMethodsResponse": {
            "properties": {
                "methods": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/MethodWithIntervals"
                    }
                },
                "success": {
                    "type": "boolean"
                },
                "error": {
                    "$ref": "#/definitions/ResponseError"
                }
            },
            "type": "object"
        },
        "GetPickupResponse": {
            "properties": {
                "pickup_point": {
                    "$ref": "#/definitions/PickupResponse"
                },
                "success": {
                    "type": "boolean"
                },
                "error": {
                    "$ref": "#/definitions/ResponseError"
                }
            },
            "type": "object"
        },
        "GetPupDetailsResponse": {
            "properties": {
                "pickup_point": {
                    "$ref": "#/definitions/PupDetailsResponse"
                },
                "success": {
                    "type": "boolean"
                },
                "error": {
                    "$ref": "#/definitions/ResponseError"
                }
            },
            "type": "object"
        },
        "GetPupListResponse": {
            "properties": {
                "pickups": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/PupListResponse"
                    }
                },
                "success": {
                    "type": "boolean"
                },
                "error": {
                    "$ref": "#/definitions/ResponseError"
                }
            },
            "type": "object"
        },
        "GetReturnMethodsResponse": {
            "properties": {
                "methods": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/ReturnMethod"
                    }
                },
                "success": {
                    "type": "boolean"
                },
                "error": {
                    "$ref": "#/definitions/ResponseError"
                }
            },
            "type": "object"
        },
        "GetZoneAttributesResponse": {
            "properties": {
                "attributes": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/ArrayItem"
                    }
                },
                "success": {
                    "type": "boolean"
                },
                "error": {
                    "$ref": "#/definitions/ResponseError"
                }
            },
            "type": "object"
        },
        "GetZoneRateResponse": {
            "properties": {
                "price": {
                    "description": "Total",
                    "type": "number",
                    "format": "float",
                    "example": 189
                },
                "price_tax": {
                    "type": "number",
                    "format": "float",
                    "example": 28.83
                },
                "mass_rate": {
                    "type": "number",
                    "format": "float",
                    "example": 189
                },
                "mass_rate_tax": {
                    "type": "number",
                    "format": "float",
                    "example": 28.83
                },
                "air_rate": {
                    "type": "number",
                    "format": "float",
                    "example": 0
                },
                "air_rate_tax": {
                    "type": "number",
                    "format": "float",
                    "example": 0
                },
                "insurance_rate": {
                    "type": "number",
                    "format": "float",
                    "example": 0
                },
                "insurance_rate_tax": {
                    "type": "number",
                    "format": "float",
                    "example": 0
                },
                "transportation_type": {
                    "type": "string",
                    "example": "ground"
                },
                "parcel_online": {
                    "type": "boolean",
                    "example": true
                },
                "success": {
                    "type": "boolean"
                },
                "error": {
                    "$ref": "#/definitions/ResponseError"
                }
            },
            "type": "object"
        },
        "Interval": {
            "required": [
                "id",
                "start",
                "end"
            ],
            "properties": {
                "id": {
                    "type": "integer",
                    "example": 40704062
                },
                "start": {
                    "type": "string",
                    "example": "2017-03-04 10:00:00"
                },
                "end": {
                    "type": "string",
                    "example": "2017-03-04 22:00:00"
                }
            },
            "type": "object"
        },
        "IntervalWithPalletMethod": {
            "required": [
                "pallet_method_code"
            ],
            "properties": {
                "pallet_method_code": {
                    "type": "string",
                    "example": "lamoda_showroom_novokuznetskaya"
                },
                "is_available": {
                    "type": "boolean",
                    "example": true
                },
                "id": {
                    "type": "integer",
                    "example": 40704062
                },
                "start": {
                    "type": "string",
                    "example": "2017-03-04 10:00:00"
                },
                "end": {
                    "type": "string",
                    "example": "2017-03-04 22:00:00"
                }
            },
            "type": "object"
        },
        "Method": {
            "required": [
                "code",
                "name",
                "type",
                "group_name",
                "checkout_name"
            ],
            "properties": {
                "code": {
                    "type": "string",
                    "example": "lamoda_showroom_novokuznetskaya"
                },
                "name": {
                    "type": "string",
                    "example": "\u041d\u043e\u0432\u043e\u043a\u0443\u0437\u043d\u0435\u0446\u043a\u0430\u044f, \u043c. \u041d\u043e\u0432\u043e\u043a\u0443\u0437\u043d\u0435\u0446\u043a\u0430\u044f, \u043c.\u0422\u0440\u0435\u0442\u044c\u044f\u043a\u043e\u0432\u0441\u043a\u0430\u044f, \u043c. \u041f\u0430\u0432\u0435\u043b\u0435\u0446\u043a\u0430\u044f"
                },
                "type": {
                    "description": "1 - \u043a\u0443\u0440\u044c\u0435\u0440\u0441\u043a\u0430\u044f \u0434\u043e\u0441\u0442\u0430\u0432\u043a\u0430, 2 - \u0441\u0430\u043c\u043e\u0432\u044b\u0432\u043e\u0437, 3 - \u043f\u043e\u0447\u0442\u043e\u0432\u0430\u044f \u0434\u043e\u0441\u0442\u0430\u0432\u043a\u0430",
                    "type": "integer",
                    "example": 2
                },
                "description": {
                    "type": "string",
                    "example": "- \u041f\u0440\u0438 \u0437\u0430\u043a\u0430\u0437\u0435 \u0434\u043e 24:00, \u0434\u043e\u0441\u0442\u0430\u0432\u043a\u0430 \u043d\u0430 \u0441\u043b\u0435\u0434\u0443\u044e\u0449\u0438\u0439 \u0434\u0435\u043d\u044c..."
                },
                "is_active": {
                    "type": "boolean"
                },
                "is_deleted": {
                    "type": "boolean"
                },
                "group_name": {
                    "type": "string",
                    "example": "\u0421\u0430\u043c\u043e\u0432\u044b\u0432\u043e\u0437"
                },
                "carrier": {
                    "type": "string",
                    "example": "leos-express"
                },
                "is_own": {
                    "type": "boolean",
                    "example": true
                },
                "is_client_name_required": {
                    "type": "boolean",
                    "example": false
                },
                "checkout_name": {
                    "type": "string",
                    "example": "\u0414\u043e\u0441\u0442\u0430\u0432\u043a\u0430 \u0432 \u043f\u0443\u043d\u043a\u0442 \u0432\u044b\u0434\u0430\u0447\u0438 \u0437\u0430\u043a\u0430\u0437\u043e\u0432 Lamoda"
                },
                "max_postpones": {
                    "description": "available in get-all-methods only",
                    "type": "integer",
                    "example": 3
                }
            },
            "type": "object"
        },
        "MethodType": {
            "required": [
                "code",
                "name",
                "service_level_types"
            ],
            "properties": {
                "code": {
                    "type": "string",
                    "example": "pickup"
                },
                "name": {
                    "type": "string",
                    "example": "\u0421\u0430\u043c\u043e\u0432\u044b\u0432\u043e\u0437"
                },
                "service_level_types": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/ServiceLevelType"
                    }
                }
            },
            "type": "object"
        },
        "MethodWithIntervals": {
            "required": [
                "checkout_description",
                "is_callcenter_confirm",
                "is_autoreserve_allowed",
                "has_horizon",
                "has_intervals",
                "delivery_price",
                "zone_id",
                "available_days",
                "is_bankcard_accepted",
                "is_tryon_allowed",
                "is_rejection_allowed",
                "rejection_fee",
                "service_level_code",
                "service_level_name",
                "service_level_description"
            ],
            "properties": {
                "checkout_description": {
                    "type": "string",
                    "example": "\u0414\u043e\u0441\u0442\u0430\u0432\u043a\u0430 \u0437\u0430\u043a\u0430\u0437\u0430 \u0432 \u043f\u0443\u043d\u043a\u0442 \u0432\u044b\u0434\u0430\u0447\u0438 \u0437\u0430\u043a\u0430\u0437\u043e\u0432 \u0431\u0435\u0441\u043f\u043b\u0430\u0442\u043d\u0430 \u0432 \u0441\u043b\u0443\u0447\u0430\u0435 ..."
                },
                "is_callcenter_confirm": {
                    "type": "boolean",
                    "example": false
                },
                "is_autoreserve_allowed": {
                    "description": "If True, the chosen interval gets to be reserved right after the order is created",
                    "type": "boolean"
                },
                "has_horizon": {
                    "type": "boolean"
                },
                "has_intervals": {
                    "type": "boolean"
                },
                "delivery_price": {
                    "type": "number",
                    "format": "float",
                    "example": 250
                },
                "free_delivery_net_threshold": {
                    "type": "number",
                    "format": "float",
                    "example": 2500
                },
                "zone_id": {
                    "type": "string",
                    "example": "104185"
                },
                "macrozone_code": {
                    "type": "string",
                    "example": "e1"
                },
                "pickup_point": {
                    "$ref": "#/definitions/PupResponse"
                },
                "available_days": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/DayWithPalletMethod"
                    }
                },
                "is_bankcard_accepted": {
                    "type": "boolean"
                },
                "is_tryon_allowed": {
                    "type": "boolean"
                },
                "is_rejection_allowed": {
                    "type": "boolean"
                },
                "rejection_fee": {
                    "type": "number",
                    "format": "float",
                    "example": 250
                },
                "delivery_date_min": {
                    "type": "string",
                    "example": "2017-03-04"
                },
                "delivery_date_max": {
                    "type": "string",
                    "example": "2017-03-08"
                },
                "service_level_code": {
                    "type": "string",
                    "example": "plus"
                },
                "service_level_name": {
                    "type": "string",
                    "example": "\u041f\u043b\u044e\u0441"
                },
                "service_level_description": {
                    "type": "string",
                    "example": "\u0414\u043e\u0441\u0442\u0430\u0432\u043a\u0430 \u0431\u0435\u0441\u043f\u043b\u0430\u0442\u043d\u0430\u044f. \u041f\u0440\u0438\u043c\u0435\u0440\u043a\u0430 \u043f\u0435\u0440\u0435\u0434 \u043f\u043e\u043a\u0443\u043f\u043a\u043e\u0439, \u0432\u043e\u0437\u043c\u043e\u0436\u043d\u043e\u0441\u0442\u044c \u0447\u0430\u0441\u0442\u0438\u0447\u043d\u043e\u0433\u043e \u0432\u044b\u043a\u0443\u043f\u0430..."
                },
                "communication_description": {
                    "type": "string",
                    "example": "{\u041a\u043e\u0433\u0434\u0430 \u0437\u0430\u043a\u0430\u0437 \u043f\u043e\u0441\u0442\u0443\u043f\u0438\u0442 \u0432 \u043f\u0443\u043d\u043a\u0442 \u0441\u0430\u043c\u043e\u0432\u044b\u0432\u043e\u0437\u0430, \u0432\u0430\u043c \u043f\u0440\u0438\u0434\u0435\u0442 SMS..."
                },
                "code": {
                    "type": "string",
                    "example": "lamoda_showroom_novokuznetskaya"
                },
                "name": {
                    "type": "string",
                    "example": "\u041d\u043e\u0432\u043e\u043a\u0443\u0437\u043d\u0435\u0446\u043a\u0430\u044f, \u043c. \u041d\u043e\u0432\u043e\u043a\u0443\u0437\u043d\u0435\u0446\u043a\u0430\u044f, \u043c.\u0422\u0440\u0435\u0442\u044c\u044f\u043a\u043e\u0432\u0441\u043a\u0430\u044f, \u043c. \u041f\u0430\u0432\u0435\u043b\u0435\u0446\u043a\u0430\u044f"
                },
                "type": {
                    "description": "1 - \u043a\u0443\u0440\u044c\u0435\u0440\u0441\u043a\u0430\u044f \u0434\u043e\u0441\u0442\u0430\u0432\u043a\u0430, 2 - \u0441\u0430\u043c\u043e\u0432\u044b\u0432\u043e\u0437, 3 - \u043f\u043e\u0447\u0442\u043e\u0432\u0430\u044f \u0434\u043e\u0441\u0442\u0430\u0432\u043a\u0430",
                    "type": "integer",
                    "example": 2
                },
                "description": {
                    "type": "string",
                    "example": "- \u041f\u0440\u0438 \u0437\u0430\u043a\u0430\u0437\u0435 \u0434\u043e 24:00, \u0434\u043e\u0441\u0442\u0430\u0432\u043a\u0430 \u043d\u0430 \u0441\u043b\u0435\u0434\u0443\u044e\u0449\u0438\u0439 \u0434\u0435\u043d\u044c..."
                },
                "is_active": {
                    "type": "boolean"
                },
                "is_deleted": {
                    "type": "boolean"
                },
                "group_name": {
                    "type": "string",
                    "example": "\u0421\u0430\u043c\u043e\u0432\u044b\u0432\u043e\u0437"
                },
                "carrier": {
                    "type": "string",
                    "example": "leos-express"
                },
                "is_own": {
                    "type": "boolean",
                    "example": true
                },
                "is_client_name_required": {
                    "type": "boolean",
                    "example": false
                },
                "checkout_name": {
                    "type": "string",
                    "example": "\u0414\u043e\u0441\u0442\u0430\u0432\u043a\u0430 \u0432 \u043f\u0443\u043d\u043a\u0442 \u0432\u044b\u0434\u0430\u0447\u0438 \u0437\u0430\u043a\u0430\u0437\u043e\u0432 Lamoda"
                },
                "max_postpones": {
                    "description": "available in get-all-methods only",
                    "type": "integer",
                    "example": 3
                }
            },
            "type": "object"
        },
        "PickupResponse": {
            "required": [
                "storage_days",
                "is_returns_accepted",
                "method_code",
                "method_name",
                "method_type",
                "is_bankcard_accepted",
                "is_24hours",
                "checkout_name"
            ],
            "properties": {
                "work_time": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/WorkTime"
                    }
                },
                "storage_days": {
                    "type": "string",
                    "example": 2
                },
                "is_returns_accepted": {
                    "type": "boolean",
                    "example": true
                },
                "widget_description": {
                    "type": "string",
                    "example": "\u0412\u044b\u0445\u043e\u0434\u0438\u0442\u0435 \u0438\u0437 \u0441\u0442\u0430\u043d\u0446\u0438\u0438 \u043c\u0435\u0442\u0440\u043e <b>\u041d\u043e\u0432\u043e\u043a\u0443\u0437\u043d\u0435\u0446\u043a\u0430\u044f </b>\u043d\u0430 \u0443\u043b\u0438\u0446\u0443 \u041d\u043e\u0432\u043e\u043a\u0443\u0437\u043d\u0435\u0446\u043a\u0430\u044f..."
                },
                "phones": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    },
                    "example": [
                        "+7(495) 363-63-93"
                    ]
                },
                "photos": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    },
                    "example": [
                        "//files.lmcdn.ru/delivery/ru/75/6978cffab68e1e0ce126fa0d42368e75.png"
                    ]
                },
                "method_code": {
                    "type": "string",
                    "example": "lamoda_showroom_novokuznetskaya"
                },
                "method_name": {
                    "type": "string",
                    "example": "\u041d\u043e\u0432\u043e\u043a\u0443\u0437\u043d\u0435\u0446\u043a\u0430\u044f, \u043c. \u041d\u043e\u0432\u043e\u043a\u0443\u0437\u043d\u0435\u0446\u043a\u0430\u044f, \u043c.\u0422\u0440\u0435\u0442\u044c\u044f\u043a\u043e\u0432\u0441\u043a\u0430\u044f, \u043c. \u041f\u0430\u0432\u0435\u043b\u0435\u0446\u043a\u0430\u044f"
                },
                "method_type": {
                    "description": "1 - \u043a\u0443\u0440\u044c\u0435\u0440\u0441\u043a\u0430\u044f \u0434\u043e\u0441\u0442\u0430\u0432\u043a\u0430, 2 - \u0441\u0430\u043c\u043e\u0432\u044b\u0432\u043e\u0437, 3 - \u043f\u043e\u0447\u0442\u043e\u0432\u0430\u044f \u0434\u043e\u0441\u0442\u0430\u0432\u043a\u0430",
                    "type": "string",
                    "example": 2
                },
                "latitude": {
                    "type": "number",
                    "format": "float",
                    "example": 55.738176
                },
                "longitude": {
                    "type": "number",
                    "format": "float",
                    "example": 37.631595
                },
                "is_bankcard_accepted": {
                    "type": "boolean",
                    "example": true
                },
                "is_24hours": {
                    "type": "boolean",
                    "example": false
                },
                "group_id": {
                    "type": "integer",
                    "example": 4
                },
                "checkout_name": {
                    "type": "string",
                    "example": "\u0414\u043e\u0441\u0442\u0430\u0432\u043a\u0430 \u0432 \u043f\u0443\u043d\u043a\u0442 \u0432\u044b\u0434\u0430\u0447\u0438 \u0437\u0430\u043a\u0430\u0437\u043e\u0432 Lamoda"
                },
                "communication_description": {
                    "type": "string",
                    "example": "{\u041a\u043e\u0433\u0434\u0430 \u0437\u0430\u043a\u0430\u0437 \u043f\u043e\u0441\u0442\u0443\u043f\u0438\u0442 \u0432 \u043f\u0443\u043d\u043a\u0442 \u0441\u0430\u043c\u043e\u0432\u044b\u0432\u043e\u0437\u0430, \u0432\u0430\u043c \u043f\u0440\u0438\u0434\u0435\u0442 SMS..."
                },
                "lead_time": {
                    "$ref": "#/definitions/LeadTime"
                },
                "service_levels": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/GetMethodDetails\\ServiceLevel"
                    }
                },
                "is_own": {
                    "type": "boolean"
                },
                "tracking_url": {
                    "type": "string"
                },
                "id": {
                    "type": "integer",
                    "example": 16213
                },
                "external_id": {
                    "type": "string"
                },
                "name": {
                    "type": "string",
                    "example": "\u041d\u043e\u0432\u043e\u043a\u0443\u0437\u043d\u0435\u0446\u043a\u0430\u044f"
                },
                "city": {
                    "type": "string",
                    "example": "\u041c\u043e\u0441\u043a\u0432\u0430"
                },
                "street": {
                    "type": "string",
                    "example": "\u041d\u043e\u0432\u043e\u043a\u0443\u0437\u043d\u0435\u0446\u043a\u0430\u044f"
                },
                "house": {
                    "type": "string",
                    "example": "11/13"
                },
                "office": {
                    "type": "string"
                },
                "underground_station": {
                    "type": "string",
                    "example": "\u043c. \u041d\u043e\u0432\u043e\u043a\u0443\u0437\u043d\u0435\u0446\u043a\u0430\u044f/\u0422\u0440\u0435\u0442\u044c\u044f\u043a\u043e\u0432\u0441\u043a\u0430\u044f/\u041f\u0430\u0432\u0435\u043b\u0435\u0446\u043a\u0430\u044f"
                }
            },
            "type": "object"
        },
        "PupDetailsResponse": {
            "required": [
                "storage_days",
                "is_returns_accepted",
                "method_code",
                "method_name",
                "method_type",
                "delivery_price",
                "available_days",
                "rejection_fee",
                "is_client_name_required",
                "service_level_types"
            ],
            "properties": {
                "storage_days": {
                    "type": "string",
                    "example": 2
                },
                "is_returns_accepted": {
                    "type": "boolean",
                    "example": true
                },
                "widget_description": {
                    "type": "string",
                    "example": "\u0412\u044b\u0445\u043e\u0434\u0438\u0442\u0435 \u0438\u0437 \u0441\u0442\u0430\u043d\u0446\u0438\u0438 \u043c\u0435\u0442\u0440\u043e <b>\u041d\u043e\u0432\u043e\u043a\u0443\u0437\u043d\u0435\u0446\u043a\u0430\u044f </b>\u043d\u0430 \u0443\u043b\u0438\u0446\u0443 \u041d\u043e\u0432\u043e\u043a\u0443\u0437\u043d\u0435\u0446\u043a\u0430\u044f..."
                },
                "phones": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    },
                    "example": [
                        "+7(495) 363-63-93"
                    ]
                },
                "photos": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    },
                    "example": [
                        "//files.lmcdn.ru/delivery/ru/75/6978cffab68e1e0ce126fa0d42368e75.png"
                    ]
                },
                "method_code": {
                    "type": "string",
                    "example": "lamoda_showroom_novokuznetskaya"
                },
                "method_name": {
                    "type": "string",
                    "example": "\u041d\u043e\u0432\u043e\u043a\u0443\u0437\u043d\u0435\u0446\u043a\u0430\u044f, \u043c. \u041d\u043e\u0432\u043e\u043a\u0443\u0437\u043d\u0435\u0446\u043a\u0430\u044f, \u043c.\u0422\u0440\u0435\u0442\u044c\u044f\u043a\u043e\u0432\u0441\u043a\u0430\u044f, \u043c. \u041f\u0430\u0432\u0435\u043b\u0435\u0446\u043a\u0430\u044f"
                },
                "method_type": {
                    "description": "1 - \u043a\u0443\u0440\u044c\u0435\u0440\u0441\u043a\u0430\u044f \u0434\u043e\u0441\u0442\u0430\u0432\u043a\u0430, 2 - \u0441\u0430\u043c\u043e\u0432\u044b\u0432\u043e\u0437, 3 - \u043f\u043e\u0447\u0442\u043e\u0432\u0430\u044f \u0434\u043e\u0441\u0442\u0430\u0432\u043a\u0430",
                    "type": "string",
                    "example": 2
                },
                "communication_description": {
                    "type": "string",
                    "example": "{\u041a\u043e\u0433\u0434\u0430 \u0437\u0430\u043a\u0430\u0437 \u043f\u043e\u0441\u0442\u0443\u043f\u0438\u0442 \u0432 \u043f\u0443\u043d\u043a\u0442 \u0441\u0430\u043c\u043e\u0432\u044b\u0432\u043e\u0437\u0430, \u0432\u0430\u043c \u043f\u0440\u0438\u0434\u0435\u0442 SMS..."
                },
                "delivery_price": {
                    "type": "number",
                    "format": "float",
                    "example": 250
                },
                "free_delivery_net_threshold": {
                    "type": "number",
                    "format": "float",
                    "example": 2500
                },
                "available_days": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/DayWithPalletMethod"
                    }
                },
                "checkout_description": {
                    "type": "string",
                    "example": "\u0414\u043e\u0441\u0442\u0430\u0432\u043a\u0430 \u0437\u0430\u043a\u0430\u0437\u0430 \u0432 \u043f\u0443\u043d\u043a\u0442 \u0432\u044b\u0434\u0430\u0447\u0438 \u0437\u0430\u043a\u0430\u0437\u043e\u0432 \u0431\u0435\u0441\u043f\u043b\u0430\u0442\u043d\u0430 \u0432 \u0441\u043b\u0443\u0447\u0430\u0435 ..."
                },
                "callcenter_description": {
                    "type": "string",
                    "example": "\u041f\u0440\u0438 \u0432\u044b\u043a\u0443\u043f\u0435 \u043d\u0430 \u0441\u0443\u043c\u043c\u0443 \u0431\u043e\u043b\u0435\u0435 2500 \u0440\u0443\u0431\u043b\u0435\u0439 \u0434\u043e\u0441\u0442\u0430\u0432\u043a\u0430 \u0431\u0435\u0441\u043f\u043b\u0430\u0442\u043d\u0430\u044f ..."
                },
                "rejection_fee": {
                    "type": "number",
                    "format": "float",
                    "example": 250
                },
                "is_client_name_required": {
                    "type": "boolean",
                    "example": false
                },
                "service_level_types": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/PupDetailsServiceLevelType"
                    }
                },
                "latitude": {
                    "type": "number",
                    "format": "float",
                    "example": 55.738176
                },
                "longitude": {
                    "type": "number",
                    "format": "float",
                    "example": 37.631595
                },
                "is_bankcard_accepted": {
                    "type": "boolean",
                    "example": true
                },
                "is_24hours": {
                    "type": "boolean",
                    "example": false
                },
                "group_id": {
                    "type": "integer",
                    "example": 4
                },
                "work_time": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/WorkTime"
                    }
                },
                "delivery_date_min": {
                    "type": "string",
                    "example": "2017-03-04"
                },
                "delivery_date_max": {
                    "type": "string",
                    "example": "2017-03-08"
                },
                "lead_time": {
                    "$ref": "#/definitions/LeadTime"
                },
                "has_intervals": {
                    "type": "boolean"
                },
                "has_horizon": {
                    "type": "boolean"
                },
                "zone_id": {
                    "type": "integer"
                },
                "id": {
                    "type": "integer",
                    "example": 16213
                },
                "external_id": {
                    "type": "string"
                },
                "name": {
                    "type": "string",
                    "example": "\u041d\u043e\u0432\u043e\u043a\u0443\u0437\u043d\u0435\u0446\u043a\u0430\u044f"
                },
                "city": {
                    "type": "string",
                    "example": "\u041c\u043e\u0441\u043a\u0432\u0430"
                },
                "street": {
                    "type": "string",
                    "example": "\u041d\u043e\u0432\u043e\u043a\u0443\u0437\u043d\u0435\u0446\u043a\u0430\u044f"
                },
                "house": {
                    "type": "string",
                    "example": "11/13"
                },
                "office": {
                    "type": "string"
                },
                "underground_station": {
                    "type": "string",
                    "example": "\u043c. \u041d\u043e\u0432\u043e\u043a\u0443\u0437\u043d\u0435\u0446\u043a\u0430\u044f/\u0422\u0440\u0435\u0442\u044c\u044f\u043a\u043e\u0432\u0441\u043a\u0430\u044f/\u041f\u0430\u0432\u0435\u043b\u0435\u0446\u043a\u0430\u044f"
                }
            },
            "type": "object"
        },
        "PupDetailsServiceLevelType": {
            "required": [
                "code",
                "name",
                "is_tryon_allowed",
                "is_rejection_allowed",
                "has_horizon",
                "has_intervals",
                "available_days",
                "delivery_price",
                "rejection_fee"
            ],
            "properties": {
                "code": {
                    "type": "string",
                    "example": "plus"
                },
                "name": {
                    "type": "string",
                    "example": "\u041f\u043b\u044e\u0441"
                },
                "description": {
                    "type": "string"
                },
                "is_tryon_allowed": {
                    "type": "boolean",
                    "example": true
                },
                "is_rejection_allowed": {
                    "type": "boolean",
                    "example": true
                },
                "free_delivery_net_threshold": {
                    "type": "number",
                    "format": "float",
                    "example": 2500
                },
                "checkout_description": {
                    "type": "string",
                    "example": "\u0414\u043e\u0441\u0442\u0430\u0432\u043a\u0430 \u0437\u0430\u043a\u0430\u0437\u0430 \u0432 \u043f\u0443\u043d\u043a\u0442 \u0432\u044b\u0434\u0430\u0447\u0438 \u0437\u0430\u043a\u0430\u0437\u043e\u0432 \u0431\u0435\u0441\u043f\u043b\u0430\u0442\u043d\u0430 \u0432 \u0441\u043b\u0443\u0447\u0430\u0435 ..."
                },
                "callcenter_description": {
                    "type": "string",
                    "example": "\u041f\u0440\u0438 \u0432\u044b\u043a\u0443\u043f\u0435 \u043d\u0430 \u0441\u0443\u043c\u043c\u0443 \u0431\u043e\u043b\u0435\u0435 2500 \u0440\u0443\u0431\u043b\u0435\u0439 \u0434\u043e\u0441\u0442\u0430\u0432\u043a\u0430 \u0431\u0435\u0441\u043f\u043b\u0430\u0442\u043d\u0430\u044f ..."
                },
                "has_horizon": {
                    "type": "boolean"
                },
                "has_intervals": {
                    "type": "boolean"
                },
                "available_days": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/DayWithPalletMethod"
                    }
                },
                "delivery_price": {
                    "type": "number",
                    "format": "float",
                    "example": 250
                },
                "rejection_fee": {
                    "type": "number",
                    "format": "float",
                    "example": 250
                },
                "day_min": {
                    "type": "integer",
                    "example": 1
                },
                "day_max": {
                    "type": "integer",
                    "example": 5
                },
                "is_callcenter_confirm": {
                    "type": "boolean",
                    "example": false
                }
            },
            "type": "object"
        },
        "PupListResponse": {
            "required": [
                "is_bankcard_accepted",
                "is_24hours",
                "has_intervals",
                "has_horizon"
            ],
            "properties": {
                "latitude": {
                    "type": "number",
                    "format": "float",
                    "example": 55.738176
                },
                "longitude": {
                    "type": "number",
                    "format": "float",
                    "example": 37.631595
                },
                "is_bankcard_accepted": {
                    "type": "boolean",
                    "example": true
                },
                "is_24hours": {
                    "type": "boolean",
                    "example": false
                },
                "group_id": {
                    "type": "integer",
                    "example": 4
                },
                "work_time": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/WorkTime"
                    }
                },
                "delivery_date_min": {
                    "type": "string",
                    "example": "2017-03-04"
                },
                "delivery_date_max": {
                    "type": "string",
                    "example": "2017-03-08"
                },
                "lead_time": {
                    "$ref": "#/definitions/LeadTime"
                },
                "has_intervals": {
                    "type": "boolean"
                },
                "has_horizon": {
                    "type": "boolean"
                },
                "zone_id": {
                    "type": "integer"
                },
                "id": {
                    "type": "integer",
                    "example": 16213
                },
                "external_id": {
                    "type": "string"
                },
                "name": {
                    "type": "string",
                    "example": "\u041d\u043e\u0432\u043e\u043a\u0443\u0437\u043d\u0435\u0446\u043a\u0430\u044f"
                },
                "city": {
                    "type": "string",
                    "example": "\u041c\u043e\u0441\u043a\u0432\u0430"
                },
                "street": {
                    "type": "string",
                    "example": "\u041d\u043e\u0432\u043e\u043a\u0443\u0437\u043d\u0435\u0446\u043a\u0430\u044f"
                },
                "house": {
                    "type": "string",
                    "example": "11/13"
                },
                "office": {
                    "type": "string"
                },
                "underground_station": {
                    "type": "string",
                    "example": "\u043c. \u041d\u043e\u0432\u043e\u043a\u0443\u0437\u043d\u0435\u0446\u043a\u0430\u044f/\u0422\u0440\u0435\u0442\u044c\u044f\u043a\u043e\u0432\u0441\u043a\u0430\u044f/\u041f\u0430\u0432\u0435\u043b\u0435\u0446\u043a\u0430\u044f"
                }
            },
            "type": "object"
        },
        "PupResponse": {
            "required": [
                "id",
                "name",
                "city",
                "street",
                "house"
            ],
            "properties": {
                "id": {
                    "type": "integer",
                    "example": 16213
                },
                "external_id": {
                    "type": "string"
                },
                "name": {
                    "type": "string",
                    "example": "\u041d\u043e\u0432\u043e\u043a\u0443\u0437\u043d\u0435\u0446\u043a\u0430\u044f"
                },
                "city": {
                    "type": "string",
                    "example": "\u041c\u043e\u0441\u043a\u0432\u0430"
                },
                "street": {
                    "type": "string",
                    "example": "\u041d\u043e\u0432\u043e\u043a\u0443\u0437\u043d\u0435\u0446\u043a\u0430\u044f"
                },
                "house": {
                    "type": "string",
                    "example": "11/13"
                },
                "office": {
                    "type": "string"
                },
                "underground_station": {
                    "type": "string",
                    "example": "\u043c. \u041d\u043e\u0432\u043e\u043a\u0443\u0437\u043d\u0435\u0446\u043a\u0430\u044f/\u0422\u0440\u0435\u0442\u044c\u044f\u043a\u043e\u0432\u0441\u043a\u0430\u044f/\u041f\u0430\u0432\u0435\u043b\u0435\u0446\u043a\u0430\u044f"
                }
            },
            "type": "object"
        },
        "WorkTime": {
            "required": [
                "day",
                "time_from",
                "time_to"
            ],
            "properties": {
                "day": {
                    "description": "day of week from 1 to 7",
                    "type": "integer",
                    "example": 1
                },
                "time_from": {
                    "type": "string",
                    "example": "10:00"
                },
                "time_to": {
                    "type": "string",
                    "example": "22:00"
                }
            },
            "type": "object"
        },
        "ReserveResponse": {
            "properties": {
                "interval_id": {
                    "type": "integer",
                    "example": 40704062
                },
                "success": {
                    "type": "boolean"
                },
                "error": {
                    "$ref": "#/definitions/ResponseError"
                }
            },
            "type": "object"
        },
        "Response": {
            "required": [
                "success"
            ],
            "properties": {
                "success": {
                    "type": "boolean"
                },
                "error": {
                    "$ref": "#/definitions/ResponseError"
                }
            },
            "type": "object"
        },
        "ResponseError": {
            "required": [
                "code",
                "message"
            ],
            "properties": {
                "code": {
                    "type": "integer"
                },
                "message": {
                    "type": "string"
                }
            },
            "type": "object"
        },
        "ReturnMethod": {
            "required": [
                "code",
                "name",
                "zones"
            ],
            "properties": {
                "code": {
                    "type": "string",
                    "example": "pups_return"
                },
                "name": {
                    "type": "string",
                    "example": "\u041f\u0443\u043d\u043a\u0442\u044b \u0441\u0430\u043c\u043e\u0432\u044b\u0432\u043e\u0437\u0430"
                },
                "description": {
                    "type": "string"
                },
                "zones": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/ReturnMethodZone"
                    }
                }
            },
            "type": "object"
        },
        "ReturnMethodZone": {
            "required": [
                "id",
                "name"
            ],
            "properties": {
                "id": {
                    "type": "integer",
                    "example": 1
                },
                "name": {
                    "type": "string",
                    "example": "\u041f\u0412\u0417 \u041e\u0440\u0434\u0436\u043e\u043d\u0438\u043a\u0438\u0434\u0436\u0435"
                },
                "description": {
                    "type": "string"
                }
            },
            "type": "object"
        },
        "ServiceLevelType": {
            "required": [
                "code",
                "name",
                "methods"
            ],
            "properties": {
                "code": {
                    "type": "string",
                    "example": "plus"
                },
                "name": {
                    "type": "string",
                    "example": "\u041f\u043b\u044e\u0441"
                },
                "description": {
                    "type": "string"
                },
                "is_tryon_allowed": {
                    "type": "boolean"
                },
                "is_rejection_allowed": {
                    "type": "boolean"
                },
                "methods": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/CheckoutMethod"
                    }
                }
            },
            "type": "object"
        },
        "UpdateZoneRate\\Rate": {
            "required": [
                "weight_min",
                "weight_max",
                "price"
            ],
            "properties": {
                "weight_min": {
                    "type": "number",
                    "format": "float",
                    "example": 0
                },
                "weight_max": {
                    "type": "number",
                    "format": "float",
                    "example": 1
                },
                "price": {
                    "type": "number",
                    "format": "float",
                    "example": 150
                }
            },
            "type": "object"
        },
        "UpdateZoneRate\\Request": {
            "required": [
                "method_code",
                "zones"
            ],
            "properties": {
                "method_code": {
                    "type": "string",
                    "example": "lamoda_showroom_novokuznetskaya"
                },
                "is_update": {
                    "type": "boolean",
                    "example": false
                },
                "zones": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/UpdateZoneRate\\Zone"
                    }
                }
            },
            "type": "object"
        },
        "UpdateZoneRate\\Response": {
            "properties": {
                "success": {
                    "type": "boolean"
                },
                "error": {
                    "$ref": "#/definitions/ResponseError"
                }
            },
            "type": "object"
        },
        "UpdateZoneRate\\Zone": {
            "properties": {
                "zone_id": {
                    "type": "integer",
                    "example": 104185
                },
                "pickup_id": {
                    "type": "integer",
                    "example": 16213
                },
                "pickup_external_id": {
                    "type": "string"
                },
                "rates": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/UpdateZoneRate\\Rate"
                    }
                }
            },
            "type": "object"
        },
        "JsonRpcError": {
            "required": [
                "code",
                "message"
            ],
            "properties": {
                "code": {
                    "type": "integer"
                },
                "message": {
                    "type": "string"
                }
            },
            "type": "object"
        },
        "JsonRpcResponse": {
            "required": [
                "jsonrpc",
                "id"
            ],
            "properties": {
                "jsonrpc": {
                    "type": "string"
                },
                "id": {
                    "type": "string"
                },
                "error": {
                    "$ref": "#/definitions/JsonRpcError"
                }
            },
            "type": "object"
        },
        "JsonRpcOkResponse": {
            "allOf": [
                {
                    "$ref": "#/definitions/JsonRpcResponse"
                },
                {
                    "properties": {
                        "result": {
                            "$ref": "#/definitions/Response"
                        }
                    },
                    "type": "object"
                }
            ]
        },
        "JsonRpcRequest": {
            "required": [
                "jsonrpc",
                "id"
            ],
            "properties": {
                "jsonrpc": {
                    "type": "string"
                },
                "id": {
                    "type": "string"
                }
            },
            "type": "object"
        },
        "JsonRpcMethodsGetResponse": {
            "allOf": [
                {
                    "$ref": "#/definitions/JsonRpcResponse"
                },
                {
                    "properties": {
                        "result": {
                            "$ref": "#/definitions/GetMethodsResponse"
                        }
                    },
                    "type": "object"
                }
            ]
        },
        "JsonRpcMethodsGetCheckoutResponse": {
            "allOf": [
                {
                    "$ref": "#/definitions/JsonRpcResponse"
                },
                {
                    "properties": {
                        "result": {
                            "$ref": "#/definitions/GetCheckoutMethodsResponse"
                        }
                    },
                    "type": "object"
                }
            ]
        },
        "JsonRpcMethodsGetCheckoutMultipleResponse": {
            "allOf": [
                {
                    "$ref": "#/definitions/JsonRpcResponse"
                },
                {
                    "properties": {
                        "result": {
                            "$ref": "#/definitions/GetCheckoutMethodsMultiple\\Response"
                        }
                    },
                    "type": "object"
                }
            ]
        },
        "JsonRpcMethodsGetDetailsResponse": {
            "allOf": [
                {
                    "$ref": "#/definitions/JsonRpcResponse"
                },
                {
                    "properties": {
                        "result": {
                            "$ref": "#/definitions/GetMethodDetails\\Response"
                        }
                    },
                    "type": "object"
                }
            ]
        },
        "JsonRpcMethodsGetDetailsMultipleResponse": {
            "allOf": [
                {
                    "$ref": "#/definitions/JsonRpcResponse"
                },
                {
                    "properties": {
                        "result": {
                            "$ref": "#/definitions/GetMethodDetailsMultiple\\Response"
                        }
                    },
                    "type": "object"
                }
            ]
        },
        "JsonRpcMethodsGetAllResponse": {
            "allOf": [
                {
                    "$ref": "#/definitions/JsonRpcResponse"
                },
                {
                    "properties": {
                        "result": {
                            "$ref": "#/definitions/GetAllMethodsResponse"
                        }
                    },
                    "type": "object"
                }
            ]
        },
        "JsonRpcDeliveryGetInfoResponse": {
            "allOf": [
                {
                    "$ref": "#/definitions/JsonRpcResponse"
                },
                {
                    "properties": {
                        "result": {
                            "$ref": "#/definitions/GetDeliveryInfoResponse"
                        }
                    },
                    "type": "object"
                }
            ]
        },
        "JsonRpcDeliveryGetInfoShortResponse": {
            "allOf": [
                {
                    "$ref": "#/definitions/JsonRpcResponse"
                },
                {
                    "properties": {
                        "result": {
                            "$ref": "#/definitions/GetDeliveryInfoShortResponse"
                        }
                    },
                    "type": "object"
                }
            ]
        },
        "JsonRpcIntervalsReserveResponse": {
            "allOf": [
                {
                    "$ref": "#/definitions/JsonRpcResponse"
                },
                {
                    "properties": {
                        "result": {
                            "$ref": "#/definitions/ReserveResponse"
                        }
                    },
                    "type": "object"
                }
            ]
        },
        "JsonRpcIntervalsGetResponse": {
            "allOf": [
                {
                    "$ref": "#/definitions/JsonRpcResponse"
                },
                {
                    "properties": {
                        "result": {
                            "$ref": "#/definitions/GetIntervalResponse"
                        }
                    },
                    "type": "object"
                }
            ]
        },
        "JsonRpcIntervalsFindResponse": {
            "allOf": [
                {
                    "$ref": "#/definitions/JsonRpcResponse"
                },
                {
                    "properties": {
                        "result": {
                            "$ref": "#/definitions/FindIntervalResponse"
                        }
                    },
                    "type": "object"
                }
            ]
        },
        "JsonRpcPickupsGetListResponse": {
            "allOf": [
                {
                    "$ref": "#/definitions/JsonRpcResponse"
                },
                {
                    "properties": {
                        "result": {
                            "$ref": "#/definitions/GetPupListResponse"
                        }
                    },
                    "type": "object"
                }
            ]
        },
        "JsonRpcPickupsGetResponse": {
            "allOf": [
                {
                    "$ref": "#/definitions/JsonRpcResponse"
                },
                {
                    "properties": {
                        "result": {
                            "$ref": "#/definitions/GetPickupResponse"
                        }
                    },
                    "type": "object"
                }
            ]
        },
        "JsonRpcPickupsGetDetailsResponse": {
            "allOf": [
                {
                    "$ref": "#/definitions/JsonRpcResponse"
                },
                {
                    "properties": {
                        "result": {
                            "$ref": "#/definitions/GetPupDetailsResponse"
                        }
                    },
                    "type": "object"
                }
            ]
        },
        "JsonRpcPickupsCreateRequest": {
            "allOf": [
                {
                    "$ref": "#/definitions/JsonRpcRequest"
                },
                {
                    "properties": {
                        "params": {
                            "$ref": "#/definitions/CreatePickups\\Request"
                        }
                    },
                    "type": "object"
                }
            ]
        },
        "JsonRpcPickupsCreateResponse": {
            "allOf": [
                {
                    "$ref": "#/definitions/JsonRpcResponse"
                },
                {
                    "properties": {
                        "result": {
                            "$ref": "#/definitions/CreatePickups\\Response"
                        }
                    },
                    "type": "object"
                }
            ]
        },
        "JsonRpcProfilesImportRequest": {
            "allOf": [
                {
                    "$ref": "#/definitions/JsonRpcRequest"
                },
                {
                    "properties": {
                        "params": {
                            "type": "object"
                        }
                    },
                    "type": "object"
                }
            ]
        },
        "JsonRpcZonesUpdateRateRequest": {
            "allOf": [
                {
                    "$ref": "#/definitions/JsonRpcRequest"
                },
                {
                    "properties": {
                        "params": {
                            "$ref": "#/definitions/UpdateZoneRate\\Request"
                        }
                    },
                    "type": "object"
                }
            ]
        },
        "JsonRpcZonesUpdateRateResponse": {
            "allOf": [
                {
                    "$ref": "#/definitions/JsonRpcResponse"
                },
                {
                    "properties": {
                        "result": {
                            "$ref": "#/definitions/UpdateZoneRate\\Response"
                        }
                    },
                    "type": "object"
                }
            ]
        },
        "JsonRpcZonesGetAttributesResponse": {
            "allOf": [
                {
                    "$ref": "#/definitions/JsonRpcResponse"
                },
                {
                    "properties": {
                        "result": {
                            "$ref": "#/definitions/GetZoneAttributesResponse"
                        }
                    },
                    "type": "object"
                }
            ]
        },
        "JsonRpcZonesGetRateResponse": {
            "allOf": [
                {
                    "$ref": "#/definitions/JsonRpcResponse"
                },
                {
                    "properties": {
                        "result": {
                            "$ref": "#/definitions/GetCheckoutMethodsMultiple\\Response"
                        }
                    },
                    "type": "object"
                }
            ]
        }
    },
    "parameters": {
        "aoid": {
            "name": "aoid",
            "in": "query",
            "type": "string"
        },
        "cart_amount": {
            "name": "cart_amount",
            "in": "query",
            "type": "integer"
        },
        "checkout": {
            "name": "checkout",
            "in": "query",
            "type": "boolean",
            "default": false
        },
        "code": {
            "name": "code",
            "in": "query",
            "required": true,
            "type": "string"
        },
        "company": {
            "name": "company",
            "in": "query",
            "type": "string"
        },
        "current_id": {
            "name": "current_id",
            "in": "query",
            "type": "integer"
        },
        "desired_end": {
            "name": "desired_end",
            "in": "query",
            "description": "YYYY-MM-DD HH:MM:SS",
            "required": true,
            "type": "string"
        },
        "desired_start": {
            "name": "desired_start",
            "in": "query",
            "description": "YYYY-MM-DD HH:MM:SS",
            "required": true,
            "type": "string"
        },
        "force": {
            "name": "force",
            "in": "query",
            "type": "boolean",
            "default": false
        },
        "get_method_details_multiple_aoid": {
            "name": "getMethodDetails[0][aoid]",
            "in": "query",
            "type": "string"
        },
        "get_method_details_multiple_code": {
            "name": "getMethodDetails[0][code]",
            "in": "query",
            "required": true,
            "type": "string"
        },
        "get_method_details_multiple_join": {
            "name": "getMethodDetails[0][join]",
            "in": "query",
            "description": "Comma-separated list of values. Supported values may vary in methods. Possible values: company, zone, pickup, service_level, lead_time.",
            "type": "string"
        },
        "get_method_details_multiple_origin_date": {
            "name": "getMethodDetails[0][origin_date]",
            "in": "query",
            "description": "The origin of horizon, default is current date",
            "type": "string",
            "format": "date"
        },
        "get_method_details_multiple_service_level": {
            "name": "getMethodDetails[0][service_level]",
            "in": "query",
            "type": "string"
        },
        "get_method_details_multiple_zipcode": {
            "name": "getMethodDetails[0][zipcode]",
            "in": "query",
            "type": "string"
        },
        "get_method_details_multiple_zone_id": {
            "name": "getMethodDetails[0][zone_id]",
            "in": "query",
            "type": "integer"
        },
        "ignore_capacity": {
            "name": "ignore_capacity",
            "in": "query",
            "type": "boolean",
            "default": false
        },
        "ignore_cutoff": {
            "name": "ignore_cutoff",
            "in": "query",
            "type": "boolean",
            "default": false
        },
        "insurance_sum": {
            "name": "insurance_sum",
            "in": "query",
            "type": "number",
            "format": "float"
        },
        "interval_id": {
            "name": "interval_id",
            "in": "query",
            "required": true,
            "type": "integer"
        },
        "item_count": {
            "name": "item_count",
            "in": "query",
            "type": "integer"
        },
        "join": {
            "name": "join",
            "in": "query",
            "description": "Comma-separated list of values. Supported values may vary in methods. Possible values: company, zone, pickup, service_level, lead_time.",
            "type": "string"
        },
        "latitude": {
            "name": "latitude",
            "in": "query",
            "type": "number"
        },
        "liquids": {
            "name": "liquids",
            "in": "query",
            "type": "boolean",
            "default": false
        },
        "longitude": {
            "name": "longitude",
            "in": "query",
            "type": "number"
        },
        "method_code": {
            "name": "method_code",
            "in": "query",
            "type": "string"
        },
        "method_code_required": {
            "name": "method_code",
            "in": "query",
            "required": true,
            "type": "string"
        },
        "no_airfreight": {
            "name": "no_airfreight",
            "in": "query",
            "type": "boolean",
            "default": false
        },
        "orders_cart_amount": {
            "name": "orders[0][cart_amount]",
            "in": "query",
            "type": "integer"
        },
        "orders_item_count": {
            "name": "orders[0][item_count]",
            "in": "query",
            "type": "integer"
        },
        "orders_no_airfreight": {
            "name": "orders[0][no_airfreight]",
            "in": "query",
            "type": "boolean",
            "default": false
        },
        "orders_payment_type": {
            "name": "orders[0][payment_type]",
            "in": "query",
            "description": "0 - all types<br />1 - prepayment<br />2 - postpayment",
            "type": "integer",
            "default": 2
        },
        "orders_profile_id": {
            "name": "orders[0][profile_id]",
            "in": "query",
            "type": "string",
            "default": "LM"
        },
        "origin_date": {
            "name": "origin_date",
            "in": "query",
            "description": "The origin of horizon, default is current date",
            "type": "string",
            "format": "date"
        },
        "original_delivery_date": {
            "name": "original_delivery_date",
            "in": "query",
            "description": "If provided, results will be filtered to content only dates equal to or later than this",
            "type": "string",
            "format": "date"
        },
        "pallet_method_code": {
            "name": "pallet_method_code",
            "in": "query",
            "type": "string"
        },
        "payment_type": {
            "name": "payment_type",
            "in": "query",
            "description": "0 - all types<br />1 - prepayment<br />2 - postpayment",
            "type": "integer",
            "default": 2
        },
        "pickup_id": {
            "name": "pickup_id",
            "in": "query",
            "required": true,
            "type": "integer"
        },
        "profile_id": {
            "name": "profile_id",
            "in": "query",
            "type": "string",
            "default": "LM"
        },
        "reserved_interval_id": {
            "name": "reserved_interval_id",
            "in": "query",
            "type": "integer"
        },
        "service_level": {
            "name": "service_level",
            "in": "query",
            "type": "string"
        },
        "service_level_type_code": {
            "name": "service_level_type_code",
            "in": "query",
            "type": "string"
        },
        "tag": {
            "name": "tag",
            "in": "query",
            "type": "string"
        },
        "weight": {
            "name": "weight",
            "in": "query",
            "required": true,
            "type": "number",
            "format": "float"
        },
        "zipcode": {
            "name": "zipcode",
            "in": "query",
            "type": "string"
        },
        "zipcode_required": {
            "name": "zipcode",
            "in": "query",
            "required": true,
            "type": "string"
        },
        "zone_id": {
            "name": "zone_id",
            "in": "query",
            "type": "integer"
        },
        "zone_id_required": {
            "name": "zone_id",
            "in": "query",
            "required": true,
            "type": "integer"
        },
        "jsonrpc_id": {
            "name": "id",
            "in": "query",
            "required": true,
            "type": "string"
        }
    },
    "tags": [
        {
            "name": "rest",
            "description": "Rest API"
        },
        {
            "name": "json-rpc",
            "description": "JSON-RPC API"
        },
        {
            "name": "delivery",
            "description": "Delivery info API"
        },
        {
            "name": "intervals",
            "description": "Intervals API"
        },
        {
            "name": "methods",
            "description": "Methods API"
        },
        {
            "name": "pickups",
            "description": "Pickups API"
        },
        {
            "name": "profiles",
            "description": "Profiles API"
        },
        {
            "name": "zones",
            "description": "Zones API"
        }
    ]
}
`
