package eth_test

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/lmittmann/w3"
	"github.com/lmittmann/w3/core"
	"github.com/lmittmann/w3/rpctest"
)

type testCase[T any] struct {
	Golden  string
	Call    core.CallerFactory[T]
	GotRet  T
	WantRet T
	WantErr error
}

func comp[T any](t *testing.T, wantVal, gotVal T, wantErr, gotErr error) {
	t.Helper()

	if diff := cmp.Diff(wantVal, gotVal,
		cmp.AllowUnexported(big.Int{}),
		cmpopts.EquateEmpty(),
	); diff != "" {
		t.Fatalf("(-want, +got)\n%s", diff)
	}

	if diff := cmp.Diff(wantErr, gotErr); diff != "" {
		t.Fatalf("Err: (-want, +got)\n%s", diff)
	}
}

func runTestCase[T any](t *testing.T, tt testCase[T]) func(t *testing.T) {
	return func(t *testing.T) {
		t.Helper()

		srv := rpctest.NewFileServer(t, fmt.Sprintf("testdata/%s.golden", tt.Golden))
		defer srv.Close()

		client := w3.MustDial(srv.URL())
		defer client.Close()

		gotErr := client.Call(tt.Call.Returns(&tt.GotRet))
		comp(t, tt.WantRet, tt.GotRet, tt.WantErr, gotErr)
	}
}
