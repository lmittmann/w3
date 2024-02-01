package eth

import (
	"encoding/json"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/lmittmann/w3/internal/module"
	"github.com/lmittmann/w3/w3types"
)

// BlockByHash requests the block with the given hash with full transactions.
func BlockByHash(hash common.Hash) w3types.RPCCallerFactory[types.Block] {
	return module.NewFactory(
		"eth_getBlockByHash",
		[]any{hash, true},
		module.WithRetWrapper(blockRetWrapper),
	)
}

// BlockByNumber requests the block with the given number with full
// transactions. If number is nil, the latest block is requested.
func BlockByNumber(number *big.Int) w3types.RPCCallerFactory[types.Block] {
	return module.NewFactory(
		"eth_getBlockByNumber",
		[]any{module.BlockNumberArg(number), true},
		module.WithRetWrapper(blockRetWrapper),
	)
}

// BlockTxCountByHash requests the number of transactions in the block with the
// given hash.
func BlockTxCountByHash(hash common.Hash) w3types.RPCCallerFactory[uint] {
	return module.NewFactory(
		"eth_getBlockTransactionCountByHash",
		[]any{hash},
		module.WithRetWrapper(module.HexUintRetWrapper),
	)
}

// BlockTxCountByNumber requests the number of transactions in the block with
// the given number.
func BlockTxCountByNumber(number *big.Int) w3types.RPCCallerFactory[uint] {
	return module.NewFactory(
		"eth_getBlockTransactionCountByNumber",
		[]any{module.BlockNumberArg(number)},
		module.WithRetWrapper(module.HexUintRetWrapper),
	)
}

// HeaderByHash requests the header with the given hash.
func HeaderByHash(hash common.Hash) w3types.RPCCallerFactory[types.Header] {
	return module.NewFactory[types.Header](
		"eth_getBlockByHash",
		[]any{hash, false},
	)
}

// HeaderByNumber requests the header with the given number. If number is nil,
// the latest header is requested.
func HeaderByNumber(number *big.Int) w3types.RPCCallerFactory[types.Header] {
	return module.NewFactory[types.Header](
		"eth_getBlockByNumber",
		[]any{module.BlockNumberArg(number), false},
	)
}

var blockRetWrapper = func(ret *types.Block) any { return (*rpcBlock)(ret) }

type rpcBlock types.Block

func (b *rpcBlock) UnmarshalJSON(data []byte) error {
	var header types.Header
	if err := json.Unmarshal(data, &header); err != nil {
		return err
	}

	var blockExtraData struct {
		Transactions []*types.Transaction `json:"transactions"`
		Withdrawals  []*types.Withdrawal  `json:"withdrawals"`
	}
	if err := json.Unmarshal(data, &blockExtraData); err != nil {
		return err
	}

	block := types.NewBlockWithHeader(&header).
		WithBody(blockExtraData.Transactions, nil).
		WithWithdrawals(blockExtraData.Withdrawals)
	*b = (rpcBlock)(*block)
	return nil
}
