package w3vm

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/params"
	"github.com/lmittmann/w3"
	"github.com/lmittmann/w3/w3types"
	"github.com/lmittmann/w3/w3vm/state"
)

type VM struct {
	opts *vmOptions
}

type vmOptions struct {
	chainConfig     *params.ChainConfig
	preState        w3types.State
	blockCtx        *vm.BlockContext
	header          *types.Header
	forkClient      *w3.Client
	forkBlockNumber *big.Int
	fetcher         state.Fetcher
	tb              testing.TB
}

func New(opts ...Option) *VM {
	c := &VM{opts: new(vmOptions)}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func (vm *VM) Apply(msg *w3types.Message, tracers ...vm.EVMLogger) (*Receipt, error) {
	return nil, nil
}

func (vm *VM) Call(msg *w3types.Message, tracers ...vm.EVMLogger) (*Receipt, error) {
	return nil, nil
}

func (vm *VM) CallFunc(contract common.Address, f w3types.Func, args ...any) Returner {
	return &returner{
		vm: vm,
		msg: &w3types.Message{
			To:   &contract,
			Func: f,
			Args: args,
		},
	}
}

type Returner interface {
	Returns(...any) error
}

type returner struct {
	vm  *VM
	msg *w3types.Message
}

func (r *returner) Returns(returns ...any) error {
	receipt, err := r.vm.Call(r.msg)
	if err != nil {
		return err
	}
	return receipt.DecodeReturns(returns...)
}

// Nonce returns the nonce of Address addr.
func (vm *VM) Nonce(addr common.Address) uint64 {
	return 0
}

// Balance returns the balance of Address addr.
func (vm *VM) Balance(addr common.Address) *big.Int {
	return nil
}

// Code returns the code of Address addr.
func (vm *VM) Code(addr common.Address) []byte {
	return nil
}

// StorageAt returns the state of Address addr at the give storage Hash slot.
func (vm *VM) StorageAt(addr common.Address, slot common.Hash) common.Hash {
	return common.Hash{}
}

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
