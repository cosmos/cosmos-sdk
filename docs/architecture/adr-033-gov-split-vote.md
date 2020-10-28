# ADR 033: Governance split votes

## Changelog

- 2020/10/28: Intial draft

## Status

Proposed

## Abstract

This ADR defines a validator to split votes into several pieces. Assuming a group with multisig address vote on a proposal, the members could have different opinion. In this case spliting voting power into pieces could be useful.

## Context

Currently, an address can cast vote with only one of vote options(yes/no/abstrain/no_with_veto) and full votingPower of that address goes to one vote option.

To be more accurate in voting process, allowing the split of voting power could be very helpful.

Assume that there are 100 members associated to a single validator address and they participate in voting process via that address.

60 members cast yes vote, 30 members cast no vote, 5 members abstrain, 5 members no_with_veto.
If they just cast yes vote for 100% of voting power, it won't be fair enough for 40 members especially when that validator's voting power is quite high.

If total voting power is 10000, and on above case we split voting power into 6000, 3000, 500, 500 pieces.

## Decision

We will modify cli command from 
```sh
simd tx gov vote 1 yes --from mykey
```
to 
```sh
simd tx gov vote 1 "yes=60,no=30,abstain=5,no_with_veto=5" --from mykey
```

```
simd tx gov vote 1 yes --from mykey
```
Old command still works and it's automatically converted to like below
```
simd tx gov vote 1 "yes=1" --from mykey
```

If you want to set VoteOption=1, you can omit `=1`.
```
simd tx gov vote 1 "yes,no" --from mykey
``` 
is same as
```
simd tx gov vote 1 "yes=1,no=1" --from mykey
```

## Consequences

### Positive
- It can make more accurate voting process especially for big validators.

### Negative

### Neutral

## References
