package debug

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/holiman/uint256"
	"github.com/lmittmann/w3/internal/hexutil"
	"github.com/lmittmann/w3/internal/module"
	"github.com/lmittmann/w3/w3types"
)

// TraceCall requests the trace of the given message.
func TraceCall(msg *w3types.Message, blockNumber *big.Int, config *TraceConfig) w3types.RPCCallerFactory[*Trace] {
	if config == nil {
		config = &TraceConfig{}
	}
	return module.NewFactory(
		"debug_traceCall",
		[]any{msg, module.BlockNumberArg(blockNumber), config},
		module.WithArgsWrapper[*Trace](msgArgsWrapper),
	)
}

// TraceTx requests the trace of the transaction with the given hash.
func TraceTx(txHash common.Hash, config *TraceConfig) w3types.RPCCallerFactory[*Trace] {
	if config == nil {
		config = &TraceConfig{}
	}
	return module.NewFactory[*Trace](
		"debug_traceTransaction",
		[]any{txHash, config},
	)
}

type TraceConfig struct {
	Overrides      w3types.State           // Override account state
	BlockOverrides *w3types.BlockOverrides // Override block state
	EnableStack    bool                    // Enable stack capture
	EnableMemory   bool                    // Enable memory capture
	EnableStorage  bool                    // Enable storage capture
	Limit          uint64                  // Maximum number of StructLog's to capture (all if zero)
}

// MarshalJSON implements the [json.Marshaler].
func (c *TraceConfig) MarshalJSON() ([]byte, error) {
	type config struct {
		Overrides        w3types.State           `json:"stateOverrides,omitempty"`
		BlockOverrides   *w3types.BlockOverrides `json:"blockOverrides,omitempty"`
		DisableStorage   bool                    `json:"disableStorage,omitempty"`
		DisableStack     bool                    `json:"disableStack,omitempty"`
		EnableMemory     bool                    `json:"enableMemory,omitempty"`
		EnableReturnData bool                    `json:"enableReturnData,omitempty"`
		Limit            uint64                  `json:"limit,omitempty"`
	}

	return json.Marshal(config{
		Overrides:        c.Overrides,
		BlockOverrides:   c.BlockOverrides,
		DisableStorage:   !c.EnableStorage,
		DisableStack:     !c.EnableStack,
		EnableMemory:     c.EnableMemory,
		EnableReturnData: true,
		Limit:            c.Limit,
	})
}

type Trace struct {
	Gas        uint64       `json:"gas"`
	Failed     bool         `json:"failed"`
	Output     []byte       `json:"returnValue"`
	StructLogs []*StructLog `json:"structLogs"`
}

func (t *Trace) UnmarshalJSON(data []byte) error {
	type trace struct {
		Gas        uint64        `json:"gas"`
		Failed     bool          `json:"failed"`
		Output     hexutil.Bytes `json:"returnValue"`
		StructLogs []*StructLog  `json:"structLogs"`
	}

	var dec trace
	if err := json.Unmarshal(data, &dec); err != nil {
		return err
	}

	t.Gas = dec.Gas
	t.Failed = dec.Failed
	t.Output = dec.Output
	t.StructLogs = dec.StructLogs
	return nil
}

type StructLog struct {
	Pc      uint64
	Depth   uint
	Gas     uint64
	GasCost uint
	Op      vm.OpCode
	Stack   []uint256.Int
	Memory  []byte
	Storage w3types.Storage
}

func (l *StructLog) UnmarshalJSON(data []byte) error {
	type structLog struct {
		Pc      uint64
		Depth   uint
		Gas     uint64
		GasCost uint
		Op      string
		Stack   []uint256.Int
		Memory  memory
		Storage map[optionalPrefixedHash]optionalPrefixedHash
	}

	var dec structLog
	if err := json.Unmarshal(data, &dec); err != nil {
		return err
	}

	l.Pc = dec.Pc
	l.Depth = dec.Depth
	l.Gas = dec.Gas
	l.GasCost = dec.GasCost
	l.Op = vm.StringToOp(dec.Op)
	l.Stack = dec.Stack
	l.Memory = dec.Memory

	if len(dec.Storage) > 0 {
		l.Storage = make(w3types.Storage, len(dec.Storage))
		for k, v := range dec.Storage {
			l.Storage[(common.Hash)(k)] = (common.Hash)(v)
		}
	}
	return nil
}

// optionalPrefixedHash is a helper for unmarshaling hashes with or without
// "0x"-prefix.
type optionalPrefixedHash common.Hash

func (h *optionalPrefixedHash) UnmarshalText(data []byte) error {
	if len(data) > 2 && data[0] == '0' && (data[1] == 'x' || data[1] == 'X') {
		data = data[2:]
	}

	if len(data) != 2*common.HashLength {
		return fmt.Errorf("hex string has length %d, want 64", len(data))
	}

	_, err := hex.Decode((*h)[:], data)
	return err
}

type memory []byte

func (m *memory) UnmarshalJSON(data []byte) error {
	var dec []optionalPrefixedHash
	if err := json.Unmarshal(data, &dec); err != nil {
		return err
	}

	*m = make([]byte, 0, 32*len(dec))
	for _, data := range dec {
		*m = append(*m, data[:]...)
	}
	return nil
}

func msgArgsWrapper(slice []any) ([]any, error) {
	msg := slice[0].(*w3types.Message)
	if msg.Input != nil || msg.Func == nil {
		return slice, nil
	}

	input, err := msg.Func.EncodeArgs(msg.Args...)
	if err != nil {
		return nil, err
	}
	msg.Input = input
	slice[0] = msg
	return slice, nil
}
