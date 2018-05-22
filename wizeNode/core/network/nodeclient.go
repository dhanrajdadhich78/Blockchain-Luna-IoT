package network

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"

	"wizeBlock/wizeNode/core/blockchain"
	"wizeBlock/wizeNode/core/log"
)

// TODO: rethink with SendVersion
// TODO: rethink with SendAddr
// TODO: add NodeAuthStr and BuildCommandDataWithAuth
// TODO: rethink with Data messages

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

// Set currrent node address , to include itin requests to other nodes
func (c *NodeClient) SetNodeAddress(address NodeAddr) {
	c.NodeAddress = address
}

// Check if node address looks fine
func (c *NodeClient) CheckNodeAddress(address NodeAddr) error {
	if address.Port < 1024 {
		return errors.New("Node Address Port has wrong value")
	}
	if address.Port > 65536 {
		return errors.New("Node Address Port has wrong value")
	}
	if address.Host == "" {
		return errors.New("Node Address Host has wrong value")
	}
	return nil
}

// Builds a command data. It prepares a slice of bytes from given data
func (c *NodeClient) BuildCommandData(command string, data interface{}) ([]byte, error) {
	return c.doBuildCommandData(command, data)
}

// Builds a command data. It prepares a slice of bytes from given data
func (c *NodeClient) doBuildCommandData(command string, data interface{}) ([]byte, error) {
	var payload []byte
	var err error

	if data != nil {
		payload, err = GobEncode(data)

		if err != nil {
			return nil, err
		}
	} else {
		//payload = []byte{}
		return nil, fmt.Errorf("Empty data")
	}

	request := append(CommandToBytes(command), payload...)

	return request, nil
}

func (c *NodeClient) SendData(address NodeAddr, data []byte) error {
	err := c.CheckNodeAddress(address)
	if err != nil {
		return err
	}

	log.Debug.Printf("Sending %d bytes to %s", len(data), address)
	conn, err := net.Dial(Protocol, address.String())
	if err != nil {
		log.Warn.Println("Dial error: ", err.Error())

		// we can not connect
		// we could remove this node from known
		// but this is not always good. we need something more smart here
		// TODO this needs analysis. if removing of a node is good idea
		//c.Network.RemoveNodeFromKnown(address)

		return fmt.Errorf("%s is not available\n", address)
	}
	defer conn.Close()

	_, err = io.Copy(conn, bytes.NewReader(data))
	if err != nil {
		log.Warn.Printf("Send data error: %s", err)
	}

	log.Debug.Printf("Send data: %x", data)

	return nil
}

func (c *NodeClient) SendAddr(address NodeAddr, addresses []NodeAddr) error {
	nodes := ComAddr{addresses}

	request, err := c.BuildCommandData("addr", &nodes)
	if err != nil {
		return err
	}

	return c.SendData(address, request)
}

func (c *NodeClient) SendBlock(address NodeAddr, block *blockchain.Block) error {
	data := ComBlock{c.NodeAddress, block.Serialize()}

	request, err := c.BuildCommandData("block", &data)
	if err != nil {
		return err
	}

	return c.SendData(address, request)
}

func (c *NodeClient) SendInv(address NodeAddr, kind string, items [][]byte) error {
	data := ComInv{c.NodeAddress, kind, items}

	request, err := c.BuildCommandData("inv", &data)
	if err != nil {
		return err
	}

	return c.SendData(address, request)
}

func (c *NodeClient) SendGetBlocks(address NodeAddr) error {
	data := ComGetBlocks{c.NodeAddress}

	request, err := c.BuildCommandData("getblocks", &data)
	if err != nil {
		return err
	}

	return c.SendData(address, request)
}

func (c *NodeClient) SendGetData(address NodeAddr, kind string, id []byte) error {
	data := ComGetData{c.NodeAddress, kind, id}

	request, err := c.BuildCommandData("getdata", &data)
	if err != nil {
		return err
	}

	return c.SendData(address, request)
}

func (c *NodeClient) SendTx(address NodeAddr, tnx *blockchain.Transaction) error {
	data := ComTx{c.NodeAddress, tnx.Serialize()}

	request, err := c.BuildCommandData("tx", &data)
	if err != nil {
		return err
	}

	return c.SendData(address, request)
}

func (c *NodeClient) SendVersion(address NodeAddr, bestHeight int) error {
	data := ComVersion{NodeVersion, bestHeight, c.NodeAddress}

	request, err := c.BuildCommandData("version", &data)
	if err != nil {
		return err
	}

	return c.SendData(address, request)
}
