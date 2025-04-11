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
	"github.com/google/go-cmp/cmp"
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
				"Tuple1": reflect.TypeOf(Tuple1{}),
			},
		},
		{
			Tuples:  []any{Tuple1{}, Tuple1{}},
			WantErr: errDuplicateTuple,
		},
		{
			Tuples: []any{Tuple1{}, Tuple2{}},
			WantTuples: map[string]reflect.Type{
				"Tuple1": reflect.TypeOf(Tuple1{}),
				"Tuple2": reflect.TypeOf(Tuple2{}),
			},
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			gotTuples, gotErr := tupleMap(test.Tuples...)
			if !errors.Is(gotErr, test.WantErr) {
				t.Fatalf("Err: want %v, got %v", test.WantErr, gotErr)
			}

			want := slices.Sorted(maps.Keys(test.WantTuples))
			got := slices.Sorted(maps.Keys(gotTuples))
			if diff := cmp.Diff(want, got); diff != "" {
				t.Fatalf("Tuples (-want, +got):\n%s", diff)
			}
		})
	}
}

func TestTupleX(t *testing.T) {
	tuple := Tuple2{}
	tupleType := reflect.TypeOf(tuple)

	for i := 0; i < tupleType.NumField(); i++ {
		field := tupleType.Field(i)
		t.Logf("Field: %s, Type: %s", field.Name, field.Type)

		// Also log any abitype tag if present
		if tag, ok := field.Tag.Lookup("abitype"); ok {
			t.Logf("Field: %s has abitype tag: %s", field.Name, tag)
		}
	}

	typeUint256 := abi.Type{T: abi.UintTy, Size: 256}
	typeInt256 := abi.Type{T: abi.IntTy, Size: 256}
	arg := abi.Argument{
		Name: "tuple",
		Type: abi.Type{
			T:             abi.TupleTy,
			TupleElems:    []*abi.Type{&typeUint256, &typeInt256},
			TupleRawNames: []string{"arg0", "arg1"},
			TupleType:     reflect.TypeFor[Tuple2](),
		},
	}

	args := abi.Arguments{arg}

	packed, err := args.Pack(Tuple2{
		Arg0: big.NewInt(1),
		Arg1: big.NewInt(2),
	})
	if err != nil {
		t.Fatalf("Pack: %v", err)
	}

	unpacked, err := args.Unpack(packed)
	if err != nil {
		t.Fatalf("Unpack: %v", err)
	}

	t.Logf("Packed: %x", packed)
	t.Logf("Unpacked: %+v %T", unpacked, unpacked[0])
}

type Tuple1 struct {
	Arg0 *big.Int
}

type Tuple2 struct {
	Arg0 *big.Int
	Arg1 *big.Int `abitype:"int256"`
}

func TestConvert(t *testing.T) {
	valSrc := struct {
		Arg0 *big.Int
		Arg1 *big.Int
	}{
		Arg0: big.NewInt(123),
		Arg1: big.NewInt(456),
	}

	valDst := abi.ConvertType(valSrc, Tuple2{})
	t.Logf("valDst: %+v", valDst)
}
