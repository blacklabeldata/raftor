package raftor

// Stopper stops whatever you need to be stopped. It is meant to be used inside a select statement.
type Stopper interface {
	Stop() <-chan struct{}
}
