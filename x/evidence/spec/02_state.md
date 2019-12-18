<!--
order: 2
-->

# State

Currently the `x/evidence` module only stores valid submitted `Evidence` in state.
The evidence state is also stored and exported in the `x/evidence` module's `GenesisState`.

```go
type GenesisState struct {
  Evidence []Evidence `json:"evidence" yaml:"evidence"`
}
```

All `Evidence` is retrieved and stored via a prefix `KVStore` using prefix `0x00` (`KeyPrefixEvidence`).
