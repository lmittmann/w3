package eth_test

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/lmittmann/w3"
	"github.com/lmittmann/w3/internal/rpctest"
	"github.com/lmittmann/w3/module/eth"
)

func TestStorageAt(t *testing.T) {
	t.Parallel()

	srv := rpctest.NewFileServer(t, "testdata/get_storage_at.golden")
	defer srv.Close()

	client := w3.MustDial(srv.URL())
	defer client.Close()

	var (
		storage     common.Hash
		wantStorage = w3.H("0x0000000000000000000000000000000000000000000000000000000000000042")
	)

	if err := client.Call(eth.StorageAt(w3.A("0x000000000000000000000000000000000000c0DE"), w3.H("0x0000000000000000000000000000000000000000000000000000000000000001")).Returns(&storage)); err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	if wantStorage != storage {
		t.Fatalf("want %v, got %v", wantStorage, storage)
	}
}
