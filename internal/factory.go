package internal

import (
	"errors"

	"github.com/ethereum/go-ethereum/rpc"
	"github.com/lmittmann/w3/core"
)

var errNotFound = errors.New("not found")

type Factory[T any] struct {
	method string
	args   []any
	ret    *T
}

func NewFactory[T any](method string, args []any) *Factory[T] {
	f := &Factory[T]{
		method: method,
		args:   args,
	}
	return f
}

func (f Factory[T]) Returns(ret *T) core.Caller {
	f.ret = ret
	return f
}

func (f Factory[T]) CreateRequest() (rpc.BatchElem, error) {
	return rpc.BatchElem{
		Method: f.method,
		Args:   f.args,
		Result: &f.ret,
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
