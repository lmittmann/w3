/*
Package net implements RPC API bindings for methods in the "net" namespace.
*/
package net

import (
	"math/big"

	"github.com/lmittmann/w3/internal/module"
	"github.com/lmittmann/w3/w3types"
)

// Listening returns whether the client is actively listening for network connections.
func Listening() w3types.RPCCallerFactory[bool] {
	return module.NewFactory[bool](
		"net_listening",
		[]any{},
	)
}

// PeerCount returns the number of peers connected to the node.
func PeerCount() w3types.RPCCallerFactory[*big.Int] {
	return module.NewFactory[*big.Int](
		"net_peerCount",
		[]any{},
	)
}

// Version returns the network ID (e.g. 1 for mainnet).
func Version() w3types.RPCCallerFactory[int] {
	return module.NewFactory[int](
		"net_version",
		[]any{},
	)
}
