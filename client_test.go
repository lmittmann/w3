package w3_test

import (
	"context"
	"flag"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/lmittmann/w3"
	"github.com/lmittmann/w3/module/eth"
	"github.com/lmittmann/w3/w3types"
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

	addr := common.Address{}

	b.Run("Batch", func(b *testing.B) {
		var (
			nonce   uint64
			balance big.Int
		)
		for i := 0; i < b.N; i++ {
			w3Client.Call(
				eth.Nonce(addr, nil).Returns(&nonce),
				eth.Balance(addr, nil).Returns(&balance),
			)
		}
	})

	b.Run("Sequential", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ethClient.NonceAt(context.Background(), addr, nil)
			ethClient.BalanceAt(context.Background(), addr, nil)
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

	b.Run("Batch", func(b *testing.B) {
		var balance big.Int
		for i := 0; i < b.N; i++ {
			requests := make([]w3types.Caller, len(addr100))
			for j := 0; j < len(requests); j++ {
				requests[j] = eth.Balance(addr100[j], nil).Returns(&balance)
			}
			w3Client.Call(requests...)
		}
	})

	b.Run("Sequential", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, addr := range addr100 {
				ethClient.BalanceAt(context.Background(), addr, nil)
			}
		}
	})
}

func BenchmarkCall_BalanceOf100(b *testing.B) {
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

	funcBalanceOf := w3.MustNewFunc("balanceOf(address)", "uint256")
	addrWeth9 := w3.A("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2")

	b.Run("Batch", func(b *testing.B) {
		var balance big.Int
		for i := 0; i < b.N; i++ {
			requests := make([]w3types.Caller, len(addr100))
			for j := 0; j < len(requests); j++ {
				requests[j] = eth.CallFunc(funcBalanceOf, addrWeth9, addr100[j]).Returns(&balance)
			}
			w3Client.Call(requests...)
		}
	})

	b.Run("Sequential", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, addr := range addr100 {
				input, err := funcBalanceOf.EncodeArgs(addr)
				if err != nil {
					b.Fatalf("Failed to encode args: %v", err)
				}
				ethClient.CallContract(context.Background(), ethereum.CallMsg{
					To:   &addrWeth9,
					Data: input,
				}, nil)
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

	b.Run("Batch", func(b *testing.B) {
		var block types.Block
		for i := 0; i < b.N; i++ {
			requests := make([]w3types.Caller, len(block100))
			for j := 0; j < len(requests); j++ {
				requests[j] = eth.BlockByNumber(block100[j]).Returns(&block)
			}
			w3Client.Call(requests...)
		}
	})

	b.Run("Sequential", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, number := range block100 {
				ethClient.BlockByNumber(context.Background(), number)
			}
		}
	})
}
