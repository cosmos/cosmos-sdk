## PENDING

BREAKING CHANGES

* Gaia REST API (`gaiacli advanced rest-server`)
  * [\#3069](https://github.com/cosmos/cosmos-sdk/pull/3069) `gas` type on `base_req` changed from `string` to `uint64`

* Gaia CLI  (`gaiacli`)
  * [\#810](https://github.com/cosmos/cosmos-sdk/issues/810) Don't fallback to any default values for chain ID.
    - Users need to supply chain ID either via config file or the `--chain-id` flag.
    - Change `chain_id` and `trust_node` in `gaiacli` configuration to `chain-id` and `trust-node` respectively.

* Gaia

* SDK
  * [\#3064](https://github.com/cosmos/cosmos-sdk/issues/3064) Sanitize `sdk.Coin` denom. Coins denoms are now case insensitive, i.e. 100fooToken equals to 100FOOTOKEN.

* Tendermint


FEATURES

* Gaia REST API (`gaiacli advanced rest-server`)
  * [\#3067](https://github.com/cosmos/cosmos-sdk/issues/3067) Add support for fees on transactions
  * [\#3069](https://github.com/cosmos/cosmos-sdk/pull/3069) Add a custom memo on transactions

* Gaia CLI  (`gaiacli`)
  * [\#3069](https://github.com/cosmos/cosmos-sdk/pull/3069) fee flag now supports multiple coins and was renamed to `--fees`
  * \#2399 Implement `params` command to query slashing parameters.

* Gaia

* SDK

* Tendermint


IMPROVEMENTS

* Gaia REST API (`gaiacli advanced rest-server`)

* Gaia CLI  (`gaiacli`)

* Gaia

* SDK

* Tendermint


BUG FIXES

* Gaia REST API (`gaiacli advanced rest-server`)

* Gaia CLI  (`gaiacli`)

* Gaia

  * \#3148 Fix `gaiad export` by adding a boolean to `NewGaiaApp` determining whether or not to load the latest version

* SDK

* Tendermint
