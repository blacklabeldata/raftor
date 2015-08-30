package raftor

// Notifier notifies the receiver of the cluster change.
type Notifier interface {

	// Notify returns a read-only channel of ClusterChangeEvents which can be read when a node joins, leaves or is updated in the cluster.
	Notify(ClusterChangeEvent)
}

// // NotifyChange sends ClusterChangeEvents over the given channel when a node joins, leaves or is updated in the cluster.
// func (n *notifier) Notify(evt ClusterChangeEvent) <-chan ClusterChangeEvent {
// 	return n.channel
// }

// // NotifyJoin is invoked when a node is detected to have joined.
// // The Node argument must not be modified.
// func (n *notifier) NotifyJoin(other *memberlist.Node) {
// 	n.send(ClusterChangeEvent{
// 		Type:   AddMember,
// 		Member: NewMember(n.hash(other), other.Meta),
// 	})
// 	return
// }

// // NotifyLeave is invoked when a node is detected to have left.
// // The Node argument must not be modified.
// func (n *notifier) NotifyLeave(other *memberlist.Node) {
// 	n.send(ClusterChangeEvent{
// 		Type:   RemoveMember,
// 		Member: NewMember(n.hash(other), other.Meta),
// 	})
// 	return
// }

// // NotifyUpdate is invoked when a node is detected to have
// // updated, usually involving the meta data. The Node argument
// // must not be modified.
// func (n *notifier) NotifyUpdate(other *memberlist.Node) {
// 	n.send(ClusterChangeEvent{
// 		Type:   UpdateMember,
// 		Member: NewMember(n.hash(other), other.Meta),
// 	})
// 	return
// }
