package eth

import (
	"encoding/json"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

type Account struct {
	Nonce     *uint64
	Code      []byte
	Balance   *big.Int
	State     map[common.Hash]common.Hash
	StateDiff map[common.Hash]common.Hash
}

type account struct {
	Nonce     *hexutil.Uint64             `json:"nonce,omitempty"`
	Code      *hexutil.Bytes              `json:"code,omitempty"`
	Balance   *hexutil.Big                `json:"balance,omitempty"`
	State     map[common.Hash]common.Hash `json:"state,omitempty"`
	StateDiff map[common.Hash]common.Hash `json:"stateDiff,omitempty"`
}

// MarshalJSON implements the json.Marshaler.
func (oa Account) MarshalJSON() ([]byte, error) {
	var enc account
	if oa.Nonce != nil {
		hexNonce := hexutil.Uint64(*oa.Nonce)
		enc.Nonce = &hexNonce
	}
	if oa.Code != nil {
		hexCode := hexutil.Bytes(oa.Code)
		enc.Code = &hexCode
	}
	enc.Balance = (*hexutil.Big)(oa.Balance)
	enc.State = oa.State
	enc.StateDiff = oa.StateDiff
	return json.Marshal(&enc)
}

// AccountOverrides is the collection of overridden accounts.
type AccountOverrides map[common.Address]Account
