package w3types

import (
	"encoding/json"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

type BlockOverrides struct {
	Number        *big.Int
	Difficulty    *big.Int
	Time          uint64
	GasLimit      uint64
	FeeRecipient  common.Address
	PrevRandao    common.Hash
	BaseFeePerGas *big.Int
	BlobBaseFee   *big.Int
}

type blockOverrides struct {
	Number        *hexutil.Big    `json:"number,omitempty"`
	Difficulty    *hexutil.Big    `json:"difficulty,omitempty"`
	Time          hexutil.Uint64  `json:"time,omitempty"`
	GasLimit      hexutil.Uint64  `json:"gasLimit,omitempty"`
	FeeRecipient  *common.Address `json:"feeRecipient,omitempty"` // TODO: omitzero (Go1.24)
	PrevRandao    *common.Hash    `json:"prevRandao,omitempty"`   // TODO: omitzero (Go1.24)
	BaseFeePerGas *hexutil.Big    `json:"baseFeePerGas,omitempty"`
	BlobBaseFee   *hexutil.Big    `json:"blobBaseFee,omitempty"`
}

func (o BlockOverrides) MarshalJSON() ([]byte, error) {
	dec := &blockOverrides{
		Number:        (*hexutil.Big)(o.Number),
		Difficulty:    (*hexutil.Big)(o.Difficulty),
		Time:          hexutil.Uint64(o.Time),
		GasLimit:      hexutil.Uint64(o.GasLimit),
		BaseFeePerGas: (*hexutil.Big)(o.BaseFeePerGas),
		BlobBaseFee:   (*hexutil.Big)(o.BlobBaseFee),
	}
	if o.FeeRecipient != (common.Address{}) {
		dec.FeeRecipient = &o.FeeRecipient
	}
	if o.PrevRandao != (common.Hash{}) {
		dec.PrevRandao = &o.PrevRandao
	}
	return json.Marshal(dec)
}
