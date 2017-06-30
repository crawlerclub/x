package store

import (
	"github.com/blevesearch/bleve"
	"github.com/crawlerclub/x/types"
	"github.com/syndtr/goleveldb/leveldb"
	"os"
)

type Store struct {
	DataDir    string
	db         *leveldb.DB
	BleveIndex bleve.Index
	isOpen     bool
}

func Open(dataDir string) (*Store, error) {
	var err error
	s := &Store{
		DataDir: dataDir,
		db:      &leveldb.DB{},
		isOpen:  false,
	}
	dbDir := dataDir + "/leveldb"
	indexDir := dataDir + "/bleve.index"
	s.db, err = leveldb.OpenFile(dbDir, nil)
	if err != nil {
		return s, err
	}
	s.BleveIndex, err = bleve.Open(indexDir)
	if err == bleve.ErrorIndexPathDoesNotExist {
		indexMapping := buildIndexMapping()
		s.BleveIndex, err = bleve.New(indexDir, indexMapping)
		if err != nil {
			return s, err
		}
	} else if err != nil {
		return s, err
	}
	s.isOpen = true
	return s, nil
}

func (self *Store) Close() error {
	if !self.isOpen {
		return nil
	}
	if err := self.db.Close(); err != nil {
		return err
	}
	self.isOpen = false
	return nil
}

func (self *Store) Drop() error {
	if err := self.Close(); err != nil {
		return err
	}
	return os.RemoveAll(self.DataDir)
}

func (self *Store) Get(key []byte) ([]byte, error) {
	return self.db.Get(key, nil)
}

func (self *Store) UpdateItem(item types.StoreItem) error {
	key := []byte(item.Id())
	value, err := ObjectToBytes(item)
	if err != nil {
		return err
	}
	err = self.db.Put(key, value, nil)
	if err != nil {
		return err
	}
	err = self.BleveIndex.Index(item.Id(), item)
	return err
}
