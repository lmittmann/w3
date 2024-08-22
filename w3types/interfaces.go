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

// RPCCaller is the interface that groups the basic CreateRequest and
// HandleResponse methods.
type RPCCaller interface {
	// Create a new rpc.BatchElem for doing the RPC call.
	CreateRequest() (elem rpc.BatchElem, err error)

	// Handle the response from the rpc.BatchElem to handle its result.
	HandleResponse(elem rpc.BatchElem) (err error)
}

// RPCCallerFactory is the interface that wraps the basic Returns method.
type RPCCallerFactory[T any] interface {
	// Returns given argument points to the variable in which to store the
	// calls result.
	Returns(*T) RPCCaller
}

// RPCSubscriber is the interface that wraps the basic CreateRequest method.
type RPCSubscriber interface {
	// CreateRequest returns the namespace, channel, params for starting a new
	// subscription and an error if the request cannot be created.
	CreateRequest() (namespace string, ch any, params []any, err error)
}
