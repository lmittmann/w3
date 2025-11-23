package main

import (
	"flag"
	"fmt"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/lmittmann/w3"
	"github.com/lmittmann/w3/module/eth"
	"github.com/lmittmann/w3/w3types"
	"github.com/lmittmann/w3/w3vm"
	"github.com/lmittmann/w3/w3vm/hooks"
)

var chainConfigs = map[uint64]*params.ChainConfig{
	1: params.MainnetChainConfig,
}

func main() {
	if err := run(flag.CommandLine, os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
}

func run(fs *flag.FlagSet, args []string) error {
	var (
		rpcURL            string
		displayOps        bool
		displayStaticcall bool
		overrideIndex     int
		decodeABI         bool
	)
	// parse flags
	fs.StringVar(&rpcURL, "rpc", "http://localhost:8545", "RPC endpoint")
	fs.BoolVar(&displayOps, "ops", false, "Display opcodes")
	fs.BoolVar(&displayStaticcall, "staticcall", false, "Display STATICCALL's")
	fs.IntVar(&overrideIndex, "i", -1, "Override transaction index")
	fs.BoolVar(&decodeABI, "abi", false, "Decode ABI")
	fs.Usage = func() {
		fmt.Fprintln(fs.Output(), "usage: trace [flags] [tx]")
		flag.PrintDefaults()
	}
	fs.Parse(args)

	// parse args
	if flag.NArg() != 1 {
		flag.Usage()
		return fmt.Errorf("missing tx hash")
	}
	txHash := common.HexToHash(flag.Arg(0))

	// connect to RPC endpoint
	client, err := w3.Dial(rpcURL)
	if err != nil {
		return fmt.Errorf("failed to connect to RPC endpoint: %w", err)
	}

	// populate overrides
	o := overrides{}
	if overrideIndex >= 0 {
		o.Index = &overrideIndex
	}

	tracer := hooks.NewCallTracer(os.Stdout, &hooks.CallTracerOptions{
		ShowOp:         func(op byte, pc uint64, addr common.Address) bool { return displayOps },
		ShowStaticcall: displayStaticcall,
		DecodeABI:      decodeABI,
	})

	if err := execTrace(client, txHash, o, tracer); err != nil {
		return fmt.Errorf("failed to execute trace: %w", err)
	}
	return nil
}

func execTrace(client *w3.Client, txHash common.Hash, o overrides, tracer *tracing.Hooks) error {
	// fetch tx and its receipt
	var (
		tx      *types.Transaction
		receipt *types.Receipt
		chainID uint64
	)
	if err := client.Call(
		eth.Tx(txHash).Returns(&tx),
		eth.TxReceipt(txHash).Returns(&receipt),
		eth.ChainID().Returns(&chainID),
	); err != nil {
		return fmt.Errorf("fetch tx and receipt: %w", err)
	}

	// fetch block
	var block *types.Block
	if err := client.Call(
		eth.BlockByNumber(receipt.BlockNumber).Returns(&block),
	); err != nil {
		return fmt.Errorf("fetch block: %w", err)
	}

	config := chainConfigs[chainID]
	signer := types.MakeSigner(config, block.Number(), block.Time())

	// prepare fetcher and chain
	fetcher := w3vm.NewRPCFetcher(client, new(big.Int).Sub(receipt.BlockNumber, w3.Big1))
	vm, err := w3vm.New(
		w3vm.WithFetcher(fetcher),
		w3vm.WithHeader(block.Header()),
		w3vm.WithChainConfig(config),
	)
	if err != nil {
		return fmt.Errorf("create vm: %w", err)
	}

	// index of the last tx to apply before the target tx (if -1 then apply none)
	prestateEndIndex := int(receipt.TransactionIndex)
	if o.Index != nil && *o.Index < prestateEndIndex {
		prestateEndIndex = *o.Index
	}

	// prefetch state asynchronously
	var prefetch []chan struct{}
	if prestateEndIndex > 0 {
		prefetch = make([]chan struct{}, prestateEndIndex)
	}

	for i := range prestateEndIndex {
		tx := block.Transactions()[i]
		prefetch[i] = make(chan struct{})
		go func(tx *types.Transaction, ch chan struct{}) {
			tempVM, _ := w3vm.New(
				w3vm.WithFetcher(fetcher),
				w3vm.WithHeader(block.Header()),
				w3vm.WithChainConfig(config),
			)
			msg := new(w3types.Message).MustSetTx(tx, signer)
			tempVM.Apply(msg)

			close(ch) // signal that the tx has been prefetched
		}(tx, prefetch[i])
	}

	// apply all prestate txs
	for i := range prestateEndIndex {
		<-prefetch[i]
		msg := new(w3types.Message).MustSetTx(block.Transactions()[i], signer)
		vm.Apply(msg)
	}
	// apply target tx
	_, err = vm.ApplyTx(tx, tracer)

	return err
}

// overrides stores overrides for the tx execution.
type overrides struct {
	Index *int // transaction index in the block
}
