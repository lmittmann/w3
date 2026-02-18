package hooks

import (
	"bytes"
	"fmt"
	"io"
	"iter"
	"math/big"
	"os"
	"slices"

	"github.com/charmbracelet/lipgloss"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/lmittmann/w3"
	"github.com/lmittmann/w3/internal/fourbyte"
)

type CallFrame struct {
	Type  vm.OpCode
	Depth int
	PC    uint64

	From    common.Address
	Gas     uint64
	GasUsed uint64
	To      common.Address
	Value   *big.Int
	Input   []byte
	Output  []byte

	Reverted bool
	Err      error
	Logs     []*types.Log
	Calls    []*CallFrame
}

func (cf *CallFrame) Iter() iter.Seq[*CallFrame] {
	return func(yield func(*CallFrame) bool) {
		stack := []*CallFrame{cf}
		for len(stack) > 0 {
			peek := stack[len(stack)-1]
			stack = stack[:len(stack)-1] // pop
			if !yield(peek) {
				return
			}

			for _, call := range slices.Backward(peek.Calls) {
				stack = append(stack, call) // push
			}
		}
	}
}

func (cf *CallFrame) IterLogs() iter.Seq[*types.Log] {
	return func(yield func(*types.Log) bool) {
		logs := make([]*types.Log, 0)
		for call := range cf.Iter() {
			logs = append(logs, call.Logs...)
		}

		slices.SortFunc(logs, func(a, b *types.Log) int {
			return int(a.Index - b.Index)
		})

		for _, log := range logs {
			if !yield(log) {
				return
			}
		}
	}
}

type CallFrameTracer struct {
	Call *CallFrame

	lastPC uint64
	logIdx int
	stack  []*CallFrame
}

func NewCallFrameTracer() *CallFrameTracer { return new(CallFrameTracer) }

func (t *CallFrameTracer) Hooks() *tracing.Hooks {
	t.lastPC = 0
	t.logIdx = 0
	if t.stack == nil {
		t.stack = make([]*CallFrame, 0)
	} else {
		t.stack = t.stack[:0]
	}

	return &tracing.Hooks{
		OnEnter:  t.onEnter,
		OnExit:   t.onExit,
		OnOpcode: t.onOpcodes,
		OnLog:    t.onLog,
	}
}

func (t *CallFrameTracer) onEnter(depth int, typ byte, from common.Address, to common.Address, input []byte, gas uint64, value *big.Int) {
	call := &CallFrame{
		Type:  vm.OpCode(typ),
		Depth: depth,
		PC:    t.lastPC,
		From:  from,
		To:    to,
		Input: bytes.Clone(input),
		Gas:   gas,
		Value: value,
	}
	if depth == 0 {
		t.Call = call
	} else {
		peek := t.stack[len(t.stack)-1]
		peek.Calls = append(peek.Calls, call)
	}

	t.stack = append(t.stack, call) // push
}

func (t *CallFrameTracer) onExit(depth int, output []byte, gasUsed uint64, err error, reverted bool) {
	peek := t.stack[len(t.stack)-1]
	peek.GasUsed = gasUsed
	peek.Output = bytes.Clone(output)
	peek.Reverted = reverted
	peek.Err = err
	t.stack = t.stack[:len(t.stack)-1] // pop
}

func (t *CallFrameTracer) onOpcodes(pc uint64, op byte, gas, cost uint64, scope tracing.OpContext, rData []byte, depth int, err error) {
	t.lastPC = pc
}

func (t *CallFrameTracer) onLog(log *types.Log) {
	peek := t.stack[len(t.stack)-1]
	peek.Logs = append(peek.Logs, log)
}

type PrintOptions struct {
	TargetStyler func(addr common.Address) lipgloss.Style
	targetAddr   common.Address

	ShowStaticcall bool
	ShowEvent      bool

	DecodeABI bool
}

func (opts *PrintOptions) targetStyler(addr common.Address) lipgloss.Style {
	if addr == opts.targetAddr {
		addr = TargetAddress
	}

	if opts.TargetStyler == nil {
		return defaultTargetStyler(addr)
	}
	return opts.TargetStyler(addr)
}

func Print(cf *CallFrame, ops *PrintOptions) {
	Fprint(os.Stdout, cf, ops)
}

func Fprint(w io.Writer, cf *CallFrame, opts *PrintOptions) {
	opts.targetAddr = cf.To

	callStack := make([]call, 0)
	prettyPrint(w, cf, callStack, opts)
}

func prettyPrint(w io.Writer, cf *CallFrame, callStack []call, opts *PrintOptions) {
	var (
		fn           *w3.Func
		isPrecompile bool
	)

	if opts.DecodeABI && len(cf.Input) >= 4 {
		sig := ([4]byte)(cf.Input[:4])
		fn, isPrecompile = fourbyte.Function(sig, cf.To)
	}

	// print call start
	fmt.Fprint(w,
		renderIdent(callStack, opts.targetStyler, 1),
		renderAddr(cf.To, opts.targetStyler),
		" ",
		renderCallType(byte(cf.Type)),
		renderValue(opts.DecodeABI, cf.Value),
		renderInput(fn, isPrecompile, cf.Input, opts.targetStyler),
		"\n",
	)

	// push call to callStack
	callStack = append(callStack, call{cf.Type, cf.To, fn})

	for _, call := range cf.Calls {
		if call.Type == vm.STATICCALL && !opts.ShowStaticcall {
			continue
		}
		prettyPrint(w, call, callStack, opts)
	}

	// print call end
	fmt.Fprint(w, renderIdent(callStack, opts.targetStyler, -1))
	gasUsed := fmt.Sprintf("[%d]", cf.GasUsed)
	if cf.Reverted {
		fmt.Fprint(w,
			styleRevert.Render(gasUsed),
			" ",
			renderRevert(cf.Err, cf.Output, opts.DecodeABI),
		)
	} else {
		fmt.Fprint(w,
			gasUsed,
			" ",
			renderOutput(fn, cf.Output, opts.targetStyler),
		)
	}
	fmt.Fprintln(w)
}
