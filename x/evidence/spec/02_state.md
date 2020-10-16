<!--
order: 2
-->

# State

Currently the `x/evidence` module only stores valid submitted `Evidence` in state.
The evidence state is also stored and exported in the `x/evidence` module's `GenesisState`.

```protobuf
// GenesisState defines the evidence module's genesis state.
message GenesisState {
  // evidence defines all the evidence at genesis.
  repeated google.protobuf.Any evidence = 1;
}

```

All `Evidence` is retrieved and stored via a prefix `KVStore` using prefix `0x00` (`KeyPrefixEvidence`).
