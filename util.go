package w3

import (
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// Common big.Int's.
var (
	Big0     = big.NewInt(0)
	Big1     = big.NewInt(1)
	Big2     = big.NewInt(2)
	BigGwei  = big.NewInt(1_000000000)
	BigEther = big.NewInt(1_000000000_000000000)
)

var addrZero = common.Address{} // zero address 0x00…00

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

// I returns a big.Int from a number string in decimal or hex format. Nil is
// returned if the number parsing fails.
//
// I supports the units "ether" or "eth" and "gwei" for decimal number strings.
// E.g.:
//     w3.I("1 ether")   -> 1000000000000000000
//     w3.I("1.2 ether") -> 1200000000000000000
//
// Fractional digits that exeed the units maximum number of fractional digits
// are ignored. E.g.:
//     w3.I("0.000000123456 gwei") -> 123
func I(strInt string) *big.Int {
	if len(strInt) >= 2 && strInt[0] == '0' && (strInt[1] == 'x' || strInt[1] == 'X') {
		// hex int
		bigint, ok := new(big.Int).SetString(strInt[2:], 16)
		if !ok {
			return nil
		}
		return bigint
	}

	// decimal int
	return parseDecimal(strInt)
}

func parseDecimal(s string) *big.Int {
	var (
		state int // parse state: 0=int, 1=frac

		intEnd    int
		fracStart int
		fracEnd   int
		unitStart int

		decimals int
		intPart  = new(big.Int)
		fracPart = new(big.Int)
	)

	// find ranges of int, frac, and unit parts
Outer:
	for i, c := range s {
		switch state {
		case 0:
			if c == '.' {
				fracStart = i + 1
				fracEnd = i + 1
				state++
				break
			} else if c == ' ' {
				unitStart = i + 1
				break Outer
			} else if c < '0' || '9' < c {
				return nil // invalid char
			}
			intEnd = i + 1
		case 1:
			if c == ' ' {
				unitStart = i + 1
				break Outer
			} else if c < '0' || '9' < c {
				return nil // invalid char
			}
			fracEnd = i + 1
		}
	}
	// set parts
	if unitStart < len(s) && unitStart > fracEnd {
		switch unitPart := strings.ToLower(s[unitStart:]); unitPart {
		case "ether", "eth":
			decimals = 18
		case "gwei":
			decimals = 9
		case "":
		default:
			return nil
		}
	}

	if _, ok := intPart.SetString(s[:intEnd], 10); !ok && intEnd > 0 {
		return nil
	}

	if decimals > 0 {
		intPart.Mul(
			intPart,
			new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil),
		)

		var ok bool
		if fracDigits := fracEnd - fracStart; fracDigits > decimals {
			fracEnd = fracStart + decimals
			_, ok = fracPart.SetString(s[fracStart:fracEnd], 10)
		} else if fracDigits < decimals {
			_, ok = fracPart.SetString(s[fracStart:fracEnd]+strings.Repeat("0", decimals-fracDigits), 10)
		} else {
			_, ok = fracPart.SetString(s[fracStart:fracEnd], 10)
		}
		if !ok {
			return nil
		}
	}

	return new(big.Int).Add(intPart, fracPart)
}

// Keccak returns the Keccak256 hash of data. It is short for crypto.Keccak256Hash(…)
func Keccak(data []byte) common.Hash {
	return crypto.Keccak256Hash(data)
}
