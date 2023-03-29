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
CometBFT consensus process. Specifically, it allows the application to perform
custom business logic that extends the pre-commit vote and supply additional data
as part of the vote, although they are signed separately by the same key.

The data, called vote extension, will be broadcast and received together with the
vote it is extending, and will be made available to the application in the next
height. Specifically, the proposer of the next block will receive the vote extensions
in `RequestPrepareProposal.local_last_commit.votes`.

If the application does not have vote extension information to provide, it
returns a 0-length byte array as its vote extension.

**NOTE**: 

* Although each validator process submits its own vote extension, ONLY the *proposer*
  of the *next* block will receive all the vote extensions included as part of the
  pre-commit phase of the previous block. This means only the proposer will
  implicitly have access to all the vote extensions, via `RequestPrepareProposal`,
  and that not all vote extensions may be included, since a validator does not
  have to wait for all pre-commits, only 2/3.
* The pre-commit vote is signed independently from the vote extension.

### `VerifyVoteExtension`

This method allows validators to validate the vote extension data attached to
each pre-commit message it receives. If the validation fails, the whole pre-commit
message will be deemed invalid and ignored by CometBFT.

CometBFT uses `VerifyVoteExtension` when validating a pre-commit vote. Specifically,
for a pre-commit, CometBFT will:

* Reject the message if it doesn't contain a signed vote AND a signed vote extension
* Reject the message if the vote's signature OR the vote extension's signature fails to verify
* Reject the message if `VerifyVoteExtension` was rejected by the app

Otherwise, CometBFT will accept the pre-commit message.

Note, this has important consequences on liveness, i.e., if vote extensions repeatedly
cannot be verified by correct validators, CometBFT may not be able to finalize
a block even if sufficiently many (+2/3) validators send pre-commit votes for
that block. Thus, `VerifyVoteExtension` should be used with special care.

CometBFT recommends that an application that detects an invalid vote extension
SHOULD accept it in `ResponseVerifyVoteExtension` and ignore it in its own logic.

### `FinalizeBlock`

This method delivers a decided block to the application. The application must
execute the transactions in the block deterministically and update its state
accordingly. Cryptographic commitments to the block and transaction results,
returned via the corresponding parameters in `ResponseFinalizeBlock`, are
included in the header of the next block. CometBFT calls it when a new block
is decided.

In other words, `FinalizeBlock` encapsulates the current ABCI execution flow of
`BeginBlock`, one or more `DeliverTx`, and `EndBlock` into a single ABCI method.
CometBFT will no longer execute requests for these legacy methods and instead
will just simply call `FinalizeBlock`.

## Decision

We will discuss changes to the Cosmos SDK to implement ABCI 2.0 in two distinct
phases, `VoteExtensions` and `FinalizeBlock`.

### `VoteExtensions`

Similarly for `PrepareProposal` and `ProcessProposal`, we propose to introduce
two new handlers that an application can implement in order to provide and verify
vote extensions.

We propose the following new handlers for applications to implement:

```go
type ExtendVoteHandler func(sdk.Context, abci.RequestExtendVote) abci.ResponseExtendVote
type VerifyVoteExtensionHandler func(sdk.Context, abci.RequestVerifyVoteExtension) abci.ResponseVerifyVoteExtension
```

A new execution state, `voteExtensionState`, will be introduced and provided as
the `Context` that is supplied to both handlers. It will contain relevant metadata
such as the block height and block hash. Note, `voteExtensionState` is never
committed and will exist as ephemeral state only in the context of a single block.

If an application decides to implement `ExtendVoteHandler`, it must return a
non-nil `ResponseExtendVote.VoteExtension`.

Recall, an implementation of `ExtendVoteHandler` does NOT need to be deterministic,
however, given a set of vote extensions, `VerifyVoteExtensionHandler` must be
deterministic, otherwise the chain may suffer from liveness faults. In addition,
recall CometBFT proceeds in rounds for each height, so if a decision cannot be
made about about a block proposal at a given height, CometBFT will proceed to the
next round and thus will execute `ExtendVote` and `VerifyVoteExtension` again for
the new round for each validator until 2/3 valid pre-commits can be obtained.

Given the broad scope of potential implementations and use-cases of vote extensions,
and how to verify them, most applications should choose to implement the handlers
through a single handler type, which can have any number of dependencies injected
such as keepers. In addition, this handler type could contain some notion of
volatile vote extension state management which would assist in vote extension
verification. This state management could be ephemeral or could be some form of
on-disk persistence.

Example:

```go
// VoteExtensionHandler implements an Oracle vote extension handler.
type VoteExtensionHandler struct {
	cdc   Codec
	mk    MyKeeper
	state VoteExtState // This could be a map or a DB connection object
}

// ExtendVoteHandler can do something with h.mk and possibly h.state to create
// a vote extension, such as fetching a series of prices for supported assets.
func (h VoteExtensionHandler) ExtendVoteHandler(ctx sdk.Context, req abci.RequestExtendVote) abci.ResponseExtendVote {
	prices := GetPrices(ctx, h.mk.Assets())
	bz, err := EncodePrices(h.cdc, prices)
	if err != nil {
		panic(fmt.Errorf("failed to encode prices for vote extension: %w", err))
	}

	// store our vote extension at the given height
	//
	// NOTE: Vote extensions can be overridden since we can timeout in a round.
	SetPrices(h.state, req, bz)

	return abci.ResponseExtendVote{VoteExtension: bz}
}

// VerifyVoteExtensionHandler can do something with h.state and req to verify
// the req.VoteExtension field, such as ensuring the provided oracle prices are
// within some valid range of our prices.
func (h VoteExtensionHandler) VerifyVoteExtensionHandler(ctx sdk.Context, req abci.RequestVerifyVoteExtension) abci.ResponseVerifyVoteExtension {
	prices, err := DecodePrices(h.cdc, req.VoteExtension)
	if err != nil {
		log("failed to decode vote extension", "err", err)
		return abci.ResponseVerifyVoteExtension{Status: REJECT}
	}

	if err := ValidatePrices(h.state, req, prices); err != nil {
		log("failed to validate vote extension", "prices", prices, "err", err)
		return abci.ResponseVerifyVoteExtension{Status: REJECT}
	}

	// store updated vote extensions at the given height
	//
	// NOTE: Vote extensions can be overridden since we can timeout in a round.
	SetPrices(h.state, req, req.VoteExtension)

	return abci.ResponseVerifyVoteExtension{Status: ACCEPT}
}
```

#### Vote Extension Propagation & Verification

As mentioned previously, vote extensions for height `H` are only made available
to the proposer at height `H+1` during `PrepareProposal`. However, in order to
make vote extensions useful, all validators should have access to the agreed upon
vote extensions at height `H` during `H+1`.

Since CometBFT includes all the vote extension signatures in `RequestPrepareProposal`,
we propose that the proposing validator manually "inject" the vote extensions
along with their respective signatures via a special transaction, `VoteExtsTx`,
into the block proposal during `PrepareProposal`. The `VoteExtsTx` will be
populated with a single `ExtendedCommitInfo` object which is received directly
from `RequestPrepareProposal`.

For convention, the `VoteExtsTx` transaction should be the first transaction in
the block proposal, although chains can implement their own preferences. For
safety purposes, we also propose that the proposer itself verify all the vote
extension signatures it receives in `RequestPrepareProposal`.

A validator, upon a `RequestProcessProposal`, will receive the injected `VoteExtsTx`
which includes the vote extensions along with their signatures. If no such transaction
exists, the validator MUST REJECT the proposal.

When a validator inspects a `VoteExtsTx`, it will evaluate each `SignedVoteExtension`.
For each signed vote extension, the validator will generate the signed bytes and
verify the signature. At least 2/3 valid signatures, based on voting power, must
be received in order for the block proposal to be valid, otherwise the validator
MUST REJECT the proposal.

In order to have the ability to validate signatures, `BaseApp` must have access
to the `x/staking` module, since this module stores an index from consensus
address to public key. However, we will avoid a direct dependency on `x/staking`
and instead rely on an interface instead. In addition, the Cosmos SDK will expose
a default signature verification method which applications can use:

```go
type ValidatorStore interface {
	GetValidatorByConsAddr(sdk.Context, cryptotypes.Address) (cryptotypes.PubKey, error)
}

// ValidateVoteExtensions is a function that an application can execute in
// ProcessProposal to verify vote extension signatures.
func (app *BaseApp) ValidateVoteExtensions(ctx sdk.Context, currentHeight int64, extCommit abci.ExtendedCommitInfo) error {
	for _, vote := range extCommit.Votes {
		if !vote.SignedLastBlock || len(vote.VoteExtension) == 0 {
			continue
		}

		valConsAddr := cmtcrypto.Address(vote.Validator.Address)

		validator, err := app.validatorStore.GetValidatorByConsAddr(ctx, valConsAddr)
		if err != nil {
			return fmt.Errorf("failed to get validator %s for vote extension", valConsAddr)
		}

		cmtPubKey, err := validator.CmtConsPublicKey()
		if err != nil {
			return fmt.Errorf("failed to convert public key: %w", err)
		}

		if len(vote.ExtensionSignature) == 0 {
			return fmt.Errorf("received a non-empty vote extension with empty signature for validator %s", valConsAddr)
		}

		cve := cmtproto.CanonicalVoteExtension{
			Extension: vote.VoteExtension,
			Height:    currentHeight - 1, // the vote extension was signed in the previous height
			Round:     int64(extCommit.Round),
			ChainId:   app.GetChainID(),
		}

		extSignBytes, err := cosmosio.MarshalDelimited(&cve)
		if err != nil {
			return fmt.Errorf("failed to encode CanonicalVoteExtension: %w", err)
		}

		if !cmtPubKey.VerifySignature(extSignBytes, vote.ExtensionSignature) {
			return errors.New("received vote with invalid signature")
		}

		return nil
	}
}
```

Once at least 2/3 signatures, by voting power, are received and verified, the
validator can use the vote extensions to derive additional data or come to some
decision based on the vote extensions.

> NOTE: It is very important to state, that neither the vote propagation technique
> nor the vote extension verification mechanism described above is required for
> applications to implement. In other words, a proposer is not required to verify
> and propagate vote extensions along with their signatures nor are proposers
> required to verify those signatures. An application can implement it's own
> PKI mechanism and use that to sign and verify vote extensions.

#### Vote Extension Persistence

In order to make any data derived from vote extensions persistent, we propose to
allow application developers to "merge" the `processProposalState` into the
`finalizeState`, such that when processing transactions during `FinalizeBlock`,
the base state includes any state changes from processing votes extensions
(see [`FinalizeBlock`](#finalizeblock-1) below).

However, we DO NOT want to pollute `finalizeState` with any state changes from
verifying transactions during `ProcessProposal`, so we will introduce a new API,
`ResetProcessProposalState`, that will allow applications to essentially reset
the `ProcessProposal` state, effectively allowing any state transitions after to
be committed.

What this means is that any explicit state changes in a `ProcessProposal` handler
using the injected `processProposalState.Context` will be written to state! An
application must be careful to execute `ResetProcessProposalState` where and when
appropriate.

A `ProcessProposal` handler could look like the following:

```go
func (h MyHandler) ProcessProposalHandler() sdk.ProcessProposalHandler {
	return func(ctx sdk.Context, req abci.RequestProcessProposal) abci.ResponseProcessProposal {
		for _, txBytes := range req.Txs {
			_, err := h.app.ProcessProposalVerifyTx(txBytes)
			if err != nil {
				return abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}
			}
		}

		// Reset the ProposalProposal state, which, up to this point, has accumulated
		// possible state transitions from transaction verification and processing.
		// Any subsequent state transitions after this will be merged into the
		// FinalizeState before processing Begin/EndBlock and transactions.
		// 
		// NOTE: Applications must call this to ensure no state transactions up to now
		// are committed!
		h.app.ResetProcessProposalState()

		// Any state changes that occur on the provided ctx WILL be written to state!
		h.myKeeper.SetVoteExtResult(ctx, ...)
	
		return abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_ACCEPT}
	}
}
```

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

What will then exist is a single `FinalizeBlock` execution flow. Specifically, in
`FinalizeBlock` we will execute the application's `BeginBlock`, followed by
execution of all the transactions, finally followed by execution of the application's
`EndBlock`.

Note, we will still keep the existing transaction execution mechanics within
`BaseApp`, but all notions of `DeliverTx` will be removed, i.e. `deliverState`
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
	// merge any state changes from ProcessProposal into the FinalizeBlock state
	app.MergeProcessProposalState()

	beginBlockResp := app.beginBlock(ctx, req)
	appendBlockEventAttr(beginBlockResp.Events, "begin_block")

	txExecResults := make([]abci.ExecTxResult, 0, len(req.Txs))
	for _, tx := range req.Txs {
		result := app.runTx(runTxModeFinalize, tx)
		txExecResults = append(txExecResults, result)
	}

	endBlockResp := app.endBlock(ctx, req)
	appendBlockEventAttr(beginBlockResp.Events, "end_block")

	return abci.ResponseFinalizeBlock{
		TxResults:             txExecResults,
		Events:                joinEvents(beginBlockResp.Events, endBlockResp.Events),
		ValidatorUpdates:      endBlockResp.ValidatorUpdates,
		ConsensusParamUpdates: endBlockResp.ConsensusParamUpdates,
		AppHash:               nil,
	}
}
```

#### Events

Many tools, indexers and ecosystem libraries rely on the existence `BeginBlock`
and `EndBlock` events. Since CometBFT now only exposes `FinalizeBlockEvents`, we
find that it will still be useful for these clients and tools to still query for
and rely on existing events, especially since applications will still define
`BeginBlock` and `EndBlock` implementations.

In order to facilitate existing event functionality, we propose that all `BeginBlock`
and `EndBlock` events have a dedicated `EventAttribute` with `key=block` and
`value=begin_block|end_block`. The `EventAttribute` will be appended to each event
in both `BeginBlock` and `EndBlock` events`. 

## Consequences

### Backwards Compatibility

ABCI 2.0 is naturally not backwards compatible with prior versions of the Cosmos SDK
and CometBFT. For example, an application that requests `RequestFinalizeBlock`
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
* Vote extensions allow for an entirely new set of application primitives to be
  developed, such as in-process price oracles and encrypted mempools.

### Negative

* Some existing Cosmos SDK core APIs may need to be modified and thus broken.
* Signature verification in `ProcessProposal` of 100+ vote extension signatures
  will add significant performance overhead to `ProcessProposal`. Granted, the
	signature verification process can happen concurrently using an error group
	with `GOMAXPROCS` goroutines.

### Neutral

* Having to manually "inject" vote extensions into the block proposal during
  `PrepareProposal` is an awkward approach and takes up block space unnecessarily.
* The requirement of `ResetProcessProposalState` can create a footgun for
  application developers if they're not careful, but this is necessary in order
	for applications to be able to commit state from vote extension computation.

## Further Discussions

Future discussions include design and implementation of ABCI 3.0, which is a
continuation of ABCI++ and the general discussion of optimistic execution.

## References

* [ADR 060: ABCI 1.0 (Phase I)](adr-060-abci-1.0.md)
