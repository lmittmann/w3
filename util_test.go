package w3

import (
	"math/big"
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestI(t *testing.T) {
	t.Parallel()

	tests := []struct {
		IntString string
		Want      *big.Int
	}{
		{"0", big.NewInt(0)},
		{"0x0", big.NewInt(0)},
		{"1", big.NewInt(1)},
		{"0x1", big.NewInt(1)},
		{"255", big.NewInt(255)},
		{"0xff", big.NewInt(255)},
		{"test", nil},
		{"0xtest", nil},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			got := I(test.IntString)
			if diff := cmp.Diff(test.Want, got, cmp.AllowUnexported(big.Int{})); diff != "" {
				t.Fatalf("(-want, +got)\n%s", diff)
			}
		})
	}
}
