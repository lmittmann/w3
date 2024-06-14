package abi

import (
	"errors"
	"fmt"
	"math/big"
	"reflect"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

var (
	errUnassignable = errors.New("unassignable")

	// src non slice/array/struct types
	srcBasicTypes = map[reflect.Type]struct{}{
		reflect.TypeOf(true):             {},
		reflect.TypeOf(uint(0)):          {},
		reflect.TypeOf(uint8(0)):         {},
		reflect.TypeOf(uint16(0)):        {},
		reflect.TypeOf(uint32(0)):        {},
		reflect.TypeOf(uint64(0)):        {},
		reflect.TypeOf(int(0)):           {},
		reflect.TypeOf(int8(0)):          {},
		reflect.TypeOf(int16(0)):         {},
		reflect.TypeOf(int32(0)):         {},
		reflect.TypeOf(int64(0)):         {},
		reflect.TypeOf([1]byte{}):        {},
		reflect.TypeOf([2]byte{}):        {},
		reflect.TypeOf([3]byte{}):        {},
		reflect.TypeOf([4]byte{}):        {},
		reflect.TypeOf([5]byte{}):        {},
		reflect.TypeOf([6]byte{}):        {},
		reflect.TypeOf([7]byte{}):        {},
		reflect.TypeOf([8]byte{}):        {},
		reflect.TypeOf([9]byte{}):        {},
		reflect.TypeOf([10]byte{}):       {},
		reflect.TypeOf([11]byte{}):       {},
		reflect.TypeOf([12]byte{}):       {},
		reflect.TypeOf([13]byte{}):       {},
		reflect.TypeOf([14]byte{}):       {},
		reflect.TypeOf([15]byte{}):       {},
		reflect.TypeOf([16]byte{}):       {},
		reflect.TypeOf([17]byte{}):       {},
		reflect.TypeOf([18]byte{}):       {},
		reflect.TypeOf([19]byte{}):       {},
		reflect.TypeOf([20]byte{}):       {},
		reflect.TypeOf([21]byte{}):       {},
		reflect.TypeOf([22]byte{}):       {},
		reflect.TypeOf([23]byte{}):       {},
		reflect.TypeOf([24]byte{}):       {},
		reflect.TypeOf([25]byte{}):       {},
		reflect.TypeOf([26]byte{}):       {},
		reflect.TypeOf([27]byte{}):       {},
		reflect.TypeOf([28]byte{}):       {},
		reflect.TypeOf([29]byte{}):       {},
		reflect.TypeOf([30]byte{}):       {},
		reflect.TypeOf([31]byte{}):       {},
		reflect.TypeOf([32]byte{}):       {},
		reflect.TypeOf(common.Address{}): {},
		reflect.TypeOf(common.Hash{}):    {},
		reflect.TypeOf(""):               {},
		reflect.TypeOf([]byte{}):         {},
		reflect.TypeOf(&big.Int{}):       {},
		reflect.TypeOf(big.Int{}):        {},
	}
)

// Copy shallow copies the value src to dst. If src is an anonymous struct or an
// array/slice of anonymous structs, the fields of the anonymous struct are
// copied to dst.
func Copy(dst, src any) error {
	// check if dst is valid
	if dst == nil {
		return fmt.Errorf("abi: decode nil")
	}

	rDst := reflect.ValueOf(dst)
	if rDst.Kind() != reflect.Pointer {
		return fmt.Errorf("abi: decode non-pointer %T", dst)
	}
	if rDst.IsNil() {
		return fmt.Errorf("abi: decode nil %T", dst)
	}

	err := rCopy(
		dereference(rDst),
		reflect.ValueOf(src),
	)
	if errors.Is(err, errUnassignable) {
		return fmt.Errorf("abi: can't assign %T to %T", src, dst)
	} else if err != nil {
		return fmt.Errorf("abi: %w", err)
	}

	return nil
}

func rCopy(dst, src reflect.Value) error {
	if _, ok := srcBasicTypes[src.Type()]; ok {
		return set(dst, reference(src))
	} else if k := src.Kind(); k == reflect.Struct {
		return setStruct(dst, src)
	} else if k == reflect.Slice {
		return setSlice(dst, src)
	} else if k == reflect.Array {
		return setArray(dst, src)
	} else {
		return fmt.Errorf("unsupported src type %T", src.Interface())
	}
}

func set(dst, src reflect.Value) error {
	if src.Kind() != reflect.Ptr && dst.Kind() == reflect.Ptr {
		src = reference(src)
	} else if src.Kind() == reflect.Pointer && dst.Kind() != reflect.Pointer {
		src = src.Elem()
	}

	st, dt := src.Type(), dst.Type()
	if !st.AssignableTo(dt) {
		if st.ConvertibleTo(dt) {
			src = src.Convert(dt)
		} else {
			return errUnassignable
		}
	}

	if dst.CanAddr() {
		dst.Set(src)
	} else {
		dst.Elem().Set(src.Elem())
	}
	return nil
}

func setStruct(dst, src reflect.Value) error {
	if dst.Kind() == reflect.Pointer {
		if dst.IsNil() {
			dst.Set(reflect.New(dst.Type().Elem()))
		}
		dst = dst.Elem()
	}

	st, dt := src.Type(), dst.Type()

	// field tag mapping (tags take precedence over names)
	srcFields := make(map[string]reflect.StructField)
	for i := range src.NumField() {
		field := st.Field(i)
		srcFields[field.Name] = field
	}

	for i := range dst.NumField() {
		dstField := dt.Field(i)
		srcField, ok := srcFields[dstField.Name]
		if !ok {
			if tag, ok := dstField.Tag.Lookup("abi"); ok {
				name := abi.ToCamelCase(tag)
				if srcField, ok = srcFields[name]; !ok {
					continue
				}
			} else {
				continue
			}
		}

		rCopy(
			dst.FieldByName(dstField.Name),
			src.FieldByName(srcField.Name),
		)
	}
	return nil
}

func setSlice(dst, src reflect.Value) error {
	if dst.IsNil() && dst.Kind() == reflect.Pointer {
		dst = reflect.New(dst.Type().Elem())
	}
	if dst.Kind() == reflect.Pointer {
		dst.Elem().Set(reflect.MakeSlice(dst.Elem().Type(), src.Len(), src.Len()))
	} else {
		dst.Set(reflect.MakeSlice(dst.Type(), src.Len(), src.Len()))
	}

	for i := range src.Len() {
		if dst.Kind() == reflect.Pointer {
			rCopy(dst.Elem().Index(i), src.Index(i))
		} else {
			rCopy(dst.Index(i), src.Index(i))
		}
	}
	return nil
}

func setArray(dst, src reflect.Value) error {
	if dst.Kind() == reflect.Pointer && dst.IsNil() {
		dst = reflect.New(dst.Type().Elem())
	}

	for i := range src.Len() {
		if dst.Kind() == reflect.Pointer {
			rCopy(dst.Elem().Index(i), src.Index(i))
		} else {
			rCopy(dst.Index(i), src.Index(i))
		}
	}
	return nil
}

func dereference(v reflect.Value) reflect.Value {
	for v.Kind() == reflect.Pointer && v.Elem().Kind() == reflect.Pointer {
		v = v.Elem()
	}
	return v
}

func reference(v reflect.Value) reflect.Value {
	if v.Kind() != reflect.Pointer {
		if v.CanAddr() {
			v = v.Addr()
		} else {
			switch vv := v.Interface().(type) {
			case bool:
				v = reflect.ValueOf(&vv)
			case uint:
				v = reflect.ValueOf(&vv)
			case uint8:
				v = reflect.ValueOf(&vv)
			case uint16:
				v = reflect.ValueOf(&vv)
			case uint32:
				v = reflect.ValueOf(&vv)
			case uint64:
				v = reflect.ValueOf(&vv)
			case int:
				v = reflect.ValueOf(&vv)
			case int8:
				v = reflect.ValueOf(&vv)
			case int16:
				v = reflect.ValueOf(&vv)
			case int32:
				v = reflect.ValueOf(&vv)
			case int64:
				v = reflect.ValueOf(&vv)
			case [1]byte:
				v = reflect.ValueOf(&vv)
			case [2]byte:
				v = reflect.ValueOf(&vv)
			case [3]byte:
				v = reflect.ValueOf(&vv)
			case [4]byte:
				v = reflect.ValueOf(&vv)
			case [5]byte:
				v = reflect.ValueOf(&vv)
			case [6]byte:
				v = reflect.ValueOf(&vv)
			case [7]byte:
				v = reflect.ValueOf(&vv)
			case [8]byte:
				v = reflect.ValueOf(&vv)
			case [9]byte:
				v = reflect.ValueOf(&vv)
			case [10]byte:
				v = reflect.ValueOf(&vv)
			case [11]byte:
				v = reflect.ValueOf(&vv)
			case [12]byte:
				v = reflect.ValueOf(&vv)
			case [13]byte:
				v = reflect.ValueOf(&vv)
			case [14]byte:
				v = reflect.ValueOf(&vv)
			case [15]byte:
				v = reflect.ValueOf(&vv)
			case [16]byte:
				v = reflect.ValueOf(&vv)
			case [17]byte:
				v = reflect.ValueOf(&vv)
			case [18]byte:
				v = reflect.ValueOf(&vv)
			case [19]byte:
				v = reflect.ValueOf(&vv)
			case [20]byte:
				v = reflect.ValueOf(&vv)
			case [21]byte:
				v = reflect.ValueOf(&vv)
			case [22]byte:
				v = reflect.ValueOf(&vv)
			case [23]byte:
				v = reflect.ValueOf(&vv)
			case [24]byte:
				v = reflect.ValueOf(&vv)
			case [25]byte:
				v = reflect.ValueOf(&vv)
			case [26]byte:
				v = reflect.ValueOf(&vv)
			case [27]byte:
				v = reflect.ValueOf(&vv)
			case [28]byte:
				v = reflect.ValueOf(&vv)
			case [29]byte:
				v = reflect.ValueOf(&vv)
			case [30]byte:
				v = reflect.ValueOf(&vv)
			case [31]byte:
				v = reflect.ValueOf(&vv)
			case [32]byte:
				v = reflect.ValueOf(&vv)
			case common.Address:
				v = reflect.ValueOf(&vv)
			case common.Hash:
				v = reflect.ValueOf(&vv)
			case string:
				v = reflect.ValueOf(&vv)
			case []byte:
				v = reflect.ValueOf(&vv)
			}
		}
	}
	return v
}
