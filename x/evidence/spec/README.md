# `x/evidence`

## Table of Contents

<!-- TOC -->

- **[01. Abstract](#01-abstract)**
- **[02. Concepts](#02-concepts)**
- **[03. State](#03-state)**
- **[04. Messages](#04-messages)**
- **[05. Events](#05-events)**
- **[06. Parameters](#06-parameters)**
- **[07. BeginBlock](#07-beginblock)**

## 01. Abstract

`x/evidence` is an implementation of a Cosmos SDK module, per [ADR 009](./../../../docs/architecture/adr-009-evidence-module.md),
that allows for the submission and handling of arbitrary evidence of misbehavior such
as equivocation and counterfactual signing.

The evidence module differs from standard evidence handling which typically expects the
underlying consensus engine, e.g. Tendermint, to automatically submit evidence when
it is discovered by allowing clients and foreign chains to submit more complex evidence
directly.

All concrete evidence types must implement the `Evidence` interface contract. Submitted
`Evidence` is first routed through the evidence module's `Router` in which it attempts
to find a corresponding registered `Handler` for that specific `Evidence` type.
Each `Evidence` type must have a `Handler` registered with the evidence module's
keeper in order for it to be successfully routed and executed.

Each corresponding handler must also fulfill the `Handler` interface contract. The
`Handler` for a given `Evidence` type can perform any arbitrary state transitions
such as slashing, jailing, and tombstoning.

## 02. Concepts

### Evidence

Any concrete type of evidence submitted to the `x/evidence` module must fulfill the
`Evidence` contract outlined below. Not all concrete types of evidence will fulfill
this contract in the same way and some data may be entirely irrelevant to certain
types of evidence. An additional `ValidatorEvidence`, which extends `Evidence`, has also
been created to define a contract for evidence against malicious validators.

```go
// Evidence defines the contract which concrete evidence types of misbehavior
// must implement.
type Evidence interface {
	Route() string
	Type() string
	String() string
	Hash() tmbytes.HexBytes
	ValidateBasic() error

	// Height at which the infraction occurred
	GetHeight() int64
}

// ValidatorEvidence extends Evidence interface to define contract
// for evidence against malicious validators
type ValidatorEvidence interface {
	Evidence

	// The consensus address of the malicious validator at time of infraction
	GetConsensusAddress() sdk.ConsAddress

	// The total power of the malicious validator at time of infraction
	GetValidatorPower() int64

	// The total validator set power at time of infraction
	GetTotalPower() int64
}
```

### Registration & Handling

The `x/evidence` module must first know about all types of evidence it is expected
to handle. This is accomplished by registering the `Route` method in the `Evidence`
contract with what is known as a `Router` (defined below). The `Router` accepts
`Evidence` and attempts to find the corresponding `Handler` for the `Evidence`
via the `Route` method.

```go
// Router defines a contract for which any Evidence handling module must
// implement in order to route Evidence to registered Handlers.
type Router interface {
  AddRoute(r string, h Handler) Router
  HasRoute(r string) bool
  GetRoute(path string) Handler
  Seal()
  Sealed() bool
}
```

The `Handler` (defined below) is responsible for executing the entirety of the
business logic for handling `Evidence`. This typically includes validating the
evidence, both stateless checks via `ValidateBasic` and stateful checks via any
keepers provided to the `Handler`. In addition, the `Handler` may also perform
capabilities such as slashing and jailing a validator.

```go
// Handler defines an agnostic Evidence handler. The handler is responsible
// for executing all corresponding business logic necessary for verifying the
// evidence as valid. In addition, the Handler may execute any necessary
// slashing and potential jailing.
type Handler func(Context, Evidence) error
```

## 03. State

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

## 04. Messages

### MsgSubmitEvidence

Evidence is submitted through a `MsgSubmitEvidence` message:

```protobuf
// MsgSubmitEvidence represents a message that supports submitting arbitrary
// Evidence of misbehavior such as equivocation or counterfactual signing.
message MsgSubmitEvidence {
  string              submitter = 1;
  google.protobuf.Any evidence  = 2;
}
```

Note, the `Evidence` of a `MsgSubmitEvidence` message must have a corresponding
`Handler` registered with the `x/evidence` module's `Router` in order to be processed
and routed correctly.

Given the `Evidence` is registered with a corresponding `Handler`, it is processed
as follows:

```go
func SubmitEvidence(ctx Context, evidence Evidence) error {
  if _, ok := GetEvidence(ctx, evidence.Hash()); ok {
    return ErrEvidenceExists(codespace, evidence.Hash().String())
  }
  if !router.HasRoute(evidence.Route()) {
    return ErrNoEvidenceHandlerExists(codespace, evidence.Route())
  }

  handler := router.GetRoute(evidence.Route())
  if err := handler(ctx, evidence); err != nil {
    return ErrInvalidEvidence(codespace, err.Error())
  }

  SetEvidence(ctx, evidence)
  return nil
}
```

First, there must not already exist valid submitted `Evidence` of the exact same
type. Secondly, the `Evidence` is routed to the `Handler` and executed. Finally,
if there is no error in handling the `Evidence`, it is persisted to state.

## 05. Events

The `x/evidence` module emits the following events:

### Handlers

#### MsgSubmitEvidence

| Type            | Attribute Key | Attribute Value |
| --------------- | ------------- | --------------- |
| submit_evidence | evidence_hash | {evidenceHash}  |
| message         | module        | evidence        |
| message         | sender        | {senderAddress} |
| message         | action        | submit_evidence |

## 06. Parameters

The evidence module does not have any parameters.

## 07. BeginBlock

### Evidence Handling

Tendermint blocks can include
[Evidence](https://github.com/tendermint/tendermint/blob/master/docs/spec/blockchain/blockchain.md#evidence),
which indicates that a validator committed malicious behavior. The relevant information is
forwarded to the application as ABCI Evidence in `abci.RequestBeginBlock` so that
the validator an be accordingly punished.

#### Equivocation

Currently, the evidence module only handles evidence of type `Equivocation` which is derived from
Tendermint's `ABCIEvidenceTypeDuplicateVote` during `BeginBlock`.

For some `Equivocation` submitted in `block` to be valid, it must satisfy:

`Evidence.Timestamp >= block.Timestamp - MaxEvidenceAge`

Where `Evidence.Timestamp` is the timestamp in the block at height `Evidence.Height` and
`block.Timestamp` is the current block timestamp.

If valid `Equivocation` evidence is included in a block, the validator's stake is
reduced (slashed) by `SlashFractionDoubleSign`, which is defined by the `x/slashing` module,
of what their stake was when the infraction occurred (rather than when the evidence was discovered).
We want to "follow the stake", i.e. the stake which contributed to the infraction
should be slashed, even if it has since been redelegated or started unbonding.

In addition, the validator is permanently jailed and tombstoned making it impossible for that
validator to ever re-enter the validator set.

The `Equivocation` evidence is handled as follows:

```go
func (k Keeper) HandleDoubleSign(ctx Context, evidence Equivocation) {
  consAddr := evidence.GetConsensusAddress()
  infractionHeight := evidence.GetHeight()

  // calculate the age of the evidence
  blockTime := ctx.BlockHeader().Time
  age := blockTime.Sub(evidence.GetTime())

  // reject evidence we cannot handle
  if _, err := k.slashingKeeper.GetPubkey(ctx, consAddr.Bytes()); err != nil {
    return
  }

  // reject evidence if it is too old
  if age > k.MaxEvidenceAge(ctx) {
    return
  }

  // reject evidence if the validator is already unbonded
  validator := k.stakingKeeper.ValidatorByConsAddr(ctx, consAddr)
  if validator == nil || validator.IsUnbonded() {
    return
  }

  // verify the validator has signing info in order to be slashed and tombstoned
  if ok := k.slashingKeeper.HasValidatorSigningInfo(ctx, consAddr); !ok {
    panic(...)
  }

  // reject evidence if the validator is already tombstoned
  if k.slashingKeeper.IsTombstoned(ctx, consAddr) {
    return
  }

  // We need to retrieve the stake distribution which signed the block, so we
  // subtract ValidatorUpdateDelay from the evidence height.
  // Note, that this *can* result in a negative "distributionHeight", up to
  // -ValidatorUpdateDelay, i.e. at the end of the
  // pre-genesis block (none) = at the beginning of the genesis block.
  // That's fine since this is just used to filter unbonding delegations & redelegations.
  distributionHeight := infractionHeight - sdk.ValidatorUpdateDelay

  // Slash validator. The `power` is the int64 power of the validator as provided
  // to/by Tendermint. This value is validator.Tokens as sent to Tendermint via
  // ABCI, and now received as evidence. The fraction is passed in to separately
  // to slash unbonding and rebonding delegations.
  k.slashingKeeper.Slash(ctx, consAddr, evidence.GetValidatorPower(), distributionHeight)

  // Jail the validator if not already jailed. This will begin unbonding the
  // validator if not already unbonding (tombstoned).
  if !validator.IsJailed() {
    k.slashingKeeper.Jail(ctx, consAddr)
  }

  k.slashingKeeper.JailUntil(ctx, consAddr, types.DoubleSignJailEndTime)
  k.slashingKeeper.Tombstone(ctx, consAddr)
}
```

Note, the slashing, jailing, and tombstoning calls are delegated through the `x/slashing` module
which emit informative events and finally delegate calls to the `x/staking` module. Documentation
on slashing and jailing can be found in the [x/staking spec](/.././cosmos-sdk/x/staking/spec/02_state_transitions.md)
