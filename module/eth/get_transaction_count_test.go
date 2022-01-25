package eth_test

import (
	"testing"

	"github.com/lmittmann/w3"
	"github.com/lmittmann/w3/internal/rpctest"
	"github.com/lmittmann/w3/module/eth"
)

func TestNonce(t *testing.T) {
	t.Parallel()

	srv := rpctest.NewFileServer(t, "testdata/get_transaction_count.golden")
	defer srv.Close()

	client := w3.MustDial(srv.URL())
	defer client.Close()

	var (
		nonce     uint64
		wantNonce uint64 = 1
	)

	if err := client.Call(eth.Nonce(w3.A("0x000000000000000000000000000000000000c0Fe")).Returns(&nonce)); err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	if wantNonce != nonce {
		t.Fatalf("want %d, got %d", wantNonce, nonce)
	}
}
