package hexutil

import (
	"bytes"
	"encoding/hex"

	"github.com/ethereum/go-ethereum/common"
)

// Hash is a wrapper type for [common.Hash] that marshals and unmarshals hex strings.
// It can decode hex strings with or without the 0x prefix and with or without leading
// zeros and encodes hashes without leading zeros.
type Hash common.Hash

func (h *Hash) UnmarshalText(text []byte) error {
	if len(text) >= 2 && text[0] == '0' && (text[1] == 'x' || text[1] == 'X') {
		text = text[2:]
	}
	if len(text)%2 != 0 {
		text = append([]byte{'0'}, text...)
	}

	bytes, _ := hex.DecodeString(string(text))
	*h = Hash(common.BytesToHash(bytes))
	return nil
}

func (h Hash) MarshalText() ([]byte, error) {
	bytes := bytes.TrimLeft(h[:], "\x00")
	if len(bytes) == 0 {
		return []byte("0x0"), nil
	}
	hexStr := hex.EncodeToString(bytes)
	if bytes[0]&0xf0 == 0 {
		hexStr = hexStr[1:]
	}
	return []byte("0x" + hexStr), nil
}
