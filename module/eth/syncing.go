package eth

import (
	"github.com/lmittmann/w3/internal/module"
	"github.com/lmittmann/w3/w3types"
)

// Syncing requests the syncing status of the node.
func Syncing() w3types.RPCCallerFactory[bool] {
	return module.NewFactory(
		"eth_syncing",
		[]any{},
		module.WithRetWrapper(syncingRetWrapper),
	)
}

func syncingRetWrapper(ret *bool) any {
	return (*syncingBool)(ret)
}

type syncingBool bool

func (s *syncingBool) UnmarshalJSON(b []byte) error {
	if string(b) == "false" {
		*s = false
	} else {
		*s = true
	}
	return nil
}
