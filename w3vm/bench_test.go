package w3vm_test

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/core/types"
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
		endBlock   int64 = startBlock + 10
	)

	// fetch blocks
	blocks := make([]*types.Block, 0, endBlock-startBlock)
	calls := make([]w3types.RPCCaller, 0, endBlock-startBlock)
	for i := startBlock; i < endBlock; i++ {
		block := new(types.Block)
		blocks = append(blocks, block)
		calls = append(calls, eth.BlockByNumber(big.NewInt(i)).Returns(block))
	}

	if err := client.Call(calls...); err != nil {
		b.Fatalf("Failed to fetch blocks: %v", err)
	}

	// execute blocks once to fetch the required state
	fetchers := make([]w3vm.Fetcher, 0, len(blocks))
	for _, block := range blocks {
		fetcher := w3vm.NewTestingRPCFetcher(b, 1, client, new(big.Int).Sub(block.Number(), w3.Big1))
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
		tx      = block.Transactions()[txI]
		vm, err = w3vm.New(
			w3vm.WithFetcher(fetchers[blockI]),
			w3vm.WithHeader(block.Header()),
		)
		gasSimulated uint64
	)
	if err != nil {
		b.Fatalf("Failed to build VM for block %s: %v", block.Number(), err)
	}

	for i := 0; i < b.N; i++ {
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
		tx = block.Transactions()[txI]

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
