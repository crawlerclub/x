package types

import (
	"fmt"
)

type StoreItem interface {
	Id() string
}

type Task struct {
	Id              int    `json:"id" bson:"id"`
	CrawlerName     string `json:"crawler_name" bson:"crawler_name"`
	ParserName      string `json:"parser_name" bson:"parser_name"`
	IsSeedUrl       bool   `json:"is_seed_url" bson:"is_seed_url"`
	Url             string `json:"url" bson:"url"`
	Data            string `json:"data" bson:"data"`
	LastAccessTime  int64  `json:"last_access_time" bson:"last_access_time"`
	RevisitInterval int64  `json:"revisit_interval" bson:"revisit_interval"`
}

func (this *Task) String() string {
	return fmt.Sprintf("{CrawlerName:%s, ParserName:%s, Url:%s, LastAccessTime:%d}",
		this.CrawlerName, this.ParserName, this.Url, this.LastAccessTime)
}

type CrawlerItem struct {
	Id          int         `json:"id" bson:"id"`
	CrawlerName string      `json:"crawler_name" bson:"crawler_name"`
	Conf        CrawlerConf `json:"conf" bson:"conf"`
	Weight      int         `json:"weight" bson:"weight"`
	Status      string      `json:"status" bson:"status"`
	CreateTime  int64       `json:"create_time" bson:"create_time"`
	ModifyTime  int64       `json:"modify_time" bson:"modify_time"`
	Author      string      `json:"author" bson:"author"`
}
