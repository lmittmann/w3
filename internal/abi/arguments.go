package abi

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/lmittmann/w3/internal/crypto"
)

var (
	errInvalidType = errors.New("abi: invalid type")
)

// Arguments represents a slice of abi.Argument's.
type Arguments []abi.Argument

// Parse parses the given Solidity function/event signature and returns its
// name and arguments.
func Parse(s string) (name string, a Arguments, err error) {
	name, args, err := parse(s)
	if err != nil {
		return "", nil, err
	}

	a = (Arguments)(args)
	return
}

func (a Arguments) Signature() string {
	if len(a) <= 0 {
		return ""
	}

	fields := make([]string, len(a))
	for i, arg := range a {
		fields[i] = typeToString(arg.Type)
	}
	return strings.Join(fields, ",")
}

func (a Arguments) SignatureWithName(name string) string {
	if len(a) <= 0 {
		return name + "()"
	}

	fields := make([]string, len(a))
	for i, arg := range a {
		fields[i] = typeToString(arg.Type)
	}
	return name + "(" + strings.Join(fields, ",") + ")"
}

// Encode ABI-encodes the given arguments args.
func (a Arguments) Encode(args ...any) ([]byte, error) {
	data, err := abi.Arguments(a).PackValues(args)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// EncodeWithSelector ABI-encodes the given arguments args prepended by the
// given selector.
func (a Arguments) EncodeWithSelector(selector [4]byte, args ...any) ([]byte, error) {
	data, err := a.Encode(args...)
	if err != nil {
		return nil, err
	}

	data = append(selector[:], data...)
	return data, nil
}

// EncodeWithSignature ABI-encodes the given arguments args prepended by the
// first 4 bytes of the hash of the given signature.
func (a Arguments) EncodeWithSignature(signature string, args ...any) ([]byte, error) {
	var selector [4]byte
	copy(selector[:], crypto.Keccak256([]byte(signature))[:4])

	return a.EncodeWithSelector(selector, args...)
}

// Decode ABI-decodes the given data to the given arguments args.
func (a Arguments) Decode(data []byte, args ...any) error {
	values, err := abi.Arguments(a).UnpackValues(data)
	if err != nil {
		return err
	}

	for i, val := range values {
		if err := copyVal(a[i].Type.T, args[i], val); err != nil {
			return err
		}
	}

	return nil
}

// typeToTypeString maps from a abi.Type t to its string representation.
func typeToString(t abi.Type) string {
	switch t.T {
	case abi.IntTy:
		return fmt.Sprintf("int%d", t.Size)
	case abi.UintTy:
		return fmt.Sprintf("uint%d", t.Size)
	case abi.BoolTy:
		return "bool"
	case abi.StringTy:
		return "string"
	case abi.SliceTy:
		return typeToString(*t.Elem) + "[]"
	case abi.ArrayTy:
		return typeToString(*t.Elem) + fmt.Sprintf("[%d]", t.Size)
	case abi.TupleTy:
		fields := make([]string, len(t.TupleElems))
		for i, elem := range t.TupleElems {
			fields[i] = typeToString(*elem)
		}
		return "(" + strings.Join(fields, ",") + ")"
	case abi.AddressTy:
		return "address"
	case abi.FixedBytesTy:
		return fmt.Sprintf("bytes%d", t.Size)
	case abi.BytesTy:
		return "bytes"
	case abi.HashTy:
		return "hash"
	default:
		panic(fmt.Sprintf("unsupported type %v", t))
	}
}

func copyVal(t byte, dst, src any) (err error) {
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
		return fmt.Errorf("%w: can not copy to non-pointer value", errInvalidType)
	}
	if rDst.IsNil() {
		return fmt.Errorf("%w: requires non-nil pointer", errInvalidType)
	}

	if !(rDst.Type().AssignableTo(rSrc.Type()) ||
		rDst.Elem().Type().AssignableTo(rSrc.Type())) {
		return fmt.Errorf("%w: can not copy %v to %v", errInvalidType, rSrc.Type(), rDst.Type())
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
		return fmt.Errorf("%w: can not copy to non-pointer value", errInvalidType)
	}
	if rSrc.Kind() != reflect.Struct {
		return fmt.Errorf("%w: tuple is no struct", errInvalidType)
	}

	for i := 0; i < rSrc.NumField(); i++ {
		fieldName := rSrc.Type().Field(i).Name
		if rDst.Elem().FieldByName(fieldName).Kind() == reflect.Invalid {
			return fmt.Errorf("%w: field %q does not exist on dest struct", errInvalidType, fieldName)
		}
		rDst.Elem().FieldByName(fieldName).Set(rSrc.Field(i))
	}
	return nil
}
