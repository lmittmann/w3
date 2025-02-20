/*
Package web3 implements RPC API bindings for methods in the "web3" namespace.
*/
package web3

import (
	"github.com/lmittmann/w3/internal/module"
	"github.com/lmittmann/w3/w3types"
)

// ClientVersion requests the endpoints client version.
func ClientVersion() w3types.RPCCallerFactory[string] {
	return module.NewFactory[string](
		"web3_clientVersion",
		nil,
	)
}
