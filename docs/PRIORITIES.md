# Post-0.25/GoS Pre-Release

## Staking/Slashing/Stability

- Other slashing issues blocking for launch - [#1256](https://github.com/cosmos/cosmos-sdk/issues/1256)
- Miscellaneous minor staking issues
  - [List here](https://github.com/cosmos/cosmos-sdk/issues?utf8=%E2%9C%93&q=is%3Aissue+is%3Aopen+label%3Astaking+label%3Aprelaunch)
  - Need to figure out scope of work here to estimate time
  - @rigelrozanski to start next
- Consider "tombstone" / "prison" - double-sign and you can never validate again - https://github.com/cosmos/cosmos-sdk/issues/2363

## Multisig

- Need to test changes in https://github.com/cosmos/cosmos-sdk/pull/2165
- Spam prevention - https://github.com/cosmos/cosmos-sdk/issues/2019

## ABCI Changes

- Need to update for new ABCI changes when/if they land - error string, tags are list of lists
- Need to verify correct proposer reward semantics
- CheckEvidence/DeliverEvidence, CheckTx/DeliverTx ordering semantics

## Gas

- Charge for transaction size
- Decide what "one gas" corresponds to (standard hardware benchmarks?)
- More benchmarking
- Consider charging based on maximum depth of IAVL tree iteration
- Test out gas estimation in CLI and LCD and ensure the UX works

## LCD

- Bianje working on implementation of ICS standards
- Additional PR incoming for ICS 22 and ICS 23
- Decide what ought to be ICS-standardized and what ought not to

# Lower priority

## Governance v2

- Circuit breaker - https://github.com/cosmos/cosmos-sdk/issues/926
- Parameter change proposals (roughly the same implementation as circuit breaker)

## Documentation

- gaiad / gaiacli / gaialite documentation!
