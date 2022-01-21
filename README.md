# w3

[![Go Reference](https://pkg.go.dev/badge/github.com/lmittmann/w3.svg)](https://pkg.go.dev/github.com/lmittmann/w3)
[![Go Report Card](https://goreportcard.com/badge/github.com/lmittmann/w3)](https://goreportcard.com/report/github.com/lmittmann/w3)


Package **w3** implements a modular and fast Ethereum JSON RPC client with
first-class ABI support.

* **Modular** API allows to create custom RPC method integrations that can be
  used alongside the methods implemented by the package.
* **Batch request** support significantly reduces the duration of requests to
  both remote and local endpoints.
* **ABI** bindings are specified for individual functions with Solidity syntax.
  No need for `abigen` and ABI JSON files.


## Install

```
go get github.com/lmittmann/w3@latest
```


## Getting Started

```go
// Connect to RPC endpoint (or panic on error) and
// close the connection when you are done.
client := w3.MustDial("https://cloudflare-eth.com")
defer client.Close()

var (
	addr  = w3.A("0x000000000000000000000000000000000000dEaD")
	weth9 = w3.A("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2")

	// Declare a Smart Contract function with Solidity syntax,
	// no "abigen" and ABI JSON file needed.
	balanceOf = w3.MustNewFunc("balanceOf(address)", "uint256")

	// Declare variables for the RPC responses.
	ethBalance   big.Int
	weth9Balance big.Int
)

// Do batch request (both RPC requests are send in the same
// HTTP request).
if err := client.Call(
	eth.Balance(addr).Returns(&ethBalance),
	eth.CallFunc(balanceOf, weth9, addr).Returns(&weth9Balance),
); err != nil {
	fmt.Printf("Requst failed: %v\n", err)
	return
}

fmt.Printf("Combined balance: %v wei\n",
	new(big.Int).Add(&ethBalance, &weth9Balance),
)
```
