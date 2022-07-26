package eth

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/lmittmann/w3/core"
)

// Balance requests the balance of the given common.Address addr at the given
// blockNumber. If blockNumber is nil, the balance at the latest known block is
// requested.
func Balance(addr common.Address, blockNumber *big.Int) core.CallerFactory[big.Int] {
	return &balanceFactory{addr: addr, atBlock: blockNumber}
}

type balanceFactory struct {
	// args
	addr    common.Address
	atBlock *big.Int

	// returns
	returns *big.Int
}

func (f *balanceFactory) Returns(balance *big.Int) core.Caller {
	f.returns = balance
	return f
}

// CreateRequest implements the core.RequestCreator interface.
func (f *balanceFactory) CreateRequest() (rpc.BatchElem, error) {
	return rpc.BatchElem{
		Method: "eth_getBalance",
		Args:   []any{f.addr, toBlockNumberArg(f.atBlock)},
		Result: (*hexutil.Big)(f.returns),
	}, nil
}

// HandleResponse implements the core.ResponseHandler interface.
func (f *balanceFactory) HandleResponse(elem rpc.BatchElem) error {
	if err := elem.Error; err != nil {
		return err
	}
	return nil
}
