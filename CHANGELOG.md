# Changelog

## 0.3.0 (March 13, 2017)

BREAKING CHANGES:

- Remove `--data` flag and use `BASECOIN_ROOT` to set the home directory (defaults to `~/.basecoin`)
- Remove `--in-proc` flag and start Tendermint in-process by default (expect Tendermint files in $BASECOIN_ROOT/tendermint).
To start just the ABCI app/server, use `basecoin start --abci-server`.

FEATURES:

- Introduce `basecoin init` and `basecoin unsafe_reset_all` 

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



