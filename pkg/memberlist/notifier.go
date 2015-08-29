// Package memberlist implements the raftor.ClusterChangeNotifier interface along with the memberlist.EventDelegate interface. The Notifier interface provides for a simple solution allowing the two packages to communicate membership changes.
package memberlist

import (
	"github.com/blacklabeldata/raftor"
	"github.com/blacklabeldata/raftor/pkg/murmur"
	"github.com/hashicorp/memberlist"
)

// BlockingNotifier blocks until each event has been consumed. It also implements the memberlist.EventDelegate interface which is used to process cluster membership events.
func BlockingNotifier() Notifier {
	return &notifier{make(chan raftor.ClusterChangeEvent), false}
}

// BufferedNotifier will buffer events up to the size given. After the buffered channel is full it will block until an event has been consumed.
func BufferedNotifier(size int) Notifier {
	return &notifier{make(chan raftor.ClusterChangeEvent, size), false}
}

// A Notifier subscribes to memberlist notifications and emits them as raftor.ClusterChangeEvents.
type Notifier interface {
	raftor.ClusterChangeNotifier
	memberlist.EventDelegate
}

// notifier notifies the receiver of the cluster change.
type notifier struct {
	channel chan raftor.ClusterChangeEvent
	closed  bool
}

// NotifyChange sends raftor.ClusterChangeEvents over the given channel when a node joins, leaves or is updated in the cluster.
func (n *notifier) NotifyChange() <-chan raftor.ClusterChangeEvent {
	return n.channel
}

// NotifyJoin is invoked when a node is detected to have joined.
// The Node argument must not be modified.
func (n *notifier) NotifyJoin(other *memberlist.Node) {
	n.send(raftor.ClusterChangeEvent{
		Type:   raftor.AddMember,
		Member: raftor.NewMember(n.hash(other), other.Meta),
	})
	return
}

// NotifyLeave is invoked when a node is detected to have left.
// The Node argument must not be modified.
func (n *notifier) NotifyLeave(other *memberlist.Node) {
	n.send(raftor.ClusterChangeEvent{
		Type:   raftor.RemoveMember,
		Member: raftor.NewMember(n.hash(other), other.Meta),
	})
	return
}

// NotifyUpdate is invoked when a node is detected to have
// updated, usually involving the meta data. The Node argument
// must not be modified.
func (n *notifier) NotifyUpdate(other *memberlist.Node) {
	n.send(raftor.ClusterChangeEvent{
		Type:   raftor.UpdateMember,
		Member: raftor.NewMember(n.hash(other), other.Meta),
	})
	return
}

// Stop closes the notifier channel
func (n *notifier) Stop() {
	n.closed = true
	close(n.channel)
}

// send sends an event over the channel
func (n *notifier) send(evt raftor.ClusterChangeEvent) {
	if !n.closed {
		n.channel <- evt
	}
}

// hash performs a Murmur3 hash on the memberlist.Node
func (n *notifier) hash(other *memberlist.Node) uint64 {
	return uint64(murmur.Murmur3(other.Meta, murmur.M3Seed))
}
