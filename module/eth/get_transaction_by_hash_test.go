package eth_test

import (
	"fmt"
	"math/big"
	"sync/atomic"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/lmittmann/w3"
	"github.com/lmittmann/w3/module/eth"
	"github.com/lmittmann/w3/rpctest"
)

func TestTransactionByHash_Type0(t *testing.T) {
	t.Parallel()

	srv := rpctest.NewFileServer(t, "testdata/get_transaction_by_hash__type0.golden")
	defer srv.Close()

	client := w3.MustDial(srv.URL())
	defer client.Close()

	var (
		tx     = new(types.Transaction)
		wantTx = types.NewTx(&types.LegacyTx{
			Nonce:    3292,
			GasPrice: w3.I("1559 gwei"),
			Gas:      21000,
			To:       w3.APtr("0x46499275b5c4d67dfa46B92D89aADA3158ea392e"),
			V:        w3.I("0x26"),
			R:        w3.I("0xcfaab0b753c1d71f695029e5b5da2f2f619370f5f224a42e1c19dcdcb9e814da"),
			S:        w3.I("0x606961e8b1dce9439df856ef1d1243f81f45938bac647568253260473efe7cc1"),
		})
	)

	if err := client.Call(eth.TransactionByHash(w3.H("0x2ecd08e86079f08cfc27c326aa01b1c8d62f288d5961118056bac7da315f94d9")).Returns(tx)); err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	if diff := cmp.Diff(wantTx, tx,
		cmp.AllowUnexported(types.Transaction{}, big.Int{}, atomic.Value{}),
		cmpopts.IgnoreFields(types.Transaction{}, "time"),
		cmpopts.EquateEmpty()); diff != "" {
		t.Fatalf("(-want, +got)\n%s", diff)
	}
}

func TestTransactionByHash_Type2(t *testing.T) {
	t.Parallel()

	srv := rpctest.NewFileServer(t, "testdata/get_transaction_by_hash__type2.golden")
	defer srv.Close()

	client := w3.MustDial(srv.URL())
	defer client.Close()

	var (
		tx     = new(types.Transaction)
		wantTx = types.NewTx(&types.DynamicFeeTx{
			ChainID:   big.NewInt(1),
			Nonce:     261,
			GasTipCap: w3.I("30.38 gwei"),
			GasFeeCap: w3.I("32.38 gwei"),
			Gas:       47238,
			To:        w3.APtr("0x491D6b7D6822d5d4BC88a1264E1b47791Fd8E904"),
			Data:      w3.B("0x095ea7b30000000000000000000000007645eec8bb51862a5aa855c40971b2877dae81afffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"),
			V:         w3.I("0x1"),
			R:         w3.I("0x416470241b7db89c67526881b6fd8e145416b294a35bf4280d3079f6308c2d11"),
			S:         w3.I("0x2c0af1cc55c22c0bab79ec083801da63253453156356fcd4291f50d0f425a0ee"),
		})
	)

	if err := client.Call(eth.TransactionByHash(w3.H("0xed382cb554ad10e94921d263a56c670669d6c380bbdacdbf96fed625b7132a1d")).Returns(tx)); err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	if diff := cmp.Diff(wantTx, tx,
		cmp.AllowUnexported(types.Transaction{}, big.Int{}, atomic.Value{}),
		cmpopts.IgnoreFields(types.Transaction{}, "time"),
		cmpopts.EquateEmpty()); diff != "" {
		t.Fatalf("(-want, +got)\n%s", diff)
	}
}

func TestTransactionByHash_0x00(t *testing.T) {
	t.Parallel()

	srv := rpctest.NewFileServer(t, "testdata/get_transaction_by_hash__0x00.golden")
	defer srv.Close()

	client := w3.MustDial(srv.URL())
	defer client.Close()

	var (
		tx      = new(types.Transaction)
		wantErr = fmt.Errorf("w3: response handling failed: not found")
	)

	if gotErr := client.Call(eth.TransactionByHash(common.Hash{}).Returns(tx)); wantErr.Error() != gotErr.Error() {
		t.Fatalf("want %v, got %v", wantErr, gotErr)
	}
}
