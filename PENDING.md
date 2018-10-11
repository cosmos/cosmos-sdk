## PENDING

BREAKING CHANGES

* Gaia REST API (`gaiacli advanced rest-server`)
    * [x/stake] Validator.Owner renamed to Validator.Operator
    * [\#595](https://github.com/cosmos/cosmos-sdk/issues/595) Connections to the REST server are now secured using Transport Layer Security by default. The --insecure flag is provided to switch back to insecure HTTP.

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
    * [cli] \#2073 --from can now be either an address or a key name
    * [cli] [\#1184](https://github.com/cosmos/cosmos-sdk/issues/1184) Subcommands reorganisation, see [\#2390](https://github.com/cosmos/cosmos-sdk/pull/2390) for a comprehensive list of changes.

* Gaia
    * Make the transient store key use a distinct store key. [#2013](https://github.com/cosmos/cosmos-sdk/pull/2013)
    * [x/stake] [\#1901](https://github.com/cosmos/cosmos-sdk/issues/1901) Validator type's Owner field renamed to Operator; Validator's GetOwner() renamed accordingly to comply with the SDK's Validator interface.
    * [docs] [#2001](https://github.com/cosmos/cosmos-sdk/pull/2001) Update slashing spec for slashing period
    * [x/stake, x/slashing] [#1305](https://github.com/cosmos/cosmos-sdk/issues/1305) - Rename "revoked" to "jailed"
    * [x/stake] [#1676] Revoked and jailed validators put into the unbonding state
    * [x/stake] [#1877] Redelegations/unbonding-delegation from unbonding validator have reduced time
    * [x/slashing] \#1789 Slashing changes for Tendermint validator set offset (NextValSet)
    * [x/stake] [\#2040](https://github.com/cosmos/cosmos-sdk/issues/2040) Validator
    operator type has now changed to `sdk.ValAddress`
    * [x/stake] [\#2221](https://github.com/cosmos/cosmos-sdk/issues/2221) New
    Bech32 prefixes have been introduced for a validator's consensus address and
    public key: `cosmosvalcons` and `cosmosvalconspub` respectively. Also, existing Bech32 prefixes have been
    renamed for accounts and validator operators:
      * `cosmosaccaddr` / `cosmosaccpub` => `cosmos` / `cosmospub`
      * `cosmosvaladdr` / `cosmosvalpub` => `cosmosvaloper` / `cosmosvaloperpub`
    * [x/stake] [#1013] TendermintUpdates now uses transient store
    * [x/stake] \#2435 Remove empty bytes from the ValidatorPowerRank store key
    * [x/gov] [#2195] Governance uses BFT Time
    * [x/gov] \#2256 Removed slashing for governance non-voting validators
    * [simulation] \#2162 Added back correct supply invariants
    * [x/slashing] \#2430 Simulate more slashes, check if validator is jailed before jailing
    * [x/stake] \#2393 Removed `CompleteUnbonding` and `CompleteRedelegation` Msg types, and instead added unbonding/redelegation queues to endblocker
    
* SDK
    * [core] \#2219 Update to Tendermint 0.24.0
      * Validator set updates delayed by one block
      * BFT timestamp that can safely be used by applications
      * Fixed maximum block size enforcement
    * [core] [\#1807](https://github.com/cosmos/cosmos-sdk/issues/1807) Switch from use of rational to decimal
    * [types] [\#1901](https://github.com/cosmos/cosmos-sdk/issues/1901) Validator interface's GetOwner() renamed to GetOperator()
    * [x/slashing] [#2122](https://github.com/cosmos/cosmos-sdk/pull/2122) - Implement slashing period
    * [types] [\#2119](https://github.com/cosmos/cosmos-sdk/issues/2119) Parsed error messages and ABCI log errors to make     them more human readable.
    * [types] \#2407 MulInt method added to big decimal in order to improve efficiency of slashing
    * [simulation] Rename TestAndRunTx to Operation [#2153](https://github.com/cosmos/cosmos-sdk/pull/2153)
    * [simulation] Remove log and testing.TB from Operation and Invariants, in favor of using errors \#2282
    * [simulation] Remove usage of keys and addrs in the types, in favor of simulation.Account \#2384
    * [tools] Removed gocyclo [#2211](https://github.com/cosmos/cosmos-sdk/issues/2211)
    * [baseapp] Remove `SetTxDecoder` in favor of requiring the decoder be set in baseapp initialization. [#1441](https://github.com/cosmos/cosmos-sdk/issues/1441)
    * [baseapp] [\#1921](https://github.com/cosmos/cosmos-sdk/issues/1921) Add minimumFees field to BaseApp.
    * [store] Change storeInfo within the root multistore to use tmhash instead of ripemd160 \#2308
    * [codec] \#2324 All referrences to wire have been renamed to codec. Additionally, wire.NewCodec is now codec.New().
    * [types] \#2343 Make sdk.Msg have a names field, to facilitate automatic tagging.
    * [baseapp] \#2366 Automatically add action tags to all messages
    * [x/auth] \#2377 auth.StdSignMsg -> txbuilder.StdSignMsg
    * [x/staking] \#2244 staking now holds a consensus-address-index instead of a consensus-pubkey-index
    * [x/staking] \#2236 more distribution hooks for distribution
    * [x/stake] \#2394 Split up UpdateValidator into distinct state transitions applied only in EndBlock
    * [x/slashing] \#2480 Fix signing info handling bugs & faulty slashing

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
  * [gaia-lite] [\#2113](https://github.com/cosmos/cosmos-sdk/issues/2113) Rename `/accounts/{address}/send` to `/bank/accounts/{address}/transfers`, rename `/accounts/{address}` to `/auth/accounts/{address}`

* Gaia CLI  (`gaiacli`)
  * [cli] Cmds to query staking pool and params
  * [gov][cli] #2062 added `--proposal` flag to `submit-proposal` that allows a JSON file containing a proposal to be passed in
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
  * [cli] \#2220 Add `gaiacli config` feature to interactively create CLI config files to reduce the number of required flags
  * [stake][cli] [\#1672](https://github.com/cosmos/cosmos-sdk/issues/1672) Introduced
  new commission flags for validator commands `create-validator` and `edit-validator`.

* Gaia
  * [cli] #2170 added ability to show the node's address via `gaiad tendermint show-address`
  * [simulation] #2313 Reworked `make test_sim_gaia_slow` to `make test_sim_gaia_full`, now simulates from multiple starting seeds in parallel
  * [cli] [\#1921] (https://github.com/cosmos/cosmos-sdk/issues/1921)
    * New configuration file `gaiad.toml` is now created to host Gaia-specific configuration.
    * New --minimum_fees/minimum_fees flag/config option to set a minimum fee.

* SDK
  * [querier] added custom querier functionality, so ABCI query requests can be handled by keepers
  * [simulation] [\#1924](https://github.com/cosmos/cosmos-sdk/issues/1924) allow operations to specify future operations
  * [simulation] [\#1924](https://github.com/cosmos/cosmos-sdk/issues/1924) Add benchmarking capabilities, with makefile commands "test_sim_gaia_benchmark, test_sim_gaia_profile"
  * [simulation] [\#2349](https://github.com/cosmos/cosmos-sdk/issues/2349) Add time-based future scheduled operations to simulator
  * [x/auth] \#2376 Remove FeePayer() from StdTx
  * [x/stake] [\#1672](https://github.com/cosmos/cosmos-sdk/issues/1672) Implement
  basis for the validator commission model.
  * [x/auth] Support account removal in the account mapper.

* Tendermint


IMPROVEMENTS
* [tools] Improved terraform and ansible scripts for infrastructure deployment
* [tools] Added ansible script to enable process core dumps

* Gaia REST API (`gaiacli advanced rest-server`)
    * [x/stake] [\#2000](https://github.com/cosmos/cosmos-sdk/issues/2000) Added tests for new staking endpoints

* Gaia CLI  (`gaiacli`)
    * [cli] #2060 removed `--select` from `block` command
    * [cli] #2128 fixed segfault when exporting directly after `gaiad init`

* Gaia
    * [x/stake] [#2023](https://github.com/cosmos/cosmos-sdk/pull/2023) Terminate iteration loop in `UpdateBondedValidators` and `UpdateBondedValidatorsFull` when the first revoked validator is encountered and perform a sanity check.
    * [x/auth] Signature verification's gas cost now accounts for pubkey type. [#2046](https://github.com/tendermint/tendermint/pull/2046)
    * [x/stake] [x/slashing] Ensure delegation invariants to jailed validators [#1883](https://github.com/cosmos/cosmos-sdk/issues/1883).
    * [x/stake] Improve speed of GetValidator, which was shown to be a performance bottleneck. [#2046](https://github.com/tendermint/tendermint/pull/2200)
    * [x/stake] \#2435 Improve memory efficiency of getting the various store keys
    * [genesis] \#2229 Ensure that there are no duplicate accounts or validators in the genesis state.
    * [genesis] \#2450 Validate staking genesis parameters.
    * Add SDK validation to `config.toml` (namely disabling `create_empty_blocks`) \#1571
    * \#1941(https://github.com/cosmos/cosmos-sdk/issues/1941) Version is now inferred via `git describe --tags`.
    * [x/distribution] \#1671 add distribution types and tests

* SDK
    * [tools] Make get_vendor_deps deletes `.vendor-new` directories, in case scratch files are present.
    * [spec] Added simple piggy bank distribution spec
    * [cli] [\#1632](https://github.com/cosmos/cosmos-sdk/issues/1632) Add integration tests to ensure `basecoind init && basecoind` start sequences run successfully for both `democoin` and `basecoin` examples.
    * [store] Speedup IAVL iteration, and consequently everything that requires IAVL iteration. [#2143](https://github.com/cosmos/cosmos-sdk/issues/2143)
    * [store] \#1952, \#2281 Update IAVL dependency to v0.11.0
    * [simulation] Make timestamps randomized [#2153](https://github.com/cosmos/cosmos-sdk/pull/2153)
    * [simulation] Make logs not just pure strings, speeding it up by a large factor at greater block heights \#2282
    * [simulation] Add a concept of weighting the operations \#2303
    * [simulation] Logs get written to file if large, and also get printed on panics \#2285
    * [simulation] Bank simulations now makes testing auth configurable \#2425
    * [gaiad] \#1992 Add optional flag to `gaiad testnet` to make config directory of daemon (default `gaiad`) and cli (default `gaiacli`) configurable
    * [x/stake] Add stake `Queriers` for Gaia-lite endpoints. This increases the staking endpoints performance by reusing the staking `keeper` logic for queries. [#2249](https://github.com/cosmos/cosmos-sdk/pull/2149)
    * [store] [\#2017](https://github.com/cosmos/cosmos-sdk/issues/2017) Refactor
    gas iterator gas consumption to only consume gas for iterator creation and `Next`
    calls which includes dynamic consumption of value length.
    * [types/decimal] \#2378 - Added truncate functionality to decimal
    * [client] [\#1184](https://github.com/cosmos/cosmos-sdk/issues/1184) Remove unused `client/tx/sign.go`.

* Tendermint

BUG FIXES

* Gaia REST API (`gaiacli advanced rest-server`)

* Gaia CLI  (`gaiacli`)
    * [cli] [\#1997](https://github.com/cosmos/cosmos-sdk/issues/1997) Handle panics gracefully when `gaiacli stake {delegation,unbond}` fail to unmarshal delegation.
    * [cli] [\#2265](https://github.com/cosmos/cosmos-sdk/issues/2265) Fix JSON formatting of the `gaiacli send` command.

* Gaia
  * [x/stake] Return correct Tendermint validator update set on `EndBlocker` by not
  including non previously bonded validators that have zero power. [#2189](https://github.com/cosmos/cosmos-sdk/issues/2189)

* SDK
    * [\#1988](https://github.com/cosmos/cosmos-sdk/issues/1988) Make us compile on OpenBSD (disable ledger) [#1988] (https://github.com/cosmos/cosmos-sdk/issues/1988)
    * [\#2105](https://github.com/cosmos/cosmos-sdk/issues/2105) Fix DB Iterator leak, which may leak a go routine.
    * [ledger] [\#2064](https://github.com/cosmos/cosmos-sdk/issues/2064) Fix inability to sign and send transactions via the LCD by
    loading a Ledger device at runtime.
    * [\#2158](https://github.com/cosmos/cosmos-sdk/issues/2158) Fix non-deterministic ordering of validator iteration when slashing in `gov EndBlocker`
    * [simulation] \#1924 Make simulation stop on SIGTERM
    * [\#2388](https://github.com/cosmos/cosmos-sdk/issues/2388) Remove dependency on deprecated tendermint/tmlibs repository.
    * [\#2416](https://github.com/cosmos/cosmos-sdk/issues/2416) Refactored
    `InitializeTestLCD` to properly include proposing validator in genesis state.

* Tendermint
