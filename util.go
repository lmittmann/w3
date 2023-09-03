package w3

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
)

// Common [big.Int]'s.
var (
	Big0     = big.NewInt(0)
	Big1     = big.NewInt(1)
	Big2     = big.NewInt(2)
	BigGwei  = big.NewInt(1_000000000)
	BigEther = big.NewInt(1_000000000_000000000)
)

// A returns an address from a hexstring or panics if the hexstring does not
// represent a valid checksum encoded address.
//
// Use [common.HexToAddress] to get the address from a hexstring without
// panicking.
func A(hexAddress string) (addr common.Address) {
	if !has0xPrefix(hexAddress) {
		panic(fmt.Sprintf("hex address %q must have 0x prefix", hexAddress))
	}
	if !isHex(hexAddress[2:]) {
		panic(fmt.Sprintf("hex address %q must be hex", hexAddress))
	}
	if len(hexAddress) != 42 {
		panic(fmt.Sprintf("hex address %q must have 20 bytes", hexAddress))
	}

	hex.Decode(addr[:], []byte(hexAddress[2:]))

	if hexAddress != addr.Hex() {
		panic(fmt.Sprintf("hex address %q must be checksum encoded", hexAddress))
	}
	return addr
}

// APtr returns an address pointer from a hexstring or panics if the hexstring
// does not represent a valid checksum encoded address.
func APtr(hexAddress string) *common.Address {
	addr := A(hexAddress)
	return &addr
}

// B returns a byte slice from a hexstring or panics if the hexstring does not
// represent a valid byte slice.
//
// Use [common.FromHex] to get the byte slice from a hexstring without
// panicking.
func B(hexBytes string) (bytes []byte) {
	if !has0xPrefix(hexBytes) {
		panic(fmt.Sprintf("hex bytes %q must have 0x prefix", hexBytes))
	}
	if !isHex(hexBytes[2:]) {
		panic(fmt.Sprintf("hex bytes %q must be hex", hexBytes))
	}
	if len(hexBytes)%2 != 0 {
		panic(fmt.Sprintf("hex bytes %q must have even number of hex chars", hexBytes))
	}

	bytes, _ = hex.DecodeString(hexBytes[2:])
	return bytes
}

// H returns a hash from a hexstring or panics if the hexstring does not
// represent a valid hash.
//
// Use [common.HexToHash] to get the hash from a hexstring without panicking.
func H(hexHash string) (hash common.Hash) {
	if !has0xPrefix(hexHash) {
		panic(fmt.Sprintf("hex hash %q must have 0x prefix", hexHash))
	}
	if !isHex(hexHash[2:]) {
		panic(fmt.Sprintf("hex hash %q must be hex", hexHash))
	}
	if len(hexHash) != 66 {
		panic(fmt.Sprintf("hex hash %q must have 32 bytes", hexHash))
	}

	hex.Decode(hash[:], []byte(hexHash[2:]))
	return hash
}

// I returns a [big.Int] from a hexstring or decimal number string (with
// optional unit) or panics if the parsing fails.
//
// I supports the units "ether" or "eth" and "gwei" for decimal number strings.
// E.g.:
//
//	w3.I("1 ether")   -> 1000000000000000000
//	w3.I("1.2 ether") -> 1200000000000000000
//
// Fractional digits that exceed the units maximum number of fractional digits
// are ignored. E.g.:
//
//	w3.I("0.000000123456 gwei") -> 123
func I(strInt string) *big.Int {
	if has0xPrefix(strInt) {
		return parseHexBig(strInt)
	}
	return parseDecimal(strInt)
}

func parseHexBig(hexBig string) *big.Int {
	if !isHex(hexBig[2:]) {
		panic(fmt.Sprintf("hex big %q must be hex", hexBig))
	}
	bigInt, _ := new(big.Int).SetString(hexBig[2:], 16)
	return bigInt
}

func parseDecimal(strBig string) *big.Int {
	numberUnit := strings.SplitN(strBig, " ", 2)
	integerFraction := strings.SplitN(numberUnit[0], ".", 2)
	integer, ok := new(big.Int).SetString(integerFraction[0], 10)
	if !ok {
		panic(fmt.Sprintf("str big %q must be number", strBig))
	}

	// len == 1
	if len(numberUnit) == 1 {
		if len(integerFraction) > 1 {
			panic(fmt.Sprintf("str big %q without unit must be integer", strBig))
		}
		return integer
	}

	// len == 2
	unit := strings.ToLower(numberUnit[1])
	switch unit {
	case "ether", "eth":
		integer.Mul(integer, BigEther)
	case "gwei":
		integer.Mul(integer, BigGwei)
	default:
		panic(fmt.Sprintf("str big %q has invalid unit %q", strBig, unit))
	}

	// integer
	if len(integerFraction) == 1 {
		return integer
	}

	// float
	fraction, ok := new(big.Int).SetString(integerFraction[1], 10)
	if !ok {
		panic(fmt.Sprintf("str big %q must be number", strBig))
	}

	decimals := len(integerFraction[1])
	switch unit {
	case "ether", "eth":
		if fraction.Cmp(BigEther) >= 0 {
			panic(fmt.Sprintf("str big %q exceeds precision", strBig))
		}
		fraction.Mul(fraction, new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(18-decimals)), nil))
	case "gwei":
		if fraction.Cmp(BigGwei) >= 0 {
			panic(fmt.Sprintf("str big %q exceeds precision", strBig))
		}
		fraction.Mul(fraction, new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(9-decimals)), nil))
	}

	return integer.Add(integer, fraction)
}

// FromWei returns the given Wei as decimal with the given number of decimals.
func FromWei(wei *big.Int, decimals uint8) string {
	if wei == nil {
		return fmt.Sprint(nil)
	}

	d := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)

	sign := ""
	if wei.Sign() < 0 {
		sign = "-"
	}
	wei = new(big.Int).Abs(wei)

	z, m := new(big.Int).DivMod(wei, d, new(big.Int))
	if m.Cmp(new(big.Int)) == 0 {
		return sign + z.String()
	}
	s := strings.TrimRight(fmt.Sprintf("%0*s", decimals, m.String()), "0")
	return sign + z.String() + "." + s
}

// has0xPrefix validates hexStr begins with '0x' or '0X'.
func has0xPrefix(hexStr string) bool {
	return len(hexStr) >= 2 && hexStr[0] == '0' && hexStr[1] == 'x'
}

// isHexCharacter returns bool of c being a valid hexadecimal.
func isHexCharacter(c rune) bool {
	return ('0' <= c && c <= '9') || ('a' <= c && c <= 'f') || ('A' <= c && c <= 'F')
}

// isHex validates whether each byte is valid hexadecimal string.
func isHex(str string) bool {
	for _, c := range str {
		if !isHexCharacter(c) {
			return false
		}
	}
	return true
}
