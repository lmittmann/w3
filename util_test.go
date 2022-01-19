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
		{"0x0", big.NewInt(0)},
		{"0x1", big.NewInt(1)},
		{"0xff", big.NewInt(255)},
		{"0xtest", nil},
		{"0", big.NewInt(0)},
		{"1", big.NewInt(1)},
		{"255", big.NewInt(255)},
		{"test", nil},
		{"1 ether", big.NewInt(1_000000000_000000000)},
		{"1 eth", big.NewInt(1_000000000_000000000)},
		{"1ether", nil},
		{"1.2 ether", big.NewInt(1_200000000_000000000)},
		{"01.2 ether", big.NewInt(1_200000000_000000000)},
		{"1.20 ether", big.NewInt(1_200000000_000000000)},
		{"1.200000000000000003 ether", big.NewInt(1_200000000_000000003)},
		{"1.2000000000000000030 ether", big.NewInt(1_200000000_000000003)},
		{"1.2000000000000000034 ether", big.NewInt(1_200000000_000000003)},
		{"1 gwei", big.NewInt(1_000000000)},
		{"1.2 gwei", big.NewInt(1_200000000)},
		{"1.200000003 gwei", big.NewInt(1_200000003)},
		{"1.2000000034 gwei", big.NewInt(1_200000003)},
		{".", big.NewInt(0)},
		{". ether", big.NewInt(0)},
		{"1.", big.NewInt(1)},
		{"1. ether", big.NewInt(1_000000000_000000000)},
		{".1", big.NewInt(0)},
		{".1 ether", big.NewInt(100000000_000000000)},
		{"0.1 ether", big.NewInt(100000000_000000000)},
		{"0.10 ether", big.NewInt(100000000_000000000)},
		{"00.10 ether", big.NewInt(100000000_000000000)},
		{" 1 ether", nil},
		{"1 ether ", nil},
		{"1  ether", nil},
		{"-1", nil},
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

func BenchmarkI(b *testing.B) {
	benchmarks := []string{
		"0x123456",
		"1.23456 ether",
		"1.000000000000000000 ether",
		"1.000000000000000000123456 ether",
	}

	for _, bench := range benchmarks {
		b.Run(bench, func(b *testing.B) {
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				I(bench)
			}
		})
	}
}
