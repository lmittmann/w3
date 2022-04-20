package eth

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/lmittmann/w3/core"
)

// Code requests the code of the given common.Address addr at the given
// blockNumber. If blockNumber is nil, the code at the latest known block is
// requested.
func Code(addr common.Address, blockNumber *big.Int) core.CallFactoryReturns[[]byte] {
	return &codeFactory{addr: addr, atBlock: blockNumber}
}

type codeFactory struct {
	// args
	addr    common.Address
	atBlock *big.Int

	// returns
	returns *[]byte
}

func (f *codeFactory) Returns(code *[]byte) core.Caller {
	f.returns = code
	return f
}

// CreateRequest implements the core.RequestCreator interface.
func (f *codeFactory) CreateRequest() (rpc.BatchElem, error) {
	return rpc.BatchElem{
		Method: "eth_getCode",
		Args:   []any{f.addr, toBlockNumberArg(f.atBlock)},
		Result: (*hexutil.Bytes)(f.returns),
	}, nil
}

// HandleResponse implements the core.ResponseHandler interface.
func (f *codeFactory) HandleResponse(elem rpc.BatchElem) error {
	if err := elem.Error; err != nil {
		return err
	}
	return nil
}
