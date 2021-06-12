<!--
order: 3
-->

# Begin Block

At each `BeginBlock`, all fees received in the previous block are transferred to
the distribution `ModuleAccount` account. When a delegator or validator
withdraws their rewards, they are taken out of the `ModuleAccount`. During begin
block, the different claims on the fees collected are updated as follows:

- The block proposer of the previous height and its delegators receive between 1% and 5% of fee rewards.
- The reserve community tax is charged.
- The remainder is distributed proportionally by voting power to all bonded validators

To incentivize validators to wait and include additional pre-commits in the block, the block proposer reward is calculated from Tendermint pre-commit messages.

## The Distribution Scheme

See [params](07_params.md) for description of parameters.

Let `fees` be the total fees collected in the previous block, including
inflationary rewards to the stake. All fees are collected in a specific module
account during the block. During `BeginBlock`, they are sent to the
`"distribution"` `ModuleAccount`. No other sending of tokens occurs. Instead, the
rewards each account is entitled to are stored, and withdrawals can be triggered
through the messages `FundCommunityPool`, `WithdrawValidatorCommission` and
`WithdrawDelegatorReward`.

### Reward to the Community Pool

The community pool gets `community_tax * fees`, plus any remaining dust after
validators get their rewards that are always rounded down to the nearest
integer value.

### Reward To the Validators

The proposer receives a base reward of `fees * baseproposerreward` and a bonus
of `fees * bonusproposerreward * P`, where `P = (total power of validators with
included precommits / total bonded validator power)`. The more precommits the
proposer includes, the larger `P` is. `P` can never be larger than `1.00` (since
only bonded validators can supply valid precommits) and is always larger than
`2/3`.

Any remaining fees are distributed among all the bonded validators, including
the proposer, in proportion to their consensus power.

```
powFrac = validator power / total bonded validator power
proposerMul = baseproposerreward + bonusproposerreward * P
voteMul = 1 - communitytax - proposerMul
```

In total, the proposer receives `fees  * (voteMul * powFrac + proposerMul)`.
All other validators receive `fees * voteMul * powFrac`.

### Rewards to Delegators

Each validator's rewards are distributed to its delegators. The validator also
has a self-delegation that is treated like a regular delegation in
distribution calculations.

The validator sets a commission rate. The commission rate is flexible, but each
validator sets a maximum rate and a maximum daily increase. These maximums cannot be exceeded and protect delegators from sudden increases of validator commission rates to prevent validators from taking all of the rewards.

The outstanding rewards that the operator is entitled to are stored in
`ValidatorAccumulatedCommission`, while the rewards the delegators are entitled
to are stored in `ValidatorCurrentRewards`. The [F1 fee distribution
scheme](01_concepts.md) is used to calculate the rewards per delegator as they
withdraw or update their delegation, and is thus not handled in `BeginBlock`.

### Example Distribution

For this example distribution, the underlying consensus engine selects block proposers in
proportion to their power relative to the entire bonded power.

All validators are equally performant at including pre-commits in their proposed
blocks. Then hold `(precommits included) / (total bonded validator power)`
constant so that the amortized block reward for the validator is `( validator power / total bonded power) * (1 - community tax rate)` of
the total rewards. Consequently, the reward for a single delegator is:

```
(delegator proportion of the validator power / validator power) * (validator power / total bonded power)
  * (1 - community tax rate) * (1 - validator commision rate)
= (delegator proportion of the validator power / total bonded power) * (1 -
community tax rate) * (1 - validator commision rate)
```
