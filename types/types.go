package types

import (
	"errors"
)

type ParseRule struct {
	// four RuleTypes: url, dom, string, html
	RuleType string `json:"rule_type" bson:"rule_type"`

	// when RuleType is dom, ItemKey stores the next RuleName
	ItemKey string `json:"item_key" bson:"item_key"`
	Xpath   string `json:"xpath" bson:"xpath"`
	Regex   string `json:"regex" bson:"regex"`
	Js      string `json:"js" bson:"js"`
}

type ParseConf struct {
	ParserType      string                 `json:"parser_type" bson:"parser_type"`
	ParserName      string                 `json:"parser_name" bson:"parser_name"`
	NoDefaultFields bool                   `json:"no_default_fields" bson:"no_default_fields"`
	ExampleUrl      string                 `json:"example_url" bson:"example_url"`
	Rules           map[string][]ParseRule `json:"rules" bson:"rules"` // RuleName to ParseRules
	PostProcessor   string                 `json:"post_processor" bson:"post_processor"`
}

type CrawlerConf struct {
	CrawlerType     string               `json:"crawler_type" bson:"crawler_type"`
	CrawlerName     string               `json:"crawler_name" bson:"crawler_name"`
	CrawlerDesp     string               `json:"crawler_desp" bson:"crawler_desp"`
	StartUrls       []string             `json:"start_urls" bson:"start_urls"`
	UrlsFile        string               `json:"urls_file" bson:"urls_file"`
	ParseConfs      map[string]ParseConf `json:"parse_confs" bson:"parse_confs"`
	StartParserName string               `json:"start_parser_name" bson:"start_parser_name"`
	EsUri           string               `json:"es_uri" bson:"es_uri"`
}

var (
	ErrEmptyCrawlerName       = errors.New("empty crawler_name of crawler conf")
	ErrUnSupportedCrawlerType = errors.New("unsupported crawler_type of crawler conf")
	ErrEmptyStartUrls         = errors.New("empty start_urls of crawler conf")
	ErrEmptyUrlsFile          = errors.New("empty urls_file of crawler conf")
	ErrNoStartRule            = errors.New("empty start task conf rule of crawler conf")
)

func (conf *CrawlerConf) IsValid() (bool, error) {
	if conf.CrawlerName == "" {
		return false, ErrEmptyCrawlerName
	}
	switch conf.CrawlerType {
	case "navigation":
		if conf.StartUrls == nil || len(conf.StartUrls) == 0 {
			return false, ErrEmptyStartUrls
		}
	case "url_set":
		if conf.UrlsFile == "" {
			return false, ErrEmptyUrlsFile
		}
	default:
		return false, ErrUnSupportedCrawlerType
	}
	if _, ok := conf.ParseConfs[conf.StartParserName]; !ok {
		return false, ErrNoStartRule
	}
	return true, nil
}

//type Item map[string]interface{}

type Task struct {
	// crawler_name:parser_name
	ParserName string `json:"parser_name" bson:"parser_name"`
	Url        string `json:"url" bson:"url"`
	Data       string `json:"data" bson:"data"`
}

type HttpRequest struct {
	Url      string
	Method   string
	PostData string
	UseProxy bool
	Proxy    string
	Timeout  int
	MaxLen   int64
	Platform string
}

type HttpResponse struct {
	Url        string
	Text       string
	Content    []byte
	Encoding   string
	StatusCode int
	Proxy      string
	Cookies    map[string]string
	Error      error
}
