package crawler

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/crawlerclub/x/downloader"
	"github.com/crawlerclub/x/ds"
	"github.com/crawlerclub/x/parser"
	"github.com/crawlerclub/x/types"
	"github.com/tkuchiki/parsetime"
	"golang.org/x/net/context"
	"gopkg.in/olivere/elastic.v5"
	"io/ioutil"
	"time"
)

var (
	ErrEmptyCrawlerConf = errors.New("crawler/crawler.go empty CrawlerConf")
	ErrNilTask          = errors.New("crawler/crawler.go nil Task")
	ErrNilItem          = errors.New("crawler/crawler.go nil Item")
	ErrNilTaskQueue     = errors.New("crawler/crawler.go nil TaskQueue")
)

type Crawler struct {
	Conf      *types.CrawlerConf
	TaskQueue *ds.Queue
	es        *elastic.Client
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
	return self.InitEs()
}

func (self *Crawler) InitEs() error {
	if self.Conf == nil {
		return errors.New("CrawlerConf is nil")
	}
	var err error
	if len(self.Conf.EsUri) > 0 {
		self.es, err = elastic.NewClient(
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

func (self *Crawler) InitTaskQueue(dir string) error {
	if self.Conf == nil {
		return ErrEmptyCrawlerConf
	}
	if ok, err := self.Conf.IsValid(); !ok {
		return err
	}
	taskDir := dir + "/" + self.Conf.CrawlerName
	var err error
	self.TaskQueue, err = ds.OpenQueue(taskDir)
	if err != nil {
		return err
	}
	return nil
}

func (self *Crawler) Close() {
	if self.TaskQueue != nil {
		self.TaskQueue.Close()
	}
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
	if self.es == nil {
		return nil
	}
	if esType, ok := item["es_type"]; ok {
		ctx := context.Background()
		t, _ := json.Marshal(item)
		var err error
		if id, has := item["id"]; has {
			_, err = self.es.Index().Index(self.Conf.CrawlerName).
				Type(esType.(string)).Id(id.(string)).BodyString(string(t)).Do(ctx)
		} else {
			_, err = self.es.Index().Index(self.Conf.CrawlerName).
				Type(esType.(string)).BodyString(string(t)).Do(ctx)
		}
		return err
	}
	return nil
}
