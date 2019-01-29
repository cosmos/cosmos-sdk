## PENDING

BREAKING CHANGES

* Gaia REST API (`gaiacli advanced rest-server`)

* Gaia CLI  (`gaiacli`)
  - [#3399](https://github.com/cosmos/cosmos-sdk/pull/3399) Add `gaiad validate-genesis` command to facilitate checking of genesis files

* Gaia

* SDK

* Tendermint


FEATURES

* Gaia REST API

* Gaia CLI  (`gaiacli`)

* Gaia
  - [\#3397](https://github.com/cosmos/cosmos-sdk/pull/3397) Implement genesis file sanitization to avoid failures at chain init.

* SDK
  * \#3270 [x/staking] limit number of ongoing unbonding delegations /redelegations per pair/trio

* Tendermint


IMPROVEMENTS

* Gaia REST API

* Gaia CLI  (`gaiacli`)

* Gaia

* SDK

* Tendermint


BUG FIXES

* Gaia REST API

* Gaia CLI  (`gaiacli`)
  - [\#3417](https://github.com/cosmos/cosmos-sdk/pull/3417) Fix `q slashing signing-info` panic by ensuring safety of user input and properly returning not found error
  - [\#3345](https://github.com/cosmos/cosmos-sdk/issues/3345) Upgrade ledger-cosmos-go dependency to v0.9.3 to pull
    https://github.com/ZondaX/ledger-cosmos-go/commit/ed9aa39ce8df31bad1448c72d3d226bf2cb1a8d1 in order to fix a derivation path issue that causes `gaiacli keys add --recover`
    to malfunction.


* Gaia

* SDK

* Tendermint
