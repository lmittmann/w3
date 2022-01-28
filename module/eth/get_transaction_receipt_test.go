package eth_test

import (
	"math/big"
	"sync/atomic"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/lmittmann/w3"
	"github.com/lmittmann/w3/internal/rpctest"
	"github.com/lmittmann/w3/module/eth"
)

func TestTransactionReceipt(t *testing.T) {
	t.Parallel()

	srv := rpctest.NewFileServer(t, "testdata/get_transaction_receipt.golden")
	defer srv.Close()

	client := w3.MustDial(srv.URL())
	defer client.Close()

	var (
		receipt     = new(types.Receipt)
		wantReceipt = &types.Receipt{
			Type:              2,
			Status:            types.ReceiptStatusSuccessful,
			CumulativeGasUsed: 8726063,
			Bloom:             types.BytesToBloom(w3.B("0x00000000000000000000000000000000000000000000000000000000000800000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000000000000000000000000000040000000010000000000000000000000006000000000000002000000000020000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000004000000000000000000000000010000000000000000000000000000000000000000000000000040000000000")),
			Logs: []*types.Log{
				{
					Address: w3.A("0x491D6b7D6822d5d4BC88a1264E1b47791Fd8E904"),
					Topics: []common.Hash{
						w3.H("0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925"),
						w3.H("0x0000000000000000000000002e419a06feb47d5f640636a55a814757fa10edf9"),
						w3.H("0x0000000000000000000000007645eec8bb51862a5aa855c40971b2877dae81af"),
					},
					Data:        w3.B("0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"),
					BlockNumber: 12965001,
					TxHash:      w3.H("0xed382cb554ad10e94921d263a56c670669d6c380bbdacdbf96fed625b7132a1d"),
					TxIndex:     98,
					BlockHash:   w3.H("0xa32d159805750cbe428b799a49b85dcb2300f61d806786f317260e721727d162"),
					Index:       187,
				},
			},

			TxHash:           w3.H("0xed382cb554ad10e94921d263a56c670669d6c380bbdacdbf96fed625b7132a1d"),
			GasUsed:          47238,
			BlockHash:        w3.H("0xa32d159805750cbe428b799a49b85dcb2300f61d806786f317260e721727d162"),
			BlockNumber:      big.NewInt(12965001),
			TransactionIndex: 98,
		}
	)

	if err := client.Call(eth.TransactionReceipt(w3.H("0xed382cb554ad10e94921d263a56c670669d6c380bbdacdbf96fed625b7132a1d")).Returns(receipt)); err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	if diff := cmp.Diff(wantReceipt, receipt,
		cmp.AllowUnexported(types.Transaction{}, big.Int{}, atomic.Value{}),
		cmpopts.IgnoreFields(types.Transaction{}, "time"),
		cmpopts.EquateEmpty()); diff != "" {
		t.Fatalf("(-want, +got)\n%s", diff)
	}
}
