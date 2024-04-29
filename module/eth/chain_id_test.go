package eth_test

import (
	"testing"

	"github.com/lmittmann/w3/module/eth"
	"github.com/lmittmann/w3/rpctest"
)

func TestChainID(t *testing.T) {
	tests := []rpctest.TestCase[uint64]{
		{
			Golden:  "chain_id",
			Call:    eth.ChainID(),
			WantRet: ptr[uint64](1),
		},
	}

	rpctest.RunTestCases(t, tests)
}
