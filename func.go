package w3

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/crypto"
	_abi "github.com/lmittmann/w3/internal/abi"
)

var (
	ErrInvalidABI       = errors.New("w3: invalid ABI")
	ErrArgumentMismatch = errors.New("w3: argument mismatch")
	ErrReturnsMismatch  = errors.New("w3: returns mismatch")
	ErrInvalidType      = errors.New("w3: invalid type")
	ErrEvmRevert        = errors.New("w3: evm reverted")

	revertSelector       = selector("Error(string)")
	approveSelector      = selector("approve(address,uint256)")
	transferSelector     = selector("transfer(address,uint256)")
	transferFromSelector = selector("transferFrom(address,address,uint256)")
)

// Func represents a Smart Contract function ABI binding.
//
// Func implements the core.Func interface.
type Func struct {
	Signature string        // Function signature
	Selector  [4]byte       // 4-byte selector
	Args      abi.Arguments // Input arguments
	Returns   abi.Arguments // Output returns
}

// NewFunc returns a new Smart Contract function ABI binding from the given
// Solidity function signature and its returns.
//
// An error is returned if the signature or returns parsing fails.
func NewFunc(signature, returns string) (*Func, error) {
	args, err := _abi.Parse(signature)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidABI, err)
	}
	if args.FuncName == "" {
		return nil, fmt.Errorf("%w: missing function name", ErrInvalidABI)
	}

	returnArgs, err := _abi.Parse(returns)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidABI, err)
	}
	if returnArgs.FuncName != "" {
		return nil, fmt.Errorf("%w: returns must not have a function name", ErrInvalidABI)
	}

	return &Func{
		Signature: args.Sig,
		Selector:  selector(args.Sig),
		Args:      args.Args,
		Returns:   returnArgs.Args,
	}, nil
}

// MustNewFunc is like NewFunc but panics if the signature or returns parsing
// fails.
func MustNewFunc(signature, returns string) *Func {
	fn, err := NewFunc(signature, returns)
	if err != nil {
		panic(err)
	}
	return fn
}

// EncodeArgs ABI-encodes the given args and prepends the Func's 4-byte
// selector.
func (f *Func) EncodeArgs(args ...interface{}) ([]byte, error) {
	if len(f.Args) != len(args) {
		return nil, fmt.Errorf("%w: expected %d arguments, got %d", ErrArgumentMismatch, len(f.Args), len(args))
	}

	input, err := f.Args.PackValues(args)
	if err != nil {
		return nil, err
	}

	return append(f.Selector[:], input...), nil
}

// DecodeArgs ABI-decodes the given input to the given args.
func (f *Func) DecodeArgs(input []byte, args ...interface{}) error {
	if len(f.Args) != len(args) {
		return fmt.Errorf("%w: expected %d arguments, got %d", ErrArgumentMismatch, len(f.Args), len(args))
	}

	values, err := f.Args.UnpackValues(input[4:])
	if err != nil {
		return err
	}
	for i, val := range values {
		if err := copyVal(f.Args[i].Type.T, args[i], val); err != nil {
			return err
		}
	}

	return nil
}

// DecodeReturns ABI-decodes the given output to the given returns.
func (f *Func) DecodeReturns(output []byte, returns ...interface{}) error {
	if len(f.Returns) != len(returns) {
		return fmt.Errorf("%w: expected %d returns, got %d", ErrReturnsMismatch, len(f.Returns), len(returns))
	}

	// check the output for a revert reason
	if bytes.HasPrefix(output, revertSelector[:]) {
		if reason, err := abi.UnpackRevert(output); err != nil {
			return err
		} else {
			return fmt.Errorf("%w: %s", ErrEvmRevert, reason)
		}
	}

	// Gracefully handle uncompliant ERC20 returns
	if len(returns) == 1 && len(output) == 0 &&
		(f.Selector == approveSelector ||
			f.Selector == transferSelector ||
			f.Selector == transferFromSelector) {

		if err := copyVal(abi.BoolTy, returns[0], true); err != nil {
			return err
		}
		return nil
	}

	values, err := f.Returns.UnpackValues(output)
	if err != nil {
		return err
	}
	for i, val := range values {
		if err := copyVal(f.Returns[i].Type.T, returns[i], val); err != nil {
			return err
		}
	}

	return nil
}

func copyVal(t byte, dst, src interface{}) (err error) {
	// skip copying if dst is nil
	if dst == nil {
		return
	}

	rDst := reflect.ValueOf(dst)
	rSrc := reflect.ValueOf(src)

	switch t {
	case abi.TupleTy:
		err = copyTuple(rDst, rSrc)
	default:
		err = copyNonTuple(rDst, rSrc)
	}
	return
}

func copyNonTuple(rDst, rSrc reflect.Value) error {
	if rDst.Kind() != reflect.Ptr {
		return fmt.Errorf("%w: can not copy to non-pointer value", ErrInvalidType)
	}
	if rDst.IsNil() {
		return fmt.Errorf("%w: requires non-nil pointer", ErrInvalidType)
	}

	if !(rDst.Type().AssignableTo(rSrc.Type()) ||
		rDst.Elem().Type().AssignableTo(rSrc.Type())) {
		return fmt.Errorf("%w: can not copy %v to %v", ErrInvalidType, rSrc.Type(), rDst.Type())
	}

	if rSrc.Kind() == reflect.Ptr {
		rDst.Elem().Set(rSrc.Elem())
	} else {
		rDst.Elem().Set(rSrc)
	}
	return nil
}

func copyTuple(rDst, rSrc reflect.Value) error {
	if !(rDst.Kind() == reflect.Ptr && rDst.Elem().Kind() == reflect.Struct) {
		return fmt.Errorf("%w: can not copy to non-pointer value", ErrInvalidType)
	}
	if rSrc.Kind() != reflect.Struct {
		return fmt.Errorf("%w: tuple is no struct", ErrInvalidType)
	}

	for i := 0; i < rSrc.NumField(); i++ {
		fieldName := rSrc.Type().Field(i).Name
		rDst.Elem().FieldByName(fieldName).Set(rSrc.Field(i))
	}
	return nil
}

// selector returns the 4-byte selector of the given signature.
func selector(signature string) (selector [4]byte) {
	copy(selector[:], crypto.Keccak256([]byte(signature)))
	return
}
