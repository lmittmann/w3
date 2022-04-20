package eth

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/lmittmann/w3/core"
)

// SendTransaction sends a signed transaction to the network.
func SendTransaction(tx *types.Transaction) core.CallFactoryReturns[common.Hash] {
	return &sendRawTransactionFactory{tx: tx}
}

// SendRawTransaction sends a raw transaction to the network.
func SendRawTransaction(rawTx []byte) core.CallFactoryReturns[common.Hash] {
	return &sendRawTransactionFactory{rawTx: rawTx}
}

type sendRawTransactionFactory struct {
	// args
	tx    *types.Transaction
	rawTx []byte

	// returns
	returns *common.Hash
}

func (f *sendRawTransactionFactory) Returns(hash *common.Hash) core.Caller {
	f.returns = hash
	return f
}

// CreateRequest implements the core.RequestCreator interface.
func (f *sendRawTransactionFactory) CreateRequest() (rpc.BatchElem, error) {
	if f.tx != nil {
		rawTx, err := f.tx.MarshalBinary()
		if err != nil {
			return rpc.BatchElem{}, err
		}
		f.rawTx = rawTx
	}

	return rpc.BatchElem{
		Method: "eth_sendRawTransaction",
		Args:   []any{hexutil.Encode(f.rawTx)},
		Result: f.returns,
	}, nil
}

// HandleResponse implements the core.ResponseHandler interface.
func (f *sendRawTransactionFactory) HandleResponse(elem rpc.BatchElem) error {
	if err := elem.Error; err != nil {
		return err
	}
	return nil
}
