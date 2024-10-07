package w3types_test

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/lmittmann/w3"
	"github.com/lmittmann/w3/w3types"
)

// TxBySenderAndNonceFactory requests the senders transaction hash by the nonce.
func TxBySenderAndNonceFactory(sender common.Address, nonce uint64) w3types.RPCCallerFactory[common.Hash] {
	return &getTransactionBySenderAndNonceFactory{
		sender: sender,
		nonce:  nonce,
	}
}

// getTransactionBySenderAndNonceFactory implements the w3types.RPCCaller and
// w3types.RPCCallerFactory interfaces. It stores the method parameters and
// the reference to the return value.
type getTransactionBySenderAndNonceFactory struct {
	// params
	sender common.Address
	nonce  uint64

	// returns
	returns *common.Hash
}

// Returns sets the reference to the return value.
//
// Return implements the [w3types.RPCCallerFactory] interface.
func (f *getTransactionBySenderAndNonceFactory) Returns(txHash *common.Hash) w3types.RPCCaller {
	f.returns = txHash
	return f
}

// CreateRequest creates a batch request element for the Otterscan getTransactionBySenderAndNonce method.
//
// CreateRequest implements the [w3types.RPCCaller] interface.
func (f *getTransactionBySenderAndNonceFactory) CreateRequest() (rpc.BatchElem, error) {
	return rpc.BatchElem{
		Method: "ots_getTransactionBySenderAndNonce",
		Args:   []any{f.sender, f.nonce},
		Result: f.returns,
	}, nil
}

// HandleResponse handles the response of the Otterscan getTransactionBySenderAndNonce method.
//
// HandleResponse implements the [w3types.RPCCaller] interface.
func (f *getTransactionBySenderAndNonceFactory) HandleResponse(elem rpc.BatchElem) error {
	if err := elem.Error; err != nil {
		return err
	}
	return nil
}

func ExampleRPCCaller_getTransactionBySenderAndNonce() {
	client := w3.MustDial("https://docs-demo.quiknode.pro")
	defer client.Close()

	addr := w3.A("0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045")

	var firstTxHash common.Hash
	if err := client.Call(
		TxBySenderAndNonceFactory(addr, 0).Returns(&firstTxHash),
	); err != nil {
		fmt.Printf("Request failed: %v\n", err)
		return
	}

	fmt.Printf("First Tx Hash: %s\n", firstTxHash)
	// Output:
	// First Tx Hash: 0x6ff0860e202c61189cb2a3a38286bffd694acbc50577df6cb5a7ff40e21ea074
}
