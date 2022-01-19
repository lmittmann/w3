// Package eth implements RPC API bindings to methods in the "eth" namespace.
package eth

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common/hexutil"
)

func toBlockNumberArg(blockNumber *big.Int) string {
	if blockNumber == nil {
		return "latest"
	} else if blockNumber.Sign() < 0 {
		return "pending"
	}
	return hexutil.EncodeBig(blockNumber)
}
