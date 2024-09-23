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

type CallTracerOptions struct {
	TargetStyler func(addr common.Address) lipgloss.Style
	targetAddr   common.Address

	DecodeABI bool

	NoColor bool
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

var TargetAddress = common.BytesToAddress([]byte("target"))

func defaultTargetStyler(addr common.Address) lipgloss.Style {
	switch addr {
	case TargetAddress:
		return styleTarget
	default:
		return lipgloss.NewStyle()
	}
}

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
}

func (c *callTracer) EnterHook(depth int, typ byte, from common.Address, to common.Address, input []byte, gas uint64, value *big.Int) {
	c.once.Do(func() {
		c.opts.targetAddr = to
	})

	var fn *w3.Func
	if c.opts.DecodeABI && len(input) >= 4 {
		sig := ([4]byte)(input[:4])
		fn = fourbyte.Function(sig)
	}
	defer func() { c.callStack = append(c.callStack, call{to, fn}) }()

	fmt.Fprintf(c.w, "%s%s %s%s%s\n", renderIdent(c.callStack, c.opts.targetStyler, 1), renderAddr(to, c.opts.targetStyler), renderCallType(typ), renderValue(c.opts.DecodeABI, value), renderInput(fn, input, c.opts.targetStyler))
}

func (c *callTracer) ExitHook(depth int, output []byte, gasUsed uint64, err error, reverted bool) {
	call := c.callStack[len(c.callStack)-1]
	defer func() { c.callStack = c.callStack[:depth] }()

	if reverted {
		reason, err := abi.UnpackRevert(output)
		if err != nil {
			reason = hex.EncodeToString(output)
		}
		fmt.Fprintf(c.w, "%s%s\n", renderIdent(c.callStack, c.opts.targetStyler, -1), styleRevert.Render(fmt.Sprintf("[%d]", gasUsed), err.Error()+":", reason))
	} else {
		fmt.Fprintf(c.w, "%s[%d] %s\n", renderIdent(c.callStack, c.opts.targetStyler, -1), gasUsed, renderOutput(call.Func, output, c.opts.targetStyler))
	}
}

func (c *callTracer) OpcodeHook(pc uint64, op byte, gas, cost uint64, scope tracing.OpContext, rData []byte, depth int, err error) {
}

func (c *callTracer) OnLog(log *types.Log) {}

func renderIdent(callStack []call, styler func(addr common.Address) lipgloss.Style, kind int) (ident string) {
	for i, call := range callStack {
		style := styler(call.Target)

		s := "│ "
		if isLast := i == len(callStack)-1; isLast {
			if kind > 0 {
				s = "├╴"
			} else {
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

func renderInput(fn *w3.Func, input []byte, styler func(addr common.Address) lipgloss.Style) string {
	if fn != nil && len(input) >= 4 {
		s, err := renderAbiInput(fn, input, styler)
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

func renderAbiInput(fn *w3.Func, input []byte, styler func(addr common.Address) lipgloss.Style) (string, error) {
	args, err := fn.Args.Unpack(input[4:])
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

type call struct {
	Target common.Address
	Func   *w3.Func
}
