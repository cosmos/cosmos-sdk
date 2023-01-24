# ADR 64: ABCI 2.0 Integration (Phase II)

## Changelog

* 2023-01-17: Initial Draft (@alexanderbez)

## Status

PROPOSED

## Abstract

This ADR outlines the continuation of the efforts to implement ABCI++ in the Cosmos
SDK outlined in [ADR 060: ABCI 1.0 (Phase I)](adr-060-abci-1.0.md).

Specifically, this ADR outlines the design and implementation of ABCI 2.0, which
includes `ExtendVote`, `VerifyVoteExtension` and `FinalizeBlock`.

## Context

ABCI 2.0 continues the promised updates from ABCI++, specifically three additional
ABCI methods that the application can implement in order to gain further control,
insight and customization of the consensus process, unlocking many novel use-cases
that previously not possible. We describe these three new methods below:

### `ExtendVote`

This method allows each validator process to extend the pre-commit phase of the
Tendermint consensus process. Specifically, it allows the application to perform
custom business logic that extends the pre-commit vote and supply additional data
as part of the vote.

The data, called vote extension, will be broadcast and received together with the
vote it is extending, and will be made available to the application in the next
height. Specifically, the proposer of the next block will receive the vote extensions
in `RequestPrepareProposal.local_last_commit.votes`.

If the application does not have vote extension information to provide,
it returns a 0-length byte array as its vote extension.

**NOTE**: 

* Although each validator process submits its own vote extension, ONLY the *proposer*
  of the *next* block will receive all the vote extensions included as part of the
  pre-commit phase of the previous block. This means only the proposer will
  implicitly have access to all the vote extensions, via `RequestPrepareProposal`,
  and that not all vote extensions may be included, since a validator does not
  have to wait for all pre-commits.
* The pre-commit vote is signed independently from the vote extension.

### `VerifyVoteExtension`

This method allows validators to validate the vote extension data attached to
each pre-commit message it receives. If the validation fails, the whole pre-commit
message will be deemed invalid and ignored by Tendermint.

Tendermint uses `VerifyVoteExtension` when validating a pre-commit vote. Specifically,
for a pre-commit, Tendermint will:

* Reject the message if it doesn't contain a signed vote AND a signed vote extension
* Reject the message if the vote's signature OR the vote extension's signature fails to verify
* Reject the message if `VerifyVoteExtension` was rejected

Otherwise, Tendermint will accept the pre-commit message.

Note, this has important consequences on liveness, i.e., if vote extensions repeatedly
cannot be verified by correct validators, Tendermint may not be able to finalize
a block even if sufficiently many (+2/3) validators send pre-commit votes for
that block. Thus, `VerifyVoteExtension` should be used with special care.

Tendermint recommends that an application that detects an invalid vote extension
SHOULD accept it in `ResponseVerifyVoteExtension` and ignore it in its own logic.

### `FinalizeBlock`

This method delivers a decided block to the application. The application must
execute the transactions in the block deterministically and update its state
accordingly. Cryptographic commitments to the block and transaction results,
returned via the corresponding parameters in `ResponseFinalizeBlock`, are
included in the header of the next block. Tendermint calls it when a new block
is decided.

In other words, `FinalizeBlock` encapsulates the current ABCI execution flow of
`BeginBlock`, one or more `DeliverTx`, and `EndBlock` into a single ABCI method.
Tendermint will no longer execute requests for these legacy methods and instead
will just simply call `FinalizeBlock`.

## Decision

We will discuss changes to the Cosmos SDK to implement ABCI 2.0 in two distinct
phases, `VoteExtensions` and `FinalizeBlock`.

### `VoteExtensions`

### `FinalizeBlock`

The existing ABCI methods `BeginBlock`, `DeliverTx`, and `EndBlock` have existed
since the dawn of ABCI-based applications. Thus, applications, tooling, and developers
have grown used to these methods and their use-cases. Specifically, `BeginBlock`
and `EndBlock` have grown to be pretty integral and powerful within ABCI-based
applications. E.g. an application might want to run distribution and inflation
related operations prior to executing transactions and then have staking related
changes to happen after executing all transactions.

We propose to keep `BeginBlock` and `EndBlock` within the SDK's core module
interfaces only so application developers can continue to build against existing
execution flows. However, we will remove `BeginBlock`, `DeliverTx` and `EndBlock`
from the SDK's `BaseApp` implementation and thus the ABCI surface area.

What will exist is a single `FinalizeBlock` execution flow. Specifically, in
`FinalizeBlock` we will execute the application's `BeginBlock`, followed by
execution of all the transactions, finally followed by execution of the application's
`EndBlock`.

Note, we will still keep the existing transaction execution mechanics within
`BaseApp`, but all notion of `DeliverTx` will be removed, i.e. `deliverState`
will be replace with `finalizeState`, which will be committed on `Commit`.

However, there are current parameters and fields that exist in the existing
`BeginBlock` and `EndBlock` ABCI types, such as votes that are used in distribution
and byzantine validators used in evidence handling. These parameters exist in the
`FinalizeBlock` request type, and will need to be passed to the application's
implementations of `BeginBlock` and `EndBlock`.

This means the Cosmos SDK's core module interfaces will need to be updated to
reflect these parameters. The easiest and most straightforward way to achieve
this is to just pass `RequestFinalizeBlock` to `BeginBlock` and `EndBlock`.
Alternatively, we can create dedicated proxy types in the SDK that reflect these
legacy ABCI types, e.g. `LegacyBeginBlockRequest` and `LegacyEndBlockRequest`. Or,
we can come up with new types and names altogether.

```go
func (app *BaseApp) FinalizeBlock(req abci.RequestFinalizeBlock) abci.ResponseFinalizeBlock {
	beginBlockResp := app.beginBlock(ctx, req)

	txExecResults := make([]abci.ExecTxResult, 0, len(req.Txs))
	for _, tx := range req.Txs {
		result := app.runTx(runTxModeFinalize, tx)
		txExecResults = append(txExecResults, result)
	}

	endBlockResp := app.endBlock(ctx, req)

	return abci.ResponseFinalizeBlock{
		TxResults:             txExecResults,
		Events:                joinEvents(beginBlockResp.Events, endBlockResp.Events),
		ValidatorUpdates:      endBlockResp.ValidatorUpdates,
		ConsensusParamUpdates: endBlockResp.ConsensusParamUpdates,
		AppHash:               nil,
	}
}
```

## Consequences

### Backwards Compatibility

ABCI 2.0 is naturally not backwards compatible with prior versions of the Cosmos SDK
and Tendermint. For example, an application that requests `RequestFinalizeBlock`
to the same application that does not speak ABCI 2.0 will naturally fail.

In addition, `BeginBlock`, `DeliverTx` and `EndBlock` will be removed from the
application ABCI interfaces and along with the inputs and outputs being modified
in the module interfaces.

### Positive

* `BeginBlock` and `EndBlock` semantics remain, so burden on application developers
  should be limited.
* Less communication overhead as multiple ABCI requests are condensed into a single
  request.
* Sets the groundwork for optimistic execution.

### Negative

### Neutral

## Further Discussions

Future discussions include design and implementation of ABCI 3.0, which is a
continuation of ABCI++ and the general discussion of optimistic execution.

## References

* [ADR 060: ABCI 1.0 (Phase I)](adr-060-abci-1.0.md)
