package eth

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"
)

// Nonce requests the nonce of the given common.Address addr.
func Nonce(addr common.Address) *GetTransactionCountFactory {
	return &GetTransactionCountFactory{addr: addr}
}

type GetTransactionCountFactory struct {
	// args
	addr    common.Address
	atBlock *big.Int

	// returns
	result  hexutil.Uint64
	returns *uint64
}

func (f *GetTransactionCountFactory) AtBlock(blockNumber *big.Int) *GetTransactionCountFactory {
	f.atBlock = blockNumber
	return f
}

func (f *GetTransactionCountFactory) Returns(nonce *uint64) *GetTransactionCountFactory {
	f.returns = nonce
	return f
}

func (f *GetTransactionCountFactory) CreateRequest() (rpc.BatchElem, error) {
	return rpc.BatchElem{
		Method: "eth_getTransactionCount",
		Args:   []interface{}{f.addr, toBlockNumberArg(f.atBlock)},
		Result: &f.result,
	}, nil
}

func (f *GetTransactionCountFactory) HandleResponse(elem rpc.BatchElem) error {
	if err := elem.Error; err != nil {
		return err
	}
	*f.returns = uint64(f.result)
	return nil
}
