package hexutil_test

import (
	"math/big"
	"strconv"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/lmittmann/w3/internal/hexutil"
)

var hashTests = []struct {
	Raw     string
	Val     hexutil.Hash
	WantEnc string
}{
	{"0x0", hexutil.Hash{}, "0x0"},
	{"0x00", hexutil.Hash{}, "0x0"},
	{"0xc0fe", (hexutil.Hash)(common.BigToHash(big.NewInt(0xc0fe))), "0xc0fe"},
	{"0x000000000000000000000000000000000000000000000000000000000000c0fe", (hexutil.Hash)(common.BigToHash(big.NewInt(0xc0fe))), "0xc0fe"},
}

func TestHashUnmarshalText(t *testing.T) {
	for i, test := range hashTests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			got := hexutil.Hash{}
			if err := got.UnmarshalText([]byte(test.Raw)); err != nil {
				t.Fatal(err)
			}

			if want := (common.Hash)(test.Val); want.Cmp((common.Hash)(got)) != 0 {
				t.Fatalf("want %v, got %v", want, (common.Hash)(got))
			}
		})
	}
}

func TestHashMarshalText(t *testing.T) {
	for i, test := range hashTests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			got, err := test.Val.MarshalText()
			if err != nil {
				t.Fatal(err)
			}

			if want := test.WantEnc; string(got) != want {
				t.Fatalf("want %q, got %q", want, string(got))
			}
		})
	}
}
