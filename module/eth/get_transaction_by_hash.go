package eth

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/lmittmann/w3/core"
)

// TransactionByHash requests the transaction with the given hash.
func TransactionByHash(hash common.Hash) interface {
	core.CallFactoryReturns[types.Transaction]
	core.CallFactoryReturnsRAW[RPCTransaction]
} {
	return &transactionByHashFactory{hash: hash}
}

type transactionByHashFactory struct {
	// args
	hash common.Hash

	// returns
	result     *types.Transaction
	returns    *types.Transaction
	resultRAW  *RPCTransaction
	returnsRAW *RPCTransaction
}

func (f *transactionByHashFactory) Returns(tx *types.Transaction) core.Caller {
	f.returns = tx
	return f
}

func (f *transactionByHashFactory) ReturnsRAW(tx *RPCTransaction) core.Caller {
	f.returnsRAW = tx
	return f
}

// CreateRequest implements the core.RequestCreator interface.
func (f *transactionByHashFactory) CreateRequest() (rpc.BatchElem, error) {
	if f.returns != nil {
		return rpc.BatchElem{
			Method: "eth_getTransactionByHash",
			Args:   []any{f.hash},
			Result: &f.result,
		}, nil
	}
	return rpc.BatchElem{
		Method: "eth_getTransactionByHash",
		Args:   []any{f.hash},
		Result: &f.resultRAW,
	}, nil
}

// HandleResponse implements the core.ResponseHandler interface.
func (f *transactionByHashFactory) HandleResponse(elem rpc.BatchElem) error {
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
