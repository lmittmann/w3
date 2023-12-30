package eth

import (
	"math/big"

	"github.com/lmittmann/w3/internal/module"
	"github.com/lmittmann/w3/w3types"
)

// GasTipCap requests the currently suggested gas tip cap after EIP-1559 to
// allow a timely execution of a transaction.
func GasTipCap() w3types.CallerFactory[big.Int] {
	return module.NewFactory(
		"eth_maxPriorityFeePerGas",
		nil,
		module.WithRetWrapper(module.HexBigRetWrapper),
	)
}
