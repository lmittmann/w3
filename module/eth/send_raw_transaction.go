package eth

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/lmittmann/w3/w3types"
)

// Deprecated: SendTransaction sends a signed transaction to the network.
//
// Use [SendTx] instead.
func SendTransaction(tx *types.Transaction) w3types.CallerFactory[common.Hash] {
	return SendTx(tx)
}

// Deprecated: SendRawTransaction sends a raw transaction to the network.
//
// Use [SendRawTx] instead.
func SendRawTransaction(rawTx []byte) w3types.CallerFactory[common.Hash] {
	return SendRawTx(rawTx)
}
