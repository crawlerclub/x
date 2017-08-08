package handlers

import (
	"github.com/crawlerclub/x/controller"
	"github.com/crawlerclub/x/types"
	"github.com/gorilla/mux"
	"github.com/liuzl/store"
	"github.com/syndtr/goleveldb/leveldb/util"
	"net/http"
	"strconv"
)

type ListHandler struct {
	ctl *controller.Controller
}

func NewListHandler(ctl *controller.Controller) *ListHandler {
	return &ListHandler{ctl: ctl}
}

func (self *ListHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if self.ctl == nil || self.ctl.Stores == nil {
		showError(w, r, "controller is nil", 500)
		return
	}
	vars := mux.Vars(r)
	s := vars["type"]
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
	var items []interface{}
	var count int64
	count = 0
	err := self.ctl.Stores[s].ForEach(filter, func(key, value []byte) (bool, error) {
		if count >= offset && count < offset+limit {
			if s == "crawler" {
				var item types.CrawlerItem
				e := store.BytesToObject(value, &item)
				if e != nil {
					return false, e
				}
				items = append(items, &item)
			} else {
				var item types.Task
				e := store.BytesToObject(value, &item)
				if e != nil {
					return false, e
				}
				items = append(items, &item)
			}
		}
		count += 1
		return true, nil
	})

	if err != nil {
		showError(w, r, err.Error(), 500)
		return
	}

	rv := struct {
		Total int64         `json:"total"`
		Rows  []interface{} `json:"rows"`
	}{
		Total: count,
		Rows:  items,
	}
	mustEncode(w, rv)
}
