package w3_test

import (
	"bytes"
	"fmt"
	"math/big"
	"strconv"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/lmittmann/w3"
)

func ExampleI() {
	fmt.Printf("%v wei\n", w3.I("0x2b98d99b09e3c000"))
	fmt.Printf("%v wei\n", w3.I("3141500000000000000"))
	fmt.Printf("%v wei\n", w3.I("3.1415 ether"))
	fmt.Printf("%v wei\n", w3.I("31.415 gwei"))
	// Output:
	// 3141500000000000000 wei
	// 3141500000000000000 wei
	// 3141500000000000000 wei
	// 31415000000 wei
}

func ExampleFromWei() {
	wei := big.NewInt(1_230000000_000000000)
	fmt.Printf("%s Ether\n", w3.FromWei(wei, 18))
	// Output:
	// 1.23 Ether
}

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
		{
			HexAddress:  "000000000000000000000000000000000000c0Fe",
			WantAddress: common.HexToAddress("0x000000000000000000000000000000000000c0Fe"),
		},
		{HexAddress: "0xcofe", WantPanic: `invalid address "cofe": encoding/hex: invalid byte: U+006F 'o'`},
		{HexAddress: "0xc0Fe", WantPanic: `invalid address "c0Fe": must have 20 bytes`},
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

			gotAddr := w3.A(test.HexAddress)
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
		{HexBytes: "0xc0fe", WantBytes: []byte{0xc0, 0xfe}},
		{HexBytes: "c0fe", WantBytes: []byte{0xc0, 0xfe}},
		{HexBytes: "0xcofe", WantPanic: `invalid bytes "cofe": encoding/hex: invalid byte: U+006F 'o'`},
		{HexBytes: "0xc0f", WantPanic: `invalid bytes "c0f": encoding/hex: odd length hex string`},
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

			gotBytes := w3.B(test.HexBytes)
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
		{
			HexHash:  "000000000000000000000000000000000000000000000000000000000000c0fe",
			WantHash: common.HexToHash("0x000000000000000000000000000000000000000000000000000000000000c0fe"),
		},
		{HexHash: "0xcofe", WantPanic: `invalid hash "cofe": encoding/hex: invalid byte: U+006F 'o'`},
		{HexHash: "0xc0fe", WantPanic: `invalid hash "c0fe": must have 32 bytes`},
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

			gotHash := w3.H(test.HexHash)
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
		{StrInt: "0xO", WantPanic: `invalid hex big "0xO"`},

		// int big's
		{StrInt: "0", WantBig: big.NewInt(0)},
		{StrInt: "1", WantBig: big.NewInt(1)},
		{StrInt: "255", WantBig: big.NewInt(255)},
		{StrInt: "X", WantPanic: `str big "X" must be number`},
		{StrInt: "888916043834286157872", WantBig: strBig("888916043834286157872")},

		// unit big's
		{StrInt: "0 ether", WantBig: big.NewInt(0)},
		{StrInt: "0 eth", WantBig: big.NewInt(0)},
		{StrInt: "0 gwei", WantBig: big.NewInt(0)},
		{StrInt: "1 ether", WantBig: big.NewInt(1_000000000_000000000)},
		{StrInt: "1 eth", WantBig: big.NewInt(1_000000000_000000000)},
		{StrInt: "1ether", WantPanic: `str big "1ether" must be number`},
		{StrInt: "1.2 ether", WantBig: big.NewInt(1_200000000_000000000)},
		{StrInt: "01.2 ether", WantBig: big.NewInt(1_200000000_000000000)},
		{StrInt: "1.20 ether", WantBig: big.NewInt(1_200000000_000000000)},
		{StrInt: "1.200000000000000003 ether", WantBig: big.NewInt(1_200000000_000000003)},
		{StrInt: "1.2000000000000000034 ether", WantPanic: `str big "1.2000000000000000034 ether" exceeds precision`},
		{StrInt: "1 gwei", WantBig: big.NewInt(1_000000000)},
		{StrInt: "1.2 gwei", WantBig: big.NewInt(1_200000000)},
		{StrInt: "1.200000003 gwei", WantBig: big.NewInt(1_200000003)},
		{StrInt: "1.2000000034 gwei", WantPanic: `str big "1.2000000034 gwei" exceeds precision`},
		{StrInt: ".", WantPanic: `str big "." must be number`},
		{StrInt: ". ether", WantPanic: `str big ". ether" must be number`},
		{StrInt: "1.", WantPanic: `str big "1." without unit must be integer`},
		{StrInt: "1. ether", WantPanic: `str big "1. ether" must be number`},
		{StrInt: ".1", WantPanic: `str big ".1" must be number`},
		{StrInt: ".1 ether", WantPanic: `str big ".1 ether" must be number`},
		{StrInt: "0.1 ether", WantBig: big.NewInt(100000000_000000000)},
		{StrInt: "0.10 ether", WantBig: big.NewInt(100000000_000000000)},
		{StrInt: "00.10 ether", WantBig: big.NewInt(100000000_000000000)},
		{StrInt: " 1 ether", WantPanic: `str big " 1 ether" must be number`},
		{StrInt: "1 ether ", WantPanic: `str big "1 ether " has invalid unit "ether "`},
		{StrInt: "1  ether", WantPanic: `str big "1  ether" has invalid unit " ether"`},
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

			gotBig := w3.I(test.StrInt)
			if test.WantPanic == "" && test.WantBig.Cmp(gotBig) != 0 {
				t.Fatalf("want %v, got %v", test.WantBig, gotBig)
			}
		})
	}
}

func strBig(s string) *big.Int {
	big, _ := new(big.Int).SetString(s, 10)
	return big
}

func FuzzI(f *testing.F) {
	f.Add([]byte{0x2a})
	f.Fuzz(func(t *testing.T, b []byte) {
		wantBig := new(big.Int).SetBytes(b)
		bigStr := wantBig.String()
		if gotBig := w3.I(bigStr); wantBig.Cmp(gotBig) != 0 {
			t.Fatalf("want %v, got %v", wantBig, gotBig)
		}

		bigHexstr := wantBig.Text(16)
		if gotBig := w3.I("0x" + bigHexstr); wantBig.Cmp(gotBig) != 0 {
			t.Fatalf("want %v, got %v", wantBig, gotBig)
		}
	})
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
				w3.I(bench)
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
		{nil, 0, "<nil>"},
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
			got := w3.FromWei(test.Wei, test.Decimals)
			if got != test.Want {
				t.Fatalf("%q != %q", got, test.Want)
			}
		})
	}
}
