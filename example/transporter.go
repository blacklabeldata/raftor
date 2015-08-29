package main

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/coreos/etcd/raft/raftpb"

	"github.com/hashicorp/memberlist"
)

type Transporter interface {

	// AddRemote tries to connect to the remote server over TCP in order to send Raft messages
	AddRemote(uint64, *memberlist.Node)

	// RemoveRemote disconnects from the remote server
	RemoveRemote(uint64, *memberlist.Node)

	// UpdateRemote disconnects from the remote server and reconnects if any config has changed
	UpdateRemote(uint64, *memberlist.Node)

	// Send sends out the given messages to the remote peers.
	// Each message has a To field, which is an id that maps
	// to an existing peer in the transport.
	// If the id cannot be found in the transport, the message
	// will be ignored.
	Send(messages []raftpb.Message)

	// Stop closes the connections and stops the transporter.
	Stop()
}

func NewTcpTransporter() Transporter {
	return &tcpTransporter{}
}

type tcpTransporter struct {
	mu      sync.Mutex
	remotes map[uint64]net.Conn
	stopped bool
}

func (t *tcpTransporter) Stop() {
	t.mu.Lock()
	for _, c := range t.remotes {
		c.Close()
	}
	t.stopped = true
	t.mu.Unlock()
}

func (t *tcpTransporter) Send(ms []raftpb.Message) {
	t.mu.Lock()
	if !t.stopped {
		log.Printf("[RaftTransport] Sending messages: %d", len(ms))

		// Message Aggregator
		var aggr map[uint64][]raftpb.Message

		// Aggregate messages
		for i := range ms {
			aggr[ms[i].To] = append(aggr[ms[i].To], ms[i])
		}

		var b []byte
		buffer := bytes.NewBuffer(b)

		// Send messages
		for remote, msgs := range aggr {
			c, ok := t.remotes[remote]
			if !ok || c == nil {
				continue
			}

			for _, m := range msgs {
				buffer.WriteByte(byte(RaftMessageType))
				enc, err := m.Marshal()
				if err != nil {
					log.Println("Error encoding raft message: ", err)
					goto reset
				}
				buffer.Write(enc)
			}

			// Send buffer
			if _, err := c.Write(buffer.Bytes()); err != nil {
				log.Printf("Error sending raft messages to %s: %s", c.RemoteAddr().String(), err.Error())
				t.removeRemote(remote)
			}

		reset:
			// Reset buffer
			buffer.Reset()
		}

		// for i := range ms {
		// 	if node := m.lookup.GetNode(ms[i].To); node != nil {
		// 		if data, err := ms[i].Marshal(); err == nil {
		// 			buffer.Write(data)
		// 			m.list.SendToTCP(node, buffer.Bytes())
		// 			buffer.Reset()
		// 		}
		// 	}
		// }
	}

	t.mu.Unlock()
}

func (t *tcpTransporter) AddRemote(id uint64, remote *memberlist.Node) {
	t.mu.Lock()
	t.addRemote(id, remote)
	t.mu.Unlock()
}

func (t *tcpTransporter) addRemote(id uint64, remote *memberlist.Node) {

	// If the remote connection is already established, don't do anything.
	if _, ok := t.remotes[id]; ok {
		return
	}

	meta := NodeMeta{}
	if err := meta.Unmarshal(remote.Meta); err != nil {
		log.Printf("Error unmarshalling metadata: %s", err.Error())
		return
	}

	// Connect to remote server
	addr := fmt.Sprintf("%s:%d", remote.Addr.String(), meta.Port)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Printf("Error connecting to remote server: %d : %s\n", id, remote.Addr.String())
		return
	}

	// Set connection
	t.remotes[meta.NameHash] = conn
	log.Printf("Connected to remote: %s\n", addr)
}

func (t *tcpTransporter) RemoveRemote(id uint64, remote *memberlist.Node) {
	t.mu.Lock()
	t.removeRemote(id)
	log.Println("Closed TCP connection to remote: " + remote.Addr.String())
	t.mu.Unlock()
}

func (t *tcpTransporter) removeRemote(id uint64) {
	if c, ok := t.remotes[id]; ok {
		c.Close()
	}
	delete(t.remotes, id)
}

func (t *tcpTransporter) UpdateRemote(id uint64, remote *memberlist.Node) {
	t.mu.Lock()
	log.Printf("Updating connection to remote server: %d : %s", id, remote.Addr.String())
	t.removeRemote(id)
	t.addRemote(id, remote)
	t.mu.Unlock()
}

// func NewMemberlistTransporter(list *memberlist.Memberlist, lookup NodeLookup) Transporter {
// 	return &memberListTransporter{list, false, lookup}
// }

// type memberListTransporter struct {
// 	list    *memberlist.Memberlist
// 	stopped bool
// 	lookup  NodeLookup
// }

// func (m *memberListTransporter) Stop() {
// 	m.stopped = true
// }

// func (m *memberListTransporter) Send(ms []raftpb.Message) {
// 	if !m.stopped {
// 		log.Printf("[RaftTransport] Sending messages: %d", len(ms))

// 		var b []byte
// 		buffer := bytes.NewBuffer(b)
// 		buffer.WriteByte(byte(RaftMessageType))
// 		for i := range ms {
// 			if node := m.lookup.GetNode(ms[i].To); node != nil {
// 				if data, err := ms[i].Marshal(); err == nil {
// 					buffer.Write(data)
// 					m.list.SendToTCP(node, buffer.Bytes())
// 					buffer.Reset()
// 				}
// 			}
// 		}
// 	}
// }
