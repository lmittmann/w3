package eth

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"
)

// GetCode requests the code of the given common.Address addr.
func GetCode(addr common.Address) *GetCodeFactory {
	return &GetCodeFactory{addr: addr}
}

type GetCodeFactory struct {
	// args
	addr    common.Address
	atBlock *big.Int

	// returns
	result  hexutil.Bytes
	returns *[]byte
}

func (f *GetCodeFactory) AtBlock(blockNumber *big.Int) *GetCodeFactory {
	f.atBlock = blockNumber
	return f
}

func (f *GetCodeFactory) Returns(code *[]byte) *GetCodeFactory {
	f.returns = code
	return f
}

func (f *GetCodeFactory) CreateRequest() (rpc.BatchElem, error) {
	return rpc.BatchElem{
		Method: "eth_getCode",
		Args:   []interface{}{f.addr, toBlockNumberArg(f.atBlock)},
		Result: &f.result,
	}, nil
}

func (f *GetCodeFactory) HandleResponse(elem rpc.BatchElem) error {
	if err := elem.Error; err != nil {
		return err
	}
	*f.returns = f.result
	return nil
}
