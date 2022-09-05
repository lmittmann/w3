package w3types

import (
	"encoding/json"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core"
	"github.com/lmittmann/w3/internal/crypto"
)

var emptyCodeHash = crypto.Keccak256Hash(nil)

type State map[common.Address]*Account

// SetGenesisAlloc copies the given [core.GenesisAlloc] to the state and
// returns it.
func (s State) SetGenesisAlloc(alloc core.GenesisAlloc) State {
	s = make(State, len(alloc))
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

type Account struct {
	Nonce   uint64
	Balance *big.Int
	Code    []byte
	Storage map[common.Hash]common.Hash

	codeHash *common.Hash // caches the code hash
}

// CodeHash returns the hash of the account's code.
func (acc *Account) CodeHash() common.Hash {
	if acc.codeHash != nil {
		return *acc.codeHash
	}
	if len(acc.Code) == 0 {
		return emptyCodeHash
	}

	h := crypto.Keccak256Hash(acc.Code)
	acc.codeHash = &h
	return h
}

// MarshalJSON implements the [json.Marshaler].
func (acc *Account) MarshalJSON() ([]byte, error) {
	type account struct {
		Nonce   hexutil.Uint64              `json:"nonce,omitempty"`
		Balance *hexutil.Big                `json:"balance,omitempty"`
		Code    hexutil.Bytes               `json:"code,omitempty"`
		Storage map[common.Hash]common.Hash `json:"stateDiff,omitempty"`
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
