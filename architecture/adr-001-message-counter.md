# ADR 001: Global Message Counter

## Context

There is a desire for modules to have a concept of orderings between messages.

One such example is in staking, we currently use an "intra bond tx counter" and
bond height.
The purpose these two serve is to providing an ordering for validators with equal stake,
for usage in the power-ranking of validators.
We can't use address here, as that would create a bad incentive to grind
addresses that optimized the sort function, which lowers the private key's
security.
Instead we order by whose transaction appeared first, as tracked by bondHeight
and intra bond tx counter. 

This logic however should not be unique to staking.
It is very conceivable that many modules in the future will want to be able to
know the ordering of messages / objects after they were initially created.

## Decision

Create a global message counter field of type int64.
Note that with int64's, there is no fear of overflow under normal use,
as it is only getting incremented by one,
and thus has a space of 9 quintillion values to go through.

This counter must be persisted in state, but can just be read and written on 
begin/end block respectively.
This field will get incremented upon every DeliverTx,
regardless if the transaction succeeds or not. 
It must also be incremented within the check state for CheckTx.
The global message ordering field should be set within the context
so that modules can access it.

## Corollary - Intra block ordering
In the event that there is desire to just have an intra block msg counter,
this can easily be derived from the global message counter.
Simply subtract current counter from first global message counter in the block.
Thus the relevant module could easily implement this.

## Status
Proposed

## Consequences

### Positive
* Moves message ordering out of the set of things staking must keep track of
* Abstracts the logic well so other modules can use it

### Negative
* Another thing to implement prelaunch. (Though this should be easy to implement)

### Neutral
