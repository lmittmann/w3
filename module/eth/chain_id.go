package eth

import (
	"github.com/lmittmann/w3/core"
	"github.com/lmittmann/w3/internal/module"
)

// ChainID requests the chains ID.
func ChainID() core.CallerFactory[uint64] {
	return module.NewFactory(
		"eth_chainId",
		nil,
		module.WithRetWrapper(module.HexUint64RetWrapper),
	)
}
