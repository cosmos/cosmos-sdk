## PENDING

BREAKING CHANGES

* Gaia REST API (`gaiacli advanced rest-server`)

* Gaia CLI  (`gaiacli`)

* Gaia

* SDK

* Tendermint


FEATURES

* Gaia REST API (`gaiacli advanced rest-server`)
  * [gov] [\#2479](https://github.com/cosmos/cosmos-sdk/issues/2479) Added governance parameter
    query REST endpoints.

* Gaia CLI  (`gaiacli`)
    * [cli] [\#2569](https://github.com/cosmos/cosmos-sdk/pull/2569) Add commands to query validator unbondings and redelegations
    * [cli] [\#2569](https://github.com/cosmos/cosmos-sdk/pull/2569) Add commands to query validator unbondings and redelegations
    * [cli] [\#2524](https://github.com/cosmos/cosmos-sdk/issues/2524) Add support offline mode to `gaiacli tx sign`. Lookups are not performed if the flag `--offline` is on.
    * [cli] [\#2558](https://github.com/cosmos/cosmos-sdk/issues/2558) Rename --print-sigs to --validate-signatures. It now performs a complete set of sanity checks and reports to the user. Also added --print-signature-only to print the signature only, not the whole transaction.
    * [gov][cli] [\#2479](https://github.com/cosmos/cosmos-sdk/issues/2479) Add governance
    parameter query commands.

* Gaia
  * [x/gov] [#2479](https://github.com/cosmos/cosmos-sdk/issues/2479) Added querier for getting
    governance parameters.

* SDK

* Tendermint


IMPROVEMENTS

* Gaia REST API (`gaiacli advanced rest-server`)

* Gaia CLI  (`gaiacli`)
  * [\#2749](https://github.com/cosmos/cosmos-sdk/pull/2749) Add --chain-id flag to gaiad testnet

* Gaia

* SDK
 - [x/mock/simulation] [\#2720] major cleanup, introduction of helper objects, reorganization

* Tendermint


BUG FIXES

* Gaia REST API (`gaiacli advanced rest-server`)

* Gaia CLI  (`gaiacli`)

* Gaia
  * [\#2723] Use `cosmosvalcons` Bech32 prefix in `tendermint show-address`
  * [\#2742](https://github.com/cosmos/cosmos-sdk/issues/2742) Fix time format of TimeoutCommit override 
  
* SDK

* Tendermint
  * [\#2797](https://github.com/tendermint/tendermint/pull/2797) AddressBook requires addresses to have IDs; Do not crap out immediately after sending pex addrs in seed mode
