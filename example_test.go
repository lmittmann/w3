package w3_test

import (
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/lmittmann/w3"
	"github.com/lmittmann/w3/module/eth"
)

func ExampleDial() {
	client, err := w3.Dial("https://cloudflare-eth.com")
	if err != nil {
		log.Fatalf("Failed to connect to RPC endpoint: %v", err)
	}
	defer client.Close()
}

func ExampleI() {
	fmt.Printf("%v wei\n", w3.I("0x2b98d99b09e3c000"))
	fmt.Printf("%v wei\n", w3.I("3141500000000000000"))
	fmt.Printf("%v wei\n", w3.I("3.1415 ether"))
	fmt.Printf("%v wei\n", w3.I("31.415 gwei"))

	// Output:
	// 3141500000000000000 wei
	// 3141500000000000000 wei
	// 3141500000000000000 wei
	// 31415000000 wei
}

func ExampleNewFunc() {
	// ABI binding to the balanceOf function of an ERC20 Token.
	funcBalanceOf, _ := w3.NewFunc("balanceOf(address)", "uint256")

	// Optionally names can be specified for function arguments. This is
	// especially useful for more complex functions with many arguments.
	funcBalanceOf, _ = w3.NewFunc("balanceOf(address who)", "uint256 amount")

	// ABI-encode the functions args.
	input, _ := funcBalanceOf.EncodeArgs(w3.A("0xAb5801a7D398351b8bE11C439e05C5B3259aeC9B"))
	fmt.Printf("balanceOf input: 0x%x\n", input)

	// ABI-decode the functions args from a given input.
	var (
		who common.Address
	)
	funcBalanceOf.DecodeArgs(input, &who)
	fmt.Printf("balanceOf args: %v\n", who)

	// ABI-decode the functions output.
	var (
		output = w3.B("0x000000000000000000000000000000000000000000000000000000000000c0fe")
		amount = new(big.Int)
	)
	funcBalanceOf.DecodeReturns(output, amount)
	fmt.Printf("balanceOf returns: %v\n", amount)

	// Output:
	// balanceOf input: 0x70a08231000000000000000000000000ab5801a7d398351b8be11c439e05c5b3259aec9b
	// balanceOf args: 0xAb5801a7D398351b8bE11C439e05C5B3259aeC9B
	// balanceOf returns: 49406
}

func ExampleClient_Call() {
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

	fmt.Printf("Combined balance: %v wei",
		new(big.Int).Add(&ethBalance, &weth9Balance),
	)
}
