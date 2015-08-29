package main

import (
	"log"

	"github.com/coreos/etcd/raft/raftpb"
)

type Storage interface {

	// Save function saves ents and state to the underlying stable storage.
	// Save MUST block until st and ents are on stable storage.
	Save(st raftpb.HardState, ents []raftpb.Entry) error

	// SaveSnap function saves snapshot to the underlying stable storage.
	SaveSnap(snap raftpb.Snapshot) error

	// Close closes the Storage and performs finalization.
	Close() error
}

type NoopStorage struct {
}

// Save function saves ents and state to the underlying stable storage.
// Save MUST block until st and ents are on stable storage.
func (n *NoopStorage) Save(st raftpb.HardState, ents []raftpb.Entry) error {
	log.Printf("[NoopStorage] Saving Term: %d Vote: %d Commit %d Entries: %d", st.Term, st.Vote, st.Commit, len(ents))
	return nil
}

// SaveSnap function saves snapshot to the underlying stable storage.
func (n *NoopStorage) SaveSnap(snap raftpb.Snapshot) error {
	log.Println("[NoopStorage] Saving snapshot")
	return nil
}

// Close closes the Storage and performs finalization.
func (n *NoopStorage) Close() error {
	log.Println("[NoopStorage] Closing")
	return nil
}
