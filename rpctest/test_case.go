package rpctest

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/lmittmann/w3"
	"github.com/lmittmann/w3/internal"
	"github.com/lmittmann/w3/w3types"
)

type TestCase[T any] struct {
	Golden  string                      // File name in local "testdata/" directory without ".golden" extension
	Call    w3types.RPCCallerFactory[T] // Call to test
	WantRet T                           // Wanted return value of the call
	WantErr error                       // Wanted error of the call
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

			var gotRet T
			gotErr := client.Call(test.Call.Returns(&gotRet))
			comp(t, test.WantRet, gotRet, test.WantErr, gotErr, opts...)
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
	if wantErr != nil {
		return
	}

	// compare values
	opts = append(opts,
		cmp.AllowUnexported(big.Int{}, types.Transaction{}, types.Block{}),
		cmpopts.IgnoreFields(types.Transaction{}, "time", "hash", "size", "from"),
		cmpopts.IgnoreFields(types.Block{}, "hash", "size"),
		cmpopts.EquateEmpty(),
	)
	if diff := cmp.Diff(wantVal, gotVal, opts...); diff != "" {
		t.Fatalf("(-want, +got)\n%s", diff)
	}
}
