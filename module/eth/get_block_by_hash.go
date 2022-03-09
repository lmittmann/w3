package eth

import (
	"encoding/json"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/lmittmann/w3/core"
)

// BlockByHash requests the block with full transactions with the given hash.
func BlockByHash(hash common.Hash) interface {
	core.CallFactoryReturns[types.Block]
	core.CallFactoryReturnsRAW[RPCBlock]
} {
	return &blockByHashFactory{hash: hash}
}

type blockByHashFactory struct {
	// args
	hash common.Hash

	// returns
	result     json.RawMessage
	returns    *types.Block
	resultRAW  *RPCBlock
	returnsRAW *RPCBlock
}

func (f *blockByHashFactory) Returns(block *types.Block) core.Caller {
	f.returns = block
	return f
}

func (f *blockByHashFactory) ReturnsRAW(block *RPCBlock) core.Caller {
	f.returnsRAW = block
	return f
}

// CreateRequest implements the core.RequestCreator interface.
func (f *blockByHashFactory) CreateRequest() (rpc.BatchElem, error) {
	if f.returns != nil {
		return rpc.BatchElem{
			Method: "eth_getBlockByHash",
			Args:   []interface{}{f.hash, true},
			Result: &f.result,
		}, nil
	}
	return rpc.BatchElem{
		Method: "eth_getBlockByHash",
		Args:   []interface{}{f.hash, true},
		Result: &f.resultRAW,
	}, nil
}

// HandleResponse implements the core.ResponseHandler interface.
func (f *blockByHashFactory) HandleResponse(elem rpc.BatchElem) error {
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

// HeaderByHash requests the header with the given hash.
func HeaderByHash(hash common.Hash) interface {
	core.CallFactoryReturns[types.Header]
	core.CallFactoryReturnsRAW[RPCHeader]
} {
	return &headerByHashFactory{hash: hash}
}

type headerByHashFactory struct {
	// args
	hash common.Hash

	// returns
	result     *types.Header
	returns    *types.Header
	resultRAW  *RPCHeader
	returnsRAW *RPCHeader
}

func (f *headerByHashFactory) Returns(header *types.Header) core.Caller {
	f.returns = header
	return f
}

func (f *headerByHashFactory) ReturnsRAW(header *RPCHeader) core.Caller {
	f.returnsRAW = header
	return f
}

// CreateRequest implements the core.RequestCreator interface.
func (f *headerByHashFactory) CreateRequest() (rpc.BatchElem, error) {
	if f.returns != nil {
		return rpc.BatchElem{
			Method: "eth_getBlockByHash",
			Args:   []interface{}{f.hash, false},
			Result: &f.result,
		}, nil
	}
	return rpc.BatchElem{
		Method: "eth_getBlockByHash",
		Args:   []interface{}{f.hash, false},
		Result: &f.resultRAW,
	}, nil
}

// HandleResponse implements the core.ResponseHandler interface.
func (f *headerByHashFactory) HandleResponse(elem rpc.BatchElem) error {
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
