package abi

import (
	"strconv"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

var (
	typeUint256, _ = abi.NewType("uint256", "", nil)
	typeAddress, _ = abi.NewType("address", "", nil)
)

func TestParser(t *testing.T) {
	t.Parallel()

	tests := []struct {
		Input    string
		WantArgs abi.Arguments
		WantErr  error
	}{
		{
			Input:    "uint256",
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
			Input:    "transfer(address,uint256)",
			WantArgs: abi.Arguments{{Type: typeAddress}, {Type: typeUint256}},
		},
		{
			Input:    "transfer(address recipient, uint256 amount)",
			WantArgs: abi.Arguments{{Type: typeAddress, Name: "recipient"}, {Type: typeUint256, Name: "amount"}},
		},
		{
			Input:    "fee()",
			WantArgs: nil,
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
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			gotArgs, gotErr := parse(test.Input)
			if diff := cmp.Diff(test.WantErr, gotErr); diff != "" {
				t.Fatalf("Err (-want, +got):\n%s", diff)
			}
			if diff := cmp.Diff(test.WantArgs, gotArgs,
				cmpopts.IgnoreUnexported(abi.Type{}),
				cmpopts.IgnoreFields(abi.Type{}, "TupleType")); diff != "" {
				t.Fatalf("Args (-want, +got):\n%s", diff)
			}
		})
	}
}

func parse(input string) (abi.Arguments, error) {
	itemCh := make(chan item, 1)
	l := newLexer(input, itemCh)
	go l.run()

	p := newParser(itemCh)
	if err := p.run(); err != nil {
		return nil, err
	}

	return p.args, nil
}

func tuple(types ...abi.ArgumentMarshaling) abi.Type {
	typ, err := abi.NewType("tuple", "", types)
	if err != nil {
		panic(err.Error())
	}
	return typ
}

func TestParserSingnature(t *testing.T) {
	t.Parallel()

	tests := []struct {
		Parser       *parser
		WantSig      string
		WantFuncName string
	}{
		{
			Parser:  &parser{items: []item{{itemTyp, "uint256"}, iEOF}},
			WantSig: "uint256",
		},
		{
			Parser:  &parser{items: []item{{itemTyp, "uint256"}, {itemID, "balance"}, iEOF}},
			WantSig: "uint256",
		},
		{
			Parser:  &parser{items: []item{{itemTyp, "address"}, iDelim, {itemTyp, "uint256"}, iEOF}},
			WantSig: "address,uint256",
		},
		{
			Parser:  &parser{items: []item{{itemTyp, "address"}, {itemID, "who"}, iDelim, {itemTyp, "uint256"}, {itemID, "balance"}, iEOF}},
			WantSig: "address,uint256",
		},
		{
			Parser:  &parser{items: []item{iLeftParen, iRightParen, iEOF}},
			WantSig: "()",
		},
		{
			Parser:  &parser{items: []item{iLeftParen, {itemTyp, "uint256"}, iRightParen, iEOF}},
			WantSig: "(uint256)",
		},
		{
			Parser:  &parser{items: []item{iLeftParen, {itemTyp, "uint256"}, {itemID, "balance"}, iRightParen, iEOF}},
			WantSig: "(uint256)",
		},
		{
			Parser:       &parser{items: []item{{itemID, "balanceOf"}, iLeftParen, {itemTyp, "address"}, iRightParen, iEOF}},
			WantSig:      "balanceOf(address)",
			WantFuncName: "balanceOf",
		},
		{
			Parser:       &parser{items: []item{{itemID, "balanceOf"}, iLeftParen, {itemTyp, "address"}, {itemID, "who"}, iRightParen, iEOF}},
			WantSig:      "balanceOf(address)",
			WantFuncName: "balanceOf",
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			gotSig, gotFuncName := test.Parser.singnature()
			if test.WantSig != gotSig {
				t.Fatalf("Sig: want %s, got %s", gotSig, test.WantSig)
			}
			if test.WantFuncName != gotFuncName {
				t.Fatalf("FuncName: want %s, got %s", gotFuncName, test.WantFuncName)
			}
		})
	}
}
