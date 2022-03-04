package eth

import (
	"encoding/json"
	"math/big"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/lmittmann/w3/core"
)

// BlockByNumber requests the block with full transactions with the given
// number.
func BlockByNumber(number *big.Int) interface {
	core.CallFactoryReturns[*types.Block]
	core.CallFactoryReturnsRAW[*RPCBlock]
} {
	return &blockByNumberFactory{number: number}
}

type blockByNumberFactory struct {
	// args
	number *big.Int

	// returns
	result     json.RawMessage
	returns    *types.Block
	resultRAW  *RPCBlock
	returnsRAW *RPCBlock
}

func (f *blockByNumberFactory) Returns(block *types.Block) core.Caller {
	f.returns = block
	return f
}

func (f *blockByNumberFactory) ReturnsRAW(block *RPCBlock) core.Caller {
	f.returnsRAW = block
	return f
}

// CreateRequest implements the core.RequestCreator interface.
func (f *blockByNumberFactory) CreateRequest() (rpc.BatchElem, error) {
	if f.returns != nil {
		return rpc.BatchElem{
			Method: "eth_getBlockByNumber",
			Args:   []interface{}{toBlockNumberArg(f.number), true},
			Result: &f.result,
		}, nil
	}
	return rpc.BatchElem{
		Method: "eth_getBlockByNumber",
		Args:   []interface{}{toBlockNumberArg(f.number), true},
		Result: &f.resultRAW,
	}, nil
}

// HandleResponse implements the core.ResponseHandler interface.
func (f *blockByNumberFactory) HandleResponse(elem rpc.BatchElem) error {
	if err := elem.Error; err != nil {
		return err
	}

	if f.returns != nil && len(f.result) <= 4 || f.returnsRAW != nil && f.resultRAW == nil {
		return errNotFound
	}

	if f.returns != nil {
		var head *types.Header
		var body rpcBlock
		if err := json.Unmarshal(f.result, &head); err != nil {
			return err
		}
		if err := json.Unmarshal(f.result, &body); err != nil {
			return err
		}

		block := types.NewBlockWithHeader(head).WithBody(body.Transactions, nil)
		*f.returns = *block
	} else {
		*f.returnsRAW = *f.resultRAW
	}
	return nil
}

// HeaderByNumber requests the header with the given number.
func HeaderByNumber(number *big.Int) interface {
	core.CallFactoryReturns[*types.Header]
	core.CallFactoryReturnsRAW[*RPCHeader]
} {
	return &headerByNumberFactory{number: number}
}

type headerByNumberFactory struct {
	// args
	number *big.Int

	// returns
	result     *types.Header
	returns    *types.Header
	resultRAW  *RPCHeader
	returnsRAW *RPCHeader
}

func (f *headerByNumberFactory) Returns(header *types.Header) core.Caller {
	f.returns = header
	return f
}

func (f *headerByNumberFactory) ReturnsRAW(header *RPCHeader) core.Caller {
	f.returnsRAW = header
	return f
}

// CreateRequest implements the core.RequestCreator interface.
func (f *headerByNumberFactory) CreateRequest() (rpc.BatchElem, error) {
	if f.returns != nil {
		return rpc.BatchElem{
			Method: "eth_getBlockByNumber",
			Args:   []interface{}{toBlockNumberArg(f.number), false},
			Result: &f.result,
		}, nil
	}
	return rpc.BatchElem{
		Method: "eth_getBlockByNumber",
		Args:   []interface{}{toBlockNumberArg(f.number), false},
		Result: &f.resultRAW,
	}, nil
}

// HandleResponse implements the core.ResponseHandler interface.
func (f *headerByNumberFactory) HandleResponse(elem rpc.BatchElem) error {
	if err := elem.Error; err != nil {
		return err
	}

	if f.returns != nil && f.result == nil || f.returnsRAW != nil && f.resultRAW == nil {
		return errNotFound
	}

	if f.returns != nil {
		*f.returns = *f.result
	} else {
		*f.returnsRAW = *f.resultRAW
	}
	return nil
}
