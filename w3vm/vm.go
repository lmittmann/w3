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
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
	"github.com/lmittmann/w3"
	"github.com/lmittmann/w3/module/eth"
	"github.com/lmittmann/w3/w3types"
	"github.com/lmittmann/w3/w3vm/state"
)

var (
	big1               = big.NewInt(1)
	pendingBlockNumber = big.NewInt(-1)

	ErrFetch  = errors.New("fetch error")
	ErrRevert = errors.New("revert error")
)

type VM struct {
	opts *vmOptions

	chainConfig *params.ChainConfig
	blockCtx    *vm.BlockContext
	noBaseFee   bool
	txIndex     uint64
	db          *gethState.StateDB

	fetcher state.Fetcher
}

type vmOptions struct {
	chainConfig *params.ChainConfig
	preState    w3types.State
	noBaseFee   bool

	blockCtx *vm.BlockContext
	header   *types.Header

	forkClient      *w3.Client
	forkBlockNumber *big.Int
	fetcher         state.Fetcher
	tb              testing.TB
}

func New(opts ...Option) (*VM, error) {
	vm := &VM{opts: new(vmOptions)}
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt(vm)
	}

	vm.fetcher = vm.opts.fetcher
	if vm.fetcher == nil && vm.opts.forkClient != nil {
		var calls []w3types.Caller

		latest := vm.opts.forkBlockNumber == nil
		if latest {
			vm.opts.forkBlockNumber = new(big.Int)
			calls = append(calls, eth.BlockNumber().Returns(vm.opts.forkBlockNumber))
		}
		if vm.opts.header == nil && vm.opts.blockCtx == nil {
			vm.opts.header = new(types.Header)
			if latest {
				calls = append(calls, eth.HeaderByNumber(pendingBlockNumber).Returns(vm.opts.header))
			} else {
				calls = append(calls, eth.HeaderByNumber(vm.opts.forkBlockNumber).Returns(vm.opts.header))
			}
		}

		if err := vm.opts.forkClient.Call(calls...); err != nil {
			return nil, fmt.Errorf("%w: failed to fetch header: %v", ErrFetch, err)
		}

		if latest || vm.opts.tb == nil {
			vm.fetcher = state.NewRPCFetcher(vm.opts.forkClient, vm.opts.forkBlockNumber)
		} else {
			vm.fetcher = state.NewTestingRPCFetcher(vm.opts.tb, vm.opts.forkClient, new(big.Int).Sub(vm.opts.forkBlockNumber, big1))
		}
	}

	// set chain config
	vm.chainConfig = vm.opts.chainConfig
	if vm.chainConfig == nil {
		if vm.fetcher != nil {
			vm.chainConfig = params.MainnetChainConfig
		} else {
			vm.chainConfig = allEthashProtocolChanges
		}
	}

	vm.blockCtx = vm.opts.blockCtx
	if vm.blockCtx == nil {
		if vm.opts.header != nil {
			vm.blockCtx = newBlockContext(vm.opts.header, vm.fetcherHashFunc(vm.fetcher))
		} else {
			vm.blockCtx = defaultBlockContext()
		}
	}

	// set DB
	db := newDB(vm.fetcher)
	vm.db, _ = gethState.New(hash0, db, nil)
	for addr, acc := range vm.opts.preState {
		vm.db.SetNonce(addr, acc.Nonce)
		if acc.Balance != nil {
			vm.db.SetBalance(addr, acc.Balance)
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

func (vm *VM) Apply(msg *w3types.Message, tracers ...vm.EVMLogger) (*Receipt, error) {
	return vm.apply(msg, false, newMultiEVMLogger(tracers))
}

func (v *VM) apply(msg *w3types.Message, isCall bool, tracer vm.EVMLogger) (*Receipt, error) {
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
	evm := vm.NewEVM(*v.blockCtx, *txCtx, v.db, v.chainConfig, vm.Config{
		Tracer:    tracer,
		NoBaseFee: v.noBaseFee || isCall,
	})

	snap := v.db.Snapshot()

	// apply the message to the evm
	result, err := core.ApplyMessage(evm, coreMsg, gp)
	if err != nil {
		return nil, err
	}

	// build receipt
	receipt := &Receipt{
		f:        msg.Func,
		GasUsed:  result.UsedGas,
		GasLimit: result.UsedGas + v.db.GetRefund(),
		Output:   result.ReturnData,
		Logs:     v.db.GetLogs(txHash, 0, hash0),
	}

	if err := result.Err; err != nil {
		if reason, unpackErr := abi.UnpackRevert(result.ReturnData); unpackErr != nil {
			receipt.Err = err
		} else {
			receipt.Err = fmt.Errorf("%w: %s", err, reason)
		}
	}
	if msg.To == nil {
		contractAddr := crypto.CreateAddress(msg.From, msg.Nonce)
		receipt.ContractAddress = &contractAddr
	}

	if isCall && !result.Failed() {
		v.db.RevertToSnapshot(snap)
	}
	v.db.Finalise(false)

	return receipt, receipt.Err
}

func (vm *VM) Call(msg *w3types.Message, tracers ...vm.EVMLogger) (*Receipt, error) {
	return vm.apply(msg, true, newMultiEVMLogger(tracers))
}

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

// Nonce returns the nonce of Address addr.
func (vm *VM) Nonce(addr common.Address) (uint64, error) {
	nonce := vm.db.GetNonce(addr)
	if vm.db.Error() != nil {
		return 0, fmt.Errorf("%w: failed to fetch nonce of %s", ErrFetch, addr)
	}
	return nonce, nil
}

// Balance returns the balance of Address addr.
func (vm *VM) Balance(addr common.Address) (*big.Int, error) {
	balance := vm.db.GetBalance(addr)
	if vm.db.Error() != nil {
		return nil, fmt.Errorf("%w: failed to fetch balance of %s", ErrFetch, addr)
	}
	return balance, nil
}

// Code returns the code of Address addr.
func (vm *VM) Code(addr common.Address) ([]byte, error) {
	code := vm.db.GetCode(addr)
	if vm.db.Error() != nil {
		return nil, fmt.Errorf("%w: failed to fetch code of %s", ErrFetch, addr)
	}
	return code, nil
}

// StorageAt returns the state of Address addr at the give storage Hash slot.
func (vm *VM) StorageAt(addr common.Address, slot common.Hash) (common.Hash, error) {
	val := vm.db.GetState(addr, slot)
	if vm.db.Error() != nil {
		return hash0, fmt.Errorf("%w: failed to fetch storage of %s at %s", ErrFetch, addr, slot)
	}
	return val, nil
}

func (v *VM) buildMessage(msg *w3types.Message, skipAccChecks bool) (*core.Message, *vm.TxContext, error) {
	nonce := msg.Nonce
	if nonce == 0 && !skipAccChecks && msg.From != addr0 {
		var err error
		nonce, err = v.Nonce(msg.From)
		if err != nil {
			return nil, nil, err
		}
	}

	gasLimit := msg.Gas
	if maxGasLimit := v.blockCtx.GasLimit; gasLimit == 0 {
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
	if baseFee := v.blockCtx.BaseFee; baseFee != nil && baseFee.Sign() > 0 {
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
			SkipAccountChecks: skipAccChecks,
		},
		&vm.TxContext{
			Origin:   msg.From,
			GasPrice: gasPrice,
		},
		nil
}

func (vm *VM) fetcherHashFunc(fetcher state.Fetcher) vm.GetHashFunc {
	return func(n uint64) common.Hash {
		blockNumber := new(big.Int).SetUint64(n)
		hash, _ := fetcher.HeaderHash(blockNumber)
		return hash
	}
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

// An Option configures a VM.
type Option func(*VM)

// WithChainConfig configures the VM's chain config.
func WithChainConfig(cfg *params.ChainConfig) Option {
	return func(vm *VM) { vm.opts.chainConfig = cfg }
}

// WithBlockContext configures the VM's block context.
func WithBlockContext(ctx *vm.BlockContext) Option {
	return func(vm *VM) { vm.opts.blockCtx = ctx }
}

// WithState sets the VM's pre state.
func WithState(state w3types.State) Option {
	return func(vm *VM) { vm.opts.preState = state }
}

func WithNoBaseFee() Option {
	return func(vm *VM) { vm.opts.noBaseFee = true }
}

func WithFork(client *w3.Client, blockNumber *big.Int) Option {
	return func(vm *VM) {
		vm.opts.forkClient = client
		vm.opts.forkBlockNumber = blockNumber
	}
}

// WithHeader configures the VM's block context based on the given header.
func WithHeader(header *types.Header) Option {
	return func(vm *VM) { vm.opts.header = header }
}

func WithFetcher(fetcher state.Fetcher) Option {
	return func(vm *VM) { vm.opts.fetcher = fetcher }
}

func WithTB(tb testing.TB) Option {
	return func(vm *VM) { vm.opts.tb = tb }
}
