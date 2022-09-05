package eth_test

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/lmittmann/w3"
	"github.com/lmittmann/w3/module/eth"
	"github.com/lmittmann/w3/w3types"
)

var (
	funcBalanceOf = w3.MustNewFunc("balanceOf(address)", "uint256")
)

func TestCall(t *testing.T) {
	tests := []testCase[[]byte]{
		{
			Golden: "call_func",
			Call: eth.Call(&w3types.Message{
				To:   w3.APtr("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2"),
				Func: funcBalanceOf,
				Args: []any{w3.A("0x000000000000000000000000000000000000c0Fe")},
			}, nil, nil),
			WantRet: common.BigToHash(big.NewInt(0)).Bytes(),
		},
		{
			Golden: "call_func__overrides",
			Call: eth.Call(&w3types.Message{
				To:   w3.APtr("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2"),
				Func: funcBalanceOf,
				Args: []any{w3.A("0x000000000000000000000000000000000000c0Fe")},
			}, nil, w3types.State{
				w3.A("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2"): &w3types.Account{
					Storage: map[common.Hash]common.Hash{
						w3.H("0xf68b260b81af177c0bf1a03b5d62b15aea1b486f8df26c77f33aed7538cfeb2c"): w3.H("0x000000000000000000000000000000000000000000000000000000000000002a"),
					},
				},
			}),
			WantRet: common.BigToHash(big.NewInt(42)).Bytes(),
		},
	}

	runTestCases(t, tests)
}

func TestEstimateGas(t *testing.T) {
	tests := []testCase[uint64]{
		{
			Golden: "estimate_gas",
			Call: eth.EstimateGas(&w3types.Message{
				To:   w3.APtr("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2"),
				Func: funcBalanceOf,
				Args: []any{w3.A("0x000000000000000000000000000000000000c0Fe")},
			}, nil),
			WantRet: 23750,
		},
	}

	runTestCases(t, tests)
}

func TestAccessList(t *testing.T) {
	tests := []testCase[eth.AccessListResponse]{
		{
			Golden: "create_access_list",
			Call: eth.AccessList(&w3types.Message{
				To:   w3.APtr("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2"),
				Func: funcBalanceOf,
				Args: []any{w3.A("0x000000000000000000000000000000000000c0Fe")},
			}, nil),
			WantRet: eth.AccessListResponse{
				AccessList: types.AccessList{
					{
						Address: w3.A("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2"),
						StorageKeys: []common.Hash{
							w3.H("0xf68b260b81af177c0bf1a03b5d62b15aea1b486f8df26c77f33aed7538cfeb2c"),
						},
					},
				},
				GasUsed: 26050,
			},
		},
	}

	runTestCases(t, tests)
}
