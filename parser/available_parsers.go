package parser

func GetParser(name string) Parser {
	if parser, ok := parsers[name]; ok {
		return parser
	}
	return nil
}

var parsers = map[string]Parser{
	"html": HtmlParser{"html parser"},
	"json": JsonParser{"json parser"},
}
