package abi

import (
	"errors"
	"fmt"
	"reflect"
)

var errDuplicateTuple = errors.New("duplicate tuple definition")

func tupleMap(tuples ...any) (map[string]reflect.Type, error) {
	types := make(map[string]reflect.Type)
	for _, t := range tuples {
		typ := reflect.TypeOf(t)
		if typ.Kind() != reflect.Struct {
			return nil, fmt.Errorf("expected struct, got %s", typ.Kind())
		}

		if _, ok := types[typ.Name()]; ok {
			return nil, fmt.Errorf("%w: %s", errDuplicateTuple, typ.Name())
		}
		types[typ.Name()] = typ
	}
	return types, nil
}
