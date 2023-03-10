package abi

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/lmittmann/w3/internal/crypto"
)

// Arguments represents a slice of [abi.Argument]'s.
type Arguments []abi.Argument

func (a Arguments) Signature() string {
	if len(a) <= 0 {
		return ""
	}

	fields := make([]string, len(a))
	for i, arg := range a {
		fields[i] = typeToString(&arg.Type)
	}
	return strings.Join(fields, ",")
}

func (a Arguments) SignatureWithName(name string) string {
	return name + "(" + a.Signature() + ")"
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

	for i, arg := range args {
		// discard if arg is nil
		if arg == nil {
			continue
		}

		if err := Copy(arg, values[i]); err != nil {
			return err
		}
	}
	return nil
}

// typeToString returns the string representation of a [abi.Type].
func typeToString(t *abi.Type) string {
	switch t.T {
	case abi.IntTy:
		return "int" + strconv.Itoa(t.Size)
	case abi.UintTy:
		return "uint" + strconv.Itoa(t.Size)
	case abi.BoolTy:
		return "bool"
	case abi.StringTy:
		return "string"
	case abi.SliceTy:
		return typeToString(t.Elem) + "[]"
	case abi.ArrayTy:
		return typeToString(t.Elem) + "[" + strconv.Itoa(t.Size) + "]"
	case abi.TupleTy:
		fields := make([]string, len(t.TupleElems))
		for i, elem := range t.TupleElems {
			fields[i] = typeToString(elem)
		}
		return "(" + strings.Join(fields, ",") + ")"
	case abi.AddressTy:
		return "address"
	case abi.FixedBytesTy:
		return "bytes" + strconv.Itoa(t.Size)
	case abi.BytesTy:
		return "bytes"
	case abi.HashTy:
		return "hash"
	default:
		panic(fmt.Sprintf("unsupported type %v", t))
	}
}
