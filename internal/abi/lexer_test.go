package abi

import (
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"
)

var (
	iEOF        = item{itemEOF, ""}
	iDelim      = item{itemDelim, ","}
	iLeftParen  = item{itemLeftParen, "("}
	iRightParen = item{itemRightParen, ")"}
)

func TestLexer(t *testing.T) {
	t.Parallel()

	tests := []struct {
		Input     string
		WantItems []item
	}{
		{"", []item{iEOF}},

		// single type
		{"uint256", []item{{itemTyp, "uint256"}, iEOF}},
		{"uint", []item{{itemTyp, "uint256"}, iEOF}},
		{"uint7", []item{{itemError, `unknown type "uint7"`}}},
		{"UINT256", []item{{itemError, `unknown type "UINT256"`}}},
		{"uint256 balance", []item{{itemTyp, "uint256"}, {itemID, "balance"}, iEOF}},
		{"uint256 Balance", []item{{itemTyp, "uint256"}, {itemID, "Balance"}, iEOF}},
		{"uint256 uint256", []item{{itemTyp, "uint256"}, {itemID, "uint256"}, iEOF}},
		{"uint256,", []item{{itemTyp, "uint256"}, iDelim, {itemError, "unexpected EOF after ','"}}},
		{"uint256 balance,", []item{{itemTyp, "uint256"}, {itemID, "balance"}, iDelim, {itemError, "unexpected EOF after ','"}}},

		// single array type
		{"uint256[]", []item{{itemTyp, "uint256[]"}, iEOF}},
		{"uint[]", []item{{itemTyp, "uint256[]"}, iEOF}},
		{"uint256[", []item{{itemError, "unexpected EOF, want ']'"}}},
		{"uint256[1", []item{{itemError, "unexpected EOF, want ']'"}}},
		{"uint256[X", []item{{itemError, "unexpected token 'X', want ']'"}}},
		{"uint256[7]", []item{{itemTyp, "uint256[7]"}, iEOF}},

		// multi-dimensional array type
		{"uint256[][]", []item{{itemTyp, "uint256[][]"}, iEOF}},
		{"uint256[][", []item{{itemError, "unexpected EOF, want ']'"}}},
		{"uint256[3][]", []item{{itemTyp, "uint256[3][]"}, iEOF}},

		// multiple types
		{"address,int,uint", []item{{itemTyp, "address"}, iDelim, {itemTyp, "int256"}, iDelim, {itemTyp, "uint256"}, iEOF}},
		{"address, int,  uint", []item{{itemTyp, "address"}, iDelim, {itemTyp, "int256"}, iDelim, {itemTyp, "uint256"}, iEOF}},
		{"address who, uint balance", []item{{itemTyp, "address"}, {itemID, "who"}, iDelim, {itemTyp, "uint256"}, {itemID, "balance"}, iEOF}},

		// tuple
		{"()", []item{iLeftParen, iRightParen, iEOF}},
		{"(uint256)", []item{iLeftParen, {itemTyp, "uint256"}, iRightParen, iEOF}},
		{"(uint256 balance)", []item{iLeftParen, {itemTyp, "uint256"}, {itemID, "balance"}, iRightParen, iEOF}},
		{"(uint256 balance,uint256)", []item{iLeftParen, {itemTyp, "uint256"}, {itemID, "balance"}, iDelim, {itemTyp, "uint256"}, iRightParen, iEOF}},
		{"(uint256 balance),uint256", []item{iLeftParen, {itemTyp, "uint256"}, {itemID, "balance"}, iRightParen, iDelim, {itemTyp, "uint256"}, iEOF}},

		// function
		{"fn()", []item{{itemID, "fn"}, iLeftParen, iRightParen, iEOF}},
		{"fn(", []item{{itemID, "fn"}, iLeftParen, {itemError, "unexpected EOF after '('"}}},
		{"uint256()", []item{{itemID, "uint256"}, iLeftParen, iRightParen, iEOF}},
		{"balanceOf(address)", []item{{itemID, "balanceOf"}, iLeftParen, {itemTyp, "address"}, iRightParen, iEOF}},
		{"balanceOf(address who)", []item{{itemID, "balanceOf"}, iLeftParen, {itemTyp, "address"}, {itemID, "who"}, iRightParen, iEOF}},

		// function with tuple
		{"fn((address))", []item{{itemID, "fn"}, iLeftParen, iLeftParen, {itemTyp, "address"}, iRightParen, iRightParen, iEOF}},
		{"fn((address who) param)", []item{{itemID, "fn"}, iLeftParen, iLeftParen, {itemTyp, "address"}, {itemID, "who"}, iRightParen, {itemID, "param"}, iRightParen, iEOF}},
		{"fn(uint256,(address))", []item{{itemID, "fn"}, iLeftParen, {itemTyp, "uint256"}, iDelim, iLeftParen, {itemTyp, "address"}, iRightParen, iRightParen, iEOF}},
		{"fn(uint256 value, (address who) param)", []item{{itemID, "fn"}, iLeftParen, {itemTyp, "uint256"}, {itemID, "value"}, iDelim, iLeftParen, {itemTyp, "address"}, {itemID, "who"}, iRightParen, {itemID, "param"}, iRightParen, iEOF}},
		{"fn((address),uint256)", []item{{itemID, "fn"}, iLeftParen, iLeftParen, {itemTyp, "address"}, iRightParen, iDelim, {itemTyp, "uint256"}, iRightParen, iEOF}},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			gotItems := lex(test.Input)
			if diff := cmp.Diff(test.WantItems, gotItems); diff != "" {
				t.Fatalf("(-want, +got)\n%s", diff)
			}
		})
	}
}

func lex(input string) (items []item) {
	itemCh := make(chan item, 1)
	l := newLexer(input, itemCh)
	go l.run()

	for item := range itemCh {
		items = append(items, item)
	}
	return
}
