# ADR 009: Evidence Module

## Changelog

- 31-07-2019: Initial draft

## Status

Proposed

## Context

In order to support building highly secure, robust and interoperable blockchain
applications, it is vital for the Cosmos SDK to expose a mechanism in which arbitrary
evidence can be submitted, evaluated and verified resulting in some agreed upon
penalty for any misbehaviour committed by a validator, such as equivocation (double-voting),
signing when unbonded, signing an incorrect state transition (in the future), etc.
Furthermore, such a mechanism is paramount for any
[IBC](https://github.com/cosmos/ics/blob/master/ibc/1_IBC_ARCHITECTURE.md) protocol
implementation in order to support the ability for any misbehaviour to be relayed
back from a collateralized chain to a primary chain so that the equivocating
validator(s) can be slashed.

## Decision

We will implement an evidence module in the Cosmos SDK supporting the following
functionality:

- Provide developers with the abstractions and interfaces necessary to define
custom evidence messages and types along with their slashing penalties
- Support the ability to route evidence messages to handlers in any module to
 determine the validity of submitted misbehaviour
- Support the ability through governance to modify slashing penalties of any
evidence type
- Querier implementation to support querying params, evidence types, params, and
all submitted valid misbehaviour

### Types

First, we define the `Evidence` interface type. The `x/evidence` module may implement
its own types that can be used by many chains (e.g. `CounterFactualEvidence`).
In addition, other modules may implement their own `Evidence` types in a similar
manner in which governance is extensible. It is important to note any concrete 
type implementing the `Evidence` interface may include arbitrary fields such as
an infraction time. We want the `Evidence` type to remain as flexible
as possible.

However, when submitting evidence to the `x/evidence` module, it must be submitted
as an `Infraction` which includes mandatory fields outlined below.

```go
type Evidence interface {
  Jailable() bool
  Route() string
  Type() string
  ValidateBasic() sdk.Error
  String() string
}

type Infraction struct {
  Evidence

  ConsensusAddress    sdk.ConsAddress
  InfractionHeight    int64
  Power               int64
}
```

### Routing & Handling

Each `Evidence` type must map to a specific unique route and be registered with
the `x/evidence` module. It accomplishes this through the `Router` implementation. 

```go
type Router interface {
  AddRoute(r string, h Handler) (rtr Router)
  HasRoute(r string) bool
  GetRoute(path string) (h Handler)
  Seal()
}
```

Upon successful routing through the `x/evidence` module, the `Evidence` type
is passed through a `Handler`. This `Handler` is responsible for executing all
corresponding business logic necessary for verifying the evidence. If no error
is returned, the `Evidence` is considered valid.

```go
type Handler func(ctx sdk.Context, evidence Evidence) sdk.Error
```

### Submission

Assuming the `Evidence` is valid, the corresponding slashing penalty is invoked
for the `Evidence`'s `Type`. Keep in mind the slashing penalty for any `Type` can
be configured through governance.

```go
type MsgSubmitInfraction struct {
  Infraction
}

func handleMsgSubmitInfraction(ctx sdk.Context, keeper Keeper, msg MsgSubmitEvidence) sdk.Result {
  if err := keeper.SubmitInfraction(ctx, msg.Infraction); err != nil {
    return err.Result()
  }

  // emit events...

  return sdk.Result{
    // ...
  }
}
```

The `x/evidence` module's keeper is responsible for matching the `Evidence` against
the module's router. Upon success the validator is slashed and the infraction is
persisted. In addition, the validator is jailed is the `Evidence` type is configured
to do so.

```go
func (k Keeper) SubmitInfraction(ctx sdk.Context, infraction Infraction) sdk.Error {
  handler := keeper.router.GetRoute(infraction.Evidence.Route())
  if err := handler(cacheCtx, infraction.Evidence); err != nil {
    return ErrInvalidEvidence(keeper.codespace, err.Result().Log)
  }

  keeper.stakingKeeper.Slash(
    ctx,
    infraction.ConsensusAddress,
    infraction.InfractionHeight,
    infraction.Power,
    keeper.GetSlashingPenalty(ctx, infraction.Evidence.Type()),
  )

  if infraction.Evidence.Jailable() {
    keeper.stakingKeeper.Jail(ctx, infraction.ConsensusAddress)
  }

  keeper.setInfraction(ctx, infraction)
}
```

### Genesis

We require the the `x/evidence` module's keeper to keep an internal persistent
mapping between `Evidence` types and slashing penalties represented as `InfractionPenalty`.

```go
var slashingPenaltyPrefixKey = []byte{0x01}

type InfractionPenalty struct {
  EvidenceType    string
  Penalty         sdk.Dec
}

func GetSlashingPenaltyKey(evidenceType string) []byte {
  return append(slashingPenaltyPrefixKey, []byte(evidenceType)...)
}

func (k Keeper) GetSlashingPenalty(ctx sdk.Context, evidenceType string) sdk.Dec {
  store := ctx.KVStore(k.storeKey)

  bz := store.Get(GetSlashingPenaltyKey(evidenceType))
  if len(bz) == 0 {
    return sdk.ZeroDec()
  }

  var ip InfractionPenalty
  k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &ip)

  return ip.Penalty
}
```

Finally, we need to represent the genesis state of the `x/evidence` module. The
module only needs a list of all submitted valid infractions and the infraction
penalties.

```go
type GenesisState struct {
  Infractions         []Infraction
  InfractionPenalties []InfractionPenalty
}
```

## Consequences

### Positive

- Allows the state machine to process equivocations submitted on-chain and penalize
validators based on agreed upon slashing parameters
- Does not solely rely on Tendermint to submit evidence

### Negative

- No easy way to introduce new evidence types through governance on a live chain
due to the inability to introduce the new evidence type's corresponding handler

### Neutral

- Should we persist infractions indefinitely? Or should we rather rely on events?

## References

- [ICS](https://github.com/cosmos/ics)
- [IBC Architecture](https://github.com/cosmos/ics/blob/master/ibc/1_IBC_ARCHITECTURE.md)
