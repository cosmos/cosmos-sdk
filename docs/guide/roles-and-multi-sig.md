This guide uses the roles functionality provided by `basecli` to create a multi-sig wallet. It builds upon the basecoin basics and key management guides. You should have `basecoin` started with blocks streaming in, and three accounts: `rich, poor, igor` where `rich` was the account used on `basecoin init`, _and_ run `basecli init` with the appropriate flags. Review the intro guides for more information.

In this example, `rich` will create the role and send it some coins (i.e., fill the multi-sig wallet). Then, `poor` will prepare a transaction to withdraw coins, which will be approved by `igor`. Let's look at our keys:

```
basecli keys list
```

```
All keys:
igor		5E4CB7A4E729BA0A8B18DE99E21409B6D706D0F1
poor		65D406E028319289A0706E294F3B764F44EBA3CF
rich		CB76F4092D1B13475272B36585EBD15D22A2848D
```

Using the `basecli query account` command, you'll see that `rich` has plenty of coins:

```
{
  "height": 81,
  "data": {
    "coins": [
      {
        "denom": "mycoin",
        "amount": 9007199254740992
      }
    ],
    "credit": []
  }
}
```

whereas `poor` and `igor` have no coins (in fact, the chain doesn't know about them yet):

```
ERROR: Account bytes are empty for address 65D406E028319289A0706E294F3B764F44EBA3CF
```

## Create Role

This first step defines the parameters of a new role, which will have control of any coins sent to it, and only release them if correct conditions are met. In this example, we are going to make a 2/3 multi-sig wallet. Let's look a the command and dissect it below:

```
basecli tx create-role --role="10CAFE4E" --min-sigs=2 --members=5E4CB7A4E729BA0A8B18DE99E21409B6D706D0F1,65D406E028319289A0706E294F3B764F44EBA3CF,CB76F4092D1B13475272B36585EBD15D22A2848D --sequence=1 --name=rich
```

In the first part we are sending a transaction that creates a role, rather than transfering coins. The `--role` flag is the name of the role (in hex only) and must be in double quotes. The `--min-sigs` and `--members` define your multi-sig parameters. Here, we require a minimum of 2 signatures out of 3 members but we could easily say 3 of 5 or 9 of 10, or whatever your application requires. The `--members` flag requires a comma-seperated list of addresses that will be signatories on the role. Then we set the `--sequence` number for the transaction, which will start at 1 and must be incremented by 1 for every transaction from an account. Finally, we use the name of the key/account that will be used to create the role, in this case the account `rich`. 

Remember that `rich`'s address was used on `basecoin init` and is included in the `--members` list. The command above will prompt for a password (which can also be piped into the command if desired) then - if executed correctly - return some data:

```
{
  "check_tx": {
    "code": 0,
    "data": "",
    "log": ""
  },
  "deliver_tx": {
    "code": 0,
    "data": "",
    "log": ""
  },
  "hash": "4849DA762E19CE599460B9882DD42C7F19655DC1",
  "height": 321
}
```
showing the block height at which the transaction was committed and its hash. A quick review of what we did: 1) created a role, essentially an account, that requires a minimum of two (2) signatures from three (3) accounts (members). And since it was the account named `rich`'s first transaction, the sequence was set to 1.

Let's look at the balance of the role that we've created:

```
basecli query account role:"10CAFE4E"
```

and it should be empty:

```
ERROR: Account bytes are empty for address role:10CAFE4E
```

Next, we want to send coins _to_ that role. Notice that because this is the second transaction being sent by rich, we need to increase `--sequence` to `2`: 

```
basecli tx send --fee=90mycoin --amount=10000mycoin --to=role:"10CAFE4E" --sequence=2 --name=rich
```

We need to pay a transaction fee to the validators, in this case 90 `mycoin` to send 10000 `mycoin` Notice that for the `--to` flag, to specify that we are sending to a role instead of an account, the `role:` prefix is added before the role. Because it's `rich`'s second transaction, we've incremented the sequence. The output will be nearly identical to the output from `create-role` above.

Now the role has coins (think of it like a bank). 

Double check with:

```
basecli query account role:"10CAFE4E"
```

and this time you'll see the coins in the role's account:

```
{
  "height": 2453,
  "data": {
    "coins": [
      {
        "denom": "mycoin",
        "amount": 10000
      }
    ],
    "credit": []
  }
}
```

`Poor` decides to initiate a multi-sig transaction to himself from the role's account. First, it must be prepared like so:

```
basecli tx send --amount=6000mycoin --from=role:"10CAFE4E" --to=65D406E028319289A0706E294F3B764F44EBA3CF --sequence=1 --assume-role="10CAFE4E" --name=poor --multi --prepare=tx.json
```

you'll be prompted for `poor`'s password and there won't be any `stdout` to the terminal. Note that the address in the `--to` flag matches the address of `poor`'s account from the beginning of the tutorial. The main output is the `tx.json` file that has just been created. In the above command, the `--assume-role` flag is used to (not clear, since we have --from), while the `--multi` flag is used in combination with `--prepare`, to specify the file that is prepared for a multi-sig transaction.

The `tx.json` file will look like this:

```
{
  "type": "sigs/multi",
  "data": {
    "tx": {
      "type": "chain/tx",
      "data": {
        "chain_id": "test_chain_id",
        "expires_at": 0,
        "tx": {
          "type": "nonce",
          "data": {
            "sequence": 1,
            "signers": [
              {
                "chain": "",
                "app": "sigs",
                "addr": "65D406E028319289A0706E294F3B764F44EBA3CF"
              }
            ],
            "tx": {
              "type": "role/assume",
              "data": {
                "role": "10CAFE4E",
                "tx": {
                  "type": "coin/send",
                  "data": {
                    "inputs": [
                      {
                        "address": {
                          "chain": "",
                          "app": "role",
                          "addr": "10CAFE4E"
                        },
                        "coins": [
                          {
                            "denom": "mycoin",
                            "amount": 6000
                          }
                        ]
                      }
                    ],
                    "outputs": [
                      {
                        "address": {
                          "chain": "",
                          "app": "sigs",
                          "addr": "65D406E028319289A0706E294F3B764F44EBA3CF"
                        },
                        "coins": [
                          {
                            "denom": "mycoin",
                            "amount": 6000
                          }
                        ]
                      }
                    ]
                  }
                }
              }
            }
          }
        }
      }
    },
    "signatures": [
      {
        "Sig": {
          "type": "ed25519",
          "data": "A38F73BF2D109015E4B0B6782C84875292D5FAA75F0E3362C9BD29B16CB15D57FDF0553205E7A33C740319397A434B7C31CBB10BE7F8270C9984C5567D2DC002"
        },
        "Pubkey": {
          "type": "ed25519",
          "data": "6ED38C7453148DD90DFC41D9339CE45BEFA5EB505FD7E93D85E71DFFDAFD9B8F"
        }
      }
    ]
  }
}
```

and it is loaded by the next command.

With the transaction prepared, but not sent, we'll have `igor` sign and send the prepared transaction:

```
basecli tx --in=tx.json --name=igor
```

which will give output similar to:

```
{
  "check_tx": {
    "code": 0,
    "data": "",
    "log": ""
  },
  "deliver_tx": {
    "code": 0,
    "data": "",
    "log": ""
  },
  "hash": "E345BDDED9517EB2CAAF5E30AFF3AB38A1172833",
  "height": 2673
}
```

and voila! That's the basics for creating roles and sending multi-sig transactions. For 3 of 3, you'd repeat the last step for rich as well (recall that poor initiated the multi-sig transaction). We can check the balance of the role:

```
basecli query account role:"10CAFE4E"
```

and get the result:

```
{
  "height": 2683,
  "data": {
    "coins": [
      {
        "denom": "mycoin",
        "amount": 4000
      }
    ],
    "credit": []
  }
}
```

and see that `poor` now has 6000 `mycoin`:

```
basecli query account 65D406E028319289A0706E294F3B764F44EBA3CF
```

to confirm that everything worked as expected.
