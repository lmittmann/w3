package module

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/lmittmann/w3/w3types"
)

var null = []byte("null")

type Option[T any] func(*Factory[T])

type ArgsWrapperFunc func([]any) ([]any, error)

type RetWrapperFunc[T any] func(*T) any

type Factory[T any] struct {
	method string
	args   []any
	ret    *T

	argsWrapper ArgsWrapperFunc
	retWrapper  RetWrapperFunc[T]
}

func NewFactory[T any](method string, args []any, opts ...Option[T]) *Factory[T] {
	f := &Factory[T]{
		method: method,
		args:   args,

		argsWrapper: func(args []any) ([]any, error) { return args, nil },
		retWrapper:  func(ret *T) any { return ret },
	}
	for _, opt := range opts {
		opt(f)
	}
	return f
}

func (f Factory[T]) Returns(ret *T) w3types.RPCCaller {
	f.ret = ret
	return f
}

func (f Factory[T]) CreateRequest() (rpc.BatchElem, error) {
	args, err := f.argsWrapper(f.args)
	if err != nil {
		return rpc.BatchElem{}, err
	}
	if f.ret == nil {
		return rpc.BatchElem{}, fmt.Errorf("w3: cannot return Go value of type %T: value must be passed as a non-nil pointer reference", f.ret)
	}

	return rpc.BatchElem{
		Method: f.method,
		Args:   args,
		Result: &json.RawMessage{},
	}, nil
}

func (f Factory[T]) HandleResponse(elem rpc.BatchElem) error {
	if err := elem.Error; err != nil {
		return err
	}

	ret := *(elem.Result.(*json.RawMessage))
	if len(ret) == 0 || bytes.Equal(ret, null) {
		return errNotFound
	}

	if err := json.Unmarshal(ret, f.retWrapper(f.ret)); err != nil {
		return err
	}
	return nil
}

func WithArgsWrapper[T any](fn ArgsWrapperFunc) Option[T] {
	return func(f *Factory[T]) {
		f.argsWrapper = fn
	}
}

func WithRetWrapper[T any](fn RetWrapperFunc[T]) Option[T] {
	return func(f *Factory[T]) {
		f.retWrapper = fn
	}
}

func HexBigRetWrapper(ret **big.Int) any {
	*ret = new(big.Int)
	return (*hexutil.Big)(*ret)
}
func HexUintRetWrapper(ret *uint) any     { return (*hexutil.Uint)(ret) }
func HexUint64RetWrapper(ret *uint64) any { return (*hexutil.Uint64)(ret) }
func HexBytesRetWrapper(ret *[]byte) any  { return (*hexutil.Bytes)(ret) }
