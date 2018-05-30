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

If valid evidence is included in a block, the offending validator loses
a constant `SLASH_PROPORTION` of their current stake at the beginning of the block:

```
oldShares = validator.shares
validator.shares = oldShares * (1 - SLASH_PROPORTION)
```

This ensures that offending validators are punished the same amount whether they
act as a single validator with X stake or as N validators with collectively X
stake.


## Automatic Unbonding

Every block includes a set of precommits by the validators for the previous block, 
known as the LastCommit. A LastCommit is valid so long as it contains precommits from +2/3 of voting power.

Proposers are incentivized to include precommits from all
validators in the LastCommit by receiving additional fees
proportional to the difference between the voting power included in the
LastCommit and +2/3 (see [TODO](https://github.com/cosmos/cosmos-sdk/issues/967)).

Validators are penalized for failing to be included in the LastCommit for some
number of blocks by being automatically unbonded.

The following information is stored with each validator candidate, and is only non-zero if the candidate becomes an active validator:

```go
type ValidatorSigningInfo struct {
  StartHeight           int64
  IndexOffset           int64
  JailedUntil           int64
  SignedBlocksCounter   int64
  SignedBlocksBitArray  BitArray
}
```

Where:
* `StartHeight` is set to the height that the candidate became an active validator (with non-zero voting power).
* `IndexOffset` is incremented each time the candidate was a bonded validator in a block (and may have signed a precommit or not).
* `JailedUntil` is set whenever the candidate is revoked due to downtime
* `SignedBlocksCounter` is a counter kept to avoid unnecessary array reads. `SignedBlocksBitArray.Sum() == SignedBlocksCounter` always.
* `SignedBlocksBitArray` is a bit-array of size `SIGNED_BLOCKS_WINDOW` that records, for each of the last `SIGNED_BLOCKS_WINDOW` blocks,
whether or not this validator was included in the LastCommit. It uses a `1` if the validator was included, and a `0` if it was not. Note it is initialized with all 0s.

At the beginning of each block, we update the signing info for each validator and check if they should be automatically unbonded:

```
height := block.Height

for val in block.Validators:
  signInfo = val.SignInfo
  index := signInfo.IndexOffset % SIGNED_BLOCKS_WINDOW
  signInfo.IndexOffset++
  previous = signInfo.SignedBlocksBitArray.Get(index)

  // update counter if array has changed
  if previous and val in block.AbsentValidators:
    signInfo.SignedBlocksBitArray.Set(index, 0)
    signInfo.SignedBlocksCounter--
  else if !previous and val not in block.AbsentValidators:
    signInfo.SignedBlocksBitArray.Set(index, 1)
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
```
