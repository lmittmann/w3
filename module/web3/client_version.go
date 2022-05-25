package web3

import (
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/lmittmann/w3/core"
)

// ClientVersion requests the endpoints client version.
func ClientVersion() core.CallerFactory[string] {
	return &clientVersionFactory{}
}

type clientVersionFactory struct {
	// returns
	returns *string
}

func (f *clientVersionFactory) Returns(clientVersion *string) core.Caller {
	f.returns = clientVersion
	return f
}

func (f *clientVersionFactory) CreateRequest() (rpc.BatchElem, error) {
	return rpc.BatchElem{
		Method: "web3_clientVersion",
		Result: f.returns,
	}, nil
}

func (f *clientVersionFactory) HandleResponse(elem rpc.BatchElem) error {
	if err := elem.Error; err != nil {
		return err
	}
	return nil
}
