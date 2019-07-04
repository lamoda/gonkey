package yaml_file

import (
	"io/ioutil"
	"os"
	"strings"

	"github.com/keyclaim/gonkey/models"
	"github.com/keyclaim/gonkey/testloader"
)

type YamlFileLoader struct {
	testloader.LoaderInterface

	testsLocation string
}

func NewLoader(testsLocation string) *YamlFileLoader {
	return &YamlFileLoader{
		testsLocation: testsLocation,
	}
}

func (l *YamlFileLoader) Load() (chan models.TestInterface, error) {
	fileTests, err := parseTestsWithCases(l.testsLocation)
	if err != nil {
		return nil, err
	}
	ch := make(chan models.TestInterface)
	go func() {
		for i := range fileTests {
			ch <- &fileTests[i]
		}
		close(ch)
	}()
	return ch, nil
}

func parseTestsWithCases(path string) ([]Test, error) {
	stat, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	return lookupPath(path, stat)
}

// lookupPath recursively walks over the directory and parses YML files it finds
func lookupPath(path string, fi os.FileInfo) ([]Test, error) {
	if !fi.IsDir() {
		return parseTestDefinitionFile(path)
	}
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}
	var tests []Test
	for _, fi := range files {
		if fi.IsDir() || isYmlFile(fi.Name()) {
			moreTests, err := lookupPath(path+"/"+fi.Name(), fi)
			if err != nil {
				return nil, err
			}
			tests = append(tests, moreTests...)
		}
	}
	return tests, nil
}

func isYmlFile(name string) bool {
	return strings.HasSuffix(name, ".yaml") || strings.HasSuffix(name, ".yml")
}
