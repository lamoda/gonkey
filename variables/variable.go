package variables

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/neotoolkit/faker"
)

type Variable struct {
	name         string
	value        string
	defaultValue string
	rx           *regexp.Regexp
}

// NewVariable creates new variable with given name and value
func NewVariable(name, value string) *Variable {

	name = regexp.QuoteMeta(name)
	rx := regexp.MustCompile(fmt.Sprintf(`{{\s*\$%s\s*}}`, name))

	if strings.HasPrefix(value, "faker.") {
		f := faker.NewFaker()
		fName := strings.Trim(value, "faker.")
		value = f.Faker(fName)
	}

	return &Variable{
		name:         name,
		value:        value,
		defaultValue: value,
		rx:           rx,
	}
}

func NewFromEnvironment(name string) *Variable {
	val := os.Getenv(name)
	if val == "" {
		return nil
	}

	return NewVariable(name, val)
}

// perform replaces variable in str to its value
// and returns result string
func (v *Variable) Perform(str string) string {

	res := v.rx.ReplaceAllLiteral(
		[]byte(str),
		[]byte(v.value),
	)

	return string(res)
}
