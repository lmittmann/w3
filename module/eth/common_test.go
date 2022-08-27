package eth_test

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/lmittmann/w3"
	"github.com/lmittmann/w3/core"
	"github.com/lmittmann/w3/internal"
	"github.com/lmittmann/w3/rpctest"
)

type testCase[T any] struct {
	Golden  string
	Call    core.CallerFactory[T]
	GotRet  T
	WantRet T
	WantErr error
}

func comp[T any](t *testing.T, wantVal, gotVal T, wantErr, gotErr error, opts ...cmp.Option) {
	t.Helper()

	opts = append(opts,
		cmp.AllowUnexported(big.Int{}),
		cmpopts.EquateEmpty(),
	)
	if diff := cmp.Diff(wantVal, gotVal, opts...); diff != "" {
		t.Fatalf("(-want, +got)\n%s", diff)
	}

	if diff := cmp.Diff(wantErr, gotErr,
		internal.EquateErrors(),
	); diff != "" {
		t.Fatalf("Err: (-want, +got)\n%s", diff)
	}
}

func runTestCases[T any](t *testing.T, tests []testCase[T], opts ...cmp.Option) {
	t.Helper()

	for _, test := range tests {
		t.Run(test.Golden, func(t *testing.T) {
			t.Helper()

			srv := rpctest.NewFileServer(t, fmt.Sprintf("testdata/%s.golden", test.Golden))
			defer srv.Close()

			client := w3.MustDial(srv.URL())
			defer client.Close()

			gotErr := client.Call(test.Call.Returns(&test.GotRet))
			comp(t, test.WantRet, test.GotRet, test.WantErr, gotErr, opts...)
		})
	}
}
