package module

import (
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common/hexutil"
)

var errNotFound = errors.New("not found")

func BlockNumberArg(blockNumber *big.Int) string {
	if blockNumber == nil {
		return "latest"
	}
	return hexutil.EncodeBig(blockNumber)
}
