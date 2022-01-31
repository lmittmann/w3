package eth

import (
	"encoding/json"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"
)

func BlockByHash(hash common.Hash) *BlockByHashFactory {
	return &BlockByHashFactory{hash: hash}
}

type BlockByHashFactory struct {
	// args
	hash common.Hash

	// returns
	result  json.RawMessage
	returns *types.Block
}

func (f *BlockByHashFactory) Returns(block *types.Block) *BlockByHashFactory {
	f.returns = block
	return f
}

func (f *BlockByHashFactory) CreateRequest() (rpc.BatchElem, error) {
	return rpc.BatchElem{
		Method: "eth_getBlockByHash",
		Args:   []interface{}{f.hash, true},
		Result: &f.result,
	}, nil
}

func (f *BlockByHashFactory) HandleResponse(elem rpc.BatchElem) error {
	if err := elem.Error; err != nil {
		return err
	}

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
	return nil
}

func HeaderByHash(hash common.Hash) *HeaderByHashFactory {
	return &HeaderByHashFactory{hash: hash}
}

type HeaderByHashFactory struct {
	// args
	hash common.Hash

	// returns
	result  *types.Header
	returns *types.Header
}

func (f *HeaderByHashFactory) Returns(header *types.Header) *HeaderByHashFactory {
	f.returns = header
	return f
}

func (f *HeaderByHashFactory) CreateRequest() (rpc.BatchElem, error) {
	return rpc.BatchElem{
		Method: "eth_getBlockByHash",
		Args:   []interface{}{f.hash, false},
		Result: &f.result,
	}, nil
}

func (f *HeaderByHashFactory) HandleResponse(elem rpc.BatchElem) error {
	if err := elem.Error; err != nil {
		return err
	}
	*f.returns = *f.result
	return nil
}
