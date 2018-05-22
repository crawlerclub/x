package parser

import (
	"errors"
	"github.com/crawlerclub/x/types"
	t "github.com/lestrrat/go-libxml2/types"
	"github.com/robertkrimen/otto"
	"strings"
	"time"
)

func init() {
	Parsers["html"] = HtmlParser{"html parser"}
}

type HtmlParser struct {
	Name string
}

type DOMNode struct {
	Name string // always start with name root
	Node interface{}
	Item map[string]interface{}
}

var (
	ErrEmptyXpath      = errors.New("empty xpath of node conf")
	ErrInvalidRuleType = errors.New("invalid rule_type of node conf")
	ErrEmptyRuleType   = errors.New("empty rule_type of node conf")
	ErrEmptyItemKey    = errors.New("empty item_key of node conf")
)

func (parser HtmlParser) String() string {
	return parser.Name
}

func (parser HtmlParser) parseNodeByRule(
	node interface{},
	rule types.ParseRule,
	pageUrl string) ([]interface{}, error) {
	if len(rule.RuleType) == 0 {
		return nil, ErrEmptyRuleType
	}
	if len(rule.Xpath) == 0 {
		return nil, ErrEmptyXpath
	}
	var ret []interface{}
	nodes, err := node.(t.Node).Find(rule.Xpath)
	// zliu
	defer nodes.Free()
	if err != nil {
		return nil, err
	}
	for _, domNode := range nodes.NodeList() {
		switch rule.RuleType {
		case "dom":
			ret = append(ret, interface{}(domNode))
		case "url":
			u, _ := MakeAbsoluteUrl(domNode.TextContent(), pageUrl)
			ret = append(ret, interface{}(u))
		case "string":
			ret = append(ret, interface{}(strings.TrimSpace(domNode.TextContent())))
		case "html":
			ret = append(ret, interface{}(domNode.String()))
		}
	}

	if len(rule.Regex) > 0 {
		var tmpVals []interface{}
		switch rule.RuleType {
		case "string":
			for _, v := range ret {
				res, err := ParseRegex(v.(string), rule.Regex)
				if err != nil {
					return nil, err
				}
				for _, tmpRes := range res {
					tmpVals = append(tmpVals, interface{}(tmpRes))
				}
			}
			ret = tmpVals
		case "url":
			for _, v := range ret {
				// only keep matched urls
				if MatchRegex(v.(string), rule.Regex) {
					tmpVals = append(tmpVals, v)
				}
			}
			ret = tmpVals
		}
	} // if has regex

	if len(rule.Js) > 0 {
		runtime := otto.New()
		if _, err := runtime.Run(rule.Js); err != nil {
			return nil, err
		}
		var newVals []interface{}
		for _, v := range ret {
			jsVal, err := runtime.ToValue(v)
			if err != nil {
				return nil, err
			}
			result, err := runtime.Call("process", nil, jsVal)
			if err != nil {
				return nil, err
			}
			s, err := result.Export()
			if err != nil {
				return nil, err
			}
			newVals = append(newVals, s)
		}
		ret = newVals
	} // if has js
	return ret, err
}

func (parser HtmlParser) parseNode(
	node interface{},
	rules []types.ParseRule,
	pageUrl string) ([]*DOMNode, []types.Task, map[string]interface{}, error) {
	var retDOMs []*DOMNode
	var retUrls []types.Task
	retItems := make(map[string]interface{})
	// we may get different items from one node, so need multiple rules
	for _, rule := range rules {
		if len(rule.ItemKey) == 0 {
			return nil, nil, nil, ErrEmptyItemKey
		}
		vals, err := parser.parseNodeByRule(node, rule, pageUrl)
		if err != nil {
			return nil, nil, nil, err
		}
		if rule.RuleType == "dom" {
			for _, v := range vals {
				retDOMs = append(retDOMs, &DOMNode{
					Name: rule.ItemKey,
					Node: interface{}(v),
					Item: retItems,
				})
			}
		} else {
			if rule.RuleType == "url" {
				for _, v := range vals {
					if u, ok := v.(string); ok {
						retUrls = append(retUrls, types.Task{
							ParserName: rule.ItemKey,
							Url:        u,
							IsSeedUrl:  rule.IsSeedUrl, // 20170616
						})
					}
				}
			}

			// string or url, treat tasks as items too
			if _, ok := retItems[rule.ItemKey]; !ok {
				if len(vals) == 1 {
					retItems[rule.ItemKey] = vals[0]
				} else if len(vals) > 1 {
					retItems[rule.ItemKey] = interface{}(vals)
				}
			} else {
				switch retItems[rule.ItemKey].(type) {
				case []interface{}:
					retItems[rule.ItemKey] = append(retItems[rule.ItemKey].([]interface{}), vals...)
				default:
					retItems[rule.ItemKey] = append([]interface{}{retItems[rule.ItemKey]}, vals...)
				}
				retItems[rule.ItemKey] = interface{}(retItems[rule.ItemKey])
			}
		}
	}
	return retDOMs, retUrls, retItems, nil
}

func (parser HtmlParser) Parse(
	page, pageUrl string,
	parseConf *types.ParseConf) ([]types.Task, []map[string]interface{}, error) {
	if parseConf == nil {
		return nil, nil, errors.New("parse conf is nil")
	}
	if len(page) == 0 {
		return nil, nil, errors.New("page len is 0")
	}
	conf := parseConf.Rules

	root, err := ParseHTMLString(page, "utf-8")
	if err != nil {
		return nil, nil, err
	}
	defer root.Free()

	var domList []*DOMNode
	rootNode := &DOMNode{Name: "root", Node: interface{}(root), Item: make(map[string]interface{})}
	domList = append(domList, rootNode)

	var retUrls []types.Task
	var retItems []map[string]interface{}
	for {
		if len(domList) == 0 { // no more dom to be processed
			break
		}
		domName := domList[0].Name
		domNode := domList[0].Node
		parentItems := domList[0].Item
		domList = domList[1:]

		domNode.(t.Node).MakeMortal()
		//defer domNode.(t.Node).AutoFree()

		var rules []types.ParseRule // get parse_rule via domName
		var ok bool
		if rules, ok = conf[domName]; !ok {
			continue // no conf for this dom
		}
		DOMNodes, urlList, item, err := parser.parseNode(domNode, rules, pageUrl)
		if err != nil {
			return nil, nil, err
		}
		domList = append(domList, DOMNodes...) // add more doms to be processed
		if urlList != nil {
			retUrls = append(retUrls, urlList...)
		}
		if item != nil {
			if _, ok = parentItems[domName]; !ok {
				parentItems[domName] = interface{}(item)
			} else {
				switch parentItems[domName].(type) {
				case []interface{}:
					parentItems[domName] = append(parentItems[domName].([]interface{}), interface{}(item))
				default:
					parentItems[domName] = []interface{}{parentItems[domName], interface{}(item)}
				}
				parentItems[domName] = interface{}(parentItems[domName])
			}
		}
	} // end for
	if rootItems, ok := rootNode.Item["root"]; ok {
		switch rootItems.(type) {
		case []interface{}:
			tmp, _ := rootItems.([]map[string]interface{})
			retItems = append(retItems, tmp...)
			/*
				for _, v := range tmp {
					retItems = append(retItems, types.Item(v))
				}
			*/
		default:
			retItems = append(retItems, rootItems.(map[string]interface{}))
		}
	}
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
