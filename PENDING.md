## PENDING

BREAKING CHANGES

* Gaia REST API (`gaiacli advanced rest-server`)
  * [gaia-lite] [\#2182] Renamed and merged all redelegations endpoints into `/stake/redelegations`

* Gaia CLI  (`gaiacli`)
  * [\#810](https://github.com/cosmos/cosmos-sdk/issues/810) Don't fallback to any default values for chain ID.
    * Users need to supply chain ID either via config file or the `--chain-id` flag.
    * Change `chain_id` and `trust_node` in `gaiacli` configuration to `chain-id` and `trust-node` respectively.
  * [\#3069](https://github.com/cosmos/cosmos-sdk/pull/3069) `--fee` flag renamed to `--fees` to support multiple coins
  * [\#3156](https://github.com/cosmos/cosmos-sdk/pull/3156) Remove unimplemented `gaiacli init` command

* Gaia
  * https://github.com/cosmos/cosmos-sdk/issues/2838 - Move store keys to constants
  * [\#3162](https://github.com/cosmos/cosmos-sdk/issues/3162) The `--gas` flag now takes `auto` instead of `simulate`
    in order to trigger a simulation of the tx before the actual execution.

* Gaia REST API
  * [\#3176](https://github.com/cosmos/cosmos-sdk/issues/3176) `tx/sign` endpoint now expects `BaseReq` fields as nested object.

* SDK
  * [stake] \#2513 Validator power type from Dec -> Int
  * [stake] \#3233 key and value now contain duplicate fields to simplify code
  * [\#3064](https://github.com/cosmos/cosmos-sdk/issues/3064) Sanitize `sdk.Coin` denom. Coins denoms are now case insensitive, i.e. 100fooToken equals to 100FOOTOKEN.

* Tendermint


FEATURES

* Gaia REST API (`gaiacli advanced rest-server`)
  * [\#3067](https://github.com/cosmos/cosmos-sdk/issues/3067) Add support for fees on transactions
  * [\#3069](https://github.com/cosmos/cosmos-sdk/pull/3069) Add a custom memo on transactions
  * [\#3027](https://github.com/cosmos/cosmos-sdk/issues/3027) Implement
  `/gov/proposals/{proposalID}/proposer` to query for a proposal's proposer.

* Gaia CLI  (`gaiacli`)
  * \#2399 Implement `params` command to query slashing parameters.
  * [\#3027](https://github.com/cosmos/cosmos-sdk/issues/3027) Implement
  `query gov proposer [proposal-id]` to query for a proposal's proposer.

* Gaia
    * [\#2182] [x/stake] Added querier for querying a single redelegation

* SDK
  * \#2996 Update the `AccountKeeper` to contain params used in the context of
  the ante handler.
  * [\#3179](https://github.com/cosmos/cosmos-sdk/pull/3179) New CodeNoSignatures error code.


* Tendermint


IMPROVEMENTS

* Gaia REST API
  * [\#3176](https://github.com/cosmos/cosmos-sdk/issues/3176) Validate tx/sign endpoint POST body.
  * [\#2948](https://github.com/cosmos/cosmos-sdk/issues/2948) Swagger UI now makes requests to light client node

* Gaia CLI  (`gaiacli`)
  * [\#3224](https://github.com/cosmos/cosmos-sdk/pull/3224) Support adding offline public keys to the keystore

* Gaia
  * [\#2186](https://github.com/cosmos/cosmos-sdk/issues/2186) Add Address Interface
  * [\#3158](https://github.com/cosmos/cosmos-sdk/pull/3158) Validate slashing genesis
  * [\#3172](https://github.com/cosmos/cosmos-sdk/pull/3172) Support minimum fees
  in a local testnet.

* SDK
  * [\#3137](https://github.com/cosmos/cosmos-sdk/pull/3137) Add tag documentation
    for each module along with cleaning up a few existing tags in the governance,
    slashing, and staking modules.
  * [\#3093](https://github.com/cosmos/cosmos-sdk/issues/3093) Ante handler does no longer read all accounts in one go when processing signatures as signature
    verification may fail before last signature is checked.

* Tendermint

* CI
  * \#2498 Added macos CI job to CircleCI
  * [#142](https://github.com/tendermint/devops/issues/142) Increased the number of blocks to be tested during multi-sim 

BUG FIXES

* Gaia REST API

* Gaia CLI  (`gaiacli`)
  * \#3141 Fix the bug in GetAccount when `len(res) == 0` and `err == nil`

* Gaia
  * \#3148 Fix `gaiad export` by adding a boolean to `NewGaiaApp` determining whether or not to load the latest version
  * \#3181 Correctly reset total accum update height and jailed-validator bond height / unbonding height on export-for-zero-height
  * [\#3172](https://github.com/cosmos/cosmos-sdk/pull/3172) Fix parsing `gaiad.toml`
  when it already exists.
  * \#3223 Fix unset governance proposal queues when importing state from old chain
  * [#3187](https://github.com/cosmos/cosmos-sdk/issues/3187) Fix `gaiad export`
  by resetting each validator's slashing period.

* SDK

* Tendermint
