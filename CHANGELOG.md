# Changelog

## 0.4.0 (March 30, 2017)

BREAKING CHANGES:

- CLI now uses Cobra

IMPROVEMENTS:

- `basecoin init` doesn't generate error if already initialized

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



