package raftor

import (
	"github.com/coreos/etcd/raft/raftpb"
	"golang.org/x/net/context"
)

// Applier applies either a snapshot or entries in a Commit object
type Applier interface {

	// Apply processes commit messages after being processed by Raft
	Apply() chan Commit
}

// Commit is used to send to the cluster to save either a snapshot or log entries.
type Commit struct {
	Entries  []raftpb.Entry
	Snapshot raftpb.Snapshot
	Context  context.Context
}
