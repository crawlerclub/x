package types

import (
	"errors"
	"fmt"
)

var (
	ErrEmptyCrawlerName       = errors.New("types/types.go empty crawler_name of crawler conf")
	ErrUnSupportedCrawlerType = errors.New("types/types.go unsupported crawler_type of crawler conf")
	ErrEmptyStartUrls         = errors.New("types/types.go empty start_urls of crawler conf")
	ErrEmptyUrlsFile          = errors.New("types/types.go empty urls_file of crawler conf")
	ErrNoStartRule            = errors.New("types/types.go empty start task conf rule of crawler conf")
)

type ParseRule struct {
	// four RuleTypes: url, dom, string, html
	RuleType string `json:"rule_type" bson:"rule_type"`

	// when RuleType is dom, ItemKey stores the next RuleName
	ItemKey string `json:"item_key" bson:"item_key"`
	// IsSeedUrl indicates whether the generated item is a seed or not
	IsSeedUrl bool   `json:"is_seed_url" bson:"is_seed_url"`
	Xpath     string `json:"xpath" bson:"xpath"`
	Regex     string `json:"regex" bson:"regex"`
	Js        string `json:"js" bson:"js"`
}

type ParseConf struct {
	ParserType      string                 `json:"parser_type" bson:"parser_type"`
	ParserName      string                 `json:"parser_name" bson:"parser_name"`
	NoDefaultFields bool                   `json:"no_default_fields" bson:"no_default_fields"`
	ExampleUrl      string                 `json:"example_url" bson:"example_url"`
	Rules           map[string][]ParseRule `json:"rules" bson:"rules"` // RuleName to ParseRules
	PostProcessor   string                 `json:"post_processor" bson:"post_processor"`
	RevisitInterval int64                  `json:"revisit_interval" bson:"revisit_interval"`
}

func (this *ParseConf) String() string {
	return fmt.Sprintf("{ParserType:%s, ParserName:%s, RevisitInterval:%d}",
		this.ParserType, this.ParserName, this.RevisitInterval)
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

func (self *CrawlerConf) Type() string {
	return "crawler"
}

func (self *CrawlerConf) Id() string {
	return self.CrawlerName
}

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
	RemoteAddr string
	Error      error
}
