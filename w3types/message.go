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
//
// If no input data is given, but Func is not null, the input data is
// automatically encoded from the given Func and Args arguments by many
// functions that accept a Message struct as an argument.
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

	Func Func  // Func to encode
	Args []any // Arguments for Func
}

// SetTx sets msg to the [types.Transaction] tx and returns msg.
func (msg *Message) SetTx(tx *types.Transaction, signer types.Signer) (*Message, error) {
	from, err := signer.Sender(tx)
	if err != nil {
		return nil, err
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
	return msg, nil
}

// MustSetTx is like [SetTx] but panics if the sender retrieval fails.
func (msg *Message) MustSetTx(tx *types.Transaction, signer types.Signer) *Message {
	msg, err := msg.SetTx(tx, signer)
	if err != nil {
		panic(err)
	}
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

// MarshalJSON implements the [json.Marshaler].
func (msg *Message) MarshalJSON() ([]byte, error) {
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

// UnmarshalJSON implements the [json.Unmarshaler].
func (msg *Message) UnmarshalJSON(data []byte) error {
	var dec message
	if err := json.Unmarshal(data, &dec); err != nil {
		return err
	}

	if dec.From != nil {
		msg.From = *dec.From
	}
	msg.To = dec.To
	msg.Nonce = uint64(dec.Nonce)
	if dec.GasPrice != nil {
		msg.GasPrice = (*big.Int)(dec.GasPrice)
	}
	if dec.GasFeeCap != nil {
		msg.GasFeeCap = (*big.Int)(dec.GasFeeCap)
	}
	if dec.GasTipCap != nil {
		msg.GasTipCap = (*big.Int)(dec.GasTipCap)
	}
	msg.Gas = uint64(dec.Gas)
	if dec.Value != nil {
		msg.Value = (*big.Int)(dec.Value)
	}
	if len(dec.Input) > 0 {
		msg.Input = dec.Input
	}
	if len(dec.AccessList) > 0 {
		msg.AccessList = dec.AccessList
	}
	return nil
}
