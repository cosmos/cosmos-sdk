<!--
order: 1
-->

# Concepts

Rewards are calculated per period, and the period is updated each time a validator's delegation changes, for example when it receives a new delegation.
The rewards for a single validator can then be calculated by taking the total rewards for the period before the delegation started, minus the current total rewards.
Further details can be found in [this document](/docs/spec/fee_distribution/f1_fee_distr.pdf).

In the Cosmos SDK implementation of the F1 fee distribution all rewards are distributed to a delegator when they change or remove their delegation, or when they request a withdrawal.
At that time, all slashes to the validator that occurred during the current delegation are iterated over.

The commission to the validator is paid when the validator is removed, or when the validator requests a withdrawal.
The commission is calculated and incremented at every `BeginBlock`, as part of updating accumulated fee amounts.

## Reference Counting in F1 Fee Distribution

In F1 fee distribution, in order to calculate the rewards a delegator ought to receive when they
withdraw their delegation, we must read the terms of the summation of rewards divided by tokens from
the period which they ended when they delegated, and the final period (created when they withdraw).

Additionally, as slashes change the amount of tokens a delegation will have (but we calculate this lazily,
only when a delegator un-delegates), we must calculate rewards in separate periods before / after any slashes
which occurred in between when a delegator delegated and when they withdrew their rewards. Thus slashes, like
delegations, reference the period which was ended by the slash event.

All stored historical rewards records for periods which are no longer referenced by any delegations
or any slashes can thus be safely removed, as they will never be read (future delegations and future
slashes will always reference future periods). This is implemented by tracking a `ReferenceCount`
along with each historical reward storage entry. Each time a new object (delegation or slash)
is created which might need to reference the historical record, the reference count is incremented.
Each time one object which previously needed to reference the historical record is deleted, the reference
count is decremented. If the reference count hits zero, the historical record is deleted.
