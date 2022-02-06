package eth_test

import (
	"math/big"
	"testing"

	"github.com/lmittmann/w3"
	"github.com/lmittmann/w3/module/eth"
	"github.com/lmittmann/w3/rpctest"
)

func TestGasPrice(t *testing.T) {
	t.Parallel()

	srv := rpctest.NewFileServer(t, "testdata/gas_price.golden")
	defer srv.Close()

	client := w3.MustDial(srv.URL())
	defer client.Close()

	var (
		gasPrice     = new(big.Int)
		wantGasPrice = w3.I("0xc0fe")
	)
	if err := client.Call(eth.GasPrice().Returns(gasPrice)); err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	if wantGasPrice.Cmp(gasPrice) != 0 {
		t.Fatalf("want %v, got %v", wantGasPrice, gasPrice)
	}
}
