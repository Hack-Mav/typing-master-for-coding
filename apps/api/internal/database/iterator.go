package database

import (
	"fmt"
	"reflect"

	"google.golang.org/api/iterator"
)

// Iterator defines a query result iterator.
type Iterator interface {
	Next(dst interface{}) (*Key, error)
}

// sliceIterator iterates over a list of keys and JSON payloads.
type sliceIterator struct {
	keys []*Key
	data [][]byte
	idx  int
}

func newSliceIterator(keys []*Key, data [][]byte) Iterator {
	return &sliceIterator{keys: keys, data: data}
}

func (it *sliceIterator) Next(dst interface{}) (*Key, error) {
	if it.idx >= len(it.keys) {
		return nil, iterator.Done
	}

	key := it.keys[it.idx]
	payload := it.data[it.idx]
	it.idx++

	if dst == nil {
		return key, nil
	}

	if err := unmarshalEntity(payload, dst); err != nil {
		return nil, err
	}
	return key, nil
}

// reflectSliceIterator is used by the in-memory store when it has a slice of entities.
type reflectSliceIterator struct {
	keys []*Key
	vals []reflect.Value
	idx  int
}

func (it *reflectSliceIterator) Next(dst interface{}) (*Key, error) {
	if it.idx >= len(it.keys) {
		return nil, iterator.Done
	}

	key := it.keys[it.idx]
	val := it.vals[it.idx]
	it.idx++

	if dst == nil {
		return key, nil
	}

	dv := reflect.ValueOf(dst)
	if dv.Kind() != reflect.Ptr {
		return nil, fmt.Errorf("dst must be a pointer")
	}
	dv.Elem().Set(val)
	return key, nil
}
