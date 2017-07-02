package handlers

import (
	"github.com/crawlerclub/x/controller"
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
	start, _ := strconv.ParseInt(r.FormValue("start"), 10, 32)
	limit, _ := strconv.ParseInt(r.FormValue("limit"), 10, 32)
	if start < 0 {
		start = 0
	}
	if limit <= 0 {
		limit = 10
	}
	iter := self.ctl.CrawlerDB.NewIterator(nil, nil)
	var items []interface{}
	items = append(items, "hello")
	i := int64(-1)
	cnt := int64(0)
	for iter.Next() {
		i += 1
		if i < start {
			continue
		}
		items = append(items, iter.Value)
		cnt += 1
		if cnt >= limit {
			break
		}
	}
	mustEncode(w, items)
}
