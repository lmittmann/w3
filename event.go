package w3

import (
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	_abi "github.com/lmittmann/w3/internal/abi"
)

// Event represents a Smart Contract event decoder.
type Event struct {
	Signature string        // Event signature
	Topic0    common.Hash   // Hash of event signature (Topic 0)
	Args      abi.Arguments // Arguments

	indexedArgs abi.Arguments // Subset of Args that are indexed
}

// NewEvent returns a new Smart Contract event log decoder from the given
// Solidity event signature.
//
// An error is returned if the signature parsing fails.
func NewEvent(signature string) (*Event, error) {
	name, args, err := _abi.ParseWithName(signature)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidABI, err)
	}
	if name == "" {
		return nil, fmt.Errorf("%w: missing event name", ErrInvalidABI)
	}

	indexedArgs := make(abi.Arguments, 0)
	for _, arg := range args {
		if arg.Indexed {
			arg.Indexed = false
			indexedArgs = append(indexedArgs, arg)
		}
	}

	sig := args.SignatureWithName(name)
	return &Event{
		Signature:   sig,
		Topic0:      crypto.Keccak256Hash([]byte(sig)),
		Args:        abi.Arguments(args),
		indexedArgs: indexedArgs,
	}, nil
}

// MustNewEvent is like [NewEvent] but panics if the signature parsing fails.
func MustNewEvent(signature string) *Event {
	event, err := NewEvent(signature)
	if err != nil {
		panic(err)
	}
	return event
}

// DecodeArgs decodes the topics and data of the given log to the given args.
//
// DecodeArgs is insensitiv to indexed fields.
func (e *Event) DecodeArgs(log *types.Log, args ...any) error {
	if len(log.Topics) <= 0 || e.Topic0 != log.Topics[0] {
		return fmt.Errorf("w3: topic0 mismatch")
	}

	if len(e.Args) != len(args) {
		return fmt.Errorf("%w: expected %d arguments, got %d", ErrArgumentMismatch, len(e.Args), len(args))
	}
	if len(e.indexedArgs) != len(log.Topics)-1 {
		return fmt.Errorf("%w: expected %d indexed arguments, got %d", ErrArgumentMismatch, len(e.indexedArgs), len(log.Topics)-1)
	}

	indexedArgs := make([]any, 0, len(e.indexedArgs))
	nonIndexedArgs := make([]any, 0, len(e.Args)-len(e.indexedArgs))
	for i, arg := range e.Args {
		if arg.Indexed {
			indexedArgs = append(indexedArgs, args[i])
		} else {
			nonIndexedArgs = append(nonIndexedArgs, args[i])
		}
	}

	// decode indexed args
	for i, arg := range indexedArgs {
		if err := (_abi.Arguments{e.indexedArgs[i]}).Decode(log.Topics[i+1][:], arg); err != nil {
			return err
		}
	}

	// decode non-indexed args
	if err := _abi.Arguments(e.Args).Decode(log.Data, nonIndexedArgs...); err != nil {
		return err
	}
	return nil
}
