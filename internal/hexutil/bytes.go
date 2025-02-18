package hexutil

import (
	"encoding/hex"
)

// Bytes is a byte slice that marshals/unmarshals from a hex-encoded string.
type Bytes []byte

// UnmarshalText decodes a hex string (with or without 0x prefix) into Bytes.
func (b *Bytes) UnmarshalText(data []byte) error {
	if len(data) >= 2 && data[0] == '0' && (data[1] == 'x' || data[1] == 'X') {
		data = data[2:]
	}

	*b = make([]byte, hex.DecodedLen(len(data)))
	_, err := hex.Decode(*b, data)
	return err
}

// MarshalText encodes Bytes into a hex string with a 0x prefix.
func (b Bytes) MarshalText() ([]byte, error) {
	result := make([]byte, 2+hex.EncodedLen(len(b)))
	copy(result, `0x`)
	hex.Encode(result[2:], b)
	return result, nil
}
