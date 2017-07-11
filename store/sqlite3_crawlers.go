package store

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/crawlerclub/x/types"
	_ "github.com/mattn/go-sqlite3"
)

var (
	ErrNilCrawlerDB   = errors.New("store/sqlite3_crawlers.go sql.DB is nil")
	ErrNilCrawlerItem = errors.New("store/sqlite3_crawlers.go CrawlerItem is nil")
)

const (
	setupCrawlerSql = `CREATE TABLE IF NOT EXISTS crawlers (
		id INTEGER PRIMARY KEY,
		crawler_name TEXT UNIQUE NOT NULL,
		conf TEXT NOT NULL,
		weight INTEGER NOT NULL DEFAULT 1,
		status TEXT NOT NULL DEFAULT "enabled",
		create_time INTEGER NOT NULL DEFAULT 0,
		modify_time INTEGER NOT NULL DEFAULT 0,
		author TEXT NOT NULL DEFAULT ""
	);
	CREATE INDEX IF NOT EXISTS status_index on crawlers(status);
	`

	insertCrawlerSql       = `INSERT INTO crawlers(crawler_name, conf, weight, status, create_time, author) VALUES(?, ?, ?, ?, ?, ?);`
	updateCrawlerSql       = `UPDATE crawlers SET crawler_name=?, conf=?, weight=?, status=?, modify_time=?, author=? WHERE id=?;`
	selectCrawlerSql       = `SELECT * FROM crawlers WHERE id=?;`
	selectByNameCrawlerSql = `SELECT * FROM crawlers WHERE crawler_name=?;`
	deleteCrawlerSql       = `DELETE FROM crawlers WHERE id=?;`
	deleteByNameCrawlerSql = `DELETE FROM crawlers WHERE crawler_name=?;`
	queryCrawlerSql        = `SELECT id, crawler_name, weight, status, create_time, modify_time, author FROM crawlers %s;`
	countCrawlerSql        = `SELECT COUNT(*) FROM crawlers %s;`
)

type CrawlerDB struct {
	db *sql.DB
}

func NewCrawlerDB(driverName, dbName string) (*CrawlerDB, error) {
	db, err := sql.Open(driverName, dbName)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return &CrawlerDB{db}, nil
}

func (self *CrawlerDB) Close() {
	if self.db != nil {
		self.db.Close()
	}
}

func (self *CrawlerDB) CreateTables() error {
	if self.db == nil {
		return ErrNilCrawlerDB
	}
	var err error
	_, err = self.db.Exec(setupCrawlerSql)
	return err
}

func (self *CrawlerDB) cu(item *types.CrawlerItem, action string) error {
	if item == nil {
		return ErrNilCrawlerItem
	}
	if self.db == nil {
		return ErrNilCrawlerDB
	}
	conf, err := json.Marshal(item.Conf)
	if err != nil {
		return err
	}
	sql := insertCrawlerSql
	if action == "update" {
		sql = updateCrawlerSql
	}
	stmt, err := self.db.Prepare(sql)
	if err != nil {
		return err
	}
	defer stmt.Close()
	if action == "insert" {
		_, err = stmt.Exec(item.CrawlerName, string(conf), item.Weight, item.Status, item.CreateTime, item.Author)
	} else {
		_, err = stmt.Exec(item.CrawlerName, string(conf), item.Weight, item.Status, item.ModifyTime, item.Author, item.Id)
	}
	return err
}

func (self *CrawlerDB) Insert(item *types.CrawlerItem) error {
	return self.cu(item, "insert")
}

func (self *CrawlerDB) Update(item *types.CrawlerItem) error {
	return self.cu(item, "update")
}

func (self *CrawlerDB) Select(id int) (*types.CrawlerItem, error) {
	if self.db == nil {
		return nil, ErrNilCrawlerDB
	}
	stmt, err := self.db.Prepare(selectCrawlerSql)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	item := new(types.CrawlerItem)
	var conf string
	err = stmt.QueryRow(id).Scan(&item.Id, &item.CrawlerName, &conf, &item.Weight, &item.Status, &item.CreateTime, &item.ModifyTime, &item.Author)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal([]byte(conf), &item.Conf)
	if err != nil {
		return nil, err
	}
	return item, nil
}
func (self *CrawlerDB) SelectByName(name string) (*types.CrawlerItem, error) {
	if self.db == nil {
		return nil, ErrNilCrawlerDB
	}
	stmt, err := self.db.Prepare(selectByNameCrawlerSql)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	item := new(types.CrawlerItem)
	var conf string
	err = stmt.QueryRow(name).Scan(&item.Id, &item.CrawlerName, &conf, &item.Weight, &item.Status, &item.CreateTime, &item.ModifyTime, &item.Author)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal([]byte(conf), &item.Conf)
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (self *CrawlerDB) Delete(id int) error {
	if self.db == nil {
		return ErrNilCrawlerDB
	}
	stmt, err := self.db.Prepare(deleteCrawlerSql)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(id)
	return err
}

func (self *CrawlerDB) DeleteByName(name string) error {
	if self.db == nil {
		return ErrNilCrawlerDB
	}
	stmt, err := self.db.Prepare(deleteByNameCrawlerSql)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(name)
	return err
}

func (self *CrawlerDB) Count(query string, args ...interface{}) (int64, error) {
	if self.db == nil {
		return 0, ErrNilCrawlerDB
	}
	sql := fmt.Sprintf(countCrawlerSql, query)
	rows, err := self.db.Query(sql, args...)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	var count int64
	for rows.Next() {
		err = rows.Scan(&count)
		if err != nil {
			return 0, err
		}
		break
	}
	return count, nil
}

func (self *CrawlerDB) List(query string, args ...interface{}) ([]*types.CrawlerItem, error) {
	if self.db == nil {
		return nil, ErrNilCrawlerDB
	}
	sql := fmt.Sprintf(queryCrawlerSql, query)
	rows, err := self.db.Query(sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*types.CrawlerItem
	for rows.Next() {
		item := new(types.CrawlerItem)
		if err = rows.Scan(&item.Id, &item.CrawlerName, &item.Weight, &item.Status,
			&item.CreateTime, &item.ModifyTime, &item.Author); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
