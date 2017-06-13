package main

import (
	"flag"
	"github.com/crawlerclub/x/controller"
	"github.com/crawlerclub/x/crawler"
	"github.com/golang/glog"
	"net/http"
	_ "net/http/pprof"
)

func main() {
	go func() {
		http.ListenAndServe("0.0.0.0:8309", nil)
	}()
	defer glog.Flush()
	defer glog.Info("CRAWLER exit")

	workerCnt := flag.Int("wc", 1000, "worker count")
	confFile := flag.String("conf", "./www.newsmth.net.json", "crawler conf file")
	runDir := flag.String("run", "./run", "working dir")
	flag.Parse()

	var controller controller.Controller
	controller.Init(*runDir, *workerCnt)

	var crawler crawler.Crawler
	err := crawler.LoadConfFromFile(*confFile)
	if err != nil {
		glog.Fatal(err)
	}
	controller.Crawlers[crawler.Conf.CrawlerName] = crawler
	/*
		tasks, err := crawler.GetStartTasks()
		if err != nil {
			glog.Fatal(err)
		}
		for _, task := range tasks {
			if _, err := controller.TaskQueue.EnqueueObject(task); err != nil {
				glog.Fatal(err)
			}
		}
	*/
	glog.Info("run!")
	controller.Run()
}
