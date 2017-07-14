package controller

import (
	"errors"
	"github.com/crawlerclub/x/crawler"
	"github.com/crawlerclub/x/store"
	"github.com/crawlerclub/x/types"
	"github.com/golang/glog"
	"github.com/syndtr/goleveldb/leveldb/util"
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
	ErrNameTaken      = errors.New("controller/controller.go CrawlerName already taken")
	ErrNoName         = errors.New("controller/controller.go no CrawlerName")
)

var StoreNames = []string{"crawler", "seed", "running", "crontab"}

type Controller struct {
	Crawlers    map[string]crawler.Crawler
	Schduler    CrawlerScheduler
	Stores      map[string]*store.LevelStore
	WorkerCount int

	workDir  string
	isInited bool
}

func timeStr(t int64) string {
	return time.Unix(t, 0).Format("20060102030405")
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
	self.Stores = make(map[string]*store.LevelStore)
	for _, name := range StoreNames {
		self.Stores[name], err = store.NewLevelStore(dir + "/db/" + name)
		if err != nil {
			return err
		}
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
		//return err
		glog.Error(err)
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
		task := types.Task{
			CrawlerName: item.Conf.CrawlerName,
			ParserName:  item.Conf.StartParserName,
			IsSeedUrl:   true,
			Url:         url,
		}
		if ret, _ := self.Stores["seed"].Has(task.Id()); !ret {
			// enqueue new start urls
			if _, err = crawler.TaskQueue.EnqueueObject(task); err != nil {
				return err
			}
		}
		value, _ := store.ObjectToBytes(task)
		self.Stores["seed"].Put(task.Id(), value) // update seed
	}
	return nil
}

func (self *Controller) initCrawlersFromDB() error {
	glog.Info("call initCrawlersFromDB")
	self.Crawlers = make(map[string]crawler.Crawler)
	self.Schduler.Init()

	crawlerStore := self.Stores["crawler"]
	count := 0
	err := crawlerStore.ForEach(nil, func(key, value []byte) (bool, error) {
		count += 1
		var item types.CrawlerItem
		e := store.BytesToObject(value, &item)
		if e != nil {
			return false, e
		}
		e = self.runCrawler(&item)
		if e != nil {
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		return err
	}
	glog.Info("loaded ", count, " crawler items")
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
	return self.Stores["crawler"].Delete(name)
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
	has, err := self.Stores["crawler"].Has(item.CrawlerName)
	if err != nil {
		return err
	}
	if isNew && has {
		return ErrNameTaken
	}
	if !isNew && !has {
		return ErrNoName
	}
	value, err := store.ObjectToBytes(item)
	if err != nil {
		return err
	}
	err = self.Stores["crawler"].Put(item.CrawlerName, value)
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
	for _, v := range self.Stores {
		v.Close()
	}
}

func (self *Controller) enqueueTask(wg *sync.WaitGroup, exitCh chan int, name string) {
	if !self.isInited {
		glog.Error(ErrNotInited)
		return
	}
	defer wg.Done()
	glog.Info("start ", name, " worker")
	defer glog.Info("exit ", name, " worker")
	for {
		select {
		case <-exitCh:
			return
		default:
			glog.Info("begin ", name)
			now := time.Now().Format("20060102030405")
			self.Stores[name].ForEach(&util.Range{Limit: []byte(now)},
				func(key, value []byte) (bool, error) {
					var task types.Task
					store.BytesToObject(value, &task)
					if crawler, ok := self.Crawlers[task.CrawlerName]; ok {
						crawler.TaskQueue.EnqueueObject(task)
					}
					self.Stores[name].Delete(string(key))
					return true, nil
				})
			// every 5 seconds
			time.Sleep(5 * time.Second)
		}
	}
}

func (self *Controller) cron(wg *sync.WaitGroup, exitCh chan int) {
	self.enqueueTask(wg, exitCh, "crontab")
}

func (self *Controller) retry(wg *sync.WaitGroup, exitCh chan int) {
	self.enqueueTask(wg, exitCh, "running")
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
			glog.Info("worker ", worker, " is working on ", name)
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
				key := timeStr(now+300) + "\t" + task.Id()
				value, _ := store.ObjectToBytes(task)
				self.Stores["running"].Put(key, value)

				glog.Info("process task:", task)
				tasks, items, err := crawler.Process(&task)
				if err != nil {
					glog.Error(err)
					continue
				}
				// remove task from Running
				self.Stores["running"].Delete(key)

				if parseConf, ok := crawler.Conf.ParseConfs[task.ParserName]; ok {
					if parseConf.RevisitInterval > 0 && task.IsSeedUrl {
						// add this task back to crontab
						task.LastAccessTime = now
						task.RevisitInterval = parseConf.RevisitInterval
						key = timeStr(now+task.RevisitInterval) + "\t" + task.Id()
						value, _ = store.ObjectToBytes(task)
						self.Stores["crontab"].Put(key, value)
					}
				}
				for _, t := range tasks {
					glog.Info("enqueue task:", t)
					// add SeedUrl to Seed
					if t.IsSeedUrl {
						value, _ = store.ObjectToBytes(t)
						self.Stores["seed"].Put(t.Id(), value)
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
