package raftor

// Notifier notifies the receiver of the cluster change.
type Notifier interface {

	// Notify returns a read-only channel of ClusterChangeEvents which can be read when a node joins, leaves or is updated in the cluster.
	Notify(ClusterChangeEvent)
}
