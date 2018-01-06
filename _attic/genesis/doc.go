/*
Package genesis provides some utility functions for parsing
a standard genesis file to initialize your abci application.

We wish to support using one genesis file to initialize both
tendermint and the application, so this file format is designed
to be embedable in the tendermint genesis.json file. We reuse
the same chain_id field for tendermint, ignore the other fields,
and add a special app_options field that contains information just
for the abci app (and ignored by tendermint).

The use of this file format for your application is not required by
the sdk and is only used by default in the start command, if you wish
to write your own start command, you can use any other method to
store and parse options for your abci application. The important part is
that the same data is available on every node.

Example file format:

  {
    "chain_id": "foo_bar_chain",
    "app_options": {
      "accounts": [{
        "address": "C471FB670E44D219EE6DF2FC284BE38793ACBCE1",
        "pub_key": {
          "type": "ed25519",
          "data": "6880DB93598E283A67C4D88FC67A8858AA2DE70F713FE94A5109E29C137100C2"
        },
        "coins": [
          {
            "denom": "ETH",
            "amount": 654321
          }
        ]
      }],
      "plugin_options": [
        "plugin1/key1", "value1",
        "profile/set", {"name": "john", age: 37}
      ]
    }
  }

Note that there are two subfields under app_options. The first one "accounts"
is a special case for the coin module, which is assumed to be used by most
applications. It is simply a list of accounts with an identifier and their
initial balance. The account must be identified by EITHER an address
(20 bytes in hex) or a pubkey (in the go-crypto json format), not both as in
this example. "coins" defines the initial balance of the account.

Configuration options for every other module should be placed under
"plugin_options" as key value pairs (there must be an even number of items).
The first value must be "<module>/<key>" to define the option to be set.
The second value is parsed as raw json and is the value to pass to the
application. This may be a string, an array, a map or any other valid json
structure that the module can parse.

Note that we don't use a map for plugin_options, as we will often wish
to have many values for the same key, to run this setup many times,
just as we support setting many accounts.
*/
package genesis
