package w3vm

import (
	"errors"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/lmittmann/w3/w3types"
)

var ErrMissingFunc = errors.New("missing function")

// Receipt represents the result of an applied [w3types.Message].
type Receipt struct {
	f w3types.Func // Func of corresponding message

	GasUsed         uint64          // Gas used for executing the message (including refunds)
	MaxGasUsed      uint64          // Maximum gas used during executing the message (excluding refunds)
	Logs            []*types.Log    // Logs emitted while executing the message
	Output          []byte          // Output of the executed message
	ContractAddress *common.Address // Address of the created contract, if any

	Err error // Execution error, if any
}

// DecodeReturns is like [w3types.Func.DecodeReturns], but returns [ErrMissingFunc]
// if the underlying [w3types.Message.Func] is nil.
func (r Receipt) DecodeReturns(returns ...any) error {
	if r.Err != nil {
		return r.Err
	}
	if r.f == nil {
		return ErrMissingFunc
	}
	return r.f.DecodeReturns(r.Output, returns...)
}
