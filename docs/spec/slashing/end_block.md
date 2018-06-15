# End-Block 

## Slashing

Messges which may compromise the safety of the underlying consensus protocol ("equivocations")
result in some amount of the offending validator's shares being removed ("slashed").

Currently, such messages include only the following:

- prevotes by the same validator for more than one BlockID at the same
  Height and Round 
- precommits by the same validator for more than one BlockID at the same
  Height and Round 

We call any such pair of conflicting votes `Evidence`. Full nodes in the network prioritize the 
detection and gossipping of `Evidence` so that it may be rapidly included in blocks and the offending
validators punished.

For some `evidence` to be valid, it must satisfy: 

`evidence.Timestamp >= block.Timestamp - MAX_EVIDENCE_AGE`

where `evidence.Timestamp` is the timestamp in the block at height
`evidence.Height` and `block.Timestamp` is the current block timestamp.

If valid evidence is included in a block, the validator's stake is reduced by `SLASH_PROPORTION` of 
what there stake was at the eqiuvocation occurred (rather than when it was found):

```
curVal := validator
oldVal := loadValidator(evidence.Height, evidence.Address)

slashAmount := SLASH_PROPORTION * oldVal.Shares

curVal.Shares -= slashAmount
```

This ensures that offending validators are punished the same amount whether they
act as a single validator with X stake or as N validators with collectively X
stake.

We also need to loop through the unbondings and redelegations to slash them as
well:

```
unbondings := getUnbondings(validator.Address)
for unbond in unbondings {
    if was not bonded before evidence.Height {
        continue
    }
    unbond.InitialTokens
    burn := unbond.InitialTokens * SLASH_PROPORTION
    unbond.Tokens -= burn
}

// only care if source gets slashed because we're already bonded to destination
// so if destination validator gets slashed our delegation just has same shares
// of smaller pool.
redels := getRedelegationsBySource(validator.Address)
for redel in redels {

    if was not bonded before evidence.Height {
        continue
    }

    burn := redel.InitialTokens * SLASH_PROPORTION

    amount := unbondFromValidator(redel.Destination, burn)
    destroy(amount)
}
```

## Automatic Unbonding

At the beginning of each block, we update the signing info for each validator and check if they should be automatically unbonded:

```
height := block.Height

for val in block.Validators:
  signInfo = SigningInfo.Get(val.Address)
  if signInfo == nil{
        signInfo.StartHeight = height
  }

  index := signInfo.IndexOffset % SIGNED_BLOCKS_WINDOW
  signInfo.IndexOffset++
  previous = SigningBitArray.Get(val.Address, index)

  // update counter if array has changed
  if previous and val in block.AbsentValidators:
    SigningBitArray.Set(val.Address, index, false)
    signInfo.SignedBlocksCounter--
  else if !previous and val not in block.AbsentValidators:
    SigningBitArray.Set(val.Address, index, true)
    signInfo.SignedBlocksCounter++
  // else previous == val not in block.AbsentValidators, no change

  // validator must be active for at least SIGNED_BLOCKS_WINDOW
  // before they can be automatically unbonded for failing to be 
  // included in 50% of the recent LastCommits
  minHeight = signInfo.StartHeight + SIGNED_BLOCKS_WINDOW
  minSigned = SIGNED_BLOCKS_WINDOW / 2
  if height > minHeight AND signInfo.SignedBlocksCounter < minSigned:
    signInfo.JailedUntil = block.Time + DOWNTIME_UNBOND_DURATION

    slash & unbond the validator

  SigningInfo.Set(val.Address, signInfo)
```
