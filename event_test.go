package w3

import (
	"math/big"
	"strconv"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestNewEvent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		Signature string
		WantEvent *Event
	}{
		{
			Signature: "Transfer(address,address,uint256)",
			WantEvent: &Event{
				Signature: "Transfer(address,address,uint256)",
				Topic0:    H("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"),
			},
		},
		{
			Signature: "Transfer(address from, address to, uint256 value)",
			WantEvent: &Event{
				Signature: "Transfer(address,address,uint256)",
				Topic0:    H("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"),
			},
		},
		{
			Signature: "Approval(address,address,uint256)",
			WantEvent: &Event{
				Signature: "Approval(address,address,uint256)",
				Topic0:    H("0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925"),
			},
		},
		{
			Signature: "Approval(address owner, address spender, uint256 value)",
			WantEvent: &Event{
				Signature: "Approval(address,address,uint256)",
				Topic0:    H("0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925"),
			},
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			gotEvent, err := NewEvent(test.Signature)
			if err != nil {
				t.Fatalf("Failed to create new FUnc: %v", err)
			}

			if diff := cmp.Diff(test.WantEvent, gotEvent,
				cmpopts.IgnoreFields(Event{}, "Args"),
			); diff != "" {
				t.Fatalf("(-want, +got)\n%s", diff)
			}
		})
	}
}

func TestEventDecodeArgs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		Event    *Event
		Log      *types.Log
		Args     []any
		WantArgs []any
	}{
		{
			Event: MustNewEvent("Transfer(address,address,uint256)"),
			Log: &types.Log{
				Topics: []common.Hash{
					H("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"),
				},
				Data: B("0x" +
					"000000000000000000000000000000000000000000000000000000000000c0fe" +
					"000000000000000000000000000000000000000000000000000000000000dead" +
					"000000000000000000000000000000000000000000000000000000000000002a"),
			},
			Args: []any{new(common.Address), new(common.Address), new(big.Int)},
			WantArgs: []any{
				APtr("0x000000000000000000000000000000000000c0Fe"),
				APtr("0x000000000000000000000000000000000000dEaD"),
				big.NewInt(42),
			},
		},
		{
			Event: MustNewEvent("Transfer(address,address,uint256)"),
			Log: &types.Log{
				Topics: []common.Hash{
					H("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"),
					H("0x000000000000000000000000000000000000000000000000000000000000c0fe"),
				},
				Data: B("0x" +
					"000000000000000000000000000000000000000000000000000000000000dead" +
					"000000000000000000000000000000000000000000000000000000000000002a"),
			},
			Args: []any{new(common.Address), new(common.Address), new(big.Int)},
			WantArgs: []any{
				APtr("0x000000000000000000000000000000000000c0Fe"),
				APtr("0x000000000000000000000000000000000000dEaD"),
				big.NewInt(42),
			},
		},
		{
			Event: MustNewEvent("Transfer(address,address,uint256)"),
			Log: &types.Log{
				Topics: []common.Hash{
					H("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"),
					H("0x000000000000000000000000000000000000000000000000000000000000c0fe"),
					H("0x000000000000000000000000000000000000000000000000000000000000dead"),
				},
				Data: B("0x000000000000000000000000000000000000000000000000000000000000002a"),
			},
			Args: []any{new(common.Address), new(common.Address), new(big.Int)},
			WantArgs: []any{
				APtr("0x000000000000000000000000000000000000c0Fe"),
				APtr("0x000000000000000000000000000000000000dEaD"),
				big.NewInt(42),
			},
		},
		{
			Event: MustNewEvent("Transfer(address,address,uint256)"),
			Log: &types.Log{
				Topics: []common.Hash{
					H("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"),
					H("0x000000000000000000000000000000000000000000000000000000000000c0fe"),
					H("0x000000000000000000000000000000000000000000000000000000000000dead"),
					H("0x000000000000000000000000000000000000000000000000000000000000002a"),
				},
			},
			Args: []any{new(common.Address), new(common.Address), new(big.Int)},
			WantArgs: []any{
				APtr("0x000000000000000000000000000000000000c0Fe"),
				APtr("0x000000000000000000000000000000000000dEaD"),
				big.NewInt(42),
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
			); diff != "" {
				t.Fatalf("(-want, +got)\n%s", diff)
			}
		})
	}
}
