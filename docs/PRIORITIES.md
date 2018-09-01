# High priority

## Fees

- Collection
  - Simple flat fee set in-config by validators & full nodes - ref [#1921](https://github.com/cosmos/cosmos-sdk/issues/1921)
  - @sunnya97 working on implementation
  - _*BLOCKER:*_ Blocked on [tendermint/tendermint#2275](https://github.com/tendermint/tendermint/issues/2275) @ValarDragon
- Distribution
  - "Piggy bank" fee distribution - ref [#1944](https://github.com/cosmos/cosmos-sdk/pull/1944) (spec)
  - @rigelrozanski working on implementation
- EST TIMELINE:
  - Work on fees should be completed in the `v0.25.0` release

## Staking/Slashing/Stability

- Unbonding state for validators (WIP) [#2163](https://github.com/cosmos/cosmos-sdk/pull/2163) @rigelrozanski
  - Needs :eyes: from @chris
  - Should be in `v0.25.0` release
- Slashing period PR - ref [#2122](https://github.com/cosmos/cosmos-sdk/pull/2122)
  - Needs :eyes: from @cwgoes and @jaekwon
  - Should be in `v0.25.0` release
- Other slashing issues blocking for launch - [#1256](https://github.com/cosmos/cosmos-sdk/issues/1256)
- Update staking/slashing for NextValSet change
  - @cwgoes to start next
- Miscellaneous minor staking issues
  - [List here](https://github.com/cosmos/cosmos-sdk/issues?utf8=%E2%9C%93&q=is%3Aissue+is%3Aopen+label%3Astaking+label%3Aprelaunch)
  - Need to figure out scope of work here to estimate time
  - @rigelrozanski to start next

## Vesting

- Single `VestingAccount` allowing delegation/voting but no withdrawals
- Ref [#1875](https://github.com/cosmos/cosmos-sdk/pull/1875) (spec)
- @AdityaSripal working on this.
  - Should be in `v0.25.0` release

## Multisig

- Already implemented on TM side, need simple CLI interface
- @alessio working on the SDK side of things here
- Need to schedule some time with @alessio, @ebuchman and @ValarDragon this week to finalize feature set/implementation plan

## ABCI Changes

- Need to update for new ABCI changes - error string, tags are list of lists, proposer in header (Tendermint 0.24?)
- @cwgoes has done some work here. Should be on `develop` in tendermint w/in next week.
- Include in tendermint `v0.24.0` release?

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
