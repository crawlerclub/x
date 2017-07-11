package store

import (
	"database/sql"
	"errors"
	_ "github.com/mattn/go-sqlite3"
	"time"
)

var (
	ErrNilLinkDB = errors.New("store/sqlite3_Links.go sql.DB is nil")
)

const (
	setupLinkSql = `CREATE TABLE IF NOT EXISTS Links (
		id INTEGER PRIMARY KEY,
		url TEXT UNIQUE NOT NULL,
		last_access_time INTEGER NOT NULL DEFAULT 0,
		first_access_time DATATIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		count INTEGER NOT NULL DEFAULT 1
	);
	CREATE INDEX IF NOT EXISTS last_access_time_index on Links(last_access_time);
	`

	insertLinkSql = `INSERT INTO Links(last_access_time, url) VALUES(?, ?);`
	updateLinkSql = `UPDATE Links SET last_access_time=?, count=count+1 WHERE url=?;`
	countLinkSql  = `SELECT COUNT(*) FROM Links WHERE url=?;`
	deleteLinkSql = `DELETE FROM Links WHERE last_access_time<?;`
)

type LinkDB struct {
	db *sql.DB
}

func NewLinkDB(driverName, dbName string) (*LinkDB, error) {
	db, err := sql.Open(driverName, dbName)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	_, err = db.Exec(setupLinkSql)
	if err != nil {
		return nil, err
	}
	return &LinkDB{db}, nil
}

func (self *LinkDB) Close() {
	if self.db != nil {
		self.db.Close()
	}
}

func (self *LinkDB) Has(url string) (bool, error) {
	if self.db == nil {
		return false, ErrNilLinkDB
	}
	count, err := self.Count(url)
	if err != nil {
		return false, err
	}
	var stmt *sql.Stmt
	var has bool
	if count > 0 {
		has = true
		stmt, err = self.db.Prepare(updateLinkSql)
	} else {
		has = false
		stmt, err = self.db.Prepare(insertLinkSql)
	}
	if err != nil {
		return has, err
	}
	defer stmt.Close()
	_, err = stmt.Exec(time.Now().Unix(), url)
	return has, err
}

func (self *LinkDB) Count(url string) (int, error) {
	if self.db == nil {
		return 0, ErrNilLinkDB
	}
	rows, err := self.db.Query(countLinkSql, url)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	var count int
	for rows.Next() {
		err = rows.Scan(&count)
		if err != nil {
			return 0, err
		}
		break
	}
	return count, nil
}
