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
func BlockByHash(hash common.Hash) w3types.CallerFactory[types.Block] {
	return module.NewFactory(
		"eth_getBlockByHash",
		[]any{hash, true},
		module.WithRetWrapper(blockRetWrapper),
	)
}

// BlockByNumber requests the block with the given number with full
// transactions. If number is nil, the latest block is requested.
func BlockByNumber(number *big.Int) w3types.CallerFactory[types.Block] {
	return module.NewFactory(
		"eth_getBlockByNumber",
		[]any{module.BlockNumberArg(number), true},
		module.WithRetWrapper(blockRetWrapper),
	)
}

// BlockTxCountByHash requests the number of transactions in the block with the
// given hash.
func BlockTxCountByHash(hash common.Hash) w3types.CallerFactory[uint] {
	return module.NewFactory(
		"eth_getBlockTransactionCountByHash",
		[]any{hash},
		module.WithRetWrapper(module.HexUintRetWrapper),
	)
}

// BlockTxCountByNumber requests the number of transactions in the block with
// the given number.
func BlockTxCountByNumber(number *big.Int) w3types.CallerFactory[uint] {
	return module.NewFactory(
		"eth_getBlockTransactionCountByNumber",
		[]any{module.BlockNumberArg(number)},
		module.WithRetWrapper(module.HexUintRetWrapper),
	)
}

// HeaderByHash requests the header with the given hash.
func HeaderByHash(hash common.Hash) w3types.CallerFactory[types.Header] {
	return module.NewFactory[types.Header](
		"eth_getBlockByHash",
		[]any{hash, false},
	)
}

// HeaderByNumber requests the header with the given number. If number is nil,
// the latest header is requested.
func HeaderByNumber(number *big.Int) w3types.CallerFactory[types.Header] {
	return module.NewFactory[types.Header](
		"eth_getBlockByNumber",
		[]any{module.BlockNumberArg(number), false},
	)
}

var blockRetWrapper = func(ret *types.Block) any { return (*rpcBlock)(ret) }

type rpcBlock types.Block

func (b *rpcBlock) UnmarshalJSON(data []byte) error {
	type rpcBlockTxs struct {
		Transactions []*types.Transaction `json:"transactions"`
	}

	type rpcBlockWithrawals struct {
		Withdrawals []*types.Withdrawal `json:"withdrawals"`
	}

	var header types.Header
	if err := json.Unmarshal(data, &header); err != nil {
		return err
	}

	var blockTxs rpcBlockTxs
	if err := json.Unmarshal(data, &blockTxs); err != nil {
		return err
	}

	var blockWithdrawals rpcBlockWithrawals
	if err := json.Unmarshal(data, &blockWithdrawals); err != nil {
		return err
	}

	block := types.NewBlockWithHeader(&header).
		WithBody(blockTxs.Transactions, nil).
		WithWithdrawals(blockWithdrawals.Withdrawals)
	*b = (rpcBlock)(*block)
	return nil
}
