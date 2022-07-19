<!--
order: 2
-->

# State
## In persisted KV store

1. Global unique capability index
2. Capability owners

Indexes:

* Unique index: `[]byte("index") -> []byte(currentGlobalIndex)`
* Capability Index: `[]byte("capability_index") | []byte(index) -> ProtocolBuffer(CapabilityOwners)`

## In-memory KV store

1. Initialized flag
2. Mapping between the module and capability tuple and the capability name
3. Mapping between the module and capability name and its index

Indexes:

* Initialized flag: `[]byte("mem_initialized")`
* RevCapabilityKey: `[]byte(moduleName + "/rev/" + capabilityName) -> []byte(index)`
* FwdCapabilityKey: `[]byte(moduleName + "/fwd/" + capabilityPointerAddress) -> []byte(capabilityName)`
