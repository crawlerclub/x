package store

import (
	"errors"
	"github.com/syndtr/goleveldb/leveldb"
	"os"
)

var (
	ErrNilLevelDB = errors.New("store/leveldb_store.go db is nil")
)

type LevelStore struct {
	db  *leveldb.DB
	dir string
}

func NewLevelStore(dir string) (*LevelStore, error) {
	db, err := leveldb.OpenFile(dir, nil)
	if err != nil {
		return nil, err
	}
	return &LevelStore{db: db, dir: dir}, nil
}

func (self *LevelStore) Close() error {
	if self.db != nil {
		return self.db.Close()
	}
	return nil
}

func (self *LevelStore) Drop() error {
	if err := self.Close(); err != nil {
		return err
	}
	if self.dir != "" {
		return os.RemoveAll(self.dir)
	}
	return nil
}

func (self *LevelStore) Get(key string) ([]byte, error) {
	if self.db == nil {
		return nil, ErrNilLevelDB
	}
	return self.db.Get([]byte(key), nil)
}

func (self *LevelStore) Put(key string, value []byte) error {
	if self.db == nil {
		return ErrNilLevelDB
	}
	return self.db.Put([]byte(key), value, nil)
}

func (self *LevelStore) Delete(key string) error {
	if self.db == nil {
		return ErrNilLevelDB
	}
	return self.db.Delete([]byte(key), nil)
}

func (self *LevelStore) DB() *leveldb.DB {
	return self.db
}
