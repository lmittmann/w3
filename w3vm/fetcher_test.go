package w3vm

import (
	"errors"
	"strconv"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/google/go-cmp/cmp"
	"github.com/holiman/uint256"
	"github.com/lmittmann/w3/internal"
	w3hexutil "github.com/lmittmann/w3/internal/hexutil"
)

func TestTestdataContractsMerge(t *testing.T) {
	tests := []struct {
		Contracts testdataContracts
		Other     testdataContracts
		WantErr   error
		WantLen   int
	}{
		{
			Contracts: testdataContracts{},
			Other:     testdataContracts{},
			WantLen:   0,
		},
		{
			Contracts: testdataContracts{},
			Other: testdataContracts{
				common.Hash{0x11}: []byte("code1"),
				common.Hash{0x22}: []byte("code2"),
			},
			WantLen: 2,
		},
		{
			Contracts: testdataContracts{
				common.Hash{0x11}: []byte("code1"),
			},
			Other: testdataContracts{
				common.Hash{0x22}: []byte("code2"),
			},
			WantLen: 2,
		},
		{
			Contracts: testdataContracts{
				common.Hash{0x11}: []byte("code1"),
			},
			Other: testdataContracts{
				common.Hash{0x11}: []byte("code1"),
			},
			WantLen: 1,
		},
		{
			Contracts: testdataContracts{
				common.Hash{0x11}: []byte("code1"),
			},
			Other: testdataContracts{
				common.Hash{0x11}: []byte("different_code"),
			},
			WantErr: errors.New("bytecode conflict for code hash 0x1100000000000000000000000000000000000000000000000000000000000000"),
			WantLen: 1,
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			err := test.Contracts.Merge(test.Other)
			if diff := cmp.Diff(test.WantErr, err, internal.EquateErrors()); diff != "" {
				t.Fatalf("Err: (-want +got):\n%s", diff)
			}
			if len(test.Contracts) != test.WantLen {
				t.Fatalf("Len: want %d, got %d", test.WantLen, len(test.Contracts))
			}
		})
	}
}

func TestTestdataHeaderHashesMerge(t *testing.T) {
	tests := []struct {
		HeaderHashes testdataHeaderHashes
		Other        testdataHeaderHashes
		WantErr      error
		WantLen      int
	}{
		{
			HeaderHashes: testdataHeaderHashes{},
			Other:        testdataHeaderHashes{},
			WantLen:      0,
		},
		{
			HeaderHashes: testdataHeaderHashes{},
			Other: testdataHeaderHashes{
				1: common.Hash{0x11},
				2: common.Hash{0x22},
			},
			WantLen: 2,
		},
		{
			HeaderHashes: testdataHeaderHashes{
				1: common.Hash{0x11},
			},
			Other: testdataHeaderHashes{
				2: common.Hash{0x22},
			},
			WantLen: 2,
		},
		{
			HeaderHashes: testdataHeaderHashes{
				1: common.Hash{0x11},
			},
			Other: testdataHeaderHashes{
				1: common.Hash{0x11},
			},
			WantLen: 1,
		},
		{
			HeaderHashes: testdataHeaderHashes{
				1: common.Hash{0x11},
			},
			Other: testdataHeaderHashes{
				1: common.Hash{0x22},
			},
			WantErr: errors.New("header hash conflict for block 1"),
			WantLen: 1,
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			err := test.HeaderHashes.Merge(test.Other)
			if diff := cmp.Diff(test.WantErr, err, internal.EquateErrors()); diff != "" {
				t.Fatalf("Err: (-want +got):\n%s", diff)
			}
			if len(test.HeaderHashes) != test.WantLen {
				t.Fatalf("Len: want %d, got %d", test.WantLen, len(test.HeaderHashes))
			}
		})
	}
}

func TestTestdataAccountMerge(t *testing.T) {
	tests := []struct {
		Account *testdataAccount
		Other   *testdataAccount
		WantErr error
	}{
		{
			Account: &testdataAccount{
				Nonce:    1,
				Balance:  (*hexutil.U256)(uint256.NewInt(1)),
				CodeHash: common.Hash{0x11},
				Storage: map[w3hexutil.Hash]w3hexutil.Hash{
					{0x01}: {0xaa},
				},
			},
			Other: &testdataAccount{
				Nonce:    1,
				Balance:  (*hexutil.U256)(uint256.NewInt(1)),
				CodeHash: common.Hash{0x11},
				Storage: map[w3hexutil.Hash]w3hexutil.Hash{
					{0x02}: {0xbb},
				},
			},
		},
		{
			Account: &testdataAccount{
				Nonce:    1,
				Balance:  (*hexutil.U256)(uint256.NewInt(1)),
				CodeHash: common.Hash{0x11},
			},
			Other: &testdataAccount{
				Nonce:    2,
				Balance:  (*hexutil.U256)(uint256.NewInt(1)),
				CodeHash: common.Hash{0x11},
			},
			WantErr: errors.New("nonce conflict: 1 != 2"),
		},
		{
			Account: &testdataAccount{
				Nonce:    1,
				Balance:  (*hexutil.U256)(uint256.NewInt(1)),
				CodeHash: common.Hash{0x11},
			},
			Other: &testdataAccount{
				Nonce:    1,
				Balance:  (*hexutil.U256)(uint256.NewInt(2)),
				CodeHash: common.Hash{0x11},
			},
			WantErr: errors.New("balance conflict: 0x1 != 0x2"),
		},
		{
			Account: &testdataAccount{
				Nonce:    1,
				Balance:  (*hexutil.U256)(uint256.NewInt(1)),
				CodeHash: common.Hash{0x11},
			},
			Other: &testdataAccount{
				Nonce:    1,
				Balance:  (*hexutil.U256)(uint256.NewInt(1)),
				CodeHash: common.Hash{0x22},
			},
			WantErr: errors.New("code hash conflict: 0x1100000000000000000000000000000000000000000000000000000000000000 != 0x2200000000000000000000000000000000000000000000000000000000000000"),
		},
		{
			Account: &testdataAccount{
				Nonce:    1,
				Balance:  (*hexutil.U256)(uint256.NewInt(1)),
				CodeHash: common.Hash{0x11},
				Storage: map[w3hexutil.Hash]w3hexutil.Hash{
					{0x01}: {0xaa},
				},
			},
			Other: &testdataAccount{
				Nonce:    1,
				Balance:  (*hexutil.U256)(uint256.NewInt(1)),
				CodeHash: common.Hash{0x11},
				Storage: map[w3hexutil.Hash]w3hexutil.Hash{
					{0x01}: {0xbb},
				},
			},
			WantErr: errors.New("storage conflict at slot 0x0100000000000000000000000000000000000000000000000000000000000000: 0xaa00000000000000000000000000000000000000000000000000000000000000 != 0xbb00000000000000000000000000000000000000000000000000000000000000"),
		},
		{
			Account: &testdataAccount{
				Nonce:    1,
				Balance:  (*hexutil.U256)(uint256.NewInt(1)),
				CodeHash: common.Hash{0x11},
				Storage:  nil,
			},
			Other: &testdataAccount{
				Nonce:    1,
				Balance:  (*hexutil.U256)(uint256.NewInt(1)),
				CodeHash: common.Hash{0x11},
				Storage: map[w3hexutil.Hash]w3hexutil.Hash{
					{0x01}: {0xaa},
				},
			},
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			err := test.Account.Merge(test.Other)
			if diff := cmp.Diff(test.WantErr, err, internal.EquateErrors()); diff != "" {
				t.Fatalf("Err: (-want +got):\n%s", diff)
			}

			if test.WantErr != nil {
				return
			}
			if test.Account.Storage != nil && test.Other.Storage != nil {
				for slot, value := range test.Other.Storage {
					if test.Account.Storage[slot] != value {
						t.Fatalf("Storage slot %s not properly merged", common.Hash(slot))
					}
				}
			}
		})
	}
}

func TestTestdataStateMerge(t *testing.T) {
	addr1 := common.Address{0x11}
	addr2 := common.Address{0x22}

	tests := []struct {
		State   testdataState
		Other   testdataState
		WantErr error
		WantLen int
	}{
		{
			State:   testdataState{},
			Other:   testdataState{},
			WantLen: 0,
		},
		{
			State: testdataState{},
			Other: testdataState{
				addr1: &testdataAccount{
					Nonce:    hexutil.Uint64(1),
					Balance:  (*hexutil.U256)(uint256.NewInt(1)),
					CodeHash: common.Hash{0x11},
				},
			},
			WantLen: 1,
		},
		{
			State: testdataState{
				addr1: &testdataAccount{
					Nonce:    hexutil.Uint64(1),
					Balance:  (*hexutil.U256)(uint256.NewInt(1)),
					CodeHash: common.Hash{0x11},
				},
			},
			Other: testdataState{
				addr2: &testdataAccount{
					Nonce:    hexutil.Uint64(2),
					Balance:  (*hexutil.U256)(uint256.NewInt(1)),
					CodeHash: common.Hash{0x22},
				},
			},
			WantLen: 2,
		},
		{
			State: testdataState{
				addr1: &testdataAccount{
					Nonce:    hexutil.Uint64(1),
					Balance:  (*hexutil.U256)(uint256.NewInt(1)),
					CodeHash: common.Hash{0x11},
					Storage: map[w3hexutil.Hash]w3hexutil.Hash{
						{0x01}: {0xaa},
					},
				},
			},
			Other: testdataState{
				addr1: &testdataAccount{
					Nonce:    hexutil.Uint64(1),
					Balance:  (*hexutil.U256)(uint256.NewInt(1)),
					CodeHash: common.Hash{0x11},
					Storage: map[w3hexutil.Hash]w3hexutil.Hash{
						{0x02}: {0xbb},
					},
				},
			},
			WantLen: 1,
		},
		{
			State: testdataState{
				addr1: &testdataAccount{
					Nonce:    hexutil.Uint64(1),
					Balance:  (*hexutil.U256)(uint256.NewInt(1)),
					CodeHash: common.Hash{0x11},
				},
			},
			Other: testdataState{
				addr1: &testdataAccount{
					Nonce:    hexutil.Uint64(2),
					Balance:  (*hexutil.U256)(uint256.NewInt(1)),
					CodeHash: common.Hash{0x11},
				},
			},
			WantErr: errors.New("account conflict for address 0x1100000000000000000000000000000000000000: nonce conflict: 1 != 2"),
			WantLen: 1,
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			err := test.State.Merge(test.Other)
			if diff := cmp.Diff(test.WantErr, err, internal.EquateErrors()); diff != "" {
				t.Fatalf("Err: (-want +got):\n%s", diff)
			}
			if len(test.State) != test.WantLen {
				t.Fatalf("Len: want %d, got %d", test.WantLen, len(test.State))
			}
		})
	}
}
