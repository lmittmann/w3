package hexutil_test

import (
	"bytes"
	"strconv"
	"testing"

	"github.com/lmittmann/w3/internal/hexutil"
)

var bytesTests = []struct {
	Raw     string
	Val     hexutil.Bytes
	WantEnc string
}{
	{"0xc0fe", hexutil.Bytes{0xc0, 0xfe}, "0xc0fe"},
	{"c0fe", hexutil.Bytes{0xc0, 0xfe}, "0xc0fe"},
	{"0Xc0fe", hexutil.Bytes{0xc0, 0xfe}, "0xc0fe"},
}

func TestBytesUnmarshalText(t *testing.T) {
	for i, test := range bytesTests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var got hexutil.Bytes
			if err := got.UnmarshalText([]byte(test.Raw)); err != nil {
				t.Fatal(err)
			}
			if want := test.Val; !bytes.Equal(want, got) {
				t.Fatalf("want %x, got %x", want, got)
			}
		})
	}
}

func TestBytesMarshalText(t *testing.T) {
	for i, test := range bytesTests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			got, err := test.Val.MarshalText()
			if err != nil {
				t.Fatal(err)
			}
			if want := test.WantEnc; want != string(got) {
				t.Fatalf("want %s, got %s", test.WantEnc, string(got))
			}
		})
	}
}
