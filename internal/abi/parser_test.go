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
	typeAddress = abi.Type{T: abi.AddressTy, Size: 20}
	typeBool    = abi.Type{T: abi.BoolTy}
	typeUint24  = abi.Type{T: abi.UintTy, Size: 24}
	typeUint160 = abi.Type{T: abi.UintTy, Size: 160}
	typeUint256 = abi.Type{T: abi.UintTy, Size: 256}
)

func TestParseArgs(t *testing.T) {
	tests := []struct {
		Input    string
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
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			gotArgs, gotErr := Parse(test.Input)
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
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			gotName, gotArgs, gotErr := ParseWithName(test.Input)
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
