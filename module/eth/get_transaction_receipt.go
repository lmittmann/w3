package eth

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"
)

func TransactionReceipt(hash common.Hash) *GetTransactionReceiptFactory {
	return &GetTransactionReceiptFactory{hash: hash}
}

type GetTransactionReceiptFactory struct {
	// args
	hash common.Hash

	// returns
	result  *types.Receipt
	returns *types.Receipt
}

func (f *GetTransactionReceiptFactory) Returns(receipt *types.Receipt) *GetTransactionReceiptFactory {
	f.returns = receipt
	return f
}

func (f *GetTransactionReceiptFactory) CreateRequest() (rpc.BatchElem, error) {
	return rpc.BatchElem{
		Method: "eth_getTransactionReceipt",
		Args:   []interface{}{f.hash},
		Result: &f.result,
	}, nil
}

func (f *GetTransactionReceiptFactory) HandleResponse(elem rpc.BatchElem) error {
	if err := elem.Error; err != nil {
		return err
	}
	*f.returns = *f.result
	return nil
}
