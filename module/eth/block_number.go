package eth

import (
	"math/big"

	"github.com/lmittmann/w3/core"
	"github.com/lmittmann/w3/internal"
)

// BlockNumber requests the number of the most recent block.
func BlockNumber() core.CallerFactory[big.Int] {
	return internal.NewFactory(
		"eth_blockNumber",
		nil,
		internal.WithRetWrapper(internal.HexBigWrapper),
	)
}
