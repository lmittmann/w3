package abi

import (
	"bytes"
	"math/big"
	"reflect"
	"strconv"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/google/go-cmp/cmp"
)

func TestSignature(t *testing.T) {
	tests := []struct {
		Args          Arguments
		WantSignature string
	}{
		{
			Args:          Arguments{},
			WantSignature: "",
		},
		{
			Args:          Arguments{{Type: typeUint256}},
			WantSignature: "uint256",
		},
		{
			Args:          Arguments{{Type: abi.Type{Elem: &typeUint256, T: abi.SliceTy}}},
			WantSignature: "uint256[]",
		},
		{
			Args:          Arguments{{Type: abi.Type{Elem: &typeUint256, T: abi.ArrayTy, Size: 3}}},
			WantSignature: "uint256[3]",
		},
		{
			Args: Arguments{{
				Type: abi.Type{
					T:             abi.TupleTy,
					TupleElems:    []*abi.Type{&typeUint256},
					TupleRawNames: []string{"arg0"},
				},
			}},
			WantSignature: "(uint256)",
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			gotSignature := test.Args.Signature()
			if test.WantSignature != gotSignature {
				t.Fatalf("want %q, got %q", test.WantSignature, gotSignature)
			}
		})
	}
}

func TestSignatureWithName(t *testing.T) {
	tests := []struct {
		Arguments     Arguments
		Name          string
		WantSignature string
	}{
		{
			Arguments:     Arguments{},
			Name:          "func",
			WantSignature: "func()",
		},
		{
			Arguments:     Arguments{{Type: typeUint256}},
			Name:          "func",
			WantSignature: "func(uint256)",
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			gotSignature := test.Arguments.SignatureWithName(test.Name)
			if test.WantSignature != gotSignature {
				t.Fatalf("want %q, got %q", test.WantSignature, gotSignature)
			}
		})
	}
}

func TestEncode(t *testing.T) {
	tests := []struct {
		Arguments Arguments
		Args      []any
		WantData  []byte
	}{
		{
			Arguments: Arguments{{Type: typeUint256}},
			Args:      []any{big.NewInt(1)},
			WantData:  common.FromHex("0x0000000000000000000000000000000000000000000000000000000000000001"),
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			gotData, err := test.Arguments.Encode(test.Args...)
			if err != nil {
				t.Fatalf("Failed to encode args: %v", err)
			}
			if !bytes.Equal(test.WantData, gotData) {
				t.Fatalf("\nwant 0x%x\ngot  0x%x", test.WantData, gotData)
			}
		})
	}
}

func TestEncodeWithSelector(t *testing.T) {
	tests := []struct {
		Arguments Arguments
		Selector  [4]byte
		Args      []any
		WantData  []byte
	}{
		{
			Arguments: Arguments{{Type: typeUint256}},
			Selector:  [4]byte{0x12, 0x34, 0x56, 0x78},
			Args:      []any{big.NewInt(1)},
			WantData:  common.FromHex("0x123456780000000000000000000000000000000000000000000000000000000000000001"),
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			gotData, err := test.Arguments.EncodeWithSelector(test.Selector, test.Args...)
			if err != nil {
				t.Fatalf("Failed to encode args: %v", err)
			}

			if !bytes.Equal(test.WantData, gotData) {
				t.Fatalf("\nwant 0x%x\ngot  0x%x", test.WantData, gotData)
			}
		})
	}
}

func TestEncodeWithSignature(t *testing.T) {
	tests := []struct {
		Arguments Arguments
		Args      []any
		Signature string
		WantData  []byte
	}{
		{
			Arguments: Arguments{{Type: typeUint256}},
			Args:      []any{big.NewInt(1)},
			Signature: "func(uint256)",
			WantData:  common.FromHex("0x7f98a45e0000000000000000000000000000000000000000000000000000000000000001"),
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			gotData, err := test.Arguments.EncodeWithSignature(test.Signature, test.Args...)
			if err != nil {
				t.Fatalf("Failed to encode args: %v", err)
			}
			if !bytes.Equal(test.WantData, gotData) {
				t.Fatalf("\nwant 0x%x\ngot  0x%x", test.WantData, gotData)
			}
		})
	}
}

func TestDecode(t *testing.T) {
	type tuple struct {
		A bool
		B *big.Int
	}

	var (
		dataUintBool = common.FromHex("0x0000000000000000000000000000000000000000000000000000000000000001")
		argsUint     = Arguments{{Type: typeUint256}}
		argsBool     = Arguments{{Type: typeBool}}

		dataTuple = common.FromHex("0x00000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000007")
		argsTuple = Arguments{{Type: abi.Type{
			T:             abi.TupleTy,
			TupleElems:    []*abi.Type{&typeBool, &typeUint256},
			TupleRawNames: []string{"A", "_a"},
			TupleType: reflect.TypeFor[struct {
				A bool     `abi:"a"`
				B *big.Int `abi:"b"`
			}](),
		}}}

		dataBytes2 = common.FromHex("0xc0fe000000000000000000000000000000000000000000000000000000000000")
		argsBytes2 = Arguments{{Type: abi.Type{T: abi.FixedBytesTy, Size: 2}}}

		dataSlice = common.FromHex("0x000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000007")
		argsSlice = Arguments{{Type: abi.Type{T: abi.SliceTy, Size: 1, Elem: &typeUint256}}}
	)

	t.Run("set-ptr", func(t *testing.T) {
		var arg big.Int
		err := argsUint.Decode(dataUintBool, &arg)
		if err != nil {
			t.Fatalf("Failed to decode args: %v", err)
		}
		if want := big.NewInt(1); arg.Cmp(want) != 0 {
			t.Fatalf("want %v, got %v", want, arg)
		}
	})

	t.Run("set-ptr-of-ptr", func(t *testing.T) {
		var arg *big.Int
		err := argsUint.Decode(dataUintBool, &arg)
		if err != nil {
			t.Fatalf("Failed to decode args: %v", err)
		}
		if want := big.NewInt(1); arg.Cmp(want) != 0 {
			t.Fatalf("want %v, got %v", want, arg)
		}
	})

	t.Run("nil-val", func(t *testing.T) {
		var arg *big.Int
		err := argsUint.Decode(dataUintBool, arg)
		if want := "abi: decode nil *big.Int"; err == nil || want != err.Error() {
			t.Fatalf("want %v, got %v", want, err)
		}
	})

	t.Run("nil", func(t *testing.T) {
		if err := argsUint.Decode(dataUintBool, nil); err != nil {
			t.Fatalf("want nil, got %v", err)
		}
	})

	t.Run("non-ptr", func(t *testing.T) {
		var arg bool
		err := argsBool.Decode(dataUintBool, arg)
		if want := "abi: decode non-pointer bool"; err == nil || want != err.Error() {
			t.Fatalf("want %v, got %v", want, err)
		}
	})

	t.Run("set-bool-ptr", func(t *testing.T) {
		var arg bool
		err := argsBool.Decode(dataUintBool, &arg)
		if err != nil {
			t.Fatalf("Failed to decode args: %v", err)
		}
		if want := true; arg != want {
			t.Fatalf("want %v, got %v", want, arg)
		}
	})

	t.Run("set-bool-ptr-of-ptr", func(t *testing.T) {
		var arg *bool
		err := argsBool.Decode(dataUintBool, &arg)
		if err != nil {
			t.Fatalf("Failed to decode args: %v", err)
		}
		if want := true; *arg != want {
			t.Fatalf("want %v, got %v", want, arg)
		}
	})

	t.Run("set-bytes2", func(t *testing.T) {
		var arg [2]byte
		err := argsBytes2.Decode(dataBytes2, &arg)
		if err != nil {
			t.Fatalf("Failed to decode args: %v", err)
		}
		want := [2]byte{0xc0, 0xfe}
		if diff := cmp.Diff(want, arg); diff != "" {
			t.Fatalf("(-want, +got)\n%s", diff)
		}
	})

	t.Run("set-tuple-ptr", func(t *testing.T) {
		var arg tuple
		err := argsTuple.Decode(dataTuple, &arg)
		if err != nil {
			t.Fatalf("Failed to decode args: %v", err)
		}
		want := tuple{A: true, B: big.NewInt(7)}
		if diff := cmp.Diff(want, arg, cmp.AllowUnexported(big.Int{})); diff != "" {
			t.Fatalf("(-want, +got)\n%s", diff)
		}
	})

	t.Run("set-tuple-ptr-of-ptr", func(t *testing.T) {
		var arg *tuple
		err := argsTuple.Decode(dataTuple, &arg)
		if err != nil {
			t.Fatalf("Failed to decode args: %v", err)
		}
		want := &tuple{A: true, B: big.NewInt(7)}
		if diff := cmp.Diff(want, arg, cmp.AllowUnexported(big.Int{})); diff != "" {
			t.Fatalf("(-want, +got)\n%s", diff)
		}
	})

	t.Run("set-slice", func(t *testing.T) {
		var arg []*big.Int
		err := argsSlice.Decode(dataSlice, &arg)
		if err != nil {
			t.Fatalf("Failed to decode args: %v", err)
		}
		want := []*big.Int{big.NewInt(7)}
		if diff := cmp.Diff(want, arg, cmp.AllowUnexported(big.Int{})); diff != "" {
			t.Fatalf("(-want, +got)\n%s", diff)
		}
	})
}

func TestTypeToString(t *testing.T) {
	tests := []struct {
		Type *abi.Type
		Want string
	}{
		{Type: &abi.Type{T: abi.BoolTy}, Want: "bool"},
		{Type: &abi.Type{T: abi.SliceTy, Elem: &abi.Type{T: abi.BoolTy}}, Want: "bool[]"},
		{Type: &abi.Type{Size: 2, T: abi.ArrayTy, Elem: &abi.Type{T: abi.BoolTy}}, Want: "bool[2]"},
		{Type: &abi.Type{T: abi.SliceTy, Elem: &abi.Type{T: abi.ArrayTy, Size: 2, Elem: &abi.Type{T: abi.BoolTy}}}, Want: "bool[2][]"},
		{Type: &abi.Type{T: abi.SliceTy, Elem: &abi.Type{T: abi.SliceTy, Elem: &abi.Type{T: abi.BoolTy}}}, Want: "bool[][]"},
		{Type: &abi.Type{T: abi.ArrayTy, Size: 2, Elem: &abi.Type{T: abi.SliceTy, Elem: &abi.Type{T: abi.BoolTy}}}, Want: "bool[][2]"},
		{Type: &abi.Type{T: abi.ArrayTy, Size: 2, Elem: &abi.Type{T: abi.ArrayTy, Size: 2, Elem: &abi.Type{T: abi.BoolTy}}}, Want: "bool[2][2]"},
		{Type: &abi.Type{T: abi.ArrayTy, Size: 2, Elem: &abi.Type{T: abi.SliceTy, Elem: &abi.Type{T: abi.ArrayTy, Size: 2, Elem: &abi.Type{T: abi.BoolTy}}}}, Want: "bool[2][][2]"},
		{Type: &abi.Type{T: abi.ArrayTy, Size: 2, Elem: &abi.Type{T: abi.ArrayTy, Size: 2, Elem: &abi.Type{T: abi.ArrayTy, Size: 2, Elem: &abi.Type{T: abi.BoolTy}}}}, Want: "bool[2][2][2]"},
		{Type: &abi.Type{T: abi.SliceTy, Elem: &abi.Type{T: abi.SliceTy, Elem: &abi.Type{T: abi.SliceTy, Elem: &abi.Type{T: abi.BoolTy}}}}, Want: "bool[][][]"},
		{Type: &abi.Type{T: abi.SliceTy, Elem: &abi.Type{T: abi.ArrayTy, Size: 2, Elem: &abi.Type{T: abi.SliceTy, Elem: &abi.Type{T: abi.BoolTy}}}}, Want: "bool[][2][]"},
		{Type: &abi.Type{Size: 8, T: abi.IntTy}, Want: "int8"},
		{Type: &abi.Type{Size: 16, T: abi.IntTy}, Want: "int16"},
		{Type: &abi.Type{Size: 32, T: abi.IntTy}, Want: "int32"},
		{Type: &abi.Type{Size: 64, T: abi.IntTy}, Want: "int64"},
		{Type: &abi.Type{Size: 256, T: abi.IntTy}, Want: "int256"},
		{Type: &abi.Type{T: abi.SliceTy, Elem: &abi.Type{Size: 8, T: abi.IntTy}}, Want: "int8[]"},
		{Type: &abi.Type{T: abi.ArrayTy, Size: 2, Elem: &abi.Type{Size: 8, T: abi.IntTy}}, Want: "int8[2]"},
		{Type: &abi.Type{T: abi.SliceTy, Elem: &abi.Type{Size: 16, T: abi.IntTy}}, Want: "int16[]"},
		{Type: &abi.Type{Size: 2, T: abi.ArrayTy, Elem: &abi.Type{Size: 16, T: abi.IntTy}}, Want: "int16[2]"},
		{Type: &abi.Type{T: abi.SliceTy, Elem: &abi.Type{Size: 32, T: abi.IntTy}}, Want: "int32[]"},
		{Type: &abi.Type{T: abi.ArrayTy, Size: 2, Elem: &abi.Type{Size: 32, T: abi.IntTy}}, Want: "int32[2]"},
		{Type: &abi.Type{T: abi.SliceTy, Elem: &abi.Type{Size: 64, T: abi.IntTy}}, Want: "int64[]"},
		{Type: &abi.Type{T: abi.ArrayTy, Size: 2, Elem: &abi.Type{Size: 64, T: abi.IntTy}}, Want: "int64[2]"},
		{Type: &abi.Type{T: abi.SliceTy, Elem: &abi.Type{Size: 256, T: abi.IntTy}}, Want: "int256[]"},
		{Type: &abi.Type{T: abi.ArrayTy, Size: 2, Elem: &abi.Type{Size: 256, T: abi.IntTy}}, Want: "int256[2]"},
		{Type: &abi.Type{Size: 8, T: abi.UintTy}, Want: "uint8"},
		{Type: &abi.Type{Size: 16, T: abi.UintTy}, Want: "uint16"},
		{Type: &abi.Type{Size: 32, T: abi.UintTy}, Want: "uint32"},
		{Type: &abi.Type{Size: 64, T: abi.UintTy}, Want: "uint64"},
		{Type: &abi.Type{Size: 256, T: abi.UintTy}, Want: "uint256"},
		{Type: &abi.Type{T: abi.SliceTy, Elem: &abi.Type{Size: 8, T: abi.UintTy}}, Want: "uint8[]"},
		{Type: &abi.Type{T: abi.ArrayTy, Size: 2, Elem: &abi.Type{Size: 8, T: abi.UintTy}}, Want: "uint8[2]"},
		{Type: &abi.Type{T: abi.SliceTy, Elem: &abi.Type{Size: 16, T: abi.UintTy}}, Want: "uint16[]"},
		{Type: &abi.Type{T: abi.ArrayTy, Size: 2, Elem: &abi.Type{Size: 16, T: abi.UintTy}}, Want: "uint16[2]"},
		{Type: &abi.Type{T: abi.SliceTy, Elem: &abi.Type{Size: 32, T: abi.UintTy}}, Want: "uint32[]"},
		{Type: &abi.Type{T: abi.ArrayTy, Size: 2, Elem: &abi.Type{Size: 32, T: abi.UintTy}}, Want: "uint32[2]"},
		{Type: &abi.Type{T: abi.SliceTy, Elem: &abi.Type{Size: 64, T: abi.UintTy}}, Want: "uint64[]"},
		{Type: &abi.Type{T: abi.ArrayTy, Size: 2, Elem: &abi.Type{Size: 64, T: abi.UintTy}}, Want: "uint64[2]"},
		{Type: &abi.Type{T: abi.SliceTy, Elem: &abi.Type{Size: 256, T: abi.UintTy}}, Want: "uint256[]"},
		{Type: &abi.Type{T: abi.ArrayTy, Size: 2, Elem: &abi.Type{Size: 256, T: abi.UintTy}}, Want: "uint256[2]"},
		{Type: &abi.Type{T: abi.FixedBytesTy, Size: 32}, Want: "bytes32"},
		{Type: &abi.Type{T: abi.SliceTy, Elem: &abi.Type{T: abi.BytesTy}}, Want: "bytes[]"},
		{Type: &abi.Type{T: abi.ArrayTy, Size: 2, Elem: &abi.Type{T: abi.BytesTy}}, Want: "bytes[2]"},
		{Type: &abi.Type{T: abi.SliceTy, Elem: &abi.Type{T: abi.FixedBytesTy, Size: 32}}, Want: "bytes32[]"},
		{Type: &abi.Type{T: abi.ArrayTy, Size: 2, Elem: &abi.Type{T: abi.FixedBytesTy, Size: 32}}, Want: "bytes32[2]"},
		{Type: &abi.Type{T: abi.HashTy, Size: 20}, Want: "hash"},
		{Type: &abi.Type{T: abi.StringTy}, Want: "string"},
		{Type: &abi.Type{T: abi.SliceTy, Elem: &abi.Type{T: abi.StringTy}}, Want: "string[]"},
		{Type: &abi.Type{T: abi.ArrayTy, Size: 2, Elem: &abi.Type{T: abi.StringTy}}, Want: "string[2]"},
		{Type: &abi.Type{Size: 20, T: abi.AddressTy}, Want: "address"},
		{Type: &abi.Type{T: abi.SliceTy, Elem: &abi.Type{Size: 20, T: abi.AddressTy}}, Want: "address[]"},
		{Type: &abi.Type{T: abi.ArrayTy, Size: 2, Elem: &abi.Type{Size: 20, T: abi.AddressTy}}, Want: "address[2]"},
		{Type: &abi.Type{T: abi.TupleTy, TupleElems: []*abi.Type{&typeUint256, &typeUint256}}, Want: "(uint256,uint256)"},
		{Type: &abi.Type{T: abi.TupleTy, TupleElems: []*abi.Type{{T: abi.TupleTy, TupleElems: []*abi.Type{&typeUint256, &typeUint256}}, &typeUint256}}, Want: "((uint256,uint256),uint256)"},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			got := typeToString(test.Type)
			if test.Want != got {
				t.Fatalf("want %q, got %q", test.Want, got)
			}
		})
	}
}

var (
	big1 = big.NewInt(1)
	hex1 = common.FromHex("0x0000000000000000000000000000000000000000000000000000000000000001")
)

func TestTypeMapping(t *testing.T) {
	tests := []struct {
		RAWType string
		Data    []byte
		Arg     any
		Want    any
	}{
		{RAWType: "bool", Data: hex1, Arg: new(bool), Want: ptr(true)},

		{RAWType: "int8", Data: hex1, Arg: new(int8), Want: ptr[int8](1)},
		{RAWType: "int16", Data: hex1, Arg: new(int16), Want: ptr[int16](1)},
		{RAWType: "int24", Data: hex1, Arg: new(big.Int), Want: big1},
		{RAWType: "int32", Data: hex1, Arg: new(int32), Want: ptr[int32](1)},
		{RAWType: "int40", Data: hex1, Arg: new(big.Int), Want: big1},
		{RAWType: "int48", Data: hex1, Arg: new(big.Int), Want: big1},
		{RAWType: "int56", Data: hex1, Arg: new(big.Int), Want: big1},
		{RAWType: "int64", Data: hex1, Arg: new(int64), Want: ptr[int64](1)},
		{RAWType: "int72", Data: hex1, Arg: new(big.Int), Want: big1},
		{RAWType: "int80", Data: hex1, Arg: new(big.Int), Want: big1},
		{RAWType: "int88", Data: hex1, Arg: new(big.Int), Want: big1},
		{RAWType: "int96", Data: hex1, Arg: new(big.Int), Want: big1},
		{RAWType: "int104", Data: hex1, Arg: new(big.Int), Want: big1},
		{RAWType: "int112", Data: hex1, Arg: new(big.Int), Want: big1},
		{RAWType: "int120", Data: hex1, Arg: new(big.Int), Want: big1},
		{RAWType: "int128", Data: hex1, Arg: new(big.Int), Want: big1},
		{RAWType: "int136", Data: hex1, Arg: new(big.Int), Want: big1},
		{RAWType: "int144", Data: hex1, Arg: new(big.Int), Want: big1},
		{RAWType: "int152", Data: hex1, Arg: new(big.Int), Want: big1},
		{RAWType: "int160", Data: hex1, Arg: new(big.Int), Want: big1},
		{RAWType: "int168", Data: hex1, Arg: new(big.Int), Want: big1},
		{RAWType: "int176", Data: hex1, Arg: new(big.Int), Want: big1},
		{RAWType: "int184", Data: hex1, Arg: new(big.Int), Want: big1},
		{RAWType: "int192", Data: hex1, Arg: new(big.Int), Want: big1},
		{RAWType: "int200", Data: hex1, Arg: new(big.Int), Want: big1},
		{RAWType: "int208", Data: hex1, Arg: new(big.Int), Want: big1},
		{RAWType: "int216", Data: hex1, Arg: new(big.Int), Want: big1},
		{RAWType: "int224", Data: hex1, Arg: new(big.Int), Want: big1},
		{RAWType: "int232", Data: hex1, Arg: new(big.Int), Want: big1},
		{RAWType: "int240", Data: hex1, Arg: new(big.Int), Want: big1},
		{RAWType: "int248", Data: hex1, Arg: new(big.Int), Want: big1},
		{RAWType: "int256", Data: hex1, Arg: new(big.Int), Want: big1},
		{RAWType: "int", Data: hex1, Arg: new(big.Int), Want: big1},

		{RAWType: "uint8", Data: hex1, Arg: new(uint8), Want: ptr[uint8](1)},
		{RAWType: "uint16", Data: hex1, Arg: new(uint16), Want: ptr[uint16](1)},
		{RAWType: "uint24", Data: hex1, Arg: new(big.Int), Want: big1},
		{RAWType: "uint32", Data: hex1, Arg: new(uint32), Want: ptr[uint32](1)},
		{RAWType: "uint40", Data: hex1, Arg: new(big.Int), Want: big1},
		{RAWType: "uint48", Data: hex1, Arg: new(big.Int), Want: big1},
		{RAWType: "uint56", Data: hex1, Arg: new(big.Int), Want: big1},
		{RAWType: "uint64", Data: hex1, Arg: new(uint64), Want: ptr[uint64](1)},
		{RAWType: "uint72", Data: hex1, Arg: new(big.Int), Want: big1},
		{RAWType: "uint80", Data: hex1, Arg: new(big.Int), Want: big1},
		{RAWType: "uint88", Data: hex1, Arg: new(big.Int), Want: big1},
		{RAWType: "uint96", Data: hex1, Arg: new(big.Int), Want: big1},
		{RAWType: "uint104", Data: hex1, Arg: new(big.Int), Want: big1},
		{RAWType: "uint112", Data: hex1, Arg: new(big.Int), Want: big1},
		{RAWType: "uint120", Data: hex1, Arg: new(big.Int), Want: big1},
		{RAWType: "uint128", Data: hex1, Arg: new(big.Int), Want: big1},
		{RAWType: "uint136", Data: hex1, Arg: new(big.Int), Want: big1},
		{RAWType: "uint144", Data: hex1, Arg: new(big.Int), Want: big1},
		{RAWType: "uint152", Data: hex1, Arg: new(big.Int), Want: big1},
		{RAWType: "uint160", Data: hex1, Arg: new(big.Int), Want: big1},
		{RAWType: "uint168", Data: hex1, Arg: new(big.Int), Want: big1},
		{RAWType: "uint176", Data: hex1, Arg: new(big.Int), Want: big1},
		{RAWType: "uint184", Data: hex1, Arg: new(big.Int), Want: big1},
		{RAWType: "uint192", Data: hex1, Arg: new(big.Int), Want: big1},
		{RAWType: "uint200", Data: hex1, Arg: new(big.Int), Want: big1},
		{RAWType: "uint208", Data: hex1, Arg: new(big.Int), Want: big1},
		{RAWType: "uint216", Data: hex1, Arg: new(big.Int), Want: big1},
		{RAWType: "uint224", Data: hex1, Arg: new(big.Int), Want: big1},
		{RAWType: "uint232", Data: hex1, Arg: new(big.Int), Want: big1},
		{RAWType: "uint240", Data: hex1, Arg: new(big.Int), Want: big1},
		{RAWType: "uint248", Data: hex1, Arg: new(big.Int), Want: big1},
		{RAWType: "uint256", Data: hex1, Arg: new(big.Int), Want: big1},
		{RAWType: "uint", Data: hex1, Arg: new(big.Int), Want: big1},

		{RAWType: "bytes1", Data: hex1, Arg: new([1]byte), Want: &[1]byte{}},
		{RAWType: "bytes2", Data: hex1, Arg: new([2]byte), Want: &[2]byte{}},
		{RAWType: "bytes3", Data: hex1, Arg: new([3]byte), Want: &[3]byte{}},
		{RAWType: "bytes4", Data: hex1, Arg: new([4]byte), Want: &[4]byte{}},
		{RAWType: "bytes5", Data: hex1, Arg: new([5]byte), Want: &[5]byte{}},
		{RAWType: "bytes6", Data: hex1, Arg: new([6]byte), Want: &[6]byte{}},
		{RAWType: "bytes7", Data: hex1, Arg: new([7]byte), Want: &[7]byte{}},
		{RAWType: "bytes8", Data: hex1, Arg: new([8]byte), Want: &[8]byte{}},
		{RAWType: "bytes9", Data: hex1, Arg: new([9]byte), Want: &[9]byte{}},
		{RAWType: "bytes10", Data: hex1, Arg: new([10]byte), Want: &[10]byte{}},
		{RAWType: "bytes11", Data: hex1, Arg: new([11]byte), Want: &[11]byte{}},
		{RAWType: "bytes12", Data: hex1, Arg: new([12]byte), Want: &[12]byte{}},
		{RAWType: "bytes13", Data: hex1, Arg: new([13]byte), Want: &[13]byte{}},
		{RAWType: "bytes14", Data: hex1, Arg: new([14]byte), Want: &[14]byte{}},
		{RAWType: "bytes15", Data: hex1, Arg: new([15]byte), Want: &[15]byte{}},
		{RAWType: "bytes16", Data: hex1, Arg: new([16]byte), Want: &[16]byte{}},
		{RAWType: "bytes17", Data: hex1, Arg: new([17]byte), Want: &[17]byte{}},
		{RAWType: "bytes18", Data: hex1, Arg: new([18]byte), Want: &[18]byte{}},
		{RAWType: "bytes19", Data: hex1, Arg: new([19]byte), Want: &[19]byte{}},
		{RAWType: "bytes20", Data: hex1, Arg: new([20]byte), Want: &[20]byte{}},
		{RAWType: "address", Data: hex1, Arg: new(common.Address), Want: &common.Address{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}},
		{RAWType: "bytes21", Data: hex1, Arg: new([21]byte), Want: &[21]byte{}},
		{RAWType: "bytes22", Data: hex1, Arg: new([22]byte), Want: &[22]byte{}},
		{RAWType: "bytes23", Data: hex1, Arg: new([23]byte), Want: &[23]byte{}},
		{RAWType: "bytes24", Data: hex1, Arg: new([24]byte), Want: &[24]byte{}},
		{RAWType: "bytes25", Data: hex1, Arg: new([25]byte), Want: &[25]byte{}},
		{RAWType: "bytes26", Data: hex1, Arg: new([26]byte), Want: &[26]byte{}},
		{RAWType: "bytes27", Data: hex1, Arg: new([27]byte), Want: &[27]byte{}},
		{RAWType: "bytes28", Data: hex1, Arg: new([28]byte), Want: &[28]byte{}},
		{RAWType: "bytes29", Data: hex1, Arg: new([29]byte), Want: &[29]byte{}},
		{RAWType: "bytes30", Data: hex1, Arg: new([30]byte), Want: &[30]byte{}},
		{RAWType: "bytes31", Data: hex1, Arg: new([31]byte), Want: &[31]byte{}},
		{RAWType: "bytes32", Data: hex1, Arg: new([32]byte), Want: &[32]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}},
		{RAWType: "bytes32", Data: hex1, Arg: new(common.Hash), Want: &common.Hash{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			args, err := Parse(test.RAWType)
			if err != nil {
				t.Fatalf("Failed to parse type: %v", err)
			}

			if err := args.Decode(test.Data, test.Arg); err != nil {
				t.Fatalf("Failed to decode args: %v", err)
			}

			if diff := cmp.Diff(test.Want, test.Arg, cmp.AllowUnexported(big.Int{})); diff != "" {
				t.Fatalf("(-want, +got)\n%s", diff)
			}
		})
	}
}
