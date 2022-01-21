package eth

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"
)

// Balance requests the balance of the given common.Address addr.
func Balance(addr common.Address) *GetBalanceFactory {
	return &GetBalanceFactory{addr: addr}
}

type GetBalanceFactory struct {
	// args
	addr    common.Address
	atBlock *big.Int

	// returns
	result  hexutil.Big
	returns *big.Int
}

func (f *GetBalanceFactory) AtBlock(blockNumber *big.Int) *GetBalanceFactory {
	f.atBlock = blockNumber
	return f
}

func (f *GetBalanceFactory) Returns(balance *big.Int) *GetBalanceFactory {
	f.returns = balance
	return f
}

func (f *GetBalanceFactory) CreateRequest() (rpc.BatchElem, error) {
	return rpc.BatchElem{
		Method: "eth_getBalance",
		Args:   []interface{}{f.addr, toBlockNumberArg(f.atBlock)},
		Result: &f.result,
	}, nil
}

func (f *GetBalanceFactory) HandleResponse(elem rpc.BatchElem) error {
	if err := elem.Error; err != nil {
		return err
	}
	f.returns.Set((*big.Int)(&f.result))
	return nil
}
