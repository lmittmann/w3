# `w3`: Enhanced Ethereum Integration for Go

[![Go Reference](https://pkg.go.dev/badge/github.com/lmittmann/w3.svg)](https://pkg.go.dev/github.com/lmittmann/w3)
[![Go Report Card](https://goreportcard.com/badge/github.com/lmittmann/w3)](https://goreportcard.com/report/github.com/lmittmann/w3)
[![Coverage Status](https://coveralls.io/repos/github/lmittmann/w3/badge.svg?branch=main)](https://coveralls.io/github/lmittmann/w3?branch=main)
[![Latest Release](https://img.shields.io/github/v/release/lmittmann/w3)](https://github.com/lmittmann/w3/releases)
<img src="https://w3.cool/gopher.png" align="right" alt="W3 Gopher" width="158" height="224">

`w3` is your toolbelt for integrating with Ethereum in Go. Closely linked to `go-ethereum`, it provides an ergonomic wrapper for working with **RPC**, **ABI's**, and the **EVM**.


```
go get github.com/lmittmann/w3
```


## At a Glance

* Use `w3.Client` to connect to an RPC endpoint. The client features batch request support for up to **80x faster requests** and easy extendibility. [learn more ↗](#rpc-client)
* Use `w3vm.VM` to simulate EVM execution with optional tracing and Mainnet state forking, or test Smart Contracts. [learn more ↗](#vm)
* Use `w3.Func` and `w3.Event` to create ABI bindings from Solidity function and event signatures. [learn more ↗](#abi-bindings)
* Use `w3.A`, `w3.H`, and many other utility functions to parse addresses, hashes, and other common types from strings. [learn more ↗](#utils)


## Getting Started

### RPC Client

### VM

### ABI Bindings

### Utils

<!-- --------------------------------------------------------------------------------------------------------------- -->

Package `w3` implements a blazing fast and modular Ethereum JSON RPC client with
first-class ABI support.

* **Batch request** support significantly reduces the duration of requests to
  both remote and local endpoints.
* **ABI** bindings are specified for individual functions using Solidity syntax.
  No need for `abigen` and ABI JSON files.
* **Modular** API allows to create custom RPC method integrations that can be
  used alongside the methods implemented by the package.

`w3` is closely linked to [go-ethereum](https://github.com/ethereum/go-ethereum)
and uses a variety of its types, such as [`common.Address`](https://pkg.go.dev/github.com/ethereum/go-ethereum/common#Address)
or [`types.Transaction`](https://pkg.go.dev/github.com/ethereum/go-ethereum/core/types#Transaction).

Batch requests with `w3` are up to **85x faster** than sequential requests with
`go-ethereum/ethclient`.

<details>
<summary>Benchmarks</summary>
<pre>
name               ethclient time/op  w3 time/op  delta
Call_BalanceNonce  78.3ms ± 2%        39.0ms ± 1%  -50.15%  (p=0.000 n=23+22)
Call_Balance100     3.90s ± 5%         0.05s ± 2%  -98.84%  (p=0.000 n=20+24)
Call_BalanceOf100   3.99s ± 3%         0.05s ± 2%  -98.73%  (p=0.000 n=22+23)
Call_Block100       6.89s ± 7%         1.94s ±11%  -71.77%  (p=0.000 n=24+23)
</pre>
</details>

## Getting Started

> **Note**
> Check out the [examples](https://github.com/lmittmann/w3/tree/main/examples)!

Connect to an RPC endpoint via HTTP, WebSocket, or IPC using [`Dial`](https://pkg.go.dev/github.com/lmittmann/w3#Dial)
or [`MustDial`](https://pkg.go.dev/github.com/lmittmann/w3#MustDial).

```go
// Connect (or panic on error)
client := w3.MustDial("https://rpc.ankr.com/eth")
defer client.Close()
```


## Batch Requests

Batch request support in the [`Client`](https://pkg.go.dev/github.com/lmittmann/w3#Client)
allows to send multiple RPC requests in a single HTTP request. The speed gains
to remote endpoints are huge. Fetching 100 blocks in a single batch request
with `w3` is ~80x faster compared to sequential requests with `ethclient`.

Example: Request the nonce and balance of an address in a single request

```go
var (
	addr = w3.A("0x000000000000000000000000000000000000c0Fe")

	nonce   uint64
	balance big.Int
)
err := client.Call(
	eth.Nonce(addr, nil).Returns(&nonce),
	eth.Balance(addr, nil).Returns(&balance),
)
```


## ABI Bindings

ABI bindings in `w3` are specified for individual functions using Solidity
syntax and are usable for any contract that supports that function.

Example: ABI binding for the ERC20-function `balanceOf`

```go
funcBalanceOf := w3.MustNewFunc("balanceOf(address)", "uint256")
```

A [`Func`](https://pkg.go.dev/github.com/lmittmann/w3#Func) can be used to

* encode arguments to the contracts input data ([`Func.EncodeArgs`](https://pkg.go.dev/github.com/lmittmann/w3#Func.EncodeArgs)),
* decode arguments from the contracts input data ([`Func.DecodeArgs`](https://pkg.go.dev/github.com/lmittmann/w3#Func.DecodeArgs)), and
* decode returns form the contracts output data ([`Func.DecodeReturns`](https://pkg.go.dev/github.com/lmittmann/w3#Func.DecodeReturns)).

### Reading Contracts

[`Func`](https://pkg.go.dev/github.com/lmittmann/w3#Func)'s can be used with
[`eth.CallFunc`](https://pkg.go.dev/github.com/lmittmann/w3/module/eth#CallFunc)
in the client to read contract data.

```go
var (
	weth9 = w3.A("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2")
	dai   = w3.A("0x6B175474E89094C44Da98b954EedeAC495271d0F")

	weth9Balance big.Int
	daiBalance   big.Int
)

err := client.Call(
	eth.CallFunc(weth9, funcBalanceOf, addr).Returns(&weth9Balance),
	eth.CallFunc(dai, funcBalanceOf, addr).Returns(&daiBalance),
)
```

### Writing Contracts

Sending a transaction to a contract requires three steps.

1. Encode the transaction input data using [`Func.EncodeArgs`](https://pkg.go.dev/github.com/lmittmann/w3#Func.EncodeArgs).

```go
var funcTransfer = w3.MustNewFunc("transfer(address,uint256)", "bool")

input, err := funcTransfer.EncodeArgs(w3.A("0x…"), w3.I("1 ether"))
```

2. Create a signed transaction to the contract using [go-ethereum/types](https://github.com/ethereum/go-ethereum).

```go
signer := types.LatestSigner(params.MainnetChainConfig)
tx := types.MustSignNewTx(privKey, signer, &types.DynamicFeeTx{
	To:        w3.A("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2"),
	Nonce:     0,
	Data:      input,
	Gas:       75000,
	GasFeeCap: w3.I("100 gwei"),
	GasTipCap: w3.I("1 gwei"),
})
```

3. Send the signed transaction.

```go
var txHash common.Hash
err := client.Call(
	eth.SendTx(tx).Returns(&txHash),
)
```


## Custom RPC Methods

Custom RPC methods can be called with the `w3` client by creating a
[`w3types.Caller`](https://pkg.go.dev/github.com/lmittmann/w3/w3types#Caller)
implementation.
The `w3/module/eth` package can be used as implementation reference.


## Utils

Static addresses, hashes, hex byte slices or `big.Int`'s can be parsed from
strings with the following utility functions.

```go
var (
	addr  = w3.A("0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045")
	hash  = w3.H("0xd4e56740f876aef8c010b86a40d5f56745a118d0906a34e69aec8c0db1cb8fa3")
	bytes = w3.B("0x27c5342c")
	big   = w3.I("12.34 ether")
)
```

Note that these functions panic if the string cannot be parsed. Use
[go-ethereum/common](https://pkg.go.dev/github.com/ethereum/go-ethereum/common)
to parse strings that may not be valid instead.


## RPC Methods

List of supported RPC methods.

### [`eth`](https://pkg.go.dev/github.com/lmittmann/w3/module/eth)

| Method                                    | Go Code
| :---------------------------------------- | :-------
| `eth_blockNumber`                         | `eth.BlockNumber().Returns(blockNumber *big.Int)`
| `eth_call`                                | `eth.Call(msg *w3types.Message, blockNumber *big.Int, overrides w3types.State).Returns(output *[]byte)`<br>`eth.CallFunc(contract common.Address, f w3types.Func, args ...any).Returns(returns ...any)`
| `eth_chainId`                             | `eth.ChainID().Returns(chainID *uint64)`
| `eth_createAccessList`                    | `eth.AccessList(msg *w3types.Message, blockNumber *big.Int).Returns(resp *eth.AccessListResponse)`
| `eth_estimateGas`                         | `eth.EstimateGas(msg *w3types.Message, blockNumber *big.Int).Returns(gas *uint64)`
| `eth_gasPrice`                            | `eth.GasPrice().Returns(gasPrice *big.Int)`
| `eth_maxPriorityFeePerGas`                | `eth.GasTipCap().Returns(gasTipCap *big.Int)`
| `eth_getBalance`                          | `eth.Balance(addr common.Address, blockNumber *big.Int).Returns(balance *big.Int)`
| `eth_getBlockByHash`                      | `eth.BlockByHash(hash common.Hash).Returns(block *types.Block)`<br>`eth.HeaderByHash(hash common.Hash).Returns(header *types.Header)`
| `eth_getBlockByNumber`                    | `eth.BlockByNumber(number *big.Int).Returns(block *types.Block)`<br>`eth.HeaderByNumber(number *big.Int).Returns(header *types.Header)`
| `eth_getBlockReceipts`                    | `eth.BlockReceipts(blockNumber *big.Int).Returns(receipts *types.Receipts)`
| `eth_getBlockTransactionCountByHash`      | `eth.BlockTxCountByHash(hash common.Hash).Returns(count *uint)`
| `eth_getBlockTransactionCountByNumber`    | `eth.BlockTxCountByNumber(number *big.Int).Returns(count *uint)`
| `eth_getCode`                             | `eth.Code(addr common.Address, blockNumber *big.Int).Returns(code *[]byte)`
| `eth_getLogs`                             | `eth.Logs(q ethereum.FilterQuery).Returns(logs *[]types.Log)`
| `eth_getStorageAt`                        | `eth.StorageAt(addr common.Address, slot common.Hash, blockNumber *big.Int).Returns(storage *common.Hash)`
| `eth_getTransactionByHash`                | `eth.Tx(hash common.Hash).Returns(tx *types.Transaction)`
| `eth_getTransactionByBlockHashAndIndex`   | `eth.TxByBlockHashAndIndex(blockHash common.Hash, index uint).Returns(tx *types.Transaction)`
| `eth_getTransactionByBlockNumberAndIndex` | `eth.TxByBlockNumberAndIndex(blockNumber *big.Int, index uint).Returns(tx *types.Transaction)`
| `eth_getTransactionCount`                 | `eth.Nonce(addr common.Address, blockNumber *big.Int).Returns(nonce *uint)`
| `eth_getTransactionReceipt`               | `eth.TxReceipt(txHash common.Hash).Returns(receipt *types.Receipt)`
| `eth_sendRawTransaction`                  | `eth.SendRawTx(rawTx []byte).Returns(hash *common.Hash)`<br>`eth.SendTx(tx *types.Transaction).Returns(hash *common.Hash)`
| `eth_getUncleByBlockHashAndIndex`         | `eth.UncleByBlockHashAndIndex(hash common.Hash, index uint).Returns(uncle *types.Header)`
| `eth_getUncleByBlockNumberAndIndex`       | `eth.UncleByBlockNumberAndIndex(number *big.Int, index uint).Returns(uncle *types.Header)`
| `eth_getUncleCountByBlockHash`            | `eth.UncleCountByBlockHash(hash common.Hash).Returns(count *uint)`
| `eth_getUncleCountByBlockNumber`          | `eth.UncleCountByBlockNumber(number *big.Int).Returns(count *uint)`

### [`debug`](https://pkg.go.dev/github.com/lmittmann/w3/module/debug)

| Method                   | Go Code
| :----------------------- | :-------
| `debug_traceCall`        | `debug.TraceCall(msg *w3types.Message, blockNumber *big.Int, config *debug.TraceConfig).Returns(trace *debug.Trace)`<br>`debug.CallTraceCall(msg *w3types.Message, blockNumber *big.Int, overrides w3types.State).Returns(trace *debug.CallTrace)`
| `debug_traceTransaction` | `debug.TraceTx(txHash common.Hash, config *debug.TraceConfig).Returns(trace *debug.Trace)`<br>`debug.CallTraceTx(txHash common.Hash, overrides w3types.State).Returns(trace *debug.CallTrace)`

### [`txpool`](https://pkg.go.dev/github.com/lmittmann/w3/module/txpool)

| Method               | Go Code
| :--------------------| :-------
| `txpool_content`     | `txpool.Content().Returns(resp *txpool.ContentResponse)`
| `txpool_contentFrom` | `txpool.ContentFrom(addr common.Address).Returns(resp *txpool.ContentFromResponse)`
| `txpool_status`      | `txpool.Status().Returns(resp *txpool.StatusResponse)`

### [`web3`](https://pkg.go.dev/github.com/lmittmann/w3/module/web3)

| Method               | Go Code
| :------------------- | :-------
| `web3_clientVersion` | `web3.ClientVersion().Returns(clientVersion *string)`

### Third Party RPC Method Packages

| Package                                                                  | Description
| :----------------------------------------------------------------------- | :-----------
| [github.com/lmittmann/flashbots](https://github.com/lmittmann/flashbots) | Package `flashbots` implements RPC API bindings for the Flashbots relay and mev-geth.
