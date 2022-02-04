# w3

[![Go Reference](https://pkg.go.dev/badge/github.com/lmittmann/w3.svg)](https://pkg.go.dev/github.com/lmittmann/w3)
[![Go Report Card](https://goreportcard.com/badge/github.com/lmittmann/w3)](https://goreportcard.com/report/github.com/lmittmann/w3)


Package `w3` implements a modular and fast Ethereum JSON RPC client with
first-class ABI support.

* **Modular** API allows to create custom RPC method integrations that can be
  used alongside the methods implemented by the package.
* **Batch request** support significantly reduces the duration of requests to
  both remote and local endpoints.
* **ABI** bindings are specified for individual functions using Solidity syntax.
  No need for `abigen` and ABI JSON files.

`w3` is closely linked to [go-ethereum](https://github.com/ethereum/go-ethereum)
and uses a variety of its types, such as [`common.Address`](https://pkg.go.dev/github.com/ethereum/go-ethereum/common#Address)
or [`types.Transaction`](https://pkg.go.dev/github.com/ethereum/go-ethereum/core/types#Transaction).


## Install

```
go get github.com/lmittmann/w3@latest
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
	eth.Nonce(addr).Returns(&nonce),
	eth.Balance(addr).Returns(&balance),
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
