package handlers

import (
	"encoding/json"
	"github.com/crawlerclub/x/controller"
	"github.com/crawlerclub/x/types"
    "github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
)

type CrudCrawlerHandler struct {
	ctl *controller.Controller
}

func NewCrudCrawlerHandler(ctl *controller.Controller) *CrudCrawlerHandler {
	return &CrudCrawlerHandler{ctl: ctl}
}

func (self *CrudCrawlerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    switch vars["action"] {
    case "create":
        //
    case "retrieve":
        //
    case "update":
        //
    case "delete":
        //
    default:
        showError(w, r, "unknown action", 400)
        return
    }
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		showError(w, r, err.Error(), 400)
		return
	}
    mustEncode(w, vars)
    return
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
