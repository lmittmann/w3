package w3_test

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"math/big"
	"strconv"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/google/go-cmp/cmp"
	"github.com/lmittmann/w3"
	"github.com/lmittmann/w3/internal"
	"github.com/lmittmann/w3/module/eth"
	"github.com/lmittmann/w3/rpctest"
	"github.com/lmittmann/w3/w3types"
	"golang.org/x/time/rate"
)

var (
	benchRPC = flag.String("benchRPC", "", "RPC endpoint for benchmark")

	jsonCalls1 = `> {"jsonrpc":"2.0","id":1}` + "\n" +
		`< {"jsonrpc":"2.0","id":1,"result":"0x1"}`
	jsonCalls2 = `> [{"jsonrpc":"2.0","id":1},{"jsonrpc":"2.0","id":2}]` + "\n" +
		`< [{"jsonrpc":"2.0","id":1,"result":"0x1"},{"jsonrpc":"2.0","id":2,"result":"0x1"}]`
)

func ExampleClient() {
	addr := w3.A("0x0000000000000000000000000000000000000000")

	// 1. Connect to an RPC endpoint
	client, err := w3.Dial("https://rpc.ankr.com/eth")
	if err != nil {
		// handle error
	}
	defer client.Close()

	// 2. Make a batch request
	var (
		balance big.Int
		nonce   uint64
	)
	if err := client.Call(
		eth.Balance(addr, nil).Returns(&balance),
		eth.Nonce(addr, nil).Returns(&nonce),
	); err != nil {
		// handle error
	}

	fmt.Printf("balance: %s\nnonce: %d\n", w3.FromWei(&balance, 18), nonce)
}

func ExampleClient_Call_balanceOf() {
	// Connect to RPC endpoint (or panic on error) and
	// close the connection when you are done.
	client := w3.MustDial("https://rpc.ankr.com/eth")
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
		eth.Balance(addr, nil).Returns(&ethBalance),
		eth.CallFunc(weth9, balanceOf, addr).Returns(&weth9Balance),
	); err != nil {
		fmt.Printf("Request failed: %v\n", err)
		return
	}

	fmt.Printf("Combined balance: %v wei",
		new(big.Int).Add(&ethBalance, &weth9Balance),
	)
}

func ExampleClient_Call_nonceAndBalance() {
	client := w3.MustDial("https://rpc.ankr.com/eth")
	defer client.Close()

	var (
		addr = w3.A("0x000000000000000000000000000000000000c0Fe")

		nonce   uint64
		balance big.Int
	)

	if err := client.Call(
		eth.Nonce(addr, nil).Returns(&nonce),
		eth.Balance(addr, nil).Returns(&balance),
	); err != nil {
		fmt.Printf("Request failed: %v\n", err)
		return
	}

	fmt.Printf("%s: Nonce: %d, Balance: ♦%s\n", addr, nonce, w3.FromWei(&balance, 18))
}

func ExampleClient_Call_sendERC20transferTx() {
	client := w3.MustDial("https://rpc.ankr.com/eth")
	defer client.Close()

	var (
		weth9     = w3.A("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2")
		receiver  = w3.A("0x000000000000000000000000000000000000c0Fe")
		eoaPrv, _ = crypto.GenerateKey()
	)

	funcTransfer := w3.MustNewFunc("transfer(address receiver, uint256 amount)", "bool")
	input, err := funcTransfer.EncodeArgs(receiver, w3.I("1 ether"))
	if err != nil {
		fmt.Printf("Failed to encode args: %v\n", err)
		return
	}

	signer := types.LatestSigner(params.MainnetChainConfig)
	var txHash common.Hash
	if err := client.Call(
		eth.SendTx(types.MustSignNewTx(eoaPrv, signer, &types.DynamicFeeTx{
			Nonce:     0,
			To:        &weth9,
			Data:      input,
			GasTipCap: w3.I("1 gwei"),
			GasFeeCap: w3.I("100 gwei"),
			Gas:       100_000,
		})).Returns(&txHash),
	); err != nil {
		fmt.Printf("Failed to send tx: %v\n", err)
		return
	}

	fmt.Printf("Sent tx: %s\n", txHash)
}

func ExampleCallErrors() {
	client := w3.MustDial("https://rpc.ankr.com/eth")
	defer client.Close()

	funcSymbol := w3.MustNewFunc("symbol()", "string")

	// list of addresses that might be an ERC20 token
	potentialTokens := []common.Address{
		w3.A("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2"),
		w3.A("0x00000000219ab540356cBB839Cbe05303d7705Fa"),
	}

	// build symbol()-call for each potential ERC20 token
	tokenSymbols := make([]string, len(potentialTokens))
	rpcCalls := make([]w3types.RPCCaller, len(potentialTokens))
	for i, addr := range potentialTokens {
		rpcCalls[i] = eth.CallFunc(addr, funcSymbol).Returns(&tokenSymbols[i])
	}

	// execute batch request
	var errs w3.CallErrors
	if err := client.Call(rpcCalls...); errors.As(err, &errs) {
		// handle call errors
	} else if err != nil {
		// handle other errors
		fmt.Printf("Request failed: %v\n", err)
		return
	}

	for i, addr := range potentialTokens {
		var symbol string
		if errs == nil || errs[i] == nil {
			symbol = tokenSymbols[i]
		} else {
			symbol = fmt.Sprintf("unknown symbol: %v", errs[i].Error())
		}
		fmt.Printf("%s: %s\n", addr, symbol)
	}

	// Output:
	// 0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2: WETH
	// 0x00000000219ab540356cBB839Cbe05303d7705Fa: unknown symbol: execution reverted
}

func ExampleClient_Subscribe() {
	client := w3.MustDial("wss://mainnet.gateway.tenderly.co")
	defer client.Close()

	txCh := make(chan *types.Transaction)
	sub, err := client.Subscribe(eth.PendingTransactions(txCh))
	if err != nil {
		fmt.Printf("Failed to subscribe: %v\n", err)
		return
	}

	for {
		select {
		case tx := <-txCh:
			fmt.Printf("New pending tx: %s\n", tx.Hash())
		case err := <-sub.Err():
			fmt.Printf("Subscription error: %v\n", err)
			return
		}
	}
}

func TestClientCall(t *testing.T) {
	tests := []struct {
		Buf     *bytes.Buffer
		Calls   []w3types.RPCCaller
		WantErr error
	}{
		{
			Buf:   bytes.NewBufferString(jsonCalls1),
			Calls: []w3types.RPCCaller{&testCaller{}},
		},
		{
			Buf:     bytes.NewBufferString(jsonCalls1),
			Calls:   []w3types.RPCCaller{&testCaller{RequestErr: errors.New("err")}},
			WantErr: errors.New("err"),
		},
		{
			Buf:     bytes.NewBufferString(jsonCalls1),
			Calls:   []w3types.RPCCaller{&testCaller{ReturnErr: errors.New("err")}},
			WantErr: errors.New("w3: call failed: err"),
		},
		{
			Buf: bytes.NewBufferString(jsonCalls2),
			Calls: []w3types.RPCCaller{
				&testCaller{RequestErr: errors.New("err")},
				&testCaller{},
			},
			WantErr: errors.New("err"),
		},
		{
			Buf: bytes.NewBufferString(jsonCalls2),
			Calls: []w3types.RPCCaller{
				&testCaller{ReturnErr: errors.New("err")},
				&testCaller{},
			},
			WantErr: errors.New("w3: 1 call failed:\ncall[0]: err"),
		},
		{
			Buf: bytes.NewBufferString(jsonCalls2),
			Calls: []w3types.RPCCaller{
				&testCaller{},
				&testCaller{ReturnErr: errors.New("err")},
			},
			WantErr: errors.New("w3: 1 call failed:\ncall[1]: err"),
		},
		{
			Buf: bytes.NewBufferString(jsonCalls2),
			Calls: []w3types.RPCCaller{
				&testCaller{ReturnErr: errors.New("err")},
				&testCaller{ReturnErr: errors.New("err")},
			},
			WantErr: errors.New("w3: 2 calls failed:\ncall[0]: err\ncall[1]: err"),
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			srv := rpctest.NewServer(t, test.Buf)

			client, err := w3.Dial(srv.URL())
			if err != nil {
				t.Fatalf("Failed to connect to test RPC endpoint: %v", err)
			}

			err = client.Call(test.Calls...)
			if diff := cmp.Diff(test.WantErr, err,
				internal.EquateErrors(),
			); diff != "" {
				t.Fatalf("(-want, +got)\n%s", diff)
			}
		})
	}
}

func TestClientCall_CallErrors(t *testing.T) {
	srv := rpctest.NewServer(t, bytes.NewBufferString(jsonCalls2))

	client, err := w3.Dial(srv.URL())
	if err != nil {
		t.Fatalf("Failed to connect to test RPC endpoint: %v", err)
	}

	err = client.Call(&testCaller{}, &testCaller{ReturnErr: errors.New("err")})
	if err == nil {
		t.Fatal("Want error")
	}
	if !errors.Is(err, w3.CallErrors{}) {
		t.Fatalf("Want w3.CallErrors, got %T", err)
	}
	callErrs := err.(w3.CallErrors)
	if callErrs[0] != nil {
		t.Errorf("callErrs[0]: want <nil>, got %v", callErrs[0])
	}
	if callErrs[1] == nil || callErrs[1].Error() != "err" {
		t.Errorf(`callErrs[1]: want "err", got %v`, callErrs[1])
	}
}

type testCaller struct {
	RequestErr error
	ReturnErr  error
}

func (c *testCaller) CreateRequest() (elem rpc.BatchElem, err error) {
	return rpc.BatchElem{}, c.RequestErr
}

func (c *testCaller) HandleResponse(elem rpc.BatchElem) (err error) {
	return c.ReturnErr
}

func TestClientCall_NilReference(t *testing.T) {
	client := w3.MustDial("https://rpc.ankr.com/eth")
	defer client.Close()

	var block *types.Block
	err := client.Call(
		eth.BlockByNumber(nil).Returns(block),
	)

	want := "w3: cannot return Go value of type *types.Block: value must be passed as a non-nil pointer reference"
	if diff := cmp.Diff(want, err.Error()); diff != "" {
		t.Fatalf("(-want, +got)\n%s", diff)
	}
}

func BenchmarkCall_BalanceNonce(b *testing.B) {
	if *benchRPC == "" {
		b.Skipf("Missing -benchRPC")
	}

	w3Client := w3.MustDial(*benchRPC)
	defer w3Client.Close()

	ethClient, _ := ethclient.Dial(*benchRPC)
	defer ethClient.Close()

	addr := common.Address{}

	b.Run("Batch", func(b *testing.B) {
		var (
			nonce   uint64
			balance big.Int
		)
		for range b.N {
			w3Client.Call(
				eth.Nonce(addr, nil).Returns(&nonce),
				eth.Balance(addr, nil).Returns(&balance),
			)
		}
	})

	b.Run("Sequential", func(b *testing.B) {
		for range b.N {
			ethClient.NonceAt(context.Background(), addr, nil)
			ethClient.BalanceAt(context.Background(), addr, nil)
		}
	})
}

func BenchmarkCall_Balance100(b *testing.B) {
	if *benchRPC == "" {
		b.Skipf("Missing -benchRPC")
	}

	w3Client := w3.MustDial(*benchRPC)
	defer w3Client.Close()

	ethClient, _ := ethclient.Dial(*benchRPC)
	defer ethClient.Close()

	addr100 := make([]common.Address, 100)
	for i := range len(addr100) {
		addr100[i] = common.BigToAddress(big.NewInt(int64(i)))
	}

	b.Run("Batch", func(b *testing.B) {
		var balance big.Int
		for range b.N {
			requests := make([]w3types.RPCCaller, len(addr100))
			for j := range len(requests) {
				requests[j] = eth.Balance(addr100[j], nil).Returns(&balance)
			}
			w3Client.Call(requests...)
		}
	})

	b.Run("Sequential", func(b *testing.B) {
		for range b.N {
			for _, addr := range addr100 {
				ethClient.BalanceAt(context.Background(), addr, nil)
			}
		}
	})
}

func BenchmarkCall_BalanceOf100(b *testing.B) {
	if *benchRPC == "" {
		b.Skipf("Missing -benchRPC")
	}

	w3Client := w3.MustDial(*benchRPC)
	defer w3Client.Close()

	ethClient, _ := ethclient.Dial(*benchRPC)
	defer ethClient.Close()

	addr100 := make([]common.Address, 100)
	for i := range len(addr100) {
		addr100[i] = common.BigToAddress(big.NewInt(int64(i)))
	}

	funcBalanceOf := w3.MustNewFunc("balanceOf(address)", "uint256")
	addrWeth9 := w3.A("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2")

	b.Run("Batch", func(b *testing.B) {
		var balance big.Int
		for range b.N {
			requests := make([]w3types.RPCCaller, len(addr100))
			for j := range len(requests) {
				requests[j] = eth.CallFunc(addrWeth9, funcBalanceOf, addr100[j]).Returns(&balance)
			}
			w3Client.Call(requests...)
		}
	})

	b.Run("Sequential", func(b *testing.B) {
		for range b.N {
			for _, addr := range addr100 {
				input, err := funcBalanceOf.EncodeArgs(addr)
				if err != nil {
					b.Fatalf("Failed to encode args: %v", err)
				}
				ethClient.CallContract(context.Background(), ethereum.CallMsg{
					To:   &addrWeth9,
					Data: input,
				}, nil)
			}
		}
	})
}

func BenchmarkCall_Block100(b *testing.B) {
	if *benchRPC == "" {
		b.Skipf("Missing -benchRPC")
	}

	w3Client := w3.MustDial(*benchRPC)
	defer w3Client.Close()

	ethClient, _ := ethclient.Dial(*benchRPC)
	defer ethClient.Close()

	block100 := make([]*big.Int, 100)
	for i := range len(block100) {
		block100[i] = big.NewInt(int64(14_000_000 + i))
	}

	b.Run("Batch", func(b *testing.B) {
		var block types.Block
		for range b.N {
			requests := make([]w3types.RPCCaller, len(block100))
			for j := range len(requests) {
				requests[j] = eth.BlockByNumber(block100[j]).Returns(&block)
			}
			w3Client.Call(requests...)
		}
	})

	b.Run("Sequential", func(b *testing.B) {
		for range b.N {
			for _, number := range block100 {
				ethClient.BlockByNumber(context.Background(), number)
			}
		}
	})
}

func ExampleWithRateLimiter() {
	// Limit the client to 30 requests per second and allow bursts of up to
	// 100 requests.
	client := w3.MustDial("https://rpc.ankr.com/eth",
		w3.WithRateLimiter(rate.NewLimiter(rate.Every(time.Second/30), 100), nil),
	)
	defer client.Close()
}

func ExampleWithRateLimiter_costFunc() {
	// Limit the client to 30 calls per second and allow bursts of up to
	// 100 calls using a cost function. Batch requests have an additional charge.
	client := w3.MustDial("https://rpc.ankr.com/eth",
		w3.WithRateLimiter(rate.NewLimiter(rate.Every(time.Second/30), 100),
			func(methods []string) (cost int) {
				cost = len(methods) // charge 1 CU per call
				if len(methods) > 1 {
					cost += 1 // charge 1 CU extra for the batch itself
				}
				return cost
			},
		))
	defer client.Close()
}
