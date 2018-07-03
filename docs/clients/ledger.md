# Ledger // Cosmos

### Ledger Support for Account Keys

`gaiacli` now supports derivation of account keys from a Ledger seed. To use this functionality you will need the following:

- A running `gaiad` instance connected to the network you wish to use.
- A `gaiacli` instance configured to connect to your chosen `gaiad` instance.
- A LedgerNano with the `ledger-cosmos` app installed
  * Install the Cosmos app onto your Ledger by following the instructions in the [`ledger-cosmos`](https://github.com/cosmos/ledger-cosmos/blob/master/docs/BUILD.md) repository.
  * A production-ready version of this app will soon be included in the [Ledger Apps Store](https://www.ledgerwallet.com/apps)

Once you have the Cosmos app installed on your Ledger, and the ledger is accessible from the machine you are using `gaiacli` from you can create a new Account key using the ledger:

```bash
$ gaiacli keys add {{ .Key.Name }} --ledger
NAME:	          TYPE:	  ADDRESS:						                                  PUBKEY:
{{ .Key.Name }}	ledger	cosmosaccaddr1aw64xxr80lwqqdk8u2xhlrkxqaxamkr3e2g943	cosmosaccpub1addwnpepqvhs678gh9aqrjc2tg2vezw86csnvgzqq530ujkunt5tkuc7lhjkz5mj629
```

This key will only be accessible while the ledger is plugged in and unlocked. To send some coins with this key, run the following:

```bash
$ gaiacli send --name {{ .Key.Name }} --to {{ .Destination.AccAddr }} --chain-id=gaia-7000
```

You will be asked to review and confirm the transaction on the Ledger. Once you do this you should see the result in the console! Now you can use your Ledger to manage your Atoms and Stake!
