# `w3`: Enhanced Ethereum Integration for Go

[![Go Reference](https://pkg.go.dev/badge/github.com/lmittmann/w3.svg)](https://pkg.go.dev/github.com/lmittmann/w3)
[![Go Report Card](https://goreportcard.com/badge/github.com/lmittmann/w3)](https://goreportcard.com/report/github.com/lmittmann/w3)
[![Coverage Status](https://coveralls.io/repos/github/lmittmann/w3/badge.svg?branch=main)](https://coveralls.io/github/lmittmann/w3?branch=main)
[![Latest Release](https://img.shields.io/github/v/release/lmittmann/w3)](https://github.com/lmittmann/w3/releases)
<img src="https://w3.cool/gopher.png" align="right" alt="W3 Gopher" width="158" height="224">

`w3` is your toolbelt for integrating with Ethereum in Go. Closely linked to `go‑ethereum`, it provides an ergonomic wrapper for working with **RPC**, **ABI's**, and the **EVM**.


```
go get github.com/lmittmann/w3
```


## At a Glance

* Use `w3.Client` to connect to an RPC endpoint. The client features batch request support for up to **80x faster requests** and easy extendibility. [learn&nbsp;more&nbsp;↗](#rpc-client)
* Use `w3vm.VM` to simulate EVM execution with optional tracing and Mainnet state forking, or test Smart Contracts. [learn&nbsp;more&nbsp;↗](#vm)
* Use `w3.Func` and `w3.Event` to create ABI bindings from Solidity function and event signatures. [learn&nbsp;more&nbsp;↗](#abi-bindings)
* Use `w3.A`, `w3.H`, and many other utility functions to parse addresses, hashes, and other common types from strings. [learn&nbsp;more&nbsp;↗](#utils)


## Getting Started

### RPC Client

[`w3.Client`](https://pkg.go.dev/github.com/lmittmann/w3#Client) is a batch request focused RPC client that can be used to connect to an Ethereum node via HTTP, WebSocket, or IPC. Its modular API allows to create custom RPC method integrations that can be used alongside the common methods implemented by this package.

**Example:** Batch Request ([Playground](https://pkg.go.dev/github.com/lmittmann/w3#example-Client))

```go
// 1. Connect to an RPC endpoint
client, err := w3.Dial("https://rpc.ankr.com/eth")
if err != nil {
    // handle error
}
defer client.Close()

// 2. Make a batch request
var (
    balance *big.Int
    nonce   uint64
)
if err := client.Call(
    eth.Balance(addr, nil).Returns(&balance),
    eth.Nonce(addr, nil).Returns(&nonce),
); err != nil {
    // handle error
}
```

> [!NOTE]
> #### Why send batch requests?
> Most of the time you need to call multiple RPC methods to get the data you need. When you make separate requests per RPC call you need a single round trip to the server for each call. This can be slow, especially for remote endpoints. Batching multiple RPC calls into a single request only requires a single round trip, and speeds up RPC calls significantly.

#### Error Handling

If one ore more calls in a batch request fail, `Client.Call` returns an error of type [`w3.CallErrors`](https://pkg.go.dev/github.com/lmittmann/w3#CallErrors).

**Example:** Check which RPC calls failed in a batch request ([Playground](https://pkg.go.dev/github.com/lmittmann/w3#example-CallErrors))
```go
var errs w3.CallErrors
if err := client.Call(rpcCalls...); errors.As(err, &errs) {
    // handle call errors
} else if err != nil {
    // handle other errors
}
```

#### Learn More
* List of supported [**RPC methods**](#rpc-methods)
* Learn how to create [**custom RPC method bindings**](#custom-rpc-method-bindings)

### VM

[`w3vm.VM`](https://pkg.go.dev/github.com/lmittmann/w3/w3vm#VM) is a high-level EVM environment with a simple but powerful API to simulate EVM execution, test Smart Contracts, or trace transactions. It supports Mainnet state forking via RPC and state caching for faster testing.

**Example:** Simulate an Uniswap v3 swap ([Playground](https://pkg.go.dev/github.com/lmittmann/w3/w3vm#example-VM))

```go
// 1. Create a VM that forks the Mainnet state from the latest block,
// disables the base fee, and has a fake WETH balance and approval for the router
vm, err := w3vm.New(
    w3vm.WithFork(client, nil),
    w3vm.WithNoBaseFee(),
    w3vm.WithState(w3types.State{
        addrWETH: {Storage: w3types.Storage{
            w3vm.WETHBalanceSlot(addrEOA):               common.BigToHash(w3.I("1 ether")),
            w3vm.WETHAllowanceSlot(addrEOA, addrRouter): common.BigToHash(w3.I("1 ether")),
        }},
    }),
)
if err != nil {
    // handle error
}

// 2. Simulate a Uniswap v3 swap
receipt, err := vm.Apply(&w3types.Message{
    From: addrEOA,
    To:   &addrRouter,
    Func: funcExactInput,
    Args: []any{&ExactInputParams{
        Path:             encodePath(addrWETH, 500, addrUNI),
        Recipient:        addrEOA,
        Deadline:         big.NewInt(time.Now().Unix()),
        AmountIn:         w3.I("1 ether"),
        AmountOutMinimum: w3.Big0,
    }},
})
if err != nil {
    // handle error
}

// 3. Decode output amount
var amountOut *big.Int
if err := receipt.DecodeReturns(&amountOut); err != nil {
    // handle error
}
```

### ABI Bindings

ABI bindings in `w3` are specified for individual functions using Solidity syntax and are usable for any contract that supports that function.

**Example:** ABI binding for the ERC20 `balanceOf` function ([Playground](https://pkg.go.dev/github.com/lmittmann/w3#example-NewFunc-BalanceOf))

```go
funcBalanceOf := w3.MustNewFunc("balanceOf(address)", "uint256")
```

**Example:** ABI binding for the Uniswap v4 `swap` function ([Playground](https://pkg.go.dev/github.com/lmittmann/w3#example-NewFunc-UniswapV4Swap))

```go
funcSwap := w3.MustNewFunc(`swap(
    (address currency0, address currency1, uint24 fee, int24 tickSpacing, address hooks) key,
    (bool zeroForOne, int256 amountSpecified, uint160 sqrtPriceLimitX96) params,
    bytes hookData
)`, "int256 delta")
```

A [`Func`](https://pkg.go.dev/github.com/lmittmann/w3#Func) can be used to

* encode arguments to the contracts input data ([`Func.EncodeArgs`](https://pkg.go.dev/github.com/lmittmann/w3#Func.EncodeArgs)),
* decode arguments from the contracts input data ([`Func.DecodeArgs`](https://pkg.go.dev/github.com/lmittmann/w3#Func.DecodeArgs)), and
* decode returns form the contracts output data ([`Func.DecodeReturns`](https://pkg.go.dev/github.com/lmittmann/w3#Func.DecodeReturns)).

### Utils

Static addresses, hashes, bytes or integers can be parsed from (hex-)strings with the following utility functions that panic if the string is not valid.

```go
addr := w3.A("0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045")
hash := w3.H("0xd4e56740f876aef8c010b86a40d5f56745a118d0906a34e69aec8c0db1cb8fa3")
bytes := w3.B("0x27c5342c")
amount := w3.I("12.34 ether")
```

Use [go-ethereum/common](https://pkg.go.dev/github.com/ethereum/go-ethereum/common) to parse strings that may not be valid instead.


## RPC Methods

List of supported RPC methods for [`w3.Client`](https://pkg.go.dev/github.com/lmittmann/w3#Client).

### [`eth`](https://pkg.go.dev/github.com/lmittmann/w3/module/eth)

| Method                                    | Go Code
| :---------------------------------------- | :-------
| `eth_blockNumber`                         | `eth.BlockNumber().Returns(blockNumber **big.Int)`
| `eth_call`                                | `eth.Call(msg *w3types.Message, blockNumber *big.Int, overrides w3types.State).Returns(output *[]byte)`<br>`eth.CallFunc(contract common.Address, f w3types.Func, args ...any).Returns(returns ...any)`
| `eth_chainId`                             | `eth.ChainID().Returns(chainID *uint64)`
| `eth_createAccessList`                    | `eth.AccessList(msg *w3types.Message, blockNumber *big.Int).Returns(resp **eth.AccessListResponse)`
| `eth_estimateGas`                         | `eth.EstimateGas(msg *w3types.Message, blockNumber *big.Int).Returns(gas *uint64)`
| `eth_gasPrice`                            | `eth.GasPrice().Returns(gasPrice **big.Int)`
| `eth_maxPriorityFeePerGas`                | `eth.GasTipCap().Returns(gasTipCap **big.Int)`
| `eth_getBalance`                          | `eth.Balance(addr common.Address, blockNumber *big.Int).Returns(balance **big.Int)`
| `eth_getBlockByHash`                      | `eth.BlockByHash(hash common.Hash).Returns(block *types.Block)`<br>`eth.HeaderByHash(hash common.Hash).Returns(header **types.Header)`
| `eth_getBlockByNumber`                    | `eth.BlockByNumber(number *big.Int).Returns(block *types.Block)`<br>`eth.HeaderByNumber(number *big.Int).Returns(header **types.Header)`
| `eth_getBlockReceipts`                    | `eth.BlockReceipts(blockNumber *big.Int).Returns(receipts *types.Receipts)`
| `eth_getBlockTransactionCountByHash`      | `eth.BlockTxCountByHash(hash common.Hash).Returns(count *uint)`
| `eth_getBlockTransactionCountByNumber`    | `eth.BlockTxCountByNumber(number *big.Int).Returns(count *uint)`
| `eth_getCode`                             | `eth.Code(addr common.Address, blockNumber *big.Int).Returns(code *[]byte)`
| `eth_getLogs`                             | `eth.Logs(q ethereum.FilterQuery).Returns(logs *[]types.Log)`
| `eth_getStorageAt`                        | `eth.StorageAt(addr common.Address, slot common.Hash, blockNumber *big.Int).Returns(storage *common.Hash)`
| `eth_getTransactionByHash`                | `eth.Tx(hash common.Hash).Returns(tx **types.Transaction)`
| `eth_getTransactionByBlockHashAndIndex`   | `eth.TxByBlockHashAndIndex(blockHash common.Hash, index uint).Returns(tx **types.Transaction)`
| `eth_getTransactionByBlockNumberAndIndex` | `eth.TxByBlockNumberAndIndex(blockNumber *big.Int, index uint).Returns(tx **types.Transaction)`
| `eth_getTransactionCount`                 | `eth.Nonce(addr common.Address, blockNumber *big.Int).Returns(nonce *uint)`
| `eth_getTransactionReceipt`               | `eth.TxReceipt(txHash common.Hash).Returns(receipt **types.Receipt)`
| `eth_sendRawTransaction`                  | `eth.SendRawTx(rawTx []byte).Returns(hash *common.Hash)`<br>`eth.SendTx(tx *types.Transaction).Returns(hash *common.Hash)`
| `eth_getUncleByBlockHashAndIndex`         | `eth.UncleByBlockHashAndIndex(hash common.Hash, index uint).Returns(uncle **types.Header)`
| `eth_getUncleByBlockNumberAndIndex`       | `eth.UncleByBlockNumberAndIndex(number *big.Int, index uint).Returns(uncle **types.Header)`
| `eth_getUncleCountByBlockHash`            | `eth.UncleCountByBlockHash(hash common.Hash).Returns(count *uint)`
| `eth_getUncleCountByBlockNumber`          | `eth.UncleCountByBlockNumber(number *big.Int).Returns(count *uint)`

### [`debug`](https://pkg.go.dev/github.com/lmittmann/w3/module/debug)

| Method                   | Go Code
| :----------------------- | :-------
| `debug_traceCall`        | `debug.TraceCall(msg *w3types.Message, blockNumber *big.Int, config *debug.TraceConfig).Returns(trace **debug.Trace)`<br>`debug.CallTraceCall(msg *w3types.Message, blockNumber *big.Int, overrides w3types.State).Returns(trace **debug.CallTrace)`
| `debug_traceTransaction` | `debug.TraceTx(txHash common.Hash, config *debug.TraceConfig).Returns(trace **debug.Trace)`<br>`debug.CallTraceTx(txHash common.Hash, overrides w3types.State).Returns(trace **debug.CallTrace)`

### [`txpool`](https://pkg.go.dev/github.com/lmittmann/w3/module/txpool)

| Method               | Go Code
| :--------------------| :-------
| `txpool_content`     | `txpool.Content().Returns(resp **txpool.ContentResponse)`
| `txpool_contentFrom` | `txpool.ContentFrom(addr common.Address).Returns(resp **txpool.ContentFromResponse)`
| `txpool_status`      | `txpool.Status().Returns(resp **txpool.StatusResponse)`

### [`web3`](https://pkg.go.dev/github.com/lmittmann/w3/module/web3)

| Method               | Go Code
| :------------------- | :-------
| `web3_clientVersion` | `web3.ClientVersion().Returns(clientVersion *string)`

### Third Party RPC Method Packages

| Package                                                                  | Description
| :----------------------------------------------------------------------- | :-----------
| [github.com/lmittmann/flashbots](https://github.com/lmittmann/flashbots) | Package `flashbots` implements RPC API bindings for the Flashbots relay and mev-geth.


## Custom RPC Method Bindings

Custom RPC method bindings can be created by implementing the [`w3types.RPCCaller`](https://pkg.go.dev/github.com/lmittmann/w3/w3types#RPCCaller) interface.

**Example:** RPC binding for the Otterscan `ots_getTransactionBySenderAndNonce` method ([Playground](https://pkg.go.dev/github.com/lmittmann/w3/w3types#example-RPCCaller-GetTransactionBySenderAndNonce))

```go
// TxBySenderAndNonceFactory requests the senders transaction hash by the nonce.
func TxBySenderAndNonceFactory(sender common.Address, nonce uint64) w3types.RPCCallerFactory[common.Hash] {
    return &getTransactionBySenderAndNonceFactory{
        sender: sender,
        nonce:  nonce,
    }
}

// getTransactionBySenderAndNonceFactory implements the w3types.RPCCaller and
// w3types.RPCCallerFactory interfaces. It stores the method parameters and
// the the reference to the return value.
type getTransactionBySenderAndNonceFactory struct {
    // params
    sender common.Address
    nonce  uint64

    // returns
    returns *common.Hash
}

// Returns sets the reference to the return value.
func (f *getTransactionBySenderAndNonceFactory) Returns(txHash *common.Hash) w3types.RPCCaller {
    f.returns = txHash
    return f
}

// CreateRequest creates a batch request element for the Otterscan getTransactionBySenderAndNonce method.
func (f *getTransactionBySenderAndNonceFactory) CreateRequest() (rpc.BatchElem, error) {
    return rpc.BatchElem{
        Method: "ots_getTransactionBySenderAndNonce",
        Args:   []any{f.sender, f.nonce},
        Result: f.returns,
    }, nil
}

// HandleResponse handles the response of the Otterscan getTransactionBySenderAndNonce method.
func (f *getTransactionBySenderAndNonceFactory) HandleResponse(elem rpc.BatchElem) error {
    if err := elem.Error; err != nil {
        return err
    }
    return nil
}
```
