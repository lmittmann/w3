package w3

import (
	"bytes"
	"math/big"
	"strconv"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/lmittmann/w3/w3types"
)

func TestNewFunc(t *testing.T) {
	t.Parallel()

	tests := []struct {
		Signature string
		Returns   string
		WantFunc  *Func
	}{
		{
			Signature: "transfer(address,uint256)",
			Returns:   "bool",
			WantFunc: &Func{
				Signature: "transfer(address,uint256)",
				Selector:  [4]byte{0xa9, 0x05, 0x9c, 0xbb},
			},
		},
		{
			Signature: "transfer(address recipient, uint256 amount)",
			Returns:   "bool success",
			WantFunc: &Func{
				Signature: "transfer(address,uint256)",
				Selector:  [4]byte{0xa9, 0x05, 0x9c, 0xbb},
			},
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			gotFunc, err := NewFunc(test.Signature, test.Returns)
			if err != nil {
				t.Fatalf("Failed to create new FUnc: %v", err)
			}

			if diff := cmp.Diff(test.WantFunc, gotFunc,
				cmpopts.IgnoreFields(Func{}, "Args", "Returns", "name"),
			); diff != "" {
				t.Fatalf("(-want, +got)\n%s", diff)
			}
		})
	}
}

func TestFuncEncodeArgs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		Func w3types.Func
		Args []any
		Want []byte
	}{
		{
			Func: MustNewFunc("balanceOf(address who)", "uint256 balance"),
			Args: []any{A("0x000000000000000000000000000000000000dEaD")},
			Want: B("0x70a08231000000000000000000000000000000000000000000000000000000000000dEaD"),
		},
		{
			Func: MustNewFunc("transfer(address recipient, uint256 amount)", "bool success"),
			Args: []any{A("0x000000000000000000000000000000000000dEaD"), big.NewInt(1)},
			Want: B("0xa9059cbb000000000000000000000000000000000000000000000000000000000000dEaD0000000000000000000000000000000000000000000000000000000000000001"),
		},
		{
			Func: MustNewFunc("name()", "string"),
			Args: []any{},
			Want: B("0x06fdde03"),
		},
		{
			Func: MustNewFunc("withdraw(uint256)", ""),
			Args: []any{big.NewInt(1)},
			Want: B("0x2e1a7d4d0000000000000000000000000000000000000000000000000000000000000001"),
		},
		{
			Func: MustNewFunc("getAmountsOut(uint256,address[])", "uint256[]"),
			Args: []any{big.NewInt(1), []common.Address{A("0x1111111111111111111111111111111111111111"), A("0x2222222222222222222222222222222222222222")}},
			Want: B("0xd06ca61f00000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000040000000000000000000000000000000000000000000000000000000000000000200000000000000000000000011111111111111111111111111111111111111110000000000000000000000002222222222222222222222222222222222222222"),
		},
		{
			Func: MustNewFunc("test((address arg0, uint256 arg1))", ""),
			Args: []any{&tuple{
				Arg0: A("0x000000000000000000000000000000000000c0Fe"),
				Arg1: big.NewInt(42),
			}},
			Want: B("0xba71720c000000000000000000000000000000000000000000000000000000000000c0fe000000000000000000000000000000000000000000000000000000000000002a"),
		},
		{
			Func: MustNewFunc("test((address arg0, uint256 arg1))", ""),
			Args: []any{&tupleWithWrongOrder{
				Arg1: big.NewInt(42),
				Arg0: A("0x000000000000000000000000000000000000c0Fe"),
			}},
			Want: B("0xba71720c000000000000000000000000000000000000000000000000000000000000c0fe000000000000000000000000000000000000000000000000000000000000002a"),
		},
		{
			Func: MustNewFunc("test((address arg0, uint256 arg1))", ""),
			Args: []any{&tupleWithMoreArgs{
				Arg0: A("0x000000000000000000000000000000000000c0Fe"),
				Arg1: big.NewInt(42),
				Arg2: big.NewInt(7),
			}},
			Want: B("0xba71720c000000000000000000000000000000000000000000000000000000000000c0fe000000000000000000000000000000000000000000000000000000000000002a"),
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			encodedInput, err := test.Func.EncodeArgs(test.Args...)
			if err != nil {
				t.Fatalf("Failed to encode args: %v", err)
			}

			if !bytes.Equal(test.Want, encodedInput) {
				t.Fatalf("(-want, +got)\n-%x\n+%x", test.Want, encodedInput)
			}
		})
	}
}

func TestFuncDecodeArgs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		Func     w3types.Func
		Input    []byte
		Args     []any
		WantArgs []any
	}{
		{
			Func:     MustNewFunc("test(address)", ""),
			Input:    B("0xffffffff000000000000000000000000000000000000000000000000000000000000c0fe"),
			Args:     []any{new(common.Address)},
			WantArgs: []any{APtr("0x000000000000000000000000000000000000c0Fe")},
		},
		{
			Func:     MustNewFunc("test(uint256)", ""),
			Input:    B("0xffffffff000000000000000000000000000000000000000000000000000000000000002a"),
			Args:     []any{new(big.Int)},
			WantArgs: []any{big.NewInt(42)},
		},
		{
			Func:     MustNewFunc("test(bool)", ""),
			Input:    B("0xffffffff0000000000000000000000000000000000000000000000000000000000000001"),
			Args:     []any{boolPtr(false)},
			WantArgs: []any{boolPtr(true)},
		},
		{
			Func:     MustNewFunc("test(bytes32)", ""),
			Input:    B("0xffffffff0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20"),
			Args:     []any{&[32]byte{}},
			WantArgs: []any{&[32]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32}},
		},
		{
			Func:     MustNewFunc("test(bytes32)", ""),
			Input:    B("0xffffffff0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20"),
			Args:     []any{new(common.Hash)},
			WantArgs: []any{hashPtr(H("0x0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20"))},
		},
		{
			Func:     MustNewFunc("test(bytes)", ""),
			Input:    B("0xffffffff000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000030102030000000000000000000000000000000000000000000000000000000000"),
			Args:     []any{&[]byte{}},
			WantArgs: []any{&[]byte{1, 2, 3}},
		},
		{
			Func:  MustNewFunc("test((address arg0, uint256 arg1))", ""),
			Input: B("0xffffffff000000000000000000000000000000000000000000000000000000000000c0fe000000000000000000000000000000000000000000000000000000000000002a"),
			Args:  []any{new(tuple)},
			WantArgs: []any{&tuple{
				Arg0: A("0x000000000000000000000000000000000000c0Fe"),
				Arg1: big.NewInt(42),
			}},
		},
		{
			Func:  MustNewFunc("test((address arg0, uint256 arg1))", ""),
			Input: B("0xffffffff000000000000000000000000000000000000000000000000000000000000c0fe000000000000000000000000000000000000000000000000000000000000002a"),
			Args:  []any{new(tupleWithWrongOrder)},
			WantArgs: []any{&tupleWithWrongOrder{
				Arg1: big.NewInt(42),
				Arg0: A("0x000000000000000000000000000000000000c0Fe"),
			}},
		},
		{
			Func:  MustNewFunc("test((address arg0, uint256 arg1))", ""),
			Input: B("0xffffffff000000000000000000000000000000000000000000000000000000000000c0fe000000000000000000000000000000000000000000000000000000000000002a"),
			Args:  []any{new(tupleWithMoreArgs)},
			WantArgs: []any{&tupleWithMoreArgs{
				Arg0: A("0x000000000000000000000000000000000000c0Fe"),
				Arg1: big.NewInt(42),
			}},
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			if err := test.Func.DecodeArgs(test.Input, test.Args...); err != nil {
				t.Fatalf("Failed to decode args: %v", err)
			}
			if diff := cmp.Diff(test.WantArgs, test.Args, cmp.AllowUnexported(big.Int{})); diff != "" {
				t.Fatalf("(-want, +got)\n%s", diff)
			}
		})
	}
}

func TestFuncDecodeReturns(t *testing.T) {
	t.Parallel()

	tests := []struct {
		Func        w3types.Func
		Output      []byte
		Returns     []any
		WantReturns []any
	}{
		{
			Func:        MustNewFunc("test()", "address"),
			Output:      B("0x000000000000000000000000000000000000000000000000000000000000c0fe"),
			Returns:     []any{new(common.Address)},
			WantReturns: []any{APtr("0x000000000000000000000000000000000000c0Fe")},
		},
		{
			Func:        MustNewFunc("test()", "uint256"),
			Output:      B("0x000000000000000000000000000000000000000000000000000000000000002a"),
			Returns:     []any{new(big.Int)},
			WantReturns: []any{big.NewInt(42)},
		},
		{
			Func:        MustNewFunc("test()", "bool"),
			Output:      B("0x0000000000000000000000000000000000000000000000000000000000000001"),
			Returns:     []any{boolPtr(false)},
			WantReturns: []any{boolPtr(true)},
		},
		{
			Func:        MustNewFunc("test()", "bytes32"),
			Output:      B("0x0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20"),
			Returns:     []any{&[32]byte{}},
			WantReturns: []any{&[32]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32}},
		},
		{
			Func:        MustNewFunc("test()", "bytes32"),
			Output:      B("0x0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20"),
			Returns:     []any{new(common.Hash)},
			WantReturns: []any{hashPtr(H("0x0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20"))},
		},
		{
			Func:        MustNewFunc("test()", "bytes"),
			Output:      B("0x000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000030102030000000000000000000000000000000000000000000000000000000000"),
			Returns:     []any{&[]byte{}},
			WantReturns: []any{&[]byte{1, 2, 3}},
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			if err := test.Func.DecodeReturns(test.Output, test.Returns...); err != nil {
				t.Fatalf("Failed to decode returns: %v", err)
			}
			if diff := cmp.Diff(test.WantReturns, test.Returns, cmp.AllowUnexported(big.Int{})); diff != "" {
				t.Fatalf("(-want, +got)\n%s", diff)
			}
		})
	}
}

func boolPtr(b bool) *bool               { return &b }
func hashPtr(h common.Hash) *common.Hash { return &h }

type tuple struct {
	Arg0 common.Address
	Arg1 *big.Int
}

type tupleWithWrongOrder struct {
	Arg1 *big.Int
	Arg0 common.Address
}

type tupleWithMoreArgs struct {
	Arg0 common.Address
	Arg1 *big.Int
	Arg2 *big.Int // Arg that is missing in func signature
}
