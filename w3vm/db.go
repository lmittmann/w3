package w3vm

import (
	"errors"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/state/snapshot"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/ethereum/go-ethereum/trie/trienode"
	"github.com/ethereum/go-ethereum/trie/utils"
	"github.com/ethereum/go-ethereum/triedb"
	"github.com/holiman/uint256"
)

var fakeTrieDB = triedb.NewDatabase(rawdb.NewMemoryDatabase(), &triedb.Config{})

// db implements the [state.Reader], [state.Database], and [state.Trie] interfaces.
type db struct {
	fetcher Fetcher
}

func newDB(fetcher Fetcher) *db {
	return &db{
		fetcher: fetcher,
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////
// state.Reader methods ////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////////////////

func (db *db) Account(addr common.Address) (*types.StateAccount, error) {
	if db.fetcher == nil {
		return &types.StateAccount{
			Balance:  new(uint256.Int),
			CodeHash: types.EmptyCodeHash[:],
		}, nil
	}

	return db.fetcher.Account(addr)
}

func (db *db) Storage(addr common.Address, slot common.Hash) (common.Hash, error) {
	if db.fetcher == nil {
		return common.Hash{}, nil
	}

	val, err := db.fetcher.StorageAt(addr, slot)
	if err != nil {
		return common.Hash{}, err
	}
	return val, nil
}

func (db *db) Code(addr common.Address, codeHash common.Hash) ([]byte, error) {
	if db.fetcher == nil {
		return []byte{}, nil
	}

	code, err := db.fetcher.Code(codeHash)
	if err != nil {
		return nil, errors.New("not found")
	}
	return code, nil
}

func (db *db) CodeSize(addr common.Address, codeHash common.Hash) (int, error) {
	code, err := db.Code(addr, codeHash)
	if err != nil {
		return 0, err
	}
	return len(code), nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////
// state.Database methods //////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////////////////

func (db *db) Reader(common.Hash) (state.Reader, error) { return db, nil }

func (db *db) OpenTrie(common.Hash) (state.Trie, error) { return db, nil }

func (db *db) OpenStorageTrie(common.Hash, common.Address, common.Hash, state.Trie) (state.Trie, error) {
	return db, nil
}

func (*db) PointCache() *utils.PointCache { panic("not implemented") }

func (*db) TrieDB() *triedb.Database { return fakeTrieDB }

func (*db) Snapshot() *snapshot.Tree { panic("not implemented") }

////////////////////////////////////////////////////////////////////////////////////////////////////
// state.Trie methods //////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////////////////

func (*db) GetKey([]byte) []byte                                         { panic("not implemented") }
func (*db) GetAccount(common.Address) (*types.StateAccount, error)       { panic("not implemented") }
func (*db) GetStorage(common.Address, []byte) ([]byte, error)            { panic("not implemented") }
func (*db) UpdateAccount(common.Address, *types.StateAccount, int) error { panic("not implemented") }
func (*db) UpdateStorage(common.Address, []byte, []byte) error           { panic("not implemented") }
func (*db) DeleteAccount(common.Address) error                           { panic("not implemented") }
func (*db) DeleteStorage(common.Address, []byte) error                   { panic("not implemented") }
func (*db) UpdateContractCode(common.Address, common.Hash, []byte) error { panic("not implemented") }
func (*db) Hash() common.Hash                                            { panic("not implemented") }
func (*db) Commit(bool) (common.Hash, *trienode.NodeSet)                 { panic("not implemented") }
func (*db) Witness() map[string]struct{}                                 { panic("not implemented") }
func (*db) NodeIterator([]byte) (trie.NodeIterator, error)               { panic("not implemented") }
func (*db) Prove([]byte, ethdb.KeyValueWriter) error                     { panic("not implemented") }
func (*db) IsVerkle() bool                                               { panic("not implemented") }
