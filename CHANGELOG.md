# Changelog

## 0.2.0 (March 6, 2017)

BREAKING CHANGES:

- Update to ABCI v0.4.0 and Tendermint v0.9.0
- Coins are specified on the CLI as `Xcoin`, eg. `5gold`
- `Cost` is now `Fee`

FEATURES:

- CLI for sending transactions and querying the state, 
designed to be easily extensible as plugins are implemented 
- Run Basecoin in-process with Tendermint
- Add `/account` path in Query
- IBC plugin for InterBlockchain Communication
- Demo script of IBC between two chains

IMPROVEMENTS:

- Use new Tendermint `/commit` endpoint for crafting IBC transactions
- More unit tests
- Use go-crypto S structs and go-data for more standard JSON
- Demo uses fewer sleeps

BUG FIXES:

- Various little fixes in coin arithmetic
- More commit validation in IBC
- Return results from transactions

## PreHistory

##### January 14-18, 2017

- Update to Tendermint v0.8.0
- Cleanup a bit and release blog post

##### September 22, 2016

- Basecoin compiles again



