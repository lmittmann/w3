package core

import "github.com/ethereum/go-ethereum/rpc"

// Func is the interface that wraps the methods for ABI encoding and decoding.
type Func interface {

	// EncodeArgs ABI-encodes the given args and prepends the Func's 4-byte
	// selector.
	EncodeArgs(args ...interface{}) (input []byte, err error)

	// DecodeArgs ABI-decodes the given input to the given args.
	DecodeArgs(input []byte, args ...interface{}) (err error)

	// DecodeReturns ABI-decodes the given output to the given returns.
	DecodeReturns(output []byte, returns ...interface{}) (err error)
}

// RequestCreator is the interface that wraps the basic CreateRequest method.
type RequestCreator interface {
	CreateRequest() (elem rpc.BatchElem, err error)
}

// ResponseHandler is the interface that wraps the basic HandleResponse method.
type ResponseHandler interface {
	HandleResponse(elem rpc.BatchElem) (err error)
}

// RequestCreatorResponseHandler is the interface that groups the basic CreateRequest and
// HandleResponse methods.
type RequestCreatorResponseHandler interface {
	RequestCreator
	ResponseHandler
}
