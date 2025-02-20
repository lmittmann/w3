/*
Package admin implements RPC API bindings for methods in the "admin" namespace.
*/
package admin

import (
	"math/big"
	"net"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/lmittmann/w3/internal/module"
	"github.com/lmittmann/w3/w3types"
)

// AddPeer adds the given peer to the node's peer set and returns a bool
// indicating success.
func AddPeer(url *enode.Node) w3types.RPCCallerFactory[bool] {
	return module.NewFactory[bool](
		"admin_addPeer",
		[]any{url},
	)
}

// RemovePeer disconnects from the given peer if the connection exists and
// returns a bool indicating success.
func RemovePeer(url *enode.Node) w3types.RPCCallerFactory[bool] {
	return module.NewFactory[bool](
		"admin_removePeer",
		[]any{url},
	)
}

// AddTrustedPeer adds the given peer to the trusted peers list and returns a
// bool indicating success.
func AddTrustedPeer(url *enode.Node) w3types.RPCCallerFactory[bool] {
	return module.NewFactory[bool](
		"admin_addTrustedPeer",
		[]any{url},
	)
}

// RemoveTrustedPeer removes a remote node from the trusted peers list and
// returns a bool indicating success.
func RemoveTrustedPeer(url *enode.Node) w3types.RPCCallerFactory[bool] {
	return module.NewFactory[bool](
		"admin_removeTrustedPeer",
		[]any{url},
	)
}

// NodeInfo returns information about the running node.
func NodeInfo() w3types.RPCCallerFactory[*NodeInfoResponse] {
	return module.NewFactory[*NodeInfoResponse](
		"admin_nodeInfo",
		[]any{},
	)
}

type NodeInfoResponse struct {
	Enode      *enode.Node              `json:"enode"`
	ID         string                   `json:"id"`
	IP         net.IP                   `json:"ip"`
	ListenAddr string                   `json:"listenAddr"`
	Name       string                   `json:"name"`
	Ports      *PortsInfo               `json:"ports"`
	Protocols  map[string]*ProtocolInfo `json:"protocols"`
}

type PortsInfo struct {
	Discovery int `json:"discovery"`
	Listener  int `json:"listener"`
}

type ProtocolInfo struct {
	Difficulty *big.Int    `json:"difficulty"`
	Genesis    common.Hash `json:"genesis"`
	Head       common.Hash `json:"head"`
	Network    int         `json:"network"`
}
