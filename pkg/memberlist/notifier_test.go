package memberlist

import (
	"bytes"
	"testing"

	"github.com/blacklabeldata/raftor"
	"github.com/blacklabeldata/raftor/pkg/murmur"
	"github.com/hashicorp/memberlist"
	"github.com/stretchr/testify/assert"
)

// TestBlockingNotifierJoin tests the BlockingNotifier with a NotifyJoin call. A ClusterChangeEvent should be generated and contain a Member representing the memberlist.Node that joined.
func TestBlockingNotifierJoin(t *testing.T) {
	metadata := []byte("metadata")
	node := memberlist.Node{
		Name: "Node 1",
		Meta: metadata,
	}

	// Create BlockingNotifier
	blocker := BlockingNotifier()

	// NotifyJoin
	go func() {
		blocker.NotifyJoin(&node)
	}()

	evt := <-blocker.Notify()
	blocker.Stop()

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

	// Create BlockingNotifier
	blocker := BlockingNotifier()

	// NotifyLeave
	go func() {
		blocker.NotifyLeave(&node)
	}()

	// Wait for event
	evt := <-blocker.Notify()

	// Stop notifier
	blocker.Stop()

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

	// Create BlockingNotifier
	blocker := BlockingNotifier()

	// NotifyUpdate
	go func() {
		blocker.NotifyUpdate(&node)
	}()

	// Wait for event
	evt := <-blocker.Notify()

	// Stop notifier
	blocker.Stop()

	// Test correctness
	hash := uint64(murmur.Murmur3([]byte(node.Name), murmur.M3Seed))
	assert.Equal(t, raftor.UpdateMember, evt.Type)
	assert.Equal(t, hash, evt.Member.ID())
	assert.Equal(t, metadata, evt.Member.Meta())
}

// TestBlockingNotifierStopPanic tests the BlockingNotifier with a NotifyUpdate call after the Notifier has been stopped. The update message should cause a panic as a message will be sent after the channel has been closed. In order to avoid this panic, the memberlist should be shut down before stopping a Notifier.
func TestBlockingNotifierStopPanic(t *testing.T) {
	defer func() {
		assert.NotNil(t, recover(), "Should panic")
	}()

	metadata := []byte("metadata")
	node := memberlist.Node{
		Name: "Node 1",
		Meta: metadata,
	}

	// Create BlockingNotifier
	blocker := BlockingNotifier()
	blocker.Stop()

	// NotifyUpdate
	blocker.NotifyUpdate(&node)
}

// TestBlockingNotifier tests the BlockingNotifier with multiple messages.
func TestBlockingNotifier(t *testing.T) {
	node := memberlist.Node{
		Name: "Node 1",
	}

	// Hash node
	hash := uint64(murmur.Murmur3([]byte(node.Name), murmur.M3Seed))

	// Create BufferedNotifier
	notifier := BlockingNotifier()

	// Perform notifications
	go func() {
		node.Meta = []byte("join")
		notifier.NotifyJoin(&node)

		node.Meta = []byte("update")
		notifier.NotifyUpdate(&node)

		node.Meta = []byte("leave")
		notifier.NotifyLeave(&node)
	}()

	// Test join
	evt := <-notifier.Notify()
	assert.Equal(t, raftor.AddMember, evt.Type)
	assert.Equal(t, hash, evt.Member.ID())
	assert.True(t, bytes.Equal([]byte("join"), evt.Member.Meta()))

	// Test update
	evt = <-notifier.Notify()
	assert.Equal(t, raftor.UpdateMember, evt.Type)
	assert.Equal(t, hash, evt.Member.ID())
	assert.True(t, bytes.Equal([]byte("update"), evt.Member.Meta()))

	// Test leave
	evt = <-notifier.Notify()
	assert.Equal(t, raftor.RemoveMember, evt.Type)
	assert.Equal(t, hash, evt.Member.ID())
	assert.True(t, bytes.Equal([]byte("leave"), evt.Member.Meta()))

	// Notifier
	notifier.Stop()
}

// TestBufferedNotifierJoin tests the BufferedNotifier with multiple messages.
func TestBufferedNotifier(t *testing.T) {
	node := memberlist.Node{
		Name: "Node 1",
	}

	// Hash node
	hash := uint64(murmur.Murmur3([]byte(node.Name), murmur.M3Seed))

	// Create BufferedNotifier
	notifier := BufferedNotifier(3)

	// Perform notifications
	go func() {
		node.Meta = []byte("join")
		notifier.NotifyJoin(&node)

		node.Meta = []byte("update")
		notifier.NotifyUpdate(&node)

		node.Meta = []byte("leave")
		notifier.NotifyLeave(&node)
	}()

	// Test join
	evt := <-notifier.Notify()
	assert.Equal(t, raftor.AddMember, evt.Type)
	assert.Equal(t, hash, evt.Member.ID())
	assert.True(t, bytes.Equal([]byte("join"), evt.Member.Meta()))

	// Test update
	evt = <-notifier.Notify()
	assert.Equal(t, raftor.UpdateMember, evt.Type)
	assert.Equal(t, hash, evt.Member.ID())
	assert.True(t, bytes.Equal([]byte("update"), evt.Member.Meta()))

	// Test leave
	evt = <-notifier.Notify()
	assert.Equal(t, raftor.RemoveMember, evt.Type)
	assert.Equal(t, hash, evt.Member.ID())
	assert.True(t, bytes.Equal([]byte("leave"), evt.Member.Meta()))

	// Stop notifier
	notifier.Stop()
}
