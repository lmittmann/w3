package hexutil

import (
	"encoding/hex"

	"github.com/holiman/uint256"
)

// U256 is a wrapper type for [uint256.Int] that marshals and unmarshals hex strings.
// It can decode hex strings with or without the 0x prefix and with or without leading zeros.
type U256 uint256.Int

func (u *U256) UnmarshalText(text []byte) error {
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

	(*uint256.Int)(u).SetBytes(buf)
	return nil
}

func (u U256) MarshalText() ([]byte, error) {
	return []byte((*uint256.Int)(&u).Hex()), nil
}
