# Post-0.25/GoS Pre-Release

## Staking / Slashing - Stability

- [Prelaunch Issues](https://github.com/cosmos/cosmos-sdk/issues?utf8=%E2%9C%93&q=is%3Aissue+is%3Aopen+label%3Astaking+label%3Aprelaunch-2.0)

## Multisig

- Need to test changes in https://github.com/cosmos/cosmos-sdk/pull/2165
- Spam prevention - https://github.com/cosmos/cosmos-sdk/issues/2019

## ABCI Changes

- CheckEvidence/DeliverEvidence
- CheckTx/DeliverTx ordering semantics
- ABCI Error string update (only on the SDK side)
- Need to verify correct proposer reward semantics

## Gas

- Write specification and explainer document for Gas in Cosmos
  * Charge for transaction size
  * Decide what "one gas" corresponds to (standard hardware benchmarks?)
  * Consider charging based on maximum depth of IAVL tree iteration
- Test out gas estimation in CLI and LCD and ensure the UX works

## LCD

- Bianje working with Voyager team (@fedekunze) to complete implementation and documentation.

## Documentation

- gaiad / gaiacli
- LCD
- Each module
- Tags [#1780](https://github.com/cosmos/cosmos-sdk/issues/1780)
# Lower priority

- Create some diagrams (see `docs/resources/diagrams/todo.md`) 

## Governance v2

- Circuit breaker - https://github.com/cosmos/cosmos-sdk/issues/926
- Parameter change proposals (roughly the same implementation as circuit breaker)

## Staking / Slashing - Stability

- Consider "tombstone" / "prison" - double-sign and you can never validate again - https://github.com/cosmos/cosmos-sdk/issues/2363
