package eth

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"
)

// Balance requests the balance of the given common.Address addr.
func Balance(addr common.Address) *BalanceFactory {
	return &BalanceFactory{addr: addr}
}

type BalanceFactory struct {
	// args
	addr    common.Address
	atBlock *big.Int

	// returns
	result  hexutil.Big
	returns *big.Int
}

func (f *BalanceFactory) AtBlock(blockNumber *big.Int) *BalanceFactory {
	f.atBlock = blockNumber
	return f
}

func (f *BalanceFactory) Returns(balance *big.Int) *BalanceFactory {
	f.returns = balance
	return f
}

func (f *BalanceFactory) CreateRequest() (rpc.BatchElem, error) {
	return rpc.BatchElem{
		Method: "eth_getBalance",
		Args:   []interface{}{f.addr, toBlockNumberArg(f.atBlock)},
		Result: &f.result,
	}, nil
}

func (f *BalanceFactory) HandleResponse(elem rpc.BatchElem) error {
	if err := elem.Error; err != nil {
		return err
	}
	f.returns.Set((*big.Int)(&f.result))
	return nil
}
