package eth_test

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/lmittmann/w3"
	"github.com/lmittmann/w3/module/eth"
	"github.com/lmittmann/w3/rpctest"
)

func TestSendTransaction(t *testing.T) {
	t.Parallel()

	srv := rpctest.NewFileServer(t, "testdata/send_raw_transaction.golden")
	defer srv.Close()

	client := w3.MustDial(srv.URL())
	defer client.Close()

	var (
		tx = types.NewTx(&types.DynamicFeeTx{
			ChainID:   big.NewInt(1),
			Nonce:     261,
			GasTipCap: w3.I("30.38 gwei"),
			GasFeeCap: w3.I("32.38 gwei"),
			Gas:       47238,
			To:        w3.APtr("0x491D6b7D6822d5d4BC88a1264E1b47791Fd8E904"),
			Data:      w3.B("0x095ea7b30000000000000000000000007645eec8bb51862a5aa855c40971b2877dae81afffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"),
			V:         w3.I("0x1"),
			R:         w3.I("0x416470241b7db89c67526881b6fd8e145416b294a35bf4280d3079f6308c2d11"),
			S:         w3.I("0x2c0af1cc55c22c0bab79ec083801da63253453156356fcd4291f50d0f425a0ee"),
		})
		hash     common.Hash
		wantHash = w3.H("0xed382cb554ad10e94921d263a56c670669d6c380bbdacdbf96fed625b7132a1d")
	)

	if err := client.Call(eth.SendTransaction(tx).Returns(&hash)); err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	if wantHash != hash {
		t.Fatalf("want %v, got %v", wantHash, hash)
	}
}
