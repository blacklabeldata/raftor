package raftor

import "github.com/coreos/etcd/raft"

// RaftConfig helps to configure a RaftNode
type RaftConfig struct {
	Name    string
	Cluster Cluster
	Storage raft.Storage

	Raft raft.Config
}
