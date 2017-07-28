# Changelog

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
- basecoin
  - `basecoin start` supports all flags that `tendermint node` does, such as
  `--rpc.laddr`, `--p2p.seeds`, and `--p2p.skip_upnp`
  - fully supports `--log_level` and `--trace` for logger configuration
  - merkleeyes no longers spams the logs... unless you want it
    - Example: `basecoin start --log_level="merkleeyes:info,state:info,*:error"`
    - Example: `basecoin start --log_level="merkleeyes:debug,state:info,*:error"`
- basecli
  - `basecli init` is more intelligent and only complains if there really was
  a connected chain, not just random files
  - support `localhost:46657` or `http://localhost:46657` format for nodes,
  not just `tcp://localhost:46657`
  - Add `--genesis` to init to specify chain-id and validator hash
    - Example: `basecli init --node=localhost:46657 --genesis=$HOME/.basecoin/genesis.json`
  - `basecli rpc` has a number of methods to easily accept tendermint rpc, and verifies what it can

BUG FIXES:
- basecli
  - `basecli query account` accepts hex account address with or without `0x`
  prefix
  - gives error message when running commands on an unitialized chain, rather
  than some unintelligable panic


## 0.6.0 (June 22, 2017)

Make the basecli command the only way to use client-side, to enforce best
security practices. Lots of enhancements to get it up to production quality.

BREAKING CHANGES:
- ./cmd/commands -> ./cmd/basecoin/commands
- basecli
  - `basecli proof state get` -> `basecli query key`
  - `basecli proof tx get` -> `basecli query tx`
  - `basecli proof state get --app=account` -> `basecli query account`
  - use `--chain-id` not `--chainid` for consistency
  - update to use `--trace` not `--debug` for stack traces on errors
  - complete overhaul on how tx and query subcommands are added. (see counter or trackomatron for examples)
  - no longer supports counter app (see new countercli)
- basecoin
  - `basecoin init` takes an argument, an address to allocate funds to in the genesis
  - removed key2.json
  - removed all client side functionality from it (use basecli now for proofs)
    - no tx subcommand
    - no query subcommand
    - no account (query) subcommand
    - a few other random ones...
  - enhanced relay subcommand
    - relay start did what relay used to do
    - relay init registers both chains on one another (to set it up so relay start just works)
- docs
  - removed `example-plugin`, put `counter` inside `docs/guide`
- app
  - Implements ABCI handshake by proxying merkleeyes.Info()

IMPROVEMENTS:
- `basecoin init` support `--chain-id`
- intergrates tendermint 0.10.0 (not the rc-2, but the real thing)
- commands return error code (1) on failure for easier script testing
- add `reset_all` to basecli, and never delete keys on `init`
- new shutil based unit tests, with better coverage of the cli actions
- just `make fresh` when things are getting stale ;)

BUG FIXES:
- app: no longer panics on missing app_options in genesis (thanks, anton)
- docs: updated all docs... again
- ibc: fix panic on getting BlockID from commit without 100% precommits (still a TODO)

## 0.5.2 (June 2, 2017)

BUG FIXES:
- fix parsing of the log level from Tendermint config (#97)

## 0.5.1 (May 30, 2017)

BUG FIXES:
- fix ibc demo app to use proper tendermint flags, 0.10.0-rc2 compatibility
- Make sure all cli uses new json.Marshal not wire.JSONBytes

## 0.5.0 (May 27, 2017)

BREAKING CHANGES:
- only those related to the tendermint 0.9 -> 0.10 upgrade

IMPROVEMENTS:
- basecoin cli
  - integrates tendermint 0.10.0 and unifies cli (init, unsafe_reset_all, ...)
  - integrate viper, all command line flags can also be defined in environmental variables or config.toml
- genesis file
  - you can define accounts with either address or pub_key
  - sorts coins for you, so no silent errors if not in alphabetical order
- [light-client](https://github.com/tendermint/light-client) integration
  - no longer must you trust the node you connect to, prove everything!
  - new [basecli command](./cmd/basecli/README.md)
  - integrated [key management](https://github.com/tendermint/go-crypto/blob/master/cmd/README.md), stored encrypted locally
  - tracks validator set changes and proves everything from one initial validator seed
  - `basecli proof state` gets complete proofs for any abci state
  - `basecli proof tx` gets complete proof where a tx was stored in the chain
  - `basecli proxy` exposes tendermint rpc, but only passes through results after doing complete verification

BUG FIXES:
- no more silently ignored error with invalid coin names (eg. "17.22foo coin" used to parse as "17 foo", not warning/error)


## 0.4.1 (April 26, 2017)

BUG FIXES:

- Fix bug in `basecoin unsafe_reset_X` where the `priv_validator.json` was not being reset

## 0.4.0 (April 21, 2017)

BREAKING CHANGES:

- CLI now uses Cobra, which forced changes to some of the flag names and orderings

IMPROVEMENTS:

- `basecoin init` doesn't generate error if already initialized
- Much more testing

## 0.3.1 (March 23, 2017)

IMPROVEMENTS:

- CLI returns exit code 1 and logs error before exiting

## 0.3.0 (March 23, 2017)

BREAKING CHANGES:

- Remove `--data` flag and use `BCHOME` to set the home directory (defaults to `~/.basecoin`)
- Remove `--in-proc` flag and start Tendermint in-process by default (expect Tendermint files in $BCHOME/tendermint).
To start just the ABCI app/server, use `basecoin start --without-tendermint`.
- Consolidate genesis files so the Basecoin genesis is an object under `app_options` in Tendermint genesis. For instance:

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

- Introduce `basecoin init` and `basecoin unsafe_reset_all`

## 0.2.0 (March 6, 2017)

BREAKING CHANGES:

- Update to ABCI v0.4.0 and Tendermint v0.9.0
- Coins are specified on the CLI as `Xcoin`, eg. `5gold`
- `Cost` is now `Fee`

FEATURES:

- CLI for sending transactions and querying the state,
designed to be easily extensible as plugins are implemented
- Run Basecoin in-process with Tendermint
- Add `/account` path in Query
- IBC plugin for InterBlockchain Communication
- Demo script of IBC between two chains

IMPROVEMENTS:

- Use new Tendermint `/commit` endpoint for crafting IBC transactions
- More unit tests
- Use go-crypto S structs and go-data for more standard JSON
- Demo uses fewer sleeps

BUG FIXES:

- Various little fixes in coin arithmetic
- More commit validation in IBC
- Return results from transactions

## PreHistory

##### January 14-18, 2017

- Update to Tendermint v0.8.0
- Cleanup a bit and release blog post

##### September 22, 2016

- Basecoin compiles again



