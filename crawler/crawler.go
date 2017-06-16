package crawler

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/crawlerclub/x/downloader"
	"github.com/crawlerclub/x/parser"
	"github.com/crawlerclub/x/types"
	"github.com/tkuchiki/parsetime"
	"golang.org/x/net/context"
	"gopkg.in/olivere/elastic.v5"
	"io/ioutil"
	"time"
)

var (
	ErrEmptyCrawlerConf  = errors.New("empty CrawlerConf")
	ErrNilTask           = errors.New("nil Task")
	ErrNilItem           = errors.New("nil Item")
	ErrInvalidParserName = errors.New("ParserName not in crawler_name:parser_name format")
)

type Crawler struct {
	Conf   *types.CrawlerConf
	client *elastic.Client
}

func (self *Crawler) LoadConfFromBytes(str []byte) error {
	self.Conf = new(types.CrawlerConf)
	err := json.Unmarshal(str, self.Conf)
	if err != nil {
		return err
	}
	ok, err := self.Conf.IsValid()
	if !ok {
		return err
	}
	if len(self.Conf.EsUri) > 0 {
		self.client, err = elastic.NewClient(
			elastic.SetURL(self.Conf.EsUri),
			elastic.SetMaxRetries(10))
		if err != nil {
			return err
		}
	}
	return nil
}

func (self *Crawler) LoadConfFromFile(file string) error {
	str, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}
	err = self.LoadConfFromBytes(str)
	return err
}

func (self *Crawler) GetStartTasks() ([]*types.Task, error) {
	if self.Conf == nil {
		return nil, ErrEmptyCrawlerConf
	}
	ok, err := self.Conf.IsValid()
	if !ok {
		return nil, err
	}
	var tasks []*types.Task
	for _, url := range self.Conf.StartUrls {
		tasks = append(tasks, &types.Task{Url: url, CrawlerName: self.Conf.CrawlerName, ParserName: self.Conf.StartParserName})
	}
	return tasks, nil
}

func (self *Crawler) Process(task *types.Task) ([]types.Task, []map[string]interface{}, error) {
	if task == nil {
		return nil, nil, ErrNilTask
	}
	if self.Conf == nil {
		return nil, nil, ErrEmptyCrawlerConf
	}
	if urlParser, ok := self.Conf.ParseConfs[task.ParserName]; ok {
		uParser := parser.GetParser(urlParser.ParserType)
		if uParser == nil {
			return nil, nil, errors.New(fmt.Sprintf("no parser_type %s found!", urlParser.ParserType))
		}
		req := &types.HttpRequest{Url: task.Url, Platform: "pc", Timeout: 60}
		resp := downloader.Download(req)
		if resp.Error != nil {
			return nil, nil, resp.Error
		}
		//fmt.Println(resp.Text)
		tasks, items, err := uParser.Parse(resp.Text, resp.Url, &urlParser)
		if err != nil {
			return nil, nil, err
		}

		lastModified := time.Now().Unix()
		for _, item := range items {
			if value, ok := item["last_modified_"]; ok {
				p, err := parsetime.NewParseTime()
				if err != nil {
					return nil, nil, err
				}
				t, err := p.Parse(value.(string))
				if err != nil {
					return nil, nil, err
				}
				lastModified = t.Unix()
			}
			break
		}
		if lastModified <= task.LastAccessTime {
			// stop generating tasks, since no new content found
			tasks = nil
		}

		for i, _ := range tasks {
			tasks[i].CrawlerName = self.Conf.CrawlerName // set CrawlerName for tasks
		}
		return tasks, items, err
	} else {
		return nil, nil, errors.New(fmt.Sprintf("No ParseConf for %s", task.ParserName))
	}
}

func (self *Crawler) Save(item map[string]interface{}) error {
	if self.Conf == nil {
		return ErrEmptyCrawlerConf
	}
	if item == nil {
		return ErrNilItem
	}
	if self.client == nil {
		return nil
	}
	if esType, ok := item["es_type"]; ok {
		ctx := context.Background()
		t, _ := json.Marshal(item)
		var err error
		if id, has := item["id"]; has {
			_, err = self.client.Index().Index(self.Conf.CrawlerName).
				Type(esType.(string)).Id(id.(string)).BodyString(string(t)).Do(ctx)
		} else {
			_, err = self.client.Index().Index(self.Conf.CrawlerName).
				Type(esType.(string)).BodyString(string(t)).Do(ctx)
		}
		return err
	}
	return nil
}
