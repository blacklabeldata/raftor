package raftor

// ClusterChangeHandler is used to process ClusterChangeEvents.
type ClusterChangeHandler interface {

	// HandleJoin is called when a node joins the cluster.
	HandleJoin(ClusterChangeEvent)

	// HandleJoin is called when a node leaves the cluster.
	HandleLeave(ClusterChangeEvent)

	// HandleJoin is called when a node is updated.
	HandleUpdate(ClusterChangeEvent)
}
