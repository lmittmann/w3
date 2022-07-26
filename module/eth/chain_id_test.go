package eth_test

import (
	"testing"

	"github.com/lmittmann/w3/module/eth"
)

func TestChainID(t *testing.T) {
	t.Parallel()

	tests := []testCase[uint64]{
		{
			Golden:  "chain_id",
			Call:    eth.ChainID(),
			WantRet: 1,
		},
	}

	for _, test := range tests {
		t.Run(test.Golden, runTestCase(t, test))
	}
}
