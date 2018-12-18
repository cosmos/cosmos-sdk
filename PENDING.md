## PENDING

BREAKING CHANGES

* Gaia REST API (`gaiacli advanced rest-server`)

* Gaia CLI  (`gaiacli`)
  * [\#810](https://github.com/cosmos/cosmos-sdk/issues/810) Don't fallback to any default values for chain ID.
    - Users need to supply chain ID either via config file or the `--chain-id` flag.
    - Change `chain_id` and `trust_node` in `gaiacli` configuration to `chain-id` and `trust-node` respectively.
  * [\#3156](https://github.com/cosmos/cosmos-sdk/pull/3156) Remove unimplemented `gaiacli init` command

* Gaia

* SDK
  * [\#3064](https://github.com/cosmos/cosmos-sdk/issues/3064) Sanitize `sdk.Coin` denom. Coins denoms are now case insensitive, i.e. 100fooToken equals to 100FOOTOKEN.

* Tendermint


FEATURES

* Gaia REST API (`gaiacli advanced rest-server`)
    * [gaia-lite] [\#2182] Added LCD endpoint for querying redelegations
    * [gov] [\#2479](https://github.com/cosmos/cosmos-sdk/issues/2479) Added governance parameter
    query REST endpoints.

* Gaia CLI  (`gaiacli`)

* Gaia

    * [\#2182] [x/stake] Added querier for querying a single redelegation

* SDK

* Tendermint


IMPROVEMENTS

* Gaia REST API (`gaiacli advanced rest-server`)

* Gaia CLI  (`gaiacli`)

* Gaia

* SDK
  * [\#3093](https://github.com/cosmos/cosmos-sdk/issues/3093) Ante handler does no longer read all accounts in one go when processing signatures as signature
    verification may fail before last signature is checked.

* Tendermint


BUG FIXES

* Gaia REST API (`gaiacli advanced rest-server`)

* Gaia CLI  (`gaiacli`)

* Gaia

  * \#3148 Fix `gaiad export` by adding a boolean to `NewGaiaApp` determining whether or not to load the latest version

* SDK

* Tendermint
