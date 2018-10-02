# Post-0.25 Pre-Release

## Staking/Slashing/Stability

- Other slashing issues blocking for launch - [#1256](https://github.com/cosmos/cosmos-sdk/issues/1256)
- Miscellaneous minor staking issues
  - [List here](https://github.com/cosmos/cosmos-sdk/issues?utf8=%E2%9C%93&q=is%3Aissue+is%3Aopen+label%3Astaking+label%3Aprelaunch)
  - Need to figure out scope of work here to estimate time
  - @rigelrozanski to start next

## Multisig

- Need to test changes in https://github.com/cosmos/cosmos-sdk/pull/2165
- Spam prevention - https://github.com/cosmos/cosmos-sdk/issues/2019

## ABCI Changes

- Need to update for new ABCI changes when/if they land - error string, tags are list of lists
- Need to verify correct proposer reward semantics
- CheckEvidence/DeliverEvidence, CheckTx/DeliverTx ordering semantics

## Gas

- Simple transaction benchmarking work by @jlandrews to inform additional work here
- Integrate @alessio's simulation work into CLI and LCD
- Sanity Checks

## LCD

- Bianje working on implementation ([#2147](https://github.com/cosmos/cosmos-sdk/pull/2147))
  - ICS 0,ICS 1, ICS 20 and ICS 21 implemented in this PR :point_up:
  - @fedekunze, @jackzampolin and @alexanderbez to review
- Additional PR incoming for ICS 22 and ICS 23
- Include [#382](https://github.com/cosmos/cosmos-sdk/issues/382)

# Lower priority

## Governance v2

- Simple software upgrade proposals
  - Implementation described in [#1079](https://github.com/cosmos/cosmos-sdk/issues/1079)
  - Agree upon a block height to switch to new version
- Another Governance proposal from @jaekwon [#2116](https://github.com/cosmos/cosmos-sdk/pull/2116)
- Circuit breaker
- Parameter change proposals (roughly the same implementation as circuit breaker)

## Documentation

- gaiad / gaiacli / gaialite documentation!
