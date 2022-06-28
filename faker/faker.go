package faker

import (
	"strings"

	"github.com/neotoolkit/faker"
)

func Faker(value string) string {
	key := getFakerKey(value)

	f := faker.NewFaker()

	return f.Faker(key)
}

func getFakerKey(value string) string {
	splitValue := strings.Split(value, "$faker")
	if len(splitValue) == 1 {
		return ""
	}

	if splitValue[0] != "" {
		return ""
	}

	return strings.ToLower(splitValue[1])
}
