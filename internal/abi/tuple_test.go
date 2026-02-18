package abi

import (
	"errors"
	"maps"
	"math/big"
	"reflect"
	"slices"
	"strconv"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/lmittmann/w3/internal"
)

func TestTupleMap(t *testing.T) {
	tests := []struct {
		Tuples     []any
		WantTuples map[string]reflect.Type
		WantErr    error
	}{
		{
			Tuples: []any{Tuple1{}},
			WantTuples: map[string]reflect.Type{
				"Tuple1": reflect.TypeFor[Tuple1](),
			},
		},
		{
			Tuples:  []any{Tuple1{}, Tuple1{}},
			WantErr: errors.New("duplicate tuple definition: Tuple1"),
		},
		{
			Tuples:  []any{123}, // int instead of struct
			WantErr: errors.New("expected struct, got int"),
		},
		{
			Tuples:  []any{"hello"}, // string instead of struct
			WantErr: errors.New("expected struct, got string"),
		},
		{
			Tuples: []any{Tuple1{}, Tuple2{}},
			WantTuples: map[string]reflect.Type{
				"Tuple1": reflect.TypeFor[Tuple1](),
				"Tuple2": reflect.TypeFor[Tuple2](),
			},
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			gotTuples, gotErr := tupleMap(test.Tuples...)
			if diff := cmp.Diff(test.WantErr, gotErr,
				internal.EquateErrors(),
			); diff != "" {
				t.Fatalf("Err (-want, +got):\n%s", diff)
			}

			want := slices.Sorted(maps.Keys(test.WantTuples))
			got := slices.Sorted(maps.Keys(gotTuples))
			if diff := cmp.Diff(want, got); diff != "" {
				t.Fatalf("Tuples (-want, +got):\n%s", diff)
			}
		})
	}
}

type Tuple1 struct {
	Arg0 *big.Int
}

type Tuple2 struct {
	Arg0 *big.Int
	Arg1 *big.Int `abitype:"int256"`
}

func TestTypeOfField(t *testing.T) {
	tests := []struct {
		Field   reflect.StructField
		Want    *abi.Type
		WantErr error
	}{
		// default types
		{
			Field: reflect.StructField{Name: "TestBool", Type: reflect.TypeFor[bool]()},
			Want:  &abi.Type{T: abi.BoolTy},
		},
		{
			Field: reflect.StructField{Name: "TestUint8", Type: reflect.TypeFor[uint8]()},
			Want:  &abi.Type{T: abi.UintTy, Size: 8},
		},
		{
			Field: reflect.StructField{Name: "TestUint16", Type: reflect.TypeFor[uint16]()},
			Want:  &abi.Type{T: abi.UintTy, Size: 16},
		},
		{
			Field: reflect.StructField{Name: "TestUint32", Type: reflect.TypeFor[uint32]()},
			Want:  &abi.Type{T: abi.UintTy, Size: 32},
		},
		{
			Field: reflect.StructField{Name: "TestUint64", Type: reflect.TypeFor[uint64]()},
			Want:  &abi.Type{T: abi.UintTy, Size: 64},
		},
		{
			Field: reflect.StructField{Name: "TestInt8", Type: reflect.TypeFor[int8]()},
			Want:  &abi.Type{T: abi.IntTy, Size: 8},
		},
		{
			Field: reflect.StructField{Name: "TestInt16", Type: reflect.TypeFor[int16]()},
			Want:  &abi.Type{T: abi.IntTy, Size: 16},
		},
		{
			Field: reflect.StructField{Name: "TestInt32", Type: reflect.TypeFor[int32]()},
			Want:  &abi.Type{T: abi.IntTy, Size: 32},
		},
		{
			Field: reflect.StructField{Name: "TestInt64", Type: reflect.TypeFor[int64]()},
			Want:  &abi.Type{T: abi.IntTy, Size: 64},
		},
		{
			Field: reflect.StructField{Name: "TestBigInt", Type: reflect.TypeFor[*big.Int]()},
			Want:  &abi.Type{T: abi.UintTy, Size: 256},
		},
		{
			Field: reflect.StructField{Name: "TestAddress", Type: reflect.TypeFor[common.Address]()},
			Want:  &abi.Type{T: abi.AddressTy, Size: 20},
		},
		{
			Field: reflect.StructField{Name: "TestHash", Type: reflect.TypeFor[common.Hash]()},
			Want:  &abi.Type{T: abi.FixedBytesTy, Size: 32},
		},
		{
			Field: reflect.StructField{Name: "TestBytes", Type: reflect.TypeFor[[]byte]()},
			Want:  &abi.Type{T: abi.BytesTy},
		},
		{
			Field: reflect.StructField{Name: "TestString", Type: reflect.TypeFor[string]()},
			Want:  &abi.Type{T: abi.StringTy},
		},

		// slice types
		{
			Field: reflect.StructField{Name: "TestBool", Type: reflect.TypeFor[[]bool]()},
			Want:  &abi.Type{T: abi.SliceTy, Elem: &abi.Type{T: abi.BoolTy}},
		},
		{
			Field: reflect.StructField{Name: "TestBool", Type: reflect.TypeFor[[1]bool]()},
			Want:  &abi.Type{T: abi.ArrayTy, Elem: &abi.Type{T: abi.BoolTy}, Size: 1},
		},
		{
			Field: reflect.StructField{Name: "TestBool", Type: reflect.TypeFor[[][]bool]()},
			Want:  &abi.Type{T: abi.SliceTy, Elem: &abi.Type{T: abi.SliceTy, Elem: &abi.Type{T: abi.BoolTy}}},
		},
		{
			Field: reflect.StructField{Name: "TestBool", Type: reflect.TypeFor[[1][]bool]()},
			Want:  &abi.Type{T: abi.ArrayTy, Elem: &abi.Type{T: abi.SliceTy, Elem: &abi.Type{T: abi.BoolTy}}, Size: 1},
		},
		{
			Field: reflect.StructField{Name: "TestBool", Type: reflect.TypeFor[[][1]bool]()},
			Want:  &abi.Type{T: abi.SliceTy, Elem: &abi.Type{T: abi.ArrayTy, Elem: &abi.Type{T: abi.BoolTy}, Size: 1}},
		},
		{
			Field: reflect.StructField{Name: "TestBool", Type: reflect.TypeFor[[1][1]bool]()},
			Want:  &abi.Type{T: abi.ArrayTy, Elem: &abi.Type{T: abi.ArrayTy, Elem: &abi.Type{T: abi.BoolTy}, Size: 1}, Size: 1},
		},

		// tagged types
		{
			Field: reflect.StructField{Name: "TestTagUint256", Type: reflect.TypeFor[*big.Int](), Tag: `abitype:"uint256"`},
			Want:  &abi.Type{T: abi.UintTy, Size: 256},
		},
		{
			Field: reflect.StructField{Name: "TestTagUint256", Type: reflect.TypeFor[*big.Int](), Tag: `abitype:"uint16"`},
			Want:  &abi.Type{T: abi.UintTy, Size: 16},
		},
		{
			Field: reflect.StructField{Name: "TestTagInt256", Type: reflect.TypeFor[*big.Int](), Tag: `abitype:"int256"`},
			Want:  &abi.Type{T: abi.IntTy, Size: 256},
		},
		{
			Field: reflect.StructField{Name: "TestTagAddress", Type: reflect.TypeFor[common.Address](), Tag: `abitype:"address"`},
			Want:  &abi.Type{T: abi.AddressTy, Size: 20},
		},
		{
			Field: reflect.StructField{Name: "TestTagBytes32", Type: reflect.TypeFor[[32]byte](), Tag: `abitype:"bytes32"`},
			Want:  &abi.Type{T: abi.FixedBytesTy, Size: 32},
		},
		{
			Field:   reflect.StructField{Name: "TestTagBytes32Hash", Type: reflect.TypeFor[common.Hash](), Tag: `abitype:"bytes32"`},
			WantErr: errors.New(`tagged type "bytes32" does not match type common.Hash`),
		},
		{
			Field:   reflect.StructField{Name: "TestUnknownTag", Type: reflect.TypeFor[*big.Int](), Tag: `abitype:"unknown"`},
			WantErr: errors.New(`unknown abi type "unknown"`),
		},
		{
			Field:   reflect.StructField{Name: "TestIncompatible", Type: reflect.TypeFor[string](), Tag: `abitype:"uint256"`},
			WantErr: errors.New(`tagged type "uint256" does not match type string`),
		},
		{
			Field:   reflect.StructField{Name: "TestUnsupported", Type: reflect.TypeFor[float64]()},
			WantErr: errors.New(`unknown type "float64"`),
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			got, err := typeOfField(test.Field)
			if diff := cmp.Diff(test.WantErr, err,
				internal.EquateErrors(),
			); diff != "" {
				t.Fatalf("Err: (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(test.Want, got,
				cmpopts.IgnoreUnexported(abi.Type{}),
			); diff != "" {
				t.Errorf("(-want +got):\n%s", diff)
			}
		})
	}
}

func TestToCamelCase(t *testing.T) {
	tests := []struct {
		Input string
		Want  string
	}{
		{"", ""},
		{"test", "test"},
		{"TEST", "test"},
		{"Test", "test"},
		{"testCase", "testCase"},
		{"TestCase", "testCase"},
		{"testID", "testId"},
		{"TestID", "testId"},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			got := toCamelCase(test.Input)
			if test.Want != got {
				t.Fatalf("want %q, got %q", test.Want, got)
			}
		})
	}
}
