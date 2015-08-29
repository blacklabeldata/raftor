package raftor

import "github.com/coreos/etcd/raft"

// RaftNode extends the etcd raft.Node
type RaftNode interface {
	raft.Node
	Starter
}
