package eth

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/lmittmann/w3/core"
)

// GasPrice requests the current gas price in wei.
func GasPrice() core.CallReturnsFactory[*big.Int] {
	return &gasPriceFactory{}
}

type gasPriceFactory struct {
	// returns
	result  hexutil.Big
	returns *big.Int
}

func (f *gasPriceFactory) Returns(gasPrice *big.Int) core.Caller {
	f.returns = gasPrice
	return f
}

// CreateRequest implements the core.RequestCreator interface.
func (f *gasPriceFactory) CreateRequest() (rpc.BatchElem, error) {
	return rpc.BatchElem{
		Method: "eth_gasPrice",
		Result: &f.result,
	}, nil
}

// HandleResponse implements the core.ResponseHandler interface.
func (f *gasPriceFactory) HandleResponse(elem rpc.BatchElem) error {
	if err := elem.Error; err != nil {
		return err
	}
	f.returns.Set((*big.Int)(&f.result))
	return nil
}
