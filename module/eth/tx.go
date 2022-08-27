package eth

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/lmittmann/w3/core"
	"github.com/lmittmann/w3/internal/module"
)

// Tx requests the transaction with the given hash.
func Tx(hash common.Hash) core.CallerFactory[types.Transaction] {
	return module.NewFactory[types.Transaction](
		"eth_getTransactionByHash",
		[]any{hash},
	)
}

// TxByBlockHashAndIndex requests the transaction in the given block with the given index.
func TxByBlockHashAndIndex(blockHash common.Hash, index uint64) core.CallerFactory[types.Transaction] {
	return module.NewFactory[types.Transaction](
		"eth_getTransactionByBlockHashAndIndex",
		[]any{blockHash, hexutil.Uint64(index)},
	)
}

// TxByBlockNumberAndIndex requests the transaction in the given block with the given index.
func TxByBlockNumberAndIndex(blockNumber *big.Int, index uint64) core.CallerFactory[types.Transaction] {
	return module.NewFactory[types.Transaction](
		"eth_getTransactionByBlockNumberAndIndex",
		[]any{module.BlockNumberArg(blockNumber), hexutil.Uint64(index)},
	)
}

// SendRawTx sends a raw transaction to the network and returns its hash.
func SendRawTx(rawTx []byte) core.CallerFactory[common.Hash] {
	return module.NewFactory[common.Hash](
		"eth_sendRawTransaction",
		[]any{hexutil.Encode(rawTx)},
	)
}

// SendTx sends a signed transaction to the network and returns its hash.
func SendTx(tx *types.Transaction) core.CallerFactory[common.Hash] {
	return module.NewFactory(
		"eth_sendRawTransaction",
		[]any{tx},
		module.WithArgsWrapper[common.Hash](func(args []any) ([]any, error) {
			tx := (args[0]).(*types.Transaction)

			rawTx, err := tx.MarshalBinary()
			if err != nil {
				return nil, err
			}
			return []any{hexutil.Encode(rawTx)}, nil
		}),
	)
}

// TxReceipt requests the receipt of the transaction with the given hash.
func TxReceipt(txHash common.Hash) core.CallerFactory[types.Receipt] {
	return module.NewFactory[types.Receipt](
		"eth_getTransactionReceipt",
		[]any{txHash},
	)
}

// Nonce requests the nonce of the given common.Address addr at the given
// blockNumber. If blockNumber is nil, the nonce at the latest known block is
// requested.
func Nonce(addr common.Address, blockNumber *big.Int) core.CallerFactory[uint64] {
	return module.NewFactory(
		"eth_getTransactionCount",
		[]any{addr, module.BlockNumberArg(blockNumber)},
		module.WithRetWrapper(module.HexUint64RetWrapper),
	)
}
