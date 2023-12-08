package eth

import (
	"math/big"

	"github.com/lmittmann/w3/internal/module"
	"github.com/lmittmann/w3/w3types"
)

// GasTipCap requests the current gas tip cap after 1559 in wei.
func GasTipCap() w3types.CallerFactory[big.Int] {
	return module.NewFactory(
		"eth_maxPriorityFeePerGas",
		nil,
		module.WithRetWrapper(module.HexBigRetWrapper),
	)
}
