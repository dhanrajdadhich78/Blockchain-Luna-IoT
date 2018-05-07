package app

// This manages list of known nodes by a node
type NodeNetwork struct {
	Nodes []NodeAddr
}

type NodesListJSON struct {
	Nodes   []NodeAddr
	Genesis string
}

// Set nodes list. This can be used to do initial nodes loading from config or so
func (n *NodeNetwork) SetNodes(nodes []NodeAddr, replace bool) {
	if replace {
		n.Nodes = nodes
	} else {
		n.Nodes = append(n.Nodes, nodes...)
	}
}

func (n *NodeNetwork) GetNodes() []NodeAddr {
	return n.Nodes
}

// Returns number of known nodes
func (n *NodeNetwork) GetCountOfKnownNodes() int {
	l := len(n.Nodes)

	return l
}

// Check if node address is known
func (n *NodeNetwork) CheckIsKnown(addr NodeAddr) bool {
	exists := false

	for _, node := range n.Nodes {
		if node.CompareToAddress(addr) {
			exists = true
			break
		}
	}

	return exists
}

/*
* Checks if a node exists in list of known nodes and adds it if no
* Returns true if was added
 */
func (n *NodeNetwork) AddNodeToKnown(addr NodeAddr) bool {
	exists := false

	for _, node := range n.Nodes {
		if node.CompareToAddress(addr) {
			exists = true
			break
		}
	}
	if !exists {
		n.Nodes = append(n.Nodes, addr)
	}

	return !exists
}

// Removes a node from known
func (n *NodeNetwork) RemoveNodeFromKnown(addr NodeAddr) {
	updatedlist := []NodeAddr{}

	for _, node := range n.Nodes {
		if !node.CompareToAddress(addr) {
			updatedlist = append(updatedlist, node)
		}
	}

	n.Nodes = updatedlist
}
