package eth

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/lmittmann/w3/core"
	"github.com/lmittmann/w3/internal"
)

// Balance requests the balance of the given common.Address addr at the given
// blockNumber. If blockNumber is nil, the balance at the latest known block is
// requested.
func Balance(addr common.Address, blockNumber *big.Int) core.CallerFactory[big.Int] {
	return internal.NewFactory(
		"eth_getBalance",
		[]any{addr, toBlockNumberArg(blockNumber)},
		internal.WithRetWrapper(internal.HexBigWrapper),
	)
}
