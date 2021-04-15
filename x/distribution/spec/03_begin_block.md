<!--
order: 3
-->

# Begin Block

At each `BeginBlock`, the fees received in the previous block are transferred to the distribution `ModuleAccount`, as it's the account the one who keeps track of the flow of coins in (as in this case) and out the module. The fees are also allocated to the proposer, community fund and global pool. When the validator is the proposer of the round, that validator (and their delegators) receives between 1% and 5% of fee rewards, the reserve community tax is then charged, then the remainder is distributed proportionally by voting power to all bonded validators independent of whether they voted (social distribution). Note the social distribution is applied to proposer validator in addition to the proposer reward.

The amount of proposer reward is calculated from pre-commits Tendermint messages in order to incentivize validators to wait and include additional pre-commits in the block. All provision rewards are added to a provision reward pool which validator holds individually (`ValidatorDistribution.ProvisionsRewardPool`).

## The Distribution Scheme

See [params](07_params.md) for description of parameters.

Let `fees` be the total fees collected in the previous block. All fees are
collected in a specific module account during the block. During `BeginBlock`,
they are sent to the `"distribution"` `ModuleAccount`. No other sending of
tokens occur. Instead, the rewards each account is entitled to are stored, and
withdrawals can be triggered through the messages `FundCommunityPool`,
`WithdrawValidatorCommission` and `WithdrawDelegatorReward`.

### Reward to the Community Pool

The community pool gets `communitytax * fees`, plus any remaining dust after
validators get their rewards, which will always be rounded down to nearest
integer value.

### Reward To the Validators

The proposer will receive a base reward of `fees * baseproposerreward`. In
addition, they will receive a bonus of `fees * bonusproposerreward * P`, where
`P = (precommits included / total bonded validator power)`. The more precommits
the proposer includes, the larger `P` is. `P` can never be larger than `1.00`
, and will always be larger than `2/3`.

Any remaining fees are distributed among all the bonded validators, including
the proposer, in proportion to their consensus power.

```
powFrac = validator power / total bonded validator power
proposerMul = baseproposerreward + bonusproposerreward * P
voteMul = 1 - communitytax - proposerMul
```

In total, the proposer will receive `fees * voteMul * powFrac + fees * proposerMul`.
All other validators will receive `fees * voteMul`.

### Rewards to Delegators

Each validators rewards are distributed to its delegators. Note that the
validator also has a self-delegation.

The validator sets a commission rate. The commission rate is flexible, but each
validator sets a max rate which may never be exceeded, and a max daily increase
which may never be exceeded. This protects delegators against a validator
suddenly increasing their commission rate, taking all the rewards.

The outstanding rewards that the operator is entitled to are stored
in `ValidatorCurrentRewards`, while the rewards the delegators are entitled to
are stored in `ValidatorCurrentRewards`.
The [F1 fee distribution scheme](01_concepts.md) is used to calculate the
rewards per delegator as they withdraw or update their delegation, and is thus
not handled in `BeginBlock`.

### Example Distribution

As an example, assume the underlying consensus engine selects block proposers in
proportion to their power relative to the entire bonded power. Further, assume
all validators are equally good at including pre-commits in their proposed
blocks. Then we hold `(precommits included) / (total bonded validator power)`
constant, and then the amortized block reward for the validator is
simply `( validator power / total bonded power) * (1 - community tax rate)` of
the total rewards. Consequently, the reward for a single delegator is:

```
(delegator proportion of the validator power / validator power) * (validator power / total bonded power)
  * (1 - community tax rate) * (1 - validator commision rate)
= (delegator proportion of the validator power / total bonded power) * (1 -
community tax rate) * (1 - validator commision rate)
```
