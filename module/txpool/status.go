package txpool

import (
	"encoding/json"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/lmittmann/w3/internal/module"
	"github.com/lmittmann/w3/w3types"
)

// Status requests the txpool status.
func Status() w3types.CallerFactory[StatusResponse] {
	return module.NewFactory[StatusResponse](
		"txpool_status",
		nil,
	)
}

type StatusResponse struct {
	Pending uint
	Queued  uint
}

func (s *StatusResponse) UnmarshalJSON(input []byte) error {
	type statusResponse struct {
		Pending hexutil.Uint `json:"pending"`
		Queued  hexutil.Uint `json:"queued"`
	}

	var dec statusResponse
	if err := json.Unmarshal(input, &dec); err != nil {
		return err
	}
	s.Pending = uint(dec.Pending)
	s.Queued = uint(dec.Queued)
	return nil
}
