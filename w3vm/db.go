package w3vm

import (
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	gethState "github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/ethereum/go-ethereum/trie/trienode"
	"github.com/lmittmann/w3/internal/crypto"
	"github.com/lmittmann/w3/w3vm/state"
)

var errNotFound = errors.New("not found")

// db implements the [state.Database] and [state.Trie] interfaces.
type db struct {
	fetcher state.Fetcher
}

func newDB(fetcher state.Fetcher) *db {
	return &db{
		fetcher: fetcher,
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////
// state.Database methods //////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////////////////

func (db *db) OpenTrie(root common.Hash) (gethState.Trie, error) { return db, nil }

func (db *db) OpenStorageTrie(stateRoot common.Hash, addr common.Address, root common.Hash) (gethState.Trie, error) {
	return db, nil
}

func (*db) CopyTrie(gethState.Trie) gethState.Trie { panic("not implemented") }

func (db *db) ContractCode(addr common.Address, codeHash common.Hash) ([]byte, error) {
	if db.fetcher == nil {
		return []byte{}, nil
	}

	code, err := db.fetcher.Code(addr)
	if err != nil {
		return nil, errNotFound
	}
	return code, nil
}

func (db *db) ContractCodeSize(addr common.Address, codeHash common.Hash) (int, error) {
	code, err := db.ContractCode(addr, codeHash)
	if err != nil {
		return 0, err
	}
	return len(code), nil
}

func (*db) DiskDB() ethdb.KeyValueStore { panic("not implemented") }

func (*db) TrieDB() *trie.Database { panic("not implemented") }

////////////////////////////////////////////////////////////////////////////////////////////////////
// state.Trie methods //////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////////////////

func (*db) GetKey([]byte) []byte { panic("not implemented") }

func (db *db) GetStorage(addr common.Address, key []byte) ([]byte, error) {
	if db.fetcher == nil {
		return []byte{}, nil
	}

	storageKey := common.BytesToHash(key)
	storageVal, err := db.fetcher.StorageAt(addr, storageKey)
	if err != nil {
		return nil, err
	}
	return storageVal.Bytes(), nil
}

func (db *db) GetAccount(addr common.Address) (*types.StateAccount, error) {
	if db.fetcher == nil {
		return &types.StateAccount{
			Balance:  new(big.Int),
			CodeHash: types.EmptyCodeHash[:],
		}, nil
	}

	nonce, err := db.fetcher.Nonce(addr)
	if err != nil {
		return nil, err
	}
	balance, err := db.fetcher.Balance(addr)
	if err != nil {
		return nil, err
	}
	code, err := db.fetcher.Code(addr)
	if err != nil {
		return nil, err
	}

	var codeHash []byte
	if len(code) == 0 {
		codeHash = types.EmptyCodeHash[:]
	} else {
		codeHash = crypto.Keccak256(code)
	}

	return &types.StateAccount{
		Nonce:    nonce,
		Balance:  balance,
		CodeHash: codeHash,
	}, nil
}

func (*db) UpdateStorage(addr common.Address, key, value []byte) error { panic("not implemented") }

func (*db) UpdateAccount(addr common.Address, acc *types.StateAccount) error {
	panic("not implemented")
}

func (*db) UpdateContractCode(addr common.Address, codeHash common.Hash, code []byte) error {
	panic("not implemented")
}

func (*db) DeleteStorage(addr common.Address, key []byte) error { panic("not implemented") }

func (*db) DeleteAccount(addr common.Address) error { panic("not implemented") }

func (*db) Hash() common.Hash { panic("not implemented") }

func (*db) Commit(collectLeaf bool) (common.Hash, *trienode.NodeSet, error) {
	panic("not implemented")
}

func (*db) NodeIterator(startKey []byte) (trie.NodeIterator, error) { panic("not implemented") }

func (*db) Prove(key []byte, proofDb ethdb.KeyValueWriter) error { panic("not implemented") }
