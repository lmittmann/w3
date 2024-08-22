package debug

import (
	"encoding/json"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/lmittmann/w3/internal/module"
	"github.com/lmittmann/w3/w3types"
)

// CallTraceCall requests the call trace of the given message.
func CallTraceCall(msg *w3types.Message, blockNumber *big.Int, overrides w3types.State) w3types.RPCCallerFactory[*CallTrace] {
	return module.NewFactory(
		"debug_traceCall",
		[]any{msg, module.BlockNumberArg(blockNumber), &traceConfig{Tracer: "callTracer", Overrides: overrides}},
		module.WithArgsWrapper[*CallTrace](msgArgsWrapper),
	)
}

// CallTraceTx requests the call trace of the transaction with the given hash.
func CallTraceTx(txHash common.Hash, overrides w3types.State) w3types.RPCCallerFactory[*CallTrace] {
	return module.NewFactory[*CallTrace](
		"debug_traceTransaction",
		[]any{txHash, &traceConfig{Tracer: "callTracer", Overrides: overrides}},
	)
}

type CallTrace struct {
	From         common.Address
	To           common.Address
	Type         string
	Gas          uint64
	GasUsed      uint64
	Value        *big.Int
	Input        []byte
	Output       []byte
	Error        string
	RevertReason string
	Calls        []*CallTrace
}

// UnmarshalJSON implements the [json.Unmarshaler].
func (c *CallTrace) UnmarshalJSON(data []byte) error {
	type call struct {
		From         common.Address `json:"from"`
		To           common.Address `json:"to"`
		Type         string         `json:"type"`
		Gas          hexutil.Uint64 `json:"gas"`
		GasUsed      hexutil.Uint64 `json:"gasUsed"`
		Value        *hexutil.Big   `json:"value"`
		Input        hexutil.Bytes  `json:"input"`
		Output       hexutil.Bytes  `json:"output"`
		Error        string         `json:"error"`
		RevertReason string         `json:"revertReason"`
		Calls        []*CallTrace   `json:"calls"`
	}

	var dec call
	if err := json.Unmarshal(data, &dec); err != nil {
		return err
	}

	c.From = dec.From
	c.To = dec.To
	c.Type = dec.Type
	c.Gas = uint64(dec.Gas)
	c.GasUsed = uint64(dec.GasUsed)
	if dec.Value != nil {
		c.Value = (*big.Int)(dec.Value)
	}
	c.Input = dec.Input
	c.Output = dec.Output
	c.Error = dec.Error
	c.RevertReason = dec.RevertReason
	c.Calls = dec.Calls
	return nil
}

type traceConfig struct {
	Tracer    string        `json:"tracer"`
	Overrides w3types.State `json:"stateOverrides,omitempty"`
}
