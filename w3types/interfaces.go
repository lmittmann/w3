/*
Package w3types implements common types.
*/
package w3types

import "github.com/ethereum/go-ethereum/rpc"

// Func is the interface that wraps the methods for ABI encoding and decoding.
type Func interface {

	// EncodeArgs ABI-encodes the given args and prepends the Func's 4-byte
	// selector.
	EncodeArgs(args ...any) (input []byte, err error)

	// DecodeArgs ABI-decodes the given input to the given args.
	DecodeArgs(input []byte, args ...any) (err error)

	// DecodeReturns ABI-decodes the given output to the given returns.
	DecodeReturns(output []byte, returns ...any) (err error)
}

// RequestCreator is the interface that wraps the basic CreateRequest method.
type RequestCreator interface {
	CreateRequest() (elem rpc.BatchElem, err error)
}

// ResponseHandler is the interface that wraps the basic HandleResponse method.
type ResponseHandler interface {
	HandleResponse(elem rpc.BatchElem) (err error)
}

// Caller is the interface that groups the basic CreateRequest and
// HandleResponse methods.
type Caller interface {
	RequestCreator
	ResponseHandler
}

// CallerFactory is the interface that wraps the basic Returns method.
type CallerFactory[T any] interface {

	// Returns given argument points to the variable in which to store the
	// calls result.
	Returns(*T) Caller
}
