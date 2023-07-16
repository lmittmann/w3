package state

import (
	"encoding/hex"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/holiman/uint256"
	"github.com/lmittmann/w3/internal/module"
	"github.com/lmittmann/w3/w3types"
)

var hash0 common.Hash

// ethBalance is like [eth.Balance], but returns the balance as [uint256.Int].
func ethBalance(addr common.Address, blockNumber *big.Int) w3types.CallerFactory[uint256.Int] {
	return module.NewFactory[uint256.Int](
		"eth_getBalance",
		[]any{addr, module.BlockNumberArg(blockNumber)},
	)
}

// ethStorageAt is like [eth.StorageAt], but returns the storage value as [uint256.Int].
func ethStorageAt(addr common.Address, slot uint256.Int, blockNumber *big.Int) w3types.CallerFactory[uint256.Int] {
	return module.NewFactory[uint256.Int](
		"eth_getStorageAt",
		[]any{addr, &slot, module.BlockNumberArg(blockNumber)},
		module.WithRetWrapper(func(ret *uint256.Int) any { return (*uint256OrHash)(ret) }),
	)
}

// uint256OrHash is like [uint256.Int], but can be unmarshaled from a hex number
// with leading zeros.
type uint256OrHash uint256.Int

func (i *uint256OrHash) UnmarshalText(text []byte) error {
	if len(text) >= 2 && text[0] == '0' && (text[1] == 'x' || text[1] == 'X') {
		text = text[2:]
	}

	if len(text)%2 != 0 {
		text = append([]byte{'0'}, text...)
	}
	buf := make([]byte, hex.DecodedLen(len(text)))
	if _, err := hex.Decode(buf, text); err != nil {
		return err
	}

	(*uint256.Int)(i).SetBytes(buf)
	return nil
}

// noopKeyValueStore implements a [ethdb.KeyValueStore] that does nothing.
type noopKeyValueStore struct {
	ethdb.KeyValueStore
}

// ethdb.KeyValueReader methods
func (noopKeyValueStore) Get(key []byte) ([]byte, error) { panic("not implemented") }
func (noopKeyValueStore) Has(key []byte) (bool, error)   { panic("not implemented") }

// ethdb.KeyValueWriter methods
func (noopKeyValueStore) Put(key []byte, value []byte) error { panic("not implemented") }
func (noopKeyValueStore) Delete(key []byte) error            { panic("not implemented") }

// ethdb.KeyValueStater methods
func (noopKeyValueStore) Stat(property string) (string, error) { panic("not implemented") }

// ethdb.Batcher methods
func (noopKeyValueStore) NewBatch() ethdb.Batch                 { return new(noopBatch) }
func (noopKeyValueStore) NewBatchWithSize(size int) ethdb.Batch { return new(noopBatch) }

// ethdb.Iteratee methods
func (noopKeyValueStore) NewIterator(prefix []byte, start []byte) ethdb.Iterator {
	panic("not implemented")
}

// eth.Compacter methods
func (noopKeyValueStore) Compact(start []byte, limit []byte) error { panic("not implemented") }

// ethdb.Snapshotter methods
func (noopKeyValueStore) Snapshot() int { panic("not implemented") }

// io.Closer methods
func (noopKeyValueStore) Close() error { return nil }

// noopBatch implements a [ethdb.Batch] that does nothing.
type noopBatch struct {
	ethdb.Batch
}

func (noopBatch) ValueSize() int                      { return 0 }
func (noopBatch) Write() error                        { return nil }
func (noopBatch) Reset()                              {}
func (noopBatch) Replay(w ethdb.KeyValueWriter) error { return nil }
