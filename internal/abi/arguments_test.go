package abi

import (
	"bytes"
	"math/big"
	"strconv"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
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
	t.Parallel()

	tests := []struct {
		Arguments Arguments
		Data      []byte
		GotArgs   []any
		WantArgs  []any
	}{
		{
			Arguments: Arguments{{Type: typeUint256}},
			Data:      common.FromHex("0x0000000000000000000000000000000000000000000000000000000000000001"),
			GotArgs:   []any{new(big.Int)},
			WantArgs:  []any{big.NewInt(1)},
		},
		{
			Arguments: Arguments{{Type: typeAddress}},
			Data:      common.FromHex("0x000000000000000000000000000000000000000000000000000000000000c0fe"),
			GotArgs:   []any{new(common.Address)},
			WantArgs:  []any{addrPtr(common.HexToAddress("0x000000000000000000000000000000000000c0Fe"))},
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			err := test.Arguments.Decode(test.Data, test.GotArgs...)
			if err != nil {
				t.Fatalf("Failed to decode args: %v", err)
			}
			if diff := cmp.Diff(test.WantArgs, test.GotArgs,
				cmp.AllowUnexported(big.Int{}),
			); diff != "" {
				t.Fatalf("(-want, +got)\n%s", diff)
			}
		})
	}
}

func TestTypeToString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		Type abi.Type
		Want string
	}{
		{Type: abi.Type{T: abi.BoolTy}, Want: "bool"},
		{Type: abi.Type{T: abi.SliceTy, Elem: &abi.Type{T: abi.BoolTy}}, Want: "bool[]"},
		{Type: abi.Type{Size: 2, T: abi.ArrayTy, Elem: &abi.Type{T: abi.BoolTy}}, Want: "bool[2]"},
		{Type: abi.Type{T: abi.SliceTy, Elem: &abi.Type{T: abi.ArrayTy, Size: 2, Elem: &abi.Type{T: abi.BoolTy}}}, Want: "bool[2][]"},
		{Type: abi.Type{T: abi.SliceTy, Elem: &abi.Type{T: abi.SliceTy, Elem: &abi.Type{T: abi.BoolTy}}}, Want: "bool[][]"},
		{Type: abi.Type{T: abi.ArrayTy, Size: 2, Elem: &abi.Type{T: abi.SliceTy, Elem: &abi.Type{T: abi.BoolTy}}}, Want: "bool[][2]"},
		{Type: abi.Type{T: abi.ArrayTy, Size: 2, Elem: &abi.Type{T: abi.ArrayTy, Size: 2, Elem: &abi.Type{T: abi.BoolTy}}}, Want: "bool[2][2]"},
		{Type: abi.Type{T: abi.ArrayTy, Size: 2, Elem: &abi.Type{T: abi.SliceTy, Elem: &abi.Type{T: abi.ArrayTy, Size: 2, Elem: &abi.Type{T: abi.BoolTy}}}}, Want: "bool[2][][2]"},
		{Type: abi.Type{T: abi.ArrayTy, Size: 2, Elem: &abi.Type{T: abi.ArrayTy, Size: 2, Elem: &abi.Type{T: abi.ArrayTy, Size: 2, Elem: &abi.Type{T: abi.BoolTy}}}}, Want: "bool[2][2][2]"},
		{Type: abi.Type{T: abi.SliceTy, Elem: &abi.Type{T: abi.SliceTy, Elem: &abi.Type{T: abi.SliceTy, Elem: &abi.Type{T: abi.BoolTy}}}}, Want: "bool[][][]"},
		{Type: abi.Type{T: abi.SliceTy, Elem: &abi.Type{T: abi.ArrayTy, Size: 2, Elem: &abi.Type{T: abi.SliceTy, Elem: &abi.Type{T: abi.BoolTy}}}}, Want: "bool[][2][]"},
		{Type: abi.Type{Size: 8, T: abi.IntTy}, Want: "int8"},
		{Type: abi.Type{Size: 16, T: abi.IntTy}, Want: "int16"},
		{Type: abi.Type{Size: 32, T: abi.IntTy}, Want: "int32"},
		{Type: abi.Type{Size: 64, T: abi.IntTy}, Want: "int64"},
		{Type: abi.Type{Size: 256, T: abi.IntTy}, Want: "int256"},
		{Type: abi.Type{T: abi.SliceTy, Elem: &abi.Type{Size: 8, T: abi.IntTy}}, Want: "int8[]"},
		{Type: abi.Type{T: abi.ArrayTy, Size: 2, Elem: &abi.Type{Size: 8, T: abi.IntTy}}, Want: "int8[2]"},
		{Type: abi.Type{T: abi.SliceTy, Elem: &abi.Type{Size: 16, T: abi.IntTy}}, Want: "int16[]"},
		{Type: abi.Type{Size: 2, T: abi.ArrayTy, Elem: &abi.Type{Size: 16, T: abi.IntTy}}, Want: "int16[2]"},
		{Type: abi.Type{T: abi.SliceTy, Elem: &abi.Type{Size: 32, T: abi.IntTy}}, Want: "int32[]"},
		{Type: abi.Type{T: abi.ArrayTy, Size: 2, Elem: &abi.Type{Size: 32, T: abi.IntTy}}, Want: "int32[2]"},
		{Type: abi.Type{T: abi.SliceTy, Elem: &abi.Type{Size: 64, T: abi.IntTy}}, Want: "int64[]"},
		{Type: abi.Type{T: abi.ArrayTy, Size: 2, Elem: &abi.Type{Size: 64, T: abi.IntTy}}, Want: "int64[2]"},
		{Type: abi.Type{T: abi.SliceTy, Elem: &abi.Type{Size: 256, T: abi.IntTy}}, Want: "int256[]"},
		{Type: abi.Type{T: abi.ArrayTy, Size: 2, Elem: &abi.Type{Size: 256, T: abi.IntTy}}, Want: "int256[2]"},
		{Type: abi.Type{Size: 8, T: abi.UintTy}, Want: "uint8"},
		{Type: abi.Type{Size: 16, T: abi.UintTy}, Want: "uint16"},
		{Type: abi.Type{Size: 32, T: abi.UintTy}, Want: "uint32"},
		{Type: abi.Type{Size: 64, T: abi.UintTy}, Want: "uint64"},
		{Type: abi.Type{Size: 256, T: abi.UintTy}, Want: "uint256"},
		{Type: abi.Type{T: abi.SliceTy, Elem: &abi.Type{Size: 8, T: abi.UintTy}}, Want: "uint8[]"},
		{Type: abi.Type{T: abi.ArrayTy, Size: 2, Elem: &abi.Type{Size: 8, T: abi.UintTy}}, Want: "uint8[2]"},
		{Type: abi.Type{T: abi.SliceTy, Elem: &abi.Type{Size: 16, T: abi.UintTy}}, Want: "uint16[]"},
		{Type: abi.Type{T: abi.ArrayTy, Size: 2, Elem: &abi.Type{Size: 16, T: abi.UintTy}}, Want: "uint16[2]"},
		{Type: abi.Type{T: abi.SliceTy, Elem: &abi.Type{Size: 32, T: abi.UintTy}}, Want: "uint32[]"},
		{Type: abi.Type{T: abi.ArrayTy, Size: 2, Elem: &abi.Type{Size: 32, T: abi.UintTy}}, Want: "uint32[2]"},
		{Type: abi.Type{T: abi.SliceTy, Elem: &abi.Type{Size: 64, T: abi.UintTy}}, Want: "uint64[]"},
		{Type: abi.Type{T: abi.ArrayTy, Size: 2, Elem: &abi.Type{Size: 64, T: abi.UintTy}}, Want: "uint64[2]"},
		{Type: abi.Type{T: abi.SliceTy, Elem: &abi.Type{Size: 256, T: abi.UintTy}}, Want: "uint256[]"},
		{Type: abi.Type{T: abi.ArrayTy, Size: 2, Elem: &abi.Type{Size: 256, T: abi.UintTy}}, Want: "uint256[2]"},
		{Type: abi.Type{T: abi.FixedBytesTy, Size: 32}, Want: "bytes32"},
		{Type: abi.Type{T: abi.SliceTy, Elem: &abi.Type{T: abi.BytesTy}}, Want: "bytes[]"},
		{Type: abi.Type{T: abi.ArrayTy, Size: 2, Elem: &abi.Type{T: abi.BytesTy}}, Want: "bytes[2]"},
		{Type: abi.Type{T: abi.SliceTy, Elem: &abi.Type{T: abi.FixedBytesTy, Size: 32}}, Want: "bytes32[]"},
		{Type: abi.Type{T: abi.ArrayTy, Size: 2, Elem: &abi.Type{T: abi.FixedBytesTy, Size: 32}}, Want: "bytes32[2]"},
		{Type: abi.Type{T: abi.HashTy, Size: 20}, Want: "hash"},
		{Type: abi.Type{T: abi.StringTy}, Want: "string"},
		{Type: abi.Type{T: abi.SliceTy, Elem: &abi.Type{T: abi.StringTy}}, Want: "string[]"},
		{Type: abi.Type{T: abi.ArrayTy, Size: 2, Elem: &abi.Type{T: abi.StringTy}}, Want: "string[2]"},
		{Type: abi.Type{Size: 20, T: abi.AddressTy}, Want: "address"},
		{Type: abi.Type{T: abi.SliceTy, Elem: &abi.Type{Size: 20, T: abi.AddressTy}}, Want: "address[]"},
		{Type: abi.Type{T: abi.ArrayTy, Size: 2, Elem: &abi.Type{Size: 20, T: abi.AddressTy}}, Want: "address[2]"},
		{Type: abi.Type{T: abi.TupleTy, TupleElems: []*abi.Type{&typeUint256, &typeUint256}}, Want: "(uint256,uint256)"},
		{Type: abi.Type{T: abi.TupleTy, TupleElems: []*abi.Type{{T: abi.TupleTy, TupleElems: []*abi.Type{&typeUint256, &typeUint256}}, &typeUint256}}, Want: "((uint256,uint256),uint256)"},
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

func TestCopyValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		T       byte
		Dst     any
		Src     any
		WantErr error
	}{
		{
			T:   abi.UintTy,
			Dst: new(big.Int),
			Src: big.NewInt(42),
		},
		{
			T:   abi.UintTy,
			Dst: new(big.Int),
			Src: big.NewInt(42),
		},
		{
			T:       abi.UintTy,
			Dst:     new(big.Int),
			Src:     []byte{1, 2, 3},
			WantErr: errInvalidType,
		},
		{
			T:   abi.BytesTy,
			Dst: &[]byte{},
			Src: &[]byte{1, 2, 3},
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			err := copyVal(test.T, test.Dst, test.Src)
			if diff := cmp.Diff(test.WantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("(-want, +got)\n%s", diff)
			} else if err != nil {
				return
			}

			if diff := cmp.Diff(test.Dst, test.Src, cmp.AllowUnexported(big.Int{})); diff != "" {
				t.Fatalf("(-want, +got)\n%s", diff)
			}
		})
	}
}

func addrPtr(addr common.Address) *common.Address { return &addr }
