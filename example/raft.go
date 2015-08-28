package main

import (
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/coreos/etcd/pkg/idutil"
	"github.com/coreos/etcd/pkg/wait"
	"github.com/coreos/etcd/raft"
	"github.com/coreos/etcd/raft/raftpb"
	"github.com/hashicorp/memberlist"
	"github.com/swiftkick-io/xbinary"
	"golang.org/x/net/context"
)

func NewRaftNode(name string, c *raft.Config, storage *raft.MemoryStorage, lookup NodeLookup, tr Transporter) *raftNode {
	var peers []raft.Peer
	node := raft.StartNode(c, peers)
	var mu sync.Mutex
	removed := make(map[uint64]bool)
	h := Jesteress([]byte(name))

	return &raftNode{
		index: 0,
		term:  0,
		lead:  0,
		hash:  h,

		lt:   time.Now(),
		Node: node,
		cfg:  *c,
		mu:   mu,

		lookup:  lookup,
		removed: removed,

		idgen:       idutil.NewGenerator(uint8(h), time.Now()),
		w:           wait.New(),
		ticker:      time.Tick(500 * time.Millisecond),
		raftStorage: storage,
		storage:     &NoopStorage{},
		transport:   tr,
		applyc:      make(chan apply),
		stopped:     make(chan struct{}),
		done:        make(chan struct{}),
	}
}

// apply contains entries, snapshot be applied.
// After applied all the items, the application needs
// to send notification to done chan.
type apply struct {
	entries  []raftpb.Entry
	snapshot raftpb.Snapshot
	done     chan struct{}
}

type raftNode struct {
	// Cache of the latest raft index and raft term the server has seen.
	// These three unit64 fields must be the first elements to keep 64-bit
	// alignment for atomic access to the fields.
	index uint64
	term  uint64
	lead  uint64
	hash  uint32

	// last lead elected time
	lt time.Time

	// node raft.Node
	raft.Node

	cfg raft.Config
	mu  sync.Mutex

	lookup NodeLookup
	// ids     map[uint64]*memberlist.Node
	removed map[uint64]bool

	idgen *idutil.Generator
	w     wait.Wait

	// utility
	ticker      <-chan time.Time
	raftStorage *raft.MemoryStorage
	storage     Storage

	// transport specifies the transport to send and receive msgs to members.
	// Sending messages MUST NOT block. It is okay to drop messages, since
	// clients should timeout and reissue their messages.
	// If transport is nil, server will panic.
	transport Transporter

	// a chan to send out apply
	applyc  chan apply
	stopped chan struct{}
	done    chan struct{}
}

// NotifyJoin is invoked when a node is detected to have joined.
// The Node argument must not be modified.
func (r raftNode) NotifyJoin(other *memberlist.Node) {
	r.mu.Lock()
	log.Printf("[Raft] NotifyJoin: %s", other.Name)
	if id, err := xbinary.LittleEndian.Uint32(other.Meta, 0); err == nil {
		r.lookup.SetNode(uint64(id), other)

		// Don't connect to yourself
		if id != r.hash {
			r.transport.AddRemote(uint64(id), other)
		}

		cc := raftpb.ConfChange{
			Type:    raftpb.ConfChangeAddNode,
			NodeID:  uint64(id),
			Context: other.Meta,
		}
		ctx, _ := context.WithTimeout(context.Background(), time.Second)
		err := r.configure(ctx, cc)
		log.Printf("[Raft] NotifyJoin result: %s", err.Error())
	}
	r.mu.Unlock()
	return
}

// NotifyLeave is invoked when a node is detected to have left.
// The Node argument must not be modified.
func (r raftNode) NotifyLeave(other *memberlist.Node) {
	r.mu.Lock()
	log.Printf("[Raft] NotifyLeave: %s", other.Name)
	if id, err := xbinary.LittleEndian.Uint32(other.Meta, 0); err == nil {
		r.lookup.DeleteNode(uint64(id))

		// Don't connect to yourself
		if id != r.hash {
			r.transport.RemoveRemote(uint64(id), other)
		}

		r.removed[uint64(id)] = true

		cc := raftpb.ConfChange{
			Type:   raftpb.ConfChangeRemoveNode,
			NodeID: uint64(id),
		}
		ctx, _ := context.WithTimeout(context.Background(), time.Second)
		r.configure(ctx, cc)
	}
	r.mu.Unlock()
}

// NotifyUpdate is invoked when a node is detected to have
// updated, usually involving the meta data. The Node argument
// must not be modified.
func (r raftNode) NotifyUpdate(other *memberlist.Node) {
	r.mu.Lock()
	log.Printf("[Raft] NotifyUpdate: %s", other.Name)
	if id, err := xbinary.LittleEndian.Uint32(other.Meta, 0); err == nil {
		r.lookup.SetNode(uint64(id), other)

		// Don't connect to yourself
		if id != r.hash {
			r.transport.UpdateRemote(uint64(id), other)
		}

		cc := raftpb.ConfChange{
			Type:    raftpb.ConfChangeUpdateNode,
			NodeID:  uint64(id),
			Context: other.Meta,
		}
		ctx, _ := context.WithTimeout(context.Background(), time.Second)
		r.configure(ctx, cc)
	}
	r.mu.Unlock()
}

// func (r *raftNode) SetTransport(tr Transporter) {
// 	r.transport = tr
// }

// configure sends a configuration change through consensus and
// then waits for it to be applied to the server. It
// will block until the change is performed or there is an error.
func (r *raftNode) configure(ctx context.Context, cc raftpb.ConfChange) error {
	cc.ID = r.idgen.Next()
	ch := r.w.Register(cc.ID)
	start := time.Now()
	if err := r.ProposeConfChange(ctx, cc); err != nil {
		// if err := r.node.ProposeConfChange(ctx, cc); err != nil {
		r.w.Trigger(cc.ID, nil)
		return err
	}
	select {
	case x := <-ch:
		if err, ok := x.(error); ok {
			return err
		}
		if x != nil {
			return fmt.Errorf("return type should always be error")
		}
		return nil
	case <-ctx.Done():
		r.w.Trigger(cc.ID, nil) // GC wait
		return r.parseProposeCtxErr(ctx.Err(), start)
	case <-r.done:
		return raft.ErrStopped
	}
}

func (r *raftNode) parseProposeCtxErr(err error, start time.Time) error {
	switch err {
	case context.Canceled:
		return ErrCanceled
	case context.DeadlineExceeded:
		curLeadElected := r.leadElectedTime()
		prevLeadLost := curLeadElected.Add(-2 * time.Duration(r.cfg.ElectionTick) * time.Duration(1) * time.Millisecond)
		if start.After(prevLeadLost) && start.Before(curLeadElected) {
			return ErrTimeoutDueToLeaderFail
		}
		return ErrTimeout
	default:
		return err
	}
}

func (r *raftNode) leadElectedTime() time.Time {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.lt
}

// start prepares and starts raftNode in a new goroutine. It is no longer safe
// to modify the fields after it has been started.
// TODO: Ideally raftNode should get rid of the passed in server structure.
func (r *raftNode) Start() {
	// r.applyc = make(chan apply)
	// r.stopped = make(chan struct{})
	// r.done = make(chan struct{})

	// go func() {
	var syncC <-chan time.Time

	defer r.onStop()
	for {
		select {
		case <-r.ticker:
			log.Printf("[raft] Tick")
			r.Tick()
			// r.node.Tick()
		case rd := <-r.Ready():
			// case rd := <-r.node.Ready():
			log.Printf("[raft] Ready: %#v", rd)
			if rd.SoftState != nil {
				if lead := atomic.LoadUint64(&r.lead); rd.SoftState.Lead != raft.None && lead != rd.SoftState.Lead {
					r.mu.Lock()
					r.lt = time.Now()
					r.mu.Unlock()
				}
				atomic.StoreUint64(&r.lead, rd.SoftState.Lead)
				if rd.RaftState == raft.StateLeader {
					log.Printf("[raft] Elected Leader")
					syncC = time.Tick(500 * time.Millisecond)
					// TODO: remove the nil checking
					// current test utility does not provide the stats
					// if r.s.stats != nil {
					// 	r.s.stats.BecomeLeader()
					// }
				} else {
					syncC = nil
				}
			}

			apply := apply{
				entries:  rd.CommittedEntries,
				snapshot: rd.Snapshot,
				done:     make(chan struct{}),
			}

			select {
			case r.applyc <- apply:
			case <-r.stopped:
				log.Printf("[raft] Stopped")
				return
			}

			if !raft.IsEmptySnap(rd.Snapshot) {
				if err := r.storage.SaveSnap(rd.Snapshot); err != nil {
					// plog.Fatalf("raft save snapshot error: %v", err)
				}
				r.raftStorage.ApplySnapshot(rd.Snapshot)
				// plog.Infof("raft applied incoming snapshot at index %d", rd.Snapshot.Metadata.Index)
			}
			if err := r.storage.Save(rd.HardState, rd.Entries); err != nil {
				// plog.Fatalf("raft save state and entries error: %v", err)
			}
			r.raftStorage.Append(rd.Entries)
			r.send(rd.Messages)

			select {
			case <-apply.done:
			case <-r.stopped:
				log.Printf("[raft] Stopped")
				return
			}
			r.Advance()
			// r.node.Advance()
		case <-syncC:
			log.Printf("[raft] Sync")
			// r.sync(r.s.cfg.ReqTimeout())
		case <-r.stopped:
			log.Printf("[raft] Stopped")
			return
		}

		log.Println("[raft] Inside loop")
	}
	// }()
	log.Println("[raft] OH NO!! RAFT STOPPED!!!")
}

func (r *raftNode) apply() chan apply {
	return r.applyc
}

func (r *raftNode) onStop() {
	r.Stop()
	// r.node.Stop()
	r.transport.Stop()
	if err := r.storage.Close(); err != nil {
		// plog.Panicf("raft close storage error: %v", err)
	}
	close(r.done)
}

// send relays all the messages to the transport
func (r *raftNode) send(ms []raftpb.Message) {
	for i := range ms {

		// The etcd implementation sets the msg.To field to 0 is the node has been removed.
		// I'm currently not sure if the transport is supposed to ignore that message or not based on the To field.
		if _, ok := r.removed[ms[i].To]; ok {
			ms[i].To = 0
		}
	}

	// Send messages over transport
	r.transport.Send(ms)
}

func (r *raftNode) Step(msg raftpb.Message) {
	// ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	ctx := context.TODO()
	r.Node.Step(ctx, msg)
	// r.node.Step(ctx, msg)
	// cancel()
}

func (r *raftNode) Propose(data []byte) (err error) {
	log.Println("Proposing data")
	err = r.propose(data)

	// ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	// id := r.idgen.Next()
	// ch := r.w.Register(uint64(id))
	// ctx := context.TODO()
	// r.node.Propose(ctx, data)

	// select {
	// case <-ch:
	// 	err = ctx.Err()
	// case <-ctx.Done():
	// 	r.w.Trigger(uint64(id), nil) // GC wait
	// 	err = ctx.Err()
	// case <-r.done:
	// 	err = raft.ErrStopped
	// }

	log.Println("Proposal result:", err)

	// cancel()
	return

	// for {
	// 	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	// 	err := r.node.Propose(ctx, data)
	// 	if err != nil {
	// 		log.Printf("[raft] Proposal Error: %s", err)
	// 	} else {
	// 		log.Printf("[raft] Proposal Success")
	// 	}
	// 	cancel()
	// 	return err
	// }
}

func (r *raftNode) propose(data []byte) (err error) {
	// ctx, _ := context.WithTimeout(context.Background(), time.Second)
	ctx := context.Background()
	id := r.idgen.Next()
	ch := r.w.Register(id)
	start := time.Now()
	if err := r.Node.Propose(ctx, data); err != nil {
		// if err := r.node.Propose(ctx, data); err != nil {
		r.w.Trigger(id, nil)
		return err
	}
	select {
	case x := <-ch:
		if err, ok := x.(error); ok {
			return err
		}
		if x != nil {
			return fmt.Errorf("return type should always be error")
		}
		return nil
	case <-ctx.Done():
		r.w.Trigger(id, nil) // GC wait
		return r.parseProposeCtxErr(ctx.Err(), start)
	case <-r.done:
		return raft.ErrStopped
	}
}

// // sync proposes a SYNC request and is non-blocking.
// // This makes no guarantee that the request will be proposed or performed.
// // The request will be cancelled after the given timeout.
// func (r *raftNode) sync(timeout time.Duration) {
// 	ctx, cancel := context.WithTimeout(context.Background(), timeout)
// 	req := pb.Request{
// 		Method: "SYNC",
// 		ID:     r.idgen.Next(),
// 		Time:   time.Now().UnixNano(),
// 	}
// 	data := pbutil.MustMarshal(&req)
// 	// There is no promise that node has leader when do SYNC request,
// 	// so it uses goroutine to propose.
// 	go func() {
// 		s.r.Propose(ctx, data)
// 		cancel()
// 	}()
// }
