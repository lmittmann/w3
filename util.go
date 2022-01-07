package w3

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

var (
	Big0     = big.NewInt(0)
	Big1     = big.NewInt(1)
	Big2     = big.NewInt(2)
	BigGwei  = big.NewInt(1_000000000)
	BigEther = big.NewInt(1_000000000_000000000)

	addrZero = common.Address{} // zero address 0x00…00
)

// A returns an address from a hexstring. It is short for common.HexToAddress(…).
func A(hexAddress string) common.Address {
	return common.HexToAddress(hexAddress)
}

// APtr returns an address pointer from a hexstring. The returned address is nil, if the hexstring
// address equals the zero address 0x00…00.
func APtr(hexAddress string) *common.Address {
	addr := A(hexAddress)
	if addr == addrZero {
		return nil
	}
	return &addr
}

// B returns a byte slice from a hexstring. It is short for common.FromHex(…).
func B(hexBytes string) []byte {
	return common.FromHex(hexBytes)
}

// H returns a hash from a hexstring. It is short for common.HexToHash(…).
func H(hexHash string) common.Hash {
	return common.HexToHash(hexHash)
}

// I returns a big.Int from a number string in decimal or hex format. Nil is returned if the number
// parsing fails.
func I(strInt string) *big.Int {
	var base int
	if len(strInt) >= 2 && strInt[0] == '0' && (strInt[1] == 'x' || strInt[1] == 'X') {
		strInt = strInt[2:]
		base = 16
	} else {
		base = 10
	}

	bigint, ok := new(big.Int).SetString(strInt, base)
	if !ok {
		return nil
	}
	return bigint
}

// Keccak returns the Keccak256 hash of data. It is short for crypto.Keccak256Hash(…)
func Keccak(data []byte) common.Hash {
	return crypto.Keccak256Hash(data)
}
