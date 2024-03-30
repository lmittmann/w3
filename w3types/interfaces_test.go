package w3types_test

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/lmittmann/w3"
	"github.com/lmittmann/w3/w3types"
)

func TxBySenderAndNonceFactory(sender common.Address, nonce uint64) w3types.RPCCallerFactory[common.Hash] {
	return &getTransactionBySenderAndNonceFactory{
		sender: sender,
		nonce:  nonce,
	}
}

type getTransactionBySenderAndNonceFactory struct {
	// params
	sender common.Address
	nonce  uint64

	// returns
	returns *common.Hash
}

func (f *getTransactionBySenderAndNonceFactory) Returns(txHash *common.Hash) w3types.RPCCaller {
	f.returns = txHash
	return f
}

func (f *getTransactionBySenderAndNonceFactory) CreateRequest() (rpc.BatchElem, error) {
	return rpc.BatchElem{
		Method: "ots_getTransactionBySenderAndNonce",
		Args:   []any{f.sender, f.nonce},
		Result: f.returns,
	}, nil
}

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
