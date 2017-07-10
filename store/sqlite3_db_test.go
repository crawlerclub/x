package store

import (
	"encoding/json"
	"github.com/crawlerclub/x/types"
	"io/ioutil"
	"testing"
)

func TestEngine(t *testing.T) {
	e, err := NewEngine("sqlite3", "db.sqlite3")
	if err != nil {
		t.Error(err)
	}
	var item types.CrawlerItem
	bytes, err := ioutil.ReadFile("../test/http-files/editor/default.json")
	if err != nil {
		t.Error(err)
	}
	err = json.Unmarshal(bytes, &item)
	if err != nil {
		t.Error(err)
	}
	err = e.CreateTables()
	if err != nil {
		t.Error(err)
	}
	/*
		err = e.CrawlerInsert(&item)
		if err != nil {
			t.Error(err)
		}
	*/
	i, err := e.CrawlerSelect(1)
	if err != nil {
		t.Error(err)
	}
	t.Log(i)
}
