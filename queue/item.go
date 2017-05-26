package queue

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
)

type Item struct {
	ID    uint64
	Key   []byte
	Value []byte
}

func (i *Item) ToString() string {
	return string(i.Value)
}

func (i *Item) ToObject(value interface{}) error {
	buffer := bytes.NewBuffer(i.Value)
	decoder := gob.NewDecoder(buffer)
	return decoder.Decode(value)
}

func idToKey(id uint64) []byte {
	key := make([]byte, 8)
	binary.BigEndian.PutUint64(key, id)
	return key
}
func keyToID(key []byte) uint64 {
	return binary.BigEndian.Uint64(key)
}
