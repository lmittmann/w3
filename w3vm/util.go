package w3vm

import (
	"crypto/rand"

	"github.com/ethereum/go-ethereum/common"
)

// zero values
var (
	addr0 common.Address
	hash0 common.Hash
	uint0 uint64
)

// RandA returns a random address.
func RandA() (addr common.Address) {
	rand.Read(addr[:])
	return addr
}

// nilToZero converts sets a pointer to the zero value if it is nil.
func nilToZero[T any](ptr *T) *T {
	if ptr == nil {
		return new(T)
	}
	return ptr
}

// zeroHashFunc implements a [vm.GetHashFunc] that always returns the zero hash.
func zeroHashFunc(uint64) common.Hash {
	return hash0
}
