package eth_test

import (
	"math/big"
	"testing"

	"github.com/lmittmann/w3"
	"github.com/lmittmann/w3/module/eth"
	"github.com/lmittmann/w3/rpctest"
)

func TestGasPrice(t *testing.T) {
	rpctest.RunTestCases(t, []rpctest.TestCase[*big.Int]{
		{
			Golden:  "gas_price",
			Call:    eth.GasPrice(),
			WantRet: w3.I("0xc0fe"),
		},
	})
}
