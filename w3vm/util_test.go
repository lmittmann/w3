package w3vm_test

import (
	"strconv"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/lmittmann/w3"
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
