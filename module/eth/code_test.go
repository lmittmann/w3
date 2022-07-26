package eth_test

import (
	"testing"

	"github.com/lmittmann/w3"
	"github.com/lmittmann/w3/module/eth"
)

func TestCode(t *testing.T) {
	t.Parallel()

	tests := []testCase[[]byte]{
		{
			Golden:  "get_code",
			Call:    eth.Code(w3.A("0x000000000000000000000000000000000000c0DE"), nil),
			WantRet: w3.B("0xdeadbeef"),
		},
	}

	for _, test := range tests {
		t.Run(test.Golden, runTestCase(t, test))
	}
}
