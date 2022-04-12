package eth_test

import (
	"math/big"
	"testing"

	"github.com/lmittmann/w3"
	"github.com/lmittmann/w3/module/eth"
	"github.com/lmittmann/w3/rpctest"
)

func TestBalance(t *testing.T) {
	t.Parallel()

	srv := rpctest.NewFileServer(t, "testdata/get_balance.golden")
	defer srv.Close()

	client := w3.MustDial(srv.URL())
	defer client.Close()

	var (
		balance     big.Int
		wantBalance = w3.I("1 ether")
	)
	if err := client.Call(
		eth.Balance(w3.A("0x000000000000000000000000000000000000c0Fe"), nil).Returns(&balance),
	); err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	if wantBalance.Cmp(&balance) != 0 {
		t.Fatalf("want %v, got %v", wantBalance, &balance)
	}
}

func TestBalance_AtBlock(t *testing.T) {
	t.Parallel()

	srv := rpctest.NewFileServer(t, "testdata/get_balance__at_block.golden")
	defer srv.Close()

	client := w3.MustDial(srv.URL())
	defer client.Close()

	var (
		balance     big.Int
		wantBalance = w3.I("0.1 ether")
	)
	if err := client.Call(
		eth.Balance(w3.A("0x000000000000000000000000000000000000c0Fe"), big.NewInt(255)).Returns(&balance),
	); err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	if wantBalance.Cmp(&balance) != 0 {
		t.Fatalf("want %v, got %v", wantBalance, balance)
	}
}
