package eth

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rpc"
)

// GetStorageAt requests the storage of the given common.Address addr at the
// given common.Hash slot.
func GetStorageAt(addr common.Address, slot common.Hash) *GetStorageAtFactory {
	return &GetStorageAtFactory{addr: addr, slot: slot}
}

type GetStorageAtFactory struct {
	// args
	addr    common.Address
	slot    common.Hash
	atBlock *big.Int

	// returns
	result  common.Hash
	returns *common.Hash
}

func (f *GetStorageAtFactory) AtBlock(blockNumber *big.Int) *GetStorageAtFactory {
	f.atBlock = blockNumber
	return f
}

func (f *GetStorageAtFactory) Returns(storage *common.Hash) *GetStorageAtFactory {
	f.returns = storage
	return f
}

func (f *GetStorageAtFactory) CreateRequest() (rpc.BatchElem, error) {
	return rpc.BatchElem{
		Method: "eth_getStorageAt",
		Args:   []interface{}{f.addr, f.slot, toBlockNumberArg(f.atBlock)},
		Result: &f.result,
	}, nil
}

func (f *GetStorageAtFactory) HandleResponse(elem rpc.BatchElem) error {
	if err := elem.Error; err != nil {
		return err
	}

	*f.returns = (common.Hash)(f.result)
	return nil
}
