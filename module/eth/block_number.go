package eth

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/lmittmann/w3/core"
)

// BlockNumber requests the number of the most recent block.
func BlockNumber() core.CallReturnsFactory[*big.Int] {
	return &blockNumberFactory{}
}

type blockNumberFactory struct {
	// returns
	result  hexutil.Big
	returns *big.Int
}

func (f *blockNumberFactory) Returns(blockNumber *big.Int) core.Caller {
	f.returns = blockNumber
	return f
}

// CreateRequest implements the core.RequestCreator interface.
func (f *blockNumberFactory) CreateRequest() (rpc.BatchElem, error) {
	return rpc.BatchElem{
		Method: "eth_blockNumber",
		Result: &f.result,
	}, nil
}

// HandleResponse implements the core.ResponseHandler interface.
func (f *blockNumberFactory) HandleResponse(elem rpc.BatchElem) error {
	if err := elem.Error; err != nil {
		return err
	}
	f.returns.Set((*big.Int)(&f.result))
	return nil
}
