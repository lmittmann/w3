package w3vm

import (
	"errors"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/lmittmann/w3/w3types"
)

// Receipt represents the result of an applied [w3types.Message].
type Receipt struct {
	f w3types.Func // Func of corresponding message

	GasUsed         uint64          // Gas used by the message
	GasLimit        uint64          // Minimum required gas limit (gas used + gas refund)
	Logs            []*types.Log    // Logs emitted by the message
	Output          []byte          // Output bytes of the applied message
	ContractAddress *common.Address // Contract address created by a contract creation transaction

	Err error // Revert reason
}

func (r Receipt) DecodeReturns(returns ...any) error {
	if r.Err != nil {
		return r.Err
	}
	if r.f == nil {
		return errors.New("no function")
	}
	return r.f.DecodeReturns(r.Output, returns...)
}
