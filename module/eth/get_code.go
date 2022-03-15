package eth

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/lmittmann/w3/core"
)

// Code requests the contract code of the given common.Address addr.
func Code(addr common.Address) *CodeFactory {
	return &CodeFactory{addr: addr}
}

type CodeFactory struct {
	// args
	addr    common.Address
	atBlock *big.Int

	// returns
	result  hexutil.Bytes
	returns *[]byte
}

func (f *CodeFactory) AtBlock(blockNumber *big.Int) *CodeFactory {
	f.atBlock = blockNumber
	return f
}

func (f *CodeFactory) Returns(code *[]byte) core.Caller {
	f.returns = code
	return f
}

// CreateRequest implements the core.RequestCreator interface.
func (f *CodeFactory) CreateRequest() (rpc.BatchElem, error) {
	return rpc.BatchElem{
		Method: "eth_getCode",
		Args:   []any{f.addr, toBlockNumberArg(f.atBlock)},
		Result: &f.result,
	}, nil
}

// HandleResponse implements the core.ResponseHandler interface.
func (f *CodeFactory) HandleResponse(elem rpc.BatchElem) error {
	if err := elem.Error; err != nil {
		return err
	}
	*f.returns = f.result
	return nil
}
