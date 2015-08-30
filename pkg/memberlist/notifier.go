// Package memberlist implements the raftor.ClusterChangeNotifier interface along with the memberlist.EventDelegate interface. The Notifier interface provides for a simple solution allowing the two packages to communicate membership changes.
package memberlist

import (
	"github.com/blacklabeldata/raftor"
	"github.com/blacklabeldata/raftor/pkg/murmur"
	"github.com/hashicorp/memberlist"
)

// NewEventDelegate creates a memberlist.EventDelegate and sends all
// cluster changes to the given raftor.ClusterChangeNotifier.
func NewEventDelegate(n raftor.ClusterChangeNotifier) memberlist.EventDelegate {
	return &notifier{n}
}

// notifier implements the memberlist.EventDelegate interface and forwards
// all the events to a raftor.ClusterChangeNotifier.
type notifier struct {
	n raftor.ClusterChangeNotifier
}

// NotifyJoin is invoked when a node is detected to have joined. A raftor.ClusterChangeEvent
// is created with the raftor.AddMember type and is sent to the Notifier.
func (n *notifier) NotifyJoin(other *memberlist.Node) {
	n.send(raftor.AddMember, other)
}

// NotifyLeave is invoked when a node is detected to have left. A raftor.ClusterChangeEvent
// is created with the raftor.RemoveMember type and is sent to the Notifier.
func (n *notifier) NotifyLeave(other *memberlist.Node) {
	n.send(raftor.RemoveMember, other)
}

// NotifyUpdate is invoked when a node is detected to have
// updated, usually involving the meta data. A raftor.ClusterChangeEvent
// is created with the raftor.UpdateMember type and is sent to the Notifier.
func (n *notifier) NotifyUpdate(other *memberlist.Node) {
	n.send(raftor.UpdateMember, other)
}

// send calls the Notifier to notify it of the cluster change
func (n *notifier) send(typ raftor.ClusterEventType, other *memberlist.Node) {
	n.n.Notify(raftor.ClusterChangeEvent{
		Type:   typ,
		Member: raftor.NewMember(n.hash(other), other.Meta),
	})
}

// hash performs a Murmur3 hash on the memberlist.Node name
func (n *notifier) hash(other *memberlist.Node) uint64 {
	return uint64(murmur.Murmur3([]byte(other.Name), murmur.M3Seed))
}
