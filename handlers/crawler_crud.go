package handlers

import (
	"encoding/json"
	"errors"
	"github.com/crawlerclub/x/controller"
	"github.com/crawlerclub/x/store"
	"github.com/crawlerclub/x/types"
	"github.com/gorilla/mux"
	"github.com/syndtr/goleveldb/leveldb"
	"io/ioutil"
	"net/http"
)

var (
	ErrNamesNotConsistent = errors.New("CrawlerNames not consistent")
	ErrNameExists         = errors.New("CrawlerName already exists")
)

type CrudCrawlerHandler struct {
	ctl *controller.Controller
}

func NewCrudCrawlerHandler(ctl *controller.Controller) *CrudCrawlerHandler {
	return &CrudCrawlerHandler{ctl: ctl}
}

func (self *CrudCrawlerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if self.ctl == nil || self.ctl.CrawlerDB == nil {
		showError(w, r, "controller is nil", 500)
		return
	}
	vars := mux.Vars(r)
	switch vars["action"] {
	case "create":
		has, err := self.ctl.CrawlerDB.Has([]byte(vars["name"]), nil)
		if err != nil {
			showError(w, r, err.Error(), 500)
			return
		}
		if has {
			showError(w, r, ErrNameExists.Error(), 400)
			return
		}
		err = self.saveCrawlerItem(r, vars["name"])
		if err != nil {
			showError(w, r, err.Error(), 400)
			return
		}
		ok(w)
	case "retrieve":
		data, err := self.ctl.CrawlerDB.Get([]byte(vars["name"]), nil)
		if err == nil {
			var item types.CrawlerItem
			err = store.BytesToObject(data, &item)
			if err != nil {
				showError(w, r, err.Error(), 500)
			} else {
				mustEncode(w, item)
			}
		} else if err == leveldb.ErrNotFound {
			showError(w, r, err.Error(), 404)
		} else {
			showError(w, r, err.Error(), 500)
		}
		return
	case "update":
		err := self.saveCrawlerItem(r, vars["name"])
		if err != nil {
			showError(w, r, err.Error(), 400)
			return
		}
		ok(w)
	case "delete":
		err := self.ctl.CrawlerDB.Delete([]byte(vars["name"]), nil)
		if err != nil {
			showError(w, r, err.Error(), 500)
		} else {
		}
	default:
		showError(w, r, "unknown action", 400)
		return
	}
}

func (self *CrudCrawlerHandler) saveCrawlerItem(r *http.Request, name string) error {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	var item types.CrawlerItem
	err = json.Unmarshal(b, &item)
	if err != nil {
		return err
	}
	ok, err := item.Conf.IsValid()
	if !ok {
		return err
	}
	if item.CrawlerName != item.Conf.CrawlerName || item.CrawlerName != name {
		return ErrNamesNotConsistent
	}
	data, err := store.ObjectToBytes(item)
	if err != nil {
		return err
	}
	return self.ctl.CrawlerDB.Put([]byte(name), data, nil)
}
