---
sidebar_position: 1
---

# Vote Extensions

:::note Synopsis
This sections describes how the application can define and use vote extensions
defined in ABCI++.
:::

## Extend Vote

ABCI++ allows an application to extend a pre-commit vote with arbitrary data. This
process does NOT have be deterministic and the data returned can be unique to the
validator process. The Cosmos SDK defines `baseapp.ExtendVoteHandler`:

```go
type ExtendVoteHandler func(Context, *abci.RequestExtendVote) (*abci.ResponseExtendVote, error)
```

An application can set this handler in `app.go` via the `baseapp.SetExtendVoteHandler`
`BaseApp` option function. The `sdk.ExtendVoteHandler`, if defined, is called during
the `ExtendVote` ABCI method. Note, if an application decides to implement
`baseapp.ExtendVoteHandler`, it MUST return a non-nil `VoteExtension`. However, the vote
extension can be empty. See [here](https://github.com/cometbft/cometbft/blob/v0.38.0-rc1/spec/abci/abci++_methods.md#extendvote)
for more details.

There are many decentralized censorship-resistant use cases for vote extensions.
For example, a validator may want to submit prices for a price oracle or encryption
shares for an encrypted transaction mempool. Note, an application should be careful
to consider the size of the vote extensions as they could increase latency in block
production. See [here](https://github.com/cometbft/cometbft/blob/v0.38.0-rc1/docs/qa/CometBFT-QA-38.md#vote-extensions-testbed)
for more details.

## Verify Vote Extension

Similar to extending a vote, an application can also verify vote extensions from
other validators when validating their pre-commits. For a given vote extension,
this process MUST be deterministic. The Cosmos SDK defines `sdk.VerifyVoteExtensionHandler`:

```go reference
https://github.com/cosmos/cosmos-sdk/blob/v0.50.0-alpha.0/types/abci.go#L26-L27
```

An application can set this handler in `app.go` via the `baseapp.SetVerifyVoteExtensionHandler`
`BaseApp` option function. The `sdk.VerifyVoteExtensionHandler`, if defined, is called
during the `VerifyVoteExtension` ABCI method. If an application defines a vote
extension handler, it should also define a verification handler. Note, not all
validators will share the same view of what vote extensions they verify depending
on how votes are propagated. See [here](https://github.com/cometbft/cometbft/blob/v0.38.0-rc1/spec/abci/abci++_methods.md#verifyvoteextension)
for more details.

## Vote Extension Propagation

The agreed upon vote extensions at height `H` are provided to the proposing validator
at height `H+1` during `PrepareProposal`. As a result, the vote extensions are
not natively provided or exposed to the remaining validators during `ProcessProposal`.
As a result, if an application requires that the agreed upon vote extensions from
height `H` are available to all validators at `H+1`, the application must propagate
these vote extensions manually in the block proposal itself. This can be done by
"injecting" them into the block proposal, since the `Txs` field in `PrepareProposal`
is just a slice of byte slices.

`FinalizeBlock` will ignore any byte slice that doesn't implement an `sdk.Tx` so
any injected vote extensions will safely be ignored in `FinalizeBlock`. For more
details on propagation, see the [ABCI++ 2.0 ADR](https://github.com/cosmos/cosmos-sdk/blob/main/docs/architecture/adr-064-abci-2.0.md#vote-extension-propagation--verification).
