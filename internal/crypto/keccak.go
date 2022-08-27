package crypto

import (
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

var keccakStatePool = sync.Pool{
	New: func() any {
		return crypto.NewKeccakState()
	},
}

// Keccak256Hash returns the Keccak256 hash of the input data as [common.Hash].
//
// Its implementation is similar to [crypto.Keccak256Hash], but it reuses the
// [crypto.KeccakState] to reduce the number of allocations.
func Keccak256Hash(data ...[]byte) (hash common.Hash) {
	// get Keccak state from pool
	d := keccakStatePool.Get().(crypto.KeccakState)

	for _, b := range data {
		d.Write(b)
	}
	d.Read(hash[:])

	// reset state and put it back into the pool
	d.Reset()
	keccakStatePool.Put(d)

	return hash
}

// Keccak256 returns the Keccak256 hash of the input data.
//
// Its implementation is similar to [crypto.Keccak256], but it reuses the
// [crypto.KeccakState] to reduce the number of allocations.
func Keccak256(data ...[]byte) (hash []byte) {
	return Keccak256Hash(data...).Bytes()
}
