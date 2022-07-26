package internal

import (
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/lmittmann/w3/core"
)

var errNotFound = errors.New("not found")

type Option[T any] func(*Factory[T])

type Factory[T any] struct {
	method string
	args   []any
	ret    *T

	argsWrapper func([]any) ([]any, error)
	retWrapper  func(*T) any
}

func NewFactory[T any](method string, args []any, opts ...Option[T]) *Factory[T] {
	f := &Factory[T]{
		method: method,
		args:   args,

		argsWrapper: func(args []any) ([]any, error) { return args, nil },
		retWrapper:  func(ret *T) any { return &ret },
	}
	for _, opt := range opts {
		opt(f)
	}
	return f
}

func (f Factory[T]) Returns(ret *T) core.Caller {
	f.ret = ret
	return f
}

func (f Factory[T]) CreateRequest() (rpc.BatchElem, error) {
	args, err := f.argsWrapper(f.args)
	if err != nil {
		return rpc.BatchElem{}, err
	}
	return rpc.BatchElem{
		Method: f.method,
		Args:   args,
		Result: f.retWrapper(f.ret),
	}, nil
}

func (f Factory[T]) HandleResponse(elem rpc.BatchElem) error {
	if err := elem.Error; err != nil {
		return err
	}
	if elem.Result == nil {
		return errNotFound
	}
	return nil
}

func WithArgsWrapper[T any](fn func([]any) ([]any, error)) Option[T] {
	return func(f *Factory[T]) {
		f.argsWrapper = fn
	}
}

func WithRetWrapper[T any](fn func(ret *T) any) Option[T] {
	return func(f *Factory[T]) {
		f.retWrapper = fn
	}
}

var (
	HexBigWrapper    = func(ret *big.Int) any { return (*hexutil.Big)(ret) }
	HexUint64Wrapper = func(ret *uint64) any { return (*hexutil.Uint64)(ret) }
	HexBytesWrapper  = func(ret *[]byte) any { return (*hexutil.Bytes)(ret) }
)
