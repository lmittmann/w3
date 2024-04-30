/*
Package w3vm provides a VM for executing EVM messages.
*/
package w3vm

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/core"
	gethState "github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
	"github.com/holiman/uint256"
	"github.com/lmittmann/w3"
	"github.com/lmittmann/w3/module/eth"
	"github.com/lmittmann/w3/w3types"
)

var (
	pendingBlockNumber = big.NewInt(-1)

	ErrFetch  = errors.New("fetching failed")
	ErrRevert = errors.New("execution reverted")
)

type VM struct {
	opts *options

	txIndex uint64
	db      *gethState.StateDB
}

// New creates a new VM, that is configured with the given options.
func New(opts ...Option) (*VM, error) {
	vm := &VM{opts: new(options)}
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt(vm)
	}

	if err := vm.opts.Init(); err != nil {
		return nil, err
	}

	// set DB
	db := newDB(vm.opts.fetcher)
	vm.db, _ = gethState.New(hash0, db, nil)
	for addr, acc := range vm.opts.preState {
		vm.db.SetNonce(addr, acc.Nonce)
		if acc.Balance != nil {
			vm.db.SetBalance(addr, uint256.MustFromBig(acc.Balance), tracing.BalanceIncreaseGenesisBalance)
		}
		if acc.Code != nil {
			vm.db.SetCode(addr, acc.Code)
		}
		for slot, val := range acc.Storage {
			vm.db.SetState(addr, slot, val)
		}
	}
	return vm, nil
}

// Apply the given message to the VM and return its receipt. Multiple tracing hooks
// can be given to trace the execution of the message.
func (vm *VM) Apply(msg *w3types.Message, hooks ...*tracing.Hooks) (*Receipt, error) {
	return vm.apply(msg, false, joinHooks(hooks))
}

// ApplyTx is like [VM.Apply], but takes a transaction instead of a message.
func (vm *VM) ApplyTx(tx *types.Transaction, hooks ...*tracing.Hooks) (*Receipt, error) {
	msg, err := new(w3types.Message).SetTx(tx, vm.opts.Signer())
	if err != nil {
		return nil, err
	}
	return vm.Apply(msg, hooks...)
}

func (v *VM) apply(msg *w3types.Message, isCall bool, hooks *tracing.Hooks) (*Receipt, error) {
	if v.db.Error() != nil {
		return nil, ErrFetch
	}

	coreMsg, txCtx, err := v.buildMessage(msg, isCall)
	if err != nil {
		return nil, err
	}

	var txHash common.Hash
	binary.BigEndian.PutUint64(txHash[:], v.txIndex)
	v.txIndex++
	v.db.SetTxContext(txHash, 0)

	gp := new(core.GasPool).AddGas(coreMsg.GasLimit)
	evm := vm.NewEVM(*v.opts.blockCtx, *txCtx, v.db, v.opts.chainConfig, vm.Config{
		Tracer:    hooks,
		NoBaseFee: v.opts.noBaseFee || isCall,
	})

	snap := v.db.Snapshot()

	// apply the message to the evm
	result, err := core.ApplyMessage(evm, coreMsg, gp)
	if err != nil {
		return nil, err
	}

	// build receipt
	receipt := &Receipt{
		f:         msg.Func,
		GasUsed:   result.UsedGas,
		GasRefund: result.RefundedGas,
		GasLimit:  result.UsedGas + v.db.GetRefund(),
		Output:    result.ReturnData,
		Logs:      v.db.GetLogs(txHash, 0, hash0),
	}

	if err := result.Err; err != nil {
		if reason, unpackErr := abi.UnpackRevert(result.ReturnData); unpackErr != nil {
			receipt.Err = ErrRevert
		} else {
			receipt.Err = fmt.Errorf("%w: %s", ErrRevert, reason)
		}
	}
	if msg.To == nil {
		contractAddr := crypto.CreateAddress(msg.From, coreMsg.Nonce)
		receipt.ContractAddress = &contractAddr
	}

	if isCall && !result.Failed() {
		v.db.RevertToSnapshot(snap)
	}
	v.db.Finalise(false)

	return receipt, receipt.Err
}

// Call calls the given message on the VM and returns a receipt. Any state changes
// of a call are reverted. Multiple tracing hooks can be passed to trace the execution
// of the message.
func (vm *VM) Call(msg *w3types.Message, hooks ...*tracing.Hooks) (*Receipt, error) {
	return vm.apply(msg, true, joinHooks(hooks))
}

// CallFunc is a utility function for [VM.Call] that calls the given function
// on the given contract address with the given arguments and parses the
// output into the given returns.
//
// Example:
//
//	funcBalanceOf := w3.MustNewFunc("balanceOf(address)", "uint256")
//
//	var balance *big.Int
//	err := vm.CallFunc(contractAddr, funcBalanceOf, addr).Returns(&balance)
//	if err != nil {
//		// ...
//	}
func (vm *VM) CallFunc(contract common.Address, f w3types.Func, args ...any) *CallFuncFactory {
	return &CallFuncFactory{
		vm: vm,
		msg: &w3types.Message{
			To:   &contract,
			Func: f,
			Args: args,
		},
	}
}

type CallFuncFactory struct {
	vm  *VM
	msg *w3types.Message
}

func (cff *CallFuncFactory) Returns(returns ...any) error {
	receipt, err := cff.vm.Call(cff.msg)
	if err != nil {
		return err
	}
	return receipt.DecodeReturns(returns...)
}

// Nonce returns the nonce of the given address.
func (vm *VM) Nonce(addr common.Address) (uint64, error) {
	nonce := vm.db.GetNonce(addr)
	if vm.db.Error() != nil {
		return 0, fmt.Errorf("%w: failed to fetch nonce of %s", ErrFetch, addr)
	}
	return nonce, nil
}

// Balance returns the balance of the given address.
func (vm *VM) Balance(addr common.Address) (*big.Int, error) {
	balance := vm.db.GetBalance(addr)
	if vm.db.Error() != nil {
		return nil, fmt.Errorf("%w: failed to fetch balance of %s", ErrFetch, addr)
	}
	return balance.ToBig(), nil
}

// Code returns the code of the given address.
func (vm *VM) Code(addr common.Address) ([]byte, error) {
	code := vm.db.GetCode(addr)
	if vm.db.Error() != nil {
		return nil, fmt.Errorf("%w: failed to fetch code of %s", ErrFetch, addr)
	}
	return code, nil
}

// StorageAt returns the state of the given address at the give storage slot.
func (vm *VM) StorageAt(addr common.Address, slot common.Hash) (common.Hash, error) {
	val := vm.db.GetState(addr, slot)
	if vm.db.Error() != nil {
		return hash0, fmt.Errorf("%w: failed to fetch storage of %s at %s", ErrFetch, addr, slot)
	}
	return val, nil
}

func (v *VM) buildMessage(msg *w3types.Message, skipAccChecks bool) (*core.Message, *vm.TxContext, error) {
	nonce := msg.Nonce
	if !skipAccChecks && nonce == 0 && msg.From != addr0 {
		var err error
		nonce, err = v.Nonce(msg.From)
		if err != nil {
			return nil, nil, err
		}
	}

	gasLimit := msg.Gas
	if maxGasLimit := v.opts.blockCtx.GasLimit; gasLimit == 0 {
		gasLimit = maxGasLimit
	} else if gasLimit > maxGasLimit {
		gasLimit = maxGasLimit
	}

	var input []byte
	if msg.Input == nil && msg.Func != nil {
		var err error
		input, err = msg.Func.EncodeArgs(msg.Args...)
		if err != nil {
			return nil, nil, err
		}
	} else {
		input = msg.Input
	}

	gasPrice := nilToZero(msg.GasPrice)
	gasFeeCap := nilToZero(msg.GasFeeCap)
	gasTipCap := nilToZero(msg.GasTipCap)
	if baseFee := v.opts.blockCtx.BaseFee; baseFee != nil && baseFee.Sign() > 0 {
		gasPrice = math.BigMin(gasFeeCap, new(big.Int).Add(baseFee, gasTipCap))
	}

	return &core.Message{
			To:                msg.To,
			From:              msg.From,
			Nonce:             nonce,
			Value:             nilToZero(msg.Value),
			GasLimit:          gasLimit,
			GasPrice:          gasPrice,
			GasFeeCap:         gasFeeCap,
			GasTipCap:         gasFeeCap,
			Data:              input,
			AccessList:        msg.AccessList,
			BlobGasFeeCap:     msg.BlobGasFeeCap,
			BlobHashes:        msg.BlobHashes,
			SkipAccountChecks: skipAccChecks,
		},
		&vm.TxContext{
			Origin:     msg.From,
			GasPrice:   gasPrice,
			BlobHashes: msg.BlobHashes,
			BlobFeeCap: msg.BlobGasFeeCap,
		},
		nil
}

func newBlockContext(h *types.Header, getHash vm.GetHashFunc) *vm.BlockContext {
	var random *common.Hash
	if h.Difficulty == nil || h.Difficulty.Sign() == 0 {
		random = &h.MixDigest
	}

	return &vm.BlockContext{
		CanTransfer: core.CanTransfer,
		Transfer:    core.Transfer,
		GetHash:     getHash,
		Coinbase:    h.Coinbase,
		BlockNumber: nilToZero(h.Number),
		Time:        h.Time,
		Difficulty:  nilToZero(h.Difficulty),
		BaseFee:     nilToZero(h.BaseFee),
		GasLimit:    h.GasLimit,
		Random:      random,
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////
// VM Option ///////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////////////////

type options struct {
	chainConfig *params.ChainConfig
	preState    w3types.State
	noBaseFee   bool

	blockCtx *vm.BlockContext
	header   *types.Header

	forkClient      *w3.Client
	forkBlockNumber *big.Int
	fetcher         Fetcher
	tb              testing.TB
}

func (opt *options) Signer() types.Signer {
	if opt.fetcher == nil {
		return types.LatestSigner(opt.chainConfig)
	}
	return types.MakeSigner(opt.chainConfig, opt.header.Number, opt.header.Time)
}

func (opts *options) Init() error {
	// set initial chain config
	isChainConfigSet := opts.chainConfig != nil
	if !isChainConfigSet {
		opts.chainConfig = allEthashProtocolChanges
	}

	// set fetcher
	if opts.fetcher == nil && opts.forkClient != nil {
		var calls []w3types.RPCCaller

		latest := opts.forkBlockNumber == nil
		if latest {
			opts.forkBlockNumber = new(big.Int)
			calls = append(calls, eth.BlockNumber().Returns(opts.forkBlockNumber))
		}
		if opts.header == nil && opts.blockCtx == nil {
			opts.header = new(types.Header)
			if latest {
				calls = append(calls, eth.HeaderByNumber(pendingBlockNumber).Returns(opts.header))
			} else {
				calls = append(calls, eth.HeaderByNumber(opts.forkBlockNumber).Returns(opts.header))
			}
		}

		if err := opts.forkClient.Call(calls...); err != nil {
			return fmt.Errorf("%w: failed to fetch header: %v", ErrFetch, err)
		}

		if latest {
			opts.fetcher = NewRPCFetcher(opts.forkClient, opts.forkBlockNumber)
		} else if opts.tb == nil {
			opts.fetcher = NewRPCFetcher(opts.forkClient, new(big.Int).Sub(opts.forkBlockNumber, w3.Big1))
		} else {
			opts.fetcher = NewTestingRPCFetcher(opts.tb, opts.chainConfig.ChainID.Uint64(), opts.forkClient, new(big.Int).Sub(opts.forkBlockNumber, w3.Big1))
		}
	}

	// potentially update chain config
	if !isChainConfigSet && opts.fetcher != nil {
		opts.chainConfig = params.MainnetChainConfig
	}

	if opts.blockCtx == nil {
		if opts.header != nil {
			opts.blockCtx = newBlockContext(opts.header, fetcherHashFunc(opts.fetcher))
		} else {
			opts.blockCtx = defaultBlockContext()
		}
	}
	return nil
}

func fetcherHashFunc(fetcher Fetcher) vm.GetHashFunc {
	return func(blockNumber uint64) common.Hash {
		hash, _ := fetcher.HeaderHash(blockNumber)
		return hash
	}
}

// An Option configures a [VM].
type Option func(*VM)

// WithChainConfig sets the chain config for the VM.
func WithChainConfig(cfg *params.ChainConfig) Option {
	return func(vm *VM) { vm.opts.chainConfig = cfg }
}

// WithBlockContext sets the block context for the VM.
func WithBlockContext(ctx *vm.BlockContext) Option {
	return func(vm *VM) { vm.opts.blockCtx = ctx }
}

// WithState sets the pre state of the VM.
//
// WithState can be used together with [WithFork] to only set the state of some
// accounts, or partially overwrite the storage of an account.
func WithState(state w3types.State) Option {
	return func(vm *VM) { vm.opts.preState = state }
}

// WithNoBaseFee forces the EIP-1559 base fee to 0 for the VM.
func WithNoBaseFee() Option {
	return func(vm *VM) { vm.opts.noBaseFee = true }
}

// WithFork sets the client and block number to fetch state from and sets the
// block context for the VM. If the block number is nil, the latest state is
// fetched and the pending block is used for constructing the block context.
//
// If used together with [WithTB], fetched state is stored in the testdata
// directory of the tests package.
func WithFork(client *w3.Client, blockNumber *big.Int) Option {
	return func(vm *VM) {
		vm.opts.forkClient = client
		vm.opts.forkBlockNumber = blockNumber
	}
}

// WithHeader sets the block context for the VM based on the given header
func WithHeader(header *types.Header) Option {
	return func(vm *VM) { vm.opts.header = header }
}

// WithFetcher sets the fetcher for the VM.
func WithFetcher(fetcher Fetcher) Option {
	return func(vm *VM) { vm.opts.fetcher = fetcher }
}

// WithTB enables persistent state caching when used together with [WithFork].
// State is stored in the testdata directory of the tests package.
func WithTB(tb testing.TB) Option {
	return func(vm *VM) { vm.opts.tb = tb }
}
