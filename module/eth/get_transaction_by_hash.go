package eth

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/lmittmann/w3/core"
)

// Deprecated: TransactionByHash requests the transaction with the given hash.
//
// Use [Tx] instead.
func TransactionByHash(hash common.Hash) core.CallerFactory[types.Transaction] {
	return Tx(hash)
}
