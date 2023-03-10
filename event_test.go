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
		{ // https://github.com/lmittmann/w3/issues/15
			Event: MustNewEvent("NameRegistered(string name, bytes32 indexed label, address indexed owner, uint cost, uint expires)"),
			Log: &types.Log{
				Address: A("0x283Af0B28c62C092C9727F1Ee09c02CA627EB7F5"),
				Topics: []common.Hash{
					H("0xca6abbe9d7f11422cb6ca7629fbf6fe9efb1c621f71ce8f02b9f2a230097404f"),
					H("0x4e59ffc7ae105a2b19f7f29b63e9f9c5ac28e27bce744a330804c6a89269cec0"),
					H("0x000000000000000000000000bd08f39b2523426cc1d6961e2d6a9744b3b432b5"),
				},
				Data: B("0x000000000000000000000000000000000000000000000000000000000000006000000000000000000000000000000000000000000000000000165a7d3a3e48ad0000000000000000000000000000000000000000000000000000000066e4d0a50000000000000000000000000000000000000000000000000000000000000009656c666172657373690000000000000000000000000000000000000000000000"),
			},
			Args: []any{new(string), new(common.Hash), new(common.Address), new(big.Int), new(big.Int)},
			WantArgs: []any{
				ptr("elfaressi"),
				ptr(H("0x4e59ffc7ae105a2b19f7f29b63e9f9c5ac28e27bce744a330804c6a89269cec0")),
				APtr("0xbD08F39B2523426Cc1d6961e2d6A9744B3B432b5"),
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
			); diff != "" {
				t.Fatalf("(-want, +got)\n%s", diff)
			}
		})
	}
}
