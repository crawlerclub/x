package handlers

import (
	"encoding/json"
	"github.com/golang/glog"
	"net/http"
)

func showError(w http.ResponseWriter, r *http.Request, msg string, code int) {
	glog.Error("Reporting error ", code, "/", msg)
	http.Error(w, msg, code)
}

func mustEncode(w http.ResponseWriter, i interface{}) {
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Content-type", "application/json")
	e := json.NewEncoder(w)
	if err := e.Encode(i); err != nil {
		panic(err)
	}
}

type varLookupFunc func(req *http.Request) string
