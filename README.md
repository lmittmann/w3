# w3

[![Go Reference](https://pkg.go.dev/badge/github.com/lmittmann/w3.svg)](https://pkg.go.dev/github.com/lmittmann/w3)
[![Go Report Card](https://goreportcard.com/badge/github.com/lmittmann/w3)](https://goreportcard.com/report/github.com/lmittmann/w3)
[![Coverage Status](https://coveralls.io/repos/github/lmittmann/w3/badge.svg?branch=main)](https://coveralls.io/github/lmittmann/w3?branch=main)
[![Latest Release](https://img.shields.io/github/v/release/lmittmann/w3)](https://github.com/lmittmann/w3/releases)

<img src="https://user-images.githubusercontent.com/3458786/153202258-24bf253e-5ab0-4efd-a0ed-43dc1bf093c9.png" align="right" alt="W3 Gopher" width="158" height="224">

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


## Install

```
go get github.com/lmittmann/w3
```


## Getting Started

Connect to an RPC endpoint via HTTP, WebSocket, or IPC using [`Dial`](https://pkg.go.dev/github.com/lmittmann/w3#Dial)
or [`MustDial`](https://pkg.go.dev/github.com/lmittmann/w3#MustDial).

```go
// Connect (or panic on error)
client := w3.MustDial("https://cloudflare-eth.com")
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
	eth.CallFunc(funcBalanceOf, weth9, addr).Returns(&weth9Balance),
	eth.CallFunc(funcBalanceOf, dai, addr).Returns(&daiBalance),
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
var (
	signer = types.LatestSignerForChainID(params.MainnetChainConfig.ChainID)
	weth9  = w3.A("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2")
)

tx, err := types.SignNewTx(privKey, signer, &types.DynamicFeeTx{
	To:        weth9,
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
	eth.SendTransaction(tx).Returns(&txHash),
)
```


## Custom RPC Methods

Custom RPC methods can be called with the `w3` client by creating a
[`core.Caller`](https://pkg.go.dev/github.com/lmittmann/w3/core#Caller)
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

### `eth`

| Method                      | Go Code
| :-------------------------- | :-------
| `eth_blockNumber`           | `eth.BlockNumber().Returns(blockNumber *big.Int)`
| `eth_call`                  | `eth.Call(msg ethereum.CallMsg, blockNumber *big.Int).Returns(output *[]byte)`<br>`eth.CallFunc(fn core.Func, contract common.Address, args ...any).Returns(returns ...any)`
| `eth_chainId`               | `eth.ChainID().Returns(chainID *uint64)`
| `eth_gasPrice`              | `eth.GasPrice().Returns(gasPrice *big.Int)`
| `eth_getBalance`            | `eth.Balance(addr common.Address, blockNumber *big.Int).Returns(balance *big.Int)`
| `eth_getBlockByHash`        | `eth.BlockByHash(hash common.Hash).Returns(block *types.Block)`<br>`eth.BlockByHash(hash common.Hash).ReturnsRAW(block *eth.RPCBlock)` <br>`eth.HeaderByHash(hash common.Hash).Returns(header *types.Header)`<br>`eth.HeaderByHash(hash common.Hash).ReturnsRAW(header *eth.RPCHeader)`
| `eth_getBlockByNumber`      | `eth.BlockByNumber(number *big.Int).Returns(block *types.Block)`<br>`eth.BlockByNumber(number *big.Int).ReturnsRAW(block *eth.RPCBlock)`<br>`eth.HeaderByNumber(number *big.Int).Returns(header *types.Header)`<br>`eth.HeaderByNumber(number *big.Int).ReturnsRAW(header *eth.RAWHeader)`
| `eth_getCode`               | `eth.Code(addr common.Address, blockNumber *big.Int).Returns(code *[]byte)`
| `eth_getLogs`               | `eth.Logs(q ethereum.FilterQuery).Returns(logs *[]types.Log)`
| `eth_getStorageAt`          | `eth.StorageAt(addr common.Address, slot common.Hash, blockNumber *big.Int).Returns(storage *common.Hash)`
| `eth_getTransactionByHash`  | `eth.TransactionByHash(hash common.Hash).Returns(tx *types.Transaction)`<br>`eth.TransactionByHash(hash common.Hash).ReturnsRAW(tx *eth.RPCTransaction)`
| `eth_getTransactionCount`   | `eth.Nonce(addr common.Address, blockNumber *big.Int).Returns(nonce *uint64)`
| `eth_getTransactionReceipt` | `eth.TransactionReceipt(hash common.Hash).Returns(receipt *types.Receipt)`<br>`eth.TransactionReceipt(hash common.Hash).ReturnsRAW(receipt *eth.RPCReceipt)`
| `eth_sendRawTransaction`    | `eth.SendTransaction(tx *types.Transaction).Returns(hash *common.Hash)`<br>`eth.SendRawTransaction(rawTx []byte).Returns(hash *common.Hash)`

### `debug`

| Method                   | Go Code
| :----------------------- | :-------
| `debug_traceCall`        | TODO <!-- `debug.TraceCall(msg ethereum.CallMsg).Returns(blockNumber *big.Int)` -->
| `debug_traceTransaction` | TODO <!--`debug.TraceTransaction(hash common.Hash).Returns(blockNumber *big.Int)` -->

### `web3`

| Method               | Go Code
| :------------------- | :-------
| `web3_clientVersion` | `web3.ClientVersion().Returns(clientVersion *string)`

### Third Party RPC Method Packages

| Package                                                                  | Description
| :----------------------------------------------------------------------- | :-----------
| [github.com/lmittmann/flashbots](https://github.com/lmittmann/flashbots) | Package `flashbots` implements RPC API bindings for the Flashbots relay and mev-geth.
