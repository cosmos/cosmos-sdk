## PENDING

BREAKING CHANGES

* Gaia REST API (`gaiacli advanced rest-server`)
  * [gaia-lite] [\#2182] Renamed and merged all redelegations endpoints into `/staking/redelegations`
  * [\#3176](https://github.com/cosmos/cosmos-sdk/issues/3176) `tx/sign` endpoint now expects `BaseReq` fields as nested object.
  * [\#2222] all endpoints renamed from `/stake` -> `/staking`
  * [\#1268] `LooseTokens` -> `NotBondedTokens`
  * [\#3289] misc renames:
    * `Validator.UnbondingMinTime` -> `Validator.UnbondingCompletionTime`
    * `Delegation` -> `Value` in `MsgCreateValidator` and `MsgDelegate`
    * `MsgBeginUnbonding` -> `MsgUndelegate`

* Gaia CLI  (`gaiacli`)
  * [\#810](https://github.com/cosmos/cosmos-sdk/issues/810) Don't fallback to any default values for chain ID.
    * Users need to supply chain ID either via config file or the `--chain-id` flag.
    * Change `chain_id` and `trust_node` in `gaiacli` configuration to `chain-id` and `trust-node` respectively.
  * [\#3069](https://github.com/cosmos/cosmos-sdk/pull/3069) `--fee` flag renamed to `--fees` to support multiple coins
  * [\#3156](https://github.com/cosmos/cosmos-sdk/pull/3156) Remove unimplemented `gaiacli init` command
  * [\#2222] `gaiacli tx stake` -> `gaiacli tx staking`, `gaiacli query stake` -> `gaiacli query staking`
  * [\#1894](https://github.com/cosmos/cosmos-sdk/issues/1894) `version` command now shows latest commit, vendor dir hash, and build machine info.
  * [\#3320](https://github.com/cosmos/cosmos-sdk/pull/3320) Ensure all `gaiacli query` commands respect the `--output` and `--indent` flags

* Gaia
  * https://github.com/cosmos/cosmos-sdk/issues/2838 - Move store keys to constants
  * [\#3162](https://github.com/cosmos/cosmos-sdk/issues/3162) The `--gas` flag now takes `auto` instead of `simulate`
    in order to trigger a simulation of the tx before the actual execution.
  * [\#3285](https://github.com/cosmos/cosmos-sdk/pull/3285) New `gaiad tendermint version` to print libs versions
  * [\#1894](https://github.com/cosmos/cosmos-sdk/pull/1894) `version` command now shows latest commit, vendor dir hash, and build machine info.

* SDK
  * [staking] \#2513 Validator power type from Dec -> Int
  * [staking] \#3233 key and value now contain duplicate fields to simplify code
  * [\#3064](https://github.com/cosmos/cosmos-sdk/issues/3064) Sanitize `sdk.Coin` denom. Coins denoms are now case insensitive, i.e. 100fooToken equals to 100FOOTOKEN.
  * [\#3195](https://github.com/cosmos/cosmos-sdk/issues/3195) Allows custom configuration for syncable strategy
  * [\#3242](https://github.com/cosmos/cosmos-sdk/issues/3242) Fix infinite gas
    meter utilization during aborted ante handler executions.
  * [x/distribution] \#3292 Enable or disable withdraw addresses with a parameter in the param store
  * [staking] \#2222 `/stake` -> `/staking` module rename
  * [staking] \#1268 `LooseTokens` -> `NotBondedTokens`
  * [staking] \#1402 Redelegation and unbonding-delegation structs changed to include multiple an array of entries
  * [staking] \#3289 misc renames:
    * `Validator.UnbondingMinTime` -> `Validator.UnbondingCompletionTime`
    * `Delegation` -> `Value` in `MsgCreateValidator` and `MsgDelegate`
    * `MsgBeginUnbonding` -> `MsgUndelegate`
  * [\#3315] Increase decimal precision to 18
  * \#3323 Update to Tendermint 0.29.0
  * [\#3328](https://github.com/cosmos/cosmos-sdk/issues/3328) [x/gov] Remove redundant action tag

* Tendermint
  * [\#3298](https://github.com/cosmos/cosmos-sdk/issues/3298) Upgrade to Tendermint 0.28.0

FEATURES

* Gaia REST API (`gaiacli advanced rest-server`)
  * [\#3067](https://github.com/cosmos/cosmos-sdk/issues/3067) Add support for fees on transactions
  * [\#3069](https://github.com/cosmos/cosmos-sdk/pull/3069) Add a custom memo on transactions
  * [\#3027](https://github.com/cosmos/cosmos-sdk/issues/3027) Implement
  `/gov/proposals/{proposalID}/proposer` to query for a proposal's proposer.

* Gaia CLI  (`gaiacli`)
  * \#2399 Implement `params` command to query slashing parameters.
  * [\#2730](https://github.com/cosmos/cosmos-sdk/issues/2730) Add tx search pagination parameter
  * [\#3027](https://github.com/cosmos/cosmos-sdk/issues/3027) Implement
  `query gov proposer [proposal-id]` to query for a proposal's proposer.
  * [\#3198](https://github.com/cosmos/cosmos-sdk/issues/3198) New `keys add --multisig` flag to store multisig keys locally.
  * [\#3198](https://github.com/cosmos/cosmos-sdk/issues/3198) New `multisign` command to generate multisig signatures.
  * [\#3198](https://github.com/cosmos/cosmos-sdk/issues/3198) New `sign --multisig` flag to enable multisig mode.
  * [\#2715](https://github.com/cosmos/cosmos-sdk/issues/2715) Reintroduce gaia server's insecure mode.
  * [\#3334](https://github.com/cosmos/cosmos-sdk/pull/3334) New `gaiad completion` and `gaiacli completion` to generate Bash/Zsh completion scripts.
  * [\#2607](https://github.com/cosmos/cosmos-sdk/issues/2607) Make `gaiacli config` handle the boolean `indent` flag to beautify commands JSON output.

* Gaia
  * [\#2182] [x/staking] Added querier for querying a single redelegation
  * [\#3305](https://github.com/cosmos/cosmos-sdk/issues/3305) Add support for
    vesting accounts at genesis.
  * [\#3198](https://github.com/cosmos/cosmos-sdk/issues/3198) [x/auth] Add multisig transactions support
  * [\#3198](https://github.com/cosmos/cosmos-sdk/issues/3198) `add-genesis-account` can take both account addresses and key names

* SDK
  - \#3099 Implement F1 fee distribution
  - [\#2926](https://github.com/cosmos/cosmos-sdk/issues/2926) Add TxEncoder to client TxBuilder.
  * \#2694 Vesting account implementation.
  * \#2996 Update the `AccountKeeper` to contain params used in the context of
  the ante handler.
  * [\#3179](https://github.com/cosmos/cosmos-sdk/pull/3179) New CodeNoSignatures error code.
  * \#3319 [x/distribution] Queriers for all distribution state worth querying; distribution query commands

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
  * [\#3172](https://github.com/cosmos/cosmos-sdk/pull/3172) Support minimum fees in a local testnet.
  * [\#3250](https://github.com/cosmos/cosmos-sdk/pull/3250) Refactor integration tests and increase coverage
  * [\#3248](https://github.com/cosmos/cosmos-sdk/issues/3248) Refactor tx fee
  model:
    * Validators specify minimum gas prices instead of minimum fees
    * Clients may provide either fees or gas prices directly
    * The gas prices of a tx must meet a validator's minimum
  * [\#2859](https://github.com/cosmos/cosmos-sdk/issues/2859) Rename `TallyResult` in gov proposals to `FinalTallyResult`
  * [\#3286](https://github.com/cosmos/cosmos-sdk/pull/3286) Fix `gaiad gentx` printout of account's addresses, i.e. user bech32 instead of hex.

* SDK
  * [\#3137](https://github.com/cosmos/cosmos-sdk/pull/3137) Add tag documentation
    for each module along with cleaning up a few existing tags in the governance,
    slashing, and staking modules.
  * [\#3093](https://github.com/cosmos/cosmos-sdk/issues/3093) Ante handler does no longer read all accounts in one go when processing signatures as signature
    verification may fail before last signature is checked.
  * [staking] \#1402 Add for multiple simultaneous redelegations or unbonding-delegations within an unbonding period
  * [staking] \#1268 staking spec rewrite

* Tendermint

* CI
  * \#2498 Added macos CI job to CircleCI
  * [#142](https://github.com/tendermint/devops/issues/142) Increased the number of blocks to be tested during multi-sim
  * [#147](https://github.com/tendermint/devops/issues/142) Added docker image build to CI

BUG FIXES

* Gaia REST API

* Gaia CLI  (`gaiacli`)
  * \#3141 Fix the bug in GetAccount when `len(res) == 0` and `err == nil`
  * [\#810](https://github.com/cosmos/cosmos-sdk/pull/3316) Fix regression in gaiacli config file handling

* Gaia
  * \#3148 Fix `gaiad export` by adding a boolean to `NewGaiaApp` determining whether or not to load the latest version
  * \#3181 Correctly reset total accum update height and jailed-validator bond height / unbonding height on export-for-zero-height
  * [\#3172](https://github.com/cosmos/cosmos-sdk/pull/3172) Fix parsing `gaiad.toml`
  when it already exists.
  * \#3223 Fix unset governance proposal queues when importing state from old chain
  * [#3187](https://github.com/cosmos/cosmos-sdk/issues/3187) Fix `gaiad export`
  by resetting each validator's slashing period.
  * [##3336](https://github.com/cosmos/cosmos-sdk/issues/3336) Ensure all SDK
  messages have their signature bytes contain canonical fields `value` and `type`.

* SDK

* Tendermint
