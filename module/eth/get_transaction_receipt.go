package eth

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"
)

// TransactionReceipt requests the receipt of the transaction with the given
// hash.
func TransactionReceipt(hash common.Hash) *TransactionReceiptFactory {
	return &TransactionReceiptFactory{hash: hash}
}

type TransactionReceiptFactory struct {
	// args
	hash common.Hash

	// returns
	result     *types.Receipt
	returns    *types.Receipt
	resultRAW  *RPCReceipt
	returnsRAW *RPCReceipt
}

func (f *TransactionReceiptFactory) Returns(receipt *types.Receipt) *TransactionReceiptFactory {
	f.returns = receipt
	return f
}

func (f *TransactionReceiptFactory) ReturnsRAW(receipt *RPCReceipt) *TransactionReceiptFactory {
	f.returnsRAW = receipt
	return f
}

// CreateRequest implements the core.RequestCreator interface.
func (f *TransactionReceiptFactory) CreateRequest() (rpc.BatchElem, error) {
	if f.returns != nil {
		return rpc.BatchElem{
			Method: "eth_getTransactionReceipt",
			Args:   []interface{}{f.hash},
			Result: &f.result,
		}, nil
	}
	return rpc.BatchElem{
		Method: "eth_getTransactionReceipt",
		Args:   []interface{}{f.hash},
		Result: &f.resultRAW,
	}, nil
}

// HandleResponse implements the core.ResponseHandler interface.
func (f *TransactionReceiptFactory) HandleResponse(elem rpc.BatchElem) error {
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
