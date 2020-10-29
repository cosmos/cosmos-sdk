# ADR 033: Governance split votes

## Changelog

- 2020/10/28: Intial draft

## Status

Proposed

## Abstract

This ADR defines a modification to the the governance module that would allow a staker to split their votes into several voting options. For example, it could use 70% of its voting power to vote Yes and 30% of its voting power to vote No.

## Context

Currently, an address can cast a vote with only one options (Yes/No/Abstain/NoWithVeto) and use their full voting power behind that choice.

However, often times the entity owning that address might not be a single individual.  For example, a company might have different stakeholders who want to vote differently, and so it makes sense to allow them to split their voting power.  Another example use case is exchanges.  Many centralized exchanges often stake a portion of their users' tokens in their custody.  Currently, it is not possible for them to do "passthrough voting" and giving their users voting rights over their tokens.  However, with this system, exchanges can poll their users for voting preferences, and then vote on-chain proportionally to the results of the poll.

## Decision

We modify the vote structs to be

```
type WeightedVoteOption struct {
  Option string
  Weight sdk.Dec
}

type Vote struct {
  ProposalID int64
  Voter      sdk.Address
  Options    []WeightedVoteOption
}
```

The `ValidateBasic` of a MsgVote struct would require that
1. The sum of all the Rates is equal to 1.0
2. No Option is repeated

The governance tally function will iterate over all the options in a vote and add to the tally the result of the voter's voting power * the rate for that option.

```
tally() {
    results := map[types.VoteOption]sdk.Dec

    for _, vote := range votes {
        for i, weightedOption := range vote.Options {
            results[weightedOption.Option] += getVotingPower(vote.voter) * weightedOption.Weight
        }
    }
}
```

The CLI command for creating a multi-option vote would be as such:
```sh
simd tx gov vote 1 "yes=0.6,no=0.3,abstain=0.05,no_with_veto=0.05" --from mykey
```

To create a single-option vote a user can do either
```
simd tx gov vote 1 "yes=1" --from mykey
```

or 

```sh
simd tx gov vote 1 yes --from mykey
```

to maintain backwards compatibility.


## Consequences

### Positive
- Can make the voting process more accurate for addresses representing multiple stakeholders, often some of the largest addresses.

### Negative
- 

### Neutral
- Relatively minor change to governance tally function.
