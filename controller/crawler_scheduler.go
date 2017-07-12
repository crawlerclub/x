package controller

import (
	"container/list"
	"errors"
	"github.com/crawlerclub/x/ds"
	"math/rand"
	"sync"
	"time"
)

var (
	ErrNilList      = errors.New("SortedList is nil")
	ErrNeverHappen  = errors.New("weighted random selection error")
	ErrNameNotFound = errors.New("crawler name not found")
)

type Item struct {
	CrawlerName string
	Weight      int
}

func compareWeight(left interface{}, right interface{}) int {
	// from big to small
	return right.(Item).Weight - left.(Item).Weight
}

type CrawlerScheduler struct {
	sync.RWMutex
	crawlerList *ds.SortedList
	totalWeight int
}

func (self *CrawlerScheduler) Init() {
	self.Lock()
	defer self.Unlock()
	if self.crawlerList != nil {
		return
	}
	self.totalWeight = 0
	self.crawlerList = ds.NewSortedList(compareWeight)
}

func (self *CrawlerScheduler) WeightedChoice() (string, error) {
	if self.crawlerList == nil || self.totalWeight <= 0 {
		return "", ErrNilList
	}
	rand.Seed(time.Now().UTC().UnixNano())
	self.Lock()
	defer self.Unlock()
	rnd := rand.Intn(self.totalWeight)
	for e := self.crawlerList.Front(); e != nil; e = e.Next() {
		rnd -= e.Value.(Item).Weight
		if rnd <= 0 {
			return e.Value.(Item).CrawlerName, nil
		}
	}
	return "", ErrNeverHappen
}

func (self *CrawlerScheduler) Insert(item Item) error {
	self.Lock()
	defer self.Unlock()
	if self.crawlerList == nil {
		return ErrNilList
	}
	self.crawlerList.Insert(item)
	self.totalWeight += item.Weight
	return nil
}

func (self *CrawlerScheduler) Remove(name string) error {
	self.Lock()
	defer self.Unlock()
	if self.crawlerList == nil {
		return ErrNilList
	}
	var next *list.Element
	for e := self.crawlerList.Front(); e != nil; e = next {
		next = e.Next()
		if e.Value.(Item).CrawlerName == name {
			self.crawlerList.Remove(e)
			self.totalWeight -= e.Value.(Item).Weight
			return nil
		}
	}
	return ErrNameNotFound
}
