package store

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/crawlerclub/x/types"
	_ "github.com/mattn/go-sqlite3"
)

var (
	ErrNilTaskDB   = errors.New("store/sqlite3_tasks.go sql.DB is nil")
	ErrNilTaskItem = errors.New("store/sqlite3_tasks.go Task is nil")
)

const (
	setupTaskSql = `CREATE TABLE IF NOT EXISTS tasks (
		id INTEGER PRIMARY KEY,
		crawler_name TEXT NOT NULL,
		parser_name TEXT NOT NULL,
		is_seed_url INTEGER NOT NULL DEFAULT 1,
		url TEXT NOT NULL DEFAULT "",
		data TEXT NOT NULL DEFAULT "",
		status TEXT NOT NULL DEFAULT "enabled",
		last_access_time INTEGER NOT NULL DEFAULT 0,
		revisit_interval INTEGER NOT NULL DEFAULT 0,
		next_exec_time INTEGER NOT NULL DEFAULT 0
	);
	CREATE INDEX IF NOT EXISTS crawler_name_index on tasks(crawler_name);
	CREATE INDEX IF NOT EXISTS parser_name_index on tasks(parser_name);
	CREATE INDEX IF NOT EXISTS status_index on tasks(status);
	CREATE INDEX IF NOT EXISTS next_exec_time_index on tasks(next_exec_time);
	`

	insertTaskSql = `INSERT INTO tasks(crawler_name, parser_name, is_seed_url, url, data, status, last_access_time, revisit_interval, next_exec_time) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?);`
	updateTaskSql = `UPDATE tasks SET crawler_name=?, parser_name=?, is_seed_url=?, url=?, data=?, status=?, last_access_time=?, revisit_interval=? WHERE id=?;`
	selectTaskSql = `SELECT * FROM tasks WHERE id=?;`
	deleteTaskSql = `DELETE FROM tasks WHERE id=?;`
	queryTaskSql  = `SELECT * FROM tasks %s;`
	countTaskSql  = `SELECT COUNT(*) FROM tasks %s;`
)

type TaskDB struct {
	db *sql.DB
}

func NewTaskDB(driverName, dbName string) (*TaskDB, error) {
	db, err := sql.Open(driverName, dbName)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	_, err = db.Exec(setupTaskSql)
	if err != nil {
		return nil, err
	}
	return &TaskDB{db}, nil
}

func (self *TaskDB) Close() {
	if self.db != nil {
		self.db.Close()
	}
}

func (self *TaskDB) cu(item *types.Task, action string) (int64, error) {
	if item == nil {
		return 0, ErrNilTaskItem
	}
	if self.db == nil {
		return 0, ErrNilTaskDB
	}
	sql := insertTaskSql
	if action == "update" {
		sql = updateTaskSql
	}
	stmt, err := self.db.Prepare(sql)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()
	if action == "insert" {
		result, err := stmt.Exec(item.CrawlerName, item.ParserName, item.IsSeedUrl,
			item.Url, item.Data, item.Status, item.LastAccessTime,
			item.RevisitInterval, item.NextExecTime)
		if err != nil {
			return 0, err
		}
		return result.LastInsertId()
	} else {
		_, err = stmt.Exec(item.CrawlerName, item.ParserName, item.IsSeedUrl,
			item.Url, item.Data, item.Status, item.LastAccessTime,
			item.RevisitInterval, item.NextExecTime, item.Id)
		return item.Id, err
	}
	return 0, nil
}

func (self *TaskDB) Insert(item *types.Task) (int64, error) {
	return self.cu(item, "insert")
}

func (self *TaskDB) Update(item *types.Task) (int64, error) {
	return self.cu(item, "update")
}

func (self *TaskDB) Select(id int) (*types.Task, error) {
	if self.db == nil {
		return nil, ErrNilTaskDB
	}
	stmt, err := self.db.Prepare(selectTaskSql)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	item := new(types.Task)
	err = stmt.QueryRow(id).Scan(&item.Id, &item.CrawlerName, &item.ParserName,
		&item.IsSeedUrl, &item.Url, &item.Data, &item.Status,
		&item.LastAccessTime, &item.RevisitInterval, &item.NextExecTime)
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (self *TaskDB) Delete(id int64) error {
	if self.db == nil {
		return ErrNilTaskDB
	}
	stmt, err := self.db.Prepare(deleteTaskSql)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(id)
	return err
}

func (self *TaskDB) Count(query string, args ...interface{}) (int, error) {
	if self.db == nil {
		return 0, ErrNilTaskDB
	}
	sql := fmt.Sprintf(queryTaskSql, query)
	rows, err := self.db.Query(sql, args...)
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

func (self *TaskDB) List(query string, args ...interface{}) ([]*types.Task, error) {
	if self.db == nil {
		return nil, ErrNilTaskDB
	}
	sql := fmt.Sprintf(queryTaskSql, query)
	rows, err := self.db.Query(sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*types.Task
	for rows.Next() {
		item := new(types.Task)
		if err = rows.Scan(&item.Id, &item.CrawlerName, &item.ParserName,
			&item.IsSeedUrl, &item.Url, &item.Data, &item.LastAccessTime,
			&item.RevisitInterval); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
