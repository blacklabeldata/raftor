package memberlist

import (
	"bytes"
	"testing"

	"github.com/blacklabeldata/raftor"
	"github.com/blacklabeldata/raftor/pkg/murmur"
	"github.com/hashicorp/memberlist"
	"github.com/stretchr/testify/assert"
)

// Implementation of raftor.ClusterChangeNotifier for testing
func Notifier(notifyc chan raftor.ClusterChangeEvent) raftor.ClusterChangeNotifier {
	return &testNotifier{notifyc}
}

// notifier notifies the receiver of the cluster change.
type testNotifier struct {
	channel chan raftor.ClusterChangeEvent
}

// send sends an event over the channel
func (n *testNotifier) Notify(evt raftor.ClusterChangeEvent) {
	n.channel <- evt
}

// TestBlockingNotifierJoin tests the BlockingNotifier with a NotifyJoin call. A ClusterChangeEvent should be generated and contain a Member representing the memberlist.Node that joined.
func TestBlockingNotifierJoin(t *testing.T) {
	metadata := []byte("metadata")
	node := memberlist.Node{
		Name: "Node 1",
		Meta: metadata,
	}

	// Setup notifier
	out := make(chan raftor.ClusterChangeEvent)
	defer close(out)

	// Create blocking notifier
	blocker := Notifier(out)

	// Create event delegate
	delegate := NewEventDelegate(blocker)

	// NotifyJoin
	go func() {
		delegate.NotifyJoin(&node)
	}()

	// Wait for event
	evt := <-out

	hash := uint64(murmur.Murmur3([]byte(node.Name), murmur.M3Seed))
	assert.Equal(t, raftor.AddMember, evt.Type)
	assert.Equal(t, hash, evt.Member.ID())
	assert.Equal(t, metadata, evt.Member.Meta())
}

// TestBlockingNotifierLeave tests the BlockingNotifier with a NotifyLeave call. A ClusterChangeEvent should be generated and contain a Member representing the memberlist.Node that left the cluster.
func TestBlockingNotifierLeave(t *testing.T) {

	// Create node
	metadata := []byte("metadata")
	node := memberlist.Node{
		Name: "Node 1",
		Meta: metadata,
	}

	// Setup notifier
	out := make(chan raftor.ClusterChangeEvent)
	defer close(out)

	// Create blocking notifier
	blocker := Notifier(out)

	// Create event delegate
	delegate := NewEventDelegate(blocker)

	// NotifyLeave
	go func() {
		delegate.NotifyLeave(&node)
	}()

	// Wait for event
	evt := <-out

	// Test correctness
	hash := uint64(murmur.Murmur3([]byte(node.Name), murmur.M3Seed))
	assert.Equal(t, raftor.RemoveMember, evt.Type)
	assert.Equal(t, hash, evt.Member.ID())
	assert.Equal(t, metadata, evt.Member.Meta())
}

// TestBlockingNotifierUpdate tests the BlockingNotifier with a NotifyUpdate call. A ClusterChangeEvent should be generated and contain a Member representing the memberlist.Node that has been updated.
func TestBlockingNotifierUpdate(t *testing.T) {

	// Create node
	metadata := []byte("metadata")
	node := memberlist.Node{
		Name: "Node 1",
		Meta: metadata,
	}

	// Setup notifier
	out := make(chan raftor.ClusterChangeEvent)
	defer close(out)

	// Create blocking notifier
	blocker := Notifier(out)

	// Create event delegate
	delegate := NewEventDelegate(blocker)

	// NotifyUpdate
	go func() {
		delegate.NotifyUpdate(&node)
	}()

	// Wait for event
	evt := <-out

	// Test correctness
	hash := uint64(murmur.Murmur3([]byte(node.Name), murmur.M3Seed))
	assert.Equal(t, raftor.UpdateMember, evt.Type)
	assert.Equal(t, hash, evt.Member.ID())
	assert.Equal(t, metadata, evt.Member.Meta())
}

// TestBlockingNotifier tests the BlockingNotifier with multiple messages.
func TestBlockingNotifier(t *testing.T) {
	node := memberlist.Node{
		Name: "Node 1",
	}

	// Hash node
	hash := uint64(murmur.Murmur3([]byte(node.Name), murmur.M3Seed))

	// Setup notifier
	out := make(chan raftor.ClusterChangeEvent)
	defer close(out)

	// Create blocking notifier
	notifier := Notifier(out)

	// Create event delegate
	delegate := NewEventDelegate(notifier)

	// Perform notifications
	go func() {
		node.Meta = []byte("join")
		delegate.NotifyJoin(&node)

		node.Meta = []byte("update")
		delegate.NotifyUpdate(&node)

		node.Meta = []byte("leave")
		delegate.NotifyLeave(&node)
	}()

	// Test join
	evt := <-out
	assert.Equal(t, raftor.AddMember, evt.Type)
	assert.Equal(t, hash, evt.Member.ID())
	assert.True(t, bytes.Equal([]byte("join"), evt.Member.Meta()))

	// Test update
	evt = <-out
	assert.Equal(t, raftor.UpdateMember, evt.Type)
	assert.Equal(t, hash, evt.Member.ID())
	assert.True(t, bytes.Equal([]byte("update"), evt.Member.Meta()))

	// Test leave
	evt = <-out
	assert.Equal(t, raftor.RemoveMember, evt.Type)
	assert.Equal(t, hash, evt.Member.ID())
	assert.True(t, bytes.Equal([]byte("leave"), evt.Member.Meta()))
}

// TestBufferedNotifierJoin tests the BufferedNotifier with multiple messages.
func TestBufferedNotifier(t *testing.T) {
	node := memberlist.Node{
		Name: "Node 1",
	}

	// Hash node
	hash := uint64(murmur.Murmur3([]byte(node.Name), murmur.M3Seed))

	// Setup notifier
	out := make(chan raftor.ClusterChangeEvent, 3)
	defer close(out)

	// Create blocking notifier
	notifier := Notifier(out)

	// Create event delegate
	delegate := NewEventDelegate(notifier)

	// Perform notifications
	go func() {
		node.Meta = []byte("join")
		delegate.NotifyJoin(&node)

		node.Meta = []byte("update")
		delegate.NotifyUpdate(&node)

		node.Meta = []byte("leave")
		delegate.NotifyLeave(&node)
	}()

	// Test join
	evt := <-out
	assert.Equal(t, raftor.AddMember, evt.Type)
	assert.Equal(t, hash, evt.Member.ID())
	assert.True(t, bytes.Equal([]byte("join"), evt.Member.Meta()))

	// Test update
	evt = <-out
	assert.Equal(t, raftor.UpdateMember, evt.Type)
	assert.Equal(t, hash, evt.Member.ID())
	assert.True(t, bytes.Equal([]byte("update"), evt.Member.Meta()))

	// Test leave
	evt = <-out
	assert.Equal(t, raftor.RemoveMember, evt.Type)
	assert.Equal(t, hash, evt.Member.ID())
	assert.True(t, bytes.Equal([]byte("leave"), evt.Member.Meta()))
}
