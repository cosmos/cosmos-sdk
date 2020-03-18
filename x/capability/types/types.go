package types

import "fmt"

var _ Capability = (*CapabilityKey)(nil)

// Capability defines the interface a capability must implement. The given
// capability must provide a GUID.
type Capability interface {
	GetIndex() uint64
	String() string
}

// CapabilityKey defines an implementation of a Capability. The index provided to
// a CapabilityKey must be globally unique.
type CapabilityKey struct {
	Name  string
	Index uint64
}

func NewCapabilityKey(name string, index uint64) Capability {
	return &CapabilityKey{Name: name, Index: index}
}

// GetIndex returns the capability key's index.
func (ck *CapabilityKey) GetIndex() uint64 {
	return ck.Index
}

func (ck *CapabilityKey) String() string {
	return fmt.Sprintf("CapabilityKey{%p, %d, %s}", ck, ck.Index, ck.Name)
}
