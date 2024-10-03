package w3vm

import (
	"crypto/rand"
	"math/big"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/holiman/uint256"
	"github.com/lmittmann/w3"
	"github.com/lmittmann/w3/internal/crypto"
	w3hexutil "github.com/lmittmann/w3/internal/hexutil"
	"github.com/lmittmann/w3/internal/mod"
	"github.com/lmittmann/w3/internal/module"
	"github.com/lmittmann/w3/w3types"
)

// RandA returns a random address.
func RandA() (addr common.Address) {
	rand.Read(addr[:])
	return addr
}

var (
	weth9BalancePos   = common.BytesToHash([]byte{3})
	weth9AllowancePos = common.BytesToHash([]byte{4})
)

// WETHBalanceSlot returns the storage slot that stores the WETH balance of
// the given addr.
func WETHBalanceSlot(addr common.Address) common.Hash {
	return Slot(weth9BalancePos, common.BytesToHash(addr[:]))
}

// WETHAllowanceSlot returns the storage slot that stores the WETH allowance
// of the given owner to the spender.
func WETHAllowanceSlot(owner, spender common.Address) common.Hash {
	return Slot2(weth9AllowancePos, common.BytesToHash(owner[:]), common.BytesToHash(spender[:]))
}

// Slot returns the storage slot of a mapping with the given position and key.
//
// Slot follows the Solidity storage layout for:
//
//	mapping(bytes32 => bytes32)
func Slot(pos, key common.Hash) common.Hash {
	return crypto.Keccak256Hash(key[:], pos[:])
}

// Slot2 returns the storage slot of a double mapping with the given position
// and keys.
//
// Slot2 follows the Solidity storage layout for:
//
//	mapping(bytes32 => mapping(bytes32 => bytes32))
func Slot2(pos, key, key2 common.Hash) common.Hash {
	return crypto.Keccak256Hash(
		key2[:],
		crypto.Keccak256(key[:], pos[:]),
	)
}

// nilToZero converts sets a pointer to the zero value if it is nil.
func nilToZero[T any](ptr *T) *T {
	if ptr == nil {
		return new(T)
	}
	return ptr
}

// zeroHashFunc implements a [vm.GetHashFunc] that always returns the zero hash.
func zeroHashFunc(uint64) common.Hash {
	return w3.Hash0
}

////////////////////////////////////////////////////////////////////////////////////////////////////
// w3types.RPCCaller's /////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////////////////

// ethBalance is like [eth.Balance], but returns the balance as [uint256.Int].
func ethBalance(addr common.Address, blockNumber *big.Int) w3types.RPCCallerFactory[uint256.Int] {
	return module.NewFactory(
		"eth_getBalance",
		[]any{addr, module.BlockNumberArg(blockNumber)},
		module.WithRetWrapper(func(ret *uint256.Int) any { return (*hexutil.U256)(ret) }),
	)
}

// ethStorageAt is like [eth.StorageAt], but returns the storage value as [uint256.Int].
func ethStorageAt(addr common.Address, slot common.Hash, blockNumber *big.Int) w3types.RPCCallerFactory[common.Hash] {
	return module.NewFactory(
		"eth_getStorageAt",
		[]any{addr, slot, module.BlockNumberArg(blockNumber)},
		module.WithRetWrapper(func(ret *common.Hash) any { return (*w3hexutil.Hash)(ret) }),
	)
}

// ethHeaderHash is like [eth.Header], but only parses the header hash.
func ethHeaderHash(blockNumber uint64) w3types.RPCCallerFactory[header] {
	return module.NewFactory[header](
		"eth_getBlockByNumber",
		[]any{hexutil.Uint64(blockNumber), false},
	)
}

type header struct {
	Hash common.Hash `json:"hash"`
}

////////////////////////////////////////////////////////////////////////////////////////////////////
// tracing.Hook's //////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////////////////

// joinHooks joins multiple hooks into one.
func joinHooks(hooks []*tracing.Hooks) *tracing.Hooks {
	// hot path
	switch len(hooks) {
	case 0:
		return nil
	case 1:
		return hooks[0]
	}

	// vm hooks
	var onTxStarts []tracing.TxStartHook
	var onTxEnds []tracing.TxEndHook
	var onEnters []tracing.EnterHook
	var onExits []tracing.ExitHook
	var onOpcodes []tracing.OpcodeHook
	var onFaults []tracing.FaultHook
	var onGasChanges []tracing.GasChangeHook
	// state hooks
	var onBalanceChanges []tracing.BalanceChangeHook
	var onNonceChanges []tracing.NonceChangeHook
	var onCodeChanges []tracing.CodeChangeHook
	var onStorageChanges []tracing.StorageChangeHook
	var onLogs []tracing.LogHook

	for _, h := range hooks {
		if h == nil {
			continue
		}
		// vm hooks
		if h.OnTxStart != nil {
			onTxStarts = append(onTxStarts, h.OnTxStart)
		}
		if h.OnTxEnd != nil {
			onTxEnds = append(onTxEnds, h.OnTxEnd)
		}
		if h.OnEnter != nil {
			onEnters = append(onEnters, h.OnEnter)
		}
		if h.OnExit != nil {
			onExits = append(onExits, h.OnExit)
		}
		if h.OnOpcode != nil {
			onOpcodes = append(onOpcodes, h.OnOpcode)
		}
		if h.OnFault != nil {
			onFaults = append(onFaults, h.OnFault)
		}
		if h.OnGasChange != nil {
			onGasChanges = append(onGasChanges, h.OnGasChange)
		}
		// state hooks
		if h.OnBalanceChange != nil {
			onBalanceChanges = append(onBalanceChanges, h.OnBalanceChange)
		}
		if h.OnNonceChange != nil {
			onNonceChanges = append(onNonceChanges, h.OnNonceChange)
		}
		if h.OnCodeChange != nil {
			onCodeChanges = append(onCodeChanges, h.OnCodeChange)
		}
		if h.OnStorageChange != nil {
			onStorageChanges = append(onStorageChanges, h.OnStorageChange)
		}
		if h.OnLog != nil {
			onLogs = append(onLogs, h.OnLog)
		}
	}

	hook := new(tracing.Hooks)
	// vm hooks
	if len(onTxStarts) > 0 {
		hook.OnTxStart = func(vm *tracing.VMContext, tx *types.Transaction, from common.Address) {
			for _, h := range onTxStarts {
				h(vm, tx, from)
			}
		}
	}
	if len(onTxEnds) > 0 {
		hook.OnTxEnd = func(receipt *types.Receipt, err error) {
			for _, h := range onTxEnds {
				h(receipt, err)
			}
		}
	}
	if len(onEnters) > 0 {
		hook.OnEnter = func(depth int, typ byte, from, to common.Address, input []byte, gas uint64, value *big.Int) {
			for _, h := range onEnters {
				h(depth, typ, from, to, input, gas, value)
			}
		}
	}
	if len(onExits) > 0 {
		hook.OnExit = func(depth int, output []byte, gasUsed uint64, err error, reverted bool) {
			for _, h := range onExits {
				h(depth, output, gasUsed, err, reverted)
			}
		}
	}
	if len(onOpcodes) > 0 {
		hook.OnOpcode = func(pc uint64, op byte, gas, cost uint64, scope tracing.OpContext, rData []byte, depth int, err error) {
			for _, h := range onOpcodes {
				h(pc, op, gas, cost, scope, rData, depth, err)
			}
		}
	}
	if len(onFaults) > 0 {
		hook.OnFault = func(pc uint64, op byte, gas, cost uint64, scope tracing.OpContext, depth int, err error) {
			for _, h := range onFaults {
				h(pc, op, gas, cost, scope, depth, err)
			}
		}
	}
	if len(onGasChanges) > 0 {
		hook.OnGasChange = func(old, new uint64, reason tracing.GasChangeReason) {
			for _, h := range onGasChanges {
				h(old, new, reason)
			}
		}
	}
	// state hooks
	if len(onBalanceChanges) > 0 {
		hook.OnBalanceChange = func(addr common.Address, prev, new *big.Int, reason tracing.BalanceChangeReason) {
			for _, h := range onBalanceChanges {
				h(addr, prev, new, reason)
			}
		}
	}
	if len(onNonceChanges) > 0 {
		hook.OnNonceChange = func(addr common.Address, prev, new uint64) {
			for _, h := range onNonceChanges {
				h(addr, prev, new)
			}
		}
	}
	if len(onCodeChanges) > 0 {
		hook.OnCodeChange = func(addr common.Address, prevCodeHash common.Hash, prevCode []byte, codeHash common.Hash, code []byte) {
			for _, h := range onCodeChanges {
				h(addr, prevCodeHash, prevCode, codeHash, code)
			}
		}
	}
	if len(onStorageChanges) > 0 {
		hook.OnStorageChange = func(addr common.Address, slot, prev, new common.Hash) {
			for _, h := range onStorageChanges {
				h(addr, slot, prev, new)
			}
		}
	}
	if len(onLogs) > 0 {
		hook.OnLog = func(log *types.Log) {
			for _, h := range onLogs {
				h(log)
			}
		}
	}
	return hook
}

////////////////////////////////////////////////////////////////////////////////////////////////////
// Testing /////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////////////////

func getTbFilepath(tb testing.TB) string {
	// Find test name of the root test (drop subtests from name).
	if tb == nil || tb.Name() == "" {
		return ""
	}
	tn := strings.SplitN(tb.Name(), "/", 2)[0]

	// Find the test function in the call stack. Don't go deeper than 32 frames.
	for i := range 32 {
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
