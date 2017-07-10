package store

import (
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/crawlerclub/x/types"
	_ "github.com/mattn/go-sqlite3"
)

var (
	ErrNilDB          = errors.New("store/sqlite3_db.go sql.DB is nil")
	ErrNilCrawlerItem = errors.New("store/sqlite3_db.go CrawlerItem is nil")
)

const (
	db           = "x"
	crawlerTable = "crawlers"
	taskTable    = "tasks"
	setupSql     = `CREATE TABLE IF NOT EXISTS crawlers (
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

	insertCrawlerSql = `INSERT INTO crawlers(crawler_name, conf, weight, status, create_time, author) VALUES(?, ?, ?, ?, ?, ?);`
	updateCrawlerSql = `UPDATE crawlers SET crawler_name=?, conf=?, weight=?, status=?, modify_time=?, author=? WHERE id=?;`
	selectCrawlerSql = `SELECT * FROM crawlers WHERE id=?;`
	deleteCrawlerSql = `DELETE FROM crawlers WHERE id=?;`
)

type Engine struct {
	db *sql.DB
}

func NewEngine(driverName, dbName string) (*Engine, error) {
	db, err := sql.Open(driverName, dbName)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return &Engine{db}, nil
}

func (self *Engine) Close() {
	if self.db != nil {
		self.db.Close()
	}
}

func (self *Engine) CreateTables() error {
	if self.db == nil {
		return ErrNilDB
	}
	var err error
	_, err = self.db.Exec(setupSql)
	return err
}

func (self *Engine) cu(item *types.CrawlerItem, action string) error {
	if item == nil {
		return ErrNilCrawlerItem
	}
	if self.db == nil {
		return ErrNilDB
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

func (self *Engine) CrawlerInsert(item *types.CrawlerItem) error {
	return self.cu(item, "insert")
}

func (self *Engine) CrawlerUpdate(item *types.CrawlerItem) error {
	return self.cu(item, "update")
}

func (self *Engine) CrawlerSelect(id int) (*types.CrawlerItem, error) {
	if self.db == nil {
		return nil, ErrNilDB
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

func (self *Engine) CrawlerDelete(id int) error {
	if self.db == nil {
		return ErrNilDB
	}
	stmt, err := self.db.Prepare(deleteCrawlerSql)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(id)
	return err
}
