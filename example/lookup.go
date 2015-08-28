package main

import (
	"sync"

	"github.com/hashicorp/memberlist"
)

// NodeLookup provides an interface for looking up a node based on an id.
type NodeLookup interface {
	SetNode(uint64, *memberlist.Node)
	GetNode(uint64) *memberlist.Node
	DeleteNode(uint64)
}

// NewNodeLookup creates a new lookup table for memberlist.Nodes based on the node's metadata
func NewNodeLookup() NodeLookup {
	var mu sync.Mutex
	return &nodeLookup{mu, make(map[uint64]*memberlist.Node)}
}

type nodeLookup struct {
	mu    sync.Mutex
	nodes map[uint64]*memberlist.Node
}

func (n *nodeLookup) SetNode(id uint64, node *memberlist.Node) {
	n.mu.Lock()
	n.nodes[id] = node
	n.mu.Unlock()
}

func (n *nodeLookup) GetNode(id uint64) (node *memberlist.Node) {
	n.mu.Lock()
	node = n.nodes[id]
	n.mu.Unlock()
	return
}

func (n *nodeLookup) DeleteNode(id uint64) {
	n.mu.Lock()
	delete(n.nodes, id)
	n.mu.Unlock()
}
