package abi

import (
	"errors"
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/lmittmann/w3/internal"
)

func TestLexer(t *testing.T) {
	tests := []struct {
		Input     string
		WantItems []*item
		WantErr   error
	}{
		{Input: "", WantItems: []*item{{itemTypeEOF, ""}}},
		{Input: "uint256", WantItems: []*item{{itemTypeID, "uint256"}, {itemTypeEOF, ""}}},
		{Input: "uint256[1]", WantItems: []*item{{itemTypeID, "uint256"}, {itemTypePunct, "["}, {itemTypeNum, "1"}, {itemTypePunct, "]"}, {itemTypeEOF, ""}}},
		{Input: "uint balance", WantItems: []*item{{itemTypeID, "uint"}, {itemTypeID, "balance"}, {itemTypeEOF, ""}}},
		{Input: "1", WantItems: []*item{{itemTypeNum, "1"}, {itemTypeEOF, ""}}},

		{Input: "0", WantErr: errors.New("unexpected character: 0")},
		{Input: "uint256[0]", WantErr: errors.New("unexpected character: 0")},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			items, err := lex(test.Input)
			if diff := cmp.Diff(test.WantErr, err, internal.EquateErrors()); diff != "" {
				t.Fatalf("Err: (-want, +got)\n%s", diff)
			}
			if diff := cmp.Diff(test.WantItems, items, cmpopts.EquateEmpty()); diff != "" {
				t.Fatalf("Items: (-want, +got)\n%s", diff)
			}
		})
	}
}

func lex(input string) ([]*item, error) {
	l := newLexer(input)

	var items []*item
	for {
		item, err := l.nextItem()
		if err != nil {
			return nil, err
		}

		items = append(items, item)
		if item.Typ == itemTypeEOF {
			break
		}
	}
	return items, nil
}
