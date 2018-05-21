package parser

var Parsers = make(map[string]Parser)

func GetParser(name string) Parser {
	return Parsers[name]
}
