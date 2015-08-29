package memberlist

import "github.com/blacklabeldata/raftor"
import "github.com/blacklabeldata/raftor/pkg/murmur"

// ClusterChangeNotifier notifies the receiver of the cluster change.
type notifier struct {
	channel <-chan raftor.ClusterChangeEvent
}

// NotifyChange sends raftor.ClusterChangeEvents over the given channel when a node joins, leaves or is updated in the cluster.
func (n *notifier) NotifyChange() <-chan raftor.ClusterChangeEvent {
	return n.channel
}

// NotifyJoin is invoked when a node is detected to have joined.
// The Node argument must not be modified.
func (n *notifier) NotifyJoin(other *memberlist.Node) {
	n.channel <- raftor.ClusterChangeEvent{
		Type:   raftor.AddMember,
		Member: raftor.NewMember(n.hash(other), other.Meta),
	}
	return
}

// NotifyLeave is invoked when a node is detected to have left.
// The Node argument must not be modified.
func (n *notifier) NotifyLeave(other *memberlist.Node) {
	n.channel <- raftor.ClusterChangeEvent{
		Type:   raftor.RemoveMember,
		Member: raftor.NewMember(n.hash(other), other.Meta),
	}
	return
}

// NotifyUpdate is invoked when a node is detected to have
// updated, usually involving the meta data. The Node argument
// must not be modified.
func (n *notifier) NotifyUpdate(other *memberlist.Node) {
	n.channel <- raftor.ClusterChangeEvent{
		Type:   raftor.UpdateMember,
		Member: raftor.NewMember(n.hash(other), other.Meta),
	}
	return
}

// hash performs a Murmur3 hash on the memberlist.Node
func (n *notifier) hash(other *memberlist.Node) uint64 {
	return uint64(murmur.Murmur3(other.Meta, murmur.M3Seed))
}
