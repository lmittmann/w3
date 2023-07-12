package w3vm

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/params"
	"github.com/lmittmann/w3/w3types"
)

type VM struct{}

func New(opts ...Option) *VM {
	c := &VM{}
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

func WithChainConfig(cfg *params.ChainConfig) Option {
	return func(vm *VM) {}
}

func WithBlockContext(ctx *vm.BlockContext) Option {
	return func(vm *VM) {}
}

func WithState(s w3types.State) Option {
	return func(vm *VM) {}
}
