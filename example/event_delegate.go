package main

import (
	"log"
	"sync"

	"github.com/hashicorp/memberlist"
)

// MultiEventDelegate dispatches membership events to multiple EventDelegates.
type MultiEventDelegate struct {
	mu   sync.Mutex
	list []memberlist.EventDelegate
}

// NotifyJoin is invoked when a node is detected to have joined.
// The Node argument must not be modified.
func (m *MultiEventDelegate) NotifyJoin(node *memberlist.Node) {
	m.mu.Lock()
	log.Printf("[Memberlist] NotifyJoin: %s", node.Name)
	for i := range m.list {
		m.list[i].NotifyJoin(node)
	}
	m.mu.Unlock()
}

// NotifyLeave is invoked when a node is detected to have left.
// The Node argument must not be modified.
func (m *MultiEventDelegate) NotifyLeave(node *memberlist.Node) {
	m.mu.Lock()
	log.Printf("[Memberlist] NotifyLeave: %s", node.Name)
	for i := range m.list {
		m.list[i].NotifyLeave(node)
	}
	m.mu.Unlock()
}

// NotifyUpdate is invoked when a node is detected to have
// updated, usually involving the meta data. The Node argument
// must not be modified.
func (m *MultiEventDelegate) NotifyUpdate(node *memberlist.Node) {
	m.mu.Lock()
	log.Printf("[Memberlist] NotifyUpdate: %s", node.Name)
	for i := range m.list {
		m.list[i].NotifyUpdate(node)
	}
	m.mu.Unlock()
}

func (m *MultiEventDelegate) AddEventDelegate(e memberlist.EventDelegate) {
	m.mu.Lock()
	m.list = append(m.list, e)
	m.mu.Unlock()
}
