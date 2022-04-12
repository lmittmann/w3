package w3_test

import (
	"context"
	"flag"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/lmittmann/w3"
	"github.com/lmittmann/w3/core"
	"github.com/lmittmann/w3/module/eth"
)

var benchRPC = flag.String("benchRPC", "", "RPC endpoint for benchmark")

func BenchmarkCall_BalanceNonce(b *testing.B) {
	if *benchRPC == "" {
		b.Skipf("Missing -benchRPC")
	}

	w3Client := w3.MustDial(*benchRPC)
	defer w3Client.Close()

	ethClient, _ := ethclient.Dial(*benchRPC)
	defer ethClient.Close()

	var (
		addr = common.BigToAddress(w3.Big0)

		nonce   uint64
		balance = new(big.Int)
	)

	b.Run("Batch", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			w3Client.Call(
				eth.Nonce(addr, nil).Returns(&nonce),
				eth.Balance(addr, nil).Returns(balance),
			)
		}
	})

	b.Run("Sequential", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			nonce, _ = ethClient.NonceAt(context.Background(), addr, nil)
			balance, _ = ethClient.BalanceAt(context.Background(), addr, nil)
		}
	})
}

func BenchmarkCall_Balance100(b *testing.B) {
	if *benchRPC == "" {
		b.Skipf("Missing -benchRPC")
	}

	w3Client := w3.MustDial(*benchRPC)
	defer w3Client.Close()

	ethClient, _ := ethclient.Dial(*benchRPC)
	defer ethClient.Close()

	addr100 := make([]common.Address, 100)
	for i := 0; i < len(addr100); i++ {
		addr100[i] = common.BigToAddress(big.NewInt(int64(i)))
	}

	balance := new(big.Int)

	b.Run("Batch", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			requests := make([]core.Caller, len(addr100))
			for j := 0; j < len(requests); j++ {
				requests[j] = eth.Balance(addr100[j], nil).Returns(balance)
			}
			w3Client.Call(requests...)
		}
	})

	b.Run("Sequential", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, addr := range addr100 {
				balance, _ = ethClient.BalanceAt(context.Background(), addr, nil)
			}
		}
	})
}

func BenchmarkCall_Block100(b *testing.B) {
	if *benchRPC == "" {
		b.Skipf("Missing -benchRPC")
	}

	w3Client := w3.MustDial(*benchRPC)
	defer w3Client.Close()

	ethClient, _ := ethclient.Dial(*benchRPC)
	defer ethClient.Close()

	block100 := make([]*big.Int, 100)
	for i := 0; i < len(block100); i++ {
		block100[i] = big.NewInt(int64(14_000_000 + i))
	}

	block := new(types.Block)

	b.Run("Batch", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			requests := make([]core.Caller, len(block100))
			for j := 0; j < len(requests); j++ {
				requests[j] = eth.BlockByNumber(block100[j]).Returns(block)
			}
			w3Client.Call(requests...)
		}
	})

	b.Run("Sequential", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, number := range block100 {
				block, _ = ethClient.BlockByNumber(context.Background(), number)
			}
		}
	})
}
