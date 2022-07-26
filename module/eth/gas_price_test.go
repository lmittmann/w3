package eth_test

import (
	"math/big"
	"testing"

	"github.com/lmittmann/w3"
	"github.com/lmittmann/w3/module/eth"
)

func TestGasPrice(t *testing.T) {
	t.Parallel()

	tests := []testCase[big.Int]{
		{
			Golden:  "gas_price",
			Call:    eth.GasPrice(),
			WantRet: *w3.I("0xc0fe"),
		},
	}

	for _, test := range tests {
		t.Run(test.Golden, runTestCase(t, test))
	}
}
