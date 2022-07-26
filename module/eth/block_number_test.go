package eth_test

import (
	"math/big"
	"testing"

	"github.com/lmittmann/w3"
	"github.com/lmittmann/w3/module/eth"
	"github.com/lmittmann/w3/rpctest"
)

func TestBlockNumber(t *testing.T) {
	t.Parallel()

	srv := rpctest.NewFileServer(t, "testdata/block_number.golden")
	defer srv.Close()

	client := w3.MustDial(srv.URL())
	defer client.Close()

	var (
		blockNumber     big.Int
		wantBlockNumber = w3.I("0xc0fe")
	)
	if err := client.Call(
		eth.BlockNumber().Returns(&blockNumber),
	); err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	if wantBlockNumber.Cmp(&blockNumber) != 0 {
		t.Fatalf("want %v, got %v", wantBlockNumber, blockNumber)
	}
}
