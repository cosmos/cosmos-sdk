# Changelog

## 0.26.0

BREAKING CHANGES

* Gaia
  * [gaiad init] [\#2602](https://github.com/cosmos/cosmos-sdk/issues/2602) New genesis workflow

* SDK
  * [simulation] [\#2665](https://github.com/cosmos/cosmos-sdk/issues/2665) only argument to simulation.Invariant is now app

* Tendermint
  * Upgrade to version 0.26.0

FEATURES

* Gaia CLI  (`gaiacli`)
  * [cli] [\#2569](https://github.com/cosmos/cosmos-sdk/pull/2569) Add commands to query validator unbondings and redelegations
  * [cli] [\#2569](https://github.com/cosmos/cosmos-sdk/pull/2569) Add commands to query validator unbondings and redelegations
  * [cli] [\#2524](https://github.com/cosmos/cosmos-sdk/issues/2524) Add support offline mode to `gaiacli tx sign`. Lookups are not performed if the flag `--offline` is on.
  * [cli] [\#2558](https://github.com/cosmos/cosmos-sdk/issues/2558) Rename --print-sigs to --validate-signatures. It now performs a complete set of sanity checks and reports to the user. Also added --print-signature-only to print the signature only, not the whole transaction.
  * [cli] [\#2704](https://github.com/cosmos/cosmos-sdk/pull/2704) New add-genesis-account convenience command to populate genesis.json with genesis accounts.

* SDK
  * [\#1336](https://github.com/cosmos/cosmos-sdk/issues/1336) Mechanism for SDK Users to configure their own Bech32 prefixes instead of using the default cosmos prefixes.

IMPROVEMENTS

* Gaia
 * [\#2637](https://github.com/cosmos/cosmos-sdk/issues/2637) [x/gov] Switched inactive and active proposal queues to an iterator based queue

* SDK
 * [\#2573](https://github.com/cosmos/cosmos-sdk/issues/2573) [x/distribution] add accum invariance
 * [\#2556](https://github.com/cosmos/cosmos-sdk/issues/2556) [x/mock/simulation] Fix debugging output
 * [\#2396](https://github.com/cosmos/cosmos-sdk/issues/2396) [x/mock/simulation] Change parameters to get more slashes
 * [\#2617](https://github.com/cosmos/cosmos-sdk/issues/2617) [x/mock/simulation] Randomize all genesis parameters
 * [\#2669](https://github.com/cosmos/cosmos-sdk/issues/2669) [x/stake] Added invarant check to make sure validator's power aligns with its spot in the power store.
 * [\#1924](https://github.com/cosmos/cosmos-sdk/issues/1924) [x/mock/simulation] Use a transition matrix for block size
 * [\#2660](https://github.com/cosmos/cosmos-sdk/issues/2660) [x/mock/simulation] Staking transactions get tested far more frequently
 * [\#2610](https://github.com/cosmos/cosmos-sdk/issues/2610) [x/stake] Block redelegation to and from the same validator
 * [\#2652](https://github.com/cosmos/cosmos-sdk/issues/2652) [x/auth] Add benchmark for get and set account
 * [\#2685](https://github.com/cosmos/cosmos-sdk/issues/2685) [store] Add general merkle absence proof (also for empty substores)
 * [\#2708](https://github.com/cosmos/cosmos-sdk/issues/2708) [store] Disallow setting nil values

BUG FIXES

* Gaia
 * [\#2670](https://github.com/cosmos/cosmos-sdk/issues/2670) [x/stake] fixed incorrect `IterateBondedValidators` and split into two functions: `IterateBondedValidators` and `IterateLastBlockConsValidators`
 * [\#2691](https://github.com/cosmos/cosmos-sdk/issues/2691) Fix local testnet creation by using a single canonical genesis time
 - [\#2670](https://github.com/cosmos/cosmos-sdk/issues/2670) [x/stake] fixed incorrent `IterateBondedValidators` and split into two functions: `IterateBondedValidators` and `IterateLastBlockConsValidators`
 - [\#2648](https://github.com/cosmos/cosmos-sdk/issues/2648) [gaiad] Fix `gaiad export` / `gaiad import` consistency, test in CI

* SDK
 * [\#2625](https://github.com/cosmos/cosmos-sdk/issues/2625) [x/gov] fix AppendTag function usage error
 * [\#2677](https://github.com/cosmos/cosmos-sdk/issues/2677) [x/stake, x/distribution] various staking/distribution fixes as found by the simulator
 * [\#2674](https://github.com/cosmos/cosmos-sdk/issues/2674) [types] Fix coin.IsLT() impl, coins.IsLT() impl, and renamed coins.Is\* to coins.IsAll\* (see [\#2686](https://github.com/cosmos/cosmos-sdk/issues/2686))
 * [\#2711](https://github.com/cosmos/cosmos-sdk/issues/2711) [x/stake] Add commission data to `MsgCreateValidator` signature bytes.
 * Temporarily disable insecure mode for Gaia Lite

## 0.25.0

*October 24th, 2018*

BREAKING CHANGES

* Gaia REST API (`gaiacli advanced rest-server`)
    * [x/stake] Validator.Owner renamed to Validator.Operator
    * [\#595](https://github.com/cosmos/cosmos-sdk/issues/595) Connections to the REST server are now secured using Transport Layer Security by default. The --insecure flag is provided to switch back to insecure HTTP.
    * [gaia-lite] [\#2258](https://github.com/cosmos/cosmos-sdk/issues/2258) Split `GET stake/delegators/{delegatorAddr}` into `GET stake/delegators/{delegatorAddr}/delegations`, `GET stake/delegators/{delegatorAddr}/unbonding_delegations` and `GET stake/delegators/{delegatorAddr}/redelegations`

* Gaia CLI  (`gaiacli`)
    * [x/stake] Validator.Owner renamed to Validator.Operator
    * [cli] unsafe_reset_all, show_validator, and show_node_id have been renamed to unsafe-reset-all, show-validator, and show-node-id
    * [cli] [\#1983](https://github.com/cosmos/cosmos-sdk/issues/1983) --print-response now defaults to true in commands that create and send a transaction
    * [cli] [\#1983](https://github.com/cosmos/cosmos-sdk/issues/1983) you can now pass --pubkey or --address to gaiacli keys show to return a plaintext representation of the key's address or public key for use with other commands
    * [cli] [\#2061](https://github.com/cosmos/cosmos-sdk/issues/2061) changed proposalID in governance REST endpoints to proposal-id
    * [cli] [\#2014](https://github.com/cosmos/cosmos-sdk/issues/2014) `gaiacli advanced` no longer exists - to access `ibc`, `rest-server`, and `validator-set` commands use `gaiacli ibc`, `gaiacli rest-server`, and `gaiacli tendermint`, respectively
    * [makefile] `get_vendor_deps` no longer updates lock file it just updates vendor directory. Use `update_vendor_deps` to update the lock file. [#2152](https://github.com/cosmos/cosmos-sdk/pull/2152)
    * [cli] [\#2221](https://github.com/cosmos/cosmos-sdk/issues/2221) All commands that
    utilize a validator's operator address must now use the new Bech32 prefix,
    `cosmosvaloper`.
    * [cli] [\#2190](https://github.com/cosmos/cosmos-sdk/issues/2190) `gaiacli init --gen-txs` is now `gaiacli init --with-txs` to reduce confusion
    * [cli] [\#2073](https://github.com/cosmos/cosmos-sdk/issues/2073) --from can now be either an address or a key name
    * [cli] [\#1184](https://github.com/cosmos/cosmos-sdk/issues/1184) Subcommands reorganisation, see [\#2390](https://github.com/cosmos/cosmos-sdk/pull/2390) for a comprehensive list of changes.
    * [cli] [\#2524](https://github.com/cosmos/cosmos-sdk/issues/2524) Add support offline mode to `gaiacli tx sign`. Lookups are not performed if the flag `--offline` is on.
    * [cli] [\#2570](https://github.com/cosmos/cosmos-sdk/pull/2570) Add commands to query deposits on proposals

* Gaia
    * Make the transient store key use a distinct store key. [#2013](https://github.com/cosmos/cosmos-sdk/pull/2013)
    * [x/stake] [\#1901](https://github.com/cosmos/cosmos-sdk/issues/1901) Validator type's Owner field renamed to Operator; Validator's GetOwner() renamed accordingly to comply with the SDK's Validator interface.
    * [docs] [#2001](https://github.com/cosmos/cosmos-sdk/pull/2001) Update slashing spec for slashing period
    * [x/stake, x/slashing] [#1305](https://github.com/cosmos/cosmos-sdk/issues/1305) - Rename "revoked" to "jailed"
    * [x/stake] [#1676] Revoked and jailed validators put into the unbonding state
    * [x/stake] [#1877] Redelegations/unbonding-delegation from unbonding validator have reduced time
    * [x/slashing] [\#1789](https://github.com/cosmos/cosmos-sdk/issues/1789) Slashing changes for Tendermint validator set offset (NextValSet)
    * [x/stake] [\#2040](https://github.com/cosmos/cosmos-sdk/issues/2040) Validator
    operator type has now changed to `sdk.ValAddress`
    * [x/stake] [\#2221](https://github.com/cosmos/cosmos-sdk/issues/2221) New
    Bech32 prefixes have been introduced for a validator's consensus address and
    public key: `cosmosvalcons` and `cosmosvalconspub` respectively. Also, existing Bech32 prefixes have been
    renamed for accounts and validator operators:
      * `cosmosaccaddr` / `cosmosaccpub` => `cosmos` / `cosmospub`
      * `cosmosvaladdr` / `cosmosvalpub` => `cosmosvaloper` / `cosmosvaloperpub`
    * [x/stake] [#1013] TendermintUpdates now uses transient store
    * [x/stake] [\#2435](https://github.com/cosmos/cosmos-sdk/issues/2435) Remove empty bytes from the ValidatorPowerRank store key
    * [x/gov] [\#2195](https://github.com/cosmos/cosmos-sdk/issues/2195) Governance uses BFT Time
    * [x/gov] [\#2256](https://github.com/cosmos/cosmos-sdk/issues/2256) Removed slashing for governance non-voting validators
    * [simulation] [\#2162](https://github.com/cosmos/cosmos-sdk/issues/2162) Added back correct supply invariants
    * [x/slashing] [\#2430](https://github.com/cosmos/cosmos-sdk/issues/2430) Simulate more slashes, check if validator is jailed before jailing
    * [x/stake] [\#2393](https://github.com/cosmos/cosmos-sdk/issues/2393) Removed `CompleteUnbonding` and `CompleteRedelegation` Msg types, and instead added unbonding/redelegation queues to endblocker
    * [x/mock/simulation] [\#2501](https://github.com/cosmos/cosmos-sdk/issues/2501) Simulate transactions & invariants for fee distribution, and fix bugs discovered in the process
      * [x/auth] Simulate random fee payments
      * [cmd/gaia/app] Simulate non-zero inflation
      * [x/stake] Call hooks correctly in several cases related to delegation/validator updates
      * [x/stake] Check full supply invariants, including yet-to-be-withdrawn fees
      * [x/stake] Remove no-longer-in-use store key
      * [x/slashing] Call hooks correctly when a validator is slashed
      * [x/slashing] Truncate withdrawals (unbonding, redelegation) and burn change
      * [x/mock/simulation] Ensure the simulation cannot set a proposer address of nil
      * [x/mock/simulation] Add more event logs on begin block / end block for clarity
      * [x/mock/simulation] Correctly set validator power in abci.RequestBeginBlock
      * [x/minting] Correctly call stake keeper to track inflated supply
      * [x/distribution] Sanity check for nonexistent rewards
      * [x/distribution] Truncate withdrawals and return change to the community pool
      * [x/distribution] Add sanity checks for incorrect accum / total accum relations
      * [x/distribution] Correctly calculate total power using Tendermint updates
      * [x/distribution] Simulate withdrawal transactions
      * [x/distribution] Fix a bug where the fee pool was not correctly tracked on WithdrawDelegatorRewardsAll
    * [x/stake] [\#1673](https://github.com/cosmos/cosmos-sdk/issues/1673) Validators are no longer deleted until they can no longer possibly be slashed
    * [\#1890](https://github.com/cosmos/cosmos-sdk/issues/1890) Start chain with initial state + sequence of transactions
      * [cli] Rename `gaiad init gentx` to `gaiad gentx`.
      * [cli] Add `--skip-genesis` flag to `gaiad init` to prevent `genesis.json` generation.
      * Drop `GenesisTx` in favor of a signed `StdTx` with only one `MsgCreateValidator` message.
      * [cli] Port `gaiad init` and `gaiad testnet` to work with `StdTx` genesis transactions.
      * [cli] Add `--moniker` flag to `gaiad init` to override moniker when generating `genesis.json` - i.e. it takes effect when running with the `--with-txs` flag, it is ignored otherwise.

* SDK
    * [core] [\#2219](https://github.com/cosmos/cosmos-sdk/issues/2219) Update to Tendermint 0.24.0
      * Validator set updates delayed by one block
      * BFT timestamp that can safely be used by applications
      * Fixed maximum block size enforcement
    * [core] [\#1807](https://github.com/cosmos/cosmos-sdk/issues/1807) Switch from use of rational to decimal
    * [types] [\#1901](https://github.com/cosmos/cosmos-sdk/issues/1901) Validator interface's GetOwner() renamed to GetOperator()
    * [x/slashing] [#2122](https://github.com/cosmos/cosmos-sdk/pull/2122) - Implement slashing period
    * [types] [\#2119](https://github.com/cosmos/cosmos-sdk/issues/2119) Parsed error messages and ABCI log errors to make     them more human readable.
    * [types] [\#2407](https://github.com/cosmos/cosmos-sdk/issues/2407) MulInt method added to big decimal in order to improve efficiency of slashing
    * [simulation] Rename TestAndRunTx to Operation [#2153](https://github.com/cosmos/cosmos-sdk/pull/2153)
    * [simulation] Remove log and testing.TB from Operation and Invariants, in favor of using errors [\#2282](https://github.com/cosmos/cosmos-sdk/issues/2282)
    * [simulation] Remove usage of keys and addrs in the types, in favor of simulation.Account [\#2384](https://github.com/cosmos/cosmos-sdk/issues/2384)
    * [tools] Removed gocyclo [#2211](https://github.com/cosmos/cosmos-sdk/issues/2211)
    * [baseapp] Remove `SetTxDecoder` in favor of requiring the decoder be set in baseapp initialization. [#1441](https://github.com/cosmos/cosmos-sdk/issues/1441)
    * [baseapp] [\#1921](https://github.com/cosmos/cosmos-sdk/issues/1921) Add minimumFees field to BaseApp.
    * [store] Change storeInfo within the root multistore to use tmhash instead of ripemd160 [\#2308](https://github.com/cosmos/cosmos-sdk/issues/2308)
    * [codec] [\#2324](https://github.com/cosmos/cosmos-sdk/issues/2324) All referrences to wire have been renamed to codec. Additionally, wire.NewCodec is now codec.New().
    * [types] [\#2343](https://github.com/cosmos/cosmos-sdk/issues/2343) Make sdk.Msg have a names field, to facilitate automatic tagging.
    * [baseapp] [\#2366](https://github.com/cosmos/cosmos-sdk/issues/2366) Automatically add action tags to all messages
    * [x/auth] [\#2377](https://github.com/cosmos/cosmos-sdk/issues/2377) auth.StdSignMsg -> txbuilder.StdSignMsg
    * [x/staking] [\#2244](https://github.com/cosmos/cosmos-sdk/issues/2244) staking now holds a consensus-address-index instead of a consensus-pubkey-index
    * [x/staking] [\#2236](https://github.com/cosmos/cosmos-sdk/issues/2236) more distribution hooks for distribution
    * [x/stake] [\#2394](https://github.com/cosmos/cosmos-sdk/issues/2394) Split up UpdateValidator into distinct state transitions applied only in EndBlock
    * [x/slashing] [\#2480](https://github.com/cosmos/cosmos-sdk/issues/2480) Fix signing info handling bugs & faulty slashing
    * [x/stake] [\#2412](https://github.com/cosmos/cosmos-sdk/issues/2412) Added an unbonding validator queue to EndBlock to automatically update validator.Status when finished Unbonding
    * [x/stake] [\#2500](https://github.com/cosmos/cosmos-sdk/issues/2500) Block conflicting redelegations until we add an index
    * [x/params] Global Paramstore refactored
    * [types] [\#2506](https://github.com/cosmos/cosmos-sdk/issues/2506) sdk.Dec MarshalJSON now marshals as a normal Decimal, with 10 digits of decimal precision
    * [x/stake] [\#2508](https://github.com/cosmos/cosmos-sdk/issues/2508) Utilize Tendermint power for validator power key
    * [x/stake] [\#2531](https://github.com/cosmos/cosmos-sdk/issues/2531) Remove all inflation logic
    * [x/mint] [\#2531](https://github.com/cosmos/cosmos-sdk/issues/2531) Add minting module and inflation logic
    * [x/auth] [\#2540](https://github.com/cosmos/cosmos-sdk/issues/2540) Rename `AccountMapper` to `AccountKeeper`.
    * [types] [\#2456](https://github.com/cosmos/cosmos-sdk/issues/2456) Renamed msg.Name() and msg.Type() to msg.Type() and msg.Route() respectively

* Tendermint
  * Update tendermint version from v0.23.0 to v0.25.0, notable changes
    * Mempool now won't build too large blocks, or too computationally expensive blocks
    * Maximum tx sizes and gas are now removed, and are implicitly the blocks maximums
    * ABCI validators no longer send the pubkey. The pubkey is only sent in validator updates
    * Validator set changes are now delayed by one block
    * Block header now includes the next validator sets hash
    * BFT time is implemented
    * Secp256k1 signature format has changed
    * There is now a threshold multisig format
    * See the [tendermint changelog](https://github.com/tendermint/tendermint/blob/master/CHANGELOG.md) for other changes.

FEATURES

* Gaia REST API (`gaiacli advanced rest-server`)
  * [gaia-lite] Endpoints to query staking pool and params
  * [gaia-lite] [\#2110](https://github.com/cosmos/cosmos-sdk/issues/2110) Add support for `simulate=true` requests query argument to endpoints that send txs to run simulations of transactions
  * [gaia-lite] [\#966](https://github.com/cosmos/cosmos-sdk/issues/966) Add support for `generate_only=true` query argument to generate offline unsigned transactions
  * [gaia-lite] [\#1953](https://github.com/cosmos/cosmos-sdk/issues/1953) Add /sign endpoint to sign transactions generated with `generate_only=true`.
  * [gaia-lite] [\#1954](https://github.com/cosmos/cosmos-sdk/issues/1954) Add /broadcast endpoint to broadcast transactions signed by the /sign endpoint.
  * [gaia-lite] [\#2113](https://github.com/cosmos/cosmos-sdk/issues/2113) Rename `/accounts/{address}/send` to `/bank/accounts/{address}/transfers`, rename `/accounts/{address}` to `/auth/accounts/{address}`, replace `proposal-id` with `proposalId` in all gov endpoints
  * [gaia-lite] [\#2478](https://github.com/cosmos/cosmos-sdk/issues/2478) Add query gov proposal's deposits endpoint
  * [gaia-lite] [\#2477](https://github.com/cosmos/cosmos-sdk/issues/2477) Add query validator's outgoing redelegations and unbonding delegations endpoints

* Gaia CLI  (`gaiacli`)
  * [cli] Cmds to query staking pool and params
  * [gov][cli] [\#2062](https://github.com/cosmos/cosmos-sdk/issues/2062) added `--proposal` flag to `submit-proposal` that allows a JSON file containing a proposal to be passed in
  * [\#2040](https://github.com/cosmos/cosmos-sdk/issues/2040) Add `--bech` to `gaiacli keys show` and respective REST endpoint to
  provide desired Bech32 prefix encoding
  * [cli] [\#2047](https://github.com/cosmos/cosmos-sdk/issues/2047) [\#2306](https://github.com/cosmos/cosmos-sdk/pull/2306) Passing --gas=simulate triggers a simulation of the tx before the actual execution.
  The gas estimate obtained via the simulation will be used as gas limit in the actual execution.
  * [cli] [\#2047](https://github.com/cosmos/cosmos-sdk/issues/2047) The --gas-adjustment flag can be used to adjust the estimate obtained via the simulation triggered by --gas=simulate.
  * [cli] [\#2110](https://github.com/cosmos/cosmos-sdk/issues/2110) Add --dry-run flag to perform a simulation of a transaction without broadcasting it. The --gas flag is ignored as gas would be automatically estimated.
  * [cli] [\#2204](https://github.com/cosmos/cosmos-sdk/issues/2204) Support generating and broadcasting messages with multiple signatures via command line:
    * [\#966](https://github.com/cosmos/cosmos-sdk/issues/966) Add --generate-only flag to build an unsigned transaction and write it to STDOUT.
    * [\#1953](https://github.com/cosmos/cosmos-sdk/issues/1953) New `sign` command to sign transactions generated with the --generate-only flag.
    * [\#1954](https://github.com/cosmos/cosmos-sdk/issues/1954) New `broadcast` command to broadcast transactions generated offline and signed with the `sign` command.
  * [cli] [\#2220](https://github.com/cosmos/cosmos-sdk/issues/2220) Add `gaiacli config` feature to interactively create CLI config files to reduce the number of required flags
  * [stake][cli] [\#1672](https://github.com/cosmos/cosmos-sdk/issues/1672) Introduced
  new commission flags for validator commands `create-validator` and `edit-validator`.
  * [stake][cli] [\#1890](https://github.com/cosmos/cosmos-sdk/issues/1890) Add `--genesis-format` flag to `gaiacli tx create-validator` to produce transactions in genesis-friendly format.
  * [cli][\#2554](https://github.com/cosmos/cosmos-sdk/issues/2554) Make `gaiacli keys show` multisig ready.

* Gaia
  * [cli] [\#2170](https://github.com/cosmos/cosmos-sdk/issues/2170) added ability to show the node's address via `gaiad tendermint show-address`
  * [simulation] [\#2313](https://github.com/cosmos/cosmos-sdk/issues/2313) Reworked `make test_sim_gaia_slow` to `make test_sim_gaia_full`, now simulates from multiple starting seeds in parallel
  * [cli] [\#1921] (https://github.com/cosmos/cosmos-sdk/issues/1921)
    * New configuration file `gaiad.toml` is now created to host Gaia-specific configuration.
    * New --minimum_fees/minimum_fees flag/config option to set a minimum fee.

* SDK
  * [querier] added custom querier functionality, so ABCI query requests can be handled by keepers
  * [simulation] [\#1924](https://github.com/cosmos/cosmos-sdk/issues/1924) allow operations to specify future operations
  * [simulation] [\#1924](https://github.com/cosmos/cosmos-sdk/issues/1924) Add benchmarking capabilities, with makefile commands "test_sim_gaia_benchmark, test_sim_gaia_profile"
  * [simulation] [\#2349](https://github.com/cosmos/cosmos-sdk/issues/2349) Add time-based future scheduled operations to simulator
  * [x/auth] [\#2376](https://github.com/cosmos/cosmos-sdk/issues/2376) Remove FeePayer() from StdTx
  * [x/stake] [\#1672](https://github.com/cosmos/cosmos-sdk/issues/1672) Implement
  basis for the validator commission model.
  * [x/auth] Support account removal in the account mapper.


IMPROVEMENTS
* [tools] Improved terraform and ansible scripts for infrastructure deployment
* [tools] Added ansible script to enable process core dumps

* Gaia REST API (`gaiacli advanced rest-server`)
    * [x/stake] [\#2000](https://github.com/cosmos/cosmos-sdk/issues/2000) Added tests for new staking endpoints
    * [gaia-lite] [\#2445](https://github.com/cosmos/cosmos-sdk/issues/2445) Standarized REST error responses
    * [gaia-lite] Added example to Swagger specification for /keys/seed.
    * [x/stake] Refactor REST utils

* Gaia CLI  (`gaiacli`)
    * [cli] [\#2060](https://github.com/cosmos/cosmos-sdk/issues/2060) removed `--select` from `block` command
    * [cli] [\#2128](https://github.com/cosmos/cosmos-sdk/issues/2128) fixed segfault when exporting directly after `gaiad init`
    * [cli] [\#1255](https://github.com/cosmos/cosmos-sdk/issues/1255) open KeyBase in read-only mode
     for query-purpose CLI commands

* Gaia
    * [x/stake] [#2023](https://github.com/cosmos/cosmos-sdk/pull/2023) Terminate iteration loop in `UpdateBondedValidators` and `UpdateBondedValidatorsFull` when the first revoked validator is encountered and perform a sanity check.
    * [x/auth] Signature verification's gas cost now accounts for pubkey type. [#2046](https://github.com/tendermint/tendermint/pull/2046)
    * [x/stake] [x/slashing] Ensure delegation invariants to jailed validators [#1883](https://github.com/cosmos/cosmos-sdk/issues/1883).
    * [x/stake] Improve speed of GetValidator, which was shown to be a performance bottleneck. [#2046](https://github.com/tendermint/tendermint/pull/2200)
    * [x/stake] [\#2435](https://github.com/cosmos/cosmos-sdk/issues/2435) Improve memory efficiency of getting the various store keys
    * [genesis] [\#2229](https://github.com/cosmos/cosmos-sdk/issues/2229) Ensure that there are no duplicate accounts or validators in the genesis state.
    * [genesis] [\#2450](https://github.com/cosmos/cosmos-sdk/issues/2450) Validate staking genesis parameters.
    * Add SDK validation to `config.toml` (namely disabling `create_empty_blocks`) [\#1571](https://github.com/cosmos/cosmos-sdk/issues/1571)
    * [\#1941](https://github.com/cosmos/cosmos-sdk/issues/1941)(https://github.com/cosmos/cosmos-sdk/issues/1941) Version is now inferred via `git describe --tags`.
    * [x/distribution] [\#1671](https://github.com/cosmos/cosmos-sdk/issues/1671) add distribution types and tests

* SDK
    * [tools] Make get_vendor_deps deletes `.vendor-new` directories, in case scratch files are present.
    * [spec] Added simple piggy bank distribution spec
    * [cli] [\#1632](https://github.com/cosmos/cosmos-sdk/issues/1632) Add integration tests to ensure `basecoind init && basecoind` start sequences run successfully for both `democoin` and `basecoin` examples.
    * [store] Speedup IAVL iteration, and consequently everything that requires IAVL iteration. [#2143](https://github.com/cosmos/cosmos-sdk/issues/2143)
    * [store] [\#1952](https://github.com/cosmos/cosmos-sdk/issues/1952), [\#2281](https://github.com/cosmos/cosmos-sdk/issues/2281) Update IAVL dependency to v0.11.0
    * [simulation] Make timestamps randomized [#2153](https://github.com/cosmos/cosmos-sdk/pull/2153)
    * [simulation] Make logs not just pure strings, speeding it up by a large factor at greater block heights [\#2282](https://github.com/cosmos/cosmos-sdk/issues/2282)
    * [simulation] Add a concept of weighting the operations [\#2303](https://github.com/cosmos/cosmos-sdk/issues/2303)
    * [simulation] Logs get written to file if large, and also get printed on panics [\#2285](https://github.com/cosmos/cosmos-sdk/issues/2285)
    * [simulation] Bank simulations now makes testing auth configurable [\#2425](https://github.com/cosmos/cosmos-sdk/issues/2425)
    * [gaiad] [\#1992](https://github.com/cosmos/cosmos-sdk/issues/1992) Add optional flag to `gaiad testnet` to make config directory of daemon (default `gaiad`) and cli (default `gaiacli`) configurable
    * [x/stake] Add stake `Queriers` for Gaia-lite endpoints. This increases the staking endpoints performance by reusing the staking `keeper` logic for queries. [#2249](https://github.com/cosmos/cosmos-sdk/pull/2149)
    * [store] [\#2017](https://github.com/cosmos/cosmos-sdk/issues/2017) Refactor
    gas iterator gas consumption to only consume gas for iterator creation and `Next`
    calls which includes dynamic consumption of value length.
    * [types/decimal] [\#2378](https://github.com/cosmos/cosmos-sdk/issues/2378) - Added truncate functionality to decimal
    * [client] [\#1184](https://github.com/cosmos/cosmos-sdk/issues/1184) Remove unused `client/tx/sign.go`.
    * [tools] [\#2464](https://github.com/cosmos/cosmos-sdk/issues/2464) Lock binary dependencies to a specific version
    * #2573 [x/distribution] add accum invariance

BUG FIXES

* Gaia CLI  (`gaiacli`)
    * [cli] [\#1997](https://github.com/cosmos/cosmos-sdk/issues/1997) Handle panics gracefully when `gaiacli stake {delegation,unbond}` fail to unmarshal delegation.
    * [cli] [\#2265](https://github.com/cosmos/cosmos-sdk/issues/2265) Fix JSON formatting of the `gaiacli send` command.
    * [cli] [\#2547](https://github.com/cosmos/cosmos-sdk/issues/2547) Mark --to and --amount as required flags for `gaiacli tx send`.

* Gaia
  * [x/stake] Return correct Tendermint validator update set on `EndBlocker` by not
  including non previously bonded validators that have zero power. [#2189](https://github.com/cosmos/cosmos-sdk/issues/2189)

* SDK
    * [\#1988](https://github.com/cosmos/cosmos-sdk/issues/1988) Make us compile on OpenBSD (disable ledger) [#1988] (https://github.com/cosmos/cosmos-sdk/issues/1988)
    * [\#2105](https://github.com/cosmos/cosmos-sdk/issues/2105) Fix DB Iterator leak, which may leak a go routine.
    * [ledger] [\#2064](https://github.com/cosmos/cosmos-sdk/issues/2064) Fix inability to sign and send transactions via the LCD by
    loading a Ledger device at runtime.
    * [\#2158](https://github.com/cosmos/cosmos-sdk/issues/2158) Fix non-deterministic ordering of validator iteration when slashing in `gov EndBlocker`
    * [simulation] [\#1924](https://github.com/cosmos/cosmos-sdk/issues/1924) Make simulation stop on SIGTERM
    * [\#2388](https://github.com/cosmos/cosmos-sdk/issues/2388) Remove dependency on deprecated tendermint/tmlibs repository.
    * [\#2416](https://github.com/cosmos/cosmos-sdk/issues/2416) Refactored `InitializeTestLCD` to properly include proposing validator in genesis state.
    * #2573 [x/distribution] accum invariance bugfix
    * #2573 [x/slashing] unbonding-delegation slashing invariance bugfix

## 0.24.2

*August 22nd, 2018*

BUG FIXES

* Tendermint
  - Fix unbounded consensus WAL growth

## 0.24.1

*August 21st, 2018*

BUG FIXES

* Gaia
  - [x/slashing] Evidence tracking now uses validator address instead of validator pubkey

## 0.24.0

*August 13th, 2018*

BREAKING CHANGES

* Gaia REST API (`gaiacli advanced rest-server`)
  - [x/stake] [\#1880](https://github.com/cosmos/cosmos-sdk/issues/1880) More REST-ful endpoints (large refactor)
  - [x/slashing] [\#1866](https://github.com/cosmos/cosmos-sdk/issues/1866) `/slashing/signing_info` takes cosmosvalpub instead of cosmosvaladdr
  - use time.Time instead of int64 for time. See Tendermint v0.23.0
  - Signatures are no longer Amino encoded with prefixes (just encoded as raw
    bytes) - see Tendermint v0.23.0

* Gaia CLI  (`gaiacli`)
  -  [x/stake] change `--keybase-sig` to `--identity`
  -  [x/stake] [\#1828](https://github.com/cosmos/cosmos-sdk/issues/1828) Force user to specify amount on create-validator command by removing default
  -  [x/gov] Change `--proposalID` to `--proposal-id`
  -  [x/stake, x/gov] [\#1606](https://github.com/cosmos/cosmos-sdk/issues/1606) Use `--from` instead of adhoc flags like `--address-validator`
        and `--proposer` to indicate the sender address.
  -  [\#1551](https://github.com/cosmos/cosmos-sdk/issues/1551) Remove `--name` completely
  -  Genesis/key creation (`gaiad init`) now supports user-provided key passwords

* Gaia
  - [x/stake] Inflation doesn't use rationals in calculation (performance boost)
  - [x/stake] Persist a map from `addr->pubkey` in the state since BeginBlock
    doesn't provide pubkeys.
  - [x/gov] [\#1781](https://github.com/cosmos/cosmos-sdk/issues/1781) Added tags sub-package, changed tags to use dash-case
  - [x/gov] [\#1688](https://github.com/cosmos/cosmos-sdk/issues/1688) Governance parameters are now stored in globalparams store
  - [x/gov] [\#1859](https://github.com/cosmos/cosmos-sdk/issues/1859) Slash validators who do not vote on a proposal
  - [x/gov] [\#1914](https://github.com/cosmos/cosmos-sdk/issues/1914) added TallyResult type that gets stored in Proposal after tallying is finished

* SDK
  - [baseapp] Msgs are no longer run on CheckTx, removed `ctx.IsCheckTx()`
  - [baseapp] NewBaseApp constructor takes sdk.TxDecoder as argument instead of wire.Codec
  - [types] sdk.NewCoin takes sdk.Int, sdk.NewInt64Coin takes int64
  - [x/auth] Default TxDecoder can be found in `x/auth` rather than baseapp
  - [client] [\#1551](https://github.com/cosmos/cosmos-sdk/issues/1551): Refactored `CoreContext` to `TxContext` and `QueryContext`
      - Removed all tx related fields and logic (building & signing) to separate
        structure `TxContext` in `x/auth/client/context`

* Tendermint
    - v0.22.5 -> See [Tendermint PR](https://github.com/tendermint/tendermint/pull/1966)
        - change all the cryptography imports.
    - v0.23.0 -> See
      [Changelog](https://github.com/tendermint/tendermint/blob/v0.23.0/CHANGELOG.md#0230)
      and [SDK PR](https://github.com/cosmos/cosmos-sdk/pull/1927)
        - BeginBlock no longer includes crypto.Pubkey
        - use time.Time instead of int64 for time.

FEATURES

* Gaia REST API (`gaiacli advanced rest-server`)
    - [x/gov] Can now query governance proposals by ProposalStatus

* Gaia CLI  (`gaiacli`)
    - [x/gov] added `query-proposals` command. Can filter by `depositer`, `voter`, and `status`
    - [x/stake] [\#2043](https://github.com/cosmos/cosmos-sdk/issues/2043) Added staking query cli cmds for unbonding-delegations and redelegations

* Gaia
  - [networks] Added ansible scripts to upgrade seed nodes on a network

* SDK
  - [x/mock/simulation] Randomized simulation framework
     - Modules specify invariants and operations, preferably in an x/[module]/simulation package
     - Modules can test random combinations of their own operations
     - Applications can integrate operations and invariants from modules together for an integrated simulation
     - Simulates Tendermint's algorithm for validator set updates
     - Simulates validator signing/downtime with a Markov chain, and occaisional double-signatures
     - Includes simulated operations & invariants for staking, slashing, governance, and bank modules
  - [store] [\#1481](https://github.com/cosmos/cosmos-sdk/issues/1481) Add transient store
  - [baseapp] Initialize validator set on ResponseInitChain
  - [baseapp] added BaseApp.Seal - ability to seal baseapp parameters once they've been set
  - [cosmos-sdk-cli] New `cosmos-sdk-cli` tool to quickly initialize a new
    SDK-based project
  - [scripts] added log output monitoring to DataDog using Ansible scripts

IMPROVEMENTS

* Gaia
  - [spec] [\#967](https://github.com/cosmos/cosmos-sdk/issues/967) Inflation and distribution specs drastically improved
  - [x/gov] [\#1773](https://github.com/cosmos/cosmos-sdk/issues/1773) Votes on a proposal can now be queried
  - [x/gov] Initial governance parameters can now be set in the genesis file
  - [x/stake] [\#1815](https://github.com/cosmos/cosmos-sdk/issues/1815) Sped up the processing of `EditValidator` txs.
  - [config] [\#1930](https://github.com/cosmos/cosmos-sdk/issues/1930) Transactions indexer indexes all tags by default.
  - [ci] [#2057](https://github.com/cosmos/cosmos-sdk/pull/2057) Run `make localnet-start` on every commit and ensure network reaches at least 10 blocks

* SDK
  - [baseapp] [\#1587](https://github.com/cosmos/cosmos-sdk/issues/1587) Allow any alphanumeric character in route
  - [baseapp] Allow any alphanumeric character in route
  - [tools] Remove `rm -rf vendor/` from `make get_vendor_deps`
  - [x/auth] Recover ErrorOutOfGas panic in order to set sdk.Result attributes correctly
  - [x/auth] [\#2376](https://github.com/cosmos/cosmos-sdk/issues/2376) No longer runs any signature in a multi-msg, if any account/sequence number is wrong.
  - [x/auth] [\#2376](https://github.com/cosmos/cosmos-sdk/issues/2376) No longer charge gas for subtracting fees
  - [x/bank] Unit tests are now table-driven
  - [tests] Add tests to example apps in docs
  - [tests] Fixes ansible scripts to work with AWS too
  - [tests] [\#1806](https://github.com/cosmos/cosmos-sdk/issues/1806) CLI tests are now behind the build flag 'cli_test', so go test works on a new repo

BUG FIXES

* Gaia CLI  (`gaiacli`)
  -  [\#1766](https://github.com/cosmos/cosmos-sdk/issues/1766) Fixes bad example for keybase identity
  -  [x/stake] [\#2021](https://github.com/cosmos/cosmos-sdk/issues/2021) Fixed repeated CLI commands in staking

* Gaia
  - [x/stake] [#2077](https://github.com/cosmos/cosmos-sdk/pull/2077) Fixed invalid cliff power comparison
  - [\#1804](https://github.com/cosmos/cosmos-sdk/issues/1804) Fixes gen-tx genesis generation logic temporarily until upstream updates
  - [\#1799](https://github.com/cosmos/cosmos-sdk/issues/1799) Fix `gaiad export`
  - [\#1839](https://github.com/cosmos/cosmos-sdk/issues/1839) Fixed bug where intra-tx counter wasn't set correctly for genesis validators
  - [x/stake] [\#1858](https://github.com/cosmos/cosmos-sdk/issues/1858) Fixed bug where the cliff validator was not updated correctly
  - [tests] [\#1675](https://github.com/cosmos/cosmos-sdk/issues/1675) Fix non-deterministic `test_cover`
  - [tests] [\#1551](https://github.com/cosmos/cosmos-sdk/issues/1551) Fixed invalid LCD test JSON payload in `doIBCTransfer`
  - [basecoin] Fixes coin transaction failure and account query [discussion](https://forum.cosmos.network/t/unmarshalbinarybare-expected-to-read-prefix-bytes-75fbfab8-since-it-is-registered-concrete-but-got-0a141dfa/664/6)
  - [x/gov] [\#1757](https://github.com/cosmos/cosmos-sdk/issues/1757) Fix VoteOption conversion to String
  * [x/stake] [#2083] Fix broken invariant of bonded validator power decrease

## 0.23.1

*July 27th, 2018*

BUG FIXES
  * [tendermint] Update to v0.22.8
    - [consensus, blockchain] Register the Evidence interface so it can be
      marshalled/unmarshalled by the blockchain and consensus reactors

## 0.23.0

*July 25th, 2018*

BREAKING CHANGES
* [x/stake] Fixed the period check for the inflation calculation

IMPROVEMENTS
* [cli] Improve error messages for all txs when the account doesn't exist
* [tendermint] Update to v0.22.6
    - Updates the crypto imports/API (#1966)
* [x/stake] Add revoked to human-readable validator

BUG FIXES
* [tendermint] Update to v0.22.6
    - Fixes some security vulnerabilities reported in the [Bug Bounty](https://hackerone.com/tendermint)
*  [\#1797](https://github.com/cosmos/cosmos-sdk/issues/1797) Fix off-by-one error in slashing for downtime
*  [\#1787](https://github.com/cosmos/cosmos-sdk/issues/1787) Fixed bug where Tally fails due to revoked/unbonding validator
*  [\#1666](https://github.com/cosmos/cosmos-sdk/issues/1666) Add intra-tx counter to the genesis validators

## 0.22.0

*July 16th, 2018*

BREAKING CHANGES
* [x/gov] Increase VotingPeriod, DepositPeriod, and MinDeposit

IMPROVEMENTS
* [gaiad] Default config updates:
    - `timeout_commit=5000` so blocks only made every 5s
    - `prof_listen_addr=localhost:6060` so profile server is on by default
    - `p2p.send_rate` and `p2p.recv_rate` increases 10x (~5MB/s)

BUG FIXES
* [server] Fix to actually overwrite default tendermint config

## 0.21.1

*July 14th, 2018*

BUG FIXES
* [build] Added Ledger build support via `LEDGER_ENABLED=true|false`
  * True by default except when cross-compiling

## 0.21.0

*July 13th, 2018*

BREAKING CHANGES
* [x/stake] Specify DelegatorAddress in MsgCreateValidator
* [x/stake] Remove the use of global shares in the pool
   * Remove the use of `PoolShares` type in `x/stake/validator` type - replace with `Status` `Tokens` fields
* [x/auth] NewAccountMapper takes a constructor instead of a prototype
* [keys] Keybase.Update function now takes in a function to get the newpass, rather than the password itself

FEATURES
* [baseapp] NewBaseApp now takes option functions as parameters

IMPROVEMENTS
* Updated docs folder to accommodate cosmos.network docs project
* [store] Added support for tracing multi-store operations via `--trace-store`
* [store] Pruning strategy configurable with pruning flag on gaiad start

BUG FIXES
* [\#1630](https://github.com/cosmos/cosmos-sdk/issues/1630) - redelegation nolonger removes tokens from the delegator liquid account
* [keys] [\#1629](https://github.com/cosmos/cosmos-sdk/issues/1629) - updating password no longer asks for a new password when the first entered password was incorrect
* [lcd] importing an account would create a random account
* [server] 'gaiad init' command family now writes provided name as the moniker in `config.toml`
* [build] Added Ledger build support via `LEDGER_ENABLED=true|false`
  * True by default except when cross-compiling

## 0.20.0

*July 10th, 2018*

BREAKING CHANGES
* msg.GetSignBytes() returns sorted JSON (by key)
* msg.GetSignBytes() field changes
    * `msg_bytes` -> `msgs`
    * `fee_bytes` -> `fee`
* Update Tendermint to v0.22.2
    * Default ports changed from 466xx to 266xx
    * Amino JSON uses type names instead of prefix bytes
    * ED25519 addresses are the first 20-bytes of the SHA256 of the raw 32-byte
      pubkey (Instead of RIPEMD160)
    * go-crypto, abci, tmlibs have been merged into Tendermint
      * The keys sub-module is now in the SDK
    * Various other fixes
* [auth] Signers of a transaction now only sign over their own account and sequence number
* [auth] Removed MsgChangePubKey
* [auth] Removed SetPubKey from account mapper
* [auth] AltBytes renamed to Memo, now a string, max 100 characters, costs a bit of gas
* [types] `GetMsg()` -> `GetMsgs()` as txs wrap many messages
* [types] Removed GetMemo from Tx (it is still on StdTx)
* [types] renamed rational.Evaluate to rational.Round{Int64, Int}
* [types] Renamed `sdk.Address` to `sdk.AccAddress`/`sdk.ValAddress`
* [types] `sdk.AccAddress`/`sdk.ValAddress` natively marshals to Bech32 in String, Sprintf (when used with `%s`), and MarshalJSON
* [keys] Keybase and Ledger support from go-crypto merged into the SDK in the `crypto` folder
* [cli] Rearranged commands under subcommands
* [x/slashing] Update slashing for unbonding period
  * Slash according to power at time of infraction instead of power at
    time of discovery
  * Iterate through unbonding delegations & redelegations which contributed
    to an infraction, slash them proportional to their stake at the time
  * Add REST endpoint to unrevoke a validator previously revoked for downtime
  * Add REST endpoint to retrieve liveness signing information for a validator
* [x/stake] Remove Tick and add EndBlocker
* [x/stake] most index keys nolonger hold a value - inputs are rearranged to form the desired key
* [x/stake] store-value for delegation, validator, ubd, and red do not hold duplicate information contained store-key
* [x/stake] Introduce concept of unbonding for delegations and validators
  * `gaiacli stake unbond` replaced with `gaiacli stake begin-unbonding`
  * Introduced:
    * `gaiacli stake complete-unbonding`
    * `gaiacli stake begin-redelegation`
    * `gaiacli stake complete-redelegation`
* [lcd] Switch key creation output to return bech32
* [lcd] Removed shorthand CLI flags (`a`, `c`, `n`, `o`)
* [gaiad] genesis transactions now use bech32 addresses / pubkeys
* [gov] VoteStatus renamed to ProposalStatus
* [gov] VoteOption, ProposalType, and ProposalStatus all marshal to string form in JSON

DEPRECATED
* [cli] Deprecated `--name` flag in commands that send txs, in favor of `--from`

FEATURES
* [x/gov] Implemented MVP
  * Supported proposal types: just binary (pass/fail) TextProposals for now
  * Proposals need deposits to be votable; deposits are burned if proposal fails
  * Delegators delegate votes to validator by default but can override (for their stake)
* [gaiacli] Ledger support added
  - You can now use a Ledger with `gaiacli --ledger` for all key-related commands
  - Ledger keys can be named and tracked locally in the key DB
* [gaiacli] You can now attach a simple text-only memo to any transaction, with the `--memo` flag
* [gaiacli] added the following flags for commands that post transactions to the chain:
  * async -- send the tx without waiting for a tendermint response
  * json  -- return the output in json format for increased readability
  * print-response -- return the tx response. (includes fields like gas cost)
* [lcd] Queried TXs now include the tx hash to identify each tx
* [mockapp] CompleteSetup() no longer takes a testing parameter
* [x/bank] Add benchmarks for signing and delivering a block with a single bank transaction
  * Run with `cd x/bank && go test --bench=.`
* [tools] make get_tools installs tendermint's linter, and gometalinter
* [tools] Switch gometalinter to the stable version
* [tools] Add the following linters
  * misspell
  * gofmt
  * go vet -composites=false
  * unconvert
  * ineffassign
  * errcheck
  * unparam
  * gocyclo
* [tools] Added `make format` command to automate fixing misspell and gofmt errors.
* [server] Default config now creates a profiler at port 6060, and increase p2p send/recv rates
* [types] Switches internal representation of Int/Uint/Rat to use pointers
* [types] Added MinInt and MinUint functions
* [gaiad] `unsafe_reset_all` now resets addrbook.json
* [democoin] add x/oracle, x/assoc
* [tests] created a randomized testing framework.
  - Currently bank has limited functionality in the framework
  - Auth has its invariants checked within the framework
* [tests] Add WaitForNextNBlocksTM helper method
* [keys] New keys now have 24 word recovery keys, for heightened security
- [keys] Add a temporary method for exporting the private key

IMPROVEMENTS
* [x/bank] Now uses go-wire codec instead of 'encoding/json'
* [x/auth] Now uses go-wire codec instead of 'encoding/json'
* revised use of endblock and beginblock
* [stake] module reorganized to include `types` and `keeper` package
* [stake] keeper always loads the store (instead passing around which doesn't really boost efficiency)
* [stake] edit-validator changes now can use the keyword [do-not-modify] to not modify unspecified `--flag` (aka won't set them to `""` value)
* [stake] offload more generic functionality from the handler into the keeper
* [stake] clearer staking logic
* [types] added common tag constants
* [keys] improve error message when deleting non-existent key
* [gaiacli] improve error messages on `send` and `account` commands
* added contributing guidelines
* [docs] Added commands for governance CLI on testnet README

BUG FIXES
* [x/slashing] [\#1510](https://github.com/cosmos/cosmos-sdk/issues/1510) Unrevoked validators cannot un-revoke themselves
* [x/stake] [\#1513](https://github.com/cosmos/cosmos-sdk/issues/1513) Validators slashed to zero power are unbonded and removed from the store
* [x/stake] [\#1567](https://github.com/cosmos/cosmos-sdk/issues/1567) Validators decreased in power but not unbonded are now updated in Tendermint
* [x/stake] error strings lower case
* [x/stake] pool loose tokens now accounts for unbonding and unbonding tokens not associated with any validator
* [x/stake] fix revoke bytes ordering (was putting revoked candidates at the top of the list)
* [x/stake] bond count was counting revoked validators as bonded, fixed
* [gaia] Added self delegation for validators in the genesis creation
* [lcd] tests now don't depend on raw json text
* Retry on HTTP request failure in CLI tests, add option to retry tests in Makefile
* Fixed bug where chain ID wasn't passed properly in x/bank REST handler, removed Viper hack from ante handler
* Fixed bug where `democli account` didn't decode the account data correctly
* [\#872](https://github.com/cosmos/cosmos-sdk/issues/872)  - recovery phrases no longer all end in `abandon`
* [\#887](https://github.com/cosmos/cosmos-sdk/issues/887)  - limit the size of rationals that can be passed in from user input
* [\#1052](https://github.com/cosmos/cosmos-sdk/issues/1052) - Make all now works
* [\#1258](https://github.com/cosmos/cosmos-sdk/issues/1258) - printing big.rat's can no longer overflow int64
* [\#1259](https://github.com/cosmos/cosmos-sdk/issues/1259) - fix bug where certain tests that could have a nil pointer in defer
* [\#1343](https://github.com/cosmos/cosmos-sdk/issues/1343) - fixed unnecessary parallelism in CI
* [\#1353](https://github.com/cosmos/cosmos-sdk/issues/1353) - CLI: Show pool shares fractions in human-readable format
* [\#1367](https://github.com/cosmos/cosmos-sdk/issues/1367) - set ChainID in InitChain
* [\#1461](https://github.com/cosmos/cosmos-sdk/issues/1461) - CLI tests now no longer reset your local environment data
* [\#1505](https://github.com/cosmos/cosmos-sdk/issues/1505) - `gaiacli stake validator` no longer panics if validator doesn't exist
* [\#1565](https://github.com/cosmos/cosmos-sdk/issues/1565) - fix cliff validator persisting when validator set shrinks from max
* [\#1287](https://github.com/cosmos/cosmos-sdk/issues/1287) - prevent zero power validators at genesis
* [x/stake] fix bug when unbonding/redelegating using `--shares-percent`
* [\#1010](https://github.com/cosmos/cosmos-sdk/issues/1010) - two validators can't bond with the same pubkey anymore


## 0.19.0

*June 13, 2018*

BREAKING CHANGES
* msg.GetSignBytes() now returns bech32-encoded addresses in all cases
* [lcd] REST end-points now include gas
* sdk.Coin now uses sdk.Int, a big.Int wrapper with 256bit range cap

FEATURES
* [x/auth] Added AccountNumbers to BaseAccount and StdTxs to allow for replay protection with account pruning
* [lcd] added an endpoint to query for the SDK version of the connected node

IMPROVEMENTS
* export command now writes current validator set for Tendermint
* [tests] Application module tests now use a mock application
* [gaiacli] Fix error message when account isn't found when running gaiacli account
* [lcd] refactored to eliminate use of global variables, and interdependent tests
* [tests] Added testnet command to gaiad
* [tests] Added localnet targets to Makefile
* [x/stake] More stake tests added to test ByPower index

FIXES
* Fixes consensus fault on testnet - see postmortem [here](https://github.com/cosmos/cosmos-sdk/issues/1197#issuecomment-396823021)
* [x/stake] bonded inflation removed, non-bonded inflation partially implemented
* [lcd] Switch to bech32 for addresses on all human readable inputs and outputs
* [lcd] fixed tx indexing/querying
* [cli] Added `--gas` flag to specify transaction gas limit
* [gaia] Registered slashing message handler
* [x/slashing] Set signInfo.StartHeight correctly for newly bonded validators

FEATURES
* [docs] Reorganize documentation
* [docs] Update staking spec, create WIP spec for slashing, and fees

## 0.18.0

*June 9, 2018*

BREAKING CHANGES

* [stake] candidate -> validator throughout (details in refactor comment)
* [stake] delegate-bond -> delegation throughout
* [stake] `gaiacli query validator` takes and argument instead of using the `--address-candidate` flag
* [stake] introduce `gaiacli query delegations`
* [stake] staking refactor
  * ValidatorsBonded store now take sorted pubKey-address instead of validator owner-address,
    is sorted like Tendermint by pk's address
  * store names more understandable
  * removed temporary ToKick store, just needs a local map!
  * removed distinction between candidates and validators
    * everything is now a validator
    * only validators with a status == bonded are actively validating/receiving rewards
  * Introduction of Unbonding fields, lowlevel logic throughout (not fully implemented with queue)
  * Introduction of PoolShares type within validators,
    replaces three rational fields (BondedShares, UnbondingShares, UnbondedShares
* [x/auth] move stuff specific to auth anteHandler to the auth module rather than the types folder. This includes:
  * StdTx (and its related stuff i.e. StdSignDoc, etc)
  * StdFee
  * StdSignature
  * Account interface
  * Related to this organization, I also:
* [x/auth] got rid of AccountMapper interface (in favor of the struct already in auth module)
* [x/auth] removed the FeeHandler function from the AnteHandler, Replaced with FeeKeeper
* [x/auth] Removed GetSignatures() from Tx interface (as different Tx styles might use something different than StdSignature)
* [store] Removed SubspaceIterator and ReverseSubspaceIterator from KVStore interface and replaced them with helper functions in /types
* [cli] rearranged commands under subcommands
* [stake] remove Tick and add EndBlocker
* Switch to bech32cosmos on all human readable inputs and outputs


FEATURES

* [x/auth] Added ability to change pubkey to auth module
* [baseapp] baseapp now has settable functions for filtering peers by address/port & public key
* [sdk] Gas consumption is now measured as transactions are executed
  * Transactions which run out of gas stop execution and revert state changes
  * A "simulate" query has been added to determine how much gas a transaction will need
  * Modules can include their own gas costs for execution of particular message types
* [stake] Seperation of fee distribution to a new module
* [stake] Creation of a validator/delegation generics in `/types`
* [stake] Helper Description of the store in x/stake/store.md
* [stake] removed use of caches in the stake keeper
* [stake] Added REST API
* [Makefile] Added terraform/ansible playbooks to easily create remote testnets on Digital Ocean


BUG FIXES

* [stake] staking delegator shares exchange rate now relative to equivalent-bonded-tokens the validator has instead of bonded tokens
  ^ this is important for unbonded validators in the power store!
* [cli] fixed cli-bash tests
* [ci] added cli-bash tests
* [basecoin] updated basecoin for stake and slashing
* [docs] fixed references to old cli commands
* [docs] Downgraded Swagger to v2 for downstream compatibility
* auto-sequencing transactions correctly
* query sequence via account store
* fixed duplicate pub_key in stake.Validator
* Auto-sequencing now works correctly
* [gaiacli] Fix error message when account isn't found when running gaiacli account


## 0.17.5

*June 5, 2018*

Update to Tendermint v0.19.9 (Fix evidence reactor, mempool deadlock, WAL panic,
memory leak)

## 0.17.4

*May 31, 2018*

Update to Tendermint v0.19.7 (WAL fixes and more)

## 0.17.3

*May 29, 2018*

Update to Tendermint v0.19.6 (fix fast-sync halt)

## 0.17.5

*June 5, 2018*

Update to Tendermint v0.19.9 (Fix evidence reactor, mempool deadlock, WAL panic,
memory leak)

## 0.17.4

*May 31, 2018*

Update to Tendermint v0.19.7 (WAL fixes and more)

## 0.17.3

*May 29, 2018*

Update to Tendermint v0.19.6 (fix fast-sync halt)

## 0.17.2

_May 20, 2018_

Update to Tendermint v0.19.5 (reduce WAL use, bound the mempool and some rpcs, improve logging)

## 0.17.1 (May 17, 2018)

Update to Tendermint v0.19.4 (fixes a consensus bug and improves logging)

## 0.17.0 (May 15, 2018)

BREAKING CHANGES

* [stake] MarshalJSON -> MarshalBinaryLengthPrefixed
* Queries against the store must be prefixed with the path "/store"

FEATURES

* [gaiacli] Support queries for candidates, delegator-bonds
* [gaiad] Added `gaiad export` command to export current state to JSON
* [x/bank] Tx tags with sender/recipient for indexing & later retrieval
* [x/stake] Tx tags with delegator/candidate for delegation & unbonding, and candidate info for declare candidate / edit validator

IMPROVEMENTS

* [gaiad] Update for Tendermint v0.19.3 (improve `/dump_consensus_state` and add
  `/consensus_state`)
* [spec/ibc] Added spec!
* [spec/stake] Cleanup structure, include details about slashing and
  auto-unbonding
* [spec/governance] Fixup some names and pseudocode
* NOTE: specs are still a work-in-progress ...

BUG FIXES

* Auto-sequencing now works correctly


## 0.16.0 (May 14th, 2018)

BREAKING CHANGES

* Move module REST/CLI packages to x/[module]/client/rest and x/[module]/client/cli
* Gaia simple-staking bond and unbond functions replaced
* [stake] Delegator bonds now store the height at which they were updated
* All module keepers now require a codespace, see basecoin or democoin for usage
* Many changes to names throughout
  * Type as a prefix naming convention applied (ex. BondMsg -> MsgBond)
  * Removed redundancy in names (ex. stake.StakeKeeper -> stake.Keeper)
* Removed SealedAccountMapper
* gaiad init now requires use of `--name` flag
* Removed Get from Msg interface
* types/rational now extends big.Rat

FEATURES:

* Gaia stake commands include, CreateValidator, EditValidator, Delegate, Unbond
* MountStoreWithDB without providing a custom store works.
* Repo is now lint compliant / GoMetaLinter with tendermint-lint integrated into CI
* Better key output, pubkey go-amino hex bytes now output by default
* gaiad init overhaul
  * Create genesis transactions with `gaiad init gen-tx`
  * New genesis account keys are automatically added to the client keybase (introduce `--client-home` flag)
  * Initialize with genesis txs using `--gen-txs` flag
* Context now has access to the application-configured logger
* Add (non-proof) subspace query helper functions
* Add more staking query functions: candidates, delegator-bonds

BUG FIXES

* Gaia now uses stake, ported from github.com/cosmos/gaia


## 0.15.1 (April 29, 2018)

IMPROVEMENTS:

* Update Tendermint to v0.19.1 (includes many rpc fixes)


## 0.15.0 (April 29, 2018)

NOTE: v0.15.0 is a large breaking change that updates the encoding scheme to use
[Amino](github.com/tendermint/go-amino).

For details on how this changes encoding for public keys and addresses,
see the [docs](https://github.com/tendermint/tendermint/blob/v0.19.1/docs/specification/new-spec/encoding.md#public-key-cryptography).

BREAKING CHANGES

* Remove go-wire, use go-amino
* [store] Add `SubspaceIterator` and `ReverseSubspaceIterator` to `KVStore` interface
* [basecoin] NewBasecoinApp takes a `dbm.DB` and uses namespaced DBs for substores

FEATURES:

* Add CacheContext
* Add auto sequencing to client
* Add FeeHandler to ante handler

BUG FIXES

* MountStoreWithDB without providing a custom store works.

## 0.14.1 (April 9, 2018)

BUG FIXES

* [gaiacli] Fix all commands (just a duplicate of basecli for now)

## 0.14.0 (April 9, 2018)

BREAKING CHANGES:

* [client/builder] Renamed to `client/core` and refactored to use a CoreContext
  struct
* [server] Refactor to improve useability and de-duplicate code
* [types] `Result.ToQuery -> Error.QueryResult`
* [makefile] `make build` and `make install` only build/install `gaiacli` and
  `gaiad`. Use `make build_examples` and `make install_examples` for
  `basecoind/basecli` and `democoind/democli`
* [staking] Various fixes/improvements

FEATURES:

* [democoin] Added Proof-of-Work module

BUG FIXES

* [client] Reuse Tendermint RPC client to avoid excessive open files
* [client] Fix setting log level
* [basecoin] Sort coins in genesis

## 0.13.1 (April 3, 2018)

BUG FIXES

* [x/ibc] Fix CLI and relay for IBC txs
* [x/stake] Various fixes/improvements

## 0.13.0 (April 2, 2018)

BREAKING CHANGES

* [basecoin] Remove cool/sketchy modules -> moved to new `democoin`
* [basecoin] NewBasecoinApp takes a `map[string]dbm.DB` as temporary measure
  to allow mounting multiple stores with their own DB until they can share one
* [x/staking] Renamed to `simplestake`
* [builder] Functions don't take `passphrase` as argument
* [server] GenAppParams returns generated seed and address
* [basecoind] `init` command outputs JSON of everything necessary for testnet
* [basecoind] `basecoin.db -> data/basecoin.db`
* [basecli] `data/keys.db -> keys/keys.db`

FEATURES

* [types] `Coin` supports direct arithmetic operations
* [basecoind] Add `show_validator` and `show_node_id` commands
* [x/stake] Initial merge of full staking module!
* [democoin] New example application to demo custom modules

IMPROVEMENTS

* [makefile] `make install`
* [testing] Use `/tmp` for directories so they don't get left in the repo

BUG FIXES

* [basecoin] Allow app to be restarted
* [makefile] Fix build on Windows
* [basecli] Get confirmation before overriding key with same name

## 0.12.0 (March 27 2018)

BREAKING CHANGES

* Revert to old go-wire for now
* glide -> godep
* [types] ErrBadNonce -> ErrInvalidSequence
* [types] Replace tx.GetFeePayer with FeePayer(tx) - returns the first signer
* [types] NewStdTx takes the Fee
* [types] ParseAccount -> AccountDecoder; ErrTxParse -> ErrTxDecoder
* [x/auth] AnteHandler deducts fees
* [x/bank] Move some errors to `types`
* [x/bank] Remove sequence and signature from Input

FEATURES

* [examples/basecoin] New cool module to demonstrate use of state and custom transactions
* [basecoind] `show_node_id` command
* [lcd] Implement the Light Client Daemon and endpoints
* [types/stdlib] Queue functionality
* [store] Subspace iterator on IAVLTree
* [types] StdSignDoc is the document that gets signed (chainid, msg, sequence, fee)
* [types] CodeInvalidPubKey
* [types] StdFee, and StdTx takes the StdFee
* [specs] Progression of MVPs for IBC
* [x/ibc] Initial shell of IBC functionality (no proofs)
* [x/simplestake] Simple staking module with bonding/unbonding

IMPROVEMENTS

* Lots more tests!
* [client/builder] Helpers for forming and signing transactions
* [types] sdk.Address
* [specs] Staking

BUG FIXES

* [x/auth] Fix setting pubkey on new account
* [x/auth] Require signatures to include the sequences
* [baseapp] Dont panic on nil handler
* [basecoin] Check for empty bytes in account and tx

## 0.11.0 (March 1, 2017)

BREAKING CHANGES

* [examples] dummy -> kvstore
* [examples] Remove gaia
* [examples/basecoin] MakeTxCodec -> MakeCodec
* [types] CommitMultiStore interface has new `GetCommitKVStore(key StoreKey) CommitKVStore` method

FEATURES

* [examples/basecoin] CLI for `basecli` and `basecoind` (!)
* [baseapp] router.AddRoute returns Router

IMPROVEMENTS

* [baseapp] Run msg handlers on CheckTx
* [docs] Add spec for REST API
* [all] More tests!

BUG FIXES

* [baseapp] Fix panic on app restart
* [baseapp] InitChain does not call Commit
* [basecoin] Remove IBCStore because mounting multiple stores is currently broken

## 0.10.0 (February 20, 2017)

BREAKING CHANGES

* [baseapp] NewBaseApp(logger, db)
* [baseapp] NewContext(isCheckTx, header)
* [x/bank] CoinMapper -> CoinKeeper

FEATURES

* [examples/gaia] Mock CLI !
* [baseapp] InitChainer, BeginBlocker, EndBlocker
* [baseapp] MountStoresIAVL

IMPROVEMENTS

* [docs] Various improvements.
* [basecoin] Much simpler :)

BUG FIXES

* [baseapp] initialize and reset msCheck and msDeliver properly

## 0.9.0 (February 13, 2017)

BREAKING CHANGES

* Massive refactor. Basecoin works. Still needs <3

## 0.8.1

* Updates for dependencies

## 0.8.0 (December 18, 2017)

* Updates for dependencies

## 0.7.1 (October 11, 2017)

IMPROVEMENTS:

* server/commands: GetInitCmd takes list of options

## 0.7.0 (October 11, 2017)

BREAKING CHANGES:

* Everything has changed, and it's all about to change again, so don't bother using it yet!

## 0.6.2 (July 27, 2017)

IMPROVEMENTS:

* auto-test all tutorials to detect breaking changes
* move deployment scripts from `/scripts` to `/publish` for clarity

BUG FIXES:

* `basecoin init` ensures the address in genesis.json is valid
* fix bug that certain addresses couldn't receive ibc packets

## 0.6.1 (June 28, 2017)

Make lots of small cli fixes that arose when people were using the tools for
the testnet.

IMPROVEMENTS:

* basecoin
  * `basecoin start` supports all flags that `tendermint node` does, such as
    `--rpc.laddr`, `--p2p.seeds`, and `--p2p.skip_upnp`
  * fully supports `--log_level` and `--trace` for logger configuration
  * merkleeyes no longers spams the logs... unless you want it
    * Example: `basecoin start --log_level="merkleeyes:info,state:info,*:error"`
    * Example: `basecoin start --log_level="merkleeyes:debug,state:info,*:error"`
* basecli
  * `basecli init` is more intelligent and only complains if there really was
    a connected chain, not just random files
  * support `localhost:46657` or `http://localhost:46657` format for nodes,
    not just `tcp://localhost:46657`
  * Add `--genesis` to init to specify chain-id and validator hash
    * Example: `basecli init --node=localhost:46657 --genesis=$HOME/.basecoin/genesis.json`
  * `basecli rpc` has a number of methods to easily accept tendermint rpc, and verifies what it can

BUG FIXES:

* basecli
  * `basecli query account` accepts hex account address with or without `0x`
    prefix
  * gives error message when running commands on an unitialized chain, rather
    than some unintelligable panic

## 0.6.0 (June 22, 2017)

Make the basecli command the only way to use client-side, to enforce best
security practices. Lots of enhancements to get it up to production quality.

BREAKING CHANGES:

* ./cmd/commands -> ./cmd/basecoin/commands
* basecli
  * `basecli proof state get` -> `basecli query key`
  * `basecli proof tx get` -> `basecli query tx`
  * `basecli proof state get --app=account` -> `basecli query account`
  * use `--chain-id` not `--chainid` for consistency
  * update to use `--trace` not `--debug` for stack traces on errors
  * complete overhaul on how tx and query subcommands are added. (see counter or trackomatron for examples)
  * no longer supports counter app (see new countercli)
* basecoin
  * `basecoin init` takes an argument, an address to allocate funds to in the genesis
  * removed key2.json
  * removed all client side functionality from it (use basecli now for proofs)
    * no tx subcommand
    * no query subcommand
    * no account (query) subcommand
    * a few other random ones...
  * enhanced relay subcommand
    * relay start did what relay used to do
    * relay init registers both chains on one another (to set it up so relay start just works)
* docs
  * removed `example-plugin`, put `counter` inside `docs/guide`
* app
  * Implements ABCI handshake by proxying merkleeyes.Info()

IMPROVEMENTS:

* `basecoin init` support `--chain-id`
* intergrates tendermint 0.10.0 (not the rc-2, but the real thing)
* commands return error code (1) on failure for easier script testing
* add `reset_all` to basecli, and never delete keys on `init`
* new shutil based unit tests, with better coverage of the cli actions
* just `make fresh` when things are getting stale ;)

BUG FIXES:

* app: no longer panics on missing app_options in genesis (thanks, anton)
* docs: updated all docs... again
* ibc: fix panic on getting BlockID from commit without 100% precommits (still a TODO)

## 0.5.2 (June 2, 2017)

BUG FIXES:

* fix parsing of the log level from Tendermint config (#97)

## 0.5.1 (May 30, 2017)

BUG FIXES:

* fix ibc demo app to use proper tendermint flags, 0.10.0-rc2 compatibility
* Make sure all cli uses new json.Marshal not wire.JSONBytes

## 0.5.0 (May 27, 2017)

BREAKING CHANGES:

* only those related to the tendermint 0.9 -> 0.10 upgrade

IMPROVEMENTS:

* basecoin cli
  * integrates tendermint 0.10.0 and unifies cli (init, unsafe_reset_all, ...)
  * integrate viper, all command line flags can also be defined in environmental variables or config.toml
* genesis file
  * you can define accounts with either address or pub_key
  * sorts coins for you, so no silent errors if not in alphabetical order
* [light-client](https://github.com/tendermint/light-client) integration
  * no longer must you trust the node you connect to, prove everything!
  * new [basecli command](./cmd/basecli/README.md)
  * integrated [key management](https://github.com/tendermint/go-crypto/blob/master/cmd/README.md), stored encrypted locally
  * tracks validator set changes and proves everything from one initial validator seed
  * `basecli proof state` gets complete proofs for any abci state
  * `basecli proof tx` gets complete proof where a tx was stored in the chain
  * `basecli proxy` exposes tendermint rpc, but only passes through results after doing complete verification

BUG FIXES:

* no more silently ignored error with invalid coin names (eg. "17.22foo coin" used to parse as "17 foo", not warning/error)

## 0.4.1 (April 26, 2017)

BUG FIXES:

* Fix bug in `basecoin unsafe_reset_X` where the `priv_validator.json` was not being reset

## 0.4.0 (April 21, 2017)

BREAKING CHANGES:

* CLI now uses Cobra, which forced changes to some of the flag names and orderings

IMPROVEMENTS:

* `basecoin init` doesn't generate error if already initialized
* Much more testing

## 0.3.1 (March 23, 2017)

IMPROVEMENTS:

* CLI returns exit code 1 and logs error before exiting

## 0.3.0 (March 23, 2017)

BREAKING CHANGES:

* Remove `--data` flag and use `BCHOME` to set the home directory (defaults to `~/.basecoin`)
* Remove `--in-proc` flag and start Tendermint in-process by default (expect Tendermint files in $BCHOME/tendermint).
  To start just the ABCI app/server, use `basecoin start --without-tendermint`.
* Consolidate genesis files so the Basecoin genesis is an object under `app_options` in Tendermint genesis. For instance:

```
{
  "app_hash": "",
  "chain_id": "foo_bar_chain",
  "genesis_time": "0001-01-01T00:00:00.000Z",
  "validators": [
    {
      "amount": 10,
      "name": "",
      "pub_key": [
	1,
	"7B90EA87E7DC0C7145C8C48C08992BE271C7234134343E8A8E8008E617DE7B30"
      ]
    }
  ],
  "app_options": {
    "accounts": [{
      "pub_key": {
        "type": "ed25519",
        "data": "6880db93598e283a67c4d88fc67a8858aa2de70f713fe94a5109e29c137100c2"
      },
      "coins": [
        {
          "denom": "blank",
          "amount": 12345
        },
        {
          "denom": "ETH",
          "amount": 654321
        }
      ]
    }],
    "plugin_options": ["plugin1/key1", "value1", "plugin1/key2", "value2"]
  }
}
```

Note the array of key-value pairs is now under `app_options.plugin_options` while the `app_options` themselves are well formed.
We also changed `chainID` to `chain_id` and consolidated to have just one of them.

FEATURES:

* Introduce `basecoin init` and `basecoin unsafe_reset_all`

## 0.2.0 (March 6, 2017)

BREAKING CHANGES:

* Update to ABCI v0.4.0 and Tendermint v0.9.0
* Coins are specified on the CLI as `Xcoin`, eg. `5gold`
* `Cost` is now `Fee`

FEATURES:

* CLI for sending transactions and querying the state,
  designed to be easily extensible as plugins are implemented
* Run Basecoin in-process with Tendermint
* Add `/account` path in Query
* IBC plugin for InterBlockchain Communication
* Demo script of IBC between two chains

IMPROVEMENTS:

* Use new Tendermint `/commit` endpoint for crafting IBC transactions
* More unit tests
* Use go-crypto S structs and go-data for more standard JSON
* Demo uses fewer sleeps

BUG FIXES:

* Various little fixes in coin arithmetic
* More commit validation in IBC
* Return results from transactions

## PreHistory

##### January 14-18, 2017

* Update to Tendermint v0.8.0
* Cleanup a bit and release blog post

##### September 22, 2016

* Basecoin compiles again
