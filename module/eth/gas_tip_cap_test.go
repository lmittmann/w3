package eth_test

import (
	"math/big"
	"testing"

	"github.com/lmittmann/w3"
	"github.com/lmittmann/w3/module/eth"
	"github.com/lmittmann/w3/rpctest"
)

func TesGasTipCap(t *testing.T) {
	tests := []rpctest.TestCase[big.Int]{
		{
			Golden:  "eth_maxPriorityFeePerGas",
			Call:    eth.GasTipCap(),
			WantRet: *w3.I("0xc0fe"),
		},
	}

	rpctest.RunTestCases(t, tests)
}
