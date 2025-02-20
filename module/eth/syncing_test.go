package eth_test

import (
	"testing"

	"github.com/lmittmann/w3/module/eth"
	"github.com/lmittmann/w3/rpctest"
)

func TestSyncing(t *testing.T) {
	rpctest.RunTestCases(t, []rpctest.TestCase[bool]{
		{
			Golden:  "syncing__false",
			Call:    eth.Syncing(),
			WantRet: false,
		},
		{
			Golden:  "syncing__true",
			Call:    eth.Syncing(),
			WantRet: true,
		},
	})
}
