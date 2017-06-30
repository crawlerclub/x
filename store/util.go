package store

import (
	"bytes"
	"encoding/gob"
)

func ObjectToBytes(object interface{}) ([]byte, error) {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	if err := encoder.Encode(object); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func BytesToObject(value []byte, object interface{}) error {
	buffer := bytes.NewBuffer(value)
	decoder := gob.NewDecoder(buffer)
	if err := decoder.Decode(object); err != nil {
		return err
	}
	return nil
}
