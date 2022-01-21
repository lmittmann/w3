package w3

import (
	"bytes"
	"math/big"
	"strconv"
	"testing"

	_abi "github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/lmittmann/w3/core"
)

func TestEncodeArgs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		Func core.Func
		Args []interface{}
		Want []byte
	}{
		{
			Func: MustNewFunc("balanceOf(address who)", "uint256 balance"),
			Args: []interface{}{A("0x000000000000000000000000000000000000dEaD")},
			Want: B("0x70a08231000000000000000000000000000000000000000000000000000000000000dEaD"),
		},
		{
			Func: MustNewFunc("transfer(address recipient, uint256 amount)", "bool success"),
			Args: []interface{}{A("0x000000000000000000000000000000000000dEaD"), big.NewInt(1)},
			Want: B("0xa9059cbb000000000000000000000000000000000000000000000000000000000000dEaD0000000000000000000000000000000000000000000000000000000000000001"),
		},
		{
			Func: MustNewFunc("name()", "string"),
			Args: []interface{}{},
			Want: B("0x06fdde03"),
		},
		{
			Func: MustNewFunc("withdraw(uint256)", ""),
			Args: []interface{}{big.NewInt(1)},
			Want: B("0x2e1a7d4d0000000000000000000000000000000000000000000000000000000000000001"),
		},
		{
			Func: MustNewFunc("getAmountsOut(uint256,address[])", "uint256[]"),
			Args: []interface{}{big.NewInt(1), []common.Address{A("0x1111111111111111111111111111111111111111"), A("0x2222222222222222222222222222222222222222")}},
			Want: B("0xd06ca61f00000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000040000000000000000000000000000000000000000000000000000000000000000200000000000000000000000011111111111111111111111111111111111111110000000000000000000000002222222222222222222222222222222222222222"),
		},
		{
			Func: MustNewFunc("test((address arg0, uint256 arg1))", ""),
			Args: []interface{}{&tuple{
				Arg0: A("0x000000000000000000000000000000000000c0Fe"),
				Arg1: big.NewInt(42),
			}},
			Want: B("0xba71720c000000000000000000000000000000000000000000000000000000000000c0fe000000000000000000000000000000000000000000000000000000000000002a"),
		},
		{
			Func: MustNewFunc("test((address arg0, uint256 arg1))", ""),
			Args: []interface{}{&tupleWithWrongOrder{
				Arg1: big.NewInt(42),
				Arg0: A("0x000000000000000000000000000000000000c0Fe"),
			}},
			Want: B("0xba71720c000000000000000000000000000000000000000000000000000000000000c0fe000000000000000000000000000000000000000000000000000000000000002a"),
		},
		{
			Func: MustNewFunc("test((address arg0, uint256 arg1))", ""),
			Args: []interface{}{&tupleWithMoreArgs{
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
				t.Fatalf("(-want +got):\n-%x\n+%x", test.Want, encodedInput)
			}
		})
	}
}

func TestDecodeArgs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		Func     core.Func
		Input    []byte
		Args     []interface{}
		WantArgs []interface{}
	}{
		{
			Func:     MustNewFunc("test(address)", ""),
			Input:    B("0xffffffff000000000000000000000000000000000000000000000000000000000000c0fe"),
			Args:     []interface{}{new(common.Address)},
			WantArgs: []interface{}{APtr("0x000000000000000000000000000000000000c0Fe")},
		},
		{
			Func:     MustNewFunc("test(uint256)", ""),
			Input:    B("0xffffffff000000000000000000000000000000000000000000000000000000000000002a"),
			Args:     []interface{}{new(big.Int)},
			WantArgs: []interface{}{big.NewInt(42)},
		},
		{
			Func:     MustNewFunc("test(bool)", ""),
			Input:    B("0xffffffff0000000000000000000000000000000000000000000000000000000000000001"),
			Args:     []interface{}{boolPtr(false)},
			WantArgs: []interface{}{boolPtr(true)},
		},
		{
			Func:     MustNewFunc("test(bytes32)", ""),
			Input:    B("0xffffffff0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20"),
			Args:     []interface{}{&[32]byte{}},
			WantArgs: []interface{}{&[32]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32}},
		},
		{
			Func:     MustNewFunc("test(bytes32)", ""),
			Input:    B("0xffffffff0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20"),
			Args:     []interface{}{new(common.Hash)},
			WantArgs: []interface{}{hashPtr(H("0x0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20"))},
		},
		{
			Func:     MustNewFunc("test(bytes)", ""),
			Input:    B("0xffffffff000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000030102030000000000000000000000000000000000000000000000000000000000"),
			Args:     []interface{}{&[]byte{}},
			WantArgs: []interface{}{&[]byte{1, 2, 3}},
		},
		{
			Func:  MustNewFunc("test((address arg0, uint256 arg1))", ""),
			Input: B("0xffffffff000000000000000000000000000000000000000000000000000000000000c0fe000000000000000000000000000000000000000000000000000000000000002a"),
			Args:  []interface{}{new(tuple)},
			WantArgs: []interface{}{&tuple{
				Arg0: A("0x000000000000000000000000000000000000c0Fe"),
				Arg1: big.NewInt(42),
			}},
		},
		{
			Func:  MustNewFunc("test((address arg0, uint256 arg1))", ""),
			Input: B("0xffffffff000000000000000000000000000000000000000000000000000000000000c0fe000000000000000000000000000000000000000000000000000000000000002a"),
			Args:  []interface{}{new(tupleWithWrongOrder)},
			WantArgs: []interface{}{&tupleWithWrongOrder{
				Arg1: big.NewInt(42),
				Arg0: A("0x000000000000000000000000000000000000c0Fe"),
			}},
		},
		{
			Func:  MustNewFunc("test((address arg0, uint256 arg1))", ""),
			Input: B("0xffffffff000000000000000000000000000000000000000000000000000000000000c0fe000000000000000000000000000000000000000000000000000000000000002a"),
			Args:  []interface{}{new(tupleWithMoreArgs)},
			WantArgs: []interface{}{&tupleWithMoreArgs{
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

func TestDecodeReturns(t *testing.T) {
	t.Parallel()

	tests := []struct {
		Func        core.Func
		Output      []byte
		Returns     []interface{}
		WantReturns []interface{}
	}{
		{
			Func:        MustNewFunc("test()", "address"),
			Output:      B("0x000000000000000000000000000000000000000000000000000000000000c0fe"),
			Returns:     []interface{}{new(common.Address)},
			WantReturns: []interface{}{APtr("0x000000000000000000000000000000000000c0Fe")},
		},
		{
			Func:        MustNewFunc("test()", "uint256"),
			Output:      B("0x000000000000000000000000000000000000000000000000000000000000002a"),
			Returns:     []interface{}{new(big.Int)},
			WantReturns: []interface{}{big.NewInt(42)},
		},
		{
			Func:        MustNewFunc("test()", "bool"),
			Output:      B("0x0000000000000000000000000000000000000000000000000000000000000001"),
			Returns:     []interface{}{boolPtr(false)},
			WantReturns: []interface{}{boolPtr(true)},
		},
		{
			Func:        MustNewFunc("test()", "bytes32"),
			Output:      B("0x0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20"),
			Returns:     []interface{}{&[32]byte{}},
			WantReturns: []interface{}{&[32]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32}},
		},
		{
			Func:        MustNewFunc("test()", "bytes32"),
			Output:      B("0x0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20"),
			Returns:     []interface{}{new(common.Hash)},
			WantReturns: []interface{}{hashPtr(H("0x0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20"))},
		},
		{
			Func:        MustNewFunc("test()", "bytes"),
			Output:      B("0x000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000030102030000000000000000000000000000000000000000000000000000000000"),
			Returns:     []interface{}{&[]byte{}},
			WantReturns: []interface{}{&[]byte{1, 2, 3}},
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

func TestCopyValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		T       byte
		Dst     interface{}
		Src     interface{}
		WantErr error
	}{
		{
			T:   _abi.UintTy,
			Dst: new(big.Int),
			Src: big.NewInt(42),
		},
		{
			T:   _abi.UintTy,
			Dst: new(big.Int),
			Src: big.NewInt(42),
		},
		{
			T:       _abi.UintTy,
			Dst:     new(big.Int),
			Src:     []byte{1, 2, 3},
			WantErr: ErrInvalidType,
		},
		{
			T:   _abi.BytesTy,
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
