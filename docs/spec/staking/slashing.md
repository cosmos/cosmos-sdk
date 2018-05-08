
# Slashing

A validator bond is an economic commitment made by a validator signing key to both the safety and liveness of
the consensus. Validator keys must not sign invalid messages which could
violate consensus safety, and their signed precommit messages must be regularly included in
block commits. 

The incentivization of these two goals are treated separately.

## Safety

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

`evidence.Height >= block.Height - MAX_EVIDENCE_AGE`

If valid evidence is included in a block, the offending validator loses
a constant `SLASH_PROPORTION` of their current stake:

```
oldShares = validator.shares
validator.shares = oldShares * (1 - SLASH_PROPORTION)
```

This ensures that offending validators are punished the same amount whether they
act as a single validator with X stake or as N validators with collectively X
stake.



## Liveness

Every block includes a set of precommits by the validators for the previous block, 
known as the LastCommit. A LastCommit is valid so long as it contains precommits from +2/3 of voting power.

Proposers are incentivized to include precommits from all
validators in the LastCommit by receiving additional fees
proportional to the difference between the voting power included in the
LastCommit and +2/3 (see [TODO](https://github.com/cosmos/cosmos-sdk/issues/967)).

Validators are penalized for failing to be included in the LastCommit for some
number of blocks by being automatically unbonded.


TODO: do we do this by trying to track absence directly in the state, using
something like the below, or do we let users notify the app when a validator has
been absent using the
[TxLivenessCheck](https://github.com/cosmos/cosmos-sdk/blob/develop/docs/spec/staking/spec-technical.md#txlivelinesscheck).


A list, `ValidatorAbsenceInfos`, is stored in the state and used to track how often
validators were included in a LastCommit.

```go
// Ordered by ValidatorAddress.
// One entry for each validator.
type ValidatorAbsenceInfos []ValidatorAbsenceInfo

type ValidatorAbsenceInfo struct {
    ValidatorAddress    []byte  // address of the validator 
    FirstHeight         int64   // first height the validator was absent
    Count               int64   // number of heights validator was absent since (and including) first
}
```

