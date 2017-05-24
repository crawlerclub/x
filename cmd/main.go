package main

import (
	"flag"
	"github.com/beeker1121/goque"
	"github.com/crawlerclub/x/crawler"
	"github.com/crawlerclub/x/types"
	"github.com/golang/glog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func startWorker(worker int, wg *sync.WaitGroup, exitCh chan int, queue *goque.Queue, crawler *crawler.Crawler) {
	defer wg.Done()
	glog.Info("start worker: ", worker)
	defer glog.Info("exit worker: ", worker)
	for {
		select {
		case <-exitCh:
			return
		default:
			glog.Info("Work on next task! worker: ", worker)
			time.Sleep(1 * time.Second)
			glog.Info(queue.Length())
			item, err := queue.Dequeue()
			if err != nil {
				glog.Error(err)
				time.Sleep(1 * time.Second)
				continue
			}
			var task types.Task
			err = item.ToObject(&task)
			if err != nil {
				glog.Error(err)
				continue
			}
			glog.Info("process task:", task)
			tasks, items, err := crawler.Process(&task)
			if err != nil {
				glog.Error(err)
				continue
			}
			for _, t := range tasks {
				glog.Info("enqueue task:", t)
				queue.EnqueueObject(t)
				break
			}
			for _, item := range items {
				crawler.Save(item)
			}
		}
	}
}

func Stop(sigs chan os.Signal, exitCh chan int) {
	<-sigs
	glog.Info("receive stop signal")
	close(exitCh)
}

func main() {
	defer glog.Flush()
	defer glog.Info("crawler exit")

	workerCnt := flag.Int("wc", 1, "worker count")
	confFile := flag.String("conf", "./www.newsmth.net.json", "crawler conf file")
	runDir := flag.String("run", "./run", "working dir")
	flag.Parse()
	if *workerCnt > 1000 {
		glog.Fatal("worker count too large, no larger than 1000")
	}

	var crawler crawler.Crawler
	err := crawler.LoadConfFromFile(*confFile)
	if err != nil {
		glog.Fatal(err)
	}

	queue, err := goque.OpenQueue(*runDir)
	if err != nil {
		glog.Fatal(err)
	}
	defer queue.Close()

	if queue.Length() == 0 {
		glog.Info("New round of crawl")
		tasks, err := crawler.GetStartTasks()
		if err != nil {
			glog.Fatal(err)
		}
		for _, task := range tasks {
			if _, err := queue.EnqueueObject(task); err != nil {
				glog.Fatal(err)
			}
		}
	}

	exitCh := make(chan int)
	sigs := make(chan os.Signal)
	var wg sync.WaitGroup
	for i := 0; i < *workerCnt; i++ {
		wg.Add(1)
		go startWorker(i, &wg, exitCh, queue, &crawler)
	}
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go Stop(sigs, exitCh)
	wg.Wait()
}
