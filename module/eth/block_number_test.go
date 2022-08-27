package eth_test

import (
	"math/big"
	"testing"

	"github.com/lmittmann/w3"
	"github.com/lmittmann/w3/module/eth"
)

func TestBlockNumber(t *testing.T) {
	tests := []testCase[big.Int]{
		{
			Golden:  "block_number",
			Call:    eth.BlockNumber(),
			WantRet: *w3.I("0xc0fe"),
		},
	}

	runTestCases(t, tests)
}
