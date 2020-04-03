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

// NewCapabilityKey returns a reference to a new CapabilityKey to be used as an
// actual capability.
func NewCapabilityKey(index uint64) Capability {
	return &CapabilityKey{Index: index}
}

// String returns the string representation of a CapabilityKey. The string contains
// the CapabilityKey's memory reference as the string is to be used in a composite
// key and to authenticate capabilities.
func (ck *CapabilityKey) String() string {
	return fmt.Sprintf("CapabilityKey{%p, %d}", ck, ck.Index)
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

func NewCapabilityOwners() *CapabilityOwners {
	return &CapabilityOwners{Owners: make([]Owner, 0)}
}

// Set attempts to add a given owner to the CapabilityOwners. If the owner
// already exists, an error will be returned. Set runs in O(log n) average time
// and O(n) in the worst case.
func (co *CapabilityOwners) Set(owner Owner) error {
	i, ok := co.Get(owner)
	if ok {
		// owner already exists at co.Owners[i]
		return sdkerrors.Wrapf(ErrOwnerClaimed, owner.String())
	}

	// owner does not exist in the set of owners, so we insert at position i
	co.Owners = append(co.Owners, Owner{}) // expand by 1 in amortized O(1) / O(n) worst case
	copy(co.Owners[i+1:], co.Owners[i:])
	co.Owners[i] = owner

	return nil
}

// Remove removes a provided owner from the CapabilityOwners if it exists. If the
// owner does not exist, Remove is considered a no-op.
func (co *CapabilityOwners) Remove(owner Owner) {
	if len(co.Owners) == 0 {
		return
	}

	i, ok := co.Get(owner)
	if ok {
		// owner exists at co.Owners[i]
		co.Owners = append(co.Owners[:i], co.Owners[i+1:]...)
	}
}

// Get returns (i, true) of the provided owner in the CapabilityOwners if the
// owner exists, where i indicates the owner's index in the set. Otherwise
// (i, false) where i indicates where in the set the owner should be added.
func (co *CapabilityOwners) Get(owner Owner) (int, bool) {
	// find smallest index s.t. co.Owners[i] >= owner in O(log n) time
	i := sort.Search(len(co.Owners), func(i int) bool { return co.Owners[i].Key() >= owner.Key() })
	if i < len(co.Owners) && co.Owners[i].Key() == owner.Key() {
		// owner exists at co.Owners[i]
		return i, true
	}

	return i, false
}
