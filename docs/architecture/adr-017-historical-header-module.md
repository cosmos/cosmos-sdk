# ADR 17: Historical Header Module

## Changelog

- 26 November 2019: Start of first version
- 2 December 2019: Final draft of first version

## Context

In order for the Cosmos SDK to implement the [IBC specification](https://github.com/cosmos/ics), modules within the SDK must have the ability to introspect recent consensus states (validator sets & commitment roots) as proofs of these values on other chains must be checked during the handshakes.

## Decision

The application MUST instruct Tendermint to provide requisite information for the application to track the past `n` headers by indicating as such in a field on `abci.ResponseInitChain` containing the value of `n`.
- The value of this field MAY default to `0` for ease of backwards compatibility with other ABCI clients which do not require this header-introspection functionality.

Tendermint MUST read this field when handling the `abci.ResponseInitChain` response, and then behave as follows:
- With the first `abci.RequestBeginBlock`, Tendermint MUST send the `n` most recent committed headers in order.
  - If fewer than `n` previous headers exist, Tendermint MUST send all previous headers.
- With subsequent `abci.RequestBeginBlock` calls, Tendermint MUST send the the most recent committed header.

The application MUST use this header data provided by `abci.RequestBeginBlock` to track the past `n` committed headers in the `BaseApp`:
- When handling the first `abci.RequestBeginBlock` invocation, the application must store all `n` committed headers in memory.
- When handling subsequent `abci.RequestBeginBlock` invocations, the application must remove the oldest commmited header from memory and store the most recent one (provided as a field of `abci.RequestBeginBlock`).

The application MUST make these past `n` committed headers available for querying by SDK modules through the `sdk.Context` as follows:

```golang
func (c Context) PreviousHeader(height uint64) tmtypes.Header {
  // implemented in the context
}
```

## Status

Proposed.

## Consequences

Implementation of this ADR will require synchronised changes to Tendermint & the Cosmos SDK. It should be backwards-compatible with other ABCI clients, which can simply elect to ignore the field.

### Positive

- Easy retrieval of headers & state roots for recent past heights by modules anywhere in the SDK
- Maintains existing deliver-only interface (no "backwards queries" from the SDK to Tendermint)

### Negative

- Additional memory usage (to store `n` headers)
- Additional state tracking in Tendermint & the SDK

### Neutral

(none known)

## References

- [ICS 2: "Consensus state introspection"](https://github.com/cosmos/ics/tree/master/spec/ics-024-host-requirements#consensus-state-introspection)
