package raftor

import (
	"time"

	"github.com/coreos/etcd/raft"
	"github.com/coreos/etcd/raft/raftpb"
	"golang.org/x/net/context"
)

// Commit is used to send to the cluster to save either a snapshot or log entries.
type Commit struct {
	State    RaftState
	Entries  []raftpb.Entry
	Snapshot raftpb.Snapshot
	Messages []raftpb.Message
	Context  context.Context
}

// RaftState describes the state of the Raft cluster for each commit
type RaftState struct {
	CommitID             uint64
	Vote                 uint64
	Term                 uint64
	Lead                 uint64
	LastLeadElectionTime time.Time
	RaftState            raft.StateType
}

// Applier applies either a snapshot or entries in a Commit object
type Applier interface {

	// Apply processes commit messages after being processed by Raft
	Apply(Commit)
}
