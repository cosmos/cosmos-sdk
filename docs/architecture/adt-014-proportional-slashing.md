# ADR 14: Proportional Slashing

## Changelog

- 2019-10-15: Initial draft

## Context

In Proof of Stake-based chains, centralization of consensus power amongst a small set of validators causes harm to the network due to increased risk of censorship, liveness failure, fork attacks, etc.  However, while this centralization causes a negative externality to the network, it is not directly felt by the stakers contributing towards delegating towards already large validators.  We would like a way to pass on the negative externality cost of centralization onto those large validators and their stakers.

## Decision

### Design

To solve this problem, we will implement a procedure called Proportional Slashing.  The desire is that the larger a validator is, the more they should be slashed.  The first naive attempt is to make a validator's slash percent proportional to their share of consensus voting power.

```
slash_amount = power // power is the faulting validator's voting power.
```

However, this will incentivize validators with large amounts of stake to split up their voting power amongst accounts, so that if they fault, they all get slashed at a lower percent.  The solution to this is to take into account not just a validator's own voting percentage, but also the voting percentage of all the other validators who get slashed in a specified time frame.

```
slash_amount = (power_1 + power_2 + ... + power_n) // where power_i is the voting power of the ith validator faulting in the period
```

Now, if someone splits a validator of 10% into two validators of 5% each which both fault, then they both fault in the same window, they both will still get slashed at the sum 10% amount.

However, an operator might still choose to split up their stake across multiple accounts with hopes that if any of them fault independently, they will not get slashed at the full amount.  In the case that the validators do fault together, they will get slashed the same amount as if they were one entity.  There is no con to splitting up.  However, if operators are going to split up their stake without actually decorrelating their setups, this also causes a negative externality to the network as it fills up validator slots that could have gone to others or increases the commit size.  In order to disincentivize this, we want it to be the case such that splitting up a validator into multiple validators and they fault together is punished more heavily that keeping it as a single validator that faults.

We can achieve this by not only taking into account the sum of the percentages of the validators that faulted, but also the *number* of validators that faulted in the window.  One general form for an equation that fits this desired property looks like this:

```
slash_amount = (sqrt(power_1) + sqrt(power_2) + ... + sqrt(power_n))^2
```

So now, for example, if one validator of 10% faults, it gets a 10% slash, while if two validators of 5% each fault together, they both get a 20% slash ((sqrt(0.05)+sqrt(0.05))^2).

One will note, that this model doesn't differentiate between multiple validators run by the same operators vs validators run by different operators.  This can be seen as an additional benefit in fact.  It incentivizes validators to differentiate their setups from other validators, to avoid having correlated faults with them or else they risk a higher slash.  So for example, operators should avoid using the same popular cloud hosting platforms or using the same Staking as a Service providers.  This will lead to a more resilient and decentralized network.


We can allow some parameterization by multiplying the slash by an on-chain governable parameter.

```
slash_amount = k * (sqrt(power_1) + sqrt(power_2) + ... + sqrt(power_n))^2  // where k is an on-chain parameter for this specific slash type
```

This can be used to weight different types of slashes.  For example, we may want to punish liveness faults 10% as severely as double signs.  The k factor can also be something other than a constant, there is some research on using things like inverse gini coefficients to mitigate some griefing attacks, but this an area for future research.

There can also be minimum and maximums put in place in order to bound the size of the slash percent.

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

- May require computationally expensive sqrt function in state machine
