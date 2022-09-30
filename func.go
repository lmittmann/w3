package w3

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/crypto"
	_abi "github.com/lmittmann/w3/internal/abi"
)

var (
	ErrInvalidABI       = errors.New("w3: invalid ABI")
	ErrArgumentMismatch = errors.New("w3: argument mismatch")
	ErrReturnsMismatch  = errors.New("w3: returns mismatch")
	ErrInvalidType      = errors.New("w3: invalid type")
	ErrEvmRevert        = errors.New("w3: evm reverted")

	revertSelector       = selector("Error(string)")
	outputSuccess        = B("0x0000000000000000000000000000000000000000000000000000000000000001")
	approveSelector      = selector("approve(address,uint256)")
	transferSelector     = selector("transfer(address,uint256)")
	transferFromSelector = selector("transferFrom(address,address,uint256)")
)

// Func represents a Smart Contract function ABI binding.
//
// Func implements the [w3types.Func] interface.
type Func struct {
	Signature string        // Function signature
	Selector  [4]byte       // 4-byte selector
	Args      abi.Arguments // Arguments (input)
	Returns   abi.Arguments // Returns (output)

	name string // Function name
}

// NewFunc returns a new Smart Contract function ABI binding from the given
// Solidity function signature and its returns.
//
// An error is returned if the signature or returns parsing fails.
func NewFunc(signature, returns string) (*Func, error) {
	name, args, err := _abi.Parse(signature)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidABI, err)
	}
	if name == "" {
		return nil, fmt.Errorf("%w: missing function name", ErrInvalidABI)
	}

	returnsName, returnArgs, err := _abi.Parse(returns)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidABI, err)
	}
	if returnsName != "" {
		return nil, fmt.Errorf("%w: returns must not have a function name", ErrInvalidABI)
	}

	sig := args.SignatureWithName(name)
	return &Func{
		Signature: sig,
		Selector:  selector(sig),
		Args:      abi.Arguments(args),
		Returns:   abi.Arguments(returnArgs),
		name:      name,
	}, nil
}

// MustNewFunc is like [NewFunc] but panics if the signature or returns parsing
// fails.
func MustNewFunc(signature, returns string) *Func {
	fn, err := NewFunc(signature, returns)
	if err != nil {
		panic(err)
	}
	return fn
}

// EncodeArgs ABI-encodes the given args and prepends the Func's 4-byte
// selector.
func (f *Func) EncodeArgs(args ...any) ([]byte, error) {
	return _abi.Arguments(f.Args).EncodeWithSelector(f.Selector, args...)
}

// DecodeArgs ABI-decodes the given input to the given args.
func (f *Func) DecodeArgs(input []byte, args ...any) error {
	if len(input) < 4 {
		return errors.New("w3: insufficient input length")
	}
	return _abi.Arguments(f.Args).Decode(input[4:], args...)
}

// DecodeReturns ABI-decodes the given output to the given returns.
func (f *Func) DecodeReturns(output []byte, returns ...any) error {
	// check the output for a revert reason
	if bytes.HasPrefix(output, revertSelector[:]) {
		if reason, err := abi.UnpackRevert(output); err != nil {
			return err
		} else {
			return fmt.Errorf("%w: %s", ErrEvmRevert, reason)
		}
	}

	// Gracefully handle uncompliant ERC20 returns
	if len(returns) == 1 && len(output) == 0 &&
		(f.Selector == approveSelector ||
			f.Selector == transferSelector ||
			f.Selector == transferFromSelector) {
		output = outputSuccess
	}

	return _abi.Arguments(f.Returns).Decode(output, returns...)
}

// selector returns the 4-byte selector of the given signature.
func selector(signature string) (selector [4]byte) {
	copy(selector[:], crypto.Keccak256([]byte(signature)))
	return
}
