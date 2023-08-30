package w3vm_test

import (
	"fmt"
	"math/big"
	"strconv"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/lmittmann/w3"
	"github.com/lmittmann/w3/w3types"
	"github.com/lmittmann/w3/w3vm"
)

func TestWETHBalanceSlot(t *testing.T) {
	tests := []struct {
		Addr     common.Address
		WantSlot common.Hash
	}{
		{
			Addr:     w3.A("0x000000000000000000000000000000000000dEaD"),
			WantSlot: w3.H("0x262bb27bbdd95c1cdc8e16957e36e38579ea44f7f6413dd7a9c75939def06b2c"),
		},
		{
			Addr:     w3.A("0x000000000000000000000000000000000000c0Fe"),
			WantSlot: w3.H("0xf68b260b81af177c0bf1a03b5d62b15aea1b486f8df26c77f33aed7538cfeb2c"),
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			gotSlot := w3vm.WETHBalanceSlot(test.Addr)
			if test.WantSlot != gotSlot {
				t.Fatalf("want %s, got %s", test.WantSlot, gotSlot)
			}
		})
	}
}

func ExampleWETHBalanceSlot() {
	client := w3.MustDial("https://rpc.ankr.com/eth")
	defer client.Close()

	addrC0fe := w3.A("0x000000000000000000000000000000000000c0Fe")
	addrWETH := w3.A("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2")
	funcBalanceOf := w3.MustNewFunc("balanceOf(address)", "uint256")

	vm, err := w3vm.New(
		w3vm.WithFork(client, nil),
		w3vm.WithState(w3types.State{
			addrWETH: {
				Storage: map[common.Hash]common.Hash{
					w3vm.WETHBalanceSlot(addrC0fe): common.BigToHash(w3.I("100 ether")),
				},
			},
		}),
	)
	if err != nil {
		// ...
	}

	var balance *big.Int
	err = vm.CallFunc(addrWETH, funcBalanceOf, addrC0fe).Returns(&balance)
	if err != nil {
		// ...
	}
	fmt.Printf("%s: %s WETH", addrC0fe, w3.FromWei(balance, 18))
	// Output:
	// 0x000000000000000000000000000000000000c0Fe: 100 WETH
}

func TestWETHAllowanceSlot(t *testing.T) {
	tests := []struct {
		Owner    common.Address
		Spender  common.Address
		WantSlot common.Hash
	}{
		{
			Owner:    w3.A("0x000000000000000000000000000000000000dEaD"),
			Spender:  w3.A("0x000000000000000000000000000000000000c0Fe"),
			WantSlot: w3.H("0xea3c5e9cf6f5b7aba5d41ca731cf1f8cb1373e841a1cb336cc4bfeddc27c7f8b"),
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			gotSlot := w3vm.WETHAllowanceSlot(test.Owner, test.Spender)
			if test.WantSlot != gotSlot {
				t.Fatalf("want %s, got %s", test.WantSlot, gotSlot)
			}
		})
	}
}
