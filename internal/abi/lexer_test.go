package abi

import (
	"errors"
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestLexer(t *testing.T) {
	t.Parallel()

	tests := []struct {
		Input     string
		WantItems []*item
		WantErr   error
	}{
		{Input: "", WantItems: []*item{}},
		{Input: "uint256", WantItems: []*item{{itemTypeID, "uint256"}}},
		{Input: "uint256[1]", WantItems: []*item{{itemTypeID, "uint256"}, {itemTypePunct, "["}, {itemTypeNum, "1"}, {itemTypePunct, "]"}}},
		{Input: "uint balance", WantItems: []*item{{itemTypeID, "uint"}, {itemTypeID, "balance"}}},
		{Input: "1", WantItems: []*item{{itemTypeNum, "1"}}},

		{Input: "0", WantErr: errors.New("unexpected character: 0")},
		{Input: "uint256[0]", WantErr: errors.New("unexpected character: 0")},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			items, err := lex(test.Input)
			if diff := cmp.Diff(test.WantErr, err, cmp.Comparer(equateErrors)); diff != "" {
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
		} else if item == nil {
			break
		}
		items = append(items, item)
	}
	return items, nil
}

// equateErrors compares two errors by their message.
func equateErrors(x, y error) bool {
	return x != nil && y != nil && x.Error() == y.Error()
}
