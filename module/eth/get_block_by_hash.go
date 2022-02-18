package eth

import (
	"encoding/json"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"
)

// BlockByHash requests the block with full transactions with the given hash.
func BlockByHash(hash common.Hash) *BlockByHashFactory {
	return &BlockByHashFactory{hash: hash}
}

type BlockByHashFactory struct {
	// args
	hash common.Hash

	// returns
	result     json.RawMessage
	returns    *types.Block
	resultRAW  *RPCBlock
	returnsRAW *RPCBlock
}

func (f *BlockByHashFactory) Returns(block *types.Block) *BlockByHashFactory {
	f.returns = block
	return f
}

func (f *BlockByHashFactory) ReturnsRAW(block *RPCBlock) *BlockByHashFactory {
	f.returnsRAW = block
	return f
}

// CreateRequest implements the core.RequestCreator interface.
func (f *BlockByHashFactory) CreateRequest() (rpc.BatchElem, error) {
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
func (f *BlockByHashFactory) HandleResponse(elem rpc.BatchElem) error {
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
func HeaderByHash(hash common.Hash) *HeaderByHashFactory {
	return &HeaderByHashFactory{hash: hash}
}

type HeaderByHashFactory struct {
	// args
	hash common.Hash

	// returns
	result     *types.Header
	returns    *types.Header
	resultRAW  *RPCHeader
	returnsRAW *RPCHeader
}

func (f *HeaderByHashFactory) Returns(header *types.Header) *HeaderByHashFactory {
	f.returns = header
	return f
}

func (f *HeaderByHashFactory) ReturnsRAW(header *RPCHeader) *HeaderByHashFactory {
	f.returnsRAW = header
	return f
}

// CreateRequest implements the core.RequestCreator interface.
func (f *HeaderByHashFactory) CreateRequest() (rpc.BatchElem, error) {
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
func (f *HeaderByHashFactory) HandleResponse(elem rpc.BatchElem) error {
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
