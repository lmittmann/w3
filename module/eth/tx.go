package eth

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/lmittmann/w3/internal/module"
	"github.com/lmittmann/w3/w3types"
)

// Tx requests the transaction with the given hash.
func Tx(hash common.Hash) w3types.CallerFactory[types.Transaction] {
	return module.NewFactory[types.Transaction](
		"eth_getTransactionByHash",
		[]any{hash},
	)
}

// TxByBlockHashAndIndex requests the transaction in the given block with the given index.
func TxByBlockHashAndIndex(blockHash common.Hash, index uint64) w3types.CallerFactory[types.Transaction] {
	return module.NewFactory[types.Transaction](
		"eth_getTransactionByBlockHashAndIndex",
		[]any{blockHash, hexutil.Uint64(index)},
	)
}

// TxByBlockNumberAndIndex requests the transaction in the given block with the given index.
func TxByBlockNumberAndIndex(blockNumber *big.Int, index uint64) w3types.CallerFactory[types.Transaction] {
	return module.NewFactory[types.Transaction](
		"eth_getTransactionByBlockNumberAndIndex",
		[]any{module.BlockNumberArg(blockNumber), hexutil.Uint64(index)},
	)
}

// SendRawTx sends a raw transaction to the network and returns its hash.
func SendRawTx(rawTx []byte) w3types.CallerFactory[common.Hash] {
	return module.NewFactory[common.Hash](
		"eth_sendRawTransaction",
		[]any{hexutil.Encode(rawTx)},
	)
}

// SendTx sends a signed transaction to the network and returns its hash.
func SendTx(tx *types.Transaction) w3types.CallerFactory[common.Hash] {
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
func TxReceipt(txHash common.Hash) w3types.CallerFactory[types.Receipt] {
	return module.NewFactory[types.Receipt](
		"eth_getTransactionReceipt",
		[]any{txHash},
	)
}

// BlockReceipts requests all receipts of the transactions in the given block.
func BlockReceipts(number *big.Int) w3types.CallerFactory[types.Receipts] {
	return module.NewFactory[types.Receipts](
		"eth_getBlockReceipts",
		[]any{module.BlockNumberArg(number)},
	)
}

// Nonce requests the nonce of the given common.Address addr at the given
// blockNumber. If blockNumber is nil, the nonce at the latest known block is
// requested.
func Nonce(addr common.Address, blockNumber *big.Int) w3types.CallerFactory[uint64] {
	return module.NewFactory(
		"eth_getTransactionCount",
		[]any{addr, module.BlockNumberArg(blockNumber)},
		module.WithRetWrapper(module.HexUint64RetWrapper),
	)
}
