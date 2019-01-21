# Begin-Block

## Evidence handling

Tendermint blocks can include
[Evidence](https://github.com/tendermint/tendermint/blob/develop/docs/spec/blockchain/blockchain.md#evidence), which indicates that a validator
committed malicious behavior. The relevant information is forwarded to the
application as [ABCI
Evidence](https://github.com/tendermint/tendermint/blob/develop/abci/types/types.proto#L259) in `abci.RequestBeginBlock`
so that the validator an be accordingly punished.

For some `evidence` to be valid, it must satisfy:

`evidence.Timestamp >= block.Timestamp - MAX_EVIDENCE_AGE`

where `evidence.Timestamp` is the timestamp in the block at height
`evidence.Height` and `block.Timestamp` is the current block timestamp.

If valid evidence is included in a block, the validator's stake is reduced by `SLASH_PROPORTION` of 
what their stake was when the infraction occurred (rather than when the evidence was discovered).
We want to "follow the stake": the stake which contributed to the infraction should be
slashed, even if it has since been redelegated or started unbonding. 

We first need to loop through the unbondings and redelegations from the slashed validator
and track how much stake has since moved:

```
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
stake.  The amount slashed for all double signature infractions committed within a
single slashing period is capped as described in [overview.md](overview.md) under Tombstone Caps.

## Uptime tracking

At the beginning of each block, we update the signing info for each validator and check if they've dipped below the liveness threshhold over the tracked window.  If so, they will be slashed by `LivenessSlashAmount` and will be Jailed for `LivenessJailPeriod`.  Liveness slashes do NOT lead to a tombstombing.

```
height := block.Height

for val in block.Validators:
  signInfo = SigningInfo.Get(val.Address)
  if signInfo == nil{
        signInfo.StartHeight = height
  }

  index := signInfo.IndexOffset % SIGNED_BLOCKS_WINDOW
  signInfo.IndexOffset++
  previous = MissedBlockBitArray.Get(val.Address, index)

  // update counter if array has changed
  if !previous and val in block.AbsentValidators:
    MissedBlockBitArray.Set(val.Address, index, true)
    signInfo.MissedBlocksCounter++
  else if previous and val not in block.AbsentValidators:
    MissedBlockBitArray.Set(val.Address, index, false)
    signInfo.MissedBlocksCounter--
  // else previous == val not in block.AbsentValidators, no change

  // validator must be active for at least SIGNED_BLOCKS_WINDOW
  // before they can be automatically unbonded for failing to be
  // included in 50% of the recent LastCommits
  minHeight = signInfo.StartHeight + SIGNED_BLOCKS_WINDOW
  maxMissed = SIGNED_BLOCKS_WINDOW / 2
  if height > minHeight AND signInfo.MissedBlocksCounter > maxMissed:
    signInfo.JailedUntil = block.Time + DOWNTIME_UNBOND_DURATION
    signInfo.IndexOffset = 0
    signInfo.MissedBlocksCounter = 0
    clearMissedBlockBitArray()
    slash & jail the validator

  SigningInfo.Set(val.Address, signInfo)
```
