package variables

import (
	"fmt"

	"github.com/tidwall/gjson"
)

func FromResponse(varsToSet map[string]string, body string, isJSON bool) (vars *Variables, err error) {
	names, paths := split(varsToSet)

	switch {
	case isJSON:
		vars, err = fromJSON(names, paths, body)
		if err != nil {
			return nil, err
		}
	default:
		vars, err = fromPlainText(names, body)
		if err != nil {
			return nil, err
		}
	}

	return vars, nil
}

func fromJSON(names, paths []string, body string) (*Variables, error) {
	vars := New()

	results := gjson.GetMany(body, paths...)

	for n, res := range results {
		if !res.Exists() {
			return nil,
				fmt.Errorf("path '%s' doesn't exist in given json", paths[n])
		}

		vars.Add(NewVariable(names[n], res.String()))
	}

	return vars, nil
}

func fromPlainText(names []string, body string) (*Variables, error) {
	if len(names) != 1 {
		return nil,
			fmt.Errorf(
				"count of variables for plain-text response should be 1, %d given",
				len(names),
			)
	}

	return New().Add(NewVariable(names[0], body)), nil
}

// split returns keys and values of given map as separate slices
func split(m map[string]string) (keys, values []string) {
	values = make([]string, 0, len(m))
	keys = make([]string, 0, len(m))

	for k, v := range m {
		keys = append(keys, k)
		values = append(values, v)
	}

	return keys, values
}
