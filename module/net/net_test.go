package net_test

import (
	"math/big"
	"testing"

	"github.com/lmittmann/w3"
	"github.com/lmittmann/w3/module/net"
	"github.com/lmittmann/w3/rpctest"
)

func TestListening(t *testing.T) {
	rpctest.RunTestCases(t, []rpctest.TestCase[bool]{
		{
			Golden:  "listening",
			Call:    net.Listening(),
			WantRet: true,
		},
	})
}

func TestPeerCount(t *testing.T) {
	rpctest.RunTestCases(t, []rpctest.TestCase[*big.Int]{
		{
			Golden:  "peerCount",
			Call:    net.PeerCount(),
			WantRet: w3.I("10"),
		},
	})
}

func TestVersion(t *testing.T) {
	rpctest.RunTestCases(t, []rpctest.TestCase[int]{
		{
			Golden:  "version",
			Call:    net.Version(),
			WantRet: 1,
		},
	})
}
