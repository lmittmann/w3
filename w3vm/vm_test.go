package w3vm_test

import (
	_ "embed"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/lmittmann/w3"
	"github.com/lmittmann/w3/internal"
	"github.com/lmittmann/w3/module/eth"
	"github.com/lmittmann/w3/w3types"
	"github.com/lmittmann/w3/w3vm"
	"github.com/lmittmann/w3/w3vm/state"
)

var (
	addr0 common.Address
	addr1 = common.Address{0x1}

	client = w3.MustDial("https://eth.llamarpc.com")
)

func TestVMApply(t *testing.T) {
	tests := []struct {
		PreState    w3types.State
		Message     *w3types.Message
		WantReceipt *w3vm.Receipt
		WantErr     error
	}{
		{
			Message: &w3types.Message{
				From:  addr0,
				To:    &addr1,
				Gas:   21_000,
				Value: big.NewInt(1),
			},
			WantErr: errors.New("insufficient funds for gas * price + value: address 0x0000000000000000000000000000000000000000 have 0 want 1"),
		},
		{
			Message: &w3types.Message{
				From:      addr0,
				To:        &addr1,
				Gas:       21_000,
				GasFeeCap: big.NewInt(1),
				Value:     big.NewInt(1),
			},
			WantErr: errors.New("insufficient funds for gas * price + value: address 0x0000000000000000000000000000000000000000 have 0 want 21001"),
		},
		{
			PreState: w3types.State{
				addr0: {
					Balance: w3.I("1 ether"),
				},
			},
			Message: &w3types.Message{
				From:  addr0,
				To:    &addr1,
				Gas:   21_000,
				Value: w3.I("1 ether"),
			},
			WantReceipt: &w3vm.Receipt{
				GasUsed:  21_000,
				GasLimit: 21_000,
			},
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			vm := w3vm.New(
				w3vm.WithState(test.PreState),
			)
			gotReceipt, gotErr := vm.Apply(test.Message)
			if diff := cmp.Diff(test.WantErr, gotErr,
				internal.EquateErrors(),
			); diff != "" {
				t.Fatalf("(-want +got)\n%s", diff)
			}
			if diff := cmp.Diff(test.WantReceipt, gotReceipt,
				internal.EquateErrors(),
				cmpopts.IgnoreUnexported(w3vm.Receipt{}),
			); diff != "" {
				t.Fatalf("(-want +got)\n%s", diff)
			}
		})
	}
}

func TestVMApply_Integration(t *testing.T) {
	blocks := []*big.Int{
		big.NewInt(15_054_998),
		big.NewInt(15_054_999),
		big.NewInt(15_050_000), // Gray Glacier
		big.NewInt(15_050_001),

		big.NewInt(15_537_392),
		big.NewInt(15_537_393),
		big.NewInt(15_537_394), // Paris (The Merge)
		big.NewInt(15_537_395),

		big.NewInt(17_034_868),
		big.NewInt(17_034_869),
		big.NewInt(17_034_870), // Shanghai
		big.NewInt(17_034_871),
	}

	for _, number := range blocks {
		number := number
		t.Run(number.String(), func(t *testing.T) {
			t.Parallel()

			var block types.Block
			if err := client.Call(
				eth.BlockByNumber(number).Returns(&block),
			); err != nil {
				t.Fatalf("Failed to fetch block: %v", err)
			}

			f := state.NewTestingRPCFetcher(t, client, new(big.Int).Sub(number, w3.Big1))
			vm := w3vm.New(
				w3vm.WithFetcher(f),
				w3vm.WithHeader(block.Header()),
			)
			signer := types.MakeSigner(params.MainnetChainConfig, number, block.Time())
			receipts, err := fetchReceipts(block.Transactions())
			if err != nil {
				t.Fatalf("Failed to fetch receipts: %v", err)
			}

			for i, tx := range block.Transactions() {
				t.Run(fmt.Sprintf("%d_%s", i, tx.Hash()), func(t *testing.T) {
					wantReceipt := &w3vm.Receipt{
						GasUsed: receipts[i].GasUsed,
						Logs:    receipts[i].Logs,
					}
					if receipts[i].ContractAddress != addr0 {
						wantReceipt.ContractAddress = &receipts[i].ContractAddress
					}
					if receipts[i].Status == types.ReceiptStatusFailed {
						wantReceipt.Err = cmpopts.AnyError
					}

					gotReceipt, err := vm.Apply(new(w3types.Message).MustSetTx(tx, signer))
					if err != nil && gotReceipt == nil {
						t.Fatalf("Failed to apply tx: %v", err)
					}
					if diff := cmp.Diff(wantReceipt, gotReceipt,
						cmpopts.EquateEmpty(),
						cmpopts.EquateErrors(),
						cmpopts.IgnoreUnexported(w3vm.Receipt{}),
						cmpopts.IgnoreFields(w3vm.Receipt{}, "GasLimit", "Output"),
						cmpopts.IgnoreFields(types.Log{}, "BlockHash", "BlockNumber", "TxHash", "TxIndex", "Index"),
					); diff != "" {
						t.Fatalf("(-want +got)\n%s", diff)
					}
				})
			}
		})
	}
}

func fetchReceipts(txs []*types.Transaction) ([]*types.Receipt, error) {
	const batchSize = 100

	receipts := make([]*types.Receipt, len(txs))
	caller := make([]w3types.Caller, len(txs))
	for i, tx := range txs {
		receipts[i] = new(types.Receipt)
		caller[i] = eth.TxReceipt(tx.Hash()).Returns(receipts[i])
	}

	for i := 0; i < len(txs); i += batchSize {
		if err := client.Call(caller[i:min(i+batchSize, len(txs))]...); err != nil {
			return nil, err
		}
	}
	return receipts, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
