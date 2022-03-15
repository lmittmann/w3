package w3_test

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/lmittmann/w3"
	"github.com/lmittmann/w3/module/eth"
)

func ExampleDial() {
	client, err := w3.Dial("https://cloudflare-eth.com")
	if err != nil {
		fmt.Printf("Failed to connect to RPC endpoint: %v\n", err)
		return
	}
	defer client.Close()
}

func ExampleMustDial() {
	client := w3.MustDial("https://cloudflare-eth.com")
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

func ExampleFromWei() {
	wei := big.NewInt(1_230000000_000000000)
	fmt.Printf("%s Ether\n", w3.FromWei(wei, 18))
	// Output:
	// 1.23 Ether
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

		// Declare a Smart Contract function using Solidity syntax,
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
		fmt.Printf("Request failed: %v\n", err)
		return
	}

	fmt.Printf("Combined balance: %v wei",
		new(big.Int).Add(&ethBalance, &weth9Balance),
	)
}

func ExampleClient_Call_nonceAndBalance() {
	client := w3.MustDial("https://cloudflare-eth.com")
	defer client.Close()

	var (
		addr = w3.A("0x000000000000000000000000000000000000c0Fe")

		nonce   uint64
		balance big.Int
	)

	if err := client.Call(
		eth.Nonce(addr).Returns(&nonce),
		eth.Balance(addr).Returns(&balance),
	); err != nil {
		fmt.Printf("Request failed: %v\n", err)
		return
	}

	fmt.Printf("%s: Nonce: %d, Balance: ♦%s\n", addr, nonce, w3.FromWei(&balance, 18))
}

func ExampleEvent_DecodeArgs() {
	var (
		eventTransfer = w3.MustNewEvent("Transfer(address from, address to, uint256 value)")
		log           = &types.Log{
			Address: w3.A("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2"),
			Topics: []common.Hash{
				w3.H("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"),
				w3.H("0x000000000000000000000000000000000000000000000000000000000000c0fe"),
				w3.H("0x000000000000000000000000000000000000000000000000000000000000dead"),
			},
			Data: w3.B("0x0000000000000000000000000000000000000000000000001111d67bb1bb0000"),
		}

		from  common.Address
		to    common.Address
		value big.Int
	)

	if err := eventTransfer.DecodeArgs(log, &from, &to, &value); err != nil {
		fmt.Printf("Failed to decode event log: %v\n", err)
		return
	}
	fmt.Printf("Transferred %s WETH9 from %s to %s", w3.FromWei(&value, 18), from, to)
	// Output:
	// Transferred 1.23 WETH9 from 0x000000000000000000000000000000000000c0Fe to 0x000000000000000000000000000000000000dEaD
}
