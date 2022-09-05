package w3types

import (
	"encoding/json"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
)

var addr0 = common.Address{}

// Message represents a transaction without the signature.
type Message struct {
	From       common.Address  // Sender
	To         *common.Address // Recipient
	Nonce      uint64
	GasPrice   *big.Int
	GasFeeCap  *big.Int
	GasTipCap  *big.Int
	Gas        uint64
	Value      *big.Int
	Input      []byte // Input data
	AccessList types.AccessList

	Func Func
	Args []any
}

// SetTx sets msg to the [types.Transaction] tx and returns msg.
func (msg *Message) SetTx(tx *types.Transaction, signer types.Signer) *Message {
	from, err := signer.Sender(tx)
	if err != nil {
		panic(err)
	}

	msg.From = from
	msg.To = tx.To()
	msg.Nonce = tx.Nonce()
	msg.GasPrice = tx.GasPrice()
	msg.GasFeeCap = tx.GasFeeCap()
	msg.GasTipCap = tx.GasTipCap()
	msg.Gas = tx.Gas()
	msg.Value = tx.Value()
	msg.Input = tx.Data()
	msg.AccessList = tx.AccessList()
	return msg
}

// SetCallMsg sets msg to the [ethereum.CallMsg] callMsg and returns msg.
func (msg *Message) SetCallMsg(callMsg ethereum.CallMsg) *Message {
	msg.From = callMsg.From
	msg.To = callMsg.To
	msg.Gas = callMsg.Gas
	msg.GasPrice = callMsg.GasPrice
	msg.GasFeeCap = callMsg.GasFeeCap
	msg.GasTipCap = callMsg.GasTipCap
	msg.Value = callMsg.Value
	msg.Input = callMsg.Data
	msg.AccessList = callMsg.AccessList
	return msg
}

// MarshalJSON implements the [json.Marshaler].
func (msg *Message) MarshalJSON() ([]byte, error) {
	type message struct {
		From       *common.Address  `json:"from,omitempty"`
		To         *common.Address  `json:"to,omitempty"`
		Nonce      hexutil.Uint64   `json:"nonce,omitempty"`
		GasPrice   *hexutil.Big     `json:"gasPrice,omitempty"`
		GasFeeCap  *hexutil.Big     `json:"gasFeeCap,omitempty"`
		GasTipCap  *hexutil.Big     `json:"gasTipCap,omitempty"`
		Gas        hexutil.Uint64   `json:"gas,omitempty"`
		Value      *hexutil.Big     `json:"value,omitempty"`
		Input      hexutil.Bytes    `json:"data,omitempty"`
		AccessList types.AccessList `json:"accessList,omitempty"`
	}

	var enc message
	if msg.From != addr0 {
		enc.From = &msg.From
	}
	enc.To = msg.To
	enc.Nonce = hexutil.Uint64(msg.Nonce)
	if msg.GasPrice != nil {
		enc.GasPrice = (*hexutil.Big)(msg.GasPrice)
	}
	if msg.GasFeeCap != nil {
		enc.GasFeeCap = (*hexutil.Big)(msg.GasFeeCap)
	}
	if msg.GasTipCap != nil {
		enc.GasTipCap = (*hexutil.Big)(msg.GasTipCap)
	}
	if msg.Gas > 0 {
		enc.Gas = hexutil.Uint64(msg.Gas)
	}
	if msg.Value != nil {
		enc.Value = (*hexutil.Big)(msg.Value)
	}
	if len(msg.Input) > 0 {
		enc.Input = msg.Input
	}
	if len(msg.AccessList) > 0 {
		enc.AccessList = msg.AccessList
	}
	return json.Marshal(&enc)
}
