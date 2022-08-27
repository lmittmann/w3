package eth_test

import (
	"math/big"
	"testing"

	"github.com/lmittmann/w3"
	"github.com/lmittmann/w3/module/eth"
)

func TestGasPrice(t *testing.T) {
	tests := []testCase[big.Int]{
		{
			Golden:  "gas_price",
			Call:    eth.GasPrice(),
			WantRet: *w3.I("0xc0fe"),
		},
	}

	runTestCases(t, tests)
}
