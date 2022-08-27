package eth

import (
	"math/big"

	"github.com/lmittmann/w3/core"
	"github.com/lmittmann/w3/internal/module"
)

// GasPrice requests the current gas price in wei.
func GasPrice() core.CallerFactory[big.Int] {
	return module.NewFactory(
		"eth_gasPrice",
		nil,
		module.WithRetWrapper(module.HexBigRetWrapper),
	)
}
