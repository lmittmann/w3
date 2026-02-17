package w3vm

import (
	"errors"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/state/snapshot"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/ethereum/go-ethereum/triedb"
	"github.com/holiman/uint256"
)

var fakeTrieDB = triedb.NewDatabase(rawdb.NewMemoryDatabase(), &triedb.Config{})

var fakeTrie, _ = trie.NewStateTrie(&trie.ID{}, triedb.NewDatabase(rawdb.NewMemoryDatabase(), nil))

// db implements the [state.Reader] and [state.Database] interface.
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

func (db *db) Copy() state.Reader { return db }

////////////////////////////////////////////////////////////////////////////////////////////////////
// state.Database methods //////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////////////////

func (db *db) Reader(common.Hash) (state.Reader, error) { return db, nil }

func (db *db) OpenTrie(common.Hash) (state.Trie, error) { return fakeTrie, nil }

func (db *db) OpenStorageTrie(common.Hash, common.Address, common.Hash, state.Trie) (state.Trie, error) {
	panic("not implemented")
}

func (db *db) Has(addr common.Address, codeHash common.Hash) bool {
	code, err := db.Code(addr, codeHash)
	return err == nil && len(code) > 0
}

func (*db) TrieDB() *triedb.Database { return fakeTrieDB }

func (*db) Snapshot() *snapshot.Tree { panic("not implemented") }
