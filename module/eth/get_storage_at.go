package eth

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/lmittmann/w3/core"
)

// StorageAt requests the storage of the given common.Address addr at the
// given common.Hash slot at the given blockNumber. If block number is nil, the
// slot at the latest known block is requested.
func StorageAt(addr common.Address, slot common.Hash, blockNumber *big.Int) core.CallFactoryReturns[common.Hash] {
	return &storageAtFactory{addr: addr, slot: slot, atBlock: blockNumber}
}

type storageAtFactory struct {
	// args
	addr    common.Address
	slot    common.Hash
	atBlock *big.Int

	// returns
	result  common.Hash
	returns *common.Hash
}

func (f *storageAtFactory) Returns(storage *common.Hash) core.Caller {
	f.returns = storage
	return f
}

// CreateRequest implements the core.RequestCreator interface.
func (f *storageAtFactory) CreateRequest() (rpc.BatchElem, error) {
	return rpc.BatchElem{
		Method: "eth_getStorageAt",
		Args:   []any{f.addr, f.slot, toBlockNumberArg(f.atBlock)},
		Result: &f.result,
	}, nil
}

// HandleResponse implements the core.ResponseHandler interface.
func (f *storageAtFactory) HandleResponse(elem rpc.BatchElem) error {
	if err := elem.Error; err != nil {
		return err
	}
	if f.returns != nil {
		*f.returns = (common.Hash)(f.result)
	}
	return nil
}
