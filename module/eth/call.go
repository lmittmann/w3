package eth

import (
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/lmittmann/w3/core"
	"github.com/lmittmann/w3/internal/inline"
)

// Call requests the output data of the given message at the given blockNumber.
// If blockNumber is nil, the output of the message at the latest known block is
// requested.
func Call(msg ethereum.CallMsg, blockNumber *big.Int, overrides AccountOverrides) core.CallFactoryReturns[[]byte] {
	return &callFactory{msg: msg, atBlock: blockNumber, overrides: overrides}
}

type callFactory struct {
	// args
	msg       ethereum.CallMsg
	atBlock   *big.Int
	overrides AccountOverrides

	// returns
	returns *[]byte
}

func (f *callFactory) Returns(output *[]byte) core.Caller {
	f.returns = output
	return f
}

// CreateRequest implements the core.RequestCreator interface.
func (f *callFactory) CreateRequest() (rpc.BatchElem, error) {
	return rpc.BatchElem{
		Method: "eth_call",
		Args: inline.If(f.overrides == nil,
			[]any{toCallArg(f.msg), toBlockNumberArg(f.atBlock)},
			[]any{toCallArg(f.msg), toBlockNumberArg(f.atBlock), f.overrides},
		),
		Result: (*hexutil.Bytes)(f.returns),
	}, nil
}

// HandleResponse implements the core.ResponseHandler interface.
func (f *callFactory) HandleResponse(elem rpc.BatchElem) error {
	if err := elem.Error; err != nil {
		return err
	}
	return nil
}

// CallFunc requests the returns of Func fn at common.Address contract with the
// given args.
func CallFunc(fn core.Func, contract common.Address, args ...any) *CallFuncFactory {
	return &CallFuncFactory{fn: fn, contract: contract, args: args}
}

type CallFuncFactory struct {
	// args
	fn        core.Func
	contract  common.Address
	args      []any
	from      *common.Address
	atBlock   *big.Int
	value     *big.Int
	overrides AccountOverrides

	// returns
	result  hexutil.Bytes
	returns []any
}

func (f *CallFuncFactory) AtBlock(blockNumber *big.Int) *CallFuncFactory {
	f.atBlock = blockNumber
	return f
}

func (f *CallFuncFactory) Value(value *big.Int) *CallFuncFactory {
	f.value = value
	return f
}

func (f *CallFuncFactory) Returns(returns ...any) core.Caller {
	f.returns = returns
	return f
}

func (f *CallFuncFactory) From(from common.Address) *CallFuncFactory {
	f.from = &from
	return f
}

func (f *CallFuncFactory) Overrides(overrides AccountOverrides) *CallFuncFactory {
	f.overrides = overrides
	return f
}

// CreateRequest implements the core.RequestCreator interface.
func (f *CallFuncFactory) CreateRequest() (rpc.BatchElem, error) {
	input, err := f.fn.EncodeArgs(f.args...)
	if err != nil {
		return rpc.BatchElem{}, err
	}

	msg := ethereum.CallMsg{
		To:   &f.contract,
		Data: input,
	}
	if f.from != nil {
		msg.From = *f.from
	}

	return rpc.BatchElem{
		Method: "eth_call",
		Args: inline.If(f.overrides == nil,
			[]any{toCallArg(msg), toBlockNumberArg(f.atBlock)},
			[]any{toCallArg(msg), toBlockNumberArg(f.atBlock), f.overrides},
		),
		Result: &f.result,
	}, nil
}

// HandleResponse implements the core.ResponseHandler interface.
func (f *CallFuncFactory) HandleResponse(elem rpc.BatchElem) error {
	if err := elem.Error; err != nil {
		return err
	}
	output := []byte(f.result)
	if err := f.fn.DecodeReturns(output, f.returns...); err != nil {
		return err
	}
	return nil
}

func toCallArg(msg ethereum.CallMsg) any {
	arg := map[string]any{
		"to": msg.To,
	}
	if msg.From.Hash().Big().Sign() > 0 {
		arg["from"] = msg.From
	}
	if len(msg.Data) > 0 {
		arg["data"] = hexutil.Bytes(msg.Data)
	}
	if msg.Value != nil {
		arg["value"] = (*hexutil.Big)(msg.Value)
	}
	if msg.Gas != 0 {
		arg["gas"] = hexutil.Uint64(msg.Gas)
	}
	if msg.GasPrice != nil {
		arg["gasPrice"] = (*hexutil.Big)(msg.GasPrice)
	}
	return arg
}
