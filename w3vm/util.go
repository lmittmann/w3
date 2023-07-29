package w3vm

import (
	"crypto/rand"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/lmittmann/w3/internal/crypto"
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

var (
	weth9BalancePos   = big.NewInt(3)
	weth9AllowancePos = big.NewInt(4)
)

// WETHBalanceSlot returns the storage slot that stores the WETH balance of
// the given addr.
func WETHBalanceSlot(addr common.Address) common.Hash {
	return slot(weth9BalancePos, addr)
}

// WETHAllowanceSlot returns the storage slot that stores the WETH allowance
// of the given owner and spender.
func WETHAllowanceSlot(owner, spender common.Address) common.Hash {
	return slot2(weth9AllowancePos, owner, spender)
}

func slot(pos *big.Int, acc common.Address) common.Hash {
	data := make([]byte, 64)
	copy(data[12:32], acc[:])
	pos.FillBytes(data[32:])

	return crypto.Keccak256Hash(data)
}

func slot2(pos *big.Int, acc, acc2 common.Address) common.Hash {
	data := make([]byte, 64)
	copy(data[12:32], acc[:])
	pos.FillBytes(data[32:])

	copy(data[32:], crypto.Keccak256(data))
	copy(data[12:32], acc2[:])

	return crypto.Keccak256Hash(data)
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
