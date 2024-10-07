//go:generate go run gen.go

package fourbyte

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/lmittmann/w3"
)

var (
	precompile1 = w3.MustNewFunc("ecRecover(bytes32 hash, uint8 v, bytes32 r, bytes32 s)", "address")
	precompile2 = w3.MustNewFunc("keccak(bytes)", "bytes32")
	precompile3 = w3.MustNewFunc("ripemd160(bytes)", "bytes32")
	precompile4 = w3.MustNewFunc("identity(bytes)", "bytes")

	addr1 = common.BytesToAddress([]byte{0x01})
	addr2 = common.BytesToAddress([]byte{0x02})
	addr3 = common.BytesToAddress([]byte{0x03})
	addr4 = common.BytesToAddress([]byte{0x04})
)

func Function(sig [4]byte, addr common.Address) (fn *w3.Func, isPrecompile bool) {
	switch addr {
	case addr1:
		return precompile1, true
	case addr2:
		return precompile2, true
	case addr3:
		return precompile3, true
	case addr4:
		return precompile4, true
	}
	return functions[sig], false
}

func Event(topic0 [32]byte, addr common.Address) *w3.Event {
	return events[topic0]
}
