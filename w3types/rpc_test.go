package w3types_test

import (
	"math/big"
	"strconv"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/google/go-cmp/cmp"
	"github.com/lmittmann/w3/w3types"
)

func TestBlockOverrides_MarshalJSON(t *testing.T) {
	tests := []struct {
		Overrides *w3types.BlockOverrides
		Want      string
	}{
		{
			Overrides: &w3types.BlockOverrides{},
			Want:      `{}`,
		},
		{
			Overrides: &w3types.BlockOverrides{
				Number: big.NewInt(1),
			},
			Want: `{"number":"0x1"}`,
		},
		{
			Overrides: &w3types.BlockOverrides{
				FeeRecipient: common.Address{0xc0, 0xfe},
			},
			Want: `{"feeRecipient":"0xc0fe000000000000000000000000000000000000"}`,
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			got, err := test.Overrides.MarshalJSON()
			if err != nil {
				t.Fatalf("Err: %v", err)
			}
			if diff := cmp.Diff(test.Want, string(got)); diff != "" {
				t.Fatalf("(-want, +got)\n%s", diff)
			}
		})
	}
}
