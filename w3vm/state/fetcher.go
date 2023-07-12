package state

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// Fetcher is the interface to access account state of a blockchain.
type Fetcher interface {

	// Nonce fetches the nonce of the given address.
	Nonce(common.Address) (uint64, error)

	// Balance fetches the balance of the given address.
	Balance(common.Address) (*big.Int, error)

	// Code fetches the code of the given address.
	Code(common.Address) ([]byte, error)

	// StorageAt fetches the state of the given address and storage slot.
	StorageAt(common.Address, common.Hash) (common.Hash, error)

	// HeaderHash fetches the hash of the header with the given number.
	HeaderHash(*big.Int) (common.Hash, error)
}
