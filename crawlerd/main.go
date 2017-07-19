package main

import (
	"flag"
	"github.com/GeertJohan/go.rice"
	"github.com/crawlerclub/x/controller"
	"github.com/crawlerclub/x/handlers"
	"github.com/golang/glog"
	"github.com/gorilla/mux"
	_ "github.com/mkevac/debugcharts"
	"net/http"
	_ "net/http/pprof"
)

var (
	workerCnt  = flag.Int("wc", 1, "crawler worker count")
	workingDir = flag.String("dir", "./run", "working dir")
	serverAddr = flag.String("addr", ":8080", "bind address")
)

func Web(ctl *controller.Controller, addr string) {
	router := mux.NewRouter()
	crudHandler := handlers.NewCrudCrawlerHandler(ctl)
	router.Handle("/api/crawler/{action:create|retrieve|update|delete}/{name}",
		crudHandler)
	listHandler := handlers.NewListHandler(ctl)
	router.Handle("/api/list/{type:seed|running|crontab|crawler}", listHandler)
	http.Handle("/api/", router)
	http.Handle("/", http.FileServer(rice.MustFindBox("http-files").HTTPBox()))
	http.ListenAndServe(addr, nil)
}

func main() {
	flag.Parse()
	defer glog.Flush()
	defer glog.Info("crawler exit")

	var ctl controller.Controller
	glog.Info("call ctl.Init")
	err := ctl.Init(*workingDir, *workerCnt)
	if err != nil {
		glog.Fatal(err)
	}
	defer ctl.Finish()
	go Web(&ctl, *serverAddr)
	ctl.Run()
}
