package abi

import (
	"errors"
	"strconv"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/lmittmann/w3/internal"
)

var (
	typeUint256 = abi.Type{T: abi.UintTy, Size: 256}
	typeAddress = abi.Type{T: abi.AddressTy, Size: 20}
)

func TestParser(t *testing.T) {
	t.Parallel()

	tests := []struct {
		Input        string
		WantArgs     abi.Arguments
		WantFuncName string
		WantErr      error
	}{
		{
			Input:    "",
			WantArgs: abi.Arguments{},
		},
		{
			Input:    "uint256",
			WantArgs: abi.Arguments{{Type: typeUint256}},
		},
		{
			Input:    "uint",
			WantArgs: abi.Arguments{{Type: typeUint256}},
		},
		{
			Input:    "uint256 balance",
			WantArgs: abi.Arguments{{Type: typeUint256, Name: "balance"}},
		},
		{
			Input:    "address,uint256",
			WantArgs: abi.Arguments{{Type: typeAddress}, {Type: typeUint256}},
		},
		{
			Input:    "address recipient, uint256 amount",
			WantArgs: abi.Arguments{{Type: typeAddress, Name: "recipient"}, {Type: typeUint256, Name: "amount"}},
		},
		{
			Input:    "uint256[]",
			WantArgs: abi.Arguments{{Type: abi.Type{Elem: &typeUint256, T: abi.SliceTy}}},
		},
		{
			Input: "uint256[][]",
			WantArgs: abi.Arguments{{
				Type: abi.Type{
					Elem: &abi.Type{Elem: &typeUint256, T: abi.SliceTy},
					T:    abi.SliceTy,
				},
			}},
		},
		{
			Input: "uint256[3][]",
			WantArgs: abi.Arguments{{
				Type: abi.Type{
					Elem: &abi.Type{Elem: &typeUint256, T: abi.ArrayTy, Size: 3},
					T:    abi.SliceTy,
				},
			}},
		},
		{
			Input: "(address arg0, uint256 arg1)",
			WantArgs: abi.Arguments{{
				Type: tuple(
					abi.ArgumentMarshaling{Type: "address", Name: "arg0"},
					abi.ArgumentMarshaling{Type: "uint256", Name: "arg1"},
				),
			}},
		},
		{
			Input:        "transfer(address,uint256)",
			WantArgs:     abi.Arguments{{Type: typeAddress}, {Type: typeUint256}},
			WantFuncName: "transfer",
		},
		{
			Input:        "transfer(address recipient, uint256 amount)",
			WantArgs:     abi.Arguments{{Type: typeAddress, Name: "recipient"}, {Type: typeUint256, Name: "amount"}},
			WantFuncName: "transfer",
		},
		{
			Input:        "fee()",
			WantArgs:     nil,
			WantFuncName: "fee",
		},
		{
			Input: "exactInputSingle((address tokenIn, address tokenOut, uint24 fee, address recipient, uint256 deadline, uint256 amountIn, uint256 amountOutMinimum, uint160 sqrtPriceLimitX96) params)",
			WantArgs: abi.Arguments{
				{
					Type: tuple(
						abi.ArgumentMarshaling{Type: "address", Name: "tokenIn"},
						abi.ArgumentMarshaling{Type: "address", Name: "tokenOut"},
						abi.ArgumentMarshaling{Type: "uint24", Name: "fee"},
						abi.ArgumentMarshaling{Type: "address", Name: "recipient"},
						abi.ArgumentMarshaling{Type: "uint256", Name: "deadline"},
						abi.ArgumentMarshaling{Type: "uint256", Name: "amountIn"},
						abi.ArgumentMarshaling{Type: "uint256", Name: "amountOutMinimum"},
						abi.ArgumentMarshaling{Type: "uint160", Name: "sqrtPriceLimitX96"},
					),
					Name: "params",
				},
			},
			WantFuncName: "exactInputSingle",
		},
		{
			Input:   "xxx",
			WantErr: errors.New(`lex error: unknown type "xxx"`),
		},
		{
			Input:   "f(",
			WantErr: errors.New(`unexpected EOF after '('`),
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			gotName, gotArgs, gotErr := parse(test.Input)
			if diff := cmp.Diff(test.WantErr, gotErr,
				internal.EquateErrors(),
			); diff != "" {
				t.Fatalf("Err (-want, +got):\n%s", diff)
			}
			if test.WantFuncName != gotName {
				t.Errorf("FuncName want: %s, got: %s", test.WantFuncName, gotName)
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

func tuple(types ...abi.ArgumentMarshaling) abi.Type {
	typ, err := abi.NewType("tuple", "", types)
	if err != nil {
		panic(err.Error())
	}
	return typ
}
