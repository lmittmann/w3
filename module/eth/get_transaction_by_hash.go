package eth

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/lmittmann/w3/w3types"
)

// Deprecated: TransactionByHash requests the transaction with the given hash.
//
// Use [Tx] instead.
func TransactionByHash(hash common.Hash) w3types.CallerFactory[types.Transaction] {
	return Tx(hash)
}
