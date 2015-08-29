package raftor

// ClusterChangeNotifier notifies the receiver of the cluster change.
type ClusterChangeNotifier interface {

	// NotifyChange sends ClusterChangeEvents over the given channel when a node joins, leaves or is updated in the cluster.
	NotifyChange() <-chan ClusterChangeEvent

	// Stop stops the notifier from sending out any more notifications.
	Stop()
}
