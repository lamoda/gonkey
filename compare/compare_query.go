package compare

import (
	"fmt"
	"regexp"
)

func CompareQuery(expected, actual []string) (bool, error) {
	if len(expected) != len(actual) {
		return false, fmt.Errorf("expected and actual query params have different lengths")
	}

	remove := func(array []string, i int) []string {
		array[i] = array[len(array)-1]
		return array[:len(array)-1]
	}

	var expectedCopy = make([]string, len(expected))
	copy(expectedCopy, expected)
	var actualCopy = make([]string, len(actual))
	copy(actualCopy, actual)

	for len(expectedCopy) != 0 {
		found := false

		for i, expectedValue := range expectedCopy {
			for j, actualValue := range actualCopy {
				if matches := regexExprRx.FindStringSubmatch(expectedValue); matches != nil {
					rx, err := regexp.Compile(matches[1])
					if err != nil {
						return false, err
					}

					found = rx.MatchString(actualValue)
				} else {
					found = expectedValue == actualValue
				}

				if found {
					expectedCopy = remove(expectedCopy, i)
					actualCopy = remove(actualCopy, j)
					break
				}
			}

			if found {
				break
			}
		}

		if !found {
			return false, nil
		}
	}

	return true, nil
}
