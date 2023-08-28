package w3vm

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/holiman/uint256"
)

var (
	uint1 = *uint256.NewInt(1)
	uint2 = *uint256.NewInt(2)
)

func TestReadTestdataState(t *testing.T) {
	t.Run("read non-existent", func(t *testing.T) {
		dir := t.TempDir()
		fp := filepath.Join(dir, "1_0.json")

		gotState, err := readTestdataState(fp)
		if err != nil {
			t.Fatalf("Failed to read state: %v", err)
		}

		wantState := new(forkState)
		if diff := cmp.Diff(wantState, gotState); diff != "" {
			t.Fatalf("(-want +got):\n%s", diff)
		}
	})

	t.Run("read", func(t *testing.T) {
		dir := t.TempDir()
		fp := filepath.Join(dir, "1_0.json")

		stateContent := []byte(`{"accounts":{"0x0100000000000000000000000000000000000000":{"balance":"0x1"}}}`)
		if err := os.WriteFile(fp, stateContent, 0644); err != nil {
			t.Fatalf("Failed to create state file: %v", err)

		}

		gotState, err := readTestdataState(fp)
		if err != nil {
			t.Fatalf("Failed to read state: %q", err)
		}

		wantState := &forkState{
			Accounts: map[common.Address]*account{
				{0x1}: {Balance: uint1},
			},
		}
		if diff := cmp.Diff(wantState, gotState,
			cmpopts.EquateEmpty(),
		); diff != "" {
			t.Fatalf("(-want +got):\n%s", diff)
		}
	})
}

func TestWriteTestdataState(t *testing.T) {
	t.Run("write non-existent", func(t *testing.T) {
		dir := t.TempDir()
		fp := filepath.Join(dir, "1_0.json")
		wantState := &forkState{
			Accounts: map[common.Address]*account{
				{0x1}: {Balance: uint1},
			},
		}

		if err := writeTestdataState(fp, wantState); err != nil {
			t.Fatalf("Failed to write state: %v", err)
		}

		gotState, err := readTestdataState(fp)
		if err != nil {
			t.Fatalf("Failed to read state: %q", err)
		}
		if diff := cmp.Diff(wantState, gotState,
			cmpopts.EquateEmpty(),
		); diff != "" {
			t.Fatalf("(-want +got):\n%s", diff)
		}
	})

	t.Run("write", func(t *testing.T) {
		dir := t.TempDir()
		fp := filepath.Join(dir, "1_0.json")
		preState := &forkState{
			Accounts: map[common.Address]*account{
				{0x1}: {Balance: uint1},
			},
		}
		newState := &forkState{
			Accounts: map[common.Address]*account{
				{0x2}: {Balance: uint2},
			},
		}
		wantState := &forkState{
			Accounts: map[common.Address]*account{
				{0x1}: {Balance: uint1},
				{0x2}: {Balance: uint2},
			},
		}

		if err := writeTestdataState(fp, preState); err != nil {
			t.Fatalf("Failed to write pre-state: %v", err)
		}
		if err := writeTestdataState(fp, newState); err != nil {
			t.Fatalf("Failed to write new-state: %v", err)
		}

		gotState, err := readTestdataState(fp)
		if err != nil {
			t.Fatalf("Failed to read state: %q", err)
		}
		if diff := cmp.Diff(wantState, gotState,
			cmpopts.EquateEmpty(),
		); diff != "" {
			t.Fatalf("(-want +got):\n%s", diff)
		}
	})
}

func TestForkStateMerge(t *testing.T) {
	tests := []struct {
		S1 *forkState
		S2 *forkState

		Want        *forkState
		WantChanged bool
	}{
		{
			S1:          &forkState{},
			S2:          &forkState{},
			Want:        &forkState{},
			WantChanged: false,
		},
		{
			S1:          &forkState{HeaderHashes: map[hexutil.Uint64]common.Hash{1: {0x1}}},
			S2:          &forkState{},
			Want:        &forkState{HeaderHashes: map[hexutil.Uint64]common.Hash{1: {0x1}}},
			WantChanged: false,
		},
		{
			S1:          &forkState{},
			S2:          &forkState{HeaderHashes: map[hexutil.Uint64]common.Hash{1: {0x1}}},
			Want:        &forkState{HeaderHashes: map[hexutil.Uint64]common.Hash{1: {0x1}}},
			WantChanged: true,
		},
		{
			S1:          &forkState{HeaderHashes: map[hexutil.Uint64]common.Hash{1: {0x1}}},
			S2:          &forkState{HeaderHashes: map[hexutil.Uint64]common.Hash{1: {0x1}}},
			Want:        &forkState{HeaderHashes: map[hexutil.Uint64]common.Hash{1: {0x1}}},
			WantChanged: false,
		},
		{
			S1:          &forkState{HeaderHashes: map[hexutil.Uint64]common.Hash{1: {0x1}}},
			S2:          &forkState{HeaderHashes: map[hexutil.Uint64]common.Hash{2: {0x2}}},
			Want:        &forkState{HeaderHashes: map[hexutil.Uint64]common.Hash{1: {0x1}, 2: {0x2}}},
			WantChanged: true,
		},
		{ // If the same key is present in both states, the value of S1 is NOT changed.
			S1:          &forkState{Accounts: map[common.Address]*account{{0x1}: {Balance: uint1}}},
			S2:          &forkState{Accounts: map[common.Address]*account{{0x1}: {Balance: uint2}}},
			Want:        &forkState{Accounts: map[common.Address]*account{{0x1}: {Balance: uint1}}},
			WantChanged: false,
		},
		{
			S1:          &forkState{Accounts: map[common.Address]*account{{0x1}: {}}},
			S2:          &forkState{Accounts: map[common.Address]*account{{0x1}: {Storage: map[uint256.Int]uint256.Int{uint1: uint1}}}},
			Want:        &forkState{Accounts: map[common.Address]*account{{0x1}: {Storage: map[uint256.Int]uint256.Int{uint1: uint1}}}},
			WantChanged: true,
		},
		{
			S1:          &forkState{Accounts: map[common.Address]*account{{0x1}: {Storage: map[uint256.Int]uint256.Int{uint1: uint1}}}},
			S2:          &forkState{Accounts: map[common.Address]*account{{0x1}: {}}},
			Want:        &forkState{Accounts: map[common.Address]*account{{0x1}: {Storage: map[uint256.Int]uint256.Int{uint1: uint1}}}},
			WantChanged: false,
		},
		{
			S1:          &forkState{Accounts: map[common.Address]*account{{0x1}: {Storage: map[uint256.Int]uint256.Int{}}}},
			S2:          &forkState{Accounts: map[common.Address]*account{{0x1}: {Storage: map[uint256.Int]uint256.Int{uint1: uint1}}}},
			Want:        &forkState{Accounts: map[common.Address]*account{{0x1}: {Storage: map[uint256.Int]uint256.Int{uint1: uint1}}}},
			WantChanged: true,
		},
		{
			S1:          &forkState{Accounts: map[common.Address]*account{{0x1}: {Storage: map[uint256.Int]uint256.Int{uint1: uint1}}}},
			S2:          &forkState{Accounts: map[common.Address]*account{{0x1}: {Storage: map[uint256.Int]uint256.Int{}}}},
			Want:        &forkState{Accounts: map[common.Address]*account{{0x1}: {Storage: map[uint256.Int]uint256.Int{uint1: uint1}}}},
			WantChanged: false,
		},
		{
			S1: &forkState{Accounts: map[common.Address]*account{{0x1}: {Storage: map[uint256.Int]uint256.Int{
				uint1: uint1,
			}}}},
			S2: &forkState{Accounts: map[common.Address]*account{{0x1}: {Storage: map[uint256.Int]uint256.Int{
				uint2: uint2,
			}}}},
			Want: &forkState{Accounts: map[common.Address]*account{{0x1}: {Storage: map[uint256.Int]uint256.Int{
				uint1: uint1,
				uint2: uint2,
			}}}},
			WantChanged: true,
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			gotChanged := test.S1.Merge(test.S2)
			if test.WantChanged != gotChanged {
				t.Errorf("Changed: want %v, got %v", test.WantChanged, gotChanged)
			}

			if diff := cmp.Diff(test.Want, test.S1); diff != "" {
				t.Errorf("(-want +got):\n%s", diff)
			}
		})
	}
}
