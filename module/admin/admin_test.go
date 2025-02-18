package admin_test

import (
	"math/big"
	"net"
	"testing"

	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/p2p/enr"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/lmittmann/w3"
	"github.com/lmittmann/w3/module/admin"
	"github.com/lmittmann/w3/rpctest"
)

func TestAddPeer(t *testing.T) {
	rpctest.RunTestCases(t, []rpctest.TestCase[bool]{
		{
			Golden:  "addPeer",
			Call:    admin.AddPeer(enode.MustParse("enode://a979fb575495b8d6db44f750317d0f4622bf4c2aa3365d6af7c284339968eef29b69ad0dce72a4d8db5ebb4968de0e3bec910127f134779fbcb0cb6d3331163c@52.16.188.185:30303")),
			WantRet: true,
		},
	})
}

func TestRemovePeer(t *testing.T) {
	rpctest.RunTestCases(t, []rpctest.TestCase[bool]{
		{
			Golden:  "removePeer",
			Call:    admin.RemovePeer(enode.MustParse("enode://a979fb575495b8d6db44f750317d0f4622bf4c2aa3365d6af7c284339968eef29b69ad0dce72a4d8db5ebb4968de0e3bec910127f134779fbcb0cb6d3331163c@52.16.188.185:30303")),
			WantRet: true,
		},
	})
}

func TestAddTrustedPeer(t *testing.T) {
	rpctest.RunTestCases(t, []rpctest.TestCase[bool]{
		{
			Golden:  "addTrustedPeer",
			Call:    admin.AddTrustedPeer(enode.MustParse("enode://a979fb575495b8d6db44f750317d0f4622bf4c2aa3365d6af7c284339968eef29b69ad0dce72a4d8db5ebb4968de0e3bec910127f134779fbcb0cb6d3331163c@52.16.188.185:30303")),
			WantRet: true,
		},
	})
}

func TestRemoveTrustedPeer(t *testing.T) {
	rpctest.RunTestCases(t, []rpctest.TestCase[bool]{
		{
			Golden:  "removeTrustedPeer",
			Call:    admin.RemoveTrustedPeer(enode.MustParse("enode://a979fb575495b8d6db44f750317d0f4622bf4c2aa3365d6af7c284339968eef29b69ad0dce72a4d8db5ebb4968de0e3bec910127f134779fbcb0cb6d3331163c@52.16.188.185:30303")),
			WantRet: true,
		},
	})
}

func TestNodeInfo(t *testing.T) {
	rpctest.RunTestCases(t, []rpctest.TestCase[*admin.NodeInfoResponse]{
		{
			Golden: "nodeInfo",
			Call:   admin.NodeInfo(),
			WantRet: &admin.NodeInfoResponse{
				Enode:      enode.MustParse("enode://44826a5d6a55f88a18298bca4773fca5749cdc3a5c9f308aa7d810e9b31123f3e7c5fba0b1d70aac5308426f47df2a128a6747040a3815cc7dd7167d03be320d@[::]:30303"),
				ID:         "44826a5d6a55f88a18298bca4773fca5749cdc3a5c9f308aa7d810e9b31123f3e7c5fba0b1d70aac5308426f47df2a128a6747040a3815cc7dd7167d03be320d",
				IP:         net.ParseIP("::"),
				ListenAddr: "[::]:30303",
				Name:       "reth/v0.0.1/x86_64-unknown-linux-gnu",
				Ports: &admin.PortsInfo{
					Discovery: 30303,
					Listener:  30303,
				},
				Protocols: map[string]*admin.ProtocolInfo{
					"eth": {
						Difficulty: w3.I("17334254859343145000"),
						Genesis:    w3.H("0xd4e56740f876aef8c010b86a40d5f56745a118d0906a34e69aec8c0db1cb8fa3"),
						Head:       w3.H("0xb83f73fbe6220c111136aefd27b160bf4a34085c65ba89f24246b3162257c36a"),
						Network:    1,
					},
				},
			},
		},
	}, cmpopts.IgnoreUnexported(big.Int{}, enode.Node{}, enr.Record{}))
}
