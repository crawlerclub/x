package handlers

import (
	"github.com/crawlerclub/x/controller"
	"github.com/crawlerclub/x/store"
	"github.com/crawlerclub/x/types"
	"net/http"
	"strconv"
)

type ListCrawlerHandler struct {
	ctl *controller.Controller
}

func NewListCrawlerHandler(ctl *controller.Controller) *ListCrawlerHandler {
	return &ListCrawlerHandler{ctl: ctl}
}

func (self *ListCrawlerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if self.ctl == nil || self.ctl.CrawlerDB == nil {
		showError(w, r, "controller is nil", 500)
		return
	}
	//vars := mux.Vars(r)
	r.ParseForm()
	offset, _ := strconv.ParseInt(r.FormValue("offset"), 10, 32)
	limit, _ := strconv.ParseInt(r.FormValue("limit"), 10, 32)
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 {
		limit = 10
	}
	iter := self.ctl.CrawlerDB.NewIterator(nil, nil)
	var items []interface{}
	i := int64(-1)
	cnt := int64(0)
	var item types.CrawlerItem
	for iter.Next() {
		i += 1
		if i < offset {
			continue
		}
		err := store.BytesToObject(iter.Value(), &item)
		if err != nil {
			showError(w, r, err.Error(), 500)
			return
		}
		items = append(items, item)
		cnt += 1
		if cnt >= limit {
			break
		}
	}
	rv := struct {
		Total int           `json:"total"`
		Rows  []interface{} `json:"rows"`
	}{
		Total: 100,
		Rows:  items,
	}
	mustEncode(w, rv)
}
