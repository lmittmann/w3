package state

import (
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/common"
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

type RPCFetcher struct {
	client      *w3.Client
	blockNumber *big.Int

	g            *singleflight.Group
	mux          sync.RWMutex
	accounts     map[common.Address]account2
	mux2         sync.RWMutex
	headerHashes map[uint64]common.Hash
}

func NewRPCFetcher(client *w3.Client, blockNumber *big.Int) *RPCFetcher {
	return &RPCFetcher{
		client:       client,
		blockNumber:  blockNumber,
		g:            new(singleflight.Group),
		accounts:     make(map[common.Address]account2),
		headerHashes: make(map[uint64]common.Hash),
	}
}

func (f *RPCFetcher) Nonce(addr common.Address) (uint64, error) {
	acc, err := f.fetchAccount(addr)
	if err != nil {
		return 0, err
	}
	return acc.Nonce, nil
}

func (f *RPCFetcher) Balance(addr common.Address) (*big.Int, error) {
	acc, err := f.fetchAccount(addr)
	if err != nil {
		return nil, err
	}
	return acc.Balance.ToBig(), nil
}

func (f *RPCFetcher) Code(addr common.Address) ([]byte, error) {
	acc, err := f.fetchAccount(addr)
	if err != nil {
		return nil, err
	}
	return acc.Code, nil
}

func (f *RPCFetcher) StorageAt(addr common.Address, slot common.Hash) (common.Hash, error) {
	storageVal, err := f.fetchStorageAt(addr, *new(uint256.Int).SetBytes32(slot[:]))
	if err != nil {
		return hash0, err
	}
	return storageVal.Bytes32(), nil
}

func (f *RPCFetcher) HeaderHash(blockNumber *big.Int) (common.Hash, error) {
	return f.fetchHeaderHash(blockNumber)
}

func (f *RPCFetcher) fetchAccount(addr common.Address) (account2, error) {
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
		acc = account2{
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
		return account2{}, err
	}
	return accAny.(account2), nil
}

func (f *RPCFetcher) fetchStorageAt(addr common.Address, slot uint256.Int) (uint256.Int, error) {
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

func (f *RPCFetcher) fetchHeaderHash(blockNumber *big.Int) (common.Hash, error) {
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

func (f *RPCFetcher) call(calls ...w3types.Caller) error {
	return f.client.Call(calls...)
}

type account2 struct {
	Nonce   uint64
	Balance uint256.Int
	Code    []byte
	Storage map[uint256.Int]uint256.Int
}
