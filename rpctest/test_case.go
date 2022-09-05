package rpctest

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/lmittmann/w3"
	"github.com/lmittmann/w3/internal"
	"github.com/lmittmann/w3/w3types"
)

type TestCase[T any] struct {
	Golden  string                   // File name in local "testdata/" directory without ".golden" extension
	Call    w3types.CallerFactory[T] // Call to test
	GotRet  T                        // Actual return value of the call
	WantRet T                        // Wanted return value of the call
	WantErr error                    // Wanted error of the call
}

func RunTestCases[T any](t *testing.T, tests []TestCase[T], opts ...cmp.Option) {
	t.Helper()

	for _, test := range tests {
		t.Run(test.Golden, func(t *testing.T) {
			t.Helper()

			srv := NewFileServer(t, fmt.Sprintf("testdata/%s.golden", test.Golden))
			defer srv.Close()

			client := w3.MustDial(srv.URL())
			defer client.Close()

			gotErr := client.Call(test.Call.Returns(&test.GotRet))
			comp(t, test.WantRet, test.GotRet, test.WantErr, gotErr, opts...)
		})
	}
}

func comp[T any](t *testing.T, wantVal, gotVal T, wantErr, gotErr error, opts ...cmp.Option) {
	t.Helper()

	// compare errors
	if diff := cmp.Diff(wantErr, gotErr,
		internal.EquateErrors(),
	); diff != "" {
		t.Fatalf("Err: (-want, +got)\n%s", diff)
	}

	// compare values
	opts = append(opts,
		cmp.AllowUnexported(big.Int{}),
		cmpopts.EquateEmpty(),
	)
	if diff := cmp.Diff(wantVal, gotVal, opts...); diff != "" {
		t.Fatalf("(-want, +got)\n%s", diff)
	}
}
