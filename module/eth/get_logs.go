package eth

import (
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"
)

func Logs(filterQuery ethereum.FilterQuery) *LogsFactory {
	return &LogsFactory{filterQuery: filterQuery}
}

type LogsFactory struct {
	// args
	filterQuery ethereum.FilterQuery

	// returns
	result  []types.Log
	returns *[]types.Log
}

func (f *LogsFactory) Returns(logs *[]types.Log) *LogsFactory {
	f.returns = logs
	return f
}

func (f *LogsFactory) CreateRequest() (rpc.BatchElem, error) {
	arg, err := toFilterArg(f.filterQuery)
	if err != nil {
		return rpc.BatchElem{}, err
	}

	return rpc.BatchElem{
		Method: "eth_getLogs",
		Args:   []interface{}{arg},
		Result: &f.result,
	}, nil
}

func (f *LogsFactory) HandleResponse(elem rpc.BatchElem) error {
	if err := elem.Error; err != nil {
		return err
	}
	*f.returns = f.result
	return nil
}
