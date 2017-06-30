package handlers

import (
	"encoding/json"
	"github.com/crawlerclub/x/controller"
	"github.com/crawlerclub/x/types"
	"io/ioutil"
	"net/http"
)

type AddCrawlerHandler struct {
	ctl *controller.Controller
}

func NewAddCrawlerHandler(ctl *controller.Controller) *AddCrawlerHandler {
	return &AddCrawlerHandler{ctl: ctl}
}

func (self *AddCrawlerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		showError(w, r, err.Error(), 400)
		return
	}
	var item types.CrawlerItem
	err = json.Unmarshal(b, &item)
	if err != nil {
		showError(w, r, err.Error(), 400)
		return
	}
	ok, err := item.Conf.IsValid()
	if !ok {
		showError(w, r, err.Error(), 400)
		return
	}
	if item.CrawlerName != item.Conf.CrawlerName {
		showError(w, r, "CrawlerNames are not consistent", 400)
		return
	}
	mustEncode(w, item)
}
