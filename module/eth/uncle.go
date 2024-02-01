package eth

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/lmittmann/w3/internal/module"
	"github.com/lmittmann/w3/w3types"
)

// UncleByBlockHashAndIndex requests the uncle of the block with the given hash
// at the given index.
func UncleByBlockHashAndIndex(hash common.Hash, index uint) w3types.RPCCallerFactory[types.Header] {
	return module.NewFactory[types.Header](
		"eth_getUncleByBlockHashAndIndex",
		[]any{hash, hexutil.Uint(index)},
	)
}

// UncleByBlockNumberAndIndex requests the uncle of the block with the given
// number at the given index.
func UncleByBlockNumberAndIndex(number *big.Int, index uint) w3types.RPCCallerFactory[types.Header] {
	return module.NewFactory[types.Header](
		"eth_getUncleByBlockNumberAndIndex",
		[]any{module.BlockNumberArg(number), hexutil.Uint(index)},
	)
}

// UncleCountByBlockHash requests the number of uncles of the block with the
// given hash.
func UncleCountByBlockHash(hash common.Hash) w3types.RPCCallerFactory[uint] {
	return module.NewFactory(
		"eth_getUncleCountByBlockHash",
		[]any{hash},
		module.WithRetWrapper(module.HexUintRetWrapper),
	)
}

// UncleCountByBlockNumber requests the number of uncles of the block with the
// given number.
func UncleCountByBlockNumber(number *big.Int) w3types.RPCCallerFactory[uint] {
	return module.NewFactory(
		"eth_getUncleCountByBlockNumber",
		[]any{module.BlockNumberArg(number)},
		module.WithRetWrapper(module.HexUintRetWrapper),
	)
}
