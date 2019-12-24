# ADR 14: Proportional Slashing

## Changelog

- 2019-10-15: Initial draft

## Context

In Proof of Stake-based chains, centralization of consensus power amongst a small set of validators can cause harm to the network due to increased risk of censorship, liveness failure, fork attacks, etc.  However, while this centralization causes a negative externality to the network, it is not directly felt by the delegators contributing towards delegating towards already large validators.  We would like a way to pass on the negative externality cost of centralization onto those large validators and their delegators.

## Decision

### Design

To solve this problem, we will implement a procedure called Proportional Slashing.  The desire is that the larger a validator is, the more they should be slashed.  The first naive attempt is to make a validator's slash percent proportional to their share of consensus voting power.

```
slash_amount = k * power // power is the faulting validator's voting power and k is some on-chain constant
```

However, this will incentivize validators with large amounts of stake to split up their voting power amongst accounts, so that if they fault, they all get slashed at a lower percent.  The solution to this is to take into account not just a validator's own voting percentage, but also the voting percentage of all the other validators who get slashed in a specified time frame.

```
slash_amount = k * (power_1 + power_2 + ... + power_n) // where power_i is the voting power of the ith validator faulting in the specified time frame and k is some on-chain constant
```

Now, if someone splits a validator of 10% into two validators of 5% each which both fault, then they both fault in the same time frame, they both will still get slashed at the sum 10% amount.

However, an operator might still choose to split up their stake across multiple accounts with hopes that if any of them fault independently, they will not get slashed at the full amount.  In the case that the validators do fault together, they will get slashed the same amount as if they were one entity.  There is no con to splitting up.  However, if operators are going to split up their stake without actually decorrelating their setups, this also causes a negative externality to the network as it fills up validator slots that could have gone to others or increases the commit size.  In order to disincentivize this, we want it to be the case such that splitting up a validator into multiple validators and they fault together is punished more heavily that keeping it as a single validator that faults.

We can achieve this by not only taking into account the sum of the percentages of the validators that faulted, but also the *number* of validators that faulted in the window.  One general form for an equation that fits this desired property looks like this:

```
slash_amount = k * ((power_1)^(1/r) + (power_2)^(1/r) + ... + (power_n)^(1/r))^r // where k and r are both on-chain constants
```

So now, for example, assuming k=1 and r=2, if one validator of 10% faults, it gets a 10% slash, while if two validators of 5% each fault together, they both get a 20% slash ((sqrt(0.05)+sqrt(0.05))^2).

#### Correlation across non-sybil validators

One will note, that this model doesn't differentiate between multiple validators run by the same operators vs validators run by different operators.  This can be seen as an additional benefit in fact.  It incentivizes validators to differentiate their setups from other validators, to avoid having correlated faults with them or else they risk a higher slash.  So for example, operators should avoid using the same popular cloud hosting platforms or using the same Staking as a Service providers.  This will lead to a more resilient and decentralized network.

#### Parameterization

The value of k and r can be different for different types of slashable faults.  For example, we may want to punish liveness faults 10% as severely as double signs.

There can also be minimum and maximums put in place in order to bound the size of the slash percent.

#### Griefing

Griefing, the act of intentionally being slashed to make another's slash worse, could be a concern here.  However, using the protocol described here, the attacker could not substantially grief without getting slashed a substantial amount themselves.  The larger the validator is, the more heavily it can impact the slash, it needs to be non-trivial to have a significant impact on the slash percent.  Furthermore, the larger the grief, the griefer loses quadratically more.

It may also be possible to, rather than the k and r factors being constants, perhaps using an inverse gini coefficient may mitigate some griefing attacks, but this an area for future research.

### Implementation

In the slashing module, we will add two queues that will track all of the recent slash events.  For double sign faults, we will define "recent slashes" as ones that have occured within the last `unbonding period`.  For liveness faults, we will define "recent slashes" as ones that have occured withing the last `jail period`.

```
type SlashEvent struct {
    Address                     sdk.ValAddress
    SqrtValidatorVotingPercent  sdk.Dec
    SlashedSoFar                sdk.Dec
}
```

These slash events will be pruned from the queue once they are older than their respective "recent slash period".

Whenever a new slash occurs, a `SlashEvent` struct is created with the faulting validator's voting percent and a `SlashedSoFar` of 0.  Because recent slash events are pruned before the unbonding period and unjail period expires, it should not be possible for the same validator to have multiple SlashEvents in the same Queue at the same time.

We then will iterate over all the SlashEvents in the queue, adding their `SqrtValidatorVotingPercent` and squaring the result to calculate the new percent to slash all the validators in the queue at, using the "Square of Sum of Roots" formula introduced above.

Once we have the `NewSlashPercent`, we then iterate over all the `SlashEvent`s in the queue once again, and if `NewSlashPercent > SlashedSoFar` for that SlashEvent, we call the `staking.Slash(slashEvent.Address, slashEvent.Power, Math.Min(Math.Max(minSlashPercent, NewSlashPercent - SlashedSoFar), maxSlashPercent)` (we pass in the power of the validator before any slashes occured, so that we slash the right amount of tokens).  We then set `SlashEvent.SlashedSoFar` amount to `NewSlashPercent`.


## Status

Proposed

## Consequences

### Positive

- Increases decentralization by disincentivizing delegating to large validators
- Incentivizes Decorrelation of Validators
- More severely punishes attacks than accidental faults

### Negative

- May require computationally expensive root function in state machine
