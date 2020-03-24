package types

import (
	"fmt"
	"sort"

	"gopkg.in/yaml.v2"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

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
	Index uint64 `json:"index" yaml:"index"`
}

// NewCapabilityKey returns a reference to a new CapabilityKey to be used as an
// actual capability.
func NewCapabilityKey(index uint64) Capability {
	return &CapabilityKey{Index: index}
}

// GetIndex returns the capability key's index.
func (ck *CapabilityKey) GetIndex() uint64 {
	return ck.Index
}

// String returns the string representation of a CapabilityKey. The string contains
// the CapabilityKey's memory reference as the string is to be used in a composite
// key and to authenticate capabilities.
func (ck *CapabilityKey) String() string {
	return fmt.Sprintf("CapabilityKey{%p, %d}", ck, ck.Index)
}

// Owner defines a single capability owner. An owner is defined by the name of
// capability and the module name.
type Owner struct {
	Module string `json:"module" yaml:"module"`
	Name   string `json:"name" yaml:"name"`
}

func NewOwner(module, name string) Owner {
	return Owner{Module: module, Name: name}
}

// Key returns a composite key for an Owner.
func (o Owner) Key() string {
	return fmt.Sprintf("%s/%s", o.Module, o.Name)
}

func (o Owner) String() string {
	bz, _ := yaml.Marshal(o)
	return string(bz)
}

// CapabilityOwners defines a set of owners of a single Capability. The set of
// owners must be unique.
type CapabilityOwners struct {
	Owners []Owner `json:"owners" yaml:"owners"`
}

func NewCapabilityOwners() *CapabilityOwners {
	return &CapabilityOwners{Owners: make([]Owner, 0)}
}

// Set attempts to add a given owner to the CapabilityOwners. If the owner
// already exists, an error will be returned. Set runs in O(log n) average time
// and O(n) in the worst case.
func (co *CapabilityOwners) Set(owner Owner) error {
	// find smallest index s.t. co.Owners[i] >= owner in O(log n) time
	i := sort.Search(len(co.Owners), func(i int) bool { return co.Owners[i].Key() >= owner.Key() })
	if i < len(co.Owners) && co.Owners[i].Key() == owner.Key() {
		// owner already exists at co.Owners[i]
		return sdkerrors.Wrapf(ErrOwnerClaimed, owner.String())
	}

	// owner does not exist in the set of owners, so we insert at position i
	co.Owners = append(co.Owners, Owner{}) // expand by 1 in amortized O(1) / O(n) worst case
	copy(co.Owners[i+1:], co.Owners[i:])
	co.Owners[i] = owner

	return nil
}
