# Validator Set Changes

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

Every block includes a set of precommits by the validators for the previous block, 
known as the LastCommit. A LastCommit is valid so long as it contains precommits from +2/3 of voting power.

Proposers are incentivized to include precommits from all
validators in the LastCommit by receiving additional fees
proportional to the difference between the voting power included in the
LastCommit and +2/3 (see [TODO](https://github.com/cosmos/cosmos-sdk/issues/967)).

Validators are penalized for failing to be included in the LastCommit for some
number of blocks by being automatically unbonded.

Maps:

- map1: < prefix-info | tm val addr > -> <validator signing info>
- map2: < prefix-bit-array | tm val addr | LE uint64 index in sign bit array > -> < signed bool >

The following information is stored with each validator candidate, and is only non-zero if the candidate becomes an active validator:

```go
type ValidatorSigningInfo struct {
  StartHeight           int64
  IndexOffset           int64
  JailedUntil           int64
  SignedBlocksCounter   int64
}

```

Where:
* `StartHeight` is set to the height that the candidate became an active validator (with non-zero voting power).
* `IndexOffset` is incremented each time the candidate was a bonded validator in a block (and may have signed a precommit or not).
* `JailedUntil` is set whenever the candidate is revoked due to downtime
* `SignedBlocksCounter` is a counter kept to avoid unnecessary array reads. `SignedBlocksBitArray.Sum() == SignedBlocksCounter` always.


Map2 simulates a bit array - better to do the lookup rather than read/write the
bitarray every time. Size of bit-array is `SIGNED_BLOCKS_WINDOW`. It records, for each of the last `SIGNED_BLOCKS_WINDOW` blocks,
whether or not this validator was included in the LastCommit. 
It sets the value to true if the validator was included and false if not.
Note it is not explicilty initialized (the keys wont exist).

At the beginning of each block, we update the signing info for each validator and check if they should be automatically unbonded:

```
height := block.Height

for val in block.Validators:
  signInfo = getSignInfo(val.Address)
  if signInfo == nil{
        signInfo.StartHeight = height
  }

  index := signInfo.IndexOffset % SIGNED_BLOCKS_WINDOW
  signInfo.IndexOffset++
  previous = getDidSign(val.Address, index)

  // update counter if array has changed
  if previous and val in block.AbsentValidators:
    setDidSign(val.Address, index, false)
    signInfo.SignedBlocksCounter--
  else if !previous and val not in block.AbsentValidators:
    setDidSign(val.Address, index, true)
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

  setSignInfo(val.Address, signInfo)
```
