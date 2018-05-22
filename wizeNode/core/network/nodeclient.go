package network

import (
	"bytes"
	"io"
	"net"

	"wizeBlock/wizeNode/core/blockchain"
	"wizeBlock/wizeNode/core/log"
)

type NodeClient struct {
	WalletAddress string
	NodeAddress   NodeAddr
	Network       *NodeNetwork
}

type ComAddr struct {
	AddrList []NodeAddr
}

type ComBlock struct {
	AddrFrom NodeAddr
	Block    []byte
}

type ComGetBlocks struct {
	AddrFrom NodeAddr
}

type ComGetData struct {
	AddrFrom NodeAddr
	Type     string
	ID       []byte
}

type ComInv struct {
	AddrFrom NodeAddr
	Type     string
	Items    [][]byte
}

type ComTx struct {
	AddFrom     NodeAddr
	Transaction []byte
}

type ComVersion struct {
	Version    int
	BestHeight int
	AddrFrom   NodeAddr
}

func (c *NodeClient) SetNodeAddress(address NodeAddr) {
	c.NodeAddress = address
}

func (c *NodeClient) SendData(address NodeAddr, data []byte) {
	conn, err := net.Dial(Protocol, address.NodeAddrToString())
	if err != nil {
		log.Warn.Printf("%s is not available\n", address)

		// TODO: should we update known nodes
		//var updatedNodes []string
		//for _, node := range KnownNodes {
		//	if node != address {
		//		updatedNodes = append(updatedNodes, node)
		//	}
		//}
		//KnownNodes = updatedNodes

		return
	}
	defer conn.Close()

	_, err = io.Copy(conn, bytes.NewReader(data))
	if err != nil {
		log.Warn.Printf("Send data error: %s", err)
	}
	log.Debug.Printf("Send data: %x", data)
}

func (c *NodeClient) SendAddr(address NodeAddr, addresses []NodeAddr) {
	nodes := ComAddr{addresses}
	payload, _ := GobEncode(nodes)
	request := append(CommandToBytes("addr"), payload...)
	c.SendData(address, request)
}

func (c *NodeClient) SendBlock(address NodeAddr, block *blockchain.Block) {
	data := ComBlock{c.NodeAddress, block.Serialize()}
	payload, _ := GobEncode(data)
	request := append(CommandToBytes("block"), payload...)
	c.SendData(address, request)
}

func (c *NodeClient) SendInv(address NodeAddr, kind string, items [][]byte) {
	inventory := ComInv{c.NodeAddress, kind, items}
	payload, _ := GobEncode(inventory)
	request := append(CommandToBytes("inv"), payload...)
	c.SendData(address, request)
}

func (c *NodeClient) SendGetBlocks(address NodeAddr) {
	payload, _ := GobEncode(ComGetBlocks{c.NodeAddress})
	request := append(CommandToBytes("getblocks"), payload...)
	c.SendData(address, request)
}

func (c *NodeClient) SendGetData(address NodeAddr, kind string, id []byte) {
	payload, _ := GobEncode(ComGetData{c.NodeAddress, kind, id})
	request := append(CommandToBytes("getdata"), payload...)
	c.SendData(address, request)
}

func (c *NodeClient) SendTx(address NodeAddr, tnx *blockchain.Transaction) {
	data := ComTx{c.NodeAddress, tnx.Serialize()}
	payload, _ := GobEncode(data)
	request := append(CommandToBytes("tx"), payload...)
	c.SendData(address, request)
}

func (c *NodeClient) SendVersion(address NodeAddr, bestHeight int) {
	payload, _ := GobEncode(ComVersion{NodeVersion, bestHeight, c.NodeAddress})
	request := append(CommandToBytes("version"), payload...)
	c.SendData(address, request)
}
