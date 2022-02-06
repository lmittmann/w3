package eth_test

import (
	"testing"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/google/go-cmp/cmp"
	"github.com/lmittmann/w3"
	"github.com/lmittmann/w3/module/eth"
	"github.com/lmittmann/w3/rpctest"
)

func TestLogs(t *testing.T) {
	t.Parallel()

	srv := rpctest.NewFileServer(t, "testdata/get_logs.golden")
	defer srv.Close()

	client := w3.MustDial(srv.URL())
	defer client.Close()

	var (
		filterQuery = ethereum.FilterQuery{
			FromBlock: w3.I("10000000"),
			ToBlock:   w3.I("10010000"),
			Topics:    [][]common.Hash{{w3.H("0x0d3648bd0f6ba80134a33ba9275ac585d9d315f0ad8355cddefde31afa28d0e9")}}}

		logs     []types.Log
		wantLogs = []types.Log{
			{
				Address: w3.A("0x5C69bEe701ef814a2B6a3EDD4B1652CB9cc5aA6f"),
				Topics: []common.Hash{
					w3.H("0x0d3648bd0f6ba80134a33ba9275ac585d9d315f0ad8355cddefde31afa28d0e9"),
					w3.H("0x000000000000000000000000a0b86991c6218b36c1d19d4a2e9eb0ce3606eb48"),
					w3.H("0x000000000000000000000000c02aaa39b223fe8d0a0e5c4f27ead9083c756cc2"),
				},
				Data:        w3.B("0x000000000000000000000000b4e16d0168e52d35cacd2c6185b44281ec28c9dc0000000000000000000000000000000000000000000000000000000000000001"),
				BlockNumber: 10008355,
				TxHash:      w3.H("0xd07cbde817318492092cc7a27b3064a69bd893c01cb593d6029683ffd290ab3a"),
				TxIndex:     38,
				BlockHash:   w3.H("0x359d1dc4f14f9a07cba3ae8416958978ce98f78ad7b8d505925dad9722081f04"),
				Index:       34,
			},
			{
				Address: w3.A("0x5C69bEe701ef814a2B6a3EDD4B1652CB9cc5aA6f"),
				Topics: []common.Hash{
					w3.H("0x0d3648bd0f6ba80134a33ba9275ac585d9d315f0ad8355cddefde31afa28d0e9"),
					w3.H("0x0000000000000000000000008e870d67f660d95d5be530380d0ec0bd388289e1"),
					w3.H("0x000000000000000000000000a0b86991c6218b36c1d19d4a2e9eb0ce3606eb48"),
				},
				Data:        w3.B("0x0000000000000000000000003139ffc91b99aa94da8a2dc13f1fc36f9bdc98ee0000000000000000000000000000000000000000000000000000000000000002"),
				BlockNumber: 10008500,
				TxHash:      w3.H("0xb0621ca74cee9f540dda6d575f6a7b876133b42684c1259aaeb59c831410ccb2"),
				TxIndex:     35,
				BlockHash:   w3.H("0x27ff22f242123ca65b93d3886f1fa62bfa8f2a5d1c224750a7356b1c18b821f4"),
				Index:       28,
			},
		}
	)
	if err := client.Call(eth.Logs(filterQuery).Returns(&logs)); err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	if diff := cmp.Diff(wantLogs, logs); diff != "" {
		t.Fatalf("(-want, +got)\n%s", diff)
	}
}
