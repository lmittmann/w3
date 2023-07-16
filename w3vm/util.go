package w3vm

import "github.com/ethereum/go-ethereum/common"

// zero values
var (
	addr0 common.Address
	hash0 common.Hash
	uint0 uint64
)

// nilToZero converts sets a pointer to the zero value if it is nil.
func nilToZero[T any](ptr *T) *T {
	if ptr == nil {
		return new(T)
	}
	return ptr
}
