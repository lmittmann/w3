package state_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/holiman/uint256"
	"github.com/lmittmann/w3/w3vm/state"
)

func TestTestAccountMarshaling(t *testing.T) {
	acc := &state.Account{
		Nonce:   1,
		Balance: *uint256.NewInt(1),
		Code:    []byte{0xc0, 0xfe},
		Storage: map[uint256.Int]uint256.Int{
			*uint256.NewInt(0): *uint256.NewInt(1),
		},
	}
	enc := []byte(`{"nonce":"0x1","balance":"0x1","code":"0xc0fe","storage":{"0x0":"0x1"}}`)

	t.Run("MarshalJSON", func(t *testing.T) {
		got, err := json.Marshal(acc)
		if err != nil {
			t.Fatalf("Failed to marshal account: %v", err)
		}
		want := enc
		if !bytes.Equal(want, got) {
			t.Fatalf("(-want +got):\n- %s\n+ %s", want, got)
		}
	})

	t.Run("UnmarshalJSON", func(t *testing.T) {
		var got state.Account
		err := json.Unmarshal(enc, &got)
		if err != nil {
			t.Fatalf("Failed to unmarshal account: %v", err)
		}
		want := acc
		if diff := cmp.Diff(want, &got); diff != "" {
			t.Fatalf("(-want +got):\n%s", diff)
		}
	})

}