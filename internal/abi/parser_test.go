package abi

import (
	"errors"
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/lmittmann/w3/internal"
)

var (
	typeAddress = abi.Type{T: abi.AddressTy, Size: 20}
	typeBool    = abi.Type{T: abi.BoolTy}
	typeUint24  = abi.Type{T: abi.UintTy, Size: 24}
	typeUint160 = abi.Type{T: abi.UintTy, Size: 160}
	typeUint256 = abi.Type{T: abi.UintTy, Size: 256}
)

func TestParseArgs(t *testing.T) {
	tests := []struct {
		Input    string
		Tuples   []any
		WantArgs Arguments
		WantErr  error
	}{
		{
			Input:    "",
			WantArgs: Arguments{},
		},
		{
			Input:   "xxx",
			WantErr: errors.New(`syntax error: unexpected "xxx", expecting type`),
		},
		{
			Input:    "uint256",
			WantArgs: Arguments{{Type: typeUint256}},
		},
		{
			Input:    "uint",
			WantArgs: Arguments{{Type: typeUint256}},
		},
		{
			Input:    "uint256 balance",
			WantArgs: Arguments{{Type: typeUint256, Name: "balance"}},
		},
		{
			Input:    "uint256 indexed balance",
			WantArgs: Arguments{{Type: typeUint256, Indexed: true, Name: "balance"}},
		},
		{
			Input:    "uint256 indexed",
			WantArgs: Arguments{{Type: typeUint256, Indexed: true}},
		},
		{
			Input:    "uint256[]",
			WantArgs: Arguments{{Type: abi.Type{Elem: &typeUint256, T: abi.SliceTy}}},
		},
		{
			Input:    "uint256[3]",
			WantArgs: Arguments{{Type: abi.Type{Elem: &typeUint256, T: abi.ArrayTy, Size: 3}}},
		},
		{
			Input: "uint256[][]",
			WantArgs: Arguments{{
				Type: abi.Type{
					Elem: &abi.Type{Elem: &typeUint256, T: abi.SliceTy},
					T:    abi.SliceTy,
				},
			}},
		},
		{
			Input: "uint256[][3]",
			WantArgs: Arguments{{
				Type: abi.Type{
					Elem: &abi.Type{Elem: &typeUint256, T: abi.SliceTy},
					T:    abi.ArrayTy,
					Size: 3,
				},
			}},
		},
		{
			Input: "uint256[3][]",
			WantArgs: Arguments{{
				Type: abi.Type{
					Elem: &abi.Type{Elem: &typeUint256, T: abi.ArrayTy, Size: 3},
					T:    abi.SliceTy,
				},
			}},
		},
		{
			Input: "uint256[],uint256[3],uint256[][3],uint256[3][]",
			WantArgs: Arguments{
				{Type: abi.Type{T: abi.SliceTy, Elem: &typeUint256}},
				{Type: abi.Type{T: abi.ArrayTy, Size: 3, Elem: &typeUint256}},
				{Type: abi.Type{T: abi.ArrayTy, Size: 3, Elem: &abi.Type{T: abi.SliceTy, Elem: &typeUint256}}},
				{Type: abi.Type{T: abi.SliceTy, Elem: &abi.Type{T: abi.ArrayTy, Size: 3, Elem: &typeUint256}}},
			},
		},
		{
			Input:   "uint256[",
			WantErr: errors.New(`syntax error: unexpected EOF, expecting "]"`),
		},
		{
			Input:   "uint256[3",
			WantErr: errors.New(`syntax error: unexpected EOF, expecting "]"`),
		},
		{
			Input: "(uint256 arg0)",
			WantArgs: Arguments{{
				Type: abi.Type{
					T:             abi.TupleTy,
					TupleElems:    []*abi.Type{&typeUint256},
					TupleRawNames: []string{"arg0"},
				},
			}},
		},
		{
			Input: "(uint256 arg0)[]",
			WantArgs: Arguments{{
				Type: abi.Type{
					Elem: &abi.Type{
						T:             abi.TupleTy,
						TupleElems:    []*abi.Type{&typeUint256},
						TupleRawNames: []string{"arg0"},
					},
					T: abi.SliceTy,
				},
			}},
		},
		{
			Input: "(uint256 arg0)[3]",
			WantArgs: Arguments{{
				Type: abi.Type{
					Elem: &abi.Type{
						T:             abi.TupleTy,
						TupleElems:    []*abi.Type{&typeUint256},
						TupleRawNames: []string{"arg0"},
					},
					T:    abi.ArrayTy,
					Size: 3,
				},
			}},
		},
		{
			Input: "uint256,(uint256 v0,uint256 v1),((uint256 v00,uint256 v01) v0,uint256 v1)",
			WantArgs: Arguments{
				{Type: typeUint256},
				{Type: abi.Type{T: abi.TupleTy, TupleElems: []*abi.Type{&typeUint256, &typeUint256}, TupleRawNames: []string{"v0", "v1"}}},
				{Type: abi.Type{T: abi.TupleTy, TupleElems: []*abi.Type{
					{T: abi.TupleTy, TupleElems: []*abi.Type{&typeUint256, &typeUint256}, TupleRawNames: []string{"v00", "v01"}},
					&typeUint256,
				}, TupleRawNames: []string{"v0", "v1"}}},
			},
		},
		{
			Input:  "simpleStruct",
			Tuples: []any{simpleStruct{}},
			WantArgs: Arguments{{
				Type: abi.Type{
					T:             abi.TupleTy,
					TupleRawName:  "simpleStruct",
					TupleElems:    []*abi.Type{&typeUint256, &typeAddress},
					TupleRawNames: []string{"amount", "token"},
				},
			}},
		},
		{
			Input:  "simpleStructWithoutTags",
			Tuples: []any{simpleStructWithoutTags{}},
			WantArgs: Arguments{{
				Type: abi.Type{
					T:             abi.TupleTy,
					TupleRawName:  "simpleStructWithoutTags",
					TupleElems:    []*abi.Type{&typeUint256, &typeAddress},
					TupleRawNames: []string{"amount", "token"},
				},
			}},
		},
		{
			Input:  "simpleStruct[]",
			Tuples: []any{simpleStruct{}},
			WantArgs: Arguments{{
				Type: abi.Type{
					T: abi.SliceTy,
					Elem: &abi.Type{
						T:             abi.TupleTy,
						TupleRawName:  "simpleStruct",
						TupleElems:    []*abi.Type{&typeUint256, &typeAddress},
						TupleRawNames: []string{"amount", "token"},
					},
				},
			}},
		},
		{
			Input:  "simpleStruct[3]",
			Tuples: []any{simpleStruct{}},
			WantArgs: Arguments{{
				Type: abi.Type{
					T:    abi.ArrayTy,
					Size: 3,
					Elem: &abi.Type{
						T:             abi.TupleTy,
						TupleRawName:  "simpleStruct",
						TupleElems:    []*abi.Type{&typeUint256, &typeAddress},
						TupleRawNames: []string{"amount", "token"},
					},
				},
			}},
		},
		{
			Input:  "uint256,simpleStruct,address",
			Tuples: []any{simpleStruct{}},
			WantArgs: Arguments{
				{Type: typeUint256},
				{
					Type: abi.Type{
						T:             abi.TupleTy,
						TupleRawName:  "simpleStruct",
						TupleElems:    []*abi.Type{&typeUint256, &typeAddress},
						TupleRawNames: []string{"amount", "token"},
					},
				},
				{Type: typeAddress},
			},
		},
		{
			Input:  "emptyStruct",
			Tuples: []any{emptyStruct{}},
			WantArgs: Arguments{{
				Type: abi.Type{
					T:             abi.TupleTy,
					TupleRawName:  "emptyStruct",
					TupleElems:    []*abi.Type{},
					TupleRawNames: []string{},
				},
			}},
		},
		{
			Input:   "unknownStruct",
			Tuples:  []any{simpleStruct{}},
			WantErr: errors.New(`syntax error: unexpected "unknownStruct", expecting type`),
		},
		{
			Input:   "simpleStruct",
			Tuples:  []any{}, // no tuples provided
			WantErr: errors.New(`syntax error: unexpected "simpleStruct", expecting type`),
		},
		{
			Input:   "simpleStruct",
			Tuples:  []any{simpleStruct{}, simpleStruct{}},
			WantErr: errors.New(`syntax error: duplicate tuple definition: simpleStruct`),
		},
		{
			Input:  "nestedStruct",
			Tuples: []any{nestedStruct{}},
			WantArgs: Arguments{{
				Type: abi.Type{
					T:            abi.TupleTy,
					TupleRawName: "nestedStruct",
					TupleElems: []*abi.Type{
						{
							T:             abi.TupleTy,
							TupleRawName:  "simpleStruct",
							TupleElems:    []*abi.Type{&typeUint256, &typeAddress},
							TupleRawNames: []string{"amount", "token"},
						},
						&typeBool,
					},
					TupleRawNames: []string{"inner", "active"},
				},
			}},
		},
		{
			Input:  "nestedStruct[]",
			Tuples: []any{nestedStruct{}},
			WantArgs: Arguments{{
				Type: abi.Type{
					T: abi.SliceTy,
					Elem: &abi.Type{
						T:            abi.TupleTy,
						TupleRawName: "nestedStruct",
						TupleElems: []*abi.Type{
							{
								T:             abi.TupleTy,
								TupleRawName:  "simpleStruct",
								TupleElems:    []*abi.Type{&typeUint256, &typeAddress},
								TupleRawNames: []string{"amount", "token"},
							},
							&typeBool,
						},
						TupleRawNames: []string{"inner", "active"},
					},
				},
			}},
		},
		{
			Input:  "complexStruct",
			Tuples: []any{complexStruct{}},
			WantArgs: Arguments{{
				Type: abi.Type{
					T:            abi.TupleTy,
					TupleRawName: "complexStruct",
					TupleElems: []*abi.Type{
						&typeUint256,
						&typeAddress,
						{T: abi.BytesTy},
						{
							T:            abi.TupleTy,
							TupleRawName: "nestedStruct",
							TupleElems: []*abi.Type{
								{
									T:             abi.TupleTy,
									TupleRawName:  "simpleStruct",
									TupleElems:    []*abi.Type{&typeUint256, &typeAddress},
									TupleRawNames: []string{"amount", "token"},
								},
								&typeBool,
							},
							TupleRawNames: []string{"inner", "active"},
						},
					},
					TupleRawNames: []string{"id", "owner", "data", "metadata"},
				},
			}},
		},
		{
			Input:  "complexStruct[2]",
			Tuples: []any{complexStruct{}},
			WantArgs: Arguments{{
				Type: abi.Type{
					T:    abi.ArrayTy,
					Size: 2,
					Elem: &abi.Type{
						T:            abi.TupleTy,
						TupleRawName: "complexStruct",
						TupleElems: []*abi.Type{
							&typeUint256,
							&typeAddress,
							{T: abi.BytesTy},
							{
								T:            abi.TupleTy,
								TupleRawName: "nestedStruct",
								TupleElems: []*abi.Type{
									{
										T:             abi.TupleTy,
										TupleRawName:  "simpleStruct",
										TupleElems:    []*abi.Type{&typeUint256, &typeAddress},
										TupleRawNames: []string{"amount", "token"},
									},
									&typeBool,
								},
								TupleRawNames: []string{"inner", "active"},
							},
						},
						TupleRawNames: []string{"id", "owner", "data", "metadata"},
					},
				},
			}},
		},
		{
			Input:  "arrayStruct",
			Tuples: []any{arrayStruct{}},
			WantArgs: Arguments{{
				Type: abi.Type{
					T:            abi.TupleTy,
					TupleRawName: "arrayStruct",
					TupleElems: []*abi.Type{
						{T: abi.SliceTy, Elem: &typeUint256},
						&typeUint256,
					},
					TupleRawNames: []string{"values", "count"},
				},
			}},
		},
		{
			Input:  "arrayStruct[]",
			Tuples: []any{arrayStruct{}},
			WantArgs: Arguments{{
				Type: abi.Type{
					T: abi.SliceTy,
					Elem: &abi.Type{
						T:            abi.TupleTy,
						TupleRawName: "arrayStruct",
						TupleElems: []*abi.Type{
							{T: abi.SliceTy, Elem: &typeUint256},
							&typeUint256,
						},
						TupleRawNames: []string{"values", "count"},
					},
				},
			}},
		},
		{
			Input:  "uint256,nestedStruct,complexStruct,arrayStruct",
			Tuples: []any{nestedStruct{}, complexStruct{}, arrayStruct{}},
			WantArgs: Arguments{
				{Type: typeUint256},
				{
					Type: abi.Type{
						T:            abi.TupleTy,
						TupleRawName: "nestedStruct",
						TupleElems: []*abi.Type{
							{
								T:             abi.TupleTy,
								TupleRawName:  "simpleStruct",
								TupleElems:    []*abi.Type{&typeUint256, &typeAddress},
								TupleRawNames: []string{"amount", "token"},
							},
							&typeBool,
						},
						TupleRawNames: []string{"inner", "active"},
					},
				},
				{
					Type: abi.Type{
						T:            abi.TupleTy,
						TupleRawName: "complexStruct",
						TupleElems: []*abi.Type{
							&typeUint256,
							&typeAddress,
							{T: abi.BytesTy},
							{
								T:            abi.TupleTy,
								TupleRawName: "nestedStruct",
								TupleElems: []*abi.Type{
									{
										T:             abi.TupleTy,
										TupleRawName:  "simpleStruct",
										TupleElems:    []*abi.Type{&typeUint256, &typeAddress},
										TupleRawNames: []string{"amount", "token"},
									},
									&typeBool,
								},
								TupleRawNames: []string{"inner", "active"},
							},
						},
						TupleRawNames: []string{"id", "owner", "data", "metadata"},
					},
				},
				{
					Type: abi.Type{
						T:            abi.TupleTy,
						TupleRawName: "arrayStruct",
						TupleElems: []*abi.Type{
							{T: abi.SliceTy, Elem: &typeUint256},
							&typeUint256,
						},
						TupleRawNames: []string{"values", "count"},
					},
				},
			},
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%d_%s", i, test.Input), func(t *testing.T) {
			gotArgs, gotErr := Parse(test.Input, test.Tuples...)
			if diff := cmp.Diff(test.WantErr, gotErr,
				internal.EquateErrors(),
			); diff != "" {
				t.Fatalf("Err (-want, +got):\n%s", diff)
			}
			if diff := cmp.Diff(test.WantArgs, gotArgs,
				cmpopts.EquateEmpty(),
				cmpopts.IgnoreUnexported(abi.Type{}),
				cmpopts.IgnoreFields(abi.Type{}, "TupleType")); diff != "" {
				t.Errorf("Args (-want, +got):\n%s", diff)
			}
		})
	}
}

func TestParseArgsWithName(t *testing.T) {
	tests := []struct {
		Input    string
		Tuples   []any
		WantArgs Arguments
		WantName string
		WantErr  error
	}{
		{
			Input:   "",
			WantErr: errors.New(`syntax error: unexpected EOF, expecting name`),
		},
		{
			Input:   "uint",
			WantErr: errors.New(`syntax error: unexpected EOF, expecting "("`),
		},
		{
			Input:    "f()",
			WantName: "f",
		},
		{
			Input:   "f(",
			WantErr: errors.New(`syntax error: unexpected EOF, expecting type`),
		},
		{
			Input:   "f(uint256",
			WantErr: errors.New(`syntax error: unexpected EOF, want "," or ")"`),
		},
		{
			Input:   "f(uint256 indexed",
			WantErr: errors.New(`syntax error: unexpected EOF, want "," or ")"`),
		},
		{
			Input:   "f(uint256 arg0",
			WantErr: errors.New(`syntax error: unexpected EOF, want "," or ")"`),
		},
		{
			Input:   "f(uint256 indexed arg0",
			WantErr: errors.New(`syntax error: unexpected EOF, want "," or ")"`),
		},
		{
			Input:   "f(uint256,",
			WantErr: errors.New(`syntax error: unexpected EOF, expecting type`),
		},
		{
			Input:    "transfer(address,uint256)",
			WantArgs: Arguments{{Type: typeAddress}, {Type: typeUint256}},
			WantName: "transfer",
		},
		{
			Input:    "transfer(address recipient, uint256 amount)",
			WantArgs: Arguments{{Type: typeAddress, Name: "recipient"}, {Type: typeUint256, Name: "amount"}},
			WantName: "transfer",
		},
		{
			Input: "exactInputSingle((address tokenIn, address tokenOut, uint24 fee, address recipient, uint256 deadline, uint256 amountIn, uint256 amountOutMinimum, uint160 sqrtPriceLimitX96) params)",
			WantArgs: Arguments{{
				Type: abi.Type{
					T:             abi.TupleTy,
					TupleElems:    []*abi.Type{&typeAddress, &typeAddress, &typeUint24, &typeAddress, &typeUint256, &typeUint256, &typeUint256, &typeUint160},
					TupleRawNames: []string{"tokenIn", "tokenOut", "fee", "recipient", "deadline", "amountIn", "amountOutMinimum", "sqrtPriceLimitX96"},
				},
				Name: "params",
			}},
			WantName: "exactInputSingle",
		},
		{
			Input:    "transfer(simpleStruct)",
			Tuples:   []any{simpleStruct{}},
			WantName: "transfer",
			WantArgs: Arguments{{
				Type: abi.Type{
					T:             abi.TupleTy,
					TupleRawName:  "simpleStruct",
					TupleElems:    []*abi.Type{&typeUint256, &typeAddress},
					TupleRawNames: []string{"amount", "token"},
				},
			}},
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%d_%s", i, test.Input), func(t *testing.T) {
			gotName, gotArgs, gotErr := ParseWithName(test.Input, test.Tuples...)
			if diff := cmp.Diff(test.WantErr, gotErr,
				internal.EquateErrors(),
			); diff != "" {
				t.Fatalf("Err (-want, +got):\n%s", diff)
			}
			if test.WantName != gotName {
				t.Errorf("Name want: %s, got: %s", test.WantName, gotName)
			}
			if diff := cmp.Diff(test.WantArgs, gotArgs,
				cmpopts.EquateEmpty(),
				cmpopts.IgnoreUnexported(abi.Type{}),
				cmpopts.IgnoreFields(abi.Type{}, "TupleType")); diff != "" {
				t.Errorf("Args (-want, +got):\n%s", diff)
			}
		})
	}
}

func BenchmarkParseArgsWithName(b *testing.B) {
	benchmarks := []struct {
		Name  string
		Input string
	}{
		{Name: "symbol", Input: "symbol()"},
		{Name: "transfer", Input: "transfer(address,uint256)"},
		{Name: "exactInputSingle", Input: "exactInputSingle((address tokenIn, address tokenOut, uint24 fee, address recipient, uint256 deadline, uint256 amountIn, uint256 amountOutMinimum, uint160 sqrtPriceLimitX96) params)"},
		{Name: "swap", Input: `swap(
			(address currency0, address currency1, uint24 fee, int24 tickSpacing, address hooks) key,
			(bool zeroForOne, int256 amountSpecified, uint160 sqrtPriceLimitX96) params,
			bytes hookData
		)`},
		{Name: "settle", Input: `settle(
			address[] tokens,
			uint256[] clearingPrices,
			(
				uint256 sellTokenIndex,
				uint256 buyTokenIndex,
				address receiver,
				uint256 sellAmount,
				uint256 buyAmount,
				uint32 validTo,
				bytes32 appData,
				uint256 feeAmount,
				uint256 flags,
				uint256 executedAmount,
				bytes signature
			)[] trades,
			(address target, uint256 value, bytes callData)[][3] interactions
		)`},
	}

	for _, bench := range benchmarks {
		b.Run(bench.Name, func(b *testing.B) {
			b.ReportAllocs()
			for range b.N {
				ParseWithName(bench.Input)
			}
		})
	}
}

// test structs for tuple functionality
type simpleStruct struct {
	Amount *big.Int       `abitype:"uint256"`
	Token  common.Address `abitype:"address"`
}

type simpleStructWithoutTags struct {
	Amount *big.Int
	Token  common.Address
}

type nestedStruct struct {
	Inner  simpleStruct
	Active bool
}

type complexStruct struct {
	ID       *big.Int `abitype:"uint256"`
	Owner    common.Address
	Data     []byte
	Metadata nestedStruct
}

type arrayStruct struct {
	Values []*big.Int `abitype:"uint256"`
	Count  *big.Int   `abitype:"uint256"`
}

type emptyStruct struct{}
