package eth_test

import (
	"testing"

	"github.com/lmittmann/w3"
	"github.com/lmittmann/w3/module/eth"
	"github.com/lmittmann/w3/rpctest"
)

func TestCode(t *testing.T) {
	tests := []rpctest.TestCase[[]byte]{
		{
			Golden:  "get_code",
			Call:    eth.Code(w3.A("0x000000000000000000000000000000000000c0DE"), nil),
			WantRet: w3.B("0xdeadbeef"),
		},
	}

	rpctest.RunTestCases(t, tests)
}
