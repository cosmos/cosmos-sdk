## PENDING

BREAKING CHANGES

* Gaia REST API (`gaiacli advanced rest-server`)
  * [\#3284](https://github.com/cosmos/cosmos-sdk/issues/3284) Rename the `name`
  field to `from` in the `base_req` body.

* Gaia CLI  (`gaiacli`)
  - [#3399](https://github.com/cosmos/cosmos-sdk/pull/3399) Add `gaiad validate-genesis` command to facilitate checking of genesis files
  - [\#1894](https://github.com/cosmos/cosmos-sdk/issues/1894) `version` prints out short info by default. Add `--long` flag. Proper handling of `--format` flag introduced.

* Gaia

* SDK

* Tendermint


FEATURES

* Gaia REST API

* Gaia CLI  (`gaiacli`)
  * [\#3429](https://github.com/cosmos/cosmos-sdk/issues/3429) Support querying
  for all delegator distribution rewards.
  * \#3449 Proof verification now works with absence proofs

* Gaia
  - [\#3397](https://github.com/cosmos/cosmos-sdk/pull/3397) Implement genesis file sanitization to avoid failures at chain init.
  * \#3428 Run the simulation from a particular genesis state loaded from a file

* SDK
  * \#3270 [x/staking] limit number of ongoing unbonding delegations /redelegations per pair/trio

* Tendermint


IMPROVEMENTS

* Gaia REST API
  * [\#3284](https://github.com/cosmos/cosmos-sdk/issues/3284) Update Gaia Lite
  REST service to support the following:
    * Automatic account number and sequence population when fields are omitted
    * Generate only functionality no longer requires access to a local Keybase
    * `from` field in the `base_req` body can be a Keybase name or account address

* Gaia CLI  (`gaiacli`)

* Gaia
  * [\#3418](https://github.com/cosmos/cosmos-sdk/issues/3418) Add vesting account
  genesis validation checks to `GaiaValidateGenesisState`.
  * [\#3420](https://github.com/cosmos/cosmos-sdk/issues/3420) Added maximum length to governance proposal descriptions and titles

* SDK
  * \#3435 Test that store implementations do not allow nil values

* Tendermint


BUG FIXES

* Gaia REST API

* Gaia CLI  (`gaiacli`)
  - [\#3417](https://github.com/cosmos/cosmos-sdk/pull/3417) Fix `q slashing signing-info` panic by ensuring safety of user input and properly returning not found error
  - [\#3345](https://github.com/cosmos/cosmos-sdk/issues/3345) Upgrade ledger-cosmos-go dependency to v0.9.3 to pull
    https://github.com/ZondaX/ledger-cosmos-go/commit/ed9aa39ce8df31bad1448c72d3d226bf2cb1a8d1 in order to fix a derivation path issue that causes `gaiacli keys add --recover`
    to malfunction.
  - [\#3419](https://github.com/cosmos/cosmos-sdk/pull/3419) Fix `q distr slashes` panic 
  - [\#3453](https://github.com/cosmos/cosmos-sdk/pull/3453) The `rest-server` command didn't respect persistent flags such as `--chain-id` and `--trust-node` if they were
    passed on the command line.

* Gaia

* SDK

* Tendermint
