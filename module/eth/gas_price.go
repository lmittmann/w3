package eth

import (
	"math/big"

	"github.com/lmittmann/w3/core"
	"github.com/lmittmann/w3/internal"
)

// GasPrice requests the current gas price in wei.
func GasPrice() core.CallerFactory[big.Int] {
	return internal.NewFactory(
		"eth_gasPrice",
		nil,
		internal.WithRetWrapper(internal.HexBigWrapper),
	)
}
