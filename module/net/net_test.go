package net_test

import (
	"testing"

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
	rpctest.RunTestCases(t, []rpctest.TestCase[int]{
		{
			Golden:  "peer_count",
			Call:    net.PeerCount(),
			WantRet: 10,
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
