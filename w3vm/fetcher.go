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
	"sync/atomic"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/gofrs/flock"
	"github.com/holiman/uint256"
	"github.com/lmittmann/w3"
	"github.com/lmittmann/w3/internal/crypto"
	w3hexutil "github.com/lmittmann/w3/internal/hexutil"
	"github.com/lmittmann/w3/internal/mod"
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

	dirty uint32 // indicates whether new state has been fetched (0=false, 1=true)

	// file modification times for testdata files
	stateFileModTime        time.Time
	contractsFileModTime    time.Time
	headerHashesFileModTime time.Time
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
	atomic.StoreUint32(&f.dirty, 1)

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
	atomic.StoreUint32(&f.dirty, 1)

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
	atomic.StoreUint32(&f.dirty, 1)

	var (
		header       header
		headerHashCh = make(chan func() (common.Hash, error), 1)
	)
	go func() {
		err := f.call(ethHeaderHash(blockNumber).Returns(&header))
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
	if mod.Root == "" {
		panic("w3vm: NewTestingRPCFetcher must be used in a module test")
	}

	fetcher := newRPCFetcher(client, blockNumber)
	if err := fetcher.loadTestdataState(chainID); err != nil {
		tb.Fatalf("w3vm: failed to load state from testdata: %v", err)
	}

	tb.Cleanup(func() {
		if err := fetcher.storeTestdataState(chainID); err != nil {
			tb.Fatalf("w3vm: failed to write state to testdata: %v", err)
		}
	})
	return fetcher
}

var (
	testdataMutex sync.RWMutex                      // in-process synchronization
	testdataLock  = flock.New(testdataPath("LOCK")) // inter-process synchronization
)

func (f *rpcFetcher) loadTestdataState(chainID uint64) (err error) {
	// lock testdata files
	testdataMutex.RLock()
	defer testdataMutex.RUnlock()
	testdataLock.RLock()
	defer testdataLock.Unlock()

	// read testdata files
	stateFn := fmt.Sprintf("%d_%v.json", chainID, f.blockNumber)
	var state testdataState
	if f.stateFileModTime, err = readTestdata(stateFn, &state, time.Time{}); err != nil {
		return err
	}

	var contracts testdataContracts
	if f.contractsFileModTime, err = readTestdata("contracts.json", &contracts, time.Time{}); err != nil {
		return err
	}

	headerHashesFn := fmt.Sprintf("%d_header_hashes.json", chainID)
	var headerHashes testdataHeaderHashes
	if f.headerHashesFileModTime, err = readTestdata(headerHashesFn, &headerHashes, time.Time{}); err != nil {
		return err
	}

	// build fetcher state
	f.mux.Lock()
	f.mux2.Lock()
	f.mux3.Lock()
	defer f.mux.Unlock()
	defer f.mux2.Unlock()
	defer f.mux3.Unlock()

	for addr, acc := range state {
		codeHash := acc.codeHash()

		f.accounts[addr] = func() (*types.StateAccount, error) {
			return &types.StateAccount{
				Nonce:    uint64(acc.Nonce),
				Balance:  (*uint256.Int)(acc.Balance),
				CodeHash: codeHash[:],
			}, nil
		}
		if _, ok := f.contracts[codeHash]; codeHash != types.EmptyCodeHash && !ok {
			f.contracts[codeHash] = func() ([]byte, error) {
				return contracts[codeHash], nil
			}
		}
		for slot, val := range acc.Storage {
			f.storage[storageKey{addr, (common.Hash)(slot)}] = func() (common.Hash, error) {
				return (common.Hash)(val), nil
			}
		}
		for blockNumber, hash := range headerHashes {
			f.headerHashes[uint64(blockNumber)] = func() (common.Hash, error) {
				return hash, nil
			}
		}
	}
	return nil
}

func (f *rpcFetcher) storeTestdataState(chainID uint64) (err error) {
	if atomic.LoadUint32(&f.dirty) == 0 {
		return nil // if no new state was fetched, we do not need to store it
	}

	// read fetcher state
	f.mux.RLock()
	f.mux2.RLock()
	f.mux3.RLock()
	defer f.mux.RUnlock()
	defer f.mux2.RUnlock()
	defer f.mux3.RUnlock()

	var (
		state        = make(testdataState)
		contracts    = make(testdataContracts)
		headerHashes = make(testdataHeaderHashes)
	)
	for addr, accFunc := range f.accounts {
		acc, err := accFunc()
		if err != nil {
			continue
		}

		state[addr] = &testdataAccount{
			Nonce:   hexutil.Uint64(acc.Nonce),
			Balance: (*hexutil.U256)(acc.Balance),
		}
		if !bytes.Equal(acc.CodeHash, types.EmptyCodeHash[:]) {
			codeHash := common.BytesToHash(acc.CodeHash)
			state[addr].CodeHash = codeHash
			contracts[codeHash], _ = f.contracts[codeHash]()
		}
	}

	for storageKey, storageValFunc := range f.storage {
		storageVal, err := storageValFunc()
		if err != nil {
			continue
		}

		if _, ok := state[storageKey.addr]; !ok {
			state[storageKey.addr] = &testdataAccount{
				Storage: make(map[w3hexutil.Hash]w3hexutil.Hash),
			}
		} else if state[storageKey.addr].Storage == nil {
			state[storageKey.addr].Storage = make(map[w3hexutil.Hash]w3hexutil.Hash)
		}
		state[storageKey.addr].Storage[w3hexutil.Hash(storageKey.slot)] = w3hexutil.Hash(storageVal)
	}

	for blockNumber, hashFunc := range f.headerHashes {
		hash, err := hashFunc()
		if err != nil {
			continue
		}
		headerHashes[hexutil.Uint64(blockNumber)] = hash
	}

	// lock testdata files
	testdataMutex.Lock()
	defer testdataMutex.Unlock()
	testdataLock.Lock()
	defer testdataLock.Unlock()

	// load current testdata state
	stateFn := fmt.Sprintf("%d_%v.json", chainID, f.blockNumber)
	var otherState testdataState
	if _, err = readTestdata(stateFn, &otherState, f.stateFileModTime); err != nil {
		return err
	}

	var otherContracts testdataContracts
	if _, err = readTestdata("contracts.json", &otherContracts, f.contractsFileModTime); err != nil {
		return err
	}

	headerHashesFn := fmt.Sprintf("%d_header_hashes.json", chainID)
	var otherHeaderHashes testdataHeaderHashes
	if _, err = readTestdata(headerHashesFn, &otherHeaderHashes, f.headerHashesFileModTime); err != nil {
		return err
	}

	// merge
	if err := state.Merge(otherState); err != nil {
		return fmt.Errorf("failed to merge testdata state: %w", err)
	}

	if err := contracts.Merge(otherContracts); err != nil {
		return fmt.Errorf("failed to merge testdata contracts: %w", err)
	}

	if err := headerHashes.Merge(otherHeaderHashes); err != nil {
		return fmt.Errorf("failed to merge testdata header hashes: %w", err)
	}

	// write testdata files
	if err := writeTestdata(stateFn, state); err != nil {
		return err
	}
	if err := writeTestdata("contracts.json", contracts); err != nil {
		return err
	}
	if err := writeTestdata(headerHashesFn, headerHashes); err != nil {
		return err
	}

	return nil
}

type storageKey struct {
	addr common.Address
	slot common.Hash
}

// testdataState maps accounts to their state at a specific block in a specific
// chain.
type testdataState map[common.Address]*testdataAccount

func (s testdataState) Merge(other testdataState) error {
	for addr, otherAccount := range other {
		if existingAccount, ok := s[addr]; ok {
			if err := existingAccount.Merge(otherAccount); err != nil {
				return fmt.Errorf("account conflict for address %s: %w", addr, err)
			}
		} else {
			s[addr] = otherAccount
		}
	}
	return nil
}

// testdataAccount represents the state of a single account.
type testdataAccount struct {
	Nonce    hexutil.Uint64                    `json:"nonce"`
	Balance  *hexutil.U256                     `json:"balance"`
	CodeHash common.Hash                       `json:"codeHash,omitzero"`
	Storage  map[w3hexutil.Hash]w3hexutil.Hash `json:"storage,omitempty"`
}

func (a *testdataAccount) codeHash() common.Hash {
	if a.CodeHash == w3.Hash0 {
		return types.EmptyCodeHash
	}
	return a.CodeHash
}

func (a *testdataAccount) Merge(other *testdataAccount) error {
	if a.Nonce != other.Nonce {
		return fmt.Errorf("nonce conflict: %d != %d", a.Nonce, other.Nonce)
	}
	if (*uint256.Int)(a.Balance).Cmp((*uint256.Int)(other.Balance)) != 0 {
		return fmt.Errorf("balance conflict: %s != %s", a.Balance, other.Balance)
	}
	if a.CodeHash != other.CodeHash {
		return fmt.Errorf("code hash conflict: %s != %s", a.CodeHash, other.CodeHash)
	}

	// Merge storage maps
	if a.Storage == nil {
		a.Storage = make(map[w3hexutil.Hash]w3hexutil.Hash)
	}
	for slot, value := range other.Storage {
		if existingValue, ok := a.Storage[slot]; ok {
			if existingValue != value {
				return fmt.Errorf("storage conflict at slot %s: %s != %s",
					(common.Hash)(slot), (common.Hash)(existingValue), (common.Hash)(value),
				)
			}
		} else {
			a.Storage[slot] = value
		}
	}

	return nil
}

// testdataContracts maps code hashes to their code.
type testdataContracts map[common.Hash]hexutil.Bytes

func (c testdataContracts) Merge(other testdataContracts) error {
	for hash, code := range other {
		if existingCode, ok := c[hash]; ok {
			if !bytes.Equal(existingCode, code) {
				return fmt.Errorf("bytecode conflict for code hash %s", hash)
			}
		} else {
			c[hash] = code
		}
	}
	return nil
}

// testdataHeaderHashes maps block numbers to their hashes for a specific chain.
type testdataHeaderHashes map[hexutil.Uint64]common.Hash

func (h testdataHeaderHashes) Merge(other testdataHeaderHashes) error {
	for blockNumber, hash := range other {
		if existingHash, ok := h[blockNumber]; ok {
			if existingHash != hash {
				return fmt.Errorf("header hash conflict for block %d", blockNumber)
			}
		} else {
			h[blockNumber] = hash
		}
	}
	return nil
}

func readTestdata(filename string, data any, onlyIfModifiedAfter time.Time) (time.Time, error) {
	path := testdataPath(filename)

	// get file info first
	info, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return time.Time{}, nil
	} else if err != nil {
		return time.Time{}, err
	}

	if info.ModTime().Before(onlyIfModifiedAfter) {
		return info.ModTime(), nil // file was NOT modified after "onlyIfModifiedAfter"
	}

	// open and read file
	f, err := os.Open(path)
	if err != nil {
		return time.Time{}, err
	}
	defer f.Close()

	if err := json.NewDecoder(f).Decode(data); err != nil {
		return time.Time{}, fmt.Errorf("decode json %s: %w", filename, err)
	}
	return info.ModTime(), nil
}

func writeTestdata(filename string, data any) error {
	path := testdataPath(filename)

	// create "testdata/w3vm"-dir, if it does not exist yet
	dir := filepath.Dir(path)
	if _, err := os.Stat(dir); errors.Is(err, os.ErrNotExist) {
		if err := os.MkdirAll(dir, 0o775); err != nil {
			return err
		}
	}

	// create or open file
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o664)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "\t")
	if err := enc.Encode(data); err != nil {
		return fmt.Errorf("encode json %s: %w", filename, err)
	}
	return nil
}

func testdataPath(filename string) string {
	return filepath.Join(mod.Root, "testdata", "w3vm", filename)
}
