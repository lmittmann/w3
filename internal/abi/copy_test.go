package abi

import (
	"errors"
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/google/go-cmp/cmp"
	"github.com/lmittmann/w3/internal"
)

func TestCopy(t *testing.T) {
	tests := []struct {
		Dst, Src any
		WantDst  any
		WantErr  error
	}{
		// uint256
		{
			Dst:     nil,
			Src:     big.NewInt(42),
			WantErr: errors.New("abi: decode nil"),
		},
		{
			Dst: func() *big.Int {
				var b *big.Int
				return b
			}(),
			Src:     big.NewInt(42),
			WantErr: errors.New("abi: decode nil *big.Int"),
		},
		{
			Dst:     big.Int{},
			Src:     big.NewInt(42),
			WantErr: errors.New("abi: decode non-pointer big.Int"),
		},
		{
			Dst:     new(big.Int),
			Src:     big.NewInt(42),
			WantDst: big.NewInt(42),
		},
		{
			Dst: func() **big.Int {
				var b *big.Int
				return &b
			}(),
			Src:     big.NewInt(42),
			WantDst: ptr(big.NewInt(42)),
		},
		{
			Dst:     new(big.Int),
			Src:     val(big.NewInt(42)),
			WantErr: errors.New("abi: can't assign big.Int to *big.Int"),
		},

		// uint
		{
			Dst:     ptr[uint](0),
			Src:     uint(42),
			WantDst: ptr[uint](42),
		},
		{
			Dst:     ptr[uint](0),
			Src:     ptr[uint](42),
			WantErr: errors.New("abi: unsupported src type *uint"),
		},
		{
			Dst:     ptr(ptr[uint](0)),
			Src:     uint(42),
			WantDst: ptr(ptr[uint](42)),
		},
		{
			Dst:     uint(0),
			Src:     uint(42),
			WantErr: errors.New("abi: decode non-pointer uint"),
		},
		{
			Dst:     ptr[uint32](0),
			Src:     uint64(42),
			WantErr: errors.New("abi: can't assign uint64 to *uint32"),
		},

		// bytes
		{
			Dst:     ptr(make([]byte, 0)),
			Src:     []byte{0xc0, 0xfe},
			WantDst: ptr([]byte{0xc0, 0xfe}),
		},

		// slices
		{
			Dst:     ptr(make([]common.Address, 0)),
			Src:     []common.Address{{1}, {2}, {3}},
			WantDst: ptr([]common.Address{{1}, {2}, {3}}),
		},
		{
			Dst: func() *[]common.Address {
				var s []common.Address
				return &s
			}(),
			Src:     []common.Address{{1}, {2}, {3}},
			WantDst: ptr([]common.Address{{1}, {2}, {3}}),
		},

		// arrays
		{
			Dst:     ptr([3]common.Address{}),
			Src:     [3]common.Address{{1}, {2}, {3}},
			WantDst: ptr([3]common.Address{{1}, {2}, {3}}),
		},
		{
			Dst: func() *[3]common.Address {
				var a [3]common.Address
				return &a
			}(),
			Src:     [3]common.Address{{1}, {2}, {3}},
			WantDst: ptr([3]common.Address{{1}, {2}, {3}}),
		},

		// 2d slices/arrays
		{
			Dst:     ptr(make([][]common.Address, 0)),
			Src:     [][]common.Address{{{1}, {2}}, {{2}, {3}}},
			WantDst: ptr([][]common.Address{{{1}, {2}}, {{2}, {3}}}),
		},
		{
			Dst: func() *[][]common.Address {
				var s [][]common.Address
				return &s
			}(),
			Src:     [][]common.Address{{{1}, {2}}, {{2}, {3}}},
			WantDst: ptr([][]common.Address{{{1}, {2}}, {{2}, {3}}}),
		},
		{
			Dst:     ptr(make([][2]common.Address, 0)),
			Src:     [][2]common.Address{{{1}, {2}}, {{2}, {3}}},
			WantDst: ptr([][2]common.Address{{{1}, {2}}, {{2}, {3}}}),
		},
		{
			Dst:     ptr([2][]common.Address{}),
			Src:     [2][]common.Address{{{1}, {2}}, {{2}, {3}}},
			WantDst: ptr([2][]common.Address{{{1}, {2}}, {{2}, {3}}}),
		},
		{
			Dst:     ptr([2][2]common.Address{}),
			Src:     [2][2]common.Address{{{1}, {2}}, {{2}, {3}}},
			WantDst: ptr([2][2]common.Address{{{1}, {2}}, {{2}, {3}}}),
		},

		// tuples
		{
			Dst: new(tuple0),
			Src: struct {
				Uint *big.Int
				Addr common.Address
			}{Uint: big.NewInt(1), Addr: common.Address{1}},
			WantDst: &tuple0{Uint: big.NewInt(1), Addr: common.Address{1}},
		},
		{
			Dst: ptr(new(tuple0)),
			Src: struct {
				Uint *big.Int
				Addr common.Address
			}{Uint: big.NewInt(1), Addr: common.Address{1}},
			WantDst: ptr(&tuple0{Uint: big.NewInt(1), Addr: common.Address{1}}),
		},
		{
			Dst: new(tuple0),
			Src: struct {
				Uint *big.Int
			}{Uint: big.NewInt(1)},
			WantDst: &tuple0{Uint: big.NewInt(1)},
		},
		{
			Dst: new(tuple1),
			Src: struct {
				Uint *big.Int
				Addr common.Address
			}{Uint: big.NewInt(1), Addr: common.Address{1}},
			WantDst: &tuple1{Uint: big.NewInt(1), XAddr: common.Address{1}},
		},
		{
			Dst: new(tuple2),
			Src: struct {
				Uint  *big.Int
				Tuple struct {
					Uint *big.Int
					Addr common.Address
				}
			}{Uint: big.NewInt(1), Tuple: struct {
				Uint *big.Int
				Addr common.Address
			}{Uint: big.NewInt(1), Addr: common.Address{1}}},
			WantDst: &tuple2{Uint: big.NewInt(1), Tuple: &tuple0{Uint: big.NewInt(1), Addr: common.Address{1}}},
		},

		// tuple with nested tuple slice
		{
			Dst: new(tuple3),
			Src: struct {
				Uint  *big.Int
				Tuple []struct {
					Uint *big.Int
					Addr common.Address
				}
			}{Uint: big.NewInt(1), Tuple: []struct {
				Uint *big.Int
				Addr common.Address
			}{
				{Uint: big.NewInt(1), Addr: common.Address{1}},
				{Uint: big.NewInt(2), Addr: common.Address{2}},
			}},
			WantDst: &tuple3{Uint: big.NewInt(1), Tuple: []*tuple0{
				{Uint: big.NewInt(1), Addr: common.Address{1}},
				{Uint: big.NewInt(2), Addr: common.Address{2}},
			}},
		},

		// tuple slice
		{
			Dst: ptr(make([]*tuple0, 0)),
			Src: []struct {
				Uint *big.Int
				Addr common.Address
			}{
				{Uint: big.NewInt(1), Addr: common.Address{1}},
				{Uint: big.NewInt(2), Addr: common.Address{2}},
			},
			WantDst: ptr([]*tuple0{
				{Uint: big.NewInt(1), Addr: common.Address{1}},
				{Uint: big.NewInt(2), Addr: common.Address{2}},
			}),
		},
		{
			Dst: func() *[]*tuple0 {
				var s []*tuple0
				return &s
			}(),
			Src: []struct {
				Uint *big.Int
				Addr common.Address
			}{
				{Uint: big.NewInt(1), Addr: common.Address{1}},
				{Uint: big.NewInt(2), Addr: common.Address{2}},
			},
			WantDst: ptr([]*tuple0{
				{Uint: big.NewInt(1), Addr: common.Address{1}},
				{Uint: big.NewInt(2), Addr: common.Address{2}},
			}),
		},

		// tuple array
		{
			Dst: ptr([2]*tuple0{}),
			Src: [2]struct {
				Uint *big.Int
				Addr common.Address
			}{
				{Uint: big.NewInt(1), Addr: common.Address{1}},
				{Uint: big.NewInt(2), Addr: common.Address{2}},
			},
			WantDst: ptr([2]*tuple0{
				{Uint: big.NewInt(1), Addr: common.Address{1}},
				{Uint: big.NewInt(2), Addr: common.Address{2}},
			}),
		},
		{
			Dst: func() *[2]*tuple0 {
				var s [2]*tuple0
				return &s
			}(),
			Src: [2]struct {
				Uint *big.Int
				Addr common.Address
			}{
				{Uint: big.NewInt(1), Addr: common.Address{1}},
				{Uint: big.NewInt(2), Addr: common.Address{2}},
			},
			WantDst: ptr([2]*tuple0{
				{Uint: big.NewInt(1), Addr: common.Address{1}},
				{Uint: big.NewInt(2), Addr: common.Address{2}},
			}),
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%d_%T_%T", i, test.Dst, test.Src), func(t *testing.T) {
			err := Copy(test.Dst, test.Src)
			if diff := cmp.Diff(test.WantErr, err,
				internal.EquateErrors(),
			); diff != "" {
				t.Fatalf("Err: (-want +got)\n%s", diff)
			} else if err != nil {
				return
			}

			if diff := cmp.Diff(test.WantDst, test.Dst,
				cmp.AllowUnexported(big.Int{}),
			); diff != "" {
				t.Fatalf("Dst: (-want +got)\n%s", diff)
			}
		})
	}
}

func ptr[T any](v T) *T { return &v }
func val[T any](v *T) T { return *v }

type tuple0 struct {
	Uint *big.Int
	Addr common.Address
}

type tuple1 struct {
	Uint  *big.Int
	XAddr common.Address `abi:"addr"`
}

type tuple2 struct {
	Uint  *big.Int
	Tuple *tuple0
}

type tuple3 struct {
	Uint  *big.Int
	Tuple []*tuple0
}
