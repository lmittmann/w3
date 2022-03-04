package eth

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/lmittmann/w3/core"
)

// ChainID requests the chains ID.
func ChainID() core.CallFactoryReturns[*uint64] {
	return &chainIDFactory{}
}

type chainIDFactory struct {
	// returns
	result  hexutil.Uint64
	returns *uint64
}

func (f *chainIDFactory) Returns(chainID *uint64) core.Caller {
	f.returns = chainID
	return f
}

// CreateRequest implements the core.RequestCreator interface.
func (f *chainIDFactory) CreateRequest() (rpc.BatchElem, error) {
	return rpc.BatchElem{
		Method: "eth_chainId",
		Result: &f.result,
	}, nil
}

// HandleResponse implements the core.ResponseHandler interface.
func (f *chainIDFactory) HandleResponse(elem rpc.BatchElem) error {
	if err := elem.Error; err != nil {
		return err
	}
	*f.returns = uint64(f.result)
	return nil
}
