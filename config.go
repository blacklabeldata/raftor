package raftor

import "github.com/coreos/etcd/raft"

// ClusterConfig helps to configure a RaftNode
type ClusterConfig struct {
	Name string
	Raft raft.Config
}
