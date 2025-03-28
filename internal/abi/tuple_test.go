package abi

import (
	"errors"
	"maps"
	"math/big"
	"reflect"
	"slices"
	"strconv"
	"testing"

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

type Tuple1 struct {
	Arg0 *big.Int
}

type Tuple2 struct {
	Arg0 *big.Int
	Arg1 *big.Int `abitype:"int256"`
}
