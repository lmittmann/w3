package w3

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"

	_abi "github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/lmittmann/w3/core"
	"github.com/lmittmann/w3/internal/abi"
)

var (
	ErrInvalidABI       = errors.New("w3: invalid ABI")
	ErrArgumentMismatch = errors.New("w3: argument mismatch")
	ErrReturnsMismatch  = errors.New("w3: returns mismatch")
	ErrInvalidType      = errors.New("w3: invalid type")
	ErrEvmRevert        = errors.New("w3: evm reverted")

	revertSelector       = crypto.Keccak256([]byte("Error(string)"))[:4]
	approveSelector      = crypto.Keccak256([]byte("approve(address,uint256)"))[:4]
	transferSelector     = crypto.Keccak256([]byte("transfer(address,uint256)"))[:4]
	transferFromSelector = crypto.Keccak256([]byte("transferFrom(address,address,uint256)"))[:4]
)

type abiFunc struct {
	Signature string
	Selector  []byte         // four-byte selector
	Args      _abi.Arguments // input
	Returns   _abi.Arguments // output
}

// NewFunc returns a new Smart Contract function ABI binding from the given
// Solidity function signature and its returns.
//
// An error is returned if the signature or returns parsing fails.
func NewFunc(signature, returns string) (core.Func, error) {
	args, err := abi.Parse(signature)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidABI, err)
	}
	if args.FuncName == "" {
		return nil, fmt.Errorf("%w: missing function name", ErrInvalidABI)
	}

	returnArgs, err := abi.Parse(returns)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidABI, err)
	}
	if returnArgs.FuncName != "" {
		return nil, fmt.Errorf("%w: returns must not have a function name", ErrInvalidABI)
	}

	return &abiFunc{
		Signature: args.Sig,
		Selector:  crypto.Keccak256([]byte(args.Sig))[:4],
		Args:      args.Args,
		Returns:   returnArgs.Args,
	}, nil
}

// MustNewFunc is like NewFunc but panics if the signature or returns parsing
// fails.
func MustNewFunc(signature, returns string) core.Func {
	fn, err := NewFunc(signature, returns)
	if err != nil {
		panic(err)
	}
	return fn
}

// EncodeArgs ABI-encodes the given args and prepends the Func's four-byte
// selector.
func (f *abiFunc) EncodeArgs(args ...interface{}) ([]byte, error) {
	if len(f.Args) != len(args) {
		return nil, fmt.Errorf("%w: expected %d arguments, got %d", ErrArgumentMismatch, len(f.Args), len(args))
	}

	input, err := f.Args.PackValues(args)
	if err != nil {
		return nil, err
	}

	return append(f.Selector, input...), nil
}

// DecodeArgs ABI-decodes the given input to the given args.
func (f *abiFunc) DecodeArgs(input []byte, args ...interface{}) error {
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
func (f *abiFunc) DecodeReturns(output []byte, returns ...interface{}) error {
	if len(f.Returns) != len(returns) {
		return fmt.Errorf("%w: expected %d returns, got %d", ErrReturnsMismatch, len(f.Returns), len(returns))
	}

	// check the output for a revert reason
	if bytes.HasPrefix(output, revertSelector) {
		if reason, err := _abi.UnpackRevert(output); err != nil {
			return err
		} else {
			return fmt.Errorf("%w: %s", ErrEvmRevert, reason)
		}
	}

	// Gracefully handle uncompliant ERC20 returns
	if len(output) == 0 &&
		(bytes.Equal(f.Selector, approveSelector) ||
			bytes.Equal(f.Selector, transferSelector) ||
			bytes.Equal(f.Selector, transferFromSelector)) &&
		len(returns) == 1 {

		if err := copyVal(_abi.BoolTy, returns[0], true); err != nil {
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
	case _abi.TupleTy:
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
