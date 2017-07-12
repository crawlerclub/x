package controller

import (
	"errors"
	"github.com/crawlerclub/x/crawler"
	"github.com/crawlerclub/x/store"
	"github.com/crawlerclub/x/types"
	"github.com/golang/glog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var (
	ErrNotInited      = errors.New("controller/controller.go call Controller.Init first")
	ErrWorkerCount    = errors.New("controller/controller.go worker count must between 0 and 1000")
	ErrNilCrawlerItem = errors.New("controller/controller.go CrawlerItem is nil")
	ErrNamesNotSame   = errors.New("controller/controller.go CrawlerNames are not the same")
)

type Controller struct {
	Crawlers     map[string]crawler.Crawler
	Schduler     CrawlerScheduler
	CrawlerStore *store.CrawlerDB
	RunningStore *store.TaskDB
	CrontabStore *store.TaskDB
	LinkStore    *store.LinkDB
	WorkerCount  int

	workDir  string
	isInited bool
}

func (self *Controller) Init(dir string, wc int) error {
	glog.Info("call Init")
	if wc <= 0 || wc > 1000 {
		return ErrWorkerCount
	}
	self.workDir = dir
	self.WorkerCount = wc
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			return err
		}
	}
	self.CrawlerStore, err = store.NewCrawlerDB("sqlite3", dir+"/crawlers.sqlite3")
	if err != nil {
		return err
	}
	self.RunningStore, err = store.NewTaskDB("sqlite3", dir+"/running.sqlite3")
	if err != nil {
		return err
	}
	self.CrontabStore, err = store.NewTaskDB("sqlite3", dir+"/crontab.sqlite3")
	if err != nil {
		return err
	}
	self.LinkStore, err = store.NewLinkDB("sqlite3", dir+"/link.sqlite3")
	if err != nil {
		return err
	}
	err = self.initCrawlersFromDB()
	if err != nil {
		return err
	}
	self.isInited = true
	return nil
}

func (self *Controller) runCrawler(item *types.CrawlerItem) error {
	glog.Info("call runCrawler: ", item.CrawlerName)
	var crawler crawler.Crawler
	crawler.Conf = &item.Conf
	err := crawler.InitEs()
	if err != nil {
		return err
	}
	dir := self.workDir + "/queue"
	err = crawler.InitTaskQueue(dir)
	if err != nil {
		return err
	}
	if _, ok := self.Crawlers[item.CrawlerName]; ok {
		self.Schduler.Remove(item.CrawlerName)
	}
	self.Crawlers[item.CrawlerName] = crawler
	var sitem Item
	sitem.CrawlerName = item.CrawlerName
	sitem.Weight = item.Weight
	glog.Info("SortedList add ", sitem)
	err = self.Schduler.Insert(sitem)
	if err != nil {
		return err
	}
	for _, url := range item.Conf.StartUrls {
		if has, _ := self.LinkStore.Has(url); has {
			continue
		}
		task := types.Task{Url: url, CrawlerName: item.Conf.CrawlerName,
			ParserName: item.Conf.StartParserName, IsSeedUrl: false}
		p, _ := item.Conf.ParseConfs[item.Conf.StartParserName]
		if p.RevisitInterval > 0 {
			task.IsSeedUrl = true
			task.RevisitInterval = p.RevisitInterval
			task.NextExecTime = time.Now().Unix() + task.RevisitInterval
			id, err := self.CrontabStore.Insert(&task)
			if err != nil {
				return err
			}
			task.Id = id
		}
		if _, err = crawler.TaskQueue.EnqueueObject(task); err != nil {
			return err
		}
	}
	return nil
}

func (self *Controller) initCrawlersFromDB() error {
	glog.Info("call initCrawlersFromDB")
	self.Crawlers = make(map[string]crawler.Crawler)
	self.Schduler.Init()
	items, err := self.CrawlerStore.List("WHERE status=?", "enabled")
	glog.Info("loaded ", len(items), " crawler confs")
	if err != nil {
		return err
	}
	for _, item := range items {
		err = self.runCrawler(item)
		if err != nil {
			return err
		}
	}
	return nil
}

func (self *Controller) DelCrawler(name string) error {
	err := self.Schduler.Remove(name)
	if err != nil && err != ErrNameNotFound {
		return err
	}
	if crawler, ok := self.Crawlers[name]; ok {
		crawler.Close()
	}
	delete(self.Crawlers, name)
	return self.CrawlerStore.DeleteByName(name)
}

func (self *Controller) AddCrawler(item *types.CrawlerItem, isNew bool) error {
	if item == nil {
		return ErrNilCrawlerItem
	}
	ok, err := item.Conf.IsValid()
	if !ok {
		return err
	}
	if item.CrawlerName != item.Conf.CrawlerName {
		return ErrNamesNotSame
	}
	if isNew {
		err = self.CrawlerStore.Insert(item)
	} else {
		err = self.CrawlerStore.Update(item)
	}
	if err != nil {
		return err
	}
	if item.Status == "enabled" && item.Weight > 0 {
		err = self.runCrawler(item)
		if err != nil {
			return err
		}
	}
	return nil
}

func (self *Controller) Finish() {
	if !self.isInited {
		return
	}
	self.CrawlerStore.Close()
	self.RunningStore.Close()
	self.CrontabStore.Close()
}

func (self *Controller) enqueueTask(wg *sync.WaitGroup, exitCh chan int, name string) {
	if !self.isInited {
		glog.Error(ErrNotInited)
		return
	}
	defer wg.Done()
	glog.Info("start ", name, " worker")
	defer glog.Info("exit ", name, "worker")
	for {
		select {
		case <-exitCh:
			return
		default:
			glog.Info("begin ", name)
			now := time.Now().Unix()
			var tasks []*types.Task
			var err error
			if name == "cron" {
				tasks, err = self.CrontabStore.List("WHERE next_exec_time>?", now)
			} else if name == "retry" {
				tasks, err = self.RunningStore.List("WHERE next_exec_time>?", now)
			} else {
				glog.Error("unknown worker: ", name)
				return
			}
			if err != nil {
				glog.Error(err)
				return
			}
			for _, task := range tasks {
				if crawler, ok := self.Crawlers[task.CrawlerName]; ok {
					crawler.TaskQueue.EnqueueObject(task)
					if name == "retry" {
						continue
					}
					task.LastAccessTime = now
					task.NextExecTime = now + task.RevisitInterval
					self.CrontabStore.Update(task)
				} else {
					glog.Error("no crawler for task.CrawlerName: ", task.CrawlerName)
				}
			}
			// every 5 seconds
			time.Sleep(5 * time.Second)
		}
	}
}

func (self *Controller) cron(wg *sync.WaitGroup, exitCh chan int) {
	self.enqueueTask(wg, exitCh, "cron")
}

func (self *Controller) retry(wg *sync.WaitGroup, exitCh chan int) {
	self.enqueueTask(wg, exitCh, "retry")
}

func (self *Controller) startWorker(worker int, wg *sync.WaitGroup, exitCh chan int) {
	if !self.isInited {
		glog.Error(ErrNotInited)
		return
	}
	defer wg.Done()
	glog.Info("start crawler worker: ", worker)
	defer glog.Info("exit crawler worker: ", worker)
	for {
		select {
		case <-exitCh:
			return
		default:
			glog.Info("Work on next task! worker: ", worker)
			name, err := self.Schduler.WeightedChoice()
			if err != nil {
				glog.Error(err)
				time.Sleep(10 * time.Second)
				continue
			}
			glog.Info(worker, " is working on ", name)
			if crawler, ok := self.Crawlers[name]; ok {
				item, err := crawler.TaskQueue.Dequeue()
				if err != nil {
					glog.Error(err)
					continue
				}
				var task types.Task
				err = item.ToObject(&task)
				if err != nil {
					glog.Error(err)
					continue
				}
				now := time.Now().Unix()
				task.NextExecTime = now + 300
				id, err := self.RunningStore.Insert(&task)
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
				// remove task from RunningDB
				err = self.RunningStore.Delete(id)
				if err != nil {
					glog.Error(err)
				}
				if parseConf, ok := crawler.Conf.ParseConfs[task.ParserName]; ok {
					if parseConf.RevisitInterval > 0 && task.IsSeedUrl {
						// add this task back to cron
						task.LastAccessTime = now
						task.RevisitInterval = parseConf.RevisitInterval
						task.NextExecTime = now + task.RevisitInterval
						_, err = self.CrontabStore.Update(&task)
						if err != nil {
							glog.Error(err)
						}
					}
				}
				for _, t := range tasks {
					glog.Info("enqueue task:", t)
					// add SeedUrl to CrontabStore
					if t.IsSeedUrl {
						if has, _ := self.LinkStore.Has(t.Url); has {
							continue
						}
						if p, has := crawler.Conf.ParseConfs[t.ParserName]; has {
							if p.RevisitInterval > 0 {
								t.RevisitInterval = p.RevisitInterval
								t.NextExecTime = now + t.RevisitInterval
								_, err = self.CrontabStore.Insert(&t)
								if err != nil {
									glog.Error(err)
								}
							}
						}
					}
					crawler.TaskQueue.EnqueueObject(t)
				}
				for _, item := range items {
					crawler.Save(item)
				}
			} else {
				glog.Error("No crawler named: ", name)
			}
		}
	}
}

func (self *Controller) stop(sigs chan os.Signal, exitCh chan int) {
	if !self.isInited {
		glog.Error(ErrNotInited)
		return
	}
	<-sigs
	glog.Info("receive stop signal")
	close(exitCh)
}

func (self *Controller) Run() {
	if !self.isInited {
		glog.Error(ErrNotInited)
		return
	}
	exitCh := make(chan int)
	sigs := make(chan os.Signal)
	var wg sync.WaitGroup
	for i := 0; i < self.WorkerCount; i++ {
		wg.Add(1)
		go self.startWorker(i, &wg, exitCh)
	}
	wg.Add(1)
	go self.cron(&wg, exitCh)
	wg.Add(1)
	go self.retry(&wg, exitCh)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go self.stop(sigs, exitCh)
	wg.Wait()
}
