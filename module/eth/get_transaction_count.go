package eth

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/lmittmann/w3/core"
)

// Nonce requests the nonce of the given common.Address addr.
func Nonce(addr common.Address) *NonceFactory {
	return &NonceFactory{addr: addr}
}

type NonceFactory struct {
	// args
	addr    common.Address
	atBlock *big.Int

	// returns
	result  hexutil.Uint64
	returns *uint64
}

func (f *NonceFactory) AtBlock(blockNumber *big.Int) *NonceFactory {
	f.atBlock = blockNumber
	return f
}

func (f *NonceFactory) Returns(nonce *uint64) core.Caller {
	f.returns = nonce
	return f
}

// CreateRequest implements the core.RequestCreator interface.
func (f *NonceFactory) CreateRequest() (rpc.BatchElem, error) {
	return rpc.BatchElem{
		Method: "eth_getTransactionCount",
		Args:   []any{f.addr, toBlockNumberArg(f.atBlock)},
		Result: &f.result,
	}, nil
}

// HandleResponse implements the core.ResponseHandler interface.
func (f *NonceFactory) HandleResponse(elem rpc.BatchElem) error {
	if err := elem.Error; err != nil {
		return err
	}
	*f.returns = uint64(f.result)
	return nil
}
