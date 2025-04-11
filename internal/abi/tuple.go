package abi

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/ethereum/go-ethereum/accounts/abi"
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

func buildTuples(tuples ...any) (map[string]abi.Argument, error) {
	types := make(map[string]abi.Argument)
	for _, t := range tuples {
		typ := reflect.TypeOf(t)
		if typ.Kind() != reflect.Struct {
			return nil, fmt.Errorf("expected struct, got %s", typ.Kind())
		}

		rawNames := make([]string, typ.NumField())
		elems := make([]*abi.Type, typ.NumField())
		for i := range typ.NumField() {
			field := typ.Field(i)
			rawNames[i] = field.Name
			elem, err := typeOf(field)
			if err != nil {
				return nil, err
			}
			elems[i] = elem
		}

		arg := abi.Argument{
			Name: typ.Name(),
			Type: abi.Type{
				T:             abi.TupleTy,
				TupleRawName:  typ.Name(),
				TupleElems:    elems,
				TupleRawNames: rawNames,
				TupleType:     typ,
			},
		}
		types[typ.Name()] = arg
	}
	return types, nil
}

// typeOf returns the abi.Type of a struct field.
func typeOf(field reflect.StructField) (*abi.Type, error) {
	const tagKey = "abitype"

	tag, tagOk := field.Tag.Lookup(tagKey)
	abiType, abiTypeOk := types[tag]
	if tagOk {
		if !abiTypeOk {
			return nil, fmt.Errorf("unknown abi type %q for field %q", tag, field.Name)
		}
		if abiTypeOk && abiType.GetType() != field.Type {
			return nil, fmt.Errorf("abi type %q for field %q is not compatible with its type %q", tag, field.Name, field.Type)
		}
		return &abiType, nil
	}

	switch field.Type.Kind() {
	case reflect.Bool:
		return &abi.Type{T: abi.BoolTy}, nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return &abi.Type{T: abi.UintTy}, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return &abi.Type{T: abi.IntTy}, nil
	}

	return nil, nil
}
