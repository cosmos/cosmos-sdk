<!--
order: 6
-->

# BeginBlock

## Evidence Handling

Tendermint blocks can include
[Evidence](https://github.com/tendermint/tendermint/blob/master/docs/spec/blockchain/blockchain.md#evidence),
which indicates that a validator committed malicious behavior. The relevant information is
forwarded to the application as ABCI Evidence in `abci.RequestBeginBlock` so that
the validator an be accordingly punished.

### Equivocation

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
