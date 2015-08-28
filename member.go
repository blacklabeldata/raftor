package raftor

// Member describes a node in the cluster
type Member interface {
	ID() uint64
	Meta() []byte
}

// NewMember creates a new member type
func NewMember(id uint64, meta []byte) Member {
	return member{id, meta}
}

// member implements the Member interface
type member struct {
	id   uint64
	meta []byte
}

// Id returns the member's unique id
func (m member) ID() uint64 {
	return m.id
}

// Meta returns the nodes metadata
func (m member) Meta() []byte {
	return m.meta
}
