# BeginBlock

## Evidence Handling

Tendermint blocks can include
[Evidence](https://github.com/tendermint/tendermint/blob/master/docs/spec/blockchain/blockchain.md#evidence), which indicates that a validator committed malicious
behavior. The relevant information is forwarded to the application as ABCI Evidence
in `abci.RequestBeginBlock` so that the validator an be accordingly punished.

For some `Evidence` submitted in `block` to be valid, it must satisfy:

`Evidence.Timestamp >= block.Timestamp - MaxEvidenceAge`

Where `Evidence.Timestamp` is the timestamp in the block at height
`Evidence.Height` and `block.Timestamp` is the current block timestamp.

If valid evidence is included in a block, the validator's stake is reduced by
some penalty (`SlashFractionDoubleSign` for equivocation) of what their stake was
when the infraction occurred (rather than when the evidence was discovered). We
want to "follow the stake", i.e. the stake which contributed to the infraction
should be slashed, even if it has since been redelegated or started unbonding.

We first need to loop through the unbondings and redelegations from the slashed
validator and track how much stake has since moved:

```go
slashAmountUnbondings := 0
slashAmountRedelegations := 0

unbondings := getUnbondings(validator.Address)
for unbond in unbondings {

    if was not bonded before evidence.Height or started unbonding before unbonding period ago {
        continue
    }

    burn := unbond.InitialTokens * SLASH_PROPORTION
    slashAmountUnbondings += burn

    unbond.Tokens = max(0, unbond.Tokens - burn)
}

// only care if source gets slashed because we're already bonded to destination
// so if destination validator gets slashed our delegation just has same shares
// of smaller pool.
redels := getRedelegationsBySource(validator.Address)
for redel in redels {

    if was not bonded before evidence.Height or started redelegating before unbonding period ago {
        continue
    }

    burn := redel.InitialTokens * SLASH_PROPORTION
    slashAmountRedelegations += burn

    amount := unbondFromValidator(redel.Destination, burn)
    destroy(amount)
}
```

We then slash the validator and tombstone them:

```
curVal := validator
oldVal := loadValidator(evidence.Height, evidence.Address)

slashAmount := SLASH_PROPORTION * oldVal.Shares
slashAmount -= slashAmountUnbondings
slashAmount -= slashAmountRedelegations

curVal.Shares = max(0, curVal.Shares - slashAmount)

signInfo = SigningInfo.Get(val.Address)
signInfo.JailedUntil = MAX_TIME
signInfo.Tombstoned = true
SigningInfo.Set(val.Address, signInfo)
```

This ensures that offending validators are punished the same amount whether they
act as a single validator with X stake or as N validators with collectively X
stake. The amount slashed for all double signature infractions committed within a
single slashing period is capped as described in [overview.md](overview.md) under Tombstone Caps.

## Liveness Tracking

At the beginning of each block, we update the `ValidatorSigningInfo` for each
validator and check if they've crossed below the liveness threshold over a
sliding window. This sliding window is defined by `SignedBlocksWindow` and the
index in this window is determined by `IndexOffset` found in the validator's
`ValidatorSigningInfo`. For each block processed, the `IndexOffset` is incrimented
regardless if the validator signed or not. Once the index is determined, the
`MissedBlocksBitArray` and `MissedBlocksCounter` are updated accordingly.

Finally, in order to determine if a validator crosses below the liveness threshold,
we fetch the maximum number of blocks missed, `maxMissed`, which is
`SignedBlocksWindow - (MinSignedPerWindow * SignedBlocksWindow)` and the minimum
height at which we can determine liveness, `minHeight`. If the current block is
greater than `minHeight` and the validator's `MissedBlocksCounter` is greater than
`maxMissed`, they will be slashed by `SlashFractionDowntime`, will be jailed
for `DowntimeJailDuration`, and have the following values reset:
`MissedBlocksBitArray`, `MissedBlocksCounter`, and `IndexOffset`.

__Note__: Liveness slashes do **NOT** lead to a tombstombing.

```go
height := block.Height

for vote in block.LastCommitInfo.Votes {
  signInfo := GetValidatorSigningInfo(vote.Validator.Address)

  // This is a relative index, so we counts blocks the validator SHOULD have
  // signed. We use the 0-value default signing info if not present, except for
  // start height.
  index := signInfo.IndexOffset % SignedBlocksWindow()
  signInfo.IndexOffset++

  // Update MissedBlocksBitArray and MissedBlocksCounter. The MissedBlocksCounter
  // just tracks the sum of MissedBlocksBitArray. That way we avoid needing to
  // read/write the whole array each time.
  missedPrevious := GetValidatorMissedBlockBitArray(vote.Validator.Address, index)
  missed := !signed

  switch {
  case !missedPrevious && missed:
    // array index has changed from not missed to missed, increment counter
    SetValidatorMissedBlockBitArray(vote.Validator.Address, index, true)
    signInfo.MissedBlocksCounter++

  case missedPrevious && !missed:
    // array index has changed from missed to not missed, decrement counter
    SetValidatorMissedBlockBitArray(vote.Validator.Address, index, false)
    signInfo.MissedBlocksCounter--

  default:
    // array index at this index has not changed; no need to update counter
  }

  if missed {
    // emit events...
  }

  minHeight := signInfo.StartHeight + SignedBlocksWindow()
  maxMissed := SignedBlocksWindow() - MinSignedPerWindow()

  // If we are past the minimum height and the validator has missed too many
  // jail and slash them.
  if height > minHeight && signInfo.MissedBlocksCounter > maxMissed {
    validator := ValidatorByConsAddr(vote.Validator.Address)

    // emit events...

    // We need to retrieve the stake distribution which signed the block, so we
    // subtract ValidatorUpdateDelay from the block height, and subtract an
    // additional 1 since this is the LastCommit.
    //
    // Note, that this CAN result in a negative "distributionHeight" up to
    // -ValidatorUpdateDelay-1, i.e. at the end of the pre-genesis block (none) = at the beginning of the genesis block.
    // That's fine since this is just used to filter unbonding delegations & redelegations.
    distributionHeight := height - sdk.ValidatorUpdateDelay - 1

    Slash(vote.Validator.Address, distributionHeight, vote.Validator.Power, SlashFractionDowntime())
    Jail(vote.Validator.Address)

    signInfo.JailedUntil = block.Time.Add(DowntimeJailDuration())

    // We need to reset the counter & array so that the validator won't be
    // immediately slashed for downtime upon rebonding.
    signInfo.MissedBlocksCounter = 0
    signInfo.IndexOffset = 0
    ClearValidatorMissedBlockBitArray(vote.Validator.Address)
  }

  SetValidatorSigningInfo(vote.Validator.Address, signInfo)
}
```
