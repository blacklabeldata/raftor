package main

import (
	"errors"
	"log"
	"time"

	"github.com/hashicorp/memberlist"
	"github.com/swiftkick-io/xbinary"
)

// NodeMeta stores metadata about the local node and will be encoded to be sent over memberlist.
type NodeMeta struct {
	NameHash uint64
	Port     uint32
}

// Marshal encodes the metadata to be sent over the wire
func (n NodeMeta) Marshal() ([]byte, error) {
	buf := make([]byte, 8)
	xbinary.LittleEndian.PutUint32(buf, 0, uint32(n.NameHash))
	xbinary.LittleEndian.PutUint32(buf, 4, uint32(n.Port))
	return buf, nil
}

// Unmarshal decodes the metadata sent over the wire
func (n NodeMeta) Unmarshal(data []byte) error {
	if len(data) < 8 {
		return errors.New("MetaData incorrect length")
	}

	// Decode node id
	h, err := xbinary.LittleEndian.Uint32(data, 0)
	if err != nil {
		return err
	}
	n.NameHash = uint64(h)

	// Decode TCP port
	port, err := xbinary.LittleEndian.Uint32(data, 4)
	if err != nil {
		return err
	}
	n.Port = port
	return nil
}

func NewDelegate(name string, tcpport int) memberlist.Delegate {
	return delegate{
		name: name,
		meta: NodeMeta{
			NameHash: uint64(Jesteress([]byte(name))),
			Port:     uint32(tcpport),
		},
		timer: time.Tick(time.Second),
		// raft:  raft,
	}
}

type delegate struct {
	name  string
	meta  NodeMeta
	timer <-chan time.Time
	// raft  *raftNode
}

// NodeMeta is used to retrieve meta-data about the current node
// when broadcasting an alive message. It's length is limited to
// the given byte size. This metadata is available in the Node structure.
func (e delegate) NodeMeta(limit int) (meta []byte) {
	log.Printf("Node Metadata limit: %d", limit)

	if buf, err := e.meta.Marshal(); err == nil {
		meta = buf
	}
	return
}

// NotifyMsg is called when a user-data message is received.
// Care should be taken that this method does not block, since doing
// so would block the entire UDP packet receive loop. Additionally, the byte
// slice may be modified after the call returns, so it should be copied if needed.
func (d delegate) NotifyMsg(data []byte) {

	// Ignore empty message
	if len(data) == 0 {
		return
	}

	switch UserMessageType(data[0]) {
	case RaftMessageType:
		// log.Printf("Received Raft message type")
		// m := raftpb.Message{}
		// err := m.Unmarshal(data[1:])
		// if err != nil {
		// 	return
		// }
		// go d.raft.Step(m)
	case TimestampMessageType:
		log.Printf("Received message from %s", string(data))

		// Send to raft
		// go d.raft.Propose(data)
	case UnknownMessageType:
		log.Printf("Received unknown message type")
	}
}

// GetBroadcasts is called when user data messages can be broadcast.
// It can return a list of buffers to send. Each buffer should assume an
// overhead as provided with a limit on the total byte size allowed.
// The total byte size of the resulting data to send must not exceed
// the limit.
func (d delegate) GetBroadcasts(overhead, limit int) [][]byte {
	select {
	case <-d.timer:
		data := []byte{byte(TimestampMessageType)}
		data = append(data, []byte(d.name+": "+time.Now().String())...)

		// go d.raft.Propose(data)
		return [][]byte{data}
	default:
		return nil
	}
}

// LocalState is used for a TCP Push/Pull. This is sent to
// the remote side in addition to the membership information. Any
// data can be sent here. See MergeRemoteState as well. The `join`
// boolean indicates this is for a join instead of a push/pull.
func (delegate) LocalState(join bool) []byte {
	return nil
}

// MergeRemoteState is invoked after a TCP Push/Pull. This is the
// state received from the remote side and is the result of the
// remote side's LocalState call. The 'join'
// boolean indicates this is for a join instead of a push/pull.
func (delegate) MergeRemoteState(buf []byte, join bool) {
}
