package w3

import (
	"bytes"
	"math/big"
	"strconv"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func TestA(t *testing.T) {
	t.Parallel()

	tests := []struct {
		HexAddress  string
		WantPanic   string
		WantAddress common.Address
	}{
		{
			HexAddress:  "0x000000000000000000000000000000000000c0Fe",
			WantAddress: common.HexToAddress("0x000000000000000000000000000000000000c0Fe"),
		},
		{HexAddress: "000000000000000000000000000000000000c0Fe", WantPanic: `hex address "000000000000000000000000000000000000c0Fe" must have 0x prefix`},
		{HexAddress: "0xcofe", WantPanic: `hex address "0xcofe" must be hex`},
		{HexAddress: "0xc0Fe", WantPanic: `hex address "0xc0Fe" must have 20 bytes`},
		{HexAddress: "0x000000000000000000000000000000000000c0fe", WantPanic: `hex address "0x000000000000000000000000000000000000c0fe" must be checksum encoded`},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			defer func() {
				var gotPanic string
				if r := recover(); r != nil {
					gotPanic = r.(string)
				}
				if test.WantPanic != gotPanic {
					t.Fatalf("want %q, got %q", test.WantPanic, gotPanic)
				}
			}()

			gotAddr := A(test.HexAddress)
			if test.WantPanic == "" && test.WantAddress != gotAddr {
				t.Fatalf("want: %s, got: %s", test.WantAddress, gotAddr)
			}
		})
	}
}

func TestB(t *testing.T) {
	t.Parallel()

	tests := []struct {
		HexBytes  string
		WantPanic string
		WantBytes []byte
	}{
		{
			HexBytes:  "0xc0fe",
			WantBytes: []byte{0xc0, 0xfe},
		},
		{HexBytes: "c0fe", WantPanic: `hex bytes "c0fe" must have 0x prefix`},
		{HexBytes: "0xcofe", WantPanic: `hex bytes "0xcofe" must be hex`},
		{HexBytes: "0xc0f", WantPanic: `hex bytes "0xc0f" must have even number of hex chars`},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			defer func() {
				var gotPanic string
				if r := recover(); r != nil {
					gotPanic = r.(string)
				}
				if test.WantPanic != gotPanic {
					t.Fatalf("want %q, got %q", test.WantPanic, gotPanic)
				}
			}()

			gotBytes := B(test.HexBytes)
			if test.WantPanic == "" && !bytes.Equal(test.WantBytes, gotBytes) {
				t.Fatalf("want: %x, got: %x", test.WantBytes, gotBytes)
			}
		})
	}
}

func TestH(t *testing.T) {
	t.Parallel()

	tests := []struct {
		HexHash   string
		WantPanic string
		WantHash  common.Hash
	}{
		{
			HexHash:  "0x000000000000000000000000000000000000000000000000000000000000c0fe",
			WantHash: common.HexToHash("0x000000000000000000000000000000000000000000000000000000000000c0fe"),
		},
		{HexHash: "000000000000000000000000000000000000000000000000000000000000c0fe", WantPanic: `hex hash "000000000000000000000000000000000000000000000000000000000000c0fe" must have 0x prefix`},
		{HexHash: "0xcofe", WantPanic: `hex hash "0xcofe" must be hex`},
		{HexHash: "0xc0fe", WantPanic: `hex hash "0xc0fe" must have 32 bytes`},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			defer func() {
				var gotPanic string
				if r := recover(); r != nil {
					gotPanic = r.(string)
				}
				if test.WantPanic != gotPanic {
					t.Fatalf("want %q, got %q", test.WantPanic, gotPanic)
				}
			}()

			gotHash := H(test.HexHash)
			if test.WantPanic == "" && test.WantHash != gotHash {
				t.Fatalf("want: %s, got: %s", test.WantHash, gotHash)
			}
		})
	}
}

func TestI(t *testing.T) {
	t.Parallel()

	tests := []struct {
		StrInt    string
		WantPanic string
		WantBig   *big.Int
	}{
		// hex big's
		{StrInt: "0x0", WantBig: big.NewInt(0)},
		{StrInt: "0x1", WantBig: big.NewInt(1)},
		{StrInt: "0xff", WantBig: big.NewInt(255)},
		{StrInt: "0xO", WantPanic: `hex big "0xO" must be hex`},

		// int big's
		{StrInt: "0", WantBig: big.NewInt(0)},
		{StrInt: "1", WantBig: big.NewInt(1)},
		{StrInt: "255", WantBig: big.NewInt(255)},
		{StrInt: "X", WantPanic: `invalid str big "X": number has no digits`},

		// unit big's
		{StrInt: "0 ether", WantBig: big.NewInt(0)},
		{StrInt: "0 eth", WantBig: big.NewInt(0)},
		{StrInt: "0 gwei", WantBig: big.NewInt(0)},
		{StrInt: "1 ether", WantBig: big.NewInt(1_000000000_000000000)},
		{StrInt: "1 eth", WantBig: big.NewInt(1_000000000_000000000)},
		{StrInt: "1ether", WantPanic: `invalid str big "1ether": number has no digits`},
		{StrInt: "1.2 ether", WantBig: big.NewInt(1_200000000_000000000)},
		{StrInt: "01.2 ether", WantBig: big.NewInt(1_200000000_000000000)},
		{StrInt: "1.20 ether", WantBig: big.NewInt(1_200000000_000000000)},
		{StrInt: "1.200000000000000003 ether", WantBig: big.NewInt(1_200000000_000000003)},
		{StrInt: "1.2000000000000000030 ether", WantBig: big.NewInt(1_200000000_000000003)},
		{StrInt: "1.2000000000000000034 ether", WantBig: big.NewInt(1_200000000_000000003)},
		{StrInt: "1 gwei", WantBig: big.NewInt(1_000000000)},
		{StrInt: "1.2 gwei", WantBig: big.NewInt(1_200000000)},
		{StrInt: "1.200000003 gwei", WantBig: big.NewInt(1_200000003)},
		{StrInt: "1.2000000034 gwei", WantBig: big.NewInt(1_200000003)},
		{StrInt: ".", WantPanic: `invalid str big ".": number has no digits`},
		{StrInt: ". ether", WantPanic: `invalid str big ". ether": number has no digits`},
		{StrInt: "1.", WantBig: big.NewInt(1)},
		{StrInt: "1. ether", WantBig: big.NewInt(1_000000000_000000000)},
		{StrInt: ".1", WantPanic: `invalid str big ".1": must be an integer`},
		{StrInt: ".1 ether", WantBig: big.NewInt(100000000_000000000)},
		{StrInt: "0.1 ether", WantBig: big.NewInt(100000000_000000000)},
		{StrInt: "0.10 ether", WantBig: big.NewInt(100000000_000000000)},
		{StrInt: "00.10 ether", WantBig: big.NewInt(100000000_000000000)},
		{StrInt: " 1 ether", WantPanic: `invalid str big " 1 ether": EOF`},
		{StrInt: "1 ether ", WantPanic: `invalid str big "1 ether ": unknown unit "ether "`},
		{StrInt: "1  ether", WantPanic: `invalid str big "1  ether": unknown unit " ether"`},
		{StrInt: "-1", WantBig: big.NewInt(-1)},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			defer func() {
				var gotPanic string
				if r := recover(); r != nil {
					gotPanic = r.(string)
				}
				if test.WantPanic != gotPanic {
					t.Fatalf("want %q, got %q", test.WantPanic, gotPanic)
				}
			}()

			gotBig := I(test.StrInt)
			if test.WantPanic == "" && test.WantBig.Cmp(gotBig) != 0 {
				t.Fatalf("want %v, got %v", test.WantBig, gotBig)
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

func TestFromWei(t *testing.T) {
	t.Parallel()

	tests := []struct {
		Wei      *big.Int
		Decimals uint8
		Want     string
	}{
		{big.NewInt(0), 0, "0"},
		{big.NewInt(1), 0, "1"},
		{big.NewInt(0), 18, "0"},
		{big.NewInt(1), 18, "0.000000000000000001"},
		{big.NewInt(1000), 18, "0.000000000000001"},
		{big.NewInt(1000000), 18, "0.000000000001"},
		{big.NewInt(1000000000), 18, "0.000000001"},
		{big.NewInt(1000000000000), 18, "0.000001"},
		{big.NewInt(1000000000000000), 18, "0.001"},
		{big.NewInt(1000000000000000000), 18, "1"},
		{big.NewInt(-1), 18, "-0.000000000000000001"},
		{big.NewInt(-1000), 18, "-0.000000000000001"},
		{big.NewInt(-1000000), 18, "-0.000000000001"},
		{big.NewInt(-1000000000), 18, "-0.000000001"},
		{big.NewInt(-1000000000000), 18, "-0.000001"},
		{big.NewInt(-1000000000000000), 18, "-0.001"},
		{big.NewInt(-1000000000000000000), 18, "-1"},
		{big.NewInt(1000000000000000000), 15, "1000"},
		{big.NewInt(1000000000000000000), 12, "1000000"},
		{big.NewInt(1000000000000000000), 9, "1000000000"},
		{big.NewInt(1000000000000000000), 6, "1000000000000"},
		{big.NewInt(1000000000000000000), 3, "1000000000000000"},
		{big.NewInt(1000000000000000000), 0, "1000000000000000000"},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			got := FromWei(test.Wei, test.Decimals)
			if got != test.Want {
				t.Fatalf("%q != %q", got, test.Want)
			}
		})
	}
}
