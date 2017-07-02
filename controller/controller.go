package controller

import (
	"bytes"
	"encoding/gob"
	"errors"
	"github.com/crawlerclub/x/crawler"
	"github.com/crawlerclub/x/store"
	"github.com/crawlerclub/x/types"
	"github.com/golang/glog"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var (
	ErrNotInited      = errors.New("call Controller.Init first")
	ErrWorkerCount    = errors.New("worker count must between 0 and 1000")
	ErrNilCrawlerItem = errors.New("CrawlerItem is nil")
	ErrNamesNotSame   = errors.New("CrawlerNames are not the same")
)

type Controller struct {
	Crawlers  map[string]crawler.Crawler
	Schduler  CrawlerScheduler
	CrawlerDB *leveldb.DB
	CrontabDB *leveldb.DB
	RunningDB *leveldb.DB

	WorkerCount int

	workDir  string
	isInited bool
}

func (self *Controller) Init(dir string, wc int) error {
	if wc <= 0 || wc > 1000 {
		return ErrWorkerCount
	}
	self.workDir = dir
	self.WorkerCount = wc
	var err error

	crawlerDir := dir + "/crawlerdb"
	self.CrawlerDB, err = leveldb.OpenFile(crawlerDir, nil)
	if err != nil {
		return err
	}

	cronDir := dir + "/crontabdb"
	self.CrontabDB, err = leveldb.OpenFile(cronDir, nil)
	if err != nil {
		return err
	}

	runDir := dir + "/runningdb"
	self.RunningDB, err = leveldb.OpenFile(runDir, nil)
	if err != nil {
		return err
	}

	self.initCrawlersFromDB()

	self.isInited = true
	return nil
}

func (self *Controller) DelCrawler(name string) error {
	err := self.Schduler.Remove(name)
	if err != nil && err != ErrNameNotFound {
		return err
	}
	delete(self.Crawlers, name)
	return self.CrawlerDB.Delete([]byte(name), nil)
}

func (self *Controller) AddCrawler(item *types.CrawlerItem, name string) error {
	if item == nil {
		return ErrNilCrawlerItem
	}
	ok, err := item.Conf.IsValid()
	if !ok {
		return err
	}
	if item.CrawlerName != item.Conf.CrawlerName || item.CrawlerName != name {
		return ErrNamesNotSame
	}
	data, err := store.ObjectToBytes(*item)
	if err != nil {
		return nil
	}
	err = self.CrawlerDB.Put([]byte(name), data, nil)
	if err != nil {
		return err
	}
	if item.Status == "enable" && item.Weight > 0 {
		var crawler crawler.Crawler
		crawler.Conf = &item.Conf
		crawler.InitEs()

		dir := self.workDir + "/queue/" + name
		err = crawler.InitTaskQueue(dir)
		if err != nil {
			return err
		}
		self.Crawlers[name] = crawler

		err := self.Schduler.Remove(name)
		if err != nil && err != ErrNameNotFound {
			return err
		}
		var sitem Item
		sitem.CrawlerName = name
		sitem.Weight = item.Weight
		err = self.Schduler.Insert(sitem)
		if err != nil {
			return err
		}
	}
	return nil
}

func (self *Controller) initCrawlersFromDB() error {
	self.Crawlers = make(map[string]crawler.Crawler)
	self.Schduler.Init()
	var err error
	iter := self.CrawlerDB.NewIterator(nil, nil)
	defer iter.Release()
	for iter.Next() {
		name := string(iter.Key())
		value := iter.Value()
		var item types.CrawlerItem
		err = store.BytesToObject(value, &item)
		if err != nil {
			glog.Error(err)
			continue
		}

		if item.Status != "enabled" {
			continue
		}

		var crawler crawler.Crawler
		crawler.Conf = &item.Conf
		crawler.InitEs()

		dir := self.workDir + "/queue/" + name
		err = crawler.InitTaskQueue(dir)
		if err != nil {
			return err
		}
		self.Crawlers[name] = crawler

		var sitem Item
		sitem.CrawlerName = name
		sitem.Weight = item.Weight
		err = self.Schduler.Insert(sitem)
		if err != nil {
			return err
		}
	}
	return iter.Error()
}

func (self *Controller) addToDB(ts int64, task types.Task, db *leveldb.DB) error {
	if !self.isInited {
		return ErrNotInited
	}
	key := time.Unix(ts, 0).Format("20060102030405") + ":" + task.Url
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	if err := encoder.Encode(task); err != nil {
		return err
	}
	err := db.Put([]byte(key), buffer.Bytes(), nil)
	return err
}

func (self *Controller) AddCrontab(ts int64, task types.Task) error {
	glog.Info("add crontab: ", task)
	return self.addToDB(ts, task, self.CrontabDB)
}

func (self *Controller) AddRunning(ts int64, task types.Task) error {
	return self.addToDB(ts, task, self.RunningDB)
}

func (self *Controller) Finish() error {
	if !self.isInited {
		return nil
	}
	err := self.CrontabDB.Close()
	if err != nil {
		return err
	}
	err = self.CrawlerDB.Close()
	if err != nil {
		return err
	}
	err = self.RunningDB.Close()
	return err
}

func (self *Controller) enqueueTask(wg *sync.WaitGroup, exitCh chan int, db *leveldb.DB, diff int64, name string) {
	if !self.isInited {
		glog.Error(ErrNotInited)
		return
	}
	defer wg.Done()
	glog.Info("start ", name)
	defer glog.Info("exit ", name)
	for {
		select {
		case <-exitCh:
			return
		default:
			glog.Info("begin ", name)
			now := time.Now().Unix()
			start := time.Unix(0, 0).Format("20060102030405")
			limit := time.Unix(now+diff, 0).Format("20060102030405")
			iter := db.NewIterator(&util.Range{Start: []byte(start), Limit: []byte(limit)}, nil)
			for iter.Next() {
				key := iter.Key()
				value := iter.Value()
				buffer := bytes.NewBuffer(value)
				decoder := gob.NewDecoder(buffer)
				var task types.Task
				err := decoder.Decode(&task)
				glog.Info(string(key), task)
				if err != nil {
					glog.Error(err)
					continue
				}
				if crawler, ok := self.Crawlers[task.CrawlerName]; ok {
					crawler.TaskQueue.EnqueueObject(task)
					db.Delete(key, nil)
				} else {
					glog.Error("no crawler for task.CrawlerName: ", task.CrawlerName)
				}
			}
			iter.Release()
			err := iter.Error()
			if err != nil {
				glog.Error(err)
			}
			// every 60 seconds
			time.Sleep(5 * time.Second)
		}
	}
}

func (self *Controller) cron(wg *sync.WaitGroup, exitCh chan int) {
	self.enqueueTask(wg, exitCh, self.CrontabDB, 120, "cron worker")
}

func (self *Controller) retry(wg *sync.WaitGroup, exitCh chan int) {
	self.enqueueTask(wg, exitCh, self.RunningDB, -600, "retry worker")
}

func (self *Controller) startWorker(worker int, wg *sync.WaitGroup, exitCh chan int) {
	if !self.isInited {
		glog.Error(ErrNotInited)
		return
	}
	defer wg.Done()
	glog.Info("start worker: ", worker)
	defer glog.Info("exit worker: ", worker)
	for {
		select {
		case <-exitCh:
			return
		default:
			glog.Info("Work on next task! worker: ", worker)
			name, err := self.Schduler.WeightedChoice()
			if err != nil {
				glog.Error(err)
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
				// add task to RunningDB
				err = self.AddRunning(now, task)
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
				key := time.Unix(now, 0).Format("20060102030405") + ":" + task.Url
				err = self.RunningDB.Delete([]byte(key), nil)
				if err != nil {
					glog.Error(err)
				}
				if parseConf, ok := crawler.Conf.ParseConfs[task.ParserName]; ok {
					if parseConf.RevisitInterval > 0 {
						// add this task back to cron
						task.LastAccessTime = now
						err = self.AddCrontab(now+parseConf.RevisitInterval, task)
						if err != nil {
							glog.Error(err)
						}
					}
				}

				for _, t := range tasks {
					glog.Info("enqueue task:", t)
					crawler.TaskQueue.EnqueueObject(t)
				}
				for _, item := range items {
					crawler.Save(item)
				}
			} else {
				glog.Error("No crawler named: ", name)
			}
			//time.Sleep(1 * time.Second)
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
