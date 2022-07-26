package eth

import (
	"github.com/lmittmann/w3/core"
	"github.com/lmittmann/w3/internal"
)

// ChainID requests the chains ID.
func ChainID() core.CallerFactory[uint64] {
	return internal.NewFactory(
		"eth_chainId",
		nil,
		internal.WithRetWrapper(internal.HexUint64Wrapper),
	)
}
