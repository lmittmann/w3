package w3

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/ethereum/go-ethereum/rpc"
	"github.com/lmittmann/w3/w3types"
	"golang.org/x/time/rate"
)

// Client represents a connection to an RPC endpoint.
type Client struct {
	client *rpc.Client

	// rate limiter
	rl        *rate.Limiter
	rlPerCall bool
}

// NewClient returns a new Client given an rpc.Client client.
func NewClient(client *rpc.Client, opts ...Option) *Client {
	c := &Client{
		client: client,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Dial returns a new Client connected to the URL rawurl. An error is returned
// if the connection establishment fails.
//
// The supported URL schemes are "http", "https", "ws" and "wss". If rawurl is a
// file name with no URL scheme, a local IPC socket connection is established.
func Dial(rawurl string, opts ...Option) (*Client, error) {
	client, err := rpc.Dial(rawurl)
	if err != nil {
		return nil, err
	}
	return NewClient(client, opts...), nil
}

// MustDial is like [Dial] but panics if the connection establishment fails.
func MustDial(rawurl string, opts ...Option) *Client {
	client, err := Dial(rawurl, opts...)
	if err != nil {
		panic(fmt.Sprintf("w3: %s", err))
	}
	return client
}

// Close closes the RPC connection and cancels any in-flight requests.
//
// Close implements the [io.Closer] interface.
func (c *Client) Close() error {
	c.client.Close()
	return nil
}

// CallCtx creates the final RPC request, sends it, and handles the RPC
// response.
//
// An error is returned if RPC request creation, networking, or RPC response
// handling fails.
func (c *Client) CallCtx(ctx context.Context, calls ...w3types.Caller) error {
	// no requests = nothing to do
	if len(calls) <= 0 {
		return nil
	}

	// invoke rate limiter
	if err := c.rateLimit(ctx, len(calls)); err != nil {
		return err
	}

	// create requests
	batchElems := make([]rpc.BatchElem, len(calls))
	var err error
	for i, req := range calls {
		batchElems[i], err = req.CreateRequest()
		if err != nil {
			return err
		}
	}

	// do requests
	if len(batchElems) > 1 {
		// batch requests if >1 request
		err = c.client.BatchCallContext(ctx, batchElems)
		if err != nil {
			return err
		}
	} else {
		// non-batch requests if 1 request
		batchElem := batchElems[0]
		err = c.client.CallContext(ctx, batchElem.Result, batchElem.Method, batchElem.Args...)
		if err != nil {
			switch reflect.TypeOf(err).String() {
			case "*rpc.jsonError":
				batchElems[0].Error = err
			default:
				return err
			}
		}
	}

	// handle responses
	var callErrs CallErrors
	for i, req := range calls {
		err = req.HandleResponse(batchElems[i])
		if err != nil {
			if callErrs == nil {
				callErrs = make(CallErrors, len(calls))
			}
			callErrs[i] = err
		}
	}
	if len(callErrs) > 0 {
		return callErrs
	}
	return nil
}

// Call is like [Client.CallCtx] with ctx equal to context.Background().
func (c *Client) Call(calls ...w3types.Caller) error {
	return c.CallCtx(context.Background(), calls...)
}

func (c *Client) rateLimit(ctx context.Context, n int) error {
	if c.rl == nil {
		return nil
	}

	if c.rlPerCall {
		return c.rl.WaitN(ctx, n)
	}
	return c.rl.Wait(ctx)
}

// CallErrors is an error type that contains the errors of multiple calls. The
// length of the error slice is equal to the number of calls. Each error at a
// given index corresponds to the call at the same index. An error is nil if the
// corresponding call was successful.
type CallErrors []error

func (e CallErrors) Error() string {
	if len(e) == 1 && e[0] != nil {
		return fmt.Sprintf("w3: call failed: %s", e[0])
	}

	var errors []string
	for i, err := range e {
		if err == nil {
			continue
		}
		errors = append(errors, fmt.Sprintf("call[%d]: %s", i, err))
	}

	var plr string
	if len(errors) > 1 {
		plr = "s"
	}
	return fmt.Sprintf("w3: %d call%s failed:\n%s", len(errors), plr, strings.Join(errors, "\n"))
}

func (e CallErrors) Is(target error) bool {
	_, ok := target.(CallErrors)
	return ok
}
