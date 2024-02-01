package eth

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/lmittmann/w3/internal/module"
	"github.com/lmittmann/w3/w3types"
)

// StorageAt requests the storage of the given common.Address addr at the
// given common.Hash slot at the given blockNumber. If block number is nil, the
// slot at the latest known block is requested.
func StorageAt(addr common.Address, slot common.Hash, blockNumber *big.Int) w3types.RPCCallerFactory[common.Hash] {
	return module.NewFactory[common.Hash](
		"eth_getStorageAt",
		[]any{addr, slot, module.BlockNumberArg(blockNumber)},
	)
}
