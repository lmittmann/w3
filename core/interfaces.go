package core

import "github.com/ethereum/go-ethereum/rpc"

// RequestCreater is the interface that wraps the basic CreateRequest method.
type RequestCreater interface {
	CreateRequest() (elem rpc.BatchElem, err error)
}

// ResponseHandler is the interface that wraps the basic HandleResponse method.
type ResponseHandler interface {
	HandleResponse(elem rpc.BatchElem) (err error)
}

// RequestCreaterResponseHandler is the interface that groups the basic CreateRequest and
// HandleResponse methods.
type RequestCreaterResponseHandler interface {
	RequestCreater
	ResponseHandler
}
