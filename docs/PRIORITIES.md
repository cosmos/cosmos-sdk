# Post-0.25/GoS Pre-Release

## Multisig

- Need to test changes in https://github.com/cosmos/cosmos-sdk/pull/2165
- Spam prevention - https://github.com/cosmos/cosmos-sdk/issues/2019

## ABCI Changes

- CheckEvidence/DeliverEvidence
- CheckTx/DeliverTx ordering semantics

## Gas

- Write specification and explainer document for Gas in Cosmos
  * Charge for transaction size
  * Decide what "one gas" corresponds to (standard hardware benchmarks?)
- Test out gas estimation in CLI and LCD and ensure the UX works

## LCD

- Bianje working with Voyager team (@fedekunze) to complete implementation and documentation.

## Documentation

- gaiad / gaiacli
- LCD
- Each module

# Lower priority

## Governance v2

- Circuit breaker - https://github.com/cosmos/cosmos-sdk/issues/926
- Parameter change proposals (roughly the same implementation as circuit breaker)
