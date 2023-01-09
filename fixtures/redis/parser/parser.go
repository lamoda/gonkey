package parser

type FixtureFileParser interface {
    Parse(ctx *context, filename string) (*Fixture, error)
    Copy(parser *fileParser) FixtureFileParser
}

var fixtureParsersRegistry = make(map[string]FixtureFileParser)

func RegisterParser(format string, parser FixtureFileParser) {
    fixtureParsersRegistry[format] = parser
}

func GetParser(format string) FixtureFileParser {
    return fixtureParsersRegistry[format]
}

func init() {
    RegisterParser("yaml", &redisYamlParser{})
}

