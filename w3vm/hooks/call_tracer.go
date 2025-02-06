package hooks

import (
	"encoding/hex"
	"fmt"
	"io"
	"math/big"
	"reflect"
	"strings"
	"sync"

	"github.com/charmbracelet/lipgloss"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/lmittmann/w3"
	"github.com/lmittmann/w3/internal/fourbyte"
)

var (
	// styles
	styleDim           = lipgloss.NewStyle().Faint(true)
	styleTarget        = lipgloss.NewStyle().Foreground(lipgloss.Color("#EBFF71"))
	styleValue         = lipgloss.NewStyle().Foreground(lipgloss.Color("#71FF71"))
	styleRevert        = lipgloss.NewStyle().Foreground(lipgloss.Color("#FE5F86"))
	stylesStaticcall   = lipgloss.NewStyle().Faint(true)
	stylesDelegatecall = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFA500"))
)

// TargetAddress can be used to match the target (to) address in the TargetStyler
// of [CallTracerOptions].
var TargetAddress = common.BytesToAddress([]byte("target"))

// CallTracerOptions configures the CallTracer hook. A zero CallTracerOptions
// consists entirely of default values.
type CallTracerOptions struct {
	TargetStyler func(addr common.Address) lipgloss.Style
	targetAddr   common.Address

	ShowStaticcall bool

	ShowOp   func(op byte, pc uint64, addr common.Address) bool
	OpStyler func(op byte) lipgloss.Style

	DecodeABI bool
}

func (opts *CallTracerOptions) targetStyler(addr common.Address) lipgloss.Style {
	if addr == opts.targetAddr {
		addr = TargetAddress
	}

	if opts.TargetStyler == nil {
		return defaultTargetStyler(addr)
	}
	return opts.TargetStyler(addr)
}

func (opts *CallTracerOptions) showOp(op byte, pc uint64, addr common.Address) bool {
	if opts.ShowOp == nil {
		return false
	}
	return opts.ShowOp(op, pc, addr)
}

func (opts *CallTracerOptions) opStyler(op byte) lipgloss.Style {
	if opts.OpStyler == nil {
		return defaultOpStyler(op)
	}
	return opts.OpStyler(op)
}

func defaultTargetStyler(addr common.Address) lipgloss.Style {
	switch addr {
	case TargetAddress:
		return styleTarget
	default:
		return lipgloss.NewStyle()
	}
}

func defaultOpStyler(byte) lipgloss.Style {
	return lipgloss.NewStyle()
}

// NewCallTracer returns a new hook that writes to w and is configured with opts.
func NewCallTracer(w io.Writer, opts *CallTracerOptions) *tracing.Hooks {
	if opts == nil {
		opts = new(CallTracerOptions)
	}
	tracer := &callTracer{w: w, opts: opts}

	return &tracing.Hooks{
		OnEnter:  tracer.EnterHook,
		OnExit:   tracer.ExitHook,
		OnOpcode: tracer.OpcodeHook,
		OnLog:    tracer.OnLog,
	}
}

type callTracer struct {
	w    io.Writer
	opts *CallTracerOptions
	once sync.Once

	callStack []call

	// isInStaticcall is the number of nested staticcalls on the callStack.
	// It is only incremented if opts.ShowStatic is true.
	isInStaticcall int
}

func (c *callTracer) EnterHook(depth int, typ byte, from common.Address, to common.Address, input []byte, gas uint64, value *big.Int) {
	c.once.Do(func() {
		c.opts.targetAddr = to
	})

	var (
		fn           *w3.Func
		isPrecompile bool
	)
	if c.opts.DecodeABI && len(input) >= 4 {
		sig := ([4]byte)(input[:4])
		fn, isPrecompile = fourbyte.Function(sig, to)
	}

	callType := vm.OpCode(typ)
	defer func() { c.callStack = append(c.callStack, call{callType, to, fn}) }()
	if !c.opts.ShowStaticcall && callType == vm.STATICCALL {
		c.isInStaticcall++
	}
	if c.isInStaticcall > 0 {
		return
	}

	fmt.Fprintf(c.w, "%s%s %s%s%s\n", renderIdent(c.callStack, c.opts.targetStyler, 1), renderAddr(to, c.opts.targetStyler), renderCallType(typ), renderValue(c.opts.DecodeABI, value), renderInput(fn, isPrecompile, input, c.opts.targetStyler))
}

func (c *callTracer) ExitHook(depth int, output []byte, gasUsed uint64, err error, reverted bool) {
	call := c.callStack[len(c.callStack)-1]
	defer func() { c.callStack = c.callStack[:depth] }()

	if !c.opts.ShowStaticcall && call.Type == vm.STATICCALL {
		defer func() { c.isInStaticcall-- }()
	}
	if c.isInStaticcall > 0 {
		return
	}

	if reverted {
		fmt.Fprintf(c.w, "%s%s\n", renderIdent(c.callStack, c.opts.targetStyler, -1), styleRevert.Render(fmt.Sprintf("[%d]", gasUsed), renderRevert(err, output, c.opts.DecodeABI)))
	} else {
		fmt.Fprintf(c.w, "%s[%d] %s\n", renderIdent(c.callStack, c.opts.targetStyler, -1), gasUsed, renderOutput(call.Func, output, c.opts.targetStyler))
	}
}

func (c *callTracer) OpcodeHook(pc uint64, op byte, gas, cost uint64, scope tracing.OpContext, rData []byte, depth int, err error) {
	if c.isInStaticcall > 0 ||
		!c.opts.showOp(op, pc, scope.Address()) {
		return
	}
	fmt.Fprintln(c.w, renderIdent(c.callStack, c.opts.targetStyler, 0)+renderOp(op, c.opts.opStyler, pc, scope))
}

func (c *callTracer) OnLog(log *types.Log) {
	if c.isInStaticcall > 0 {
		return
	}
}

func renderIdent(callStack []call, styler func(addr common.Address) lipgloss.Style, kind int) (ident string) {
	for i, call := range callStack {
		style := styler(call.Target)

		s := "│ "
		if isLast := i == len(callStack)-1; isLast {
			switch kind {
			case 1:
				s = "├╴"
			case -1:
				s = "└╴"
			}
		}
		ident += style.Faint(true).Render(s)
	}
	return ident
}

func renderAddr(addr common.Address, styler func(addr common.Address) lipgloss.Style) string {
	return styler(addr).Render(addr.Hex())
}

func renderCallType(typ byte) string {
	switch vm.OpCode(typ) {
	case vm.CALL:
		return ""
	case vm.STATICCALL:
		return stylesStaticcall.Render("static") + " "
	case vm.DELEGATECALL:
		return stylesDelegatecall.Render("delegate") + " "
	case vm.CREATE:
		return "create "
	case vm.CREATE2:
		return "create2 "
	default:
		panic(fmt.Sprintf("unknown call type %92x", typ))
	}
}

func renderValue(decodeABI bool, val *big.Int) string {
	if val == nil || val.Sign() == 0 {
		return ""
	}
	if !decodeABI {
		return styleValue.Render(val.String(), "ETH") + " "
	}
	return styleValue.Render(w3.FromWei(val, 18), "ETH") + " "
}

func renderInput(fn *w3.Func, isPrecompile bool, input []byte, styler func(addr common.Address) lipgloss.Style) string {
	if fn != nil && len(input) >= 4 {
		s, err := renderAbiInput(fn, isPrecompile, input, styler)
		if err == nil {
			return s
		}
	}
	return renderRawInput(input, styler)
}

func renderOutput(fn *w3.Func, output []byte, styler func(addr common.Address) lipgloss.Style) string {
	if fn != nil && len(output) >= 4 {
		s, err := renderAbiOutput(fn, output, styler)
		if err == nil {
			return s
		}
	}
	return renderRawOutput(output, styler)
}

func renderRevert(revertErr error, output []byte, decodeABI bool) string {
	if decodeABI && len(output) >= 4 {
		sig := ([4]byte)(output[:4])
		fn, isPrecompile := fourbyte.Function(sig, w3.Addr0)
		if fn != nil && !isPrecompile {
			args, err := fn.Args.Unpack(output[4:])
			if err == nil {
				funcName := strings.Split(fn.Signature, "(")[0]
				return fmt.Sprintf("%s: %s(%s)", revertErr, funcName, renderAbiArgs(fn.Args, args, nil))
			}
		}
	}
	return fmt.Sprintf("%s: %x", revertErr, output)
}

func renderRawInput(input []byte, styler func(addr common.Address) lipgloss.Style) (s string) {
	s = "0x"
	if len(input)%32 == 4 {
		s += hex.EncodeToString(input[:4])
		for i := 4; i < len(input); i += 32 {
			s += renderWord(input[i:i+32], styler)
		}
	} else {
		s += hex.EncodeToString(input)
	}
	return
}

func renderRawOutput(output []byte, styler func(addr common.Address) lipgloss.Style) (s string) {
	s = "0x"
	if len(output)%32 == 0 {
		for i := 0; i < len(output); i += 32 {
			s += renderWord(output[i:i+32], styler)
		}
	} else {
		s += hex.EncodeToString(output)
	}
	return
}

func renderWord(word []byte, _ func(addr common.Address) lipgloss.Style) string {
	s := hex.EncodeToString(word)
	nonZeroWord := strings.TrimLeft(s, "0")
	if len(nonZeroWord) < len(s) {
		s = styleDim.Render(strings.Repeat("0", len(s)-len(nonZeroWord))) + nonZeroWord
	}
	return s
}

func renderAbiInput(fn *w3.Func, isPrecompile bool, input []byte, styler func(addr common.Address) lipgloss.Style) (string, error) {
	if !isPrecompile {
		input = input[4:]
	}

	args, err := fn.Args.Unpack(input)
	if err != nil {
		return "", err
	}

	funcName := strings.Split(fn.Signature, "(")[0]
	return funcName + "(" + renderAbiArgs(fn.Args, args, styler) + ")", nil
}

func renderAbiOutput(fn *w3.Func, output []byte, styler func(addr common.Address) lipgloss.Style) (string, error) {
	returns, err := fn.Returns.Unpack(output)
	if err != nil {
		return "", err
	}

	return renderAbiArgs(fn.Returns, returns, styler), nil
}

func renderAbiArgs(args abi.Arguments, vals []any, styler func(addr common.Address) lipgloss.Style) (s string) {
	for i, val := range vals {
		arg := args[i]
		s += renderAbiTyp(&arg.Type, arg.Name, val, styler)
		if i < len(vals)-1 {
			s += styleDim.Render(",") + " "
		}
	}
	return
}

func renderAbiTyp(typ *abi.Type, name string, val any, styler func(addr common.Address) lipgloss.Style) (s string) {
	if name != "" {
		s += styleDim.Render(name+":") + " "
	}

	if styler == nil {
		styler = func(addr common.Address) lipgloss.Style { return lipgloss.NewStyle() }
	}

	switch val := val.(type) {
	case []byte:
		s += "0x" + hex.EncodeToString(val)
	case [32]byte:
		s += "0x" + hex.EncodeToString(val[:])
	case [31]byte:
		s += "0x" + hex.EncodeToString(val[:])
	case [30]byte:
		s += "0x" + hex.EncodeToString(val[:])
	case [29]byte:
		s += "0x" + hex.EncodeToString(val[:])
	case [28]byte:
		s += "0x" + hex.EncodeToString(val[:])
	case [27]byte:
		s += "0x" + hex.EncodeToString(val[:])
	case [26]byte:
		s += "0x" + hex.EncodeToString(val[:])
	case [25]byte:
		s += "0x" + hex.EncodeToString(val[:])
	case [24]byte:
		s += "0x" + hex.EncodeToString(val[:])
	case [23]byte:
		s += "0x" + hex.EncodeToString(val[:])
	case [22]byte:
		s += "0x" + hex.EncodeToString(val[:])
	case [21]byte:
		s += "0x" + hex.EncodeToString(val[:])
	case [20]byte:
		s += "0x" + hex.EncodeToString(val[:])
	case [19]byte:
		s += "0x" + hex.EncodeToString(val[:])
	case [18]byte:
		s += "0x" + hex.EncodeToString(val[:])
	case [17]byte:
		s += "0x" + hex.EncodeToString(val[:])
	case [16]byte:
		s += "0x" + hex.EncodeToString(val[:])
	case [15]byte:
		s += "0x" + hex.EncodeToString(val[:])
	case [14]byte:
		s += "0x" + hex.EncodeToString(val[:])
	case [13]byte:
		s += "0x" + hex.EncodeToString(val[:])
	case [12]byte:
		s += "0x" + hex.EncodeToString(val[:])
	case [11]byte:
		s += "0x" + hex.EncodeToString(val[:])
	case [10]byte:
		s += "0x" + hex.EncodeToString(val[:])
	case [9]byte:
		s += "0x" + hex.EncodeToString(val[:])
	case [8]byte:
		s += "0x" + hex.EncodeToString(val[:])
	case [7]byte:
		s += "0x" + hex.EncodeToString(val[:])
	case [6]byte:
		s += "0x" + hex.EncodeToString(val[:])
	case [5]byte:
		s += "0x" + hex.EncodeToString(val[:])
	case [4]byte:
		s += "0x" + hex.EncodeToString(val[:])
	case [3]byte:
		s += "0x" + hex.EncodeToString(val[:])
	case [2]byte:
		s += "0x" + hex.EncodeToString(val[:])
	case [1]byte:
		s += "0x" + hex.EncodeToString(val[:])
	case common.Address:
		style := styler(val)
		s += style.Render(val.Hex())
	case common.Hash:
		s += val.Hex()
	case any: // tuple, array, or slice
		switch refVal := reflect.ValueOf(val); refVal.Kind() {
		case reflect.Slice:
			s += "["
			for i := range refVal.Len() {
				s += renderAbiTyp(typ.Elem, "", refVal.Index(i).Interface(), styler)

				if i < refVal.Len()-1 {
					s += styleDim.Render(",") + " "
				}
			}
			s += "]"
		case reflect.Array:
			s += "["
			for i := range refVal.Len() {
				s += renderAbiTyp(typ.Elem, "", refVal.Index(i).Interface(), styler)

				if i < refVal.Len()-1 {
					s += styleDim.Render(",") + " "
				}
			}
			s += "]"
		case reflect.Struct:
			s += "("
			for i := range refVal.NumField() {
				s += renderAbiTyp(typ.TupleElems[i], typ.TupleRawNames[i], refVal.Field(i).Interface(), styler)

				if i < refVal.NumField()-1 {
					s += styleDim.Render(",") + " "
				}
			}
			s += ")"
		default:
			s += fmt.Sprintf("%v", val)
		}
	default:
		s += fmt.Sprintf("%v", val)
	}
	return s
}

func renderOp(op byte, style func(byte) lipgloss.Style, pc uint64, scope tracing.OpContext) string {
	const maxStackDepth = 7
	sb := new(strings.Builder)
	sb.WriteString(styleDim.Render(fmt.Sprintf("0x%04x ", pc)))
	sb.WriteString(style(op).Render(fmt.Sprintf("%-12s ", vm.OpCode(op))))

	stack := scope.StackData()
	for i, j := len(stack)-1, 0; i >= 0 && i >= len(stack)-maxStackDepth; i, j = i-1, j+1 {
		notLast := i > 0 && i > len(stack)-maxStackDepth
		if isAccessed := opAccessesStackElem(op, j); isAccessed {
			sb.WriteString(stack[i].Hex())
		} else {
			sb.WriteString(styleDim.Render(stack[i].Hex()))
		}
		if notLast {
			sb.WriteString(" ")
		}
	}

	if len(stack) > maxStackDepth {
		sb.WriteString(styleDim.Render(fmt.Sprintf(" …%d", len(stack)-maxStackDepth)))
	}

	return sb.String()
}

// call stores state of the current call in execution.
type call struct {
	Type   vm.OpCode
	Target common.Address
	Func   *w3.Func
}

// opAccessesStackElem returns true, if the given opcode accesses the stack at
// index i, otherwise false.
func opAccessesStackElem(op byte, i int) bool {
	switch {
	case byte(vm.SWAP1) <= op && op <= byte(vm.SWAP16):
		return i == 0 || i == int(op)-int(vm.SWAP1)+1
	case byte(vm.DUP1) <= op && op <= byte(vm.DUP16):
		return i == int(op)-int(vm.DUP1)
	default:
		return i < pops[op]
	}
}

var pops = [256]int{
	vm.STOP: 0, vm.ADD: 2, vm.MUL: 2, vm.SUB: 2, vm.DIV: 2, vm.SDIV: 2, vm.MOD: 2, vm.SMOD: 2, vm.ADDMOD: 3, vm.MULMOD: 3, vm.EXP: 2, vm.SIGNEXTEND: 2,
	vm.LT: 2, vm.GT: 2, vm.SLT: 2, vm.SGT: 2, vm.EQ: 2, vm.ISZERO: 1, vm.AND: 2, vm.OR: 2, vm.XOR: 2, vm.NOT: 1, vm.BYTE: 2, vm.SHL: 2, vm.SHR: 2, vm.SAR: 2,
	vm.KECCAK256: 2,
	vm.BALANCE:   1, vm.CALLDATALOAD: 1, vm.CALLDATACOPY: 3, vm.CODECOPY: 3, vm.EXTCODESIZE: 1, vm.EXTCODECOPY: 4, vm.RETURNDATACOPY: 3, vm.EXTCODEHASH: 1,
	vm.BLOCKHASH: 1, vm.BLOBHASH: 1,
	vm.POP: 1, vm.MLOAD: 1, vm.MSTORE: 2, vm.MSTORE8: 2, vm.SLOAD: 1, vm.SSTORE: 2, vm.JUMP: 1, vm.JUMPI: 2, vm.TLOAD: 1, vm.TSTORE: 2, vm.MCOPY: 3,
	vm.LOG0: 2, vm.LOG1: 3, vm.LOG2: 4, vm.LOG3: 5, vm.LOG4: 6,
	vm.CREATE: 3, vm.CALL: 7, vm.CALLCODE: 7, vm.RETURN: 2, vm.DELEGATECALL: 6, vm.CREATE2: 4, vm.STATICCALL: 6, vm.REVERT: 2, vm.SELFDESTRUCT: 1,
}
