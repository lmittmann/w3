package abi

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
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

		// check if arg is valid
		dst := reflect.ValueOf(arg)
		if dst.Kind() != reflect.Ptr {
			return fmt.Errorf("abi: decode non-pointer %T", arg)
		}
		if dst.IsNil() {
			return fmt.Errorf("abi: decode nil")
		}

		var err error
		switch a[0].Type.T {
		case abi.TupleTy:
			err = copyTuple(arg, values[i])
		default:
			err = copyNonTuple(arg, values[i])
		}
		if err != nil {
			return fmt.Errorf("abi: %w", err)
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

func copyNonTuple(dst, src any) error {
	dstVal := reflect.ValueOf(dst)
	if dstVal.Elem().Kind() == reflect.Ptr || dstVal.Elem().Kind() == reflect.Slice {
		// dst is a pointer to a pointer
		dstVal = dstVal.Elem()
	}

	var err error
	switch val := src.(type) {
	case uint8:
		err = setVal(dstVal, reflect.ValueOf(&val))
	case uint16:
		err = setVal(dstVal, reflect.ValueOf(&val))
	case uint32:
		err = setVal(dstVal, reflect.ValueOf(&val))
	case uint64:
		err = setVal(dstVal, reflect.ValueOf(&val))
	case uint:
		err = setVal(dstVal, reflect.ValueOf(&val))
	case int8:
		err = setVal(dstVal, reflect.ValueOf(&val))
	case int16:
		err = setVal(dstVal, reflect.ValueOf(&val))
	case int32:
		err = setVal(dstVal, reflect.ValueOf(&val))
	case int64:
		err = setVal(dstVal, reflect.ValueOf(&val))
	case [1]byte:
		err = setVal(dstVal, reflect.ValueOf(&val))
	case [2]byte:
		err = setVal(dstVal, reflect.ValueOf(&val))
	case [3]byte:
		err = setVal(dstVal, reflect.ValueOf(&val))
	case [4]byte:
		err = setVal(dstVal, reflect.ValueOf(&val))
	case [5]byte:
		err = setVal(dstVal, reflect.ValueOf(&val))
	case [6]byte:
		err = setVal(dstVal, reflect.ValueOf(&val))
	case [7]byte:
		err = setVal(dstVal, reflect.ValueOf(&val))
	case [8]byte:
		err = setVal(dstVal, reflect.ValueOf(&val))
	case [9]byte:
		err = setVal(dstVal, reflect.ValueOf(&val))
	case [10]byte:
		err = setVal(dstVal, reflect.ValueOf(&val))
	case [11]byte:
		err = setVal(dstVal, reflect.ValueOf(&val))
	case [12]byte:
		err = setVal(dstVal, reflect.ValueOf(&val))
	case [13]byte:
		err = setVal(dstVal, reflect.ValueOf(&val))
	case [14]byte:
		err = setVal(dstVal, reflect.ValueOf(&val))
	case [15]byte:
		err = setVal(dstVal, reflect.ValueOf(&val))
	case [16]byte:
		err = setVal(dstVal, reflect.ValueOf(&val))
	case [17]byte:
		err = setVal(dstVal, reflect.ValueOf(&val))
	case [18]byte:
		err = setVal(dstVal, reflect.ValueOf(&val))
	case [19]byte:
		err = setVal(dstVal, reflect.ValueOf(&val))
	case [20]byte:
		err = setVal(dstVal, reflect.ValueOf(&val))
	case [21]byte:
		err = setVal(dstVal, reflect.ValueOf(&val))
	case [22]byte:
		err = setVal(dstVal, reflect.ValueOf(&val))
	case [23]byte:
		err = setVal(dstVal, reflect.ValueOf(&val))
	case [24]byte:
		err = setVal(dstVal, reflect.ValueOf(&val))
	case [25]byte:
		err = setVal(dstVal, reflect.ValueOf(&val))
	case [26]byte:
		err = setVal(dstVal, reflect.ValueOf(&val))
	case [27]byte:
		err = setVal(dstVal, reflect.ValueOf(&val))
	case [28]byte:
		err = setVal(dstVal, reflect.ValueOf(&val))
	case [29]byte:
		err = setVal(dstVal, reflect.ValueOf(&val))
	case [30]byte:
		err = setVal(dstVal, reflect.ValueOf(&val))
	case [31]byte:
		err = setVal(dstVal, reflect.ValueOf(&val))
	case [32]byte:
		err = setVal(dstVal, reflect.ValueOf(&val))
	case bool:
		err = setVal(dstVal, reflect.ValueOf(&val))
	case common.Address:
		err = setVal(dstVal, reflect.ValueOf(&val))
	case common.Hash:
		err = setVal(dstVal, reflect.ValueOf(&val))
	default:
		err = setVal(dstVal, reflect.ValueOf(val))
	}
	return err
}

func copyTuple(dst, src any) error {
	dstVal := reflect.ValueOf(dst)
	if dstVal.Elem().Kind() == reflect.Ptr {
		// dst is a pointer to a pointer
		dstVal = dstVal.Elem()
	}
	if !dstVal.Elem().IsValid() {
		// dst is a pointer to nil pointer
		dstVal.Set(reflect.New(dstVal.Type().Elem()))
	}

	srcVal := reflect.ValueOf(src)
	if dstVal.Elem().Kind() != reflect.Struct || srcVal.Kind() != reflect.Struct {
		panic("no struct")
	}

	st, dt := srcVal.Type(), dstVal.Type()

	// field tag mapping (tags take precedence over names)
	fieldTagMap := make(map[string]reflect.StructField)
	for i := 0; i < srcVal.NumField(); i++ {
		field := st.Field(i)
		tag, ok := field.Tag.Lookup("abi")
		if !ok {
			panic("missing field abi tag")
		}
		fieldTagMap[tag] = field
	}

	// match dst fields to src fields tags
	for i := 0; i < dstVal.Elem().NumField(); i++ {
		dstTypeField := dt.Elem().Field(i)
		dstField := dstVal.Elem().Field(i)

		// lookup src field by:
		//  1. "abi" tag, if specified
		//  2. field name, otherwise
		// ignore if there is no match.
		var srcField reflect.Value
		if tag, ok := dstTypeField.Tag.Lookup("abi"); ok {
			srcField = srcVal.FieldByIndex(fieldTagMap[tag].Index)
		} else if field := srcVal.FieldByName(dstTypeField.Name); field != (reflect.Value{}) && !field.IsZero() {
			srcField = field
		} else {
			continue
		}

		// set dst field value to src field value
		if err := setVal(dstField, srcField); err != nil {
			return err
		}
	}
	return nil
}

func setVal(dst, src reflect.Value) error {
	st, dt := src.Type(), dst.Type()
	if !st.AssignableTo(dt) {
		if st.ConvertibleTo(dt) {
			src = src.Convert(dt)
		} else {
			return fmt.Errorf("cannot assign %v to %v", st, dt)
		}
	}

	if dst.CanAddr() {
		dst.Set(src)
	} else {
		dst.Elem().Set(src.Elem())
	}
	return nil
}
