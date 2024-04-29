package w3vm

import (
	"fmt"
	"math/big"
	"sync"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/holiman/uint256"
	"github.com/lmittmann/w3"
	"github.com/lmittmann/w3/internal/crypto"
	"github.com/lmittmann/w3/module/eth"
	"github.com/lmittmann/w3/w3types"
)

// Fetcher is the interface to access account state of a blockchain.
type Fetcher interface {
	// Account fetches the account of the given address.
	Account(common.Address) (*types.StateAccount, error)

	// Code fetches the code of the given code hash.
	Code(common.Hash) ([]byte, error)

	// StorageAt fetches the state of the given address and storage slot.
	StorageAt(common.Address, common.Hash) (common.Hash, error)

	// HeaderHash fetches the hash of the header with the given number.
	HeaderHash(uint64) (common.Hash, error)
}

type rpcFetcher struct {
	client      *w3.Client
	blockNumber *big.Int

	mux          sync.RWMutex
	accounts     map[common.Address]func() (*types.StateAccount, error)
	contracts    map[common.Hash]func() ([]byte, error)
	mux2         sync.RWMutex
	storage      map[storageKey]func() (common.Hash, error)
	mux3         sync.RWMutex
	headerHashes map[uint64]func() (common.Hash, error)
}

// NewRPCFetcher returns a new [Fetcher] that fetches account state from the given
// RPC client for the given block number.
//
// Note, that the returned state for a given block number is the state after the
// execution of that block.
func NewRPCFetcher(client *w3.Client, blockNumber *big.Int) Fetcher {
	return newRPCFetcher(client, blockNumber)
}

func newRPCFetcher(client *w3.Client, blockNumber *big.Int) *rpcFetcher {
	return &rpcFetcher{
		client:       client,
		blockNumber:  blockNumber,
		accounts:     make(map[common.Address]func() (*types.StateAccount, error)),
		contracts:    make(map[common.Hash]func() ([]byte, error)),
		storage:      make(map[storageKey]func() (common.Hash, error)),
		headerHashes: make(map[uint64]func() (common.Hash, error)),
	}
}

func (f *rpcFetcher) Account(addr common.Address) (*types.StateAccount, error) {
	f.mux.RLock()
	acc, ok := f.accounts[addr]
	f.mux.RUnlock()
	if ok {
		return acc()
	}

	var (
		accNew      = &types.StateAccount{Balance: new(uint256.Int)}
		contractNew []byte

		accCh      = make(chan func() (*types.StateAccount, error), 1)
		contractCh = make(chan func() ([]byte, error), 1)
	)
	go func() {
		err := f.call(
			eth.Nonce(addr, f.blockNumber).Returns(&accNew.Nonce),
			ethBalance(addr, f.blockNumber).Returns(accNew.Balance),
			eth.Code(addr, f.blockNumber).Returns(&contractNew),
		)
		if err != nil {
			accCh <- func() (*types.StateAccount, error) { return nil, err }
			contractCh <- func() ([]byte, error) { return nil, err }
			return
		}

		if len(contractNew) == 0 {
			accNew.CodeHash = types.EmptyCodeHash[:]
		} else {
			accNew.CodeHash = crypto.Keccak256(contractNew)
		}
		accCh <- func() (*types.StateAccount, error) { return accNew, nil }
		contractCh <- func() ([]byte, error) { return contractNew, nil }
	}()

	f.mux.Lock()
	defer f.mux.Unlock()
	accOnce := sync.OnceValues(<-accCh)
	f.accounts[addr] = accOnce
	accRet, err := accOnce()
	if err != nil {
		return nil, err
	}
	f.contracts[common.BytesToHash(accRet.CodeHash)] = sync.OnceValues(<-contractCh)
	return accRet, nil
}

func (f *rpcFetcher) Code(codeHash common.Hash) ([]byte, error) {
	fmt.Println("code")
	f.mux.RLock()
	contract, ok := f.contracts[codeHash]
	f.mux.RUnlock()
	if !ok {
		panic("not implemented")
	}
	return contract()
}

func (f *rpcFetcher) StorageAt(addr common.Address, slot common.Hash) (common.Hash, error) {
	fmt.Println("storage", addr, slot)
	key := storageKey{addr, slot}

	f.mux2.RLock()
	storage, ok := f.storage[key]
	f.mux2.RUnlock()
	if ok {
		return storage()
	}

	var (
		storageVal   common.Hash
		storageValCh = make(chan func() (common.Hash, error), 1)
	)
	go func() {
		err := f.call(eth.StorageAt(addr, slot, f.blockNumber).Returns(&storageVal))
		storageValCh <- func() (common.Hash, error) { return storageVal, err }
	}()

	storageValOnce := sync.OnceValues(<-storageValCh)
	f.mux2.Lock()
	f.storage[key] = storageValOnce
	f.mux2.Unlock()
	return storageValOnce()
}

func (f *rpcFetcher) HeaderHash(blockNumber uint64) (common.Hash, error) {
	f.mux3.RLock()
	hash, ok := f.headerHashes[blockNumber]
	f.mux3.RUnlock()
	if ok {
		return hash()
	}

	var (
		header       header
		headerHashCh = make(chan func() (common.Hash, error), 1)
	)
	go func() {
		err := f.call(ethHeaderHash(f.blockNumber.Uint64()).Returns(&header))
		headerHashCh <- func() (common.Hash, error) { return header.Hash, err }
	}()

	headerHashOnce := sync.OnceValues(<-headerHashCh)
	f.mux3.Lock()
	f.headerHashes[blockNumber] = headerHashOnce
	f.mux3.Unlock()
	return headerHashOnce()
}

func (f *rpcFetcher) call(calls ...w3types.RPCCaller) error {
	return f.client.Call(calls...)
}

// func (f *rpcFetcher) setForkState(s *forkState) {
// 	f.mux.Lock()
// 	f.mux2.Lock()
// 	f.mux3.Lock()
// 	defer f.mux.Unlock()
// 	defer f.mux2.Unlock()
// 	defer f.mux3.Unlock()

// 	for addr, acc := range s.Accounts {
// 		f.accounts[addr] = acc
// 	}

// 	for blockNumber, hash := range s.HeaderHashes {
// 		f.headerHashes[uint64(blockNumber)] = hash
// 	}
// }

// func (f *rpcFetcher) getForkState() *forkState {
// 	f.mux.Lock()
// 	f.mux2.Lock()
// 	f.mux3.Lock()
// 	defer f.mux.Unlock()
// 	defer f.mux2.Unlock()
// 	defer f.mux3.Unlock()

// 	s := new(forkState)
// 	if len(f.accounts) > 0 {
// 		s.Accounts = make(map[common.Address]*account, len(f.accounts))
// 		for addr, acc := range f.accounts {
// 			accCop := *acc
// 			s.Accounts[addr] = &accCop
// 		}
// 	}

// 	if len(f.headerHashes) > 0 {
// 		s.HeaderHashes = make(map[hexutil.Uint64]common.Hash, len(f.headerHashes))
// 		for blockNumber, hash := range f.headerHashes {
// 			s.HeaderHashes[hexutil.Uint64(blockNumber)] = hash
// 		}
// 	}
// 	return s
// }

////////////////////////////////////////////////////////////////////////////////////////////////////
// TestingRPCFetcher ///////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////////////////

// NewTestingRPCFetcher returns a new [Fetcher] like [NewRPCFetcher], but caches
// the fetched state on disk in the testdata directory of the tests package.
func NewTestingRPCFetcher(tb testing.TB, client *w3.Client, blockNumber *big.Int) Fetcher {
	// fp := getTbFilepath(tb)
	// if ok := isTbInMod(fp); !ok {
	// 	panic("must be called from a test in a module")
	// }

	// if blockNumber == nil {
	// 	tb.Fatal("w3vm: block number must not be <nil>")
	// }

	// var chainID uint64
	// if err := client.Call(
	// 	eth.ChainID().Returns(&chainID),
	// ); err != nil {
	// 	tb.Fatalf("w3vm: failed to fetch chain ID: %v", err)
	// }

	// fp = filepath.Join(fp, "testdata", "w3vm", fmt.Sprintf("%d_%v.json", chainID, blockNumber))
	// testdataState, err := readTestdataState(fp)
	// if err != nil {
	// 	tb.Fatalf("w3vm: failed to read state from testdata: %v", err)
	// }

	fetcher := newRPCFetcher(client, blockNumber)
	// fetcher.setForkState(testdataState)

	// tb.Cleanup(func() {
	// 	postTestdataState := fetcher.getForkState()
	// 	if err := writeTestdataState(fp, postTestdataState); err != nil {
	// 		tb.Errorf("w3test: failed to write state to testdata: %v", err)
	// 	}
	// })

	return fetcher
}

type storageKey struct {
	addr common.Address
	slot common.Hash
}
