package raftor

import "golang.org/x/net/context"

// ClusterEventType is an enum describing how the cluster is changing.
type ClusterEventType uint8

const (
	// AddMember is used to describe a cluster change when a node is added.
	AddMember ClusterEventType = iota

	// RemoveMember is used to describe a cluster change when a node is removed.
	RemoveMember

	// UpdateMember is used to describe a cluster change when a node is updated.
	UpdateMember

	// StopCluster is sent to the cluster when the Raft server has stopped.
	StopCluster
)

// ClusterChangeEvent is used to store details about a cluster change. It is sent when a new node is detected and after the change has been applied to a raft log.
type ClusterChangeEvent struct {
	Type   ClusterEventType
	Member Member
}

// Cluster maintains an active list of nodes in the cluster. Cluster is also responsible for reporting and responding to changes in cluster membership.
type Cluster interface {

	// ID represents the cluster ID.
	ID() uint64

	// Name returns the Cluster's name
	Name() string

	// GetMember returns a Member instance based on it's ID.
	GetMember(uint64) Member

	// IsBanished checks whether the given ID has been removed from this
	// cluster at some point in the past
	IsBanished(id uint64) bool

	// NotifyChange sends ClusterChangeEvents over the given channel when a node joins, leaves or is updated in the cluster.
	NotifyChange() <-chan ClusterChangeEvent

	// ApplyChange is called after the ClusterChangeEvent has been processed and stored by Raft.
	ApplyChange() chan ClusterChangeEvent

	// Stop stops the cluster and triggers the context when finished
	Stop(context.Context)
}