package txpool_test

import (
	"testing"

	"github.com/lmittmann/w3/module/txpool"
	"github.com/lmittmann/w3/rpctest"
)

func TestStatus(t *testing.T) {
	rpctest.RunTestCases(t, []rpctest.TestCase[*txpool.StatusResponse]{
		{
			Golden:  "status",
			Call:    txpool.Status(),
			WantRet: &txpool.StatusResponse{Pending: 10, Queued: 7},
		},
	})
}
