package hexutil_test

import (
	"strconv"
	"testing"

	"github.com/holiman/uint256"
	"github.com/lmittmann/w3/internal/hexutil"
)

var u256Tests = []struct {
	Raw     string
	Val     *hexutil.U256
	WantEnc string
}{
	{"0x0", new(hexutil.U256), "0x0"},
	{"0x00", new(hexutil.U256), "0x0"},
	{"0xc0fe", (*hexutil.U256)(uint256.NewInt(0xc0fe)), "0xc0fe"},
	{"0x000000000000000000000000000000000000c0fe", (*hexutil.U256)(uint256.NewInt(0xc0fe)), "0xc0fe"},
}

func TestU256UnmarshalText(t *testing.T) {
	for i, test := range u256Tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			got := new(hexutil.U256)
			if err := got.UnmarshalText([]byte(test.Raw)); err != nil {
				t.Fatal(err)
			}

			if want := (*uint256.Int)(test.Val); want.Cmp((*uint256.Int)(got)) != 0 {
				t.Fatalf("want %v, got %v", want, (*uint256.Int)(got))
			}
		})
	}
}

func TestU256MarshalText(t *testing.T) {
	for i, test := range u256Tests {
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
