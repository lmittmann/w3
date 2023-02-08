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
	t.Parallel()

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
	t.Parallel()

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
	t.Parallel()

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
	t.Parallel()

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
	t.Parallel()

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
			TupleType: reflect.TypeOf(struct {
				A bool     `abi:"a"`
				B *big.Int `abi:"b"`
			}{}),
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
		if want := "abi: decode nil"; err == nil || want != err.Error() {
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
	t.Parallel()

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
