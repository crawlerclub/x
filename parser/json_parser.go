package parser

import (
	"encoding/json"
	"errors"
	"github.com/crawlerclub/x/types"
	"github.com/robertkrimen/otto"
	"time"
)

func init() {
	Parsers["json"] = JsonParser{"json parser"}
}

type JsonParser struct {
	Name string
}

func (parser JsonParser) String() string {
	return parser.Name
}

func (parser JsonParser) Parse(page, pageUrl string, parseConf *types.ParseConf) ([]types.Task, []map[string]interface{}, error) {
	if parseConf == nil {
		return nil, nil, errors.New("parse conf is nil")
	}
	if len(page) == 0 {
		return nil, nil, errors.New("page len is 0")
	}
	var retUrls []types.Task
	var retItems []map[string]interface{}
	item := make(map[string]interface{})
	err := json.Unmarshal([]byte(page), &item)
	if err != nil {
		return nil, nil, err
	}
	retItems = append(retItems, item)

	if !parseConf.NoDefaultFields {
		for _, v := range retItems {
			v["from_url_"] = pageUrl
			v["from_parser_name_"] = parseConf.ParserName
			v["crawl_time_"] = time.Now().Format("2006-01-02 15:04:05")
		}
	}

	if len(parseConf.PostProcessor) > 0 && len(retItems) > 0 {
		runtime := otto.New()
		if _, err := runtime.Run(parseConf.PostProcessor); err != nil {
			return nil, nil, err
		}

		jsVal, err := runtime.ToValue(retItems)
		result, err := runtime.Call("process", nil, jsVal)
		if err != nil {
			return nil, nil, err
		}

		s, err := result.Export()
		if err != nil {
			return nil, nil, err
		}
		value, ok := s.([]map[string]interface{})
		if ok && len(value) > 0 {
			retItems = value
		}
	}
	return retUrls, retItems, nil
}
