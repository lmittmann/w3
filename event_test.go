package w3_test

import (
	"fmt"
	"math/big"
	"strconv"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/lmittmann/w3"
)

func ExampleEvent_DecodeArgs() {
	var (
		eventTransfer = w3.MustNewEvent("Transfer(address indexed from, address indexed to, uint256 value)")
		log           = &types.Log{
			Address: w3.A("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2"),
			Topics: []common.Hash{
				w3.H("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"),
				w3.H("0x000000000000000000000000000000000000000000000000000000000000c0fe"),
				w3.H("0x000000000000000000000000000000000000000000000000000000000000dead"),
			},
			Data: w3.B("0x0000000000000000000000000000000000000000000000001111d67bb1bb0000"),
		}

		from  common.Address
		to    common.Address
		value big.Int
	)

	if err := eventTransfer.DecodeArgs(log, &from, &to, &value); err != nil {
		fmt.Printf("Failed to decode event log: %v\n", err)
		return
	}
	fmt.Printf("Transferred %s WETH9 from %s to %s", w3.FromWei(&value, 18), from, to)
	// Output:
	// Transferred 1.23 WETH9 from 0x000000000000000000000000000000000000c0Fe to 0x000000000000000000000000000000000000dEaD
}

func TestNewEvent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		Signature string
		WantEvent *w3.Event
	}{
		{
			Signature: "Transfer(address,address,uint256)",
			WantEvent: &w3.Event{
				Signature: "Transfer(address,address,uint256)",
				Topic0:    w3.H("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"),
			},
		},
		{
			Signature: "Transfer(address from, address to, uint256 value)",
			WantEvent: &w3.Event{
				Signature: "Transfer(address,address,uint256)",
				Topic0:    w3.H("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"),
			},
		},
		{
			Signature: "Approval(address,address,uint256)",
			WantEvent: &w3.Event{
				Signature: "Approval(address,address,uint256)",
				Topic0:    w3.H("0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925"),
			},
		},
		{
			Signature: "Approval(address owner, address spender, uint256 value)",
			WantEvent: &w3.Event{
				Signature: "Approval(address,address,uint256)",
				Topic0:    w3.H("0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925"),
			},
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			gotEvent, err := w3.NewEvent(test.Signature)
			if err != nil {
				t.Fatalf("Failed to create new FUnc: %v", err)
			}

			if diff := cmp.Diff(test.WantEvent, gotEvent,
				cmpopts.IgnoreUnexported(w3.Event{}),
				cmpopts.IgnoreFields(w3.Event{}, "Args"),
			); diff != "" {
				t.Fatalf("(-want, +got)\n%s", diff)
			}
		})
	}
}

func TestEventDecodeArgs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		Event    *w3.Event
		Log      *types.Log
		Args     []any
		WantArgs []any
	}{
		{
			Event: w3.MustNewEvent("Transfer(address,address,uint256)"),
			Log: &types.Log{
				Topics: []common.Hash{
					w3.H("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"),
				},
				Data: w3.B("0x" +
					"000000000000000000000000000000000000000000000000000000000000c0fe" +
					"000000000000000000000000000000000000000000000000000000000000dead" +
					"000000000000000000000000000000000000000000000000000000000000002a"),
			},
			Args: []any{new(common.Address), new(common.Address), new(big.Int)},
			WantArgs: []any{
				w3.APtr("0x000000000000000000000000000000000000c0Fe"),
				w3.APtr("0x000000000000000000000000000000000000dEaD"),
				big.NewInt(42),
			},
		},
		{
			Event: w3.MustNewEvent("Transfer(address indexed, address, uint256)"),
			Log: &types.Log{
				Topics: []common.Hash{
					w3.H("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"),
					w3.H("0x000000000000000000000000000000000000000000000000000000000000c0fe"),
				},
				Data: w3.B("0x" +
					"000000000000000000000000000000000000000000000000000000000000dead" +
					"000000000000000000000000000000000000000000000000000000000000002a"),
			},
			Args: []any{new(common.Address), new(common.Address), new(big.Int)},
			WantArgs: []any{
				w3.APtr("0x000000000000000000000000000000000000c0Fe"),
				w3.APtr("0x000000000000000000000000000000000000dEaD"),
				big.NewInt(42),
			},
		},
		{
			Event: w3.MustNewEvent("Transfer(address, address indexed, uint256)"),
			Log: &types.Log{
				Topics: []common.Hash{
					w3.H("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"),
					w3.H("0x000000000000000000000000000000000000000000000000000000000000dead"),
				},
				Data: w3.B("0x" +
					"000000000000000000000000000000000000000000000000000000000000c0fe" +
					"000000000000000000000000000000000000000000000000000000000000002a"),
			},
			Args: []any{new(common.Address), new(common.Address), new(big.Int)},
			WantArgs: []any{
				w3.APtr("0x000000000000000000000000000000000000c0Fe"),
				w3.APtr("0x000000000000000000000000000000000000dEaD"),
				big.NewInt(42),
			},
		},
		{
			Event: w3.MustNewEvent("Transfer(address indexed, address indexed, uint256)"),
			Log: &types.Log{
				Topics: []common.Hash{
					w3.H("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"),
					w3.H("0x000000000000000000000000000000000000000000000000000000000000c0fe"),
					w3.H("0x000000000000000000000000000000000000000000000000000000000000dead"),
				},
				Data: w3.B("0x000000000000000000000000000000000000000000000000000000000000002a"),
			},
			Args: []any{new(common.Address), new(common.Address), new(big.Int)},
			WantArgs: []any{
				w3.APtr("0x000000000000000000000000000000000000c0Fe"),
				w3.APtr("0x000000000000000000000000000000000000dEaD"),
				big.NewInt(42),
			},
		},
		{
			Event: w3.MustNewEvent("Transfer(address indexed, address indexed, uint256 indexed)"),
			Log: &types.Log{
				Topics: []common.Hash{
					w3.H("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"),
					w3.H("0x000000000000000000000000000000000000000000000000000000000000c0fe"),
					w3.H("0x000000000000000000000000000000000000000000000000000000000000dead"),
					w3.H("0x000000000000000000000000000000000000000000000000000000000000002a"),
				},
			},
			Args: []any{new(common.Address), new(common.Address), new(big.Int)},
			WantArgs: []any{
				w3.APtr("0x000000000000000000000000000000000000c0Fe"),
				w3.APtr("0x000000000000000000000000000000000000dEaD"),
				big.NewInt(42),
			},
		},
		{ // https://github.com/lmittmann/w3/issues/15
			Event: w3.MustNewEvent("NameRegistered(string name, bytes32 indexed label, address indexed owner, uint cost, uint expires)"),
			Log: &types.Log{
				Address: w3.A("0x283Af0B28c62C092C9727F1Ee09c02CA627EB7F5"),
				Topics: []common.Hash{
					w3.H("0xca6abbe9d7f11422cb6ca7629fbf6fe9efb1c621f71ce8f02b9f2a230097404f"),
					w3.H("0x4e59ffc7ae105a2b19f7f29b63e9f9c5ac28e27bce744a330804c6a89269cec0"),
					w3.H("0x000000000000000000000000bd08f39b2523426cc1d6961e2d6a9744b3b432b5"),
				},
				Data: w3.B("0x000000000000000000000000000000000000000000000000000000000000006000000000000000000000000000000000000000000000000000165a7d3a3e48ad0000000000000000000000000000000000000000000000000000000066e4d0a50000000000000000000000000000000000000000000000000000000000000009656c666172657373690000000000000000000000000000000000000000000000"),
			},
			Args: []any{new(string), new(common.Hash), new(common.Address), new(big.Int), new(big.Int)},
			WantArgs: []any{
				ptr("elfaressi"),
				ptr(w3.H("0x4e59ffc7ae105a2b19f7f29b63e9f9c5ac28e27bce744a330804c6a89269cec0")),
				w3.APtr("0xbD08F39B2523426Cc1d6961e2d6A9744B3B432b5"),
				big.NewInt(6291943382206637),
				big.NewInt(1726271653),
			},
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			if err := test.Event.DecodeArgs(test.Log, test.Args...); err != nil {
				t.Fatalf("Failed to decode args: %v", err)
			}
			if diff := cmp.Diff(test.WantArgs, test.Args,
				cmp.AllowUnexported(big.Int{}),
				cmpopts.IgnoreUnexported(w3.Event{}),
			); diff != "" {
				t.Fatalf("(-want, +got)\n%s", diff)
			}
		})
	}
}
