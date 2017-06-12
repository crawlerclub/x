package parser

import "github.com/crawlerclub/x/types"

type Parser interface {
	String() string
	Parse(page, pageUrl string, parseConf *types.ParseConf) ([]types.Task, []map[string]interface{}, error)
}
