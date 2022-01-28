package eth

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"
)

func TransactionByHash(hash common.Hash) *GetTransactionByHashFactory {
	return &GetTransactionByHashFactory{hash: hash}
}

type GetTransactionByHashFactory struct {
	// args
	hash common.Hash

	// returns
	result  *types.Transaction
	returns *types.Transaction
}

func (f *GetTransactionByHashFactory) Returns(tx *types.Transaction) *GetTransactionByHashFactory {
	f.returns = tx
	return f
}

func (f *GetTransactionByHashFactory) CreateRequest() (rpc.BatchElem, error) {
	return rpc.BatchElem{
		Method: "eth_getTransactionByHash",
		Args:   []interface{}{f.hash},
		Result: &f.result,
	}, nil
}

func (f *GetTransactionByHashFactory) HandleResponse(elem rpc.BatchElem) error {
	if err := elem.Error; err != nil {
		return err
	}
	*f.returns = *f.result
	return nil
}
