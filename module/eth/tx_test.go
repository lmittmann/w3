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

var type2Tx = types.NewTx(&types.DynamicFeeTx{
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

func TestTx(t *testing.T) {
	rpctest.RunTestCases(t,
		[]rpctest.TestCase[*types.Transaction]{
			{
				Golden: "get_transaction_by_hash__type0",
				Call:   eth.Tx(w3.H("0x2ecd08e86079f08cfc27c326aa01b1c8d62f288d5961118056bac7da315f94d9")),
				WantRet: types.NewTx(&types.LegacyTx{
					Nonce:    3292,
					GasPrice: w3.I("1559 gwei"),
					Gas:      21000,
					To:       w3.APtr("0x46499275b5c4d67dfa46B92D89aADA3158ea392e"),
					V:        w3.I("0x26"),
					R:        w3.I("0xcfaab0b753c1d71f695029e5b5da2f2f619370f5f224a42e1c19dcdcb9e814da"),
					S:        w3.I("0x606961e8b1dce9439df856ef1d1243f81f45938bac647568253260473efe7cc1"),
				}),
			},
			{
				Golden:  "get_transaction_by_hash__type2",
				Call:    eth.Tx(w3.H("0xed382cb554ad10e94921d263a56c670669d6c380bbdacdbf96fed625b7132a1d")),
				WantRet: type2Tx,
			},
			{
				Golden:  "get_transaction_by_hash__0x00",
				Call:    eth.Tx(common.Hash{}),
				WantErr: fmt.Errorf("w3: call failed: not found"),
			},
		},
		cmp.AllowUnexported(types.Transaction{}, atomic.Value{}),
		cmpopts.IgnoreFields(types.Transaction{}, "time"),
	)
}

func TestTxByBlockNumberAndIndex(t *testing.T) {
	rpctest.RunTestCases(t,
		[]rpctest.TestCase[*types.Transaction]{
			{
				Golden:  "get_transaction_by_block_number_and_index",
				Call:    eth.TxByBlockNumberAndIndex(big.NewInt(12965001), 98),
				WantRet: type2Tx,
			},
		},
		cmp.AllowUnexported(types.Transaction{}, atomic.Value{}),
		cmpopts.IgnoreFields(types.Transaction{}, "time"),
	)
}

func TestSendTx(t *testing.T) {
	rpctest.RunTestCases(t, []rpctest.TestCase[common.Hash]{
		{
			Golden:  "send_raw_transaction",
			Call:    eth.SendTx(type2Tx),
			WantRet: w3.H("0xed382cb554ad10e94921d263a56c670669d6c380bbdacdbf96fed625b7132a1d"),
		},
	})
}

func TestTxReceipt(t *testing.T) {
	rpctest.RunTestCases(t, []rpctest.TestCase[*types.Receipt]{
		{
			Golden: "get_transaction_receipt",
			Call:   eth.TxReceipt(w3.H("0xed382cb554ad10e94921d263a56c670669d6c380bbdacdbf96fed625b7132a1d")),
			WantRet: &types.Receipt{
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
				TxHash:            w3.H("0xed382cb554ad10e94921d263a56c670669d6c380bbdacdbf96fed625b7132a1d"),
				GasUsed:           47238,
				EffectiveGasPrice: w3.I("31504967822"),
				BlockHash:         w3.H("0xa32d159805750cbe428b799a49b85dcb2300f61d806786f317260e721727d162"),
				BlockNumber:       big.NewInt(12965001),
				TransactionIndex:  98,
			},
		},
		{
			Golden:  "get_transaction_receipt_0x00",
			Call:    eth.TxReceipt(common.Hash{}),
			WantErr: fmt.Errorf("w3: call failed: not found"),
		},
	})
}

func TestBlockReceipts(t *testing.T) {
	rpctest.RunTestCases(t, []rpctest.TestCase[types.Receipts]{
		{
			Golden: "get_block_receipts",
			Call:   eth.BlockReceipts(big.NewInt(0xc0fe)),
			WantRet: types.Receipts{
				{
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
					TxHash:            w3.H("0xed382cb554ad10e94921d263a56c670669d6c380bbdacdbf96fed625b7132a1d"),
					GasUsed:           47238,
					EffectiveGasPrice: w3.I("31504967822"),
					BlockHash:         w3.H("0xa32d159805750cbe428b799a49b85dcb2300f61d806786f317260e721727d162"),
					BlockNumber:       big.NewInt(12965001),
					TransactionIndex:  98,
				},
			},
		},
	})
}

func TestNonce(t *testing.T) {
	rpctest.RunTestCases(t, []rpctest.TestCase[uint64]{
		{
			Golden:  "get_transaction_count",
			Call:    eth.Nonce(w3.A("0x000000000000000000000000000000000000c0Fe"), nil),
			WantRet: 1,
		},
	})
}
