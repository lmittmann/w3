package txpool_test

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/lmittmann/w3/module/txpool"
	"github.com/lmittmann/w3/rpctest"
)

func TestContent(t *testing.T) {
	tests := []rpctest.TestCase[txpool.ContentResponse]{
		{
			Golden: "content",
			Call:   txpool.Content(),
			WantRet: txpool.ContentResponse{
				Pending: map[common.Address][]*types.Transaction{
					common.HexToAddress("0x000454307bB96E303044046a6eB2736D2aD560B6"): {
						types.NewTx(&types.DynamicFeeTx{
							ChainID:   big.NewInt(1),
							Nonce:     4652,
							GasTipCap: big.NewInt(31407912032),
							GasFeeCap: big.NewInt(202871575924),
							Gas:       1100000,
							To:        ptr(common.HexToAddress("0xEf1c6E67703c7BD7107eed8303Fbe6EC2554BF6B")),
							Value:     big.NewInt(81000000000000000),
							Data:      common.FromHex("0x3593564c000000000000000000000000000000000000000000000000000000000000006000000000000000000000000000000000000000000000000000000000000000a0000000000000000000000000000000000000000000000000000001896e196d5300000000000000000000000000000000000000000000000000000000000000030b090c00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000003000000000000000000000000000000000000000000000000000000000000006000000000000000000000000000000000000000000000000000000000000000c000000000000000000000000000000000000000000000000000000000000001e000000000000000000000000000000000000000000000000000000000000000400000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000011fc51222ce800000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000004657febe8d8000000000000000000000000000000000000000000000000000011fc51222ce800000000000000000000000000000000000000000000000000000000000000000a000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002000000000000000000000000c02aaa39b223fe8d0a0e5c4f27ead9083c756cc20000000000000000000000008d538a82c84d7003aa0e7f1098bd9dc5ea1777be000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000000"),
							R:         new(big.Int).SetBytes(common.FromHex("0xf91729846ac8bb780a7239b7f0157d53330bba310a0811f4eec2eae25172c252")),
							S:         new(big.Int).SetBytes(common.FromHex("0x1bb9a1b9aea8b128e0e8dc42b17664be50c7b4073973c730b6c4cf2a3b3503cb")),
						}),
					},
				},
				Queued: map[common.Address][]*types.Transaction{
					common.HexToAddress("0x1BA4Ca9ac6ff4CF192C11E8C1624563f302cAA87"): {
						types.NewTx(&types.DynamicFeeTx{
							ChainID:   big.NewInt(1),
							Nonce:     183,
							GasTipCap: big.NewInt(110000000),
							GasFeeCap: big.NewInt(20027736270),
							Gas:       99226,
							To:        ptr(common.HexToAddress("0x1BA4Ca9ac6ff4CF192C11E8C1624563f302cAA87")),
							Value:     big.NewInt(0),
							Data:      []byte{},
							R:         new(big.Int).SetBytes(common.FromHex("0xea35c7c0643b79664b0bbb7f42d64edd371508ae4c33c1f817a80a2655465935")),
							S:         new(big.Int).SetBytes(common.FromHex("0x76d39f794e9e1ee359d66b7d3b19b90aaf2051b2159c68f3bb8c954558989da8")),
						}),
					},
				},
			},
		},
	}

	rpctest.RunTestCases(t, tests)
}

func TestContentFrom(t *testing.T) {
	tests := []rpctest.TestCase[txpool.ContentFromResponse]{
		{
			Golden: "contentFrom",
			Call:   txpool.ContentFrom(common.HexToAddress("0x1BA4Ca9ac6ff4CF192C11E8C1624563f302cAA87")),
			WantRet: txpool.ContentFromResponse{
				Queued: []*types.Transaction{
					types.NewTx(&types.DynamicFeeTx{
						ChainID:   big.NewInt(1),
						Nonce:     183,
						GasTipCap: big.NewInt(110000000),
						GasFeeCap: big.NewInt(20027736270),
						Gas:       99226,
						To:        ptr(common.HexToAddress("0x1BA4Ca9ac6ff4CF192C11E8C1624563f302cAA87")),
						Value:     big.NewInt(0),
						Data:      []byte{},
						R:         new(big.Int).SetBytes(common.FromHex("0xea35c7c0643b79664b0bbb7f42d64edd371508ae4c33c1f817a80a2655465935")),
						S:         new(big.Int).SetBytes(common.FromHex("0x76d39f794e9e1ee359d66b7d3b19b90aaf2051b2159c68f3bb8c954558989da8")),
					}),
				},
			},
		},
	}

	rpctest.RunTestCases(t, tests)
}

func ptr[T any](x T) *T { return &x }
