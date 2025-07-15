package w3types

import (
	"bytes"
	"encoding/json"
	"maps"
	"math/big"
	"sync/atomic"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/lmittmann/w3/internal/crypto"
)

type State map[common.Address]*Account

// SetGenesisAlloc copies the given [types.GenesisAlloc] to the state and
// returns it.
func (s State) SetGenesisAlloc(alloc types.GenesisAlloc) State {
	clear(s)
	for addr, acc := range alloc {
		s[addr] = &Account{
			Nonce:   acc.Nonce,
			Balance: acc.Balance,
			Code:    acc.Code,
			Storage: acc.Storage,
		}
	}
	return s
}

// Merge returns a new state that is the result of merging the called state with the given state.
// All state in other state will overwrite the state in the called state.
func (s State) Merge(other State) (merged State) {
	merged = make(State, len(s))

	// copy all accounts from s
	for addr, acc := range s {
		merged[addr] = acc.deepCopy()
	}

	// merge all accounts from other
	for addr, acc := range other {
		if mergedAcc, ok := merged[addr]; ok {
			mergedAcc.merge(acc)
		} else {
			merged[addr] = acc.deepCopy()
		}
	}
	return merged
}

type Account struct {
	Nonce   uint64
	Balance *big.Int
	Code    []byte
	Storage Storage

	codeHash atomic.Pointer[common.Hash] // caches the code hash
}

// deepCopy returns a deep copy of the account.
func (acc *Account) deepCopy() *Account {
	newAcc := &Account{Nonce: acc.Nonce}
	if acc.Balance != nil {
		newAcc.Balance = new(big.Int).Set(acc.Balance)
	}
	if acc.Code != nil {
		newAcc.Code = bytes.Clone(acc.Code)
	}
	if len(acc.Storage) > 0 {
		newAcc.Storage = maps.Clone(acc.Storage)
	}
	return newAcc
}

// merge merges the given account into the called account.
func (dst *Account) merge(src *Account) {
	// merge account fields
	srcIsZero := src.Nonce == 0 && src.Balance == nil && len(src.Code) == 0
	if !srcIsZero {
		dst.Nonce = src.Nonce
		if src.Balance != nil {
			dst.Balance = new(big.Int).Set(src.Balance)
		} else {
			dst.Balance = nil
		}
		if len(src.Code) > 0 {
			dst.Code = bytes.Clone(src.Code)
		} else {
			dst.Code = nil
		}
	}

	// merge storage
	if dst.Storage == nil && len(src.Storage) > 0 {
		dst.Storage = maps.Clone(src.Storage)
	} else if len(src.Storage) > 0 {
		maps.Copy(dst.Storage, src.Storage)
	}
}

// CodeHash returns the hash of the account's code.
func (acc *Account) CodeHash() common.Hash {
	if codeHash := acc.codeHash.Load(); codeHash != nil {
		return *codeHash
	}

	if len(acc.Code) == 0 {
		acc.codeHash.Store(&types.EmptyCodeHash)
		return types.EmptyCodeHash
	}

	codeHash := crypto.Keccak256Hash(acc.Code)
	acc.codeHash.Store(&codeHash)
	return codeHash
}

// MarshalJSON implements the [json.Marshaler].
func (acc *Account) MarshalJSON() ([]byte, error) {
	type account struct {
		Nonce   hexutil.Uint64 `json:"nonce,omitempty"`
		Balance *hexutil.Big   `json:"balance,omitempty"`
		Code    hexutil.Bytes  `json:"code,omitempty"`
		Storage Storage        `json:"stateDiff,omitempty"`
	}

	var enc account
	if acc.Nonce > 0 {
		enc.Nonce = hexutil.Uint64(acc.Nonce)
	}
	if acc.Balance != nil {
		enc.Balance = (*hexutil.Big)(acc.Balance)
	}
	if len(acc.Code) > 0 {
		enc.Code = acc.Code
	}
	if len(acc.Storage) > 0 {
		enc.Storage = acc.Storage
	}
	return json.Marshal(&enc)
}

type Storage map[common.Hash]common.Hash
