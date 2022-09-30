package debug_test

import (
	"testing"

	"github.com/lmittmann/w3"
	"github.com/lmittmann/w3/module/debug"
	"github.com/lmittmann/w3/rpctest"
	"github.com/lmittmann/w3/w3types"
)

func TestCallTraceTx(t *testing.T) {
	tests := []rpctest.TestCase[debug.CallTrace]{
		{
			Golden: "traceCall_callTracer",
			Call: debug.CallTraceCall(&w3types.Message{
				From:  w3.A("0x000000000000000000000000000000000000c0Fe"),
				To:    w3.APtr("0x000000000000000000000000000000000000dEaD"),
				Value: w3.I("1 ether"),
			}, nil, w3types.State{
				w3.A("0x000000000000000000000000000000000000c0Fe"): {Balance: w3.I("1 ether")},
			}),
			WantRet: debug.CallTrace{
				From:  w3.A("0x000000000000000000000000000000000000c0Fe"),
				To:    w3.A("0x000000000000000000000000000000000000dEaD"),
				Type:  "CALL",
				Gas:   49979000,
				Value: w3.I("1 ether"),
			},
		},
	}

	rpctest.RunTestCases(t, tests)
}
