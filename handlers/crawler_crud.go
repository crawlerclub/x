package handlers

import (
	"encoding/json"
	"errors"
	"github.com/crawlerclub/x/controller"
	"github.com/crawlerclub/x/types"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
)

var (
	ErrNameExists = errors.New("CrawlerName already exists")
)

type CrudCrawlerHandler struct {
	ctl *controller.Controller
}

func NewCrudCrawlerHandler(ctl *controller.Controller) *CrudCrawlerHandler {
	return &CrudCrawlerHandler{ctl: ctl}
}

func (self *CrudCrawlerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if self.ctl == nil || self.ctl.CrawlerStore == nil {
		showError(w, r, "controller is nil", 500)
		return
	}
	vars := mux.Vars(r)
	switch vars["action"] {
	case "create":
		err := self.saveCrawlerItem(r, true)
		if err != nil {
			showError(w, r, err.Error(), 400)
			return
		}
		ok(w)
	case "retrieve":
		item, err := self.ctl.CrawlerStore.SelectByName(vars["name"])
		if err == nil {
			mustEncode(w, item)
		} else {
			showError(w, r, err.Error(), 500)
		}
		return
	case "update":
		err := self.saveCrawlerItem(r, false)
		if err != nil {
			showError(w, r, err.Error(), 400)
			return
		}
		ok(w)
	case "delete":
		err := self.ctl.DelCrawler(vars["name"])
		if err != nil {
			showError(w, r, err.Error(), 500)
		} else {
			ok(w)
		}
	default:
		showError(w, r, "unknown action", 400)
		return
	}
}

func (self *CrudCrawlerHandler) saveCrawlerItem(r *http.Request, isNew bool) error {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	var item types.CrawlerItem
	err = json.Unmarshal(b, &item)
	if err != nil {
		return err
	}
	return self.ctl.AddCrawler(&item, isNew)
}
