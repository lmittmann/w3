package eth

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"
)

// SendTransaction sends a signed transaction to the network.
func SendTransaction(tx *types.Transaction) *SendRawTransactionFactory {
	return &SendRawTransactionFactory{tx: tx}
}

// SendRawTransaction sends a raw transaction to the network.
func SendRawTransaction(rawTx []byte) *SendRawTransactionFactory {
	return &SendRawTransactionFactory{rawTx: rawTx}
}

type SendRawTransactionFactory struct {
	// args
	tx    *types.Transaction
	rawTx []byte

	// returns
	result  *common.Hash
	returns *common.Hash
}

func (f *SendRawTransactionFactory) Returns(hash *common.Hash) *SendRawTransactionFactory {
	f.returns = hash
	return f
}

// CreateRequest implements the core.RequestCreator interface.
func (f *SendRawTransactionFactory) CreateRequest() (rpc.BatchElem, error) {
	if f.tx != nil {
		rawTx, err := f.tx.MarshalBinary()
		if err != nil {
			return rpc.BatchElem{}, err
		}
		f.rawTx = rawTx
	}

	return rpc.BatchElem{
		Method: "eth_sendRawTransaction",
		Args:   []interface{}{hexutil.Encode(f.rawTx)},
		Result: &f.result,
	}, nil
}

// HandleResponse implements the core.ResponseHandler interface.
func (f *SendRawTransactionFactory) HandleResponse(elem rpc.BatchElem) error {
	if err := elem.Error; err != nil {
		return err
	}

	*f.returns = *f.result
	return nil
}
