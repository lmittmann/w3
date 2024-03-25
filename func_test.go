package w3_test

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/lmittmann/w3"
	"github.com/lmittmann/w3/internal"
	"github.com/lmittmann/w3/w3types"
)

func ExampleNewFunc_balanceOf() {
	// ABI binding to the balanceOf function of an ERC20 Token.
	funcBalanceOf, _ := w3.NewFunc("balanceOf(address)", "uint256")

	// Optionally names can be specified for function arguments. This is
	// especially useful for more complex functions with many arguments.
	funcBalanceOf, _ = w3.NewFunc("balanceOf(address who)", "uint256 amount")

	// ABI-encode the functions args.
	input, _ := funcBalanceOf.EncodeArgs(w3.A("0xAb5801a7D398351b8bE11C439e05C5B3259aeC9B"))
	fmt.Printf("balanceOf input: 0x%x\n", input)

	// ABI-decode the functions args from a given input.
	var (
		who common.Address
	)
	funcBalanceOf.DecodeArgs(input, &who)
	fmt.Printf("balanceOf args: %v\n", who)

	// ABI-decode the functions output.
	var (
		output = w3.B("0x000000000000000000000000000000000000000000000000000000000000c0fe")
		amount = new(big.Int)
	)
	funcBalanceOf.DecodeReturns(output, amount)
	fmt.Printf("balanceOf returns: %v\n", amount)
	// Output:
	// balanceOf input: 0x70a08231000000000000000000000000ab5801a7d398351b8be11c439e05c5b3259aec9b
	// balanceOf args: 0xAb5801a7D398351b8bE11C439e05C5B3259aeC9B
	// balanceOf returns: 49406
}

func ExampleNewFunc_uniswapV4Swap() {
	// ABI binding for the Uniswap v4 swap function.
	funcSwap, _ := w3.NewFunc(`swap(
		(address currency0, address currency1, uint24 fee, int24 tickSpacing, address hooks) key,
		(bool zeroForOne, int256 amountSpecified, uint160 sqrtPriceLimitX96) params,
		bytes hookData
	)`, "int256 delta")

	// ABI binding for the PoolKey struct.
	type PoolKey struct {
		Currency0   common.Address
		Currency1   common.Address
		Fee         *big.Int
		TickSpacing *big.Int
		Hooks       common.Address
	}

	// ABI binding for the SwapParams struct.
	type SwapParams struct {
		ZeroForOne        bool
		AmountSpecified   *big.Int
		SqrtPriceLimitX96 *big.Int
	}

	// ABI-encode the functions args.
	input, _ := funcSwap.EncodeArgs(
		&PoolKey{
			Currency0:   w3.A("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2"),
			Currency1:   w3.A("0x6B175474E89094C44Da98b954EedeAC495271d0F"),
			Fee:         big.NewInt(0),
			TickSpacing: big.NewInt(0),
		},
		&SwapParams{
			ZeroForOne:        false,
			AmountSpecified:   big.NewInt(0),
			SqrtPriceLimitX96: big.NewInt(0),
		},
		[]byte{},
	)
	fmt.Printf("swap input: 0x%x\n", input)
	// Output:
	// swap input: 0xf3cd914c000000000000000000000000c02aaa39b223fe8d0a0e5c4f27ead9083c756cc20000000000000000000000006b175474e89094c44da98b954eedeac495271d0f00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001200000000000000000000000000000000000000000000000000000000000000000
}

func TestNewFunc(t *testing.T) {
	tests := []struct {
		Signature string
		Returns   string
		WantFunc  *w3.Func
	}{
		{
			Signature: "transfer(address,uint256)",
			Returns:   "bool",
			WantFunc: &w3.Func{
				Signature: "transfer(address,uint256)",
				Selector:  [4]byte{0xa9, 0x05, 0x9c, 0xbb},
			},
		},
		{
			Signature: "transfer(address recipient, uint256 amount)",
			Returns:   "bool success",
			WantFunc: &w3.Func{
				Signature: "transfer(address,uint256)",
				Selector:  [4]byte{0xa9, 0x05, 0x9c, 0xbb},
			},
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			gotFunc, err := w3.NewFunc(test.Signature, test.Returns)
			if err != nil {
				t.Fatalf("Failed to create new FUnc: %v", err)
			}

			if diff := cmp.Diff(test.WantFunc, gotFunc,
				cmpopts.IgnoreFields(w3.Func{}, "Args", "Returns", "name"),
			); diff != "" {
				t.Fatalf("(-want, +got)\n%s", diff)
			}
		})
	}
}

func TestFuncEncodeArgs(t *testing.T) {
	tests := []struct {
		Func w3types.Func
		Args []any
		Want []byte
	}{
		{
			Func: w3.MustNewFunc("balanceOf(address who)", "uint256 balance"),
			Args: []any{w3.A("0x000000000000000000000000000000000000dEaD")},
			Want: w3.B("0x70a08231000000000000000000000000000000000000000000000000000000000000dEaD"),
		},
		{
			Func: w3.MustNewFunc("transfer(address recipient, uint256 amount)", "bool success"),
			Args: []any{w3.A("0x000000000000000000000000000000000000dEaD"), big.NewInt(1)},
			Want: w3.B("0xa9059cbb000000000000000000000000000000000000000000000000000000000000dEaD0000000000000000000000000000000000000000000000000000000000000001"),
		},
		{
			Func: w3.MustNewFunc("name()", "string"),
			Args: []any{},
			Want: w3.B("0x06fdde03"),
		},
		{
			Func: w3.MustNewFunc("withdraw(uint256)", ""),
			Args: []any{big.NewInt(1)},
			Want: w3.B("0x2e1a7d4d0000000000000000000000000000000000000000000000000000000000000001"),
		},
		{
			Func: w3.MustNewFunc("getAmountsOut(uint256,address[])", "uint256[]"),
			Args: []any{big.NewInt(1), []common.Address{w3.A("0x1111111111111111111111111111111111111111"), w3.A("0x2222222222222222222222222222222222222222")}},
			Want: w3.B("0xd06ca61f00000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000040000000000000000000000000000000000000000000000000000000000000000200000000000000000000000011111111111111111111111111111111111111110000000000000000000000002222222222222222222222222222222222222222"),
		},
		{
			Func: w3.MustNewFunc("test((address arg0, uint256 arg1))", ""),
			Args: []any{&tuple{
				Arg0: w3.A("0x000000000000000000000000000000000000c0Fe"),
				Arg1: big.NewInt(42),
			}},
			Want: w3.B("0xba71720c000000000000000000000000000000000000000000000000000000000000c0fe000000000000000000000000000000000000000000000000000000000000002a"),
		},
		{
			Func: w3.MustNewFunc("test((address arg0, uint256 arg1))", ""),
			Args: []any{&tupleWithWrongOrder{
				Arg1: big.NewInt(42),
				Arg0: w3.A("0x000000000000000000000000000000000000c0Fe"),
			}},
			Want: w3.B("0xba71720c000000000000000000000000000000000000000000000000000000000000c0fe000000000000000000000000000000000000000000000000000000000000002a"),
		},
		{
			Func: w3.MustNewFunc("test((address arg0, uint256 arg1))", ""),
			Args: []any{&tupleWithMoreArgs{
				Arg0: w3.A("0x000000000000000000000000000000000000c0Fe"),
				Arg1: big.NewInt(42),
				Arg2: big.NewInt(7),
			}},
			Want: w3.B("0xba71720c000000000000000000000000000000000000000000000000000000000000c0fe000000000000000000000000000000000000000000000000000000000000002a"),
		},
		{
			Func: w3.MustNewFunc("test((address arg0, uint256 arg1)[] args)", ""),
			Args: []any{
				[]tuple{
					{Arg0: w3.A("0x1111111111111111111111111111111111111111"), Arg1: big.NewInt(7)},
					{Arg0: w3.A("0x2222222222222222222222222222222222222222"), Arg1: big.NewInt(42)},
				},
			},
			Want: w3.B("0xae4f5efa00000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000002000000000000000000000000111111111111111111111111111111111111111100000000000000000000000000000000000000000000000000000000000000070000000000000000000000002222222222222222222222222222222222222222000000000000000000000000000000000000000000000000000000000000002a"),
		},
		{
			Func: w3.MustNewFunc("test((address arg0, bytes arg1)[] calls)", ""),
			Args: []any{
				[]tupleWithBytes{
					{Arg0: w3.A("0x1111111111111111111111111111111111111111"), Arg1: w3.B("0xc0fe")},
					{Arg0: w3.A("0x2222222222222222222222222222222222222222"), Arg1: w3.B("0xdeadbeef")},
				},
			},
			Want: w3.B("0x3a91207700000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000000c0000000000000000000000000111111111111111111111111111111111111111100000000000000000000000000000000000000000000000000000000000000400000000000000000000000000000000000000000000000000000000000000002c0fe000000000000000000000000000000000000000000000000000000000000000000000000000000000000222222222222222222222222222222222222222200000000000000000000000000000000000000000000000000000000000000400000000000000000000000000000000000000000000000000000000000000004deadbeef00000000000000000000000000000000000000000000000000000000"),
		},
		{
			Func: w3.MustNewFunc("test(uint[])", ""),
			Args: []any{
				[]*big.Int{big.NewInt(0xdead), big.NewInt(0xbeef)},
			},
			Want: w3.B("0xca16068400000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000dead000000000000000000000000000000000000000000000000000000000000beef"),
		},
		{
			Func: w3.MustNewFunc("test(uint[2])", ""),
			Args: []any{
				[2]*big.Int{big.NewInt(0xdead), big.NewInt(0xbeef)},
			},
			Want: w3.B("0xf1635056000000000000000000000000000000000000000000000000000000000000dead000000000000000000000000000000000000000000000000000000000000beef"),
		},
		{
			Func: w3.MustNewFunc("test(uint64[])", ""),
			Args: []any{
				[]uint64{0xdead, 0xbeef},
			},
			Want: w3.B("0xd3469fbd00000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000dead000000000000000000000000000000000000000000000000000000000000beef"),
		},
		{
			Func: w3.MustNewFunc("test(uint64[2])", ""),
			Args: []any{
				[2]uint64{0xdead, 0xbeef},
			},
			Want: w3.B("0x533d6285000000000000000000000000000000000000000000000000000000000000dead000000000000000000000000000000000000000000000000000000000000beef"),
		},
		{ // https://github.com/lmittmann/w3/issues/35
			Func: w3.MustNewFunc("test(((address to)[] recipients) param)", ""),
			Args: []any{
				&tupleIssue35{Recipients: []struct {
					To common.Address
				}{
					{To: w3.A("0x1111111111111111111111111111111111111111")},
					{To: w3.A("0x2222222222222222222222222222222222222222")},
				}},
			},
			Want: w3.B("0xf61d1a2a00000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000000200000000000000000000000011111111111111111111111111111111111111110000000000000000000000002222222222222222222222222222222222222222"),
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			encodedInput, err := test.Func.EncodeArgs(test.Args...)
			if err != nil {
				t.Fatalf("Failed to encode args: %v", err)
			}

			if !bytes.Equal(test.Want, encodedInput) {
				t.Fatalf("(-want, +got)\n- 0x%x\n+ 0x%x", test.Want, encodedInput)
			}
		})
	}
}

func TestFuncDecodeArgs(t *testing.T) {
	tests := []struct {
		Func     w3types.Func
		Input    []byte
		Args     []any
		WantArgs []any
		WantErr  error
	}{
		{
			Func:     w3.MustNewFunc("test(address)", ""),
			Input:    w3.B("0xbb29998e000000000000000000000000000000000000000000000000000000000000c0fe"),
			Args:     []any{new(common.Address)},
			WantArgs: []any{w3.APtr("0x000000000000000000000000000000000000c0Fe")},
		},
		{
			Func:     w3.MustNewFunc("test(uint256)", ""),
			Input:    w3.B("0x29e99f07000000000000000000000000000000000000000000000000000000000000002a"),
			Args:     []any{new(big.Int)},
			WantArgs: []any{big.NewInt(42)},
		},
		{
			Func:     w3.MustNewFunc("test(bool)", ""),
			Input:    w3.B("0x36091dff0000000000000000000000000000000000000000000000000000000000000001"),
			Args:     []any{ptr(false)},
			WantArgs: []any{ptr(true)},
		},
		{
			Func:     w3.MustNewFunc("test(bytes32)", ""),
			Input:    w3.B("0x993723210102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20"),
			Args:     []any{&[32]byte{}},
			WantArgs: []any{&[32]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32}},
		},
		{
			Func:     w3.MustNewFunc("test(bytes32)", ""),
			Input:    w3.B("0x993723210102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20"),
			Args:     []any{new(common.Hash)},
			WantArgs: []any{ptr(w3.H("0x0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20"))},
		},
		{
			Func:     w3.MustNewFunc("test(bytes)", ""),
			Input:    w3.B("0x2f570a23000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000030102030000000000000000000000000000000000000000000000000000000000"),
			Args:     []any{&[]byte{}},
			WantArgs: []any{&[]byte{1, 2, 3}},
		},
		{
			Func:  w3.MustNewFunc("test((address arg0, uint256 arg1))", ""),
			Input: w3.B("0xba71720c000000000000000000000000000000000000000000000000000000000000c0fe000000000000000000000000000000000000000000000000000000000000002a"),
			Args:  []any{new(tuple)},
			WantArgs: []any{&tuple{
				Arg0: w3.A("0x000000000000000000000000000000000000c0Fe"),
				Arg1: big.NewInt(42),
			}},
		},
		{
			// https://github.com/lmittmann/w3/issues/67
			Func:     w3.MustNewFunc("test((address, uint256))", ""),
			Input:    w3.B("0xba71720c000000000000000000000000000000000000000000000000000000000000c0fe000000000000000000000000000000000000000000000000000000000000002a"),
			Args:     []any{nil},
			WantArgs: []any{nil},
		},
		{
			// https://github.com/lmittmann/w3/issues/67
			Func:  w3.MustNewFunc("test((address, uint256))", ""),
			Input: w3.B("0xba71720c000000000000000000000000000000000000000000000000000000000000c0fe000000000000000000000000000000000000000000000000000000000000002a"),
			Args:  []any{new(tuple)},
			WantArgs: []any{&tuple{
				Arg0: w3.A("0x000000000000000000000000000000000000c0Fe"),
				Arg1: big.NewInt(42),
			}},
		},
		{
			// https://github.com/lmittmann/w3/issues/67
			Func:  w3.MustNewFunc("test((address, (address, uint256)))", ""),
			Input: w3.B("0x1a68b84c000000000000000000000000000000000000000000000000000000000000c0fe000000000000000000000000000000000000000000000000000000000000dead000000000000000000000000000000000000000000000000000000000000002a"),
			Args:  []any{new(tupleNested)},
			WantArgs: []any{&tupleNested{
				Arg0: w3.A("0x000000000000000000000000000000000000c0Fe"),
				Arg1: tuple{
					Arg0: w3.A("0x000000000000000000000000000000000000dEaD"),
					Arg1: big.NewInt(42),
				},
			}},
		},
		{
			Func:  w3.MustNewFunc("test((address arg0, uint256 arg1))", ""),
			Input: w3.B("0xba71720c000000000000000000000000000000000000000000000000000000000000c0fe000000000000000000000000000000000000000000000000000000000000002a"),
			Args:  []any{new(tupleWithWrongOrder)},
			WantArgs: []any{&tupleWithWrongOrder{
				Arg1: big.NewInt(42),
				Arg0: w3.A("0x000000000000000000000000000000000000c0Fe"),
			}},
		},
		{
			Func:  w3.MustNewFunc("test((address arg0, uint256 arg1))", ""),
			Input: w3.B("0xba71720c000000000000000000000000000000000000000000000000000000000000c0fe000000000000000000000000000000000000000000000000000000000000002a"),
			Args:  []any{new(tupleWithMoreArgs)},
			WantArgs: []any{&tupleWithMoreArgs{
				Arg0: w3.A("0x000000000000000000000000000000000000c0Fe"),
				Arg1: big.NewInt(42),
			}},
		},
		{ // https://github.com/lmittmann/w3/issues/22
			Func:    w3.MustNewFunc("transfer(address recipient, uint256 amount)", "bool success"),
			Input:   w3.B("0x"),
			Args:    []any{new(common.Address), new(big.Int)},
			WantErr: errors.New("w3: insufficient input length"),
		},
		{
			Func:  w3.MustNewFunc("test((address arg0, uint256 arg1))", ""),
			Input: w3.B("0xba71720c000000000000000000000000000000000000000000000000000000000000c0fe000000000000000000000000000000000000000000000000000000000000002a"),
			Args:  []any{new(tupleWithUnexportedProperty)},
			WantArgs: []any{&tupleWithUnexportedProperty{
				Arg1: big.NewInt(42),
			}},
		},
		{
			Func:  w3.MustNewFunc("test((address arg0, bytes arg1)[] calls)", ""),
			Input: w3.B("0x3a91207700000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000000c0000000000000000000000000111111111111111111111111111111111111111100000000000000000000000000000000000000000000000000000000000000400000000000000000000000000000000000000000000000000000000000000002c0fe000000000000000000000000000000000000000000000000000000000000000000000000000000000000222222222222222222222222222222222222222200000000000000000000000000000000000000000000000000000000000000400000000000000000000000000000000000000000000000000000000000000004deadbeef00000000000000000000000000000000000000000000000000000000"),
			Args:  []any{&[]*tupleWithBytes{}},
			WantArgs: []any{
				&[]*tupleWithBytes{
					{Arg0: w3.A("0x1111111111111111111111111111111111111111"), Arg1: w3.B("0xc0fe")},
					{Arg0: w3.A("0x2222222222222222222222222222222222222222"), Arg1: w3.B("0xdeadbeef")},
				},
			},
		},
		{
			Func:  w3.MustNewFunc("test(uint[])", ""),
			Input: w3.B("0xca16068400000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000dead000000000000000000000000000000000000000000000000000000000000beef"),
			Args:  []any{new([]*big.Int)},
			WantArgs: []any{
				&[]*big.Int{big.NewInt(0xdead), big.NewInt(0xbeef)},
			},
		},
		{
			Func:  w3.MustNewFunc("test(uint[2])", ""),
			Input: w3.B("0xf1635056000000000000000000000000000000000000000000000000000000000000dead000000000000000000000000000000000000000000000000000000000000beef"),
			Args:  []any{new([2]*big.Int)},
			WantArgs: []any{
				&[2]*big.Int{big.NewInt(0xdead), big.NewInt(0xbeef)},
			},
		},
		{
			Func:  w3.MustNewFunc("test(uint64[])", ""),
			Input: w3.B("0xd3469fbd00000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000dead000000000000000000000000000000000000000000000000000000000000beef"),
			Args:  []any{new([]uint64)},
			WantArgs: []any{
				&[]uint64{0xdead, 0xbeef},
			},
		},
		{
			Func:  w3.MustNewFunc("test(uint64[2])", ""),
			Input: w3.B("0x533d6285000000000000000000000000000000000000000000000000000000000000dead000000000000000000000000000000000000000000000000000000000000beef"),
			Args:  []any{new([2]uint64)},
			WantArgs: []any{
				&[2]uint64{0xdead, 0xbeef},
			},
		},
		{ // https://github.com/lmittmann/w3/issues/35
			Func:  w3.MustNewFunc("test(((address to)[] recipients) param)", ""),
			Input: w3.B("0xf61d1a2a00000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000000200000000000000000000000011111111111111111111111111111111111111110000000000000000000000002222222222222222222222222222222222222222"),
			Args:  []any{new(tupleIssue35)},
			WantArgs: []any{
				&tupleIssue35{Recipients: []struct {
					To common.Address
				}{
					{To: w3.A("0x1111111111111111111111111111111111111111")},
					{To: w3.A("0x2222222222222222222222222222222222222222")},
				}},
			},
		},
		{
			Func:    w3.MustNewFunc("test(address)", ""),
			Input:   w3.B("0xffffffff000000000000000000000000000000000000000000000000000000000000c0fe"),
			Args:    []any{new(common.Address)},
			WantErr: errors.New("w3: input does not match selector"),
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			err := test.Func.DecodeArgs(test.Input, test.Args...)
			if diff := cmp.Diff(test.WantErr, err, internal.EquateErrors()); diff != "" {
				t.Fatalf("Err: (-want, +got)\n%s", diff)
			}
			if err != nil {
				return
			}

			if diff := cmp.Diff(test.WantArgs, test.Args,
				cmp.AllowUnexported(big.Int{}, tupleWithUnexportedProperty{}),
			); diff != "" {
				t.Fatalf("Args: (-want, +got)\n%s", diff)
			}
		})
	}
}

func TestFuncDecodeReturns(t *testing.T) {
	tests := []struct {
		Func        w3types.Func
		Output      []byte
		Returns     []any
		WantReturns []any
	}{
		{
			Func:        w3.MustNewFunc("test()", "address"),
			Output:      w3.B("0x000000000000000000000000000000000000000000000000000000000000c0fe"),
			Returns:     []any{new(common.Address)},
			WantReturns: []any{w3.APtr("0x000000000000000000000000000000000000c0Fe")},
		},
		{
			Func:        w3.MustNewFunc("test()", "uint256"),
			Output:      w3.B("0x000000000000000000000000000000000000000000000000000000000000002a"),
			Returns:     []any{new(big.Int)},
			WantReturns: []any{big.NewInt(42)},
		},
		{
			Func:        w3.MustNewFunc("test()", "bool"),
			Output:      w3.B("0x0000000000000000000000000000000000000000000000000000000000000001"),
			Returns:     []any{ptr(false)},
			WantReturns: []any{ptr(true)},
		},
		{
			Func:        w3.MustNewFunc("test()", "bytes32"),
			Output:      w3.B("0x0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20"),
			Returns:     []any{&[32]byte{}},
			WantReturns: []any{&[32]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32}},
		},
		{
			Func:        w3.MustNewFunc("test()", "bytes32"),
			Output:      w3.B("0x0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20"),
			Returns:     []any{new(common.Hash)},
			WantReturns: []any{ptr(w3.H("0x0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20"))},
		},
		{
			Func:        w3.MustNewFunc("test()", "bytes"),
			Output:      w3.B("0x000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000030102030000000000000000000000000000000000000000000000000000000000"),
			Returns:     []any{&[]byte{}},
			WantReturns: []any{&[]byte{1, 2, 3}},
		},
		{ // https://github.com/lmittmann/w3/issues/25
			Func:    w3.MustNewFunc("test()", "(address arg0, uint256 arg1)"),
			Output:  w3.B("0x000000000000000000000000000000000000000000000000000000000000c0fe000000000000000000000000000000000000000000000000000000000000002a"),
			Returns: []any{new(tuple)},
			WantReturns: []any{&tuple{
				Arg0: w3.A("0x000000000000000000000000000000000000c0Fe"),
				Arg1: big.NewInt(42),
			}},
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

func ptr[T any](v T) *T { return &v }

type tuple struct {
	Arg0 common.Address
	Arg1 *big.Int
}

type tupleWithBytes struct {
	Arg0 common.Address
	Arg1 []byte
}

type tupleWithUnexportedProperty struct {
	arg0 common.Address
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

type tupleIssue35 struct {
	Recipients []struct {
		To common.Address
	}
}

type tupleNested struct {
	Arg0 common.Address
	Arg1 tuple
}
