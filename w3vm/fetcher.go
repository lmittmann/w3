package w3vm

import (
	"encoding/json"
	"fmt"
	"math/big"
	"path/filepath"
	"sync"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/holiman/uint256"
	"github.com/lmittmann/w3"
	"github.com/lmittmann/w3/module/eth"
	"github.com/lmittmann/w3/w3types"
	"golang.org/x/sync/singleflight"
)

// Fetcher is the interface to access account state of a blockchain.
type Fetcher interface {

	// Nonce fetches the nonce of the given address.
	Nonce(common.Address) (uint64, error)

	// Balance fetches the balance of the given address.
	Balance(common.Address) (*big.Int, error)

	// Code fetches the code of the given address.
	Code(common.Address) ([]byte, error)

	// StorageAt fetches the state of the given address and storage slot.
	StorageAt(common.Address, common.Hash) (common.Hash, error)

	// HeaderHash fetches the hash of the header with the given number.
	HeaderHash(*big.Int) (common.Hash, error)
}

type rpcFetcher struct {
	client      *w3.Client
	blockNumber *big.Int

	g            *singleflight.Group
	mux          sync.RWMutex
	accounts     map[common.Address]*account
	mux2         sync.RWMutex
	headerHashes map[uint64]common.Hash
}

func NewRPCFetcher(client *w3.Client, blockNumber *big.Int) Fetcher {
	return newRPCFetcher(client, blockNumber)
}

func newRPCFetcher(client *w3.Client, blockNumber *big.Int) *rpcFetcher {
	return &rpcFetcher{
		client:       client,
		blockNumber:  blockNumber,
		g:            new(singleflight.Group),
		accounts:     make(map[common.Address]*account),
		headerHashes: make(map[uint64]common.Hash),
	}
}

func (f *rpcFetcher) Nonce(addr common.Address) (uint64, error) {
	acc, err := f.fetchAccount(addr)
	if err != nil {
		return 0, err
	}
	return acc.Nonce, nil
}

func (f *rpcFetcher) Balance(addr common.Address) (*big.Int, error) {
	acc, err := f.fetchAccount(addr)
	if err != nil {
		return nil, err
	}
	return acc.Balance.ToBig(), nil
}

func (f *rpcFetcher) Code(addr common.Address) ([]byte, error) {
	acc, err := f.fetchAccount(addr)
	if err != nil {
		return nil, err
	}
	return acc.Code, nil
}

func (f *rpcFetcher) StorageAt(addr common.Address, slot common.Hash) (common.Hash, error) {
	storageVal, err := f.fetchStorageAt(addr, *new(uint256.Int).SetBytes32(slot[:]))
	if err != nil {
		return hash0, err
	}
	return storageVal.Bytes32(), nil
}

func (f *rpcFetcher) HeaderHash(blockNumber *big.Int) (common.Hash, error) {
	return f.fetchHeaderHash(blockNumber)
}

func (f *rpcFetcher) fetchAccount(addr common.Address) (*account, error) {
	accAny, err, _ := f.g.Do(string(addr[:]), func() (interface{}, error) {
		// check if account is already cached
		f.mux.RLock()
		acc, ok := f.accounts[addr]
		f.mux.RUnlock()
		if ok {
			return acc, nil
		}

		// fetch account from RPC
		var (
			nonce   uint64
			balance uint256.Int
			code    []byte
		)
		if err := f.call(
			eth.Nonce(addr, f.blockNumber).Returns(&nonce),
			ethBalance(addr, f.blockNumber).Returns(&balance),
			eth.Code(addr, f.blockNumber).Returns(&code),
		); err != nil {
			return nil, err
		}
		acc = &account{
			Nonce:   nonce,
			Balance: balance,
			Code:    code,
			Storage: make(map[uint256.Int]uint256.Int),
		}

		// cache account
		f.mux.Lock()
		f.accounts[addr] = acc
		f.mux.Unlock()
		return acc, nil
	})
	if err != nil {
		return nil, err
	}
	return accAny.(*account), nil
}

func (f *rpcFetcher) fetchStorageAt(addr common.Address, slot uint256.Int) (uint256.Int, error) {
	slotBytes := slot.Bytes32()
	storageValAny, err, _ := f.g.Do(string(append(addr[:], slotBytes[:]...)), func() (interface{}, error) {
		// check if account is already cached
		acc, err := f.fetchAccount(addr)
		if err != nil {
			return uint256.Int{}, err
		}

		// check if storage is already cached
		f.mux.RLock()
		storageVal, ok := acc.Storage[slot]
		f.mux.RUnlock()
		if ok {
			return storageVal, nil
		}

		// fetch storage from RPC
		if err := f.call(
			ethStorageAt(addr, slot, f.blockNumber).Returns(&storageVal),
		); err != nil {
			return uint256.Int{}, err
		}

		// cache storage
		f.mux.Lock()
		acc.Storage[slot] = storageVal
		f.mux.Unlock()
		return storageVal, nil
	})
	if err != nil {
		return uint256.Int{}, err
	}
	return storageValAny.(uint256.Int), nil
}

func (f *rpcFetcher) fetchHeaderHash(blockNumber *big.Int) (common.Hash, error) {
	hashAny, err, _ := f.g.Do(blockNumber.String(), func() (interface{}, error) {
		n := blockNumber.Uint64()

		// check if header hash is already cached
		f.mux2.RLock()
		hash, ok := f.headerHashes[n]
		f.mux2.RUnlock()
		if ok {
			return hash, nil
		}

		// fetch head hash from RPC
		var header types.Header
		if err := f.call(
			eth.HeaderByNumber(blockNumber).Returns(&header),
		); err != nil {
			return nil, err
		}
		hash = header.Hash()

		// cache account
		f.mux2.Lock()
		f.headerHashes[n] = hash
		f.mux2.Unlock()
		return hash, nil
	})
	if err != nil {
		return hash0, err
	}
	return hashAny.(common.Hash), nil
}

func (f *rpcFetcher) call(calls ...w3types.Caller) error {
	return f.client.Call(calls...)
}

func (f *rpcFetcher) setForkState(s *forkState) {
	f.mux.Lock()
	f.mux2.Lock()
	defer f.mux.Unlock()
	defer f.mux2.Unlock()

	for addr, acc := range s.Accounts {
		f.accounts[addr] = acc
	}

	for blockNumber, hash := range s.HeaderHashes {
		f.headerHashes[uint64(blockNumber)] = hash
	}
}

func (f *rpcFetcher) getForkState() *forkState {
	f.mux.Lock()
	f.mux2.Lock()
	defer f.mux.Unlock()
	defer f.mux2.Unlock()

	s := new(forkState)
	if len(f.accounts) > 0 {
		s.Accounts = make(map[common.Address]*account, len(f.accounts))
		for addr, acc := range f.accounts {
			accCop := *acc
			s.Accounts[addr] = &accCop
		}
	}

	if len(f.headerHashes) > 0 {
		s.HeaderHashes = make(map[hexutil.Uint64]common.Hash, len(f.headerHashes))
		for blockNumber, hash := range f.headerHashes {
			s.HeaderHashes[hexutil.Uint64(blockNumber)] = hash
		}
	}
	return s
}

////////////////////////////////////////////////////////////////////////////////////////////////////
// TestingRPCFetcher ///////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////////////////

func NewTestingRPCFetcher(tb testing.TB, client *w3.Client, blockNumber *big.Int) Fetcher {
	fp := getTbFilepath(tb)
	if ok := isTbInMod(fp); !ok {
		panic("must be called from a test in a module")
	}

	if blockNumber == nil {
		tb.Fatal("w3vm: block number must not be <nil>")
	}

	var chainID uint64
	if err := client.Call(
		eth.ChainID().Returns(&chainID),
	); err != nil {
		tb.Fatalf("w3vm: failed to fetch chain ID: %v", err)
	}

	fp = filepath.Join(fp, "testdata", "w3vm", fmt.Sprintf("%d_%v.json", chainID, blockNumber))
	testdataState, err := readTestdataState(fp)
	if err != nil {
		tb.Fatalf("w3vm: failed to read state from testdata: %v", err)
	}

	fetcher := newRPCFetcher(client, blockNumber)
	fetcher.setForkState(testdataState)

	tb.Cleanup(func() {
		postTestdataState := fetcher.getForkState()
		if err := writeTestdataState(fp, postTestdataState); err != nil {
			tb.Errorf("w3test: failed to write state to testdata: %v", err)
		}
	})

	return fetcher
}

////////////////////////////////////////////////////////////////////////////////////////////////////
// account /////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////////////////

type account struct {
	Nonce   uint64
	Balance uint256.Int
	Code    []byte
	Storage map[uint256.Int]uint256.Int
}

type accountMarshaling struct {
	Nonce   hexutil.Uint64                  `json:"nonce"`
	Balance uint256OrHash                   `json:"balance"`
	Code    hexutil.Bytes                   `json:"code"`
	Storage map[uint256OrHash]uint256OrHash `json:"storage,omitempty"`
}

func (acc account) MarshalJSON() ([]byte, error) {
	storage := make(map[uint256OrHash]uint256OrHash, len(acc.Storage))
	for slot, val := range acc.Storage {
		storage[uint256OrHash(slot)] = uint256OrHash(val)
	}
	return json.Marshal(accountMarshaling{
		Nonce:   hexutil.Uint64(acc.Nonce),
		Balance: uint256OrHash(acc.Balance),
		Code:    acc.Code,
		Storage: storage,
	})
}

func (acc *account) UnmarshalJSON(data []byte) error {
	var dec accountMarshaling
	if err := json.Unmarshal(data, &dec); err != nil {
		return err
	}

	acc.Nonce = uint64(dec.Nonce)
	acc.Balance = uint256.Int(dec.Balance)
	acc.Code = dec.Code
	acc.Storage = make(map[uint256.Int]uint256.Int)
	for slot, val := range dec.Storage {
		acc.Storage[uint256.Int(slot)] = uint256.Int(val)
	}
	return nil
}
