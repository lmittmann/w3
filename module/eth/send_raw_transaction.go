package eth

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/lmittmann/w3/core"
)

// Deprecated: SendTransaction sends a signed transaction to the network.
//
// Use [SendTx] instead.
func SendTransaction(tx *types.Transaction) core.CallerFactory[common.Hash] {
	return SendTx(tx)
}

// Deprecated: SendRawTransaction sends a raw transaction to the network.
//
// Use [SendRawTx] instead.
func SendRawTransaction(rawTx []byte) core.CallerFactory[common.Hash] {
	return SendRawTx(rawTx)
}
