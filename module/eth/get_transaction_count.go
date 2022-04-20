package eth

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/lmittmann/w3/core"
)

// Nonce requests the nonce of the given common.Address addr at the given
// blockNumber. If blockNumber is nil, the nonce at the latest known block is
// requested.
func Nonce(addr common.Address, blockNumber *big.Int) core.CallFactoryReturns[uint64] {
	return &nonceFactory{addr: addr, atBlock: blockNumber}
}

type nonceFactory struct {
	// args
	addr    common.Address
	atBlock *big.Int

	// returns
	returns *uint64
}

func (f *nonceFactory) Returns(nonce *uint64) core.Caller {
	f.returns = nonce
	return f
}

// CreateRequest implements the core.RequestCreator interface.
func (f *nonceFactory) CreateRequest() (rpc.BatchElem, error) {
	return rpc.BatchElem{
		Method: "eth_getTransactionCount",
		Args:   []any{f.addr, toBlockNumberArg(f.atBlock)},
		Result: (*hexutil.Uint64)(f.returns),
	}, nil
}

// HandleResponse implements the core.ResponseHandler interface.
func (f *nonceFactory) HandleResponse(elem rpc.BatchElem) error {
	if err := elem.Error; err != nil {
		return err
	}
	return nil
}
