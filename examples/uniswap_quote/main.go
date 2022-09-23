/*
uniswap_quote prints the UniSwap V3 exchange rate to swap amontIn of tokenIn for
tokenOut.

Usage:

	uniswap_quote [flags]

Flags:

	-amountIn string
		Amount of tokenIn to exchange for tokenOut (default "1 ether")
	-tokenIn string
		Token in address (default "0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2")
	-tokenOut string
		Token out address (default "0x6B175474E89094C44Da98b954EedeAC495271d0F")
	-h, --help
		help for uniswap_quote
*/
package main

import (
	"flag"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/lmittmann/w3"
	"github.com/lmittmann/w3/module/eth"
	"github.com/lmittmann/w3/w3types"
)

var (
	addrUniV3Quoter = w3.A("0xb27308f9F90D607463bb33eA1BeBb41C27CE5AB6")

	funcQuoteExactInputSingle = w3.MustNewFunc("quoteExactInputSingle(address tokenIn, address tokenOut, uint24 fee, uint256 amountIn, uint160 sqrtPriceLimitX96)", "uint256 amountOut")
	funcName                  = w3.MustNewFunc("name()", "string")
	funcSymbol                = w3.MustNewFunc("symbol()", "string")
	funcDecimals              = w3.MustNewFunc("decimals()", "uint8")

	// flags
	addrTokenIn  common.Address
	addrTokenOut common.Address
	amountIn     big.Int
)

func main() {
	// parse flags
	flag.TextVar(&amountIn, "amountIn", w3.I("1 ether"), "Token address")
	flag.TextVar(&addrTokenIn, "tokenIn", w3.A("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2"), "Token in")
	flag.TextVar(&addrTokenOut, "tokenOut", w3.A("0x6B175474E89094C44Da98b954EedeAC495271d0F"), "Token out")
	flag.Usage = func() {
		fmt.Println("uniswap_quote prints the UniSwap V3 exchange rate to swap amontIn of tokenIn for tokenOut.")
		flag.PrintDefaults()
	}
	flag.Parse()

	// connect to RPC endpoint
	client := w3.MustDial("https://rpc.ankr.com/eth")
	defer client.Close()

	// fetch token details
	var (
		tokenInName      string
		tokenInSymbol    string
		tokenInDecimals  uint8
		tokenOutName     string
		tokenOutSymbol   string
		tokenOutDecimals uint8
	)
	if err := client.Call(
		eth.CallFunc(funcName, addrTokenIn).Returns(&tokenInName),
		eth.CallFunc(funcSymbol, addrTokenIn).Returns(&tokenInSymbol),
		eth.CallFunc(funcDecimals, addrTokenIn).Returns(&tokenInDecimals),
		eth.CallFunc(funcName, addrTokenOut).Returns(&tokenOutName),
		eth.CallFunc(funcSymbol, addrTokenOut).Returns(&tokenOutSymbol),
		eth.CallFunc(funcDecimals, addrTokenOut).Returns(&tokenOutDecimals),
	); err != nil {
		fmt.Printf("Failed to fetch token details: %v\n", err)
		return
	}

	// fetch quotes
	var (
		fees       = []*big.Int{big.NewInt(100), big.NewInt(500), big.NewInt(3000), big.NewInt(10000)}
		calls      = make([]w3types.Caller, len(fees))
		amountsOut = make([]big.Int, len(fees))
	)
	for i, fee := range fees {
		calls[i] = eth.CallFunc(funcQuoteExactInputSingle, addrUniV3Quoter, addrTokenIn, addrTokenOut, fee, &amountIn, w3.Big0).Returns(&amountsOut[i])
	}
	err := client.Call(calls...)
	callErrs, ok := err.(w3.CallErrors)
	if err != nil && !ok {
		fmt.Printf("Failed to fetch quotes: %v\n", err)
		return

	}

	// print quotes
	fmt.Printf("Exchange %q for %q\n", tokenInName, tokenOutName)
	fmt.Printf("Amount in:\n  %s %s\n", w3.FromWei(&amountIn, tokenInDecimals), tokenInSymbol)
	fmt.Printf("Amount out:\n")
	for i, fee := range fees {
		if ok && callErrs[i] != nil {
			fmt.Printf("  Pool (fee=%5v): Pool does not exist\n", fee)
			continue
		}
		fmt.Printf("  Pool (fee=%5v): %s %s\n", fee, w3.FromWei(&amountsOut[i], tokenOutDecimals), tokenOutSymbol)
	}
}
