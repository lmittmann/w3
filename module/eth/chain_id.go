package eth

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"
)

// ChainID requests the chains ID.
func ChainID() *ChainIDFactory {
	return &ChainIDFactory{}
}

type ChainIDFactory struct {
	// returns
	result  hexutil.Uint64
	returns *uint64
}

func (f *ChainIDFactory) Returns(chainID *uint64) *ChainIDFactory {
	f.returns = chainID
	return f
}

// CreateRequest implements the core.RequestCreater interface.
func (f *ChainIDFactory) CreateRequest() (rpc.BatchElem, error) {
	return rpc.BatchElem{
		Method: "eth_chainId",
		Result: &f.result,
	}, nil
}

// HandleResponse implements the core.ResponseHandler interface.
func (f *ChainIDFactory) HandleResponse(elem rpc.BatchElem) error {
	if err := elem.Error; err != nil {
		return err
	}
	*f.returns = uint64(f.result)
	return nil
}
