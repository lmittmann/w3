package w3vm_test

import (
	_ "embed"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/rawdb"
	coreState "github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/lmittmann/w3"
	"github.com/lmittmann/w3/internal"
	"github.com/lmittmann/w3/module/eth"
	"github.com/lmittmann/w3/w3types"
	"github.com/lmittmann/w3/w3vm"
)

var (
	addr0    = common.Address{0x0}
	addr1    = common.Address{0x1}
	addrWETH = w3.A("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2")

	//go:embed testdata/weth9.bytecode
	hexCodeWETH string
	codeWETH    = w3.B(strings.TrimSpace(hexCodeWETH))

	funcBalanceOf = w3.MustNewFunc("balanceOf(address)", "uint256")
	funcTransfer  = w3.MustNewFunc("transfer(address,uint256)", "bool")

	client = w3.MustDial("https://rpc.ankr.com/eth")
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
		{ // WETH transfer
			PreState: w3types.State{
				addr0: {Balance: w3.I("1 ether")},
				addrWETH: {
					Code: codeWETH,
					Storage: map[common.Hash]common.Hash{
						w3vm.WETHBalanceSlot(addr0): common.BigToHash(w3.I("1 ether")),
					},
				},
			},
			Message: &w3types.Message{
				From:  addr0,
				To:    &addrWETH,
				Input: must(funcTransfer.EncodeArgs(addr1, w3.I("1 ether"))),
				Gas:   100_000,
			},
			WantReceipt: &w3vm.Receipt{
				GasUsed:   38_853,
				GasRefund: 9_713,
				GasLimit:  58_753,
				Logs: []*types.Log{
					{
						Address: addrWETH,
						Topics: []common.Hash{
							w3.H("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"),
							w3.H("0x0000000000000000000000000000000000000000000000000000000000000000"),
							w3.H("0x0000000000000000000000000100000000000000000000000000000000000000"),
						},
						Data: w3.B("0x0000000000000000000000000000000000000000000000000de0b6b3a7640000"),
					},
				},
				Output: w3.B("0x0000000000000000000000000000000000000000000000000000000000000001"),
			},
		},
		{ // WETH transfer with insufficient balance
			PreState: w3types.State{
				addr0: {Balance: w3.I("1 ether")},
				addrWETH: {
					Code: codeWETH,
					Storage: map[common.Hash]common.Hash{
						w3vm.WETHBalanceSlot(addr0): common.BigToHash(w3.I("1 ether")),
					},
				},
			},
			Message: &w3types.Message{
				From:  addr0,
				To:    &addrWETH,
				Input: must(funcTransfer.EncodeArgs(addr1, w3.I("10 ether"))),
				Gas:   100_000,
			},
			WantReceipt: &w3vm.Receipt{
				GasUsed:  24_019,
				GasLimit: 24_019,
				Err:      errors.New("execution reverted"),
			},
			WantErr: errors.New("execution reverted"),
		},
		{ // revert with output
			PreState: w3types.State{
				addr0: {Balance: w3.I("1 ether")},
				addr1: {Code: w3.B("0x60015ffd")}, // PUSH1 0x1, PUSH0, REVERT
			},
			Message: &w3types.Message{
				From: addr0,
				To:   &addr1,
			},
			WantReceipt: &w3vm.Receipt{
				GasUsed:  21_008,
				GasLimit: 21_008,
				Output:   w3.B("0x00"),
				Err:      errors.New("execution reverted"),
			},
			WantErr: errors.New("execution reverted"),
		},
		{ // deploy contract for account with nonce == 0
			Message: &w3types.Message{
				From:  addr1,
				Input: w3.B("0x00"),
			},
			WantReceipt: &w3vm.Receipt{
				GasUsed:         53_006,
				GasLimit:        53_006,
				ContractAddress: ptr(crypto.CreateAddress(addr1, 0)),
			},
		},
		{ // deploy contract for account with nonce > 0
			PreState: w3types.State{
				addr1: {Nonce: 1},
			},
			Message: &w3types.Message{
				From:  addr1,
				Input: w3.B("0x00"),
			},
			WantReceipt: &w3vm.Receipt{
				GasUsed:         53_006,
				GasLimit:        53_006,
				ContractAddress: ptr(crypto.CreateAddress(addr1, 1)),
			},
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			vm, _ := w3vm.New(
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
				cmpopts.EquateComparable(common.Address{}, common.Hash{}),
			); diff != "" {
				t.Fatalf("(-want +got)\n%s", diff)
			}
		})
	}
}

func TestVMCall(t *testing.T) {
	tests := []struct {
		PreState    w3types.State
		Message     *w3types.Message
		WantReceipt *w3vm.Receipt
		WantErr     error
	}{
		{
			PreState: w3types.State{
				addrWETH: {
					Code: codeWETH,
					Storage: map[common.Hash]common.Hash{
						w3vm.WETHBalanceSlot(addr0): common.BigToHash(w3.I("1 ether")),
					},
				},
			},
			Message: &w3types.Message{
				From:  addr0,
				To:    &addrWETH,
				Input: must(funcBalanceOf.EncodeArgs(addr0)),
			},
			WantReceipt: &w3vm.Receipt{
				GasUsed:  23_726,
				GasLimit: 23_726,
				Output:   w3.B("0x0000000000000000000000000000000000000000000000000de0b6b3a7640000"),
			},
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			vm, _ := w3vm.New(
				w3vm.WithState(test.PreState),
			)
			gotReceipt, gotErr := vm.Call(test.Message)
			if diff := cmp.Diff(test.WantErr, gotErr,
				internal.EquateErrors(),
			); diff != "" {
				t.Fatalf("(-want +got)\n%s", diff)
			}
			if diff := cmp.Diff(test.WantReceipt, gotReceipt,
				internal.EquateErrors(),
				cmpopts.IgnoreUnexported(w3vm.Receipt{}),
				cmpopts.EquateComparable(common.Address{}, common.Hash{}),
			); diff != "" {
				t.Fatalf("(-want +got)\n%s", diff)
			}
		})
	}
}

func TestVMCallFunc(t *testing.T) {
	vm, _ := w3vm.New(
		w3vm.WithState(w3types.State{
			addrWETH: {
				Code: codeWETH,
				Storage: map[common.Hash]common.Hash{
					w3vm.WETHBalanceSlot(addr0): common.BigToHash(w3.I("1 ether")),
				},
			},
		}),
	)

	var gotBalance *big.Int
	if err := vm.CallFunc(addrWETH, funcBalanceOf, addr0).Returns(&gotBalance); err != nil {
		t.Fatalf("Failed to call balanceOf: %v", err)
	}

	wantBalance := w3.I("1 ether")
	if wantBalance.Cmp(gotBalance) != 0 {
		t.Fatalf("Balance: want %s, got %s", wantBalance, gotBalance)
	}
}

func TestVM_Fetcher(t *testing.T) {
	f := new(testFetcher)
	vm, err := w3vm.New(
		w3vm.WithFetcher(f),
	)
	if err != nil {
		t.Fatalf("Failed to create VM: %v", err)
	}

	_, err = vm.Nonce(addr0)
	want := "fetching failed: failed to fetch nonce of 0x0000000000000000000000000000000000000000"
	if !errors.Is(err, w3vm.ErrFetch) || want != err.Error() {
		t.Errorf("Nonce: want %q, got %q", want, err)
	}

	_, err = vm.Balance(addr0)
	want = "fetching failed: failed to fetch balance of 0x0000000000000000000000000000000000000000"
	if !errors.Is(err, w3vm.ErrFetch) || want != err.Error() {
		t.Errorf("Balance: want %q, got %q", want, err)
	}

	_, err = vm.Code(addr0)
	want = "fetching failed: failed to fetch code of 0x0000000000000000000000000000000000000000"
	if !errors.Is(err, w3vm.ErrFetch) || want != err.Error() {
		t.Errorf("Code: want %q, got %q", want, err)
	}

	_, err = vm.StorageAt(addr0, common.Hash{})
	want = "fetching failed: failed to fetch storage of 0x0000000000000000000000000000000000000000 at 0x0000000000000000000000000000000000000000000000000000000000000000"
	if !errors.Is(err, w3vm.ErrFetch) || want != err.Error() {
		t.Errorf("StorageAt: want %q, got %q", want, err)
	}
}

type testFetcher struct{}

func (f *testFetcher) Nonce(addr common.Address) (uint64, error) {
	return 0, fmt.Errorf("%w: failed to fetch nonce", w3vm.ErrFetch)
}

func (f *testFetcher) Balance(addr common.Address) (*big.Int, error) {
	// return nil, fmt.Errorf("%w: failed to fetch balance", w3vm.ErrFetch)
	return big.NewInt(1), nil
}

func (f *testFetcher) Code(addr common.Address) ([]byte, error) {
	return nil, fmt.Errorf("%w: failed to code", w3vm.ErrFetch)
}

func (f *testFetcher) StorageAt(addr common.Address, key common.Hash) (common.Hash, error) {
	return common.Hash{}, fmt.Errorf("%w: failed to fetch storage", w3vm.ErrFetch)
}

func (f *testFetcher) HeaderHash(blockNumber *big.Int) (common.Hash, error) {
	return common.Hash{}, fmt.Errorf("%w: failed to fetch code hash", w3vm.ErrFetch)
}

func TestVMApply_Integration(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	blocks := []*big.Int{
		big.NewInt(4_369_998),
		big.NewInt(4_369_999),
		big.NewInt(4_370_000), // Byzantium
		big.NewInt(4_370_001),

		big.NewInt(7_279_998),
		big.NewInt(7_279_999),
		big.NewInt(7_280_000), // Constantinople & Petersburg
		big.NewInt(7_280_001),

		big.NewInt(9_068_998),
		big.NewInt(9_068_999),
		big.NewInt(9_069_000), // Istanbul
		big.NewInt(9_069_001),

		big.NewInt(9_199_998),
		big.NewInt(9_199_999),
		big.NewInt(9_200_000), // Muir Glacier
		big.NewInt(9_200_001),

		big.NewInt(12_243_998),
		big.NewInt(12_243_999),
		big.NewInt(12_244_000), // Berlin
		big.NewInt(12_244_001),

		big.NewInt(12_964_998),
		big.NewInt(12_964_999),
		big.NewInt(12_965_000), // London
		big.NewInt(12_965_001),

		big.NewInt(13_772_998),
		big.NewInt(13_772_999),
		big.NewInt(13_773_000), // Arrow Glacier
		big.NewInt(13_773_001),

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

			var (
				block    types.Block
				receipts types.Receipts
			)
			if err := client.Call(
				eth.BlockByNumber(number).Returns(&block),
				eth.BlockReceipts(number).Returns(&receipts),
			); err != nil {
				t.Fatalf("Failed to fetch block and receipts: %v", err)
			}

			f := w3vm.NewTestingRPCFetcher(t, client, new(big.Int).Sub(number, w3.Big1))
			vm, _ := w3vm.New(
				w3vm.WithFetcher(f),
				w3vm.WithHeader(block.Header()),
			)

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

					gotReceipt, err := vm.ApplyTx(tx)
					if err != nil && gotReceipt == nil {
						t.Fatalf("Failed to apply tx: %v", err)
					}
					if diff := cmp.Diff(wantReceipt, gotReceipt,
						cmpopts.EquateEmpty(),
						cmpopts.EquateErrors(),
						cmpopts.IgnoreUnexported(w3vm.Receipt{}),
						cmpopts.IgnoreFields(w3vm.Receipt{}, "GasRefund", "GasLimit", "Output"),
						cmpopts.IgnoreFields(types.Log{}, "BlockHash", "BlockNumber", "TxHash", "TxIndex", "Index"),
						cmpopts.EquateComparable(common.Address{}, common.Hash{}),
					); diff != "" {
						t.Fatalf("(-want +got)\n%s", diff)
					}
				})
			}
		})
	}
}

func must[T any](t T, err error) T {
	if err != nil {
		panic(err)
	}
	return t
}

func BenchmarkTransferWETH9(b *testing.B) {
	addr0 := w3vm.RandA()
	addr1 := w3vm.RandA()

	// encode input
	input := must(funcTransfer.EncodeArgs(addr1, w3.I("1 gwei")))

	blockCtx := vm.BlockContext{
		CanTransfer: core.CanTransfer,
		Transfer:    core.Transfer,
		GetHash:     func(n uint64) common.Hash { return common.Hash{} },

		BlockNumber: new(big.Int),
		Difficulty:  new(big.Int),
		BaseFee:     new(big.Int),
		GasLimit:    30_000_000,
	}

	b.Run("w3vm", func(b *testing.B) {
		vm, _ := w3vm.New(
			w3vm.WithBlockContext(&blockCtx),
			w3vm.WithChainConfig(params.AllEthashProtocolChanges),
			w3vm.WithState(w3types.State{
				addrWETH: {
					Code: codeWETH,
					Storage: map[common.Hash]common.Hash{
						w3vm.WETHBalanceSlot(addr0): common.BigToHash(w3.I("1 ether")),
					},
				},
			}),
		)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := vm.Apply(&w3types.Message{
				From:  addr0,
				To:    &addrWETH,
				Gas:   100_000,
				Nonce: uint64(i),
				Input: input,
			})
			if err != nil {
				b.Fatalf("Failed to apply msg: %v", err)
			}
		}
	})

	b.Run("geth", func(b *testing.B) {
		stateDB, _ := coreState.New(common.Hash{}, coreState.NewDatabase(rawdb.NewMemoryDatabase()), nil)
		stateDB.SetCode(addrWETH, codeWETH)
		stateDB.SetState(addrWETH, w3vm.WETHBalanceSlot(addr0), common.BigToHash(w3.I("1 ether")))

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			msg := &core.Message{
				To:                &addrWETH,
				From:              addr0,
				Nonce:             uint64(i),
				Value:             new(big.Int),
				GasLimit:          100_000,
				GasPrice:          new(big.Int),
				GasFeeCap:         new(big.Int),
				GasTipCap:         new(big.Int),
				Data:              input,
				AccessList:        nil,
				SkipAccountChecks: false,
			}
			txCtx := core.NewEVMTxContext(msg)
			evm := vm.NewEVM(blockCtx, txCtx, stateDB, params.AllEthashProtocolChanges, vm.Config{NoBaseFee: true})
			gp := new(core.GasPool).AddGas(math.MaxUint64)
			_, err := core.ApplyMessage(evm, msg, gp)
			if err != nil {
				b.Fatalf("Failed to apply msg: %v", err)
			}
			stateDB.Finalise(false)
		}
	})
}

func ptr[T any](t T) *T { return &t }

func ExampleVM() {
	var (
		addrEOA    = w3.A("0x000000000000000000000000000000000000c0Fe")
		addrWETH   = w3.A("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2")
		addrUNI    = w3.A("0x1f9840a85d5aF5bf1D1762F925BDADdC4201F984")
		addrRouter = w3.A("0xE592427A0AEce92De3Edee1F18E0157C05861564")

		funcExactInput = w3.MustNewFunc(`exactInput(
			(
			bytes path,
			address recipient,
			uint256 deadline,
			uint256 amountIn,
			uint256 amountOutMinimum
			) params
		)`, "uint256 amountOut")
	)

	type ExactInputParams struct {
		Path             []byte
		Recipient        common.Address
		Deadline         *big.Int
		AmountIn         *big.Int
		AmountOutMinimum *big.Int
	}

	encodePath := func(tokenA common.Address, fee uint32, tokenB common.Address) []byte {
		path := make([]byte, 43)
		copy(path, tokenA[:])
		path[20], path[21], path[22] = byte(fee>>16), byte(fee>>8), byte(fee)
		copy(path[23:], tokenB[:])
		return path
	}

	client, err := w3.Dial("https://rpc.ankr.com/eth")
	if err != nil {
		// handle error
	}
	defer client.Close()

	// 1. Create a VM that forks the Mainnet state from the latest block,
	// disables the base fee, and has a fake WETH balance and approval for the router
	vm, err := w3vm.New(
		w3vm.WithFork(client, nil),
		w3vm.WithNoBaseFee(),
		w3vm.WithState(w3types.State{
			addrWETH: {Storage: map[common.Hash]common.Hash{
				w3vm.WETHBalanceSlot(addrEOA):               common.BigToHash(w3.I("1 ether")),
				w3vm.WETHAllowanceSlot(addrEOA, addrRouter): common.BigToHash(w3.I("1 ether")),
			}},
		}),
	)
	if err != nil {
		// handle error
	}

	// 2. Simulate a UniSwap v3 swap
	receipt, err := vm.Apply(&w3types.Message{
		From: addrEOA,
		To:   &addrRouter,
		Func: funcExactInput,
		Args: []any{&ExactInputParams{
			Path:             encodePath(addrWETH, 500, addrUNI),
			Recipient:        addrEOA,
			Deadline:         big.NewInt(time.Now().Unix()),
			AmountIn:         w3.I("1 ether"),
			AmountOutMinimum: w3.Big0,
		}},
	})
	if err != nil {
		// handle error
	}

	// 3. Decode output amount
	var amountOut *big.Int
	if err := receipt.DecodeReturns(&amountOut); err != nil {
		// handle error
	}

	fmt.Printf("amount out: %s UNI\n", w3.FromWei(amountOut, 18))
}

func ExampleVM_Call() {
	client := w3.MustDial("https://rpc.ankr.com/eth")
	defer client.Close()

	addrWETH := w3.A("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2")
	addrEOA := w3.A("0x000000000000000000000000000000000000c0Fe")

	vm, err := w3vm.New(
		w3vm.WithFork(client, nil),
		w3vm.WithState(w3types.State{
			addrWETH: {Storage: map[common.Hash]common.Hash{
				w3vm.WETHBalanceSlot(addrEOA): common.BigToHash(w3.I("1 ether")),
			}},
		}),
	)
	if err != nil {
		// handle error
	}

	balanceOf := w3.MustNewFunc("balanceOf(address)", "uint256")
	var balance *big.Int
	if err := vm.CallFunc(addrWETH, balanceOf, addrEOA).Returns(&balance); err != nil {
		// handle error
	}
	fmt.Printf("%s: Balance: %s WETH\n", addrEOA, w3.FromWei(balance, 18))
	// Output:
	// 0x000000000000000000000000000000000000c0Fe: Balance: 1 WETH
}
