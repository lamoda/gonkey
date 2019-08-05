package yaml_file

import (
	"io/ioutil"
	"os"
	"strings"

	"github.com/lamoda/gonkey/models"
	"github.com/lamoda/gonkey/testloader"
)

type YamlFileLoader struct {
	testloader.LoaderInterface

	testsLocation string
	fileFilter    string
}

func NewLoader(testsLocation string) *YamlFileLoader {
	return &YamlFileLoader{
		testsLocation: testsLocation,
	}
}

func (l *YamlFileLoader) Load() (chan models.TestInterface, error) {
	fileTests, err := l.parseTestsWithCases(l.testsLocation)
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

func (l *YamlFileLoader) SetFileFilter(f string) {
	l.fileFilter = f
}

func (l *YamlFileLoader) parseTestsWithCases(path string) ([]Test, error) {
	stat, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	return l.lookupPath(path, stat)
}

// lookupPath recursively walks over the directory and parses YML files it finds
func (l *YamlFileLoader) lookupPath(path string, fi os.FileInfo) ([]Test, error) {
	if !fi.IsDir() {
		if !l.fitsFilter(path) {
			return []Test{}, nil
		}
		return parseTestDefinitionFile(path)
	}
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}
	var tests []Test
	for _, fi := range files {
		if !fi.IsDir() && !isYmlFile(fi.Name()) {
			continue
		}
		moreTests, err := l.lookupPath(path+"/"+fi.Name(), fi)
		if err != nil {
			return nil, err
		}
		tests = append(tests, moreTests...)
	}
	return tests, nil
}

func (l *YamlFileLoader) fitsFilter(fileName string) bool {
	if l.fileFilter == "" {
		return true
	}
	return strings.Contains(fileName, l.fileFilter)
}

func isYmlFile(name string) bool {
	return strings.HasSuffix(name, ".yaml") || strings.HasSuffix(name, ".yml")
}
