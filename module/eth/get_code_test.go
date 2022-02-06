package eth_test

import (
	"bytes"
	"testing"

	"github.com/lmittmann/w3"
	"github.com/lmittmann/w3/module/eth"
	"github.com/lmittmann/w3/rpctest"
)

func TestCode(t *testing.T) {
	t.Parallel()

	srv := rpctest.NewFileServer(t, "testdata/get_code.golden")
	defer srv.Close()

	client := w3.MustDial(srv.URL())
	defer client.Close()

	var (
		code     []byte
		wantCode = w3.B("0xdeadbeef")
	)
	if err := client.Call(eth.Code(w3.A("0x000000000000000000000000000000000000c0DE")).Returns(&code)); err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	if !bytes.Equal(wantCode, code) {
		t.Fatalf("want 0x%x, got 0x%x", wantCode, code)
	}
}
