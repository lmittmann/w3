package eth_test

import (
	"testing"

	"github.com/lmittmann/w3/module/eth"
)

func TestChainID(t *testing.T) {
	tests := []testCase[uint64]{
		{
			Golden:  "chain_id",
			Call:    eth.ChainID(),
			WantRet: 1,
		},
	}

	runTestCases(t, tests)
}
