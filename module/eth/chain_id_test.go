package eth_test

import (
	"testing"

	"github.com/lmittmann/w3"
	"github.com/lmittmann/w3/internal/rpctest"
	"github.com/lmittmann/w3/module/eth"
)

func TestChainID(t *testing.T) {
	t.Parallel()

	srv := rpctest.NewFileServer(t, "testdata/chain_id.golden")
	defer srv.Close()

	client := w3.MustDial(srv.URL())
	defer client.Close()

	var (
		chainID     uint64
		wantChainID uint64 = 1
	)
	if err := client.Call(eth.ChainID().Returns(&chainID)); err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	if wantChainID != chainID {
		t.Fatalf("want %d, got %d", wantChainID, chainID)
	}
}
