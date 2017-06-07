package controller

import (
	"bytes"
	"encoding/gob"
	"errors"
	"github.com/crawlerclub/x/crawler"
	"github.com/crawlerclub/x/ds"
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
	ErrNotInited   = errors.New("call Controller.Init first")
	ErrWorkerCount = errors.New("worker count must between 0 and 1000")
)

type Controller struct {
	Crawlers    map[string]crawler.Crawler
	TaskQueue   *ds.Queue
	CrontabDB   *leveldb.DB
	RunningDB   *leveldb.DB
	WorkerCount int
	isInited    bool
}

func (self *Controller) Init(dir string, wc int) error {
	if wc <= 0 || wc > 1000 {
		return ErrWorkerCount
	}
	self.WorkerCount = wc
	var err error
	taskDir := dir + "/taskqueue"
	self.TaskQueue, err = ds.OpenQueue(taskDir)
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

	self.Crawlers = make(map[string]crawler.Crawler)
	self.isInited = true
	return nil
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
	err := self.TaskQueue.Close()
	if err != nil {
		return err
	}
	err = self.CrontabDB.Close()
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
				self.TaskQueue.EnqueueObject(task)
				db.Delete(key, nil)
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
			glog.Info("TaskQueue length: ", self.TaskQueue.Length())
			item, err := self.TaskQueue.Dequeue()
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
			now := time.Now().Unix()
			// add task to RunningDB
			err = self.AddRunning(now, task)
			if err != nil {
				glog.Error(err)
				continue
			}

			if crawler, ok := self.Crawlers[task.CrawlerName]; ok {
				glog.Info("process task:", task)
				tasks, items, err := crawler.Process(&task)
				if err != nil {
					glog.Error(err)
					continue
				}
				// remove task from RunningDB
				err = self.RunningDB.Delete([]byte(task.Url), nil)
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
					self.TaskQueue.EnqueueObject(t)
				}
				for _, item := range items {
					crawler.Save(item)
				}
			} else {
				glog.Error("No crawler named: ", task.CrawlerName)
			}
			time.Sleep(1 * time.Second)
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
