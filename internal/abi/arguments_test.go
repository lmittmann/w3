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
	"github.com/lmittmann/w3/internal"
)

func TestParse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		Signature     string
		WantName      string
		WantArguments Arguments
		WantErr       error
	}{
		{
			Signature:     "",
			WantArguments: Arguments{},
		},
		{
			Signature:     "uint256",
			WantArguments: Arguments{{Type: typeUint256}},
		},
		{
			Signature:     "uint256 balance",
			WantArguments: Arguments{{Type: typeUint256, Name: "balance"}},
		},
		{
			Signature:     "uint256,address",
			WantArguments: Arguments{{Type: typeUint256}, {Type: typeAddress}},
		},
		{
			Signature: "uint256[],uint256[3],uint256[][3],uint256[3][]",
			WantArguments: Arguments{
				{Type: abi.Type{T: abi.SliceTy, Elem: &typeUint256}},
				{Type: abi.Type{T: abi.ArrayTy, Size: 3, Elem: &typeUint256}},
				{Type: abi.Type{T: abi.ArrayTy, Size: 3, Elem: &abi.Type{T: abi.SliceTy, Elem: &typeUint256}}},
				{Type: abi.Type{T: abi.SliceTy, Elem: &abi.Type{T: abi.ArrayTy, Size: 3, Elem: &typeUint256}}},
			},
		},
		{
			Signature: "uint256,(uint256 v0,uint256 v1),((uint256 v00,uint256 v01) v0,uint256 v1)",
			WantArguments: Arguments{
				{Type: typeUint256},
				{Type: abi.Type{T: abi.TupleTy, TupleElems: []*abi.Type{&typeUint256, &typeUint256}, TupleRawNames: []string{"v0", "v1"}}},
				{Type: abi.Type{T: abi.TupleTy, TupleElems: []*abi.Type{
					{T: abi.TupleTy, TupleElems: []*abi.Type{&typeUint256, &typeUint256}, TupleRawNames: []string{"v00", "v01"}},
					&typeUint256,
				}, TupleRawNames: []string{"v0", "v1"}}},
			},
		},
		{
			Signature:     "func()",
			WantName:      "func",
			WantArguments: Arguments{},
		},
		{
			Signature:     "transfer(address,uint256)",
			WantName:      "transfer",
			WantArguments: Arguments{{Type: typeAddress}, {Type: typeUint256}},
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			gotName, gotArguments, gotErr := Parse(test.Signature)
			if diff := cmp.Diff(test.WantErr, gotErr,
				internal.EquateErrors(),
			); diff != "" {
				t.Fatalf("Err: (-want, +got)\n%s", diff)
			} else if test.WantName != gotName {
				t.Fatalf("Name: want %q, got%q", test.WantErr, gotName)
			} else if diff := cmp.Diff(test.WantArguments, gotArguments,
				cmpopts.EquateEmpty(),
				cmpopts.IgnoreUnexported(abi.Type{}),
				cmpopts.IgnoreFields(abi.Type{}, "TupleType"),
			); diff != "" {
				t.Fatalf("Arguments: (-want, +got)\n%s", diff)
			}
		})
	}
}

func TestSignature(t *testing.T) {
	t.Parallel()

	tests := []struct {
		Arguments     Arguments
		WantSignature string
	}{
		{
			Arguments:     Arguments{},
			WantSignature: "",
		},
		{
			Arguments:     Arguments{{Type: typeUint256}},
			WantSignature: "uint256",
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			gotSignature := test.Arguments.Signature()
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
