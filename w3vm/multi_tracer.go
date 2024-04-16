package w3vm

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
)

// multiEVMLogger is a wrapper for multiple EVMLogger's.
type multiEVMLogger []vm.EVMLogger

func newMultiEVMLogger(tracers []vm.EVMLogger) vm.EVMLogger {
	// hot path
	switch len(tracers) {
	case 0:
		return nil
	case 1:
		return tracers[0]
	}

	// filter nil tracers
	// NOTE: this edits the tracers slice in place.
	var j int
	for i := range tracers {
		if tracers[i] == nil {
			continue
		} else if i > j {
			tracers[j], tracers[i] = tracers[i], tracers[j]
		}
		j++
	}
	tracers = tracers[:j]

	switch len(tracers) {
	case 0:
		return nil
	case 1:
		return tracers[0]
	default:
		return multiEVMLogger(tracers)
	}
}

func (m multiEVMLogger) CaptureTxStart(gasLimit uint64) {
	for _, tracer := range m {
		tracer.CaptureTxStart(gasLimit)
	}
}

func (m multiEVMLogger) CaptureTxEnd(restGas uint64) {
	for _, tracer := range m {
		tracer.CaptureTxEnd(restGas)
	}
}

func (m multiEVMLogger) CaptureStart(env *vm.EVM, from common.Address, to common.Address, create bool, input []byte, gas uint64, value *big.Int) {
	for _, tracer := range m {
		tracer.CaptureStart(env, from, to, create, input, gas, value)
	}
}

func (m multiEVMLogger) CaptureEnd(output []byte, gasUsed uint64, err error) {
	for _, tracer := range m {
		tracer.CaptureEnd(output, gasUsed, err)
	}
}

func (m multiEVMLogger) CaptureEnter(typ vm.OpCode, from common.Address, to common.Address, input []byte, gas uint64, value *big.Int) {
	for _, tracer := range m {
		tracer.CaptureEnter(typ, from, to, input, gas, value)
	}
}

func (m multiEVMLogger) CaptureExit(output []byte, gasUsed uint64, err error) {
	for _, tracer := range m {
		tracer.CaptureExit(output, gasUsed, err)
	}
}

func (m multiEVMLogger) CaptureState(pc uint64, op vm.OpCode, gas, cost uint64, scope *vm.ScopeContext, rData []byte, depth int, err error) {
	for _, tracer := range m {
		tracer.CaptureState(pc, op, gas, cost, scope, rData, depth, err)
	}
}

func (m multiEVMLogger) CaptureFault(pc uint64, op vm.OpCode, gas, cost uint64, scope *vm.ScopeContext, depth int, err error) {
	for _, tracer := range m {
		tracer.CaptureFault(pc, op, gas, cost, scope, depth, err)
	}
}

func (m multiEVMLogger) CaptureSystemTxEnd(intrinsicGas uint64) {

}
