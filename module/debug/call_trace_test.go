package debug_test

import (
	"testing"

	"github.com/lmittmann/w3"
	"github.com/lmittmann/w3/module/debug"
	"github.com/lmittmann/w3/rpctest"
	"github.com/lmittmann/w3/w3types"
)

func TestCallTraceTx(t *testing.T) {
	rpctest.RunTestCases(t, []rpctest.TestCase[*debug.CallTrace]{
		{
			Golden: "traceCall_callTracer",
			Call: debug.CallTraceCall(&w3types.Message{
				From:  w3.A("0x000000000000000000000000000000000000c0Fe"),
				To:    w3.APtr("0x000000000000000000000000000000000000dEaD"),
				Value: w3.I("1 ether"),
			}, nil, w3types.State{
				w3.A("0x000000000000000000000000000000000000c0Fe"): {Balance: w3.I("1 ether")},
			}),
			WantRet: &debug.CallTrace{
				From:  w3.A("0x000000000000000000000000000000000000c0Fe"),
				To:    w3.A("0x000000000000000000000000000000000000dEaD"),
				Type:  "CALL",
				Gas:   49979000,
				Value: w3.I("1 ether"),
			},
		},
		{
			Golden: "traceTx_revertReason",
			Call:   debug.CallTraceTx(w3.H("0x6ea1798a2d0d21db18d6e45ca00f230160b05f172f6022aa138a0b605831d740"), w3types.State{}),
			WantRet: &debug.CallTrace{
				From:         w3.A("0x84abea9c66d30d00549429f5f687e16708aa20c0"),
				To:           w3.A("0xd0a7333587053a5bae772bd37b9aae724e367619"),
				Type:         "CALL",
				Gas:          146604,
				GasUsed:      81510,
				Value:        w3.I("0x0"),
				Error:        "execution reverted",
				RevertReason: "BA: Insufficient gas (ETH) for refund",
			},
		},
	})
}
