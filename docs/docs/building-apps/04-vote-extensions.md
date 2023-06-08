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
validator process. The Cosmos SDK defines `ExtendVoteHandler`:

```go
type ExtendVoteHandler func(Context, *abci.RequestExtendVote) (*abci.ResponseExtendVote, error)
```

An application can set this handler in `app.go` via the `SetExtendVoteHandler`
`BaseApp` option function. The `ExtendVoteHandler`, if defined, is called during
the `ExtendVote` ABCI method. Note, if an application decides to implement
`ExtendVoteHandler`, it MUST return a non-nil `VoteExtension`. However, the vote
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
this process MUST be deterministic. The Cosmos SDK defines `VerifyVoteExtensionHandler`:

```go
type VerifyVoteExtensionHandler func(Context, *abci.RequestVerifyVoteExtension) (*abci.ResponseVerifyVoteExtension, error)
```

An application can set this handler in `app.go` via the `SetVerifyVoteExtensionHandler`
`BaseApp` option function. The `VerifyVoteExtensionHandler`, if defined, is called
during the `VerifyVoteExtension` ABCI method. If an application defines a vote
extension handler, it should also define a verification handler. Note, not all
validators will share the same view of what vote extensions they verify depending
on how votes are propagated. See [here](https://github.com/cometbft/cometbft/blob/v0.38.0-rc1/spec/abci/abci++_methods.md#verifyvoteextension)
for more details.

## Vote Extension Propagation
