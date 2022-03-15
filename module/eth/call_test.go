package eth_test

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/lmittmann/w3"
	"github.com/lmittmann/w3/module/eth"
	"github.com/lmittmann/w3/rpctest"
)

var (
	funcBalanceOf = w3.MustNewFunc("balanceOf(address)", "uint256")
)

func TestCallFunc(t *testing.T) {
	t.Parallel()

	srv := rpctest.NewFileServer(t, "testdata/call_func.golden")
	defer srv.Close()

	client := w3.MustDial(srv.URL())
	defer client.Close()

	var (
		balance     = new(big.Int)
		wantBalance = big.NewInt(0)
	)
	if err := client.Call(
		eth.CallFunc(funcBalanceOf, w3.A("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2"), w3.A("0x000000000000000000000000000000000000c0Fe")).Returns(balance),
	); err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	if wantBalance.Cmp(balance) != 0 {
		t.Fatalf("want %v, got %v", wantBalance, balance)
	}
}

func TestCallFunc_Overrides(t *testing.T) {
	t.Parallel()

	srv := rpctest.NewFileServer(t, "testdata/call_func_overrides.golden")
	defer srv.Close()

	client := w3.MustDial(srv.URL())
	defer client.Close()

	var (
		balance   = new(big.Int)
		overrides = eth.AccountOverrides{
			w3.A("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2"): eth.Account{
				StateDiff: map[common.Hash]common.Hash{
					w3.H("0xf68b260b81af177c0bf1a03b5d62b15aea1b486f8df26c77f33aed7538cfeb2c"): w3.H("0x000000000000000000000000000000000000000000000000000000000000002a"),
				},
			},
		}
		wantBalance = big.NewInt(42)
	)
	if err := client.Call(
		eth.CallFunc(funcBalanceOf, w3.A("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2"), w3.A("0x000000000000000000000000000000000000c0Fe")).Overrides(overrides).Returns(balance),
	); err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	if wantBalance.Cmp(balance) != 0 {
		t.Fatalf("want %v, got %v", wantBalance, balance)
	}
}
