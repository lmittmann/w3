package state

import (
	"errors"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/ethereum/go-ethereum/trie/trienode"
	"github.com/lmittmann/w3/internal/crypto"
	"github.com/lmittmann/w3/w3types"
)

var errNotFound = errors.New("not found")

// DB implements the [state.Database] and [state.Trie] interfaces.
type DB struct {
	fetcher  Fetcher
	accounts map[common.Address]*dbAccount
}

type dbAccount struct {
	*types.StateAccount

	Code    []byte
	Storage map[common.Hash]common.Hash
}

func NewDB(fetcher Fetcher) *DB {
	return &DB{
		fetcher:  fetcher,
		accounts: make(map[common.Address]*dbAccount),
	}
}

func (db *DB) SetState(state w3types.State) {
	db.accounts = make(map[common.Address]*dbAccount)

	for addr, acc := range state {
		db.accounts[addr] = &dbAccount{
			StateAccount: &types.StateAccount{
				Nonce:    acc.Nonce,
				Balance:  acc.Balance,
				CodeHash: acc.CodeHash().Bytes(),
			},
			Code:    acc.Code,
			Storage: acc.Storage,
		}
	}
}

func (db *DB) GetState() w3types.State {
	state := make(w3types.State, len(db.accounts))

	for addr, acc := range db.accounts {
		state[addr] = &w3types.Account{
			Nonce:   acc.Nonce,
			Balance: acc.Balance,
			Code:    acc.Code,
			Storage: acc.Storage,
		}
	}
	return state
}

////////////////////////////////////////////////////////////////////////////////////////////////////
// state.Database methods //////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////////////////

func (db *DB) OpenTrie(root common.Hash) (state.Trie, error) { return db, nil }

func (db *DB) OpenStorageTrie(stateRoot common.Hash, addr common.Address, root common.Hash) (state.Trie, error) {
	return db, nil
}

func (db *DB) CopyTrie(state.Trie) state.Trie { panic("not implemented") }

func (db *DB) ContractCode(addr common.Address, codeHash common.Hash) ([]byte, error) {
	acc, ok := db.accounts[addr]
	if !ok {
		if db.fetcher == nil {
			return nil, errNotFound
		}

		if _, err := db.GetAccount(addr); err != nil {
			return nil, err
		}
		acc = db.accounts[addr]
	}
	return acc.Code, nil
}

func (db *DB) ContractCodeSize(addr common.Address, codeHash common.Hash) (int, error) {
	code, err := db.ContractCode(addr, codeHash)
	if err != nil {
		return 0, err
	}
	return len(code), nil
}

func (db *DB) DiskDB() ethdb.KeyValueStore { return new(noopKeyValueStore) }

func (db *DB) TrieDB() *trie.Database { panic("not implemented") }

////////////////////////////////////////////////////////////////////////////////////////////////////
// state.Trie methods //////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////////////////

func (db *DB) GetKey(b []byte) []byte { panic("not implemented") }

func (db *DB) GetStorage(addr common.Address, key []byte) ([]byte, error) {
	acc, ok := db.accounts[addr]
	if !ok {
		if db.fetcher == nil {
			return nil, errNotFound
		}

		if _, err := db.GetAccount(addr); err != nil {
			return nil, err
		}

		acc = db.accounts[addr]
	}

	if acc.Storage == nil {
		if db.fetcher == nil {
			return nil, errNotFound
		}
	}

	storageKey := common.BytesToHash(key)
	storageVal, ok := acc.Storage[storageKey]
	if !ok {
		if db.fetcher == nil {
			return nil, errNotFound
		}

		var err error
		storageVal, err = db.fetcher.StorageAt(addr, storageKey)
		if err != nil {
			return nil, err
		}
		acc.Storage[storageKey] = storageVal
	}
	return storageVal.Bytes(), nil
}

func (db *DB) GetAccount(addr common.Address) (*types.StateAccount, error) {
	acc, ok := db.accounts[addr]
	if !ok {
		if db.fetcher == nil {
			return nil, errNotFound
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

		acc = &dbAccount{
			StateAccount: &types.StateAccount{
				Nonce:    nonce,
				Balance:  balance,
				CodeHash: crypto.Keccak256(code),
			},
			Code:    code,
			Storage: make(map[common.Hash]common.Hash),
		}
		db.accounts[addr] = acc
	}
	return acc.StateAccount, nil
}

func (db *DB) UpdateStorage(addr common.Address, key, value []byte) error {
	acc, ok := db.accounts[addr]
	if !ok {
		return errNotFound
	}

	if acc.Storage == nil {
		return errNotFound
	}

	acc.Storage[common.BytesToHash(key)] = common.BytesToHash(value)
	return nil
}

func (db *DB) UpdateAccount(addr common.Address, account *types.StateAccount) error {
	acc, ok := db.accounts[addr]
	if !ok {
		return errNotFound
	}

	acc.Nonce = account.Nonce
	acc.Balance = account.Balance
	acc.CodeHash = account.CodeHash[:]
	return nil
}

func (db *DB) UpdateContractCode(addr common.Address, codeHash common.Hash, code []byte) error {
	acc, ok := db.accounts[addr]
	if !ok {
		return errNotFound
	}

	acc.Code = code
	acc.CodeHash = codeHash[:]
	return nil
}

func (db *DB) DeleteStorage(addr common.Address, key []byte) error {
	acc, ok := db.accounts[addr]
	if !ok {
		return errNotFound
	}

	if acc.Storage == nil {
		return errNotFound
	}

	delete(acc.Storage, common.BytesToHash(key))
	return nil
}

// DeleteAccount abstracts an account deletion from the trie.
func (db *DB) DeleteAccount(addr common.Address) error {
	delete(db.accounts, addr)
	return nil
}

func (db *DB) Hash() common.Hash { return hash0 }

func (db *DB) Commit(collectLeaf bool) (common.Hash, *trienode.NodeSet, error) {
	return hash0, nil, nil
}

func (db *DB) NodeIterator(startKey []byte) (trie.NodeIterator, error) {
	return new(noopNodeIterator), nil
}

func (db *DB) Prove(key []byte, proofDb ethdb.KeyValueWriter) error { panic("not implemented") }
