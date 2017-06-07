package ds

import "errors"

var (
	ErrDBClosed    = errors.New("queue: database is closed")
	ErrEmpty       = errors.New("queue: queue is empty")
	ErrOutOfBounds = errors.New("queue: ID is outside range of queue")
)
