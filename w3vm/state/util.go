package state

import (
	"encoding/hex"
	"math/big"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/holiman/uint256"
	"github.com/lmittmann/w3/internal/mod"
	"github.com/lmittmann/w3/internal/module"
	"github.com/lmittmann/w3/w3types"
)

var hash0 common.Hash

////////////////////////////////////////////////////////////////////////////////////////////////////
// w3types.Caller's ////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////////////////

// ethBalance is like [eth.Balance], but returns the balance as [uint256.Int].
func ethBalance(addr common.Address, blockNumber *big.Int) w3types.CallerFactory[uint256.Int] {
	return module.NewFactory[uint256.Int](
		"eth_getBalance",
		[]any{addr, module.BlockNumberArg(blockNumber)},
	)
}

// ethStorageAt is like [eth.StorageAt], but returns the storage value as [uint256.Int].
func ethStorageAt(addr common.Address, slot uint256.Int, blockNumber *big.Int) w3types.CallerFactory[uint256.Int] {
	return module.NewFactory[uint256.Int](
		"eth_getStorageAt",
		[]any{addr, &slot, module.BlockNumberArg(blockNumber)},
		module.WithRetWrapper(func(ret *uint256.Int) any { return (*uint256OrHash)(ret) }),
	)
}

// uint256OrHash is like [uint256.Int], but can be unmarshaled from a hex number
// with leading zeros.
type uint256OrHash uint256.Int

func (i *uint256OrHash) UnmarshalText(text []byte) error {
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

	(*uint256.Int)(i).SetBytes(buf)
	return nil
}

func (i uint256OrHash) MarshalText() ([]byte, error) {
	return (*uint256.Int)(&i).MarshalText()
}

////////////////////////////////////////////////////////////////////////////////////////////////////
// Testing  ////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////////////////

func getTbFilepath(tb testing.TB) string {
	// Find test name of the root test (drop subtests from name).
	if tb == nil || tb.Name() == "" {
		return ""
	}
	tn := strings.SplitN(tb.Name(), "/", 2)[0]

	// Find the test function in the call stack. Don't go deeper than 32 frames.
	for i := 0; i < 32; i++ {
		pc, file, _, ok := runtime.Caller(i)
		if !ok {
			break
		}

		fn := runtime.FuncForPC(pc).Name()
		_, fn = filepath.Split(fn)
		fn = strings.SplitN(fn, ".", 3)[1]

		if fn == tn {
			return filepath.Dir(file)
		}
	}
	return ""
}

func isTbInMod(fp string) bool {
	return mod.Root != "" && strings.HasPrefix(fp, mod.Root)
}