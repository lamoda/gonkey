package parser

import (
    "errors"
    "fmt"
    "path/filepath"
    "strings"
)

var (
    ErrFixtureNotFound  = errors.New("fixture not found")
    ErrFixtureFileLoad  = errors.New("failed to load fixture file")
    ErrFixtureParseFile = errors.New("failed to parse fixture file")
    ErrParserNotFound   = errors.New("parser not found")
)

func loadError(fixtureName string, err error) error {
    return fmt.Errorf("%w %s: %s", ErrFixtureFileLoad, fixtureName, err)
}

func parseError(fixtureName string, err error) error {
    return fmt.Errorf("%w %s: %s", ErrFixtureParseFile, fixtureName, err)
}

type fileParser struct {
    locations []string
}

func New(locations []string) *fileParser{
    return &fileParser{
        locations: locations,
    }
}

func (l *fileParser) ParseFiles(ctx *context, names []string) ([]*Fixture, error) {
    var fileNameCache = make(map[string]struct{})
    var fixtures []*Fixture

    for _, name := range names {
        for _, loc := range l.locations {
            filename, err := l.getFirstExistsFileName(name, loc)
            if err != nil {
                return nil, loadError(name, err)
            }
            if _, ok := fileNameCache[filename]; ok {
                continue
            }

            extension := strings.Replace(filepath.Ext(filename), ".", "", -1)
            fixtureParser := GetParser(extension)
            if fixtureParser == nil {
                return nil, ErrParserNotFound
            }
            parserCopy := fixtureParser.Copy(l)

            fixture, err := parserCopy.Parse(ctx, filename)
            if err != nil {
                return nil, parseError(filename, err)
            }

            fixtures = append(fixtures, fixture)
            fileNameCache[filename] = struct{}{}
        }
    }

    return fixtures, nil
}

func (l *fileParser) getFirstExistsFileName(name string, location string) (string, error) {
    candidates := []string{
        name,
        fmt.Sprintf("%s.yaml", name),
        fmt.Sprintf("%s.yml", name),
    }

    for _, p := range candidates {
        path := filepath.Join(location, p)
        paths, err := filepath.Glob(path)
        if err != nil {
            return "", err
        }
        if len(paths) > 0 {
            return paths[0], nil
        }
    }

    return "", ErrFixtureNotFound
}
