<!--
order: 3
-->

# Begin Block

At each `BeginBlock`, the fees received are transferred to the distribution `ModuleAccount`, as it's the account the one who keeps track of the flow of coins in (as in this case) and out the module. The fees are also allocated to the proposer, community fund and global pool. When the validator is the proposer of the round, that validator (and their delegators) receives between 1% and 5% of fee rewards, the reserve community tax is then charged, then the remainder is distributed proportionally by voting power to all bonded validators independent of whether they voted (social distribution). Note the social distribution is applied to proposer validator in addition to the proposer reward.

The amount of proposer reward is calculated from pre-commits Tendermint messages in order to incentivize validators to wait and include additional pre-commits in the block. All provision rewards are added to a provision reward pool which validator holds individually (`ValidatorDistribution.ProvisionsRewardPool`).

```go
func AllocateTokens(feesCollected sdk.Coins, feePool FeePool, proposer ValidatorDistribution, 
              sumPowerPrecommitValidators, totalBondedTokens, communityTax, 
              proposerCommissionRate sdk.Dec)

     SendCoins(FeeCollectorAddr, DistributionModuleAccAddr, feesCollected)
     feesCollectedDec = MakeDecCoins(feesCollected)
     proposerReward = feesCollectedDec * (0.01 + 0.04 
                       * sumPowerPrecommitValidators / totalBondedTokens)

     commission = proposerReward * proposerCommissionRate
     proposer.PoolCommission += commission
     proposer.Pool += proposerReward - commission

     communityFunding = feesCollectedDec * communityTax
     feePool.CommunityFund += communityFunding

     poolReceived = feesCollectedDec - proposerReward - communityFunding
     feePool.Pool += poolReceived

     SetValidatorDistribution(proposer)
     SetFeePool(feePool)
```

## BeginBlock Algorithm summarized

All multiplications are truncated, meaning rounded down.
They are also checked for overflow.

The type "por" below means a real number between 0 and 1, inclusive, signifying a portion

### Protocol level parameters:
See [params](07_params.md).

### Protocol level variables:
Total bonded validator power : Tp :: nat
Community pool tokens        : Cp :: nat

# Parameters per validator:
Commission rate                     : cr :: por

# Variables per validator:
Current period                      : p  :: nat
Accumulated fees for current period : T  :: nat
Total stake                         : n  :: nat
Historical period data              : E  :: nat -> (nat, nat)
from period to T for period
and reference count
Accumulated commission              : C  :: nat
Power                               : P  :: nat

# Variables per delegation:
Receiving validator         : v :: validator (see above)
Delegated amount            : x :: nat
Last fully withdrawn period : k :: nat

# Invariants:
- forall v :: validator . v.E[i] = 0 if i >= v.p

# Create validator with initial delegation amount of x tokens:
v = new validator object in the store
v.p = 1
v.T = 0
v.n = x
v.E[0] = 0
d = new delegation object in the store
d.v = v
d.x = x
d.k = 0

# BeginBlock:
For each validator v in the active validator set:
r = reward to v for previous block     // see below
c = r * v.cr
v.C = v.C + c
v.T = v.T + r - c
r = remaining fees // due to communitytax and rounding-down of validator rewards
Cp = Cp + r


# Delegation of x tokens to validator v:
// In the staking module:
d = new delegation object in the store

if delegator has existing delegations to v:
Withdraw rewards from validator v
d.v = v
d.x = x
d.k = v.p

v.n = v.n + x

// In the distribution module
Increment period of v

# Withdraw rewards from validator v:
d = current delegation from delegator to v
r = d.x * (E[v.p] - E[d.k])
Increment period of v
d.k = v.p


# Helpers
## Increment period of v
// Conclude current period of validator v, and start a new one.
v.E[p] = v.T/v.n + v.E[p-1]
v.p = v.p + 1
v.T = 0

## Reward to validator v for previous block
p = baseproposerreward + bonusproposerreward * (precommits included / total bonded validator power)
voteMul = 1 - cr - p
powFrac = v.P / total bonded validator power
f = total fees collected
if validator is proposer:
reward = f * voteMul * powFrac + f * p
else:
reward = f * voteMul * powFrac



Description of rewards distribution, not taking into account truncation during multiplication:
(All the below results are described as proportions)

The community pool receives:
(community tax rate)

The block proposer receives a bonus of:
((proposer base rate) + (proposer bonus reward) * (precommits included) / (total bonded validator power))
If the `proposer bonus reward > 0`, it incentivizes the proposer to include all precommits it has received.
Note that for any valid block `2/3 < (precommits included) / (total bonded validator power)`.

All validators, including the block proposer, share what's left, in proportion to their power.

The validator sets a commission rate, which is how much of the received rewards it will take for itself.
The delegators to a validator receive the rest of rewards in proportion to their share of the validator's total power.

As an example, assume the underlying consensus engine selects block proposers in proportion to their power relative to the entire bonded power.
Further, assume all validators are as good at including precommits in their proposed blocks.
Then we can set `(precommits included) / (total bonded validator power)` to constant, and
Then the amortized block reward for the validator is simply (validator power / total bonded power) * (1 - community tax rate) of the total rewards.
Consequently, the reward for a single delegator is:
(delegator proportion of the validator power / validator power) * (validator power / total bonded power) * (1 - community tax rate) * (1 - validator commision rate)
= (delegator proportion of the validator power / total bonded power) * (1 - community tax rate) * (1 - validator commision rate)
