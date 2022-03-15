package eth

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/lmittmann/w3/core"
)

// StorageAt requests the storage of the given common.Address addr at the
// given common.Hash slot.
func StorageAt(addr common.Address, slot common.Hash) *StorageAtFactory {
	return &StorageAtFactory{addr: addr, slot: slot}
}

type StorageAtFactory struct {
	// args
	addr    common.Address
	slot    common.Hash
	atBlock *big.Int

	// returns
	result  common.Hash
	returns *common.Hash
}

func (f *StorageAtFactory) AtBlock(blockNumber *big.Int) *StorageAtFactory {
	f.atBlock = blockNumber
	return f
}

func (f *StorageAtFactory) Returns(storage *common.Hash) core.Caller {
	f.returns = storage
	return f
}

// CreateRequest implements the core.RequestCreator interface.
func (f *StorageAtFactory) CreateRequest() (rpc.BatchElem, error) {
	return rpc.BatchElem{
		Method: "eth_getStorageAt",
		Args:   []any{f.addr, f.slot, toBlockNumberArg(f.atBlock)},
		Result: &f.result,
	}, nil
}

// HandleResponse implements the core.ResponseHandler interface.
func (f *StorageAtFactory) HandleResponse(elem rpc.BatchElem) error {
	if err := elem.Error; err != nil {
		return err
	}

	*f.returns = (common.Hash)(f.result)
	return nil
}
