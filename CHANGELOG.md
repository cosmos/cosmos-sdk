<!--
Guiding Principles:

Changelogs are for humans, not machines.
There should be an entry for every single version.
The same types of changes should be grouped.
Versions and sections should be linkable.
The latest version comes first.
The release date of each version is displayed.
Mention whether you follow Semantic Versioning.

Usage:

Change log entries are to be added to the Unreleased section under the
appropriate stanza (see below). Each entry should ideally include a tag and
the Github issue reference in the following format:

* (<tag>) \#<issue-number> message

The issue numbers will later be link-ified during the release process so you do
not have to worry about including a link manually, but you can if you wish.

Types of changes (Stanzas):

"Features" for new features.
"Improvements" for changes in existing functionality.
"Deprecated" for soon-to-be removed features.
"Bug Fixes" for any bug fixes.
"Breaking" for breaking API changes.

Ref: https://keepachangelog.com/en/1.0.0/
-->

# Changelog

## [Unreleased]

## [v0.37.8] - 2020-03-11

### Bug Fixes

* (rest) [\#5508](https://github.com/cosmos/cosmos-sdk/pull/5508) Fix `x/distribution` endpoints to properly return height in the response.
* (x/genutil) [\#5499](https://github.com/cosmos/cosmos-sdk/pull/) Ensure `DefaultGenesis` returns valid and non-nil default genesis state.
* (x/genutil) [\#5775](https://github.com/cosmos/cosmos-sdk/pull/5775) Fix `ExportGenesis` in `x/genutil` to export default genesis state (`[]`) instead of `null`.
* (genesis) [\#5086](https://github.com/cosmos/cosmos-sdk/issues/5086) Ensure `gentxs` are always an empty array instead of `nil`.

### Improvements

* (rest) [\#5648](https://github.com/cosmos/cosmos-sdk/pull/5648) Enhance /txs usability:
  * Add `tx.minheight` key to filter transaction with an inclusive minimum block height
  * Add `tx.maxheight` key to filter transaction with an inclusive maximum block height

## [v0.37.7] - 2020-02-10

### Improvements

* (modules) [\#5597](https://github.com/cosmos/cosmos-sdk/pull/5597) Add `amount` event attribute to the `complete_unbonding`
and `complete_redelegation` events that reflect the total balances of the completed unbondings and redelegations
respectively.

### Bug Fixes

* (x/gov) [\#5622](https://github.com/cosmos/cosmos-sdk/pull/5622) Track any events emitted from a proposal's handler upon successful execution.
* (x/bank) [\#5531](https://github.com/cosmos/cosmos-sdk/issues/5531) Added missing amount event to MsgMultiSend, emitted for each output.

## [v0.37.6] - 2020-01-21

### Improvements

* (tendermint) Bump Tendermint version to [v0.32.9](https://github.com/tendermint/tendermint/releases/tag/v0.32.9)

## [v0.37.5] - 2020-01-07

### Features

* (types) [\#5360](https://github.com/cosmos/cosmos-sdk/pull/5360) Implement `SortableDecBytes` which
  allows the `Dec` type be sortable.

### Improvements

* (tendermint) Bump Tendermint version to [v0.32.8](https://github.com/tendermint/tendermint/releases/tag/v0.32.8)
* (cli) [\#5482](https://github.com/cosmos/cosmos-sdk/pull/5482) Remove old "tags" nomenclature from the `q txs` command in
  favor of the new events system. Functionality remains unchanged except that `=` is used instead of `:` to be
  consistent with the API's use of event queries.

### Bug Fixes

* (iavl) [\#5276](https://github.com/cosmos/cosmos-sdk/issues/5276) Fix potential race condition in `iavlIterator#Close`.
* (baseapp) [\#5350](https://github.com/cosmos/cosmos-sdk/issues/5350) Allow a node to restart successfully
  after a `halt-height` or `halt-time` has been triggered.
* (types) [\#5395](https://github.com/cosmos/cosmos-sdk/issues/5395) Fix `Uint#LTE`.
* (types) [\#5408](https://github.com/cosmos/cosmos-sdk/issues/5408) `NewDecCoins` constructor now sorts the coins.

## [v0.37.4] - 2019-11-04

### Improvements

* (tendermint) Bump Tendermint version to [v0.32.7](https://github.com/tendermint/tendermint/releases/tag/v0.32.7)
* (ledger) [\#4716](https://github.com/cosmos/cosmos-sdk/pull/4716) Fix ledger custom coin type support bug.

### Bug Fixes

* (baseapp) [\#5200](https://github.com/cosmos/cosmos-sdk/issues/5200) Remove duplicate events from previous messages.

## [v0.37.3] - 2019-10-10

### Bug Fixes

* (genesis) [\#5095](https://github.com/cosmos/cosmos-sdk/issues/5095) Fix genesis file migration from v0.34 to
v0.36/v0.37 not converting validator consensus pubkey to bech32 format.

### Improvements

* (tendermint) Bump Tendermint version to [v0.32.6](https://github.com/tendermint/tendermint/releases/tag/v0.32.6)

## [v0.37.1] - 2019-09-19

### Features

* (cli) [\#4973](https://github.com/cosmos/cosmos-sdk/pull/4973) Enable application CPU profiling
via the `--cpu-profile` flag.
* [\#4979](https://github.com/cosmos/cosmos-sdk/issues/4979) Introduce a new `halt-time` config and
CLI option to the `start` command. When provided, an application will halt during `Commit` when the
block time is >= the `halt-time`.

### Improvements

* [\#4990](https://github.com/cosmos/cosmos-sdk/issues/4990) Add `Events` to the `ABCIMessageLog` to
provide context and grouping of events based on the messages they correspond to. The `Events` field
in `TxResponse` is deprecated and will be removed in the next major release.

### Bug Fixes

* [\#4979](https://github.com/cosmos/cosmos-sdk/issues/4979) Use `Signal(os.Interrupt)` over
`os.Exit(0)` during configured halting to allow any `defer` calls to be executed.

## [v0.37.0] - 2019-08-21

### Bug Fixes

* (baseapp) [\#4903](https://github.com/cosmos/cosmos-sdk/issues/4903) Various height query fixes:
  * Move height with proof check from `CLIContext` to `BaseApp` as the height
  can automatically be injected there.
  * Update `handleQueryStore` to resemble `handleQueryCustom`
* (simulation) [\#4912](https://github.com/cosmos/cosmos-sdk/issues/4912) Fix SimApp ModuleAccountAddrs
to properly return black listed addresses for bank keeper initialization.
* (cli) [\#4919](https://github.com/cosmos/cosmos-sdk/pull/4919) Don't crash CLI
if user doesn't answer y/n confirmation request.
* (cli) [\#4927](https://github.com/cosmos/cosmos-sdk/issues/4927) Fix the `q gov vote`
command to handle empty (pruned) votes correctly.

### Improvements

* (rest) [\#4924](https://github.com/cosmos/cosmos-sdk/pull/4924) Return response
height even upon error as it may be useful for the downstream caller and have
`/auth/accounts/{address}` return a 200 with an empty account upon error when
that error is that the account doesn't exist.

## [v0.36.0] - 2019-08-13

### Breaking Changes

* (rest) [\#4837](https://github.com/cosmos/cosmos-sdk/pull/4837) Remove /version and /node_version
  endpoints in favor of refactoring /node_info to also include application version info.
* All REST responses now wrap the original resource/result. The response
  will contain two fields: height and result.
* [\#3565](https://github.com/cosmos/cosmos-sdk/issues/3565) Updates to the governance module:
  * Rename JSON field from `proposal_content` to `content`
  * Rename JSON field from `proposal_id` to `id`
  * Disable `ProposalTypeSoftwareUpgrade` temporarily
* [\#3775](https://github.com/cosmos/cosmos-sdk/issues/3775) unify sender transaction tag for ease of querying
* [\#4255](https://github.com/cosmos/cosmos-sdk/issues/4255) Add supply module that passively tracks the supplies of a chain
  - Renamed `x/distribution` `ModuleName`
  - Genesis JSON and CLI now use `distribution` instead of `distr`
  - Introduce `ModuleAccount` type, which tracks the flow of coins held within a module
  - Replaced `FeeCollectorKeeper` for a `ModuleAccount`
  - Replaced the staking `Pool`, which coins are now held by the `BondedPool` and `NotBonded` module accounts
  - The `NotBonded` module account now only keeps track of the not bonded tokens within staking, instead of the whole chain
  - [\#3628](https://github.com/cosmos/cosmos-sdk/issues/3628) Replaced governance's burn and deposit accounts for a `ModuleAccount`
  - Added a `ModuleAccount` for the distribution module
  - Added a `ModuleAccount` for the mint module
  [\#4472](https://github.com/cosmos/cosmos-sdk/issues/4472) validation for crisis genesis
* [\#3985](https://github.com/cosmos/cosmos-sdk/issues/3985) `ValidatorPowerRank` uses potential consensus power instead of tendermint power
* [\#4104](https://github.com/cosmos/cosmos-sdk/issues/4104) Gaia has been moved to its own repository: https://github.com/cosmos/gaia
* [\#4104](https://github.com/cosmos/cosmos-sdk/issues/4104) Rename gaiad.toml to app.toml. The internal contents of the application
  config remain unchanged.
* [\#4159](https://github.com/cosmos/cosmos-sdk/issues/4159) create the default module patterns and module manager
* [\#4230](https://github.com/cosmos/cosmos-sdk/issues/4230) Change the type of ABCIMessageLog#MsgIndex to uint16 for proper serialization.
* [\#4250](https://github.com/cosmos/cosmos-sdk/issues/4250) BaseApp.Query() returns app's version string set via BaseApp.SetAppVersion()
  when handling /app/version queries instead of the version string passed as build
  flag at compile time.
* [\#4262](https://github.com/cosmos/cosmos-sdk/issues/4262) GoSumHash is no longer returned by the version command.
* [\#4263](https://github.com/cosmos/cosmos-sdk/issues/4263) RestServer#Start now takes read and write timeout arguments.
* [\#4305](https://github.com/cosmos/cosmos-sdk/issues/4305) `GenerateOrBroadcastMsgs` no longer takes an `offline` parameter.
* [\#4342](https://github.com/cosmos/cosmos-sdk/pull/4342) Upgrade go-amino to v0.15.0
* [\#4351](https://github.com/cosmos/cosmos-sdk/issues/4351) InitCmd, AddGenesisAccountCmd, and CollectGenTxsCmd take node's and client's default home directories as arguments.
* [\#4387](https://github.com/cosmos/cosmos-sdk/issues/4387) Refactor the usage of tags (now called events) to reflect the
  new ABCI events semantics:
  - Move `x/{module}/tags/tags.go` => `x/{module}/types/events.go`
  - Update `docs/specs`
  - Refactor tags in favor of new `Event(s)` type(s)
  - Update `Context` to use new `EventManager`
  - (Begin|End)Blocker no longer return tags, but rather uses new `EventManager`
  - Message handlers no longer return tags, but rather uses new `EventManager`
  Any component (e.g. BeginBlocker, message handler, etc...) wishing to emit an event must do so
  through `ctx.EventManger().EmitEvent(s)`.
  To reset or wipe emitted events: `ctx = ctx.WithEventManager(sdk.NewEventManager())`
  To get all emitted events: `events := ctx.EventManager().Events()`
* [\#4437](https://github.com/cosmos/cosmos-sdk/issues/4437) Replace governance module store keys to use `[]byte` instead of `string`.
* [\#4451](https://github.com/cosmos/cosmos-sdk/issues/4451) Improve modularization of clients and modules:
  * Module directory structure improved and standardized
  * Aliases autogenerated
  * Auth and bank related commands are now mounted under the respective moduels
  * Client initialization and mounting standardized
* [\#4479](https://github.com/cosmos/cosmos-sdk/issues/4479) Remove codec argument redundency in client usage where
  the CLIContext's codec should be used instead.
* [\#4488](https://github.com/cosmos/cosmos-sdk/issues/4488) Decouple client tx, REST, and ultil packages from auth. These packages have
  been restructured and retrofitted into the `x/auth` module.
* [\#4521](https://github.com/cosmos/cosmos-sdk/issues/4521) Flatten x/bank structure by hiding module internals.
* [\#4525](https://github.com/cosmos/cosmos-sdk/issues/4525) Remove --cors flag, the feature is long gone.
* [\#4536](https://github.com/cosmos/cosmos-sdk/issues/4536) The `/auth/accounts/{address}` now returns a `height` in the response.
  The account is now nested under `account`.
* [\#4543](https://github.com/cosmos/cosmos-sdk/issues/4543) Account getters are no longer part of client.CLIContext() and have now moved
  to reside in the auth-specific AccountRetriever.
* [\#4588](https://github.com/cosmos/cosmos-sdk/issues/4588) Context does not depend on x/auth anymore. client/context is stripped out of the following features:
  - GetAccountDecoder()
  - CLIContext.WithAccountDecoder()
  - CLIContext.WithAccountStore()
  x/auth.AccountDecoder is unnecessary and consequently removed.
* [\#4602](https://github.com/cosmos/cosmos-sdk/issues/4602) client/input.{Buffer,Override}Stdin() functions are removed. Thanks to cobra's new release they are now redundant.
* [\#4633](https://github.com/cosmos/cosmos-sdk/issues/4633) Update old Tx search by tags APIs to use new Events
  nomenclature.
* [\#4649](https://github.com/cosmos/cosmos-sdk/issues/4649) Refactor x/crisis as per modules new specs.
* [\#3685](https://github.com/cosmos/cosmos-sdk/issues/3685) The default signature verification gas logic (`DefaultSigVerificationGasConsumer`) now specifies explicit key types rather than string pattern matching. This means that zones that depended on string matching to allow other keys will need to write a custom `SignatureVerificationGasConsumer` function.
* [\#4663](https://github.com/cosmos/cosmos-sdk/issues/4663) Refactor bank keeper by removing private functions
  - `InputOutputCoins`, `SetCoins`, `SubtractCoins` and `AddCoins` are now part of the `SendKeeper` instead of the `Keeper` interface
* (tendermint) [\#4721](https://github.com/cosmos/cosmos-sdk/pull/4721) Upgrade Tendermint to v0.32.1

### Features

* [\#4843](https://github.com/cosmos/cosmos-sdk/issues/4843) Add RegisterEvidences function in the codec package to register
  Tendermint evidence types with a given codec.
* (rest) [\#3867](https://github.com/cosmos/cosmos-sdk/issues/3867) Allow querying for genesis transaction when height query param is set to zero.
* [\#2020](https://github.com/cosmos/cosmos-sdk/issues/2020) New keys export/import command line utilities to export/import private keys in ASCII format
  that rely on Keybase's new underlying ExportPrivKey()/ImportPrivKey() API calls.
* [\#3565](https://github.com/cosmos/cosmos-sdk/issues/3565) Implement parameter change proposal support.
  Parameter change proposals can be submitted through the CLI
  or a REST endpoint. See docs for further usage.
* [\#3850](https://github.com/cosmos/cosmos-sdk/issues/3850) Add `rewards` and `commission` to distribution tx tags.
* [\#3981](https://github.com/cosmos/cosmos-sdk/issues/3981) Add support to gracefully halt a node at a given height
  via the node's `halt-height` config or CLI value.
* [\#4144](https://github.com/cosmos/cosmos-sdk/issues/4144) Allow for configurable BIP44 HD path and coin type.
* [\#4250](https://github.com/cosmos/cosmos-sdk/issues/4250) New BaseApp.{,Set}AppVersion() methods to get/set app's version string.
* [\#4263](https://github.com/cosmos/cosmos-sdk/issues/4263) Add `--read-timeout` and `--write-timeout` args to the `rest-server` command
  to support custom RPC R/W timeouts.
* [\#4271](https://github.com/cosmos/cosmos-sdk/issues/4271) Implement Coins#IsAnyGT
* [\#4318](https://github.com/cosmos/cosmos-sdk/issues/4318) Support height queries. Queries against nodes that have the queried
  height pruned will return an error.
* [\#4409](https://github.com/cosmos/cosmos-sdk/issues/4409) Implement a command that migrates exported state from one version to the next.
  The `migrate` command currently supports migrating from v0.34 to v0.36 by implementing
  necessary types for both versions.
* [\#4570](https://github.com/cosmos/cosmos-sdk/issues/4570) Move /bank/balances/{address} REST handler to x/bank/client/rest. The exposed interface is unchanged.
* Community pool spend proposal per Cosmos Hub governance proposal [\#7](https://github.com/cosmos/cosmos-sdk/issues/7) "Activate the Community Pool"

### Improvements

* (simulation) PrintAllInvariants flag will print all failed invariants
* (simulation) Add `InitialBlockHeight` flag to resume a simulation from a given block
* (simulation) [\#4670](https://github.com/cosmos/cosmos-sdk/issues/4670) Update simulation statistics to JSON format
  - Support exporting the simulation stats to a given JSON file
* [\#4775](https://github.com/cosmos/cosmos-sdk/issues/4775) Refactor CI config
* Upgrade IAVL to v0.12.4
* (tendermint) Upgrade Tendermint to v0.32.2
* (modules) [\#4751](https://github.com/cosmos/cosmos-sdk/issues/4751) update `x/genutils` to match module spec
* (keys) [\#4611](https://github.com/cosmos/cosmos-sdk/issues/4611) store keys in simapp now use a map instead of using individual literal keys
* [\#2286](https://github.com/cosmos/cosmos-sdk/issues/2286) Improve performance of CacheKVStore iterator.
* [\#3512](https://github.com/cosmos/cosmos-sdk/issues/3512) Implement Logger method on each module's keeper.
* [\#3655](https://github.com/cosmos/cosmos-sdk/issues/3655) Improve signature verification failure error message.
* [\#3774](https://github.com/cosmos/cosmos-sdk/issues/3774) add category tag to transactions for ease of filtering
* [\#3914](https://github.com/cosmos/cosmos-sdk/issues/3914) Implement invariant benchmarks and add target to makefile.
* [\#3928](https://github.com/cosmos/cosmos-sdk/issues/3928) remove staking references from types package
* [\#3978](https://github.com/cosmos/cosmos-sdk/issues/3978) Return ErrUnknownRequest in message handlers for unknown
  or invalid routed messages.
* [\#4190](https://github.com/cosmos/cosmos-sdk/issues/4190) Client responses that return (re)delegation(s) now return balances
  instead of shares.
* [\#4194](https://github.com/cosmos/cosmos-sdk/issues/4194) ValidatorSigningInfo now includes the validator's consensus address.
* [\#4235](https://github.com/cosmos/cosmos-sdk/issues/4235) Add parameter change proposal messages to simulation.
* [\#4235](https://github.com/cosmos/cosmos-sdk/issues/4235) Update the minting module params to implement params.ParamSet so
  individual keys can be set via proposals instead of passing a struct.
* [\#4259](https://github.com/cosmos/cosmos-sdk/issues/4259) `Coins` that are `nil` are now JSON encoded as an empty array `[]`.
  Decoding remains unchanged and behavior is left intact.
* [\#4305](https://github.com/cosmos/cosmos-sdk/issues/4305) The `--generate-only` CLI flag fully respects offline tx processing.
* [\#4379](https://github.com/cosmos/cosmos-sdk/issues/4379) close db write batch.
* [\#4384](https://github.com/cosmos/cosmos-sdk/issues/4384)- Allow splitting withdrawal transaction in several chunks
* [\#4403](https://github.com/cosmos/cosmos-sdk/issues/4403) Allow for parameter change proposals to supply only desired fields to be updated
  in objects instead of the entire object (only applies to values that are objects).
* [\#4415](https://github.com/cosmos/cosmos-sdk/issues/4415) /client refactor, reduce genutil dependancy on staking
* [\#4439](https://github.com/cosmos/cosmos-sdk/issues/4439) Implement governance module iterators.
* [\#4465](https://github.com/cosmos/cosmos-sdk/issues/4465) Unknown subcommands print relevant error message
* [\#4466](https://github.com/cosmos/cosmos-sdk/issues/4466) Commission validation added to validate basic of MsgCreateValidator by changing CommissionMsg to CommissionRates
* [\#4501](https://github.com/cosmos/cosmos-sdk/issues/4501) Support height queriers in rest client
* [\#4535](https://github.com/cosmos/cosmos-sdk/issues/4535) Improve import-export simulation errors by decoding the `KVPair.Value` into its
  respective type
* [\#4536](https://github.com/cosmos/cosmos-sdk/issues/4536) cli context queries return query height and accounts are returned with query height
* [\#4553](https://github.com/cosmos/cosmos-sdk/issues/4553) undelegate max entries check first
* [\#4556](https://github.com/cosmos/cosmos-sdk/issues/4556) Added IsValid function to Coin
* [\#4564](https://github.com/cosmos/cosmos-sdk/issues/4564) client/input.GetConfirmation()'s default is changed to No.
* [\#4573](https://github.com/cosmos/cosmos-sdk/issues/4573) Returns height in response for query endpoints.
* [\#4580](https://github.com/cosmos/cosmos-sdk/issues/4580) Update `Context#BlockHeight` to properly set the block height via `WithBlockHeader`.
* [\#4584](https://github.com/cosmos/cosmos-sdk/issues/4584) Update bank Keeper to use expected keeper interface of the AccountKeeper.
* [\#4584](https://github.com/cosmos/cosmos-sdk/issues/4584) Move `Account` and `VestingAccount` interface types to `x/auth/exported`.
* [\#4082](https://github.com/cosmos/cosmos-sdk/issues/4082) supply module queriers for CLI and REST endpoints
* [\#4601](https://github.com/cosmos/cosmos-sdk/issues/4601) Implement generic pangination helper function to be used in
  REST handlers and queriers.
* [\#4629](https://github.com/cosmos/cosmos-sdk/issues/4629) Added warning event that gets emitted if validator misses a block.
* [\#4674](https://github.com/cosmos/cosmos-sdk/issues/4674) Export `Simapp` genState generators and util functions by making them public
* [\#4706](https://github.com/cosmos/cosmos-sdk/issues/4706) Simplify context
  Replace complex Context construct with a simpler immutible struct.
  Only breaking change is not to support `Value` and `GetValue` as first class calls.
  We do embed ctx.Context() as a raw context.Context instead to be used as you see fit.

  Migration guide:

  ```go
  ctx = ctx.WithValue(contextKeyBadProposal, false)
  ```

  Now becomes:

  ```go
  ctx = ctx.WithContext(context.WithValue(ctx.Context(), contextKeyBadProposal, false))
  ```

  A bit more verbose, but also allows `context.WithTimeout()`, etc and only used
  in one function in this repo, in test code.
* [\#3685](https://github.com/cosmos/cosmos-sdk/issues/3685)  Add `SetAddressVerifier` and `GetAddressVerifier` to `sdk.Config` to allow SDK users to configure custom address format verification logic (to override the default limitation of 20-byte addresses).
* [\#3685](https://github.com/cosmos/cosmos-sdk/issues/3685)  Add an additional parameter to NewAnteHandler for a custom `SignatureVerificationGasConsumer` (the default logic is now in `DefaultSigVerificationGasConsumer). This allows SDK users to configure their own logic for which key types are accepted and how those key types consume gas.
* Remove `--print-response` flag as it is no longer used.
* Revert [\#2284](https://github.com/cosmos/cosmos-sdk/pull/2284) to allow create_empty_blocks in the config
* (tendermint) [\#4718](https://github.com/cosmos/cosmos-sdk/issues/4718) Upgrade tendermint/iavl to v0.12.3

### Bug Fixes

* [\#4891](https://github.com/cosmos/cosmos-sdk/issues/4891) Disable querying with proofs enabled when the query height <= 1.
* (rest) [\#4858](https://github.com/cosmos/cosmos-sdk/issues/4858) Do not return an error in BroadcastTxCommit when the tx broadcasting
  was successful. This allows the proper REST response to be returned for a
  failed tx during `block` broadcasting mode.
* (store) [\#4880](https://github.com/cosmos/cosmos-sdk/pull/4880) Fix error check in
  IAVL `Store#DeleteVersion`.
* (tendermint) [\#4879](https://github.com/cosmos/cosmos-sdk/issues/4879) Don't terminate the process immediately after startup when run in standalone mode.
* (simulation) [\#4861](https://github.com/cosmos/cosmos-sdk/pull/4861) Fix non-determinism simulation
  by using CLI flags as input and updating Makefile target.
* [\#4868](https://github.com/cosmos/cosmos-sdk/issues/4868) Context#CacheContext now sets a new EventManager. This prevents unwanted events
  from being emitted.
* (cli) [\#4870](https://github.com/cosmos/cosmos-sdk/issues/4870) Disable the `withdraw-all-rewards` command when `--generate-only` is supplied
* (modules) [\#4831](https://github.com/cosmos/cosmos-sdk/issues/4831) Prevent community spend proposal from transferring funds to a module account
* (keys) [\#4338](https://github.com/cosmos/cosmos-sdk/issues/4338) fix multisig key output for CLI
* (modules) [\#4795](https://github.com/cosmos/cosmos-sdk/issues/4795) restrict module accounts from receiving transactions.
  Allowing this would cause an invariant on the module account coins.
* (modules) [\#4823](https://github.com/cosmos/cosmos-sdk/issues/4823) Update the `DefaultUnbondingTime` from 3 days to 3 weeks to be inline with documentation.
* (abci) [\#4639](https://github.com/cosmos/cosmos-sdk/issues/4639) Fix `CheckTx` by verifying the message route
* Return height in responses when querying against BaseApp
* [\#1351](https://github.com/cosmos/cosmos-sdk/issues/1351) Stable AppHash allows no_empty_blocks
* [\#3705](https://github.com/cosmos/cosmos-sdk/issues/3705) Return `[]` instead of `null` when querying delegator rewards.
* [\#3966](https://github.com/cosmos/cosmos-sdk/issues/3966) fixed multiple assigns to action tags
  [\#3793](https://github.com/cosmos/cosmos-sdk/issues/3793) add delegator tag for MsgCreateValidator and deleted unused moniker and identity tags
* [\#4194](https://github.com/cosmos/cosmos-sdk/issues/4194) Fix pagination and results returned from /slashing/signing_infos
* [\#4230](https://github.com/cosmos/cosmos-sdk/issues/4230) Properly set and display the message index through the TxResponse.
* [\#4234](https://github.com/cosmos/cosmos-sdk/pull/4234) Allow `tx send --generate-only` to
  actually work offline.
* [\#4271](https://github.com/cosmos/cosmos-sdk/issues/4271) Fix addGenesisAccount by using Coins#IsAnyGT for vesting amount validation.
* [\#4273](https://github.com/cosmos/cosmos-sdk/issues/4273) Fix usage of AppendTags in x/staking/handler.go
* [\#4303](https://github.com/cosmos/cosmos-sdk/issues/4303) Fix NewCoins() underlying function for duplicate coins detection.
* [\#4307](https://github.com/cosmos/cosmos-sdk/pull/4307) Don't pass height to RPC calls as
  Tendermint will automatically use the latest height.
* [\#4362](https://github.com/cosmos/cosmos-sdk/issues/4362) simulation setup bugfix for multisim 7601778
* [\#4383](https://github.com/cosmos/cosmos-sdk/issues/4383) - currentStakeRoundUp is now always atleast currentStake + smallest-decimal-precision
* [\#4394](https://github.com/cosmos/cosmos-sdk/issues/4394) Fix signature count check to use the TxSigLimit param instead of
  a default.
* [\#4455](https://github.com/cosmos/cosmos-sdk/issues/4455) Use `QueryWithData()` to query unbonding delegations.
* [\#4493](https://github.com/cosmos/cosmos-sdk/issues/4493) Fix validator-outstanding-rewards command. It now takes as an argument
  a validator address.
* [\#4598](https://github.com/cosmos/cosmos-sdk/issues/4598) Fix redelegation and undelegation txs that were not checking for the correct bond denomination.
* [\#4619](https://github.com/cosmos/cosmos-sdk/issues/4619) Close iterators in `GetAllMatureValidatorQueue` and `UnbondAllMatureValidatorQueue`
  methods.
* [\#4654](https://github.com/cosmos/cosmos-sdk/issues/4654) validator slash event stored by period and height
* [\#4681](https://github.com/cosmos/cosmos-sdk/issues/4681) panic on invalid amount on `MintCoins` and `BurnCoins`
  * skip minting if inflation is set to zero
* Sort state JSON during export and initialization

## 0.35.0

### Bug Fixes

* Fix gas consumption bug in `Undelegate` preventing the ability to sync from
genesis.

## 0.34.7

### Bug Fixes

#### SDK

* Fix gas consumption bug in `Undelegate` preventing the ability to sync from
genesis.

## 0.34.6

### Bug Fixes

#### SDK

* Unbonding from a validator is now only considered "complete" after the full
unbonding period has elapsed regardless of the validator's status.

## 0.34.5

### Bug Fixes

#### SDK

* [\#4273](https://github.com/cosmos/cosmos-sdk/issues/4273) Fix usage of `AppendTags` in x/staking/handler.go

### Improvements

### SDK

* [\#2286](https://github.com/cosmos/cosmos-sdk/issues/2286) Improve performance of `CacheKVStore` iterator.
* [\#3655](https://github.com/cosmos/cosmos-sdk/issues/3655) Improve signature verification failure error message.
* [\#4384](https://github.com/cosmos/cosmos-sdk/issues/4384) Allow splitting withdrawal transaction in several chunks.

#### Gaia CLI

* [\#4227](https://github.com/cosmos/cosmos-sdk/issues/4227) Support for Ledger App v1.5.
* [#4345](https://github.com/cosmos/cosmos-sdk/pull/4345) Update `ledger-cosmos-go`
to v0.10.3.

## 0.34.4

### Bug Fixes

#### SDK

* [#4234](https://github.com/cosmos/cosmos-sdk/pull/4234) Allow `tx send --generate-only` to
actually work offline.

#### Gaia

* [\#4219](https://github.com/cosmos/cosmos-sdk/issues/4219) Return an error when an empty mnemonic is provided during key recovery.

### Improvements

#### Gaia

* [\#2007](https://github.com/cosmos/cosmos-sdk/issues/2007) Return 200 status code on empty results

### New features

#### SDK

* [\#3850](https://github.com/cosmos/cosmos-sdk/issues/3850) Add `rewards` and `commission` to distribution tx tags.

## 0.34.3

### Bug Fixes

#### Gaia

* [\#4196](https://github.com/cosmos/cosmos-sdk/pull/4196) Set default invariant
check period to zero.

## 0.34.2

### Improvements

#### SDK

* [\#4135](https://github.com/cosmos/cosmos-sdk/pull/4135) Add further clarification
to generate only usage.

### Bug Fixes

#### SDK

* [\#4135](https://github.com/cosmos/cosmos-sdk/pull/4135) Fix `NewResponseFormatBroadcastTxCommit`
* [\#4053](https://github.com/cosmos/cosmos-sdk/issues/4053) Add `--inv-check-period`
flag to gaiad to set period at which invariants checks will run.
* [\#4099](https://github.com/cosmos/cosmos-sdk/issues/4099) Update the /staking/validators endpoint to support
status and pagination query flags.

## 0.34.1

### Bug Fixes

#### Gaia

* [#4163](https://github.com/cosmos/cosmos-sdk/pull/4163) Fix v0.33.x export script to port gov data correctly.

## 0.34.0

### Breaking Changes

#### Gaia

* [\#3463](https://github.com/cosmos/cosmos-sdk/issues/3463) Revert bank module handler fork (re-enables transfers)
* [\#3875](https://github.com/cosmos/cosmos-sdk/issues/3875) Replace `async` flag with `--broadcast-mode` flag where the default
  value is `sync`. The `block` mode should not be used. The REST client now
  uses `mode` parameter instead of the `return` parameter.

#### Gaia CLI

* [\#3938](https://github.com/cosmos/cosmos-sdk/issues/3938) Remove REST server's SSL support altogether.

#### SDK

* [\#3245](https://github.com/cosmos/cosmos-sdk/issues/3245) Rename validator.GetJailed() to validator.IsJailed()
* [\#3516](https://github.com/cosmos/cosmos-sdk/issues/3516) Remove concept of shares from staking unbonding and redelegation UX;
  replaced by direct coin amount.

#### Tendermint

* [\#4029](https://github.com/cosmos/cosmos-sdk/issues/4029) Upgrade Tendermint to v0.31.3

### New features

#### SDK

* [\#2935](https://github.com/cosmos/cosmos-sdk/issues/2935) New module Crisis which can test broken invariant with messages
* [\#3813](https://github.com/cosmos/cosmos-sdk/issues/3813) New sdk.NewCoins safe constructor to replace bare sdk.Coins{} declarations.
* [\#3858](https://github.com/cosmos/cosmos-sdk/issues/3858) add website, details and identity to gentx cli command
* Implement coin conversion and denomination registration utilities

#### Gaia

* [\#2935](https://github.com/cosmos/cosmos-sdk/issues/2935) Optionally assert invariants on a blockly basis using `gaiad --assert-invariants-blockly`
* [\#3886](https://github.com/cosmos/cosmos-sdk/issues/3886) Implement minting module querier and CLI/REST clients.

#### Gaia CLI

* [\#3937](https://github.com/cosmos/cosmos-sdk/issues/3937) Add command to query community-pool

#### Gaia REST API

* [\#3937](https://github.com/cosmos/cosmos-sdk/issues/3937) Add route to fetch community-pool
* [\#3949](https://github.com/cosmos/cosmos-sdk/issues/3949) added /slashing/signing_infos to get signing_info for all validators

### Improvements

#### Gaia

* [\#3808](https://github.com/cosmos/cosmos-sdk/issues/3808) `gaiad` and `gaiacli` integration tests use ./build/ binaries.
* \[\#3819](https://github.com/cosmos/cosmos-sdk/issues/3819) Simulation refactor, log output now stored in ~/.gaiad/simulation/
  * Simulation moved to its own module (not a part of mock)
  * Logger type instead of passing function variables everywhere
  * Logger json output (for reloadable simulation running)
  * Cleanup bank simulation messages / remove dup code in bank simulation
  * Simulations saved in `~/.gaiad/simulations/`
  * "Lean" simulation output option to exclude No-ops and !ok functions (`--SimulationLean` flag)
* [\#3893](https://github.com/cosmos/cosmos-sdk/issues/3893) Improve `gaiacli tx sign` command
  * Add shorthand flags -a and -s for the account and sequence numbers respectively
  * Mark the account and sequence numbers required during "offline" mode
  * Always do an RPC query for account and sequence number during "online" mode
* [\#4018](https://github.com/cosmos/cosmos-sdk/issues/4018) create genesis port script for release v.0.34.0

#### Gaia CLI

* [\#3833](https://github.com/cosmos/cosmos-sdk/issues/3833) Modify stake to atom in gaia's doc.
* [\#3841](https://github.com/cosmos/cosmos-sdk/issues/3841) Add indent to JSON of `gaiacli keys [add|show|list]`
* [\#3859](https://github.com/cosmos/cosmos-sdk/issues/3859) Add newline to echo of `gaiacli keys ...`
* [\#3959](https://github.com/cosmos/cosmos-sdk/issues/3959) Improving error messages when signing with ledger devices fails

#### SDK

* [\#3238](https://github.com/cosmos/cosmos-sdk/issues/3238) Add block time to tx responses when querying for
  txs by tags or hash.
* \[\#3752](https://github.com/cosmos/cosmos-sdk/issues/3752) Explanatory docs for minting mechanism (`docs/spec/mint/01_concepts.md`)
* [\#3801](https://github.com/cosmos/cosmos-sdk/issues/3801) `baseapp` safety improvements
* [\#3820](https://github.com/cosmos/cosmos-sdk/issues/3820) Make Coins.IsAllGT() more robust and consistent.
* [\#3828](https://github.com/cosmos/cosmos-sdk/issues/3828) New sdkch tool to maintain changelogs
* [\#3864](https://github.com/cosmos/cosmos-sdk/issues/3864) Make Coins.IsAllGTE() more consistent.
* [\#3907](https://github.com/cosmos/cosmos-sdk/issues/3907): dep -> go mod migration
  * Drop dep in favor of go modules.
  * Upgrade to Go 1.12.1.
* [\#3917](https://github.com/cosmos/cosmos-sdk/issues/3917) Allow arbitrary decreases to validator commission rates.
* [\#3937](https://github.com/cosmos/cosmos-sdk/issues/3937) Implement community pool querier.
* [\#3940](https://github.com/cosmos/cosmos-sdk/issues/3940) Codespace should be lowercase.
* [\#3986](https://github.com/cosmos/cosmos-sdk/issues/3986) Update the Stringer implementation of the Proposal type.
* [\#926](https://github.com/cosmos/cosmos-sdk/issues/926) circuit breaker high level explanation
* [\#3896](https://github.com/cosmos/cosmos-sdk/issues/3896) Fixed various linters warnings in the context of the gometalinter -> golangci-lint migration
* [\#3916](https://github.com/cosmos/cosmos-sdk/issues/3916) Hex encode data in tx responses

### Bug Fixes

#### Gaia

* [\#3825](https://github.com/cosmos/cosmos-sdk/issues/3825) Validate genesis before running gentx
* [\#3889](https://github.com/cosmos/cosmos-sdk/issues/3889) When `--generate-only` is provided, the Keybase is not used and as a result
  the `--from` value must be a valid Bech32 cosmos address.
* 3974 Fix go env setting in installation.md
* 3996 Change 'make get_tools' to 'make tools' in DOCS_README.md.

#### Gaia CLI

* [\#3883](https://github.com/cosmos/cosmos-sdk/issues/3883) Remove Height Flag from CLI Queries
* [\#3899](https://github.com/cosmos/cosmos-sdk/issues/3899) Using 'gaiacli config node' breaks ~/config/config.toml

#### SDK

* [\#3837](https://github.com/cosmos/cosmos-sdk/issues/3837) Fix `WithdrawValidatorCommission` to properly set the validator's remaining commission.
* [\#3870](https://github.com/cosmos/cosmos-sdk/issues/3870) Fix DecCoins#TruncateDecimal to never return zero coins in
  either the truncated coins or the change coins.
* [\#3915](https://github.com/cosmos/cosmos-sdk/issues/3915) Remove ';' delimiting support from ParseDecCoins
* [\#3977](https://github.com/cosmos/cosmos-sdk/issues/3977) Fix docker image build
* [\#4020](https://github.com/cosmos/cosmos-sdk/issues/4020) Fix queryDelegationRewards by returning an error
when the validator or delegation do not exist.
* [\#4050](https://github.com/cosmos/cosmos-sdk/issues/4050) Fix DecCoins APIs
where rounding or truncation could result in zero decimal coins.
* [\#4088](https://github.com/cosmos/cosmos-sdk/issues/4088) Fix `calculateDelegationRewards`
by accounting for rounding errors when multiplying stake by slashing fractions.

## 0.33.2

### Improvements

#### Tendermint

* Upgrade Tendermint to `v0.31.0-dev0-fix0` which includes critical security fixes.

## 0.33.1

### Bug Fixes

#### Gaia

* [\#3999](https://github.com/cosmos/cosmos-sdk/pull/3999) Fix distribution delegation for zero height export bug

## 0.33.0

BREAKING CHANGES

* Gaia REST API
  * [\#3641](https://github.com/cosmos/cosmos-sdk/pull/3641) Remove the ability to use a Keybase from the REST API client:
    * `password` and `generate_only` have been removed from the `base_req` object
    * All txs that used to sign or use the Keybase now only generate the tx
    * `keys` routes completely removed
  * [\#3692](https://github.com/cosmos/cosmos-sdk/pull/3692) Update tx encoding and broadcasting endpoints:
    * Remove duplicate broadcasting endpoints in favor of POST @ `/txs`
      * The `Tx` field now accepts a `StdTx` and not raw tx bytes
    * Move encoding endpoint to `/txs/encode`

* Gaia
  * [\#3787](https://github.com/cosmos/cosmos-sdk/pull/3787) Fork the `x/bank` module into the Gaia application with only a
  modified message handler, where the modified message handler behaves the same as
  the standard `x/bank` message handler except for `MsgMultiSend` that must burn
  exactly 9 atoms and transfer 1 atom, and `MsgSend` is disabled.
  * [\#3789](https://github.com/cosmos/cosmos-sdk/pull/3789) Update validator creation flow:
    * Remove `NewMsgCreateValidatorOnBehalfOf` and corresponding business logic
    * Ensure the validator address equals the delegator address during
    `MsgCreateValidator#ValidateBasic`

* SDK
  * [\#3750](https://github.com/cosmos/cosmos-sdk/issues/3750) Track outstanding rewards per-validator instead of globally,
           and fix the main simulation issue, which was that slashes of
           re-delegations to a validator were not correctly accounted for
           in fee distribution when the redelegation in question had itself
            been slashed (from a fault committed by a different validator)
           in the same BeginBlock. Outstanding rewards are now available
           on a per-validator basis in REST.
  * [\#3669](https://github.com/cosmos/cosmos-sdk/pull/3669) Ensure consistency in message naming, codec registration, and JSON
  tags.
  * [\#3788](https://github.com/cosmos/cosmos-sdk/pull/3788) Change order of operations for greater accuracy when calculating delegation share token value
  * [\#3788](https://github.com/cosmos/cosmos-sdk/pull/3788) DecCoins.Cap -> DecCoins.Intersect
  * [\#3666](https://github.com/cosmos/cosmos-sdk/pull/3666) Improve coins denom validation.
  * [\#3751](https://github.com/cosmos/cosmos-sdk/pull/3751) Disable (temporarily) support for ED25519 account key pairs.

* Tendermint
  * [\#3804] Update to Tendermint `v0.31.0-dev0`

FEATURES

* SDK
  * [\#3719](https://github.com/cosmos/cosmos-sdk/issues/3719) DBBackend can now be set at compile time.
    Defaults: goleveldb. Supported: cleveldb.

IMPROVEMENTS

* Gaia REST API
  * Update the `TxResponse` type allowing for the `Logs` result to be JSON decoded automatically.

* Gaia CLI
  * [\#3653](https://github.com/cosmos/cosmos-sdk/pull/3653) Prompt user confirmation prior to signing and broadcasting a transaction.
  * [\#3670](https://github.com/cosmos/cosmos-sdk/pull/3670) CLI support for showing bech32 addresses in Ledger devices
  * [\#3711](https://github.com/cosmos/cosmos-sdk/pull/3711) Update `tx sign` to use `--from` instead of the deprecated `--name`
  CLI flag.
  * [\#3738](https://github.com/cosmos/cosmos-sdk/pull/3738) Improve multisig UX:
    * `gaiacli keys show -o json` now includes constituent pubkeys, respective weights and threshold
    * `gaiacli keys show --show-multisig` now displays constituent pubkeys, respective weights and threshold
    * `gaiacli tx sign --validate-signatures` now displays multisig signers with their respective weights
  * [\#3730](https://github.com/cosmos/cosmos-sdk/issues/3730) Improve workflow for
  `gaiad gentx` with offline public keys, by outputting stdtx file that needs to be signed.
  * [\#3761](https://github.com/cosmos/cosmos-sdk/issues/3761) Querying account related information using custom querier in auth module

* SDK
  * [\#3753](https://github.com/cosmos/cosmos-sdk/issues/3753) Remove no-longer-used governance penalty parameter
  * [\#3679](https://github.com/cosmos/cosmos-sdk/issues/3679) Consistent operators across Coins, DecCoins, Int, Dec
            replaced: Minus->Sub Plus->Add Div->Quo
  * [\#3665](https://github.com/cosmos/cosmos-sdk/pull/3665) Overhaul sdk.Uint type in preparation for Coins Int -> Uint migration.
  * [\#3691](https://github.com/cosmos/cosmos-sdk/issues/3691) Cleanup error messages
  * [\#3456](https://github.com/cosmos/cosmos-sdk/issues/3456) Integrate in the Int.ToDec() convenience function
  * [\#3300](https://github.com/cosmos/cosmos-sdk/pull/3300) Update the spec-spec, spec file reorg, and TOC updates.
  * [\#3694](https://github.com/cosmos/cosmos-sdk/pull/3694) Push tagged docker images on docker hub when tag is created.
  * [\#3716](https://github.com/cosmos/cosmos-sdk/pull/3716) Update file permissions the client keys directory and contents to `0700`.
  * [\#3681](https://github.com/cosmos/cosmos-sdk/issues/3681) Migrate ledger-cosmos-go from ZondaX to Cosmos organization

* Tendermint
  * [\#3699](https://github.com/cosmos/cosmos-sdk/pull/3699) Upgrade to Tendermint 0.30.1

BUG FIXES

* Gaia CLI
  * [\#3731](https://github.com/cosmos/cosmos-sdk/pull/3731) `keys add --interactive` bip32 passphrase regression fix
  * [\#3714](https://github.com/cosmos/cosmos-sdk/issues/3714) Fix USB raw access issues with gaiacli when installed via snap

* Gaia
  * [\#3777](https://github.com/cosmso/cosmos-sdk/pull/3777) `gaiad export` no longer panics when the database is empty
  * [\#3806](https://github.com/cosmos/cosmos-sdk/pull/3806) Properly return errors from a couple of struct Unmarshal functions

* SDK
  * [\#3728](https://github.com/cosmos/cosmos-sdk/issues/3728) Truncate decimal multiplication & division in distribution to ensure
           no more than the collected fees / inflation are distributed
  * [\#3727](https://github.com/cosmos/cosmos-sdk/issues/3727) Return on zero-length (including []byte{}) PrefixEndBytes() calls
  * [\#3559](https://github.com/cosmos/cosmos-sdk/issues/3559) fix occasional failing due to non-determinism in lcd test TestBonding
    where validator is unexpectedly slashed throwing off test calculations
  * [\#3411](https://github.com/cosmos/cosmos-sdk/pull/3411) Include the `RequestInitChain.Time` in the block header init during
  `InitChain`.
  * [\#3717](https://github.com/cosmos/cosmos-sdk/pull/3717) Update the vesting specification and implementation to cap deduction from
  `DelegatedVesting` by at most `DelegatedVesting`. This accounts for the case where
  the undelegation amount may exceed the original delegation amount due to
  truncation of undelegation tokens.
  * [\#3717](https://github.com/cosmos/cosmos-sdk/pull/3717) Ignore unknown proposers in allocating rewards for proposers, in case
    unbonding period was just 1 block and proposer was already deleted.
  * [\#3726](https://github.com/cosmos/cosmos-sdk/pull/3724) Cap(clip) reward to remaining coins in AllocateTokens.

## 0.32.0

BREAKING CHANGES

* Gaia REST API
  * [\#3642](https://github.com/cosmos/cosmos-sdk/pull/3642) `GET /tx/{hash}` now returns `404` instead of `500` if the transaction is not found

* SDK
 * [\#3580](https://github.com/cosmos/cosmos-sdk/issues/3580) Migrate HTTP request/response types and utilities to types/rest.
 * [\#3592](https://github.com/cosmos/cosmos-sdk/issues/3592) Drop deprecated keybase implementation's New() constructor in
   favor of a new crypto/keys.New(string, string) implementation that
   returns a lazy keybase instance. Remove client.MockKeyBase,
   superseded by crypto/keys.NewInMemory()
 * [\#3621](https://github.com/cosmos/cosmos-sdk/issues/3621) staking.GenesisState.Bonds -> Delegations

IMPROVEMENTS

* SDK
  * [\#3311](https://github.com/cosmos/cosmos-sdk/pull/3311) Reconcile the `DecCoin/s` API with the `Coin/s` API.
  * [\#3614](https://github.com/cosmos/cosmos-sdk/pull/3614) Add coin denom length checks to the coins constructors.
  * [\#3621](https://github.com/cosmos/cosmos-sdk/issues/3621) remove many inter-module dependancies
  * [\#3601](https://github.com/cosmos/cosmos-sdk/pull/3601) JSON-stringify the ABCI log response which includes the log and message
  index.
  * [\#3604](https://github.com/cosmos/cosmos-sdk/pull/3604) Improve SDK funds related error messages and allow for unicode in
  JSON ABCI log.
  * [\#3620](https://github.com/cosmos/cosmos-sdk/pull/3620) Version command shows build tags
  * [\#3638](https://github.com/cosmos/cosmos-sdk/pull/3638) Add Bcrypt benchmarks & justification of security parameter choice
  * [\#3648](https://github.com/cosmos/cosmos-sdk/pull/3648) Add JSON struct tags to vesting accounts.

* Tendermint
  * [\#3618](https://github.com/cosmos/cosmos-sdk/pull/3618) Upgrade to Tendermint 0.30.03

BUG FIXES

* SDK
  * [\#3646](https://github.com/cosmos/cosmos-sdk/issues/3646) `x/mint` now uses total token supply instead of total bonded tokens to calculate inflation


## 0.31.2

BREAKING CHANGES

* SDK
 * [\#3592](https://github.com/cosmos/cosmos-sdk/issues/3592) Drop deprecated keybase implementation's
   New constructor in favor of a new
   crypto/keys.New(string, string) implementation that
   returns a lazy keybase instance. Remove client.MockKeyBase,
   superseded by crypto/keys.NewInMemory()

IMPROVEMENTS

* SDK
  * [\#3604](https://github.com/cosmos/cosmos-sdk/pulls/3604) Improve SDK funds related error messages and allow for unicode in
  JSON ABCI log.

* Tendermint
  * [\#3563](https://github.com/cosmos/cosmos-sdk/3563) Update to Tendermint version `0.30.0-rc0`


BUG FIXES

* Gaia
  * [\#3585] Fix setting the tx hash in `NewResponseFormatBroadcastTxCommit`.
  * [\#3585] Return an empty `TxResponse` when Tendermint returns an empty
  `ResultBroadcastTx`.

* SDK
  * [\#3582](https://github.com/cosmos/cosmos-sdk/pull/3582) Running `make test_unit` was failing due to a missing tag
  * [\#3617](https://github.com/cosmos/cosmos-sdk/pull/3582) Fix fee comparison when the required fees does not contain any denom
  present in the tx fees.

## 0.31.0

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
  * [\#2513](https://github.com/cosmos/cosmos-sdk/issues/2513) Tendermint updates are adjusted by 10^-6 relative to staking tokens,
  * [\#3487](https://github.com/cosmos/cosmos-sdk/pull/3487) Move HTTP/REST utilities out of client/utils into a new dedicated client/rest package.
  * [\#3490](https://github.com/cosmos/cosmos-sdk/issues/3490) ReadRESTReq() returns bool to avoid callers to write error responses twice.
  * [\#3502](https://github.com/cosmos/cosmos-sdk/pull/3502) Fixes issue when comparing genesis states
  * [\#3514](https://github.com/cosmos/cosmos-sdk/pull/3514) Various clean ups:
    - Replace all GetKeyBase\* functions family in favor of NewKeyBaseFromDir and NewKeyBaseFromHomeFlag.
    - Remove Get prefix from all TxBuilder's getters.
  * [\#3522](https://github.com/cosmos/cosmos-sdk/pull/3522) Get rid of double negatives: Coins.IsNotNegative() -> Coins.IsAnyNegative().
  * [\#3561](https://github.com/cosmos/cosmos-sdk/issues/3561) Don't unnecessarily store denominations in staking


FEATURES

* Gaia REST API
  * [\#2358](https://github.com/cosmos/cosmos-sdk/issues/2358) Add distribution module REST interface

* Gaia CLI  (`gaiacli`)
  * [\#3429](https://github.com/cosmos/cosmos-sdk/issues/3429) Support querying
  for all delegator distribution rewards.
  * [\#3449](https://github.com/cosmos/cosmos-sdk/issues/3449) Proof verification now works with absence proofs
  * [\#3484](https://github.com/cosmos/cosmos-sdk/issues/3484) Add support
  vesting accounts to the add-genesis-account command.

* Gaia
  - [\#3397](https://github.com/cosmos/cosmos-sdk/pull/3397) Implement genesis file sanitization to avoid failures at chain init.
  * [\#3428](https://github.com/cosmos/cosmos-sdk/issues/3428) Run the simulation from a particular genesis state loaded from a file

* SDK
  * [\#3270](https://github.com/cosmos/cosmos-sdk/issues/3270) [x/staking] limit number of ongoing unbonding delegations /redelegations per pair/trio
  * [\#3477][distribution] new query endpoint "delegator_validators"
  * [\#3514](https://github.com/cosmos/cosmos-sdk/pull/3514) Provided a lazy loading implementation of Keybase that locks the underlying
    storage only for the time needed to perform the required operation. Also added Keybase reference to TxBuilder struct.
  * [types] [\#2580](https://github.com/cosmos/cosmos-sdk/issues/2580) Addresses now Bech32 empty addresses to an empty string


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
  * [\#3507](https://github.com/cosmos/cosmos-sdk/issues/3507) General cleanup, removal of unnecessary struct fields, undelegation bugfix, and comment clarification in x/staking and x/slashing

* SDK
  * [\#2605] x/params add subkey accessing
  * [\#2986](https://github.com/cosmos/cosmos-sdk/pull/2986) Store Refactor
  * [\#3435](https://github.com/cosmos/cosmos-sdk/issues/3435) Test that store implementations do not allow nil values
  * [\#2509](https://github.com/cosmos/cosmos-sdk/issues/2509) Sanitize all usage of Dec.RoundInt64()
  * [\#556](https://github.com/cosmos/cosmos-sdk/issues/556) Increase `BaseApp`
  test coverage.
  * [\#3357](https://github.com/cosmos/cosmos-sdk/issues/3357) develop state-transitions.md for staking spec, missing states added to `state.md`
  * [\#3552](https://github.com/cosmos/cosmos-sdk/pull/3552) Validate bit length when
  deserializing `Int` types.


BUG FIXES

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


## 0.30.0

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
  * [\#3249\(https://github.com/cosmos/cosmos-sdk/issues/3249) `tendermint`'s `show-validator` and `show-address` `--json` flags removed in favor of `--output-format=json`.

* SDK
  * [distribution] [\#3359](https://github.com/cosmos/cosmos-sdk/issues/3359) Always round down when calculating rewards-to-be-withdrawn in F1 fee distribution
  * [#3336](https://github.com/cosmos/cosmos-sdk/issues/3336) Ensure all SDK
  messages have their signature bytes contain canonical fields `value` and `type`.
  * [\#3333](https://github.com/cosmos/cosmos-sdk/issues/3333) - F1 storage efficiency improvements - automatic withdrawals when unbonded, historical reward reference counting
  * [staking] [\#2513](https://github.com/cosmos/cosmos-sdk/issues/2513) Validator power type from Dec -> Int
  * [staking] [\#3233](https://github.com/cosmos/cosmos-sdk/issues/3233) key and value now contain duplicate fields to simplify code
  * [\#3064](https://github.com/cosmos/cosmos-sdk/issues/3064) Sanitize `sdk.Coin` denom. Coins denoms are now case insensitive, i.e. 100fooToken equals to 100FOOTOKEN.
  * [\#3195](https://github.com/cosmos/cosmos-sdk/issues/3195) Allows custom configuration for syncable strategy
  * [\#3242](https://github.com/cosmos/cosmos-sdk/issues/3242) Fix infinite gas
    meter utilization during aborted ante handler executions.
  * [x/distribution] [\#3292](https://github.com/cosmos/cosmos-sdk/issues/3292) Enable or disable withdraw addresses with a parameter in the param store
  * [staking] [\#2222](https://github.com/cosmos/cosmos-sdk/issues/2222) `/stake` -> `/staking` module rename
  * [staking] [\#1268](https://github.com/cosmos/cosmos-sdk/issues/1268) `LooseTokens` -> `NotBondedTokens`
  * [staking] [\#1402](https://github.com/cosmos/cosmos-sdk/issues/1402) Redelegation and unbonding-delegation structs changed to include multiple an array of entries
  * [staking] [\#3289](https://github.com/cosmos/cosmos-sdk/issues/3289) misc renames:
    * `Validator.UnbondingMinTime` -> `Validator.UnbondingCompletionTime`
    * `Delegation` -> `Value` in `MsgCreateValidator` and `MsgDelegate`
    * `MsgBeginUnbonding` -> `MsgUndelegate`
  * [\#3315] Increase decimal precision to 18
  * [\#3323](https://github.com/cosmos/cosmos-sdk/issues/3323) Update to Tendermint 0.29.0
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
  * [\#2399](https://github.com/cosmos/cosmos-sdk/issues/2399) Implement `params` command to query slashing parameters.
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
  - [\#3099](https://github.com/cosmos/cosmos-sdk/issues/3099) Implement F1 fee distribution
  - [\#2926](https://github.com/cosmos/cosmos-sdk/issues/2926) Add TxEncoder to client TxBuilder.
  * [\#2694](https://github.com/cosmos/cosmos-sdk/issues/2694) Vesting account implementation.
  * [\#2996](https://github.com/cosmos/cosmos-sdk/issues/2996) Update the `AccountKeeper` to contain params used in the context of
  the ante handler.
  * [\#3179](https://github.com/cosmos/cosmos-sdk/pull/3179) New CodeNoSignatures error code.
  * [\#3319](https://github.com/cosmos/cosmos-sdk/issues/3319) [x/distribution] Queriers for all distribution state worth querying; distribution query commands
  * [\#3356](https://github.com/cosmos/cosmos-sdk/issues/3356) [x/auth] bech32-ify accounts address in error message.

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
    * `gaiad start` and `gaia.toml` take --minimum-gas-prices flag and minimum-gas-price config key respectively.
  * [\#2859](https://github.com/cosmos/cosmos-sdk/issues/2859) Rename `TallyResult` in gov proposals to `FinalTallyResult`
  * [\#3286](https://github.com/cosmos/cosmos-sdk/pull/3286) Fix `gaiad gentx` printout of account's addresses, i.e. user bech32 instead of hex.
  * [\#3249\(https://github.com/cosmos/cosmos-sdk/issues/3249) `--json` flag removed, users should use `--output=json` instead.

* SDK
  * [\#3137](https://github.com/cosmos/cosmos-sdk/pull/3137) Add tag documentation
    for each module along with cleaning up a few existing tags in the governance,
    slashing, and staking modules.
  * [\#3093](https://github.com/cosmos/cosmos-sdk/issues/3093) Ante handler does no longer read all accounts in one go when processing signatures as signature
    verification may fail before last signature is checked.
  * [staking] [\#1402](https://github.com/cosmos/cosmos-sdk/issues/1402) Add for multiple simultaneous redelegations or unbonding-delegations within an unbonding period
  * [staking] [\#1268](https://github.com/cosmos/cosmos-sdk/issues/1268) staking spec rewrite

* CI
  * [\#2498](https://github.com/cosmos/cosmos-sdk/issues/2498) Added macos CI job to CircleCI
  * [#142](https://github.com/tendermint/devops/issues/142) Increased the number of blocks to be tested during multi-sim
  * [#147](https://github.com/tendermint/devops/issues/142) Added docker image build to CI

BUG FIXES

* Gaia CLI  (`gaiacli`)
  * [\#3141](https://github.com/cosmos/cosmos-sdk/issues/3141) Fix the bug in GetAccount when `len(res) == 0` and `err == nil`
  * [\#810](https://github.com/cosmos/cosmos-sdk/pull/3316) Fix regression in gaiacli config file handling

* Gaia
  * [\#3148](https://github.com/cosmos/cosmos-sdk/issues/3148) Fix `gaiad export` by adding a boolean to `NewGaiaApp` determining whether or not to load the latest version
  * [\#3181](https://github.com/cosmos/cosmos-sdk/issues/3181) Correctly reset total accum update height and jailed-validator bond height / unbonding height on export-for-zero-height
  * [\#3172](https://github.com/cosmos/cosmos-sdk/pull/3172) Fix parsing `gaiad.toml`
  when it already exists.
  * [\#3223](https://github.com/cosmos/cosmos-sdk/issues/3223) Fix unset governance proposal queues when importing state from old chain
  * [#3187](https://github.com/cosmos/cosmos-sdk/issues/3187) Fix `gaiad export`
  by resetting each validator's slashing period.

## 0.29.1

BUG FIXES

* SDK
  * [\#3207](https://github.com/cosmos/cosmos-sdk/issues/3207) - Fix token printing bug

## 0.29.0

BREAKING CHANGES

* Gaia
  * [\#3148](https://github.com/cosmos/cosmos-sdk/issues/3148) Fix `gaiad export` by adding a boolean to `NewGaiaApp` determining whether or not to load the latest version

* SDK
  * [\#3163](https://github.com/cosmos/cosmos-sdk/issues/3163) Withdraw commission on self bond removal


## 0.28.1

BREAKING CHANGES

* Gaia REST API (`gaiacli advanced rest-server`)
  * [lcd] [\#3045](https://github.com/cosmos/cosmos-sdk/pull/3045) Fix quoted json return on GET /keys (keys list)
  * [gaia-lite] [\#2191](https://github.com/cosmos/cosmos-sdk/issues/2191) Split `POST /stake/delegators/{delegatorAddr}/delegations` into `POST /stake/delegators/{delegatorAddr}/delegations`, `POST /stake/delegators/{delegatorAddr}/unbonding_delegations` and `POST /stake/delegators/{delegatorAddr}/redelegations`
  * [gaia-lite] [\#3056](https://github.com/cosmos/cosmos-sdk/pull/3056) `generate_only` and `simulate` have moved from query arguments to POST requests body.
* Tendermint
  * [tendermint] Now using Tendermint 0.27.3

FEATURES

* Gaia REST API (`gaiacli advanced rest-server`)
  * [slashing] [\#2399](https://github.com/cosmos/cosmos-sdk/issues/2399)  Implement `/slashing/parameters` endpoint to query slashing parameters.
* Gaia CLI  (`gaiacli`)
  * [gaiacli] [\#2399](https://github.com/cosmos/cosmos-sdk/issues/2399) Implement `params` command to query slashing parameters.
* SDK
  - [client] [\#2926](https://github.com/cosmos/cosmos-sdk/issues/2926) Add TxEncoder to client TxBuilder.
* Other
  - Introduced the logjack tool for saving logs w/ rotation

IMPROVEMENTS

* Gaia REST API (`gaiacli advanced rest-server`)
  * [\#2879](https://github.com/cosmos/cosmos-sdk/issues/2879), [\#2880](https://github.com/cosmos/cosmos-sdk/issues/2880) Update deposit and vote endpoints to perform a direct txs query
    when a given proposal is inactive and thus having votes and deposits removed
    from state.
* Gaia CLI  (`gaiacli`)
  * [\#2879](https://github.com/cosmos/cosmos-sdk/issues/2879), [\#2880](https://github.com/cosmos/cosmos-sdk/issues/2880) Update deposit and vote CLI commands to perform a direct txs query
    when a given proposal is inactive and thus having votes and deposits removed
    from state.
* Gaia
  * [\#3021](https://github.com/cosmos/cosmos-sdk/pull/3021) Add `--gentx-dir` to `gaiad collect-gentxs` to specify a directory from which collect and load gentxs. Add `--output-document` to `gaiad init` to allow one to redirect output to file.


## 0.28.0

BREAKING CHANGES

* Gaia CLI  (`gaiacli`)
  * [cli] [\#2595](https://github.com/cosmos/cosmos-sdk/issues/2595) Remove `keys new` in favor of `keys add` incorporating existing functionality with addition of key recovery functionality.
  * [cli] [\#2987](https://github.com/cosmos/cosmos-sdk/pull/2987) Add shorthand `-a` to `gaiacli keys show` and update docs
  * [cli] [\#2971](https://github.com/cosmos/cosmos-sdk/pull/2971) Additional verification when running `gaiad gentx`
  * [cli] [\#2734](https://github.com/cosmos/cosmos-sdk/issues/2734) Rewrite `gaiacli config`. It is now a non-interactive config utility.

* Gaia
  * [#128](https://github.com/tendermint/devops/issues/128) Updated CircleCI job to trigger website build on every push to master/develop.
  * [\#2994](https://github.com/cosmos/cosmos-sdk/pull/2994) Change wrong-password error message.
  * [\#3009](https://github.com/cosmos/cosmos-sdk/issues/3009) Added missing Gaia genesis verification
  * [#128](https://github.com/tendermint/devops/issues/128) Updated CircleCI job to trigger website build on every push to master/develop.
  * [\#2994](https://github.com/cosmos/cosmos-sdk/pull/2994) Change wrong-password error message.
  * [\#3009](https://github.com/cosmos/cosmos-sdk/issues/3009) Added missing Gaia genesis verification
  * [gas] [\#3052](https://github.com/cosmos/cosmos-sdk/issues/3052) Updated gas costs to more reasonable numbers

* SDK
  * [auth] [\#2952](https://github.com/cosmos/cosmos-sdk/issues/2952) Signatures are no longer serialized on chain with the account number and sequence number
  * [auth] [\#2952](https://github.com/cosmos/cosmos-sdk/issues/2952) Signatures are no longer serialized on chain with the account number and sequence number
  * [stake] [\#3055](https://github.com/cosmos/cosmos-sdk/issues/3055) Use address instead of bond height / intratxcounter for deduplication

FEATURES

* Gaia CLI  (`gaiacli`)
  * [\#2961](https://github.com/cosmos/cosmos-sdk/issues/2961) Add --force flag to gaiacli keys delete command to skip passphrase check and force key deletion unconditionally.

IMPROVEMENTS

* Gaia CLI  (`gaiacli`)
  * [\#2991](https://github.com/cosmos/cosmos-sdk/issues/2991) Fully validate transaction signatures during `gaiacli tx sign --validate-signatures`

* SDK
  * [\#1277](https://github.com/cosmos/cosmos-sdk/issues/1277) Complete bank module specification
  * [\#2963](https://github.com/cosmos/cosmos-sdk/issues/2963) Complete auth module specification
  * [\#2914](https://github.com/cosmos/cosmos-sdk/issues/2914) No longer withdraw validator rewards on bond/unbond, but rather move
  the rewards to the respective validator's pools.


BUG FIXES

* Gaia CLI  (`gaiacli`)
  * [\#2921](https://github.com/cosmos/cosmos-sdk/issues/2921) Fix `keys delete` inability to delete offline and ledger keys.

* Gaia
  * [\#3003](https://github.com/cosmos/cosmos-sdk/issues/3003) CollectStdTxs() must validate DelegatorAddr against genesis accounts.

* SDK
  * [\#2967](https://github.com/cosmos/cosmos-sdk/issues/2967) Change ordering of `mint.BeginBlocker` and `distr.BeginBlocker`, recalculate inflation each block
  * [\#3068](https://github.com/cosmos/cosmos-sdk/issues/3068) check for uint64 gas overflow during `Std#ValidateBasic`.
  * [\#3071](https://github.com/cosmos/cosmos-sdk/issues/3071) Catch overflow on block gas meter


## 0.27.0

BREAKING CHANGES

* Gaia REST API (`gaiacli advanced rest-server`)
  * [gaia-lite] [\#2819](https://github.com/cosmos/cosmos-sdk/pull/2819) Txs query param format is now: `/txs?tag=value` (removed '' wrapping the query parameter `value`)

* Gaia CLI  (`gaiacli`)
  * [cli] [\#2728](https://github.com/cosmos/cosmos-sdk/pull/2728) Seperate `tx` and `query` subcommands by module
  * [cli] [\#2727](https://github.com/cosmos/cosmos-sdk/pull/2727) Fix unbonding command flow
  * [cli] [\#2786](https://github.com/cosmos/cosmos-sdk/pull/2786) Fix redelegation command flow
  * [cli] [\#2829](https://github.com/cosmos/cosmos-sdk/pull/2829) add-genesis-account command now validates state when adding accounts
  * [cli] [\#2804](https://github.com/cosmos/cosmos-sdk/issues/2804) Check whether key exists before passing it on to `tx create-validator`.
  * [cli] [\#2874](https://github.com/cosmos/cosmos-sdk/pull/2874) `gaiacli tx sign` takes an optional `--output-document` flag to support output redirection.
  * [cli] [\#2875](https://github.com/cosmos/cosmos-sdk/pull/2875) Refactor `gaiad gentx` and avoid redirection to `gaiacli tx sign` for tx signing.

* Gaia
  * [mint] [\#2825] minting now occurs every block, inflation parameter updates still hourly

* SDK
  * [\#2752](https://github.com/cosmos/cosmos-sdk/pull/2752) Don't hardcode bondable denom.
  * [\#2701](https://github.com/cosmos/cosmos-sdk/issues/2701) Account numbers and sequence numbers in `auth` are now `uint64` instead of `int64`
  * [\#2019](https://github.com/cosmos/cosmos-sdk/issues/2019) Cap total number of signatures. Current per-transaction limit is 7, and if that is exceeded transaction is rejected.
  * [\#2801](https://github.com/cosmos/cosmos-sdk/pull/2801) Remove AppInit structure.
  * [\#2798](https://github.com/cosmos/cosmos-sdk/issues/2798) Governance API has miss-spelled English word in JSON response ('depositer' -> 'depositor')
  * [\#2943](https://github.com/cosmos/cosmos-sdk/pull/2943) Transaction action tags equal the message type. Staking EndBlocker tags are included.

* Tendermint
  * Update to Tendermint 0.27.0

FEATURES

* Gaia REST API (`gaiacli advanced rest-server`)
  * [gov] [\#2479](https://github.com/cosmos/cosmos-sdk/issues/2479) Added governance parameter
    query REST endpoints.

* Gaia CLI  (`gaiacli`)
  * [gov][cli] [\#2479](https://github.com/cosmos/cosmos-sdk/issues/2479) Added governance
    parameter query commands.
  * [stake][cli] [\#2027] Add CLI query command for getting all delegations to a specific validator.
  * [\#2840](https://github.com/cosmos/cosmos-sdk/pull/2840) Standardize CLI exports from modules

* Gaia
  * [app] [\#2791](https://github.com/cosmos/cosmos-sdk/issues/2791) Support export at a specific height, with `gaiad export --height=HEIGHT`.
  * [x/gov] [#2479](https://github.com/cosmos/cosmos-sdk/issues/2479) Implemented querier
  for getting governance parameters.
  * [app] [\#2663](https://github.com/cosmos/cosmos-sdk/issues/2663) - Runtime-assertable invariants
  * [app] [\#2791](https://github.com/cosmos/cosmos-sdk/issues/2791) Support export at a specific height, with `gaiad export --height=HEIGHT`.
  * [app] [\#2812](https://github.com/cosmos/cosmos-sdk/issues/2812) Support export alterations to prepare for restarting at zero-height

* SDK
  * [simulator] [\#2682](https://github.com/cosmos/cosmos-sdk/issues/2682) MsgEditValidator now looks at the validator's max rate, thus it now succeeds a significant portion of the time
  * [core] [\#2775](https://github.com/cosmos/cosmos-sdk/issues/2775) Add deliverTx maximum block gas limit


IMPROVEMENTS

* Gaia REST API (`gaiacli advanced rest-server`)
  * [gaia-lite] [\#2819](https://github.com/cosmos/cosmos-sdk/pull/2819) Tx search now supports multiple tags as query parameters
  * [\#2836](https://github.com/cosmos/cosmos-sdk/pull/2836) Expose LCD router to allow users to register routes there.

* Gaia CLI  (`gaiacli`)
  * [\#2749](https://github.com/cosmos/cosmos-sdk/pull/2749) Add --chain-id flag to gaiad testnet
  * [\#2819](https://github.com/cosmos/cosmos-sdk/pull/2819) Tx search now supports multiple tags as query parameters

* Gaia
  * [\#2772](https://github.com/cosmos/cosmos-sdk/issues/2772) Update BaseApp to not persist state when the ante handler fails on DeliverTx.
  * [\#2773](https://github.com/cosmos/cosmos-sdk/issues/2773) Require moniker to be provided on `gaiad init`.
  * [\#2672](https://github.com/cosmos/cosmos-sdk/issues/2672) [Makefile] Updated for better Windows compatibility and ledger support logic, get_tools was rewritten as a cross-compatible Makefile.
  * [\#2766](https://github.com/cosmos/cosmos-sdk/issues/2766) [Makefile] Added goimports tool to get_tools. Get_tools now only builds new versions if binaries are missing.
  * [#110](https://github.com/tendermint/devops/issues/110) Updated CircleCI job to trigger website build when cosmos docs are updated.

* SDK
 & [x/mock/simulation] [\#2720] major cleanup, introduction of helper objects, reorganization
 * [\#2821](https://github.com/cosmos/cosmos-sdk/issues/2821) Codespaces are now strings
 * [types] [\#2776](https://github.com/cosmos/cosmos-sdk/issues/2776) Improve safety of `Coin` and `Coins` types. Various functions
 and methods will panic when a negative amount is discovered.
 * [\#2815](https://github.com/cosmos/cosmos-sdk/issues/2815) Gas unit fields changed from `int64` to `uint64`.
 * [\#2821](https://github.com/cosmos/cosmos-sdk/issues/2821) Codespaces are now strings
 * [\#2779](https://github.com/cosmos/cosmos-sdk/issues/2779) Introduce `ValidateBasic` to the `Tx` interface and call it in the ante
 handler.
 * [\#2825](https://github.com/cosmos/cosmos-sdk/issues/2825) More staking and distribution invariants
 * [\#2912](https://github.com/cosmos/cosmos-sdk/issues/2912) Print commit ID in hex when commit is synced.

* Tendermint
 * [\#2796](https://github.com/cosmos/cosmos-sdk/issues/2796) Update to go-amino 0.14.1


BUG FIXES

* Gaia REST API (`gaiacli advanced rest-server`)
  * [gaia-lite] [\#2868](https://github.com/cosmos/cosmos-sdk/issues/2868) Added handler for governance tally endpoint
  * [\#2907](https://github.com/cosmos/cosmos-sdk/issues/2907) Refactor and fix the way Gaia Lite is started.

* Gaia
  * [\#2723] Use `cosmosvalcons` Bech32 prefix in `tendermint show-address`
  * [\#2742](https://github.com/cosmos/cosmos-sdk/issues/2742) Fix time format of TimeoutCommit override
  * [\#2898](https://github.com/cosmos/cosmos-sdk/issues/2898) Remove redundant '$' in docker-compose.yml

* SDK
  * [\#2733](https://github.com/cosmos/cosmos-sdk/issues/2733) [x/gov, x/mock/simulation] Fix governance simulation, update x/gov import/export
  * [\#2854](https://github.com/cosmos/cosmos-sdk/issues/2854) [x/bank] Remove unused bank.MsgIssue, prevent possible panic
  * [\#2884](https://github.com/cosmos/cosmos-sdk/issues/2884) [docs/examples] Fix `basecli version` panic

* Tendermint
  * [\#2797](https://github.com/tendermint/tendermint/pull/2797) AddressBook requires addresses to have IDs; Do not crap out immediately after sending pex addrs in seed mode

## 0.26.0

BREAKING CHANGES

* Gaia
  * [gaiad init] [\#2602](https://github.com/cosmos/cosmos-sdk/issues/2602) New genesis workflow

* SDK
  * [simulation] [\#2665](https://github.com/cosmos/cosmos-sdk/issues/2665) only argument to sdk.Invariant is now app

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
 * [\#2648](https://github.com/cosmos/cosmos-sdk/issues/2648) [gaiad] Fix `gaiad export` / `gaiad import` consistency, test in CI

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
    * [docs] Added commands for querying governance deposits, votes and tally

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
  * [docs] Fixed light client section links

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
  * Removed redundancy in names (ex. stake.StakingKeeper -> stake.Keeper)
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

<!-- Release links -->

[Unreleased]: https://github.com/cosmos/cosmos-sdk/compare/v0.37.8...HEAD
[v0.37.8]: https://github.com/cosmos/cosmos-sdk/releases/tag/v0.37.8
[v0.37.7]: https://github.com/cosmos/cosmos-sdk/releases/tag/v0.37.7
[v0.37.6]: https://github.com/cosmos/cosmos-sdk/releases/tag/v0.37.6
[v0.37.5]: https://github.com/cosmos/cosmos-sdk/releases/tag/v0.37.5
[v0.37.4]: https://github.com/cosmos/cosmos-sdk/releases/tag/v0.37.4
[v0.37.3]: https://github.com/cosmos/cosmos-sdk/releases/tag/v0.37.3
[v0.37.1]: https://github.com/cosmos/cosmos-sdk/releases/tag/v0.37.1
[v0.37.0]: https://github.com/cosmos/cosmos-sdk/releases/tag/v0.37.0
[v0.36.0]: https://github.com/cosmos/cosmos-sdk/releases/tag/v0.36.0
