package raftor

// ClusterChangeNotifier notifies the receiver of the cluster change.
type ClusterChangeNotifier interface {

	// Change returns a channel of ClusterChangeEvents which can be read when a node joins, leaves or is updated in the cluster.
	Notify() <-chan ClusterChangeEvent

	// Stop stops the notifier from sending out any more notifications.
	Stop()
}
