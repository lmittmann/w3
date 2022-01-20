package eth

import (
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"
)

func Call(msg ethereum.CallMsg) *CallFactory {
	return &CallFactory{msg: msg}
}

type CallFactory struct {
	// args
	msg     ethereum.CallMsg
	atBlock *big.Int

	// returns
	result  hexutil.Bytes
	returns *[]byte
}

func (f *CallFactory) AtBlock(blockNumber *big.Int) *CallFactory {
	f.atBlock = blockNumber
	return f
}

func (f *CallFactory) Returns(output *[]byte) *CallFactory {
	f.returns = output
	return f
}

func (f *CallFactory) CreateRequest() (rpc.BatchElem, error) {
	return rpc.BatchElem{
		Method: "eth_call",
		Args:   []interface{}{toCallArg(f.msg), toBlockNumberArg(f.atBlock)},
		Result: &f.result,
	}, nil
}

func (f *CallFactory) HandleResponse(elem rpc.BatchElem) error {
	if err := elem.Error; err != nil {
		return err
	}
	*f.returns = []byte(f.result)
	return nil
}

func toCallArg(msg ethereum.CallMsg) interface{} {
	arg := map[string]interface{}{
		"from": msg.From,
		"to":   msg.To,
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
