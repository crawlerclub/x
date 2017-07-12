package store

import (
	"github.com/crawlerclub/x/types"
	"testing"
)

func TestTaskDB(t *testing.T) {
	e, err := NewTaskDB("sqlite3", "task.sqlite3")
	if err != nil {
		t.Error(err)
	}
	defer e.Close()
	item := types.Task{Url: "url", CrawlerName: "crawler", ParserName: "parser"}
	id, err := e.Insert(&item)
	if err != nil {
		t.Error(err)
	}
	t.Log(id)
}
