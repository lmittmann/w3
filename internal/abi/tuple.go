package abi

import (
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"unicode"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
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

		if _, ok := types[typ.Name()]; ok {
			return nil, fmt.Errorf("%w: %s", errDuplicateTuple, typ.Name())
		}

		abiTyp, err := typeOf(typ, "")
		if err != nil {
			return nil, err
		}

		arg := abi.Argument{
			Name: typ.Name(),
			Type: *abiTyp,
		}
		types[typ.Name()] = arg
	}
	return types, nil
}

// typeOfField returns the [abi.Type] of a struct field.
func typeOfField(field reflect.StructField) (*abi.Type, error) {
	const tagKey = "abitype"

	tag, _ := field.Tag.Lookup(tagKey)
	return typeOf(field.Type, tag) // tag is "" if not set
}

func typeOf(typ reflect.Type, abiType string) (*abi.Type, error) {
	abiT, isBasicT := basicTypes[typ]
	if !isBasicT {
		switch typ.Kind() {
		case reflect.Slice:
			elemAibT, err := typeOf(typ.Elem(), abiType)
			if err != nil {
				return nil, err
			}
			return &abi.Type{
				T:    abi.SliceTy,
				Elem: elemAibT,
			}, nil
		case reflect.Array:
			elemAibT, err := typeOf(typ.Elem(), abiType)
			if err != nil {
				return nil, err
			}
			return &abi.Type{
				T:    abi.ArrayTy,
				Elem: elemAibT,
				Size: typ.Len(),
			}, nil
		case reflect.Struct:
			num := typ.NumField()
			elems := make([]*abi.Type, num)
			rawNames := make([]string, num)
			for i := range num {
				f := typ.Field(i)
				elemType, err := typeOfField(f)
				if err != nil {
					return nil, err
				}
				elems[i] = elemType
				rawNames[i] = toCamelCase(f.Name)
			}
			return &abi.Type{
				T:             abi.TupleTy,
				TupleElems:    elems,
				TupleRawName:  typ.Name(),
				TupleRawNames: rawNames,
				TupleType:     typ,
			}, nil
		}
		return nil, fmt.Errorf("unknown type %q", typ)
	}

	if abiType == "" {
		// if no abiType is specified, return the basic type directly.
		return &abiT, nil
	}

	abiT, ok := types[abiType]
	if !ok {
		return nil, fmt.Errorf("unknown abi type %q", abiType)
	}
	if abiT.GetType() != typ && !(typ == reflect.TypeFor[*big.Int]() &&
		(abiT.GetType().Kind() == reflect.Int16 ||
			abiT.GetType().Kind() == reflect.Int32 ||
			abiT.GetType().Kind() == reflect.Int64 ||
			abiT.GetType().Kind() == reflect.Uint16 ||
			abiT.GetType().Kind() == reflect.Uint32 ||
			abiT.GetType().Kind() == reflect.Uint64)) {
		return nil, fmt.Errorf("tagged type %q does not match type %v", abiType, typ)
	}
	return &abiT, nil
}

var basicTypes = map[reflect.Type]abi.Type{
	reflect.TypeFor[bool]():           {T: abi.BoolTy},
	reflect.TypeFor[byte]():           {T: abi.UintTy, Size: 8},
	reflect.TypeFor[uint8]():          {T: abi.UintTy, Size: 8},
	reflect.TypeFor[uint16]():         {T: abi.UintTy, Size: 16},
	reflect.TypeFor[uint32]():         {T: abi.UintTy, Size: 32},
	reflect.TypeFor[uint64]():         {T: abi.UintTy, Size: 64},
	reflect.TypeFor[int8]():           {T: abi.IntTy, Size: 8},
	reflect.TypeFor[int16]():          {T: abi.IntTy, Size: 16},
	reflect.TypeFor[int32]():          {T: abi.IntTy, Size: 32},
	reflect.TypeFor[int64]():          {T: abi.IntTy, Size: 64},
	reflect.TypeFor[[1]byte]():        {T: abi.FixedBytesTy, Size: 1},
	reflect.TypeFor[[2]byte]():        {T: abi.FixedBytesTy, Size: 2},
	reflect.TypeFor[[3]byte]():        {T: abi.FixedBytesTy, Size: 3},
	reflect.TypeFor[[4]byte]():        {T: abi.FixedBytesTy, Size: 4},
	reflect.TypeFor[[5]byte]():        {T: abi.FixedBytesTy, Size: 5},
	reflect.TypeFor[[6]byte]():        {T: abi.FixedBytesTy, Size: 6},
	reflect.TypeFor[[7]byte]():        {T: abi.FixedBytesTy, Size: 7},
	reflect.TypeFor[[8]byte]():        {T: abi.FixedBytesTy, Size: 8},
	reflect.TypeFor[[9]byte]():        {T: abi.FixedBytesTy, Size: 9},
	reflect.TypeFor[[10]byte]():       {T: abi.FixedBytesTy, Size: 10},
	reflect.TypeFor[[11]byte]():       {T: abi.FixedBytesTy, Size: 11},
	reflect.TypeFor[[12]byte]():       {T: abi.FixedBytesTy, Size: 12},
	reflect.TypeFor[[13]byte]():       {T: abi.FixedBytesTy, Size: 13},
	reflect.TypeFor[[14]byte]():       {T: abi.FixedBytesTy, Size: 14},
	reflect.TypeFor[[15]byte]():       {T: abi.FixedBytesTy, Size: 15},
	reflect.TypeFor[[16]byte]():       {T: abi.FixedBytesTy, Size: 16},
	reflect.TypeFor[[17]byte]():       {T: abi.FixedBytesTy, Size: 17},
	reflect.TypeFor[[18]byte]():       {T: abi.FixedBytesTy, Size: 18},
	reflect.TypeFor[[19]byte]():       {T: abi.FixedBytesTy, Size: 19},
	reflect.TypeFor[[20]byte]():       {T: abi.FixedBytesTy, Size: 20},
	reflect.TypeFor[[21]byte]():       {T: abi.FixedBytesTy, Size: 21},
	reflect.TypeFor[[22]byte]():       {T: abi.FixedBytesTy, Size: 22},
	reflect.TypeFor[[23]byte]():       {T: abi.FixedBytesTy, Size: 23},
	reflect.TypeFor[[24]byte]():       {T: abi.FixedBytesTy, Size: 24},
	reflect.TypeFor[[25]byte]():       {T: abi.FixedBytesTy, Size: 25},
	reflect.TypeFor[[26]byte]():       {T: abi.FixedBytesTy, Size: 26},
	reflect.TypeFor[[27]byte]():       {T: abi.FixedBytesTy, Size: 27},
	reflect.TypeFor[[28]byte]():       {T: abi.FixedBytesTy, Size: 28},
	reflect.TypeFor[[29]byte]():       {T: abi.FixedBytesTy, Size: 29},
	reflect.TypeFor[[30]byte]():       {T: abi.FixedBytesTy, Size: 30},
	reflect.TypeFor[[31]byte]():       {T: abi.FixedBytesTy, Size: 31},
	reflect.TypeFor[[32]byte]():       {T: abi.FixedBytesTy, Size: 32},
	reflect.TypeFor[common.Address](): {T: abi.AddressTy, Size: 20},
	reflect.TypeFor[common.Hash]():    {T: abi.FixedBytesTy, Size: 32},
	reflect.TypeFor[string]():         {T: abi.StringTy},
	reflect.TypeFor[[]byte]():         {T: abi.BytesTy},
	reflect.TypeFor[*big.Int]():       {T: abi.UintTy, Size: 256},
}

func toCamelCase(s string) string {
	if len(s) == 0 {
		return s
	}

	r := []rune(s)

	var prevUp bool
	for i := range len(r) {
		if i == 0 && unicode.IsUpper(r[i]) {
			prevUp = true
			r[i] = unicode.ToLower(r[i])
		} else if unicode.IsUpper(r[i]) {
			if prevUp {
				r[i] = unicode.ToLower(r[i])
			}
			prevUp = true
		} else {
			prevUp = false
		}
	}

	return string(r)
}
