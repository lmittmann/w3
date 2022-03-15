package eth

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/lmittmann/w3/core"
)

// TransactionReceipt requests the receipt of the transaction with the given
// hash.
func TransactionReceipt(hash common.Hash) interface {
	core.CallFactoryReturns[types.Receipt]
	core.CallFactoryReturnsRAW[RPCReceipt]
} {
	return &transactionReceiptFactory{hash: hash}
}

type transactionReceiptFactory struct {
	// args
	hash common.Hash

	// returns
	result     *types.Receipt
	returns    *types.Receipt
	resultRAW  *RPCReceipt
	returnsRAW *RPCReceipt
}

func (f *transactionReceiptFactory) Returns(receipt *types.Receipt) core.Caller {
	f.returns = receipt
	return f
}

func (f *transactionReceiptFactory) ReturnsRAW(receipt *RPCReceipt) core.Caller {
	f.returnsRAW = receipt
	return f
}

// CreateRequest implements the core.RequestCreator interface.
func (f *transactionReceiptFactory) CreateRequest() (rpc.BatchElem, error) {
	if f.returns != nil {
		return rpc.BatchElem{
			Method: "eth_getTransactionReceipt",
			Args:   []any{f.hash},
			Result: &f.result,
		}, nil
	}
	return rpc.BatchElem{
		Method: "eth_getTransactionReceipt",
		Args:   []any{f.hash},
		Result: &f.resultRAW,
	}, nil
}

// HandleResponse implements the core.ResponseHandler interface.
func (f *transactionReceiptFactory) HandleResponse(elem rpc.BatchElem) error {
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
