package eth

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/lmittmann/w3/internal/module"
	"github.com/lmittmann/w3/w3types"
)

// Balance requests the balance of the given common.Address addr at the given
// blockNumber. If blockNumber is nil, the balance at the latest known block is
// requested.
func Balance(addr common.Address, blockNumber *big.Int) w3types.RPCCallerFactory[*big.Int] {
	return module.NewFactory(
		"eth_getBalance",
		[]any{addr, module.BlockNumberArg(blockNumber)},
		module.WithRetWrapper(module.HexBigRetWrapper),
	)
}
