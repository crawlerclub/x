package main

import (
	"flag"
	"github.com/GeertJohan/go.rice"
	"github.com/crawlerclub/x/controller"
	"github.com/crawlerclub/x/handlers"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

func main() {
	flag.Parse()

	var ctl controller.Controller
	ctl.Init("./run", 10)

	router := mux.NewRouter()
	crudHandler := handlers.NewCrudCrawlerHandler(&ctl)
	router.Handle("/api/crawler/{action:create|retrieve|update|delete}/{name}", crudHandler)
	listHandler := handlers.NewListCrawlerHandler(&ctl)
	router.Handle("/api/list/crawler", listHandler)
	http.Handle("/api/", router)
	http.Handle("/", http.FileServer(rice.MustFindBox("http-files").HTTPBox()))
	log.Fatal(http.ListenAndServe("0.0.0.0:8080", nil))
}
