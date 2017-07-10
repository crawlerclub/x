package handlers

import (
	"github.com/crawlerclub/x/controller"
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
	if self.ctl == nil || self.ctl.CrawlerStore == nil {
		showError(w, r, "controller is nil", 500)
		return
	}
	//vars := mux.Vars(r)
	r.ParseForm()
	offset, _ := strconv.ParseInt(r.FormValue("offset"), 10, 32)
	limit, _ := strconv.ParseInt(r.FormValue("limit"), 10, 32)

	count, err := self.ctl.CrawlerStore.Count("")
	if err != nil {
		showError(w, r, err.Error(), 500)
		return
	}

	if offset < 0 || offset >= count {
		offset = 0
	}
	if limit <= 0 {
		limit = 10
	}

	items, err := self.ctl.CrawlerStore.List("LIMIT ? OFFSET ?", limit, offset)
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
