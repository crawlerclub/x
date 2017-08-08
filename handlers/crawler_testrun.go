package handlers

import (
	"github.com/crawlerclub/x/controller"
	"github.com/crawlerclub/x/crawler"
	"github.com/crawlerclub/x/types"
	"github.com/gorilla/mux"
	"github.com/liuzl/store"
	"net/http"
)

type TestHandler struct {
	ctl *controller.Controller
}

func NewTestHandler(ctl *controller.Controller) *TestHandler {
	return &TestHandler{ctl: ctl}
}

func (self *TestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if self.ctl == nil || self.ctl.Stores == nil {
		showError(w, r, "controller is nil", 500)
		return
	}
	vars := mux.Vars(r)
	bytes, err := self.ctl.Stores["crawler"].Get(vars["name"])
	if err != nil {
		showError(w, r, err.Error(), 500)
		return
	}

	var item types.CrawlerItem
	err = store.BytesToObject(bytes, &item)
	if err != nil {
		showError(w, r, err.Error(), 500)
		return
	}

	worker := crawler.Crawler{Conf: &item.Conf}
	ret, err := worker.Test()
	if err != nil {
		showError(w, r, err.Error(), 500)
		return
	}

	mustEncode(w, ret)
}
