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

	GasUsed         uint64          // Gas used for executing the message
	GasRefund       uint64          // Gas refunded after executing the message
	GasLimit        uint64          // Deprecated: Minimum required gas limit (gas used without refund)
	Logs            []*types.Log    // Logs emitted by the message
	Output          []byte          // Output bytes of the applied message
	ContractAddress *common.Address // Contract address created by a contract creation transaction

	Err error // Revert reason
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
