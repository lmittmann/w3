package w3vm_test

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/lmittmann/w3"
	"github.com/lmittmann/w3/module/eth"
	"github.com/lmittmann/w3/w3types"
	"github.com/lmittmann/w3/w3vm"
)

// BenchmarkVM runs [VM.ApplyTx] on the given block range and returns the
// simulated gas per second.
//
// The goal of this benchmark is to compare the VM performance in a real-world
// setting.
//
// The required block pre-state of this benchmark is stored in "testdata/w3vm/".
// If the block range is changed the initial run can be quite slow (a local
// Ethereum node is recommended).
func BenchmarkVM(b *testing.B) {
	const (
		startBlock int64 = 19_000_000
		nBlocks    int   = 10
	)

	// fetch blocks
	blocks := make([]*types.Block, nBlocks)
	calls := make([]w3types.RPCCaller, nBlocks)
	for i := range blocks {
		number := big.NewInt(startBlock + int64(i))
		calls[i] = eth.BlockByNumber(number).Returns(&blocks[i])
	}

	if err := testClient.Call(calls...); err != nil {
		b.Fatalf("Failed to fetch blocks: %v", err)
	}

	// execute blocks once to fetch the required state
	fetchers := make([]w3vm.Fetcher, 0, len(blocks))
	for _, block := range blocks {
		fetcher := w3vm.NewTestingRPCFetcher(b, 1, testClient, new(big.Int).Sub(block.Number(), w3.Big1))
		fetchers = append(fetchers, fetcher)
		vm, err := w3vm.New(
			w3vm.WithFetcher(fetcher),
			w3vm.WithHeader(block.Header()),
		)
		if err != nil {
			b.Fatalf("Failed to build VM for block %s: %v", block.Number(), err)
		}

		for _, tx := range block.Transactions() {
			vm.ApplyTx(tx)
		}
	}

	// benchmark
	b.ReportAllocs()
	b.ResetTimer()
	var (
		blockI  int // block index
		txI     int // tx index
		block   = blocks[blockI]
		vm, err = w3vm.New(
			w3vm.WithFetcher(fetchers[blockI]),
			w3vm.WithHeader(block.Header()),
		)
		gasSimulated uint64
	)
	if err != nil {
		b.Fatalf("Failed to build VM for block %s: %v", block.Number(), err)
	}

	for range b.N {
		if txI >= block.Transactions().Len() {
			blockI = (blockI + 1) % len(blocks)
			txI = 0
			block = blocks[blockI]
			vm, err = w3vm.New(
				w3vm.WithFetcher(fetchers[blockI]),
				w3vm.WithHeader(block.Header()),
			)
			if err != nil {
				b.Fatalf("Failed to build VM for block %s: %v", block.Number(), err)
			}
		}
		tx := block.Transactions()[txI]

		r, err := vm.ApplyTx(tx)
		if r == nil {
			b.Fatalf("Failed to apply tx %d: %v", txI, err)
		}
		gasSimulated += r.GasUsed
		txI++
	}

	// report simulated gas per second
	dur := b.Elapsed()
	b.ReportMetric(float64(gasSimulated)/dur.Seconds(), "gas/s")
}

//go:embed testdata/burntpix.genesis.json
var rawGenesisBurntpix []byte

func BenchmarkVMCall(b *testing.B) {
	var (
		addrQuoter   = w3.A("0xb27308f9F90D607463bb33eA1BeBb41C27CE5AB6")
		addrWETH     = w3.A("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2")
		addrUSDC     = w3.A("0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48")
		addrBurntpix = w3.A("0x49206861766520746f6f206d7563682074696d65")

		funcQuote     = w3.MustNewFunc("quoteExactInputSingle(address tokenIn, address tokenOut, uint24 fee, uint256 amountIn, uint160 sqrtPriceLimitX96)", "uint256 amountOut")
		funcBalanceOf = w3.MustNewFunc("balanceOf(address)", "uint256")
		funcRun       = w3.MustNewFunc("run(uint32 seed, uint256 iterations)", "string")

		genesisBurntpix types.GenesisAlloc
	)

	if err := json.Unmarshal(rawGenesisBurntpix, &genesisBurntpix); err != nil {
		b.Fatalf("Failed to unmarshal burntpix genesis: %v", err)
	}

	benchmarks := []struct {
		Name string
		Opts []w3vm.Option
		Msg  *w3types.Message
	}{
		{
			Name: "UniswapV3Quote",
			Opts: []w3vm.Option{
				w3vm.WithFork(testClient, big.NewInt(20_000_000)),
				w3vm.WithTB(b),
			},
			Msg: &w3types.Message{
				To:    &addrQuoter,
				Input: mustEncodeArgs(funcQuote, addrWETH, addrUSDC, big.NewInt(500), w3.BigEther, w3.Big0),
			},
		},
		{
			Name: "WethBalanceOf",
			Opts: []w3vm.Option{
				w3vm.WithFork(testClient, big.NewInt(20_000_000)),
				w3vm.WithTB(b),
			},
			Msg: &w3types.Message{
				To:    &addrWETH,
				Input: mustEncodeArgs(funcBalanceOf, addrWETH),
			},
		},
		{ // https://github.com/karalabe/burntpix-benchmark
			Name: "Burntpix",
			Opts: []w3vm.Option{
				w3vm.WithChainConfig(params.AllDevChainProtocolChanges),
				w3vm.WithHeader(&types.Header{
					GasLimit: 0xffffffffffffffff,
				}),
				w3vm.WithState(w3types.State{}.SetGenesisAlloc(genesisBurntpix)),
			},
			Msg: &w3types.Message{
				From:  w3vm.RandA(),
				To:    &addrBurntpix,
				Gas:   0xffffffffffffffff,
				Input: mustEncodeArgs(funcRun, uint32(0), big.NewInt(500_000)),
			},
		},
	}

	for _, bench := range benchmarks {
		b.Run(bench.Name, func(b *testing.B) {
			// setup VM
			vm, err := w3vm.New(bench.Opts...)
			if err != nil {
				b.Fatalf("Failed to build VM: %v", err)
			}

			b.ReportAllocs()
			b.ResetTimer()

			var gasSimulated uint64
			for range b.N {
				receipt, err := vm.Call(bench.Msg)
				if err != nil {
					b.Fatal(err)
				}
				gasSimulated += receipt.GasUsed
			}

			// report simulated gas per second
			dur := b.Elapsed()
			b.ReportMetric(float64(gasSimulated)/dur.Seconds(), "gas/s")
		})
	}
}

func BenchmarkVMSnapshot(b *testing.B) {
	depositMsg := &w3types.Message{
		From:  addr0,
		To:    &addrWETH,
		Value: w3.I("1 ether"),
	}

	runs := 2
	b.Run(fmt.Sprintf("re-run %d", runs), func(b *testing.B) {
		for range b.N {
			vm, _ := w3vm.New(
				w3vm.WithState(w3types.State{
					addrWETH: {Code: codeWETH},
					addr0:    {Balance: w3.I("2 ether")},
				}),
			)

			for range runs {
				_, err := vm.Apply(depositMsg)
				if err != nil {
					b.Fatalf("Failed to deposit: %v", err)
				}
			}
		}
	})

	b.Run(fmt.Sprintf("snapshot %d", runs), func(b *testing.B) {
		vm, _ := w3vm.New(
			w3vm.WithState(w3types.State{
				addrWETH: {Code: codeWETH},
				addr0:    {Balance: w3.I("2 ether")},
			}),
		)

		for i := 0; i < runs-1; i++ {
			_, err := vm.Apply(depositMsg)
			if err != nil {
				b.Fatalf("Failed to deposit: %v", err)
			}
		}

		snap := vm.Snapshot()

		for range b.N {
			_, err := vm.Apply(depositMsg)
			if err != nil {
				b.Fatalf("Failed to deposit: %v", err)
			}

			vm.Rollback(snap.Copy())
		}
	})
}
