package w3_test

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/lmittmann/w3"
	"github.com/lmittmann/w3/module/eth"
	"github.com/lmittmann/w3/w3types"
	"github.com/lmittmann/w3/w3vm"
	"golang.org/x/time/rate"
)

var (
	funcName      = w3.MustNewFunc("name()", "string")
	funcSymbol    = w3.MustNewFunc("symbol()", "string")
	funcDecimals  = w3.MustNewFunc("decimals()", "uint8")
	funcBalanceOf = w3.MustNewFunc("balanceOf(address)", "uint256")

	addrA = common.Address{0x0a}
	addrB = common.Address{0x0b}

	prvA *ecdsa.PrivateKey // dummy private key for addrA

	addrWETH = w3.A("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2")
	addrDAI  = w3.A("0x6B175474E89094C44Da98b954EedeAC495271d0F")

	client = w3.MustDial("https://ethereum-rpc.publicnode.com")
)

// Call the name, symbol, decimals, and balanceOf functions of the Wrapped Ether
// in a single batch.
func ExampleClient_batchCallFunc() {
	blockNumber := big.NewInt(20_000_000)

	var (
		name, symbol string
		decimals     uint8
		balance      big.Int
	)
	if err := client.Call(
		eth.CallFunc(addrWETH, funcName).Returns(&name),
		eth.CallFunc(addrWETH, funcSymbol).Returns(&symbol),
		eth.CallFunc(addrWETH, funcDecimals).Returns(&decimals),
		eth.CallFunc(addrWETH, funcBalanceOf, addrWETH).AtBlock(blockNumber).Returns(&balance),
	); err != nil {
		// ...
	}

	fmt.Printf("%s's own balance: %s %s\n", name, w3.FromWei(&balance, decimals), symbol)
	// Output:
	// Wrapped Ether's own balance: 748.980125465356473638 WETH
}

// Call the Uniswap V3 Quoter for quotes on swapping 100 WETH for DAI in pools
// of all fee tiers in a single batch.
func ExampleClient_batchCallFuncUniswapQuoter() {
	blockNumber := big.NewInt(20_000_000)

	var (
		addrUniswapV3Quoter = w3.A("0xb27308f9F90D607463bb33eA1BeBb41C27CE5AB6")
		addrTokenIn         = addrWETH
		addrTokenOut        = addrDAI

		funcQuote = w3.MustNewFunc(`quoteExactInputSingle(
			address tokenIn,
			address tokenOut,
			uint24 fee,
			uint256 amountIn,
			uint160 sqrtPriceLimitX96clear
		)`, "uint256 amountOut")
	)

	var (
		amountIn       = w3.I("100 ether")
		amountOut100   *big.Int
		amountOut500   *big.Int
		amountOut3000  *big.Int
		amountOut10000 *big.Int
	)
	if err := client.Call(
		eth.CallFunc(addrUniswapV3Quoter, funcQuote, addrTokenIn, addrTokenOut, big.NewInt(100), amountIn, w3.Big0).AtBlock(blockNumber).Returns(&amountOut100),
		eth.CallFunc(addrUniswapV3Quoter, funcQuote, addrTokenIn, addrTokenOut, big.NewInt(500), amountIn, w3.Big0).AtBlock(blockNumber).Returns(&amountOut500),
		eth.CallFunc(addrUniswapV3Quoter, funcQuote, addrTokenIn, addrTokenOut, big.NewInt(3000), amountIn, w3.Big0).AtBlock(blockNumber).Returns(&amountOut3000),
		eth.CallFunc(addrUniswapV3Quoter, funcQuote, addrTokenIn, addrTokenOut, big.NewInt(10000), amountIn, w3.Big0).AtBlock(blockNumber).Returns(&amountOut10000),
	); err != nil {
		// ...
	}
	fmt.Println("Swap 100 WETH for DAI:")
	fmt.Printf("Pool with 0.01%% fee: %s DAI\n", w3.FromWei(amountOut100, 18))
	fmt.Printf("Pool with 0.05%% fee: %s DAI\n", w3.FromWei(amountOut500, 18))
	fmt.Printf("Pool with  0.3%% fee: %s DAI\n", w3.FromWei(amountOut3000, 18))
	fmt.Printf("Pool with    1%% fee: %s DAI\n", w3.FromWei(amountOut10000, 18))
	// Output:
	// Swap 100 WETH for DAI:
	// Pool with 0.01% fee: 0.840975419471618588 DAI
	// Pool with 0.05% fee: 371877.453117609415215338 DAI
	// Pool with  0.3% fee: 378532.856217317782434539 DAI
	// Pool with    1% fee: 3447.634026125332130689 DAI
}

// Fetch the nonce and balance of an EOA in a single batch.
func ExampleClient_batchEOAState() {
	var (
		nonce   uint64
		balance *big.Int
	)
	if err := client.Call(
		eth.Nonce(addrA, nil).Returns(&nonce),
		eth.Balance(addrA, nil).Returns(&balance),
	); err != nil {
		// ...
	}

	fmt.Printf("Nonce: %d\nBalance: %d\n", nonce, balance)
}

// Fetch a transaction and its receipt in a single batch.
func ExampleClient_batchTxDetails() {
	txHash := w3.H("0xc31d7e7e85cab1d38ce1b8ac17e821ccd47dbde00f9d57f2bd8613bff9428396")

	var (
		tx      *types.Transaction
		receipt *types.Receipt
	)
	if err := client.Call(
		eth.Tx(txHash).Returns(&tx),
		eth.TxReceipt(txHash).Returns(&receipt),
	); err != nil {
		// ...
	}

	fmt.Printf("Tx: %#v\nReceipt: %#v\n", tx, receipt)
}

// Fetch 1000 blocks in batches.
func ExampleClient_batchBlocks() {
	const (
		startBlock = 20_000_000
		nBlocks    = 1000
		batchSize  = 100
	)

	blocks := make([]*types.Block, nBlocks)
	calls := make([]w3types.RPCCaller, batchSize)
	for i := 0; i < nBlocks; i += batchSize {
		for j := range batchSize {
			blockNumber := new(big.Int).SetUint64(uint64(startBlock + i + j))
			calls[j] = eth.BlockByNumber(blockNumber).Returns(&blocks[i+j])
		}
		if err := client.Call(calls...); err != nil {
			// ...
		}
		fmt.Printf("Fetched %d blocks\n", i+batchSize)
	}
}

// Handle errors of individual calls in a batch.
func ExampleClient_batchHandleError() {
	tokens := []common.Address{addrWETH, addrA, addrB}
	symbols := make([]string, len(tokens))

	// build rpc calls
	calls := make([]w3types.RPCCaller, len(tokens))
	for i, token := range tokens {
		calls[i] = eth.CallFunc(token, funcSymbol).Returns(&symbols[i])
	}

	var batchErr w3.CallErrors
	if err := client.Call(calls...); errors.As(err, &batchErr) {
	} else if err != nil {
		// all calls failed
	}

	for i, symbol := range symbols {
		if len(batchErr) > 0 && batchErr[i] != nil {
			symbol = "call failed"
		}
		fmt.Printf("%s: %s\n", tokens[i], symbol)
	}
	// Output:
	// 0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2: WETH
	// 0x0a00000000000000000000000000000000000000: call failed
	// 0x0B00000000000000000000000000000000000000: call failed
}

// Fetch the token balance of an address.
func ExampleClient_callFunc() {
	var balance *big.Int
	if err := client.Call(
		eth.CallFunc(addrWETH, funcBalanceOf, addrA).Returns(&balance),
	); err != nil {
		// ...
	}

	fmt.Printf("Balance: %s WETH\n", w3.FromWei(balance, 18))
	// Output:
	// Balance: 0 WETH
}

// Fetch the token balance of an address, with state override.
func ExampleClient_callFuncWithStateOverride() {
	var balance *big.Int
	if err := client.Call(
		eth.CallFunc(addrWETH, funcBalanceOf, addrA).Overrides(w3types.State{
			addrWETH: {Storage: w3types.Storage{
				w3vm.WETHBalanceSlot(addrA): common.BigToHash(w3.I("100 ether")),
			}},
		}).Returns(&balance),
	); err != nil {
		// ...
	}

	fmt.Printf("Balance: %s WETH\n", w3.FromWei(balance, 18))
	// Output:
	// Balance: 100 WETH
}

// Send Ether transfer.
func ExampleClient_sendETHTransfer() {
	var (
		nonce    uint64
		gasPrice *big.Int
	)
	if err := client.Call(
		eth.Nonce(addrA, nil).Returns(&nonce),
		eth.GasPrice().Returns(&gasPrice),
	); err != nil {
		// ...
	}

	signer := types.LatestSigner(params.MainnetChainConfig)
	tx := types.MustSignNewTx(prvA, signer, &types.LegacyTx{
		Nonce:    nonce,
		Gas:      21_000,
		GasPrice: gasPrice,
		To:       &addrB,
		Value:    w3.I("1 ether"),
	})

	var txHash common.Hash
	if err := client.Call(eth.SendTx(tx).Returns(&txHash)); err != nil {
		// ...
	}

	fmt.Printf("Sent tx: %s\n", txHash)
}

// Send ERC20 token transfer (Wrapped Ether).
func ExampleClient_sendTokenTransfer() {
	var (
		nonce    uint64
		gasPrice *big.Int
	)
	if err := client.Call(
		eth.Nonce(addrA, nil).Returns(&nonce),
		eth.GasPrice().Returns(&gasPrice),
	); err != nil {
		// ...
	}

	funcTransfer := w3.MustNewFunc("transfer(address receiver, uint256 amount)", "bool")
	data, err := funcTransfer.EncodeArgs(addrB, w3.I("1 ether"))
	if err != nil {
		// ...
	}

	signer := types.LatestSigner(params.MainnetChainConfig)
	tx := types.MustSignNewTx(prvA, signer, &types.LegacyTx{
		Nonce:    nonce,
		Gas:      100_000,
		GasPrice: gasPrice,
		To:       &addrWETH,
		Data:     data,
	})

	var txHash common.Hash
	if err := client.Call(eth.SendTx(tx).Returns(&txHash)); err != nil {
		// ...
	}

	fmt.Printf("Sent tx: %s\n", txHash)
}

// Subscribe to pending transactions.
func ExampleClient_subscribeToPendingTransactions() {
	client, err := w3.Dial("wss://mainnet.gateway.tenderly.co")
	if err != nil {
		// ...
	}
	defer client.Close()

	pendingTxCh := make(chan *types.Transaction)
	sub, err := client.Subscribe(eth.PendingTransactions(pendingTxCh))
	if err != nil {
		// ...
	}

	for {
		select {
		case tx := <-pendingTxCh:
			fmt.Printf("New pending tx: %s\n", tx.Hash())
		case err := <-sub.Err():
			fmt.Printf("Subscription error: %v\n", err)
			return
		}
	}
}

// Rate Limit the number of requests to 10 per second, with bursts of up to 20
// requests.
func ExampleClient_rateLimitByRequest() {
	client, err := w3.Dial("https://ethereum-rpc.publicnode.com",
		w3.WithRateLimiter(rate.NewLimiter(rate.Every(time.Second/10), 20), nil),
	)
	if err != nil {
		// ...
	}
	defer client.Close()
}

// Rate Limit the number of requests to 300 compute units (CUs) per second, with
// bursts of up to 300 CUs.
// An individual CU can be charged per RPC method call.
func ExampleClient_rateLimitByComputeUnits() {
	// cu returns the CU cost for all method calls in a batch.
	cu := func(methods []string) (cost int) {
		for _, method := range methods {
			switch method {
			case "eth_blockNumber":
				cost += 5
			case "eth_getBalance",
				"eth_getBlockByNumber",
				"eth_getCode",
				"eth_getStorageAt",
				"eth_getTransactionByHash",
				"eth_getTransactionReceipt":
				cost += 15
			case "eth_call":
				cost += 20
			case "eth_getTransactionCount":
				cost += 25
			default:
				panic(fmt.Sprintf("unknown costs for %q", method))
			}
		}
		return cost
	}

	client, err := w3.Dial("https://ethereum-rpc.publicnode.com",
		w3.WithRateLimiter(rate.NewLimiter(rate.Every(time.Second/300), 300), cu),
	)
	if err != nil {
		// ...
	}
	defer client.Close()
}

// ABI bindings for the ERC20 functions.
func ExampleFunc_erc20() {
	var (
		funcTotalSupply  = w3.MustNewFunc("totalSupply()", "uint256")
		funcBalanceOf    = w3.MustNewFunc("balanceOf(address)", "uint256")
		funcTransfer     = w3.MustNewFunc("transfer(address to, uint256 amount)", "bool")
		funcAllowance    = w3.MustNewFunc("allowance(address owner, address spender)", "uint256")
		funcApprove      = w3.MustNewFunc("approve(address spender, uint256 amount)", "bool")
		funcTransferFrom = w3.MustNewFunc("transferFrom(address from, address to, uint256 amount)", "bool")
	)
	_ = funcTotalSupply
	_ = funcBalanceOf
	_ = funcTransfer
	_ = funcAllowance
	_ = funcApprove
	_ = funcTransferFrom
}

// Encode and decode the arguments of the balanceOf function.
func ExampleFunc_balanceOf() {
	// encode
	input, err := funcBalanceOf.EncodeArgs(addrA)
	if err != nil {
		// ...
	}
	fmt.Printf("encoded: 0x%x\n", input)

	// decode
	var who common.Address
	if err := funcBalanceOf.DecodeArgs(input, &who); err != nil {
		// ...
	}
	fmt.Printf("decoded: balanceOf(%s)\n", who)
	// Output:
	// encoded: 0x70a082310000000000000000000000000a00000000000000000000000000000000000000
	// decoded: balanceOf(0x0a00000000000000000000000000000000000000)
}

// ABI bindings for the Uniswap v4 swap function.
func ExampleFunc_uniswapV4Swap() {
	// ABI binding for the PoolKey struct.
	type PoolKey struct {
		Currency0   common.Address
		Currency1   common.Address
		Fee         *big.Int `abitype:"uint24"`
		TickSpacing *big.Int `abitype:"int24"`
		Hooks       common.Address
	}

	// ABI binding for the SwapParams struct.
	type SwapParams struct {
		ZeroForOne        bool
		AmountSpecified   *big.Int `abitype:"int256"`
		SqrtPriceLimitX96 *big.Int `abitype:"uint160"`
	}

	funcSwap := w3.MustNewFunc(`swap(PoolKey key, SwapParams params, bytes hookData)`, "int256 delta",
		PoolKey{}, SwapParams{},
	)

	// encode
	input, _ := funcSwap.EncodeArgs(
		&PoolKey{
			Currency0:   addrWETH,
			Currency1:   addrDAI,
			Fee:         big.NewInt(0),
			TickSpacing: big.NewInt(0),
		},
		&SwapParams{
			ZeroForOne:        false,
			AmountSpecified:   big.NewInt(0),
			SqrtPriceLimitX96: big.NewInt(0),
		},
		[]byte{},
	)
	fmt.Printf("encoded: 0x%x\n", input)
	// Output:
	// encoded: 0xf3cd914c000000000000000000000000c02aaa39b223fe8d0a0e5c4f27ead9083c756cc20000000000000000000000006b175474e89094c44da98b954eedeac495271d0f00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001200000000000000000000000000000000000000000000000000000000000000000
}

func ExampleFunc_DecodeReturns_getReserves() {
	funcGetReserves := w3.MustNewFunc("getReserves()", "uint112 reserve0, uint112 reserve1, uint32 blockTimestampLast")
	output := w3.B(
		"0x00000000000000000000000000000000000000000000003635c9adc5dea00000",
		"0x0000000000000000000000000000000000000000000000a2a15d09519be00000",
		"0x0000000000000000000000000000000000000000000000000000000064373057",
	)

	var (
		reserve0, reserve1 *big.Int
		blockTimestampLast uint32
	)
	if err := funcGetReserves.DecodeReturns(output, &reserve0, &reserve1, &blockTimestampLast); err != nil {
		// ...
	}
	fmt.Println("Reserve0:", reserve0)
	fmt.Println("Reserve1:", reserve1)
	fmt.Println("BlockTimestampLast:", blockTimestampLast)
	// Output:
	// Reserve0: 1000000000000000000000
	// Reserve1: 3000000000000000000000
	// BlockTimestampLast: 1681338455
}

func ExampleEvent_decodeTransferEvent() {
	var (
		eventTransfer = w3.MustNewEvent("Transfer(address indexed from, address indexed to, uint256 value)")
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
