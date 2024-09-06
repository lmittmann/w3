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
	Big0     = new(big.Int)
	Big1     = big.NewInt(1)
	Big2     = big.NewInt(2)
	BigGwei  = big.NewInt(1_000000000)
	BigEther = big.NewInt(1_000000000_000000000)

	// Max Uint Values.
	BigMaxUint256 = new(big.Int).Sub(new(big.Int).Lsh(Big1, 256), Big1)
	BigMaxUint248 = new(big.Int).Sub(new(big.Int).Lsh(Big1, 248), Big1)
	BigMaxUint240 = new(big.Int).Sub(new(big.Int).Lsh(Big1, 240), Big1)
	BigMaxUint232 = new(big.Int).Sub(new(big.Int).Lsh(Big1, 232), Big1)
	BigMaxUint224 = new(big.Int).Sub(new(big.Int).Lsh(Big1, 224), Big1)
	BigMaxUint216 = new(big.Int).Sub(new(big.Int).Lsh(Big1, 216), Big1)
	BigMaxUint208 = new(big.Int).Sub(new(big.Int).Lsh(Big1, 208), Big1)
	BigMaxUint200 = new(big.Int).Sub(new(big.Int).Lsh(Big1, 200), Big1)
	BigMaxUint192 = new(big.Int).Sub(new(big.Int).Lsh(Big1, 192), Big1)
	BigMaxUint184 = new(big.Int).Sub(new(big.Int).Lsh(Big1, 184), Big1)
	BigMaxUint176 = new(big.Int).Sub(new(big.Int).Lsh(Big1, 176), Big1)
	BigMaxUint168 = new(big.Int).Sub(new(big.Int).Lsh(Big1, 168), Big1)
	BigMaxUint160 = new(big.Int).Sub(new(big.Int).Lsh(Big1, 160), Big1)
	BigMaxUint152 = new(big.Int).Sub(new(big.Int).Lsh(Big1, 152), Big1)
	BigMaxUint144 = new(big.Int).Sub(new(big.Int).Lsh(Big1, 144), Big1)
	BigMaxUint136 = new(big.Int).Sub(new(big.Int).Lsh(Big1, 136), Big1)
	BigMaxUint128 = new(big.Int).Sub(new(big.Int).Lsh(Big1, 128), Big1)
	BigMaxUint120 = new(big.Int).Sub(new(big.Int).Lsh(Big1, 120), Big1)
	BigMaxUint112 = new(big.Int).Sub(new(big.Int).Lsh(Big1, 112), Big1)
	BigMaxUint104 = new(big.Int).Sub(new(big.Int).Lsh(Big1, 104), Big1)
	BigMaxUint96  = new(big.Int).Sub(new(big.Int).Lsh(Big1, 96), Big1)
	BigMaxUint88  = new(big.Int).Sub(new(big.Int).Lsh(Big1, 88), Big1)
	BigMaxUint80  = new(big.Int).Sub(new(big.Int).Lsh(Big1, 80), Big1)
	BigMaxUint72  = new(big.Int).Sub(new(big.Int).Lsh(Big1, 72), Big1)
	BigMaxUint64  = new(big.Int).Sub(new(big.Int).Lsh(Big1, 64), Big1)
	BigMaxUint56  = new(big.Int).Sub(new(big.Int).Lsh(Big1, 56), Big1)
	BigMaxUint48  = new(big.Int).Sub(new(big.Int).Lsh(Big1, 48), Big1)
	BigMaxUint40  = new(big.Int).Sub(new(big.Int).Lsh(Big1, 40), Big1)
	BigMaxUint32  = new(big.Int).Sub(new(big.Int).Lsh(Big1, 32), Big1)
	BigMaxUint24  = new(big.Int).Sub(new(big.Int).Lsh(Big1, 24), Big1)
	BigMaxUint16  = new(big.Int).Sub(new(big.Int).Lsh(Big1, 16), Big1)
	BigMaxUint8   = new(big.Int).Sub(new(big.Int).Lsh(Big1, 8), Big1)
)

// Zero Values.
var (
	Addr0 common.Address
	Hash0 common.Hash
)

// A returns an address from a hexstring or panics if the hexstring does not
// represent a valid address.
//
// Use [common.HexToAddress] to get the address from a hexstring without
// panicking.
func A(hexAddr string) (addr common.Address) {
	if has0xPrefix(hexAddr) {
		hexAddr = hexAddr[2:]
	}

	n, err := hex.Decode(addr[:], []byte(hexAddr))
	if err != nil {
		panic(fmt.Sprintf("invalid address %q: %v", hexAddr, err))
	} else if n != 20 {
		panic(fmt.Sprintf("invalid address %q: must have 20 bytes", hexAddr))
	}
	return addr
}

// APtr returns an address pointer from a hexstring or panics if the hexstring
// does not represent a valid address.
func APtr(hexAddress string) *common.Address {
	addr := A(hexAddress)
	return &addr
}

// B returns a byte slice from a hexstring or panics if the hexstring does not
// represent a valid byte slice.
//
// Use [common.FromHex] to get the byte slice from a hexstring without
// panicking.
func B(hexBytes ...string) (bytes []byte) {
	for _, s := range hexBytes {
		if has0xPrefix(s) {
			s = s[2:]
		}

		b, err := hex.DecodeString(s)
		if err != nil {
			panic(fmt.Sprintf("invalid bytes %q: %v", s, err))
		}
		bytes = append(bytes, b...)
	}
	return bytes
}

// H returns a hash from a hexstring or panics if the hexstring does not
// represent a valid hash.
//
// Use [common.HexToHash] to get the hash from a hexstring without panicking.
func H(hexHash string) (hash common.Hash) {
	if has0xPrefix(hexHash) {
		hexHash = hexHash[2:]
	}

	n, err := hex.Decode(hash[:], []byte(hexHash))
	if err != nil {
		panic(fmt.Sprintf("invalid hash %q: %v", hexHash, err))
	} else if n != 32 {
		panic(fmt.Sprintf("invalid hash %q: must have 32 bytes", hexHash))
	}
	return hash
}

// I returns a [big.Int] from a hexstring or decimal number string (with
// optional unit) or panics if the parsing fails.
//
// I supports the units "ether" or "eth" and "gwei" for decimal number strings.
// E.g.:
//
//	w3.I("1 ether")   -> 1000000000000000000
//	w3.I("10 gwei")   -> 10000000000
//
// Fractional digits that exceed the units maximum number of fractional digits
// are ignored. E.g.:
//
//	w3.I("0.000000123456 gwei") -> 123
func I(strInt string) *big.Int {
	if has0xPrefix(strInt) {
		return parseHexBig(strInt[2:])
	}
	return parseDecimal(strInt)
}

func parseHexBig(hexBig string) *big.Int {
	bigInt, ok := new(big.Int).SetString(hexBig, 16)
	if !ok {
		panic(fmt.Sprintf("invalid hex big %q", "0x"+hexBig))
	}
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
	return len(hexStr) >= 2 && hexStr[0] == '0' && (hexStr[1] == 'x' || hexStr[1] == 'X')
}
