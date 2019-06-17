package compare

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/fatih/color"
)

type CompareParams struct {
	IgnoreValues         bool
	IgnoreArraysOrdering bool
	DisallowExtraFields  bool
}

// Compare ...
func Compare(expected, actual interface{}, params CompareParams) []error {
	return compareBranch("$", expected, actual, &params)
}

func compareBranch(path string, expected, actual interface{}, params *CompareParams) []error {
	expectedType := getType(expected)
	actualType := getType(actual)
	var errors []error

	// compare types
	if expectedType != actualType {
		errors = append(errors, makeError(path, "types do not match", expectedType, actualType))
		return errors
	}

	// compare scalars
	if isScalarType(actualType) && !params.IgnoreValues && !isScalarsEqual(expected, actual) {
		errors = append(errors, makeError(path, "values do not match", expected, actual))
		return errors
	}

	// compare arrays
	if actualType == "array" {
		if params.IgnoreArraysOrdering {
			expected = sortArray(expected)
			actual = sortArray(actual)
		}

		expectedRef := reflect.ValueOf(expected)
		actualRef := reflect.ValueOf(actual)

		if expectedRef.Len() != actualRef.Len() {
			errors = append(errors, makeError(path, "array lengths do not match", expectedRef.Len(), actualRef.Len()))
			return errors
		}

		// iterate over children
		for i := 0; i < expectedRef.Len(); i++ {
			subPath := fmt.Sprintf("%s[%d]", path, i)
			res := compareBranch(subPath, expectedRef.Index(i).Interface(), actualRef.Index(i).Interface(), params)
			errors = append(errors, res...)
		}
	}

	// compare maps
	if actualType == "map" {
		expectedRef := reflect.ValueOf(expected)
		actualRef := reflect.ValueOf(actual)

		if params.DisallowExtraFields && expectedRef.Len() != actualRef.Len() {
			errors = append(errors, makeError(path, "map lengths do not match", expectedRef.Len(), actualRef.Len()))
			return errors
		}

		for _, key := range expectedRef.MapKeys() {
			// check keys presence
			if ok := actualRef.MapIndex(key); !ok.IsValid() {
				errors = append(errors, makeError(path, "key is missing", key.String(), "<missing>"))
				continue
			}

			// check values
			subPath := fmt.Sprintf("%s.%s", path, key.String())
			res := compareBranch(
				subPath,
				expectedRef.MapIndex(key).Interface(),
				actualRef.MapIndex(key).Interface(),
				params,
			)
			errors = append(errors, res...)
		}
	}

	return errors
}

func getType(value interface{}) string {
	if value == nil {
		return "nil"
	}
	rt := reflect.TypeOf(value)
	if rt.Kind() == reflect.Slice || rt.Kind() == reflect.Array {
		return "array"
	} else if rt.Kind() == reflect.Map {
		return "map"
	} else {
		return rt.String()
	}
}

func isScalarType(t string) bool {
	return !(t == "array" || t == "map")
}

func isScalarsEqual(value1, value2 interface{}) bool {
	return value1 == value2
}

func makeError(path, msg string, expected, actual interface{}) error {
	return fmt.Errorf(
		"at path %s %s:\n     expected: %s\n       actual: %s",
		color.CyanString(path),
		msg,
		color.GreenString("%v", expected),
		color.RedString("%v", actual),
	)
}

// Sort an array with respect of its elements of vary type.
func sortArray(array interface{}) interface{} {
	ref := reflect.ValueOf(array)

	interfaceSlice := make([]interface{}, 0)
	for i := 0; i < ref.Len(); i++ {
		interfaceSlice = append(interfaceSlice, ref.Index(i).Interface())
	}

	sort.Slice(interfaceSlice, func(i, j int) bool {
		str1 := representAnythingAsString(interfaceSlice[i])
		str2 := representAnythingAsString(interfaceSlice[j])
		return strings.Compare(str1, str2) < 0
	})

	return interfaceSlice
}

func representAnythingAsString(value interface{}) string {
	if value == nil {
		return ""
	}

	valueType := getType(value)

	if valueType == "array" {
		// sort array
		value = sortArray(value)
		ref := reflect.ValueOf(value)

		// represent array elements as a string
		var stringChunks []string
		for i := 0; i < ref.Len(); i++ {
			stringChunks = append(stringChunks, representAnythingAsString(ref.Index(i).Interface()))
		}
		return strings.Join(stringChunks, ".")
	}

	if valueType == "map" {
		ref := reflect.ValueOf(value)

		// sort keys ascending
		mapKeys := ref.MapKeys()
		sort.Slice(mapKeys, func(i, j int) bool {
			return strings.Compare(mapKeys[i].String(), mapKeys[j].String()) < 0
		})

		// represent map keys & elements as a string
		var stringChunks []string
		for i := 0; i < len(mapKeys); i++ {
			stringChunks = append(stringChunks, mapKeys[i].String())
			stringChunks = append(stringChunks, representAnythingAsString(ref.MapIndex(mapKeys[i]).Interface()))
		}
		return strings.Join(stringChunks, ".")
	}

	// scalars
	return fmt.Sprintf("%v", value)
}
