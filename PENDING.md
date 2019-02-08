## PENDING

BREAKING CHANGES

* Gaia REST API (`gaiacli advanced rest-server`)
  * [\#3284](https://github.com/cosmos/cosmos-sdk/issues/3284) Rename the `name`
  field to `from` in the `base_req` body.
  * [\#3485](https://github.com/cosmos/cosmos-sdk/pull/3485) Error responses are now JSON objects.
  * [\#3477][distribution] endpoint changed "all_delegation_rewards" -> "delegator_total_rewards"

* Gaia CLI  (`gaiacli`)
  - [#3399](https://github.com/cosmos/cosmos-sdk/pull/3399) Add `gaiad validate-genesis` command to facilitate checking of genesis files
  - [\#1894](https://github.com/cosmos/cosmos-sdk/issues/1894) `version` prints out short info by default. Add `--long` flag. Proper handling of `--format` flag introduced.
  - [\#3465](https://github.com/cosmos/cosmos-sdk/issues/3465) `gaiacli rest-server` switched back to insecure mode by default:
    - `--insecure` flag is removed.
    - `--tls` is now used to enable secure layer.
  - [\#3451](https://github.com/cosmos/cosmos-sdk/pull/3451) `gaiacli` now returns transactions in plain text including tags.
  - [\#3497](https://github.com/cosmos/cosmos-sdk/issues/3497) `gaiad init` now takes moniker as required arguments, not as parameter.
  * [\#3501](https://github.com/cosmos/cosmos-sdk/issues/3501) Change validator
  address Bech32 encoding to consensus address in `tendermint-validator-set`.

* Gaia
  *  [\#3457](https://github.com/cosmos/cosmos-sdk/issues/3457) Changed governance tally validatorGovInfo to use sdk.Int power instead of sdk.Dec
  *  [\#3495](https://github.com/cosmos/cosmos-sdk/issues/3495) Added Validator Minimum Self Delegation
  *  Reintroduce OR semantics for tx fees

* SDK
  * \#2513 Tendermint updates are adjusted by 10^-6 relative to staking tokens,
  * [\#3487](https://github.com/cosmos/cosmos-sdk/pull/3487) Move HTTP/REST utilities out of client/utils into a new dedicated client/rest package.
  * [\#3490](https://github.com/cosmos/cosmos-sdk/issues/3490) ReadRESTReq() returns bool to avoid callers to write error responses twice.
  * [\#3502](https://github.com/cosmos/cosmos-sdk/pull/3502) Fixes issue when comparing genesis states
  * [\#3514](https://github.com/cosmos/cosmos-sdk/pull/3514) Various clean ups:
    - Replace all GetKeyBase\* functions family in favor of NewKeyBaseFromDir and NewKeyBaseFromHomeFlag.
    - Remove Get prefix from all TxBuilder's getters.
  * [\#3522](https://github.com/cosmos/cosmos-sdk/pull/3522) Get rid of double negatives: Coins.IsNotNegative() -> Coins.IsAnyNegative().
  * \#3561 Don't unnecessarily store denominations in staking

* Tendermint


FEATURES

* Gaia REST API
  * [\#2358](https://github.com/cosmos/cosmos-sdk/issues/2358) Add distribution module REST interface

* Gaia CLI  (`gaiacli`)
  * [\#3429](https://github.com/cosmos/cosmos-sdk/issues/3429) Support querying
  for all delegator distribution rewards.
  * \#3449 Proof verification now works with absence proofs
  * [\#3484](https://github.com/cosmos/cosmos-sdk/issues/3484) Add support
  vesting accounts to the add-genesis-account command.

* Gaia
  - [\#3397](https://github.com/cosmos/cosmos-sdk/pull/3397) Implement genesis file sanitization to avoid failures at chain init.
  * \#3428 Run the simulation from a particular genesis state loaded from a file

* SDK
  * \#3270 [x/staking] limit number of ongoing unbonding delegations /redelegations per pair/trio
  * [\#3477][distribution] new query endpoint "delegator_validators"
  * [\#3514](https://github.com/cosmos/cosmos-sdk/pull/3514) Provided a lazy loading implementation of Keybase that locks the underlying
    storage only for the time needed to perform the required operation. Also added Keybase reference to TxBuilder struct.

* Tendermint


IMPROVEMENTS

* Gaia REST API
  * [\#3284](https://github.com/cosmos/cosmos-sdk/issues/3284) Update Gaia Lite
  REST service to support the following:
    * Automatic account number and sequence population when fields are omitted
    * Generate only functionality no longer requires access to a local Keybase
    * `from` field in the `base_req` body can be a Keybase name or account address
  * [\#3423](https://github.com/cosmos/cosmos-sdk/issues/3423) Allow simulation
  (auto gas) to work with generate only.
  * [\#3514](https://github.com/cosmos/cosmos-sdk/pull/3514) REST server calls to keybase does not lock the underlying storage anymore.
  * [\#3523](https://github.com/cosmos/cosmos-sdk/pull/3523) Added `/tx/encode` endpoint to serialize a JSON tx to base64-encoded Amino.

* Gaia CLI  (`gaiacli`)
  * [\#3476](https://github.com/cosmos/cosmos-sdk/issues/3476) New `withdraw-all-rewards` command to withdraw all delegations rewards for delegators.
  * [\#3497](https://github.com/cosmos/cosmos-sdk/issues/3497) `gaiad gentx` supports `--ip` and `--node-id` flags to override defaults.
  * [\#3518](https://github.com/cosmos/cosmos-sdk/issues/3518) Fix flow in
  `keys add` to show the mnemonic by default.
  * [\#3517](https://github.com/cosmos/cosmos-sdk/pull/3517) Increased test coverage
  * [\#3523](https://github.com/cosmos/cosmos-sdk/pull/3523) Added `tx encode` command to serialize a JSON tx to base64-encoded Amino.

* Gaia
  * [\#3418](https://github.com/cosmos/cosmos-sdk/issues/3418) Add vesting account
  genesis validation checks to `GaiaValidateGenesisState`.
  * [\#3420](https://github.com/cosmos/cosmos-sdk/issues/3420) Added maximum length to governance proposal descriptions and titles
  * [\#3256](https://github.com/cosmos/cosmos-sdk/issues/3256) Add gas consumption
  for tx size in the ante handler.
  * [\#3454](https://github.com/cosmos/cosmos-sdk/pull/3454) Add `--jail-whitelist` to `gaiad export` to enable testing of complex exports
  * [\#3424](https://github.com/cosmos/cosmos-sdk/issues/3424) Allow generation of gentxs with empty memo field.
  * \#3507 General cleanup, removal of unnecessary struct fields, undelegation bugfix, and comment clarification in x/staking and x/slashing

* SDK
  * [\#2605] x/params add subkey accessing
  * [\#2986](https://github.com/cosmos/cosmos-sdk/pull/2986) Store Refactor
  * \#3435 Test that store implementations do not allow nil values
  * \#2509 Sanitize all usage of Dec.RoundInt64()
  * [\#556](https://github.com/cosmos/cosmos-sdk/issues/556) Increase `BaseApp`
  test coverage.

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
  - [\#3441](https://github.com/cosmos/cosmos-sdk/pull/3431) Improved resource management and connection handling (ledger devices). Fixes issue with DER vs BER signatures.

* Gaia
  * [\#3486](https://github.com/cosmos/cosmos-sdk/pull/3486) Use AmountOf in
    vesting accounts instead of zipping/aligning denominations.

* SDK

* Tendermint
