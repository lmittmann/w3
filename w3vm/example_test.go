package w3vm_test

import (
	"fmt"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/types"
	gethVm "github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/eth/tracers/logger"
	"github.com/ethereum/go-ethereum/params"
	"github.com/lmittmann/w3"
	"github.com/lmittmann/w3/module/eth"
	"github.com/lmittmann/w3/w3types"
	"github.com/lmittmann/w3/w3vm"
	"github.com/lmittmann/w3/w3vm/hooks"
)

var (
	addrA = common.Address{0x0a}
	addrB = common.Address{0x0b}
)

// Execute an Ether transfer.
func ExampleVM_simpleTransfer() {
	vm, _ := w3vm.New(
		w3vm.WithState(w3types.State{
			addrA: {Balance: w3.I("100 ether")},
		}),
	)

	// Print balances
	balA, _ := vm.Balance(addrA)
	balB, _ := vm.Balance(addrB)
	fmt.Printf("Before transfer:\nA: %s ETH, B: %s ETH\n", w3.FromWei(balA, 18), w3.FromWei(balB, 18))

	// Transfer 10 ETH from A to B
	vm.Apply(&w3types.Message{
		From:  addrA,
		To:    &addrB,
		Value: w3.I("10 ether"),
	})

	// Print balances
	balA, _ = vm.Balance(addrA)
	balB, _ = vm.Balance(addrB)
	fmt.Printf("After transfer:\nA: %s ETH, B: %s ETH\n", w3.FromWei(balA, 18), w3.FromWei(balB, 18))
	// Output:
	// Before transfer:
	// A: 100 ETH, B: 0 ETH
	// After transfer:
	// A: 90 ETH, B: 10 ETH
}

// Execute an ERC20 token transfer with faked token balance (Wrapped Ether).
func ExampleVM_fakeTokenBalance() {
	vm, err := w3vm.New(
		w3vm.WithFork(client, nil),
		w3vm.WithNoBaseFee(),
		w3vm.WithState(w3types.State{
			addrWETH: {Storage: w3types.Storage{
				w3vm.WETHBalanceSlot(addrA): common.BigToHash(w3.I("100 ether")),
			}},
		}),
	)
	if err != nil {
		// ...
	}

	// Print WETH balance
	var balA, balB *big.Int
	if err := vm.CallFunc(addrWETH, funcBalanceOf, addrA).Returns(&balA); err != nil {
		// ...
	}
	if err := vm.CallFunc(addrWETH, funcBalanceOf, addrB).Returns(&balB); err != nil {
		// ...
	}
	fmt.Printf("Before transfer:\nA: %s WETH, B: %s WETH\n", w3.FromWei(balA, 18), w3.FromWei(balB, 18))

	// Transfer 10 WETH from A to B
	if _, err := vm.Apply(&w3types.Message{
		From: addrA,
		To:   &addrWETH,
		Func: funcTransfer,
		Args: []any{addrB, w3.I("10 ether")},
	}); err != nil {
		// ...
	}

	// Print WETH balance
	if err := vm.CallFunc(addrWETH, funcBalanceOf, addrA).Returns(&balA); err != nil {
		// ...
	}
	if err := vm.CallFunc(addrWETH, funcBalanceOf, addrB).Returns(&balB); err != nil {
		// ...
	}
	fmt.Printf("After transfer:\nA: %s WETH, B: %s WETH\n", w3.FromWei(balA, 18), w3.FromWei(balB, 18))
	// Output:
	// Before transfer:
	// A: 100 WETH, B: 0 WETH
	// After transfer:
	// A: 90 WETH, B: 10 WETH
}

// Execute an ERC20 balanceOf call with raw a [w3types.Message] using the
// messages Func and Args helper.
func ExampleVM_call() {
	vm, err := w3vm.New(
		w3vm.WithFork(client, nil),
		w3vm.WithState(w3types.State{
			addrWETH: {Storage: w3types.Storage{
				w3vm.WETHBalanceSlot(addrA): common.BigToHash(w3.I("100 ether")),
			}},
		}),
	)
	if err != nil {
		// ...
	}

	receipt, err := vm.Call(&w3types.Message{
		To:   &addrWETH,
		Func: funcBalanceOf,
		Args: []any{addrA},
	})
	if err != nil {
		// ...
	}

	var balance *big.Int
	if err := receipt.DecodeReturns(&balance); err != nil {
		// ...
	}
	fmt.Printf("Balance: %s WETH\n", w3.FromWei(balance, 18))
	// Output:
	// Balance: 100 WETH
}

// Execute an ERC20 balanceOf call using the [VM.CallFunc] helper.
func ExampleVM_callFunc() {
	vm, err := w3vm.New(
		w3vm.WithFork(client, nil),
		w3vm.WithState(w3types.State{
			addrWETH: {Storage: w3types.Storage{
				w3vm.WETHBalanceSlot(addrA): common.BigToHash(w3.I("100 ether")),
			}},
		}),
	)
	if err != nil {
		// ...
	}

	var balance *big.Int
	if err := vm.CallFunc(addrWETH, funcBalanceOf, addrA).Returns(&balance); err != nil {
		// ...
	}
	fmt.Printf("Balance: %s WETH\n", w3.FromWei(balance, 18))
	// Output:
	// Balance: 100 WETH
}

// Execute an Uniswap V3 swap.
func ExampleVM_uniswapV3Swap() {
	var (
		addrRouter = w3.A("0xE592427A0AEce92De3Edee1F18E0157C05861564")
		addrUNI    = w3.A("0x1f9840a85d5aF5bf1D1762F925BDADdC4201F984")

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

	// mapping for the exactInput-function params-tuple
	type ExactInputParams struct {
		Path             []byte
		Recipient        common.Address
		Deadline         *big.Int
		AmountIn         *big.Int
		AmountOutMinimum *big.Int
	}

	encodePath := func(tokenA, tokenB common.Address, fee uint32) []byte {
		path := make([]byte, 43)
		copy(path, tokenA[:])
		path[20], path[21], path[22] = byte(fee>>16), byte(fee>>8), byte(fee)
		copy(path[23:], tokenB[:])
		return path
	}

	// 1. Create a VM that forks the Mainnet state from the latest block,
	// disables the base fee, and has a fake WETH balance and approval for the router
	vm, err := w3vm.New(
		w3vm.WithFork(client, big.NewInt(20_000_000)),
		w3vm.WithNoBaseFee(),
		w3vm.WithState(w3types.State{
			addrWETH: {Storage: w3types.Storage{
				w3vm.WETHBalanceSlot(addrA):               common.BigToHash(w3.I("1 ether")),
				w3vm.WETHAllowanceSlot(addrA, addrRouter): common.BigToHash(w3.I("1 ether")),
			}},
		}),
	)
	if err != nil {
		// ...
	}

	// 2. Simulate a Uniswap v3 swap
	receipt, err := vm.Apply(&w3types.Message{
		From: addrA,
		To:   &addrRouter,
		Func: funcExactInput,
		Args: []any{&ExactInputParams{
			Path:             encodePath(addrWETH, addrUNI, 500),
			Recipient:        addrA,
			Deadline:         big.NewInt(time.Now().Unix()),
			AmountIn:         w3.I("1 ether"),
			AmountOutMinimum: w3.Big0,
		}},
	})
	if err != nil {
		// ...
	}

	// 3. Decode output amount
	var amountOut *big.Int
	if err := receipt.DecodeReturns(&amountOut); err != nil {
		// ...
	}

	fmt.Printf("AmountOut: %s UNI\n", w3.FromWei(amountOut, 18))
	// Output:
	// AmountOut: 278.327327986946583271 UNI
}

// Execute a message sent from the zero address.
// The [w3types.Message] sender can be freely chosen, making it possible to
// execute a message from any address.
func ExampleVM_prankZeroAddress() {
	vm, err := w3vm.New(
		w3vm.WithFork(client, big.NewInt(20_000_000)),
		w3vm.WithNoBaseFee(),
	)
	if err != nil {
		// ...
	}

	balZero, err := vm.Balance(w3.Addr0)
	if err != nil {
		// ...
	}

	_, err = vm.Apply(&w3types.Message{
		From:  w3.Addr0,
		To:    &addrA,
		Value: balZero,
	})
	if err != nil {
		// ...
	}

	balance, err := vm.Balance(addrA)
	if err != nil {
		// ...
	}

	fmt.Printf("Received %s ETH from zero address\n", w3.FromWei(balance, 18))
	// Output:
	// Received 13365.401185473565028721 ETH from zero address
}

// Trace calls (and opcodes) of a transaction.
func ExampleVM_traceCalls() {
	txHash := w3.H("0xc0679fedfe8d7c376d599cbab03de7b527347a3d135d7d8d698047f34a6611f8")

	var (
		tx      *types.Transaction
		receipt *types.Receipt
	)
	if err := client.Call(
		eth.Tx(txHash).Returns(&tx),
		eth.TxReceipt(txHash).Returns(&receipt),
	); err != nil {
		// ...
	}

	vm, err := w3vm.New(
		w3vm.WithFork(client, receipt.BlockNumber),
	)
	if err != nil {
		// ...
	}

	callTracer := hooks.NewCallTracer(os.Stdout, &hooks.CallTracerOptions{
		ShowStaticcall: true,
		DecodeABI:      true,
	})
	vm.ApplyTx(tx, callTracer)
}

// Trace a message execution to obtain the access list.
func ExampleVM_traceAccessList() {
	txHash := w3.H("0xbb4b3fc2b746877dce70862850602f1d19bd890ab4db47e6b7ee1da1fe578a0d")

	var (
		tx      *types.Transaction
		receipt *types.Receipt
	)
	if err := client.Call(
		eth.Tx(txHash).Returns(&tx),
		eth.TxReceipt(txHash).Returns(&receipt),
	); err != nil {
		// ...
	}

	var header *types.Header
	if err := client.Call(eth.HeaderByNumber(receipt.BlockNumber).Returns(&header)); err != nil {
		// ...
	}

	vm, err := w3vm.New(
		w3vm.WithFork(client, receipt.BlockNumber),
	)
	if err != nil {
		// ...
	}

	// setup access list tracer
	signer := types.MakeSigner(params.MainnetChainConfig, header.Number, header.Time)
	from, _ := signer.Sender(tx)
	accessListTracer := logger.NewAccessListTracer(
		nil,
		from, *tx.To(),
		gethVm.ActivePrecompiles(params.MainnetChainConfig.Rules(header.Number, header.Difficulty.Sign() == 0, header.Time)),
	)

	if _, err := vm.ApplyTx(tx, accessListTracer.Hooks()); err != nil {
		// ...
	}
	fmt.Println("Access List:", accessListTracer.AccessList())
}

// Trace the execution of all op's in a block.
func ExampleVM_traceBlock() {
	blockNumber := big.NewInt(20_000_000)

	var block *types.Block
	if err := client.Call(eth.BlockByNumber(blockNumber).Returns(&block)); err != nil {
		// ...
	}

	vm, err := w3vm.New(
		w3vm.WithFork(client, blockNumber),
	)
	if err != nil {
		// ...
	}

	var opCount [256]uint64
	tracer := &tracing.Hooks{
		OnOpcode: func(pc uint64, op byte, gas, cost uint64, scope tracing.OpContext, rData []byte, depth int, err error) {
			opCount[op]++
		},
	}

	for _, tx := range block.Transactions() {
		vm.ApplyTx(tx, tracer)
	}

	for op, count := range opCount {
		if count > 0 {
			fmt.Printf("0x%02x %-14s %d\n", op, gethVm.OpCode(op), count)
		}
	}
}

func TestWETHDeposit(t *testing.T) {
	// setup VM
	vm, _ := w3vm.New(
		w3vm.WithState(w3types.State{
			addrWETH: {Code: codeWETH},
			addrA:    {Balance: w3.I("1 ether")},
		}),
	)

	// pre check
	var wethBalanceBefore *big.Int
	if err := vm.CallFunc(addrWETH, funcBalanceOf, addrA).Returns(&wethBalanceBefore); err != nil {
		t.Fatal(err)
	}
	if wethBalanceBefore.Sign() != 0 {
		t.Fatal("Invalid WETH balance: want 0")
	}

	// deposit (via fallback)
	if _, err := vm.Apply(&w3types.Message{
		From:  addrA,
		To:    &addrWETH,
		Value: w3.I("1 ether"),
	}); err != nil {
		t.Fatalf("Deposit failed: %v", err)
	}

	// post check
	var wethBalanceAfter *big.Int
	if err := vm.CallFunc(addrWETH, funcBalanceOf, addrA).Returns(&wethBalanceAfter); err != nil {
		t.Fatal(err)
	}
	if w3.I("1 ether").Cmp(wethBalanceAfter) != 0 {
		t.Fatalf("Invalid WETH balance: want 1")
	}
}

func FuzzWETHDeposit(f *testing.F) {
	f.Add([]byte{1})
	f.Fuzz(func(t *testing.T, amountBytes []byte) {
		if len(amountBytes) > 32 {
			t.Skip()
		}
		amount := new(big.Int).SetBytes(amountBytes)

		// setup VM
		vm, _ := w3vm.New(
			w3vm.WithState(w3types.State{
				addrWETH: {Code: codeWETH},
				addrA:    {Balance: w3.BigMaxUint256},
			}),
		)

		// Pre-check WETH balance
		var wethBalanceBefore *big.Int
		if err := vm.CallFunc(addrWETH, funcBalanceOf, addrA).Returns(&wethBalanceBefore); err != nil {
			t.Fatal(err)
		}

		// Attempt deposit
		vm.Apply(&w3types.Message{
			From:  addrA,
			To:    &addrWETH,
			Value: amount,
		})

		// Post-check WETH balance
		var wethBalanceAfter *big.Int
		if err := vm.CallFunc(addrWETH, funcBalanceOf, addrA).Returns(&wethBalanceAfter); err != nil {
			t.Fatal(err)
		}

		// Verify balance increment
		wantBalance := new(big.Int).Add(wethBalanceBefore, amount)
		if wethBalanceAfter.Cmp(wantBalance) != 0 {
			t.Fatalf("Invalid WETH balance: want %s, got %s", wantBalance, wethBalanceAfter)
		}
	})
}

func ExampleWETHBalanceSlot() {
	addrC0fe := w3.A("0x000000000000000000000000000000000000c0Fe")
	addrWETH := w3.A("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2")
	funcBalanceOf := w3.MustNewFunc("balanceOf(address)", "uint256")

	vm, err := w3vm.New(
		w3vm.WithFork(client, nil),
		w3vm.WithState(w3types.State{
			addrWETH: {
				Storage: w3types.Storage{
					w3vm.WETHBalanceSlot(addrC0fe): common.BigToHash(w3.I("100 ether")),
				},
			},
		}),
	)
	if err != nil {
		// ...
	}

	var balance *big.Int
	err = vm.CallFunc(addrWETH, funcBalanceOf, addrC0fe).Returns(&balance)
	if err != nil {
		// ...
	}
	fmt.Printf("%s: %s WETH", addrC0fe, w3.FromWei(balance, 18))
	// Output:
	// 0x000000000000000000000000000000000000c0Fe: 100 WETH
}
