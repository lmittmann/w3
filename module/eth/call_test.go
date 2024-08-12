package eth_test

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/lmittmann/w3"
	"github.com/lmittmann/w3/module/eth"
	"github.com/lmittmann/w3/rpctest"
	"github.com/lmittmann/w3/w3types"
)

var funcBalanceOf = w3.MustNewFunc("balanceOf(address)", "uint256")

func TestCall(t *testing.T) {
	rpctest.RunTestCases(t, []rpctest.TestCase[[]byte]{
		{
			Golden: "call_func",
			Call: eth.Call(&w3types.Message{
				To:   w3.APtr("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2"),
				Func: funcBalanceOf,
				Args: []any{w3.A("0x000000000000000000000000000000000000c0Fe")},
			}, nil, nil),
			WantRet: make([]byte, 32),
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
	})
}

func TestCallFunc(t *testing.T) {
	srv := rpctest.NewFileServer(t, "testdata/call_func.golden")
	defer srv.Close()

	client := w3.MustDial(srv.URL())
	defer client.Close()

	var (
		balance     = new(big.Int)
		wantBalance = big.NewInt(0)
	)
	if err := client.Call(
		eth.CallFunc(w3.A("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2"), funcBalanceOf, w3.A("0x000000000000000000000000000000000000c0Fe")).Returns(balance),
	); err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	if wantBalance.Cmp(balance) != 0 {
		t.Fatalf("want %v, got %v", wantBalance, balance)
	}
}

func TestEstimateGas(t *testing.T) {
	rpctest.RunTestCases(t, []rpctest.TestCase[uint64]{
		{
			Golden: "estimate_gas",
			Call: eth.EstimateGas(&w3types.Message{
				To:   w3.APtr("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2"),
				Func: funcBalanceOf,
				Args: []any{w3.A("0x000000000000000000000000000000000000c0Fe")},
			}, nil),
			WantRet: 23750,
		},
	})
}

func TestAccessList(t *testing.T) {
	rpctest.RunTestCases(t, []rpctest.TestCase[*eth.AccessListResponse]{
		{
			Golden: "create_access_list",
			Call: eth.AccessList(&w3types.Message{
				To:   w3.APtr("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2"),
				Func: funcBalanceOf,
				Args: []any{w3.A("0x000000000000000000000000000000000000c0Fe")},
			}, nil),
			WantRet: &eth.AccessListResponse{
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
	})
}
