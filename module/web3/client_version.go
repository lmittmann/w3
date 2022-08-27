package web3

import (
	"github.com/lmittmann/w3/core"
	"github.com/lmittmann/w3/internal/module"
)

// ClientVersion requests the endpoints client version.
func ClientVersion() core.CallerFactory[string] {
	return module.NewFactory[string](
		"web3_clientVersion",
		nil,
	)
}
