package eth

import (
	"math/big"

	"github.com/lmittmann/w3/core"
	"github.com/lmittmann/w3/internal/module"
)

// BlockNumber requests the number of the most recent block.
func BlockNumber() core.CallerFactory[big.Int] {
	return module.NewFactory(
		"eth_blockNumber",
		nil,
		module.WithRetWrapper(module.HexBigRetWrapper),
	)
}
