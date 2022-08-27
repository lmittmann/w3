package eth

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/lmittmann/w3/core"
)

// Deprecated: TransactionReceipt requests the receipt of the transaction with
// the given hash.
//
// Use [TxReceipt] instead.
func TransactionReceipt(hash common.Hash) core.CallerFactory[types.Receipt] {
	return TxReceipt(hash)
}
