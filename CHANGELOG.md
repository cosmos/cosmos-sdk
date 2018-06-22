# Changelog

## 0.20.0

*TBD*

BREAKING CHANGES
* Change default ports from 466xx to 266xx
* AltBytes renamed to Memo, now a string, max 100 characters, costs a bit of gas
* Transactions now take a list of Messages
* Signers of a transaction now only sign over their account and sequence number

FEATURES
* [gaiacli] You can now attach a simple text-only memo to any transaction, with the `--memo` flag
* [lcd] Queried TXs now include the tx hash to identify each tx
* [mockapp] CompleteSetup() no longer takes a testing parameter
* [governance] Implemented MVP
  * Supported proposal types: just binary (pass/fail) TextProposals for now
  * Proposals need deposits to be votable; deposits are burned if proposal fails
  * Delegators delegate votes to validator by default but can override (for their stake)
* [tools] make get_tools installs tendermint's linter, and gometalinter

FIXES
* \#1259 - fix bug where certain tests that could have a nil pointer in defer
* \#1052 - Make all now works
* Retry on HTTP request failure in CLI tests, add option to retry tests in Makefile
* Fixed bug where chain ID wasn't passed properly in x/bank REST handler

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

* [stake] MarshalJSON -> MarshalBinary
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
