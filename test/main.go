package main

import (
	"flag"
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
	router.Handle("/crawler/{action:create|retrieve|update|delete}/{name}", crudHandler)
	listHandler := handlers.NewListCrawlerHandler(&ctl)
	router.Handle("/list/crawler", listHandler)
	http.Handle("/", router)
	log.Fatal(http.ListenAndServe("0.0.0.0:8888", nil))
}
