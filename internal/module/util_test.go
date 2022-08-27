package module

import (
	"math/big"
	"strconv"
	"testing"
)

func TestBlockNumberArg(t *testing.T) {
	tests := []struct {
		Number *big.Int
		Want   string
	}{
		{nil, "latest"},
		{big.NewInt(1), "0x1"},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			got := BlockNumberArg(test.Number)
			if test.Want != got {
				t.Errorf("want %q, got %q", test.Want, got)
			}
		})
	}
}
