package types

// CapabilityStore defines an ephemeral in-memory object capability store.
type CapabilityStore struct {
	revMemStore map[string]*Capability
	fwdMemStore map[string]string
}

func NewCapabilityStore() CapabilityStore {
	return CapabilityStore{
		revMemStore: make(map[string]*Capability),
		fwdMemStore: make(map[string]string),
	}
}

// GetCapability returns a Capability by module and name tuple. If no Capability
// exists, nil will be returned.
func (cs CapabilityStore) GetCapability(module, name string) *Capability {
	key := RevCapabilityKey(module, name)
	return cs.revMemStore[string(key)]
}

// GetCapabilityName returns a Capability name by module and Capability tuple. If
// no Capability name exists for the given tuple, an empty string is returned.
func (cs CapabilityStore) GetCapabilityName(module string, cap *Capability) string {
	key := FwdCapabilityKey(module, cap)
	return cs.fwdMemStore[string(key)]
}

// SetCapability sets the reverse mapping between the module and capability name
// and the capability in the in-memory store.
func (cs CapabilityStore) SetCapability(module, name string, cap *Capability) {
	key := RevCapabilityKey(module, name)
	cs.revMemStore[string(key)] = cap
}

// SetCapabilityName sets the forward mapping between the module and capability
// tuple and the capability name in the in-memory store.
func (cs CapabilityStore) SetCapabilityName(module, name string, cap *Capability) {
	key := FwdCapabilityKey(module, cap)
	cs.fwdMemStore[string(key)] = name
}

// DeleteCapability removes the reverse mapping between the module and capability
// name and the capability in the in-memory store.
func (cs CapabilityStore) DeleteCapability(module, name string) {
	key := RevCapabilityKey(module, name)
	delete(cs.revMemStore, string(key))
}

// DeleteCapabilityName removes the forward mapping between the module and capability
// tuple and the capability name in the in-memory store.
func (cs CapabilityStore) DeleteCapabilityName(module string, cap *Capability) {
	key := FwdCapabilityKey(module, cap)
	delete(cs.fwdMemStore, string(key))
}
