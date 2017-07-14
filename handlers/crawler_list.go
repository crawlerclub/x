package handlers

import (
	"github.com/crawlerclub/x/controller"
	"github.com/crawlerclub/x/store"
	"github.com/crawlerclub/x/types"
	"github.com/syndtr/goleveldb/leveldb/util"
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
	if self.ctl == nil || self.ctl.Stores == nil {
		showError(w, r, "controller is nil", 500)
		return
	}
	//vars := mux.Vars(r)
	r.ParseForm()
	offset, _ := strconv.ParseInt(r.FormValue("offset"), 10, 32)
	limit, _ := strconv.ParseInt(r.FormValue("limit"), 10, 32)
	prefix := r.FormValue("search")
	var filter *util.Range
	if prefix != "" {
		filter = util.BytesPrefix([]byte(prefix))
	}

	if offset < 0 {
		offset = 0
	}
	if limit <= 0 {
		limit = 10
	}
	var items []*types.CrawlerItem
	var count int64
	count = 0
	err := self.ctl.Stores["crawler"].ForEach(filter, func(key, value []byte) (bool, error) {
		if count >= offset && count < offset+limit {
			var item types.CrawlerItem
			e := store.BytesToObject(value, &item)
			if e != nil {
				return false, e
			}
			items = append(items, &item)
		}
		count += 1
		return true, nil
	})

	if err != nil {
		showError(w, r, err.Error(), 500)
		return
	}

	rv := struct {
		Total int64                `json:"total"`
		Rows  []*types.CrawlerItem `json:"rows"`
	}{
		Total: count,
		Rows:  items,
	}
	mustEncode(w, rv)
}
