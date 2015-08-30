package raftor

// Updater applies cluster change events
type Updater interface {

	// Update is called after the ClusterChangeEvent has been processed and stored by Raft.
	Update(ClusterChangeEvent)
}
