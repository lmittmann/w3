package web3_test

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/lmittmann/w3"
	"github.com/lmittmann/w3/internal"
	"github.com/lmittmann/w3/module/web3"
	"github.com/lmittmann/w3/rpctest"
)

func TestClientVersion(t *testing.T) {
	t.Parallel()

	srv := rpctest.NewFileServer(t, "testdata/client_version.golden")
	defer srv.Close()

	client := w3.MustDial(srv.URL())
	defer client.Close()

	var (
		clientVersion     string
		wantClientVersion = "Geth"
	)
	if err := client.Call(
		web3.ClientVersion().Returns(&clientVersion),
	); err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	if wantClientVersion != clientVersion {
		t.Fatalf("want %q, got %q", wantClientVersion, clientVersion)
	}
}

func TestClientVersion__Err(t *testing.T) {
	t.Parallel()

	srv := rpctest.NewFileServer(t, "testdata/client_version__err.golden")
	defer srv.Close()

	client := w3.MustDial(srv.URL())
	defer client.Close()

	var (
		clientVersion string
		wantErr       = errors.New("w3: response handling failed: the method web3_clientVersion does not exist/is not available")
	)
	err := client.Call(
		web3.ClientVersion().Returns(&clientVersion),
	)
	if diff := cmp.Diff(wantErr, err,
		internal.EquateErrors(),
	); diff != "" {
		t.Fatalf("(-want, +got)\n%s", diff)
	}
}
