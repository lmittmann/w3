package w3types_test

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/lmittmann/w3/w3types"
)

func TestStateMerge(t *testing.T) {
	tests := []struct {
		Name     string
		StateDst w3types.State
		StateSrc w3types.State
		Want     w3types.State
	}{
		{
			Name:     "empty",
			StateDst: w3types.State{},
			StateSrc: w3types.State{},
			Want:     w3types.State{},
		},
		{
			Name:     "empty-dst",
			StateDst: w3types.State{},
			StateSrc: w3types.State{common.Address{}: {}},
			Want:     w3types.State{common.Address{}: {}},
		},
		{
			Name:     "empty-src",
			StateDst: w3types.State{common.Address{}: {}},
			StateSrc: w3types.State{},
			Want:     w3types.State{common.Address{}: {}},
		},
		{
			Name:     "simple",
			StateDst: w3types.State{common.Address{0x01}: {}},
			StateSrc: w3types.State{common.Address{0x02}: {}},
			Want: w3types.State{
				common.Address{0x01}: {},
				common.Address{0x02}: {},
			},
		},
		{
			Name:     "simple-conflict",
			StateDst: w3types.State{common.Address{}: {Nonce: 1}},
			StateSrc: w3types.State{common.Address{}: {Nonce: 2}},
			Want:     w3types.State{common.Address{}: {Nonce: 2}},
		},
		{
			Name:     "storage-simple",
			StateDst: w3types.State{common.Address{}: {Storage: w3types.Storage{common.Hash{0x01}: common.Hash{0x01}}}},
			StateSrc: w3types.State{common.Address{}: {Storage: w3types.Storage{common.Hash{0x02}: common.Hash{0x02}}}},
			Want: w3types.State{common.Address{}: {Storage: w3types.Storage{
				common.Hash{0x01}: common.Hash{0x01},
				common.Hash{0x02}: common.Hash{0x02},
			}}},
		},
		{
			Name:     "storage-conflict",
			StateDst: w3types.State{common.Address{}: {Storage: w3types.Storage{common.Hash{}: common.Hash{0x01}}}},
			StateSrc: w3types.State{common.Address{}: {Storage: w3types.Storage{common.Hash{}: common.Hash{0x02}}}},
			Want:     w3types.State{common.Address{}: {Storage: w3types.Storage{common.Hash{}: common.Hash{0x02}}}},
		},

		// https://github.com/lmittmann/w3/pull/176
		{
			Name:     "empty-code",
			StateDst: w3types.State{common.Address{}: {Code: []byte{}}},
			StateSrc: w3types.State{},
			Want:     w3types.State{common.Address{}: {Code: []byte{}}},
		},
		{
			Name:     "empty-code2",
			StateDst: w3types.State{},
			StateSrc: w3types.State{common.Address{}: {Code: []byte{}}},
			Want:     w3types.State{common.Address{}: {Code: []byte{}}},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			got := test.StateDst.Merge(test.StateSrc)
			if diff := cmp.Diff(test.Want, got,
				cmpopts.IgnoreUnexported(w3types.Account{}),
			); diff != "" {
				t.Fatalf("(-want, +got)\n%s", diff)
			}
		})
	}
}
