package eth

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"
)

// TransactionByHash requests the transaction with the given hash.
func TransactionByHash(hash common.Hash) *TransactionByHashFactory {
	return &TransactionByHashFactory{hash: hash}
}

type TransactionByHashFactory struct {
	// args
	hash common.Hash

	// returns
	result     *types.Transaction
	returns    *types.Transaction
	resultRAW  *RPCTransaction
	returnsRAW *RPCTransaction
}

func (f *TransactionByHashFactory) Returns(tx *types.Transaction) *TransactionByHashFactory {
	f.returns = tx
	return f
}

func (f *TransactionByHashFactory) ReturnsRAW(tx *RPCTransaction) *TransactionByHashFactory {
	f.returnsRAW = tx
	return f
}

// CreateRequest implements the core.RequestCreator interface.
func (f *TransactionByHashFactory) CreateRequest() (rpc.BatchElem, error) {
	if f.returns != nil {
		return rpc.BatchElem{
			Method: "eth_getTransactionByHash",
			Args:   []interface{}{f.hash},
			Result: &f.result,
		}, nil
	}
	return rpc.BatchElem{
		Method: "eth_getTransactionByHash",
		Args:   []interface{}{f.hash},
		Result: &f.resultRAW,
	}, nil
}

// HandleResponse implements the core.ResponseHandler interface.
func (f *TransactionByHashFactory) HandleResponse(elem rpc.BatchElem) error {
	if err := elem.Error; err != nil {
		return err
	}
	if f.returns != nil && f.result == nil || f.returnsRAW != nil && f.resultRAW == nil {
		return errNotFound
	}

	if f.returns != nil {
		*f.returns = *f.result
	} else {
		*f.returnsRAW = *f.resultRAW
	}
	return nil
}
