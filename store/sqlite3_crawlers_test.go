package store

import (
	"encoding/json"
	"github.com/crawlerclub/x/types"
	"io/ioutil"
	"testing"
)

func TestCrawlerDB(t *testing.T) {
	e, err := NewCrawlerDB("sqlite3", "db.sqlite3")
	if err != nil {
		t.Error(err)
	}
	defer e.Close()
	var item types.CrawlerItem
	bytes, err := ioutil.ReadFile("../test/http-files/editor/default.json")
	if err != nil {
		t.Error(err)
	}
	err = json.Unmarshal(bytes, &item)
	if err != nil {
		t.Error(err)
	}
	err = e.Insert(&item)
	if err != nil {
		t.Error(err)
	}
	i, err := e.Select(1)
	if err != nil {
		t.Error(err)
	}
	t.Log(i)
	items, err := e.List("WHERE author=? limit 1 offset 5", "Zhanliang Liu")
	if err != nil {
		t.Error(err)
	}
	for _, j := range items {
		t.Log(j)
	}

	count, err := e.Count("")
	if err != nil {
		t.Error(err)
	}
	t.Log(count)
}
