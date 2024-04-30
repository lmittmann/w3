package w3vm

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/holiman/uint256"
	"github.com/lmittmann/w3"
	"github.com/lmittmann/w3/internal/crypto"
	w3hexutil "github.com/lmittmann/w3/internal/hexutil"
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

	dirty bool // indicates whether new state has been fetched
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

func (f *rpcFetcher) Account(addr common.Address) (a *types.StateAccount, e error) {
	f.mux.RLock()
	acc, ok := f.accounts[addr]
	f.mux.RUnlock()
	if ok {
		return acc()
	}
	f.dirty = true

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
	f.mux.RLock()
	contract, ok := f.contracts[codeHash]
	f.mux.RUnlock()
	if !ok {
		panic("not implemented")
	}
	return contract()
}

func (f *rpcFetcher) StorageAt(addr common.Address, slot common.Hash) (common.Hash, error) {
	key := storageKey{addr, slot}

	f.mux2.RLock()
	storage, ok := f.storage[key]
	f.mux2.RUnlock()
	if ok {
		return storage()
	}
	f.dirty = true

	var (
		storageVal   common.Hash
		storageValCh = make(chan func() (common.Hash, error), 1)
	)
	go func() {
		err := f.call(ethStorageAt(addr, slot, f.blockNumber).Returns(&storageVal))
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
	f.dirty = true

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

////////////////////////////////////////////////////////////////////////////////////////////////////
// TestingRPCFetcher ///////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////////////////

// NewTestingRPCFetcher returns a new [Fetcher] like [NewRPCFetcher], but caches
// the fetched state on disk in the testdata directory of the tests package.
func NewTestingRPCFetcher(tb testing.TB, chainID uint64, client *w3.Client, blockNumber *big.Int) Fetcher {
	if ok := isTbInMod(getTbFilepath(tb)); !ok {
		panic("must be called from a test in a module")
	}

	fetcher := newRPCFetcher(client, blockNumber)
	if err := fetcher.loadTestdataState(tb, chainID); err != nil {
		tb.Fatalf("w3vm: failed to load state from testdata: %v", err)
	}

	tb.Cleanup(func() {
		if err := fetcher.storeTestdataState(tb, chainID); err != nil {
			tb.Fatalf("w3vm: failed to write state to testdata: %v", err)
		}
	})
	return fetcher
}

var (
	globalStateStoreMux sync.RWMutex
	globalStateStore    = make(map[string]*state)
)

func (f *rpcFetcher) loadTestdataState(tb testing.TB, chainID uint64) error {
	dir := getTbFilepath(tb)
	fn := filepath.Join(dir,
		"testdata",
		"w3vm",
		fmt.Sprintf("%d_%v.json", chainID, f.blockNumber),
	)

	var s *state

	// check if the state has already been loaded
	globalStateStoreMux.RLock()
	s, ok := globalStateStore[fn]
	globalStateStoreMux.RUnlock()

	if !ok {
		// load state from file
		file, err := os.Open(fn)
		if errors.Is(err, os.ErrNotExist) {
			return nil
		} else if err != nil {
			return err
		}
		defer file.Close()

		if err := json.NewDecoder(file).Decode(&s); err != nil {
			return err
		}
	}

	f.mux.Lock()
	f.mux2.Lock()
	f.mux3.Lock()
	defer f.mux.Unlock()
	defer f.mux2.Unlock()
	defer f.mux3.Unlock()

	for addr, acc := range s.Accounts {
		var codeHash common.Hash
		if len(acc.Code) > 0 {
			codeHash = crypto.Keccak256Hash(acc.Code)
		} else {
			codeHash = types.EmptyCodeHash
		}

		f.accounts[addr] = func() (*types.StateAccount, error) {
			return &types.StateAccount{
				Nonce:    uint64(acc.Nonce),
				Balance:  (*uint256.Int)(acc.Balance),
				CodeHash: codeHash[:],
			}, nil
		}
		if _, ok := f.contracts[codeHash]; codeHash != types.EmptyCodeHash && !ok {
			f.contracts[codeHash] = func() ([]byte, error) {
				return acc.Code, nil
			}
		}
		for slot, val := range acc.Storage {
			f.storage[storageKey{addr, (common.Hash)(slot)}] = func() (common.Hash, error) {
				return (common.Hash)(val), nil
			}
		}
		for blockNumber, hash := range s.HeaderHashes {
			f.headerHashes[uint64(blockNumber)] = func() (common.Hash, error) {
				return hash, nil
			}
		}
	}
	return nil
}

func (f *rpcFetcher) storeTestdataState(tb testing.TB, chainID uint64) error {
	if !f.dirty {
		return nil // the state has not been modified
	}

	dir := getTbFilepath(tb)
	fn := filepath.Join(dir,
		"testdata",
		"w3vm",
		fmt.Sprintf("%d_%v.json", chainID, f.blockNumber),
	)

	// build state
	f.mux.RLock()
	f.mux2.RLock()
	f.mux3.RLock()
	defer f.mux.RUnlock()
	defer f.mux2.RUnlock()
	defer f.mux3.RUnlock()

	s := &state{
		Accounts:     make(map[common.Address]*account, len(f.accounts)),
		HeaderHashes: make(map[hexutil.Uint64]common.Hash, len(f.headerHashes)),
	}

	for addr, acc := range f.accounts {
		acc, err := acc()
		if err != nil {
			continue
		}

		s.Accounts[addr] = &account{
			Nonce:   hexutil.Uint64(acc.Nonce),
			Balance: (*hexutil.U256)(acc.Balance),
		}
		if !bytes.Equal(acc.CodeHash, types.EmptyCodeHash[:]) {
			s.Accounts[addr].Code, _ = f.contracts[common.BytesToHash(acc.CodeHash)]()
		}
	}

	for storageKey, storageVal := range f.storage {
		storageVal, err := storageVal()
		if err != nil {
			continue
		}

		if s.Accounts[storageKey.addr].Storage == nil {
			s.Accounts[storageKey.addr].Storage = make(map[w3hexutil.Hash]w3hexutil.Hash)
		}
		s.Accounts[storageKey.addr].Storage[w3hexutil.Hash(storageKey.slot)] = w3hexutil.Hash(storageVal)
	}

	globalStateStoreMux.Lock()
	defer globalStateStoreMux.Unlock()
	// merge state
	dstState, ok := globalStateStore[fn]
	if ok {
		if modified := mergeStates(dstState, s); !modified {
			return nil
		}
	} else {
		dstState = s
		globalStateStore[fn] = s
	}

	// create directory, if it does not exist
	dirPath := filepath.Dir(fn)
	if _, err := os.Stat(dirPath); errors.Is(err, os.ErrNotExist) {
		if err := os.MkdirAll(dirPath, 0775); err != nil {
			return err
		}
	}

	file, err := os.OpenFile(fn, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0664)
	if err != nil {
		return err
	}
	defer file.Close()

	dec := json.NewEncoder(file)
	dec.SetIndent("", "\t")
	if err := dec.Encode(dstState); err != nil {
		return err
	}
	return nil
}

type state struct {
	Accounts     map[common.Address]*account    `json:"accounts"`
	HeaderHashes map[hexutil.Uint64]common.Hash `json:"headerHashes,omitempty"`
}

type account struct {
	Nonce   hexutil.Uint64                    `json:"nonce"`
	Balance *hexutil.U256                     `json:"balance"`
	Code    hexutil.Bytes                     `json:"code"`
	Storage map[w3hexutil.Hash]w3hexutil.Hash `json:"storage,omitempty"`
}

// mergeStates merges the source state into the destination state and returns
// whether the destination state has been modified.
func mergeStates(dst, src *state) (modified bool) {
	// merge accounts
	for addr, acc := range src.Accounts {
		if dstAcc, ok := dst.Accounts[addr]; !ok {
			dst.Accounts[addr] = acc
			modified = true
		} else {
			if dstAcc.Storage == nil {
				dstAcc.Storage = make(map[w3hexutil.Hash]w3hexutil.Hash)
			}

			for slot, storageVal := range acc.Storage {
				if _, ok := dstAcc.Storage[slot]; !ok {
					dstAcc.Storage[slot] = storageVal
					modified = true
				}
			}
			dst.Accounts[addr] = dstAcc
		}
	}

	// merge header hashes
	for blockNumber, hash := range src.HeaderHashes {
		if _, ok := dst.HeaderHashes[blockNumber]; !ok {
			dst.HeaderHashes[blockNumber] = hash
			modified = true
		}
	}

	return modified
}

type storageKey struct {
	addr common.Address
	slot common.Hash
}
