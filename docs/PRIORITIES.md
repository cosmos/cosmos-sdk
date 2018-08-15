# High priority

## Fees

- Collection
  - Simple flat fee set in-config by validators & full nodes - ref [#1921](https://github.com/cosmos/cosmos-sdk/issues/1921)
- Distribution
  - "Piggy bank" fee distribution - ref [#1944](https://github.com/cosmos/cosmos-sdk/pull/1944) (spec)
- Reserve pool
  - Collects in a special field for now, no spending

## Governance v2

- Simple software upgrade proposals
  - Implementation described in [#1079](https://github.com/cosmos/cosmos-sdk/issues/1079)
  - Agree upon a block height to switch to new version

## Staking/Slashing/Stability

- Miscellaneous minor staking issues
  - [List here](https://github.com/cosmos/cosmos-sdk/issues?utf8=%E2%9C%93&q=is%3Aissue+is%3Aopen+label%3Astaking+label%3Aprelaunch)
- Unbonding state for validators https://github.com/cosmos/cosmos-sdk/issues/1676
- Slashing period - ref [#2001](https://github.com/cosmos/cosmos-sdk/pull/2001) (spec)
  - Various other slashing issues needing resolution - ref [#1256](https://github.com/cosmos/cosmos-sdk/issues/1256)
- Update staking/slashing for NextValSet change

## Vesting

- Single `VestingAccount` allowing delegation/voting but no withdrawals
- Ref [#1875](https://github.com/cosmos/cosmos-sdk/pull/1875) (spec)

## Multisig

- Already implemented on TM side, need simple CLI interface

## Other

- Need to update for new ABCI changes - error string, tags are list of lists, proposer in header (Tendermint 0.24?)

## Gas

- Benchmark gas, choose better constants

# Lower priority

## Governance

- Circuit breaker (parameter change proposals, roughly the same implementation)

## Documentation

- gaiad / gaiacli / gaialite documentation!
