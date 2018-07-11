# Ledger // Cosmos

### Ledger Support for account keys

`gaiacli` now supports derivation of account keys from a Ledger seed. To use this functionality you will need the following:

- A running `gaiad` instance connected to the network you wish to use.
- A `gaiacli` instance configured to connect to your chosen `gaiad` instance.
- A LedgerNano with the `ledger-cosmos` app installed
  * Install the Cosmos app onto your Ledger by following the instructions in the [`ledger-cosmos`](https://github.com/cosmos/ledger-cosmos/blob/master/docs/BUILD.md) repository.
  * A production-ready version of this app will soon be included in the [Ledger Apps Store](https://www.ledgerwallet.com/apps)

> **NOTE:** Cosmos keys are derived acording to the [BIP 44 Hierarchical Deterministic wallet spec](https://github.com/bitcoin/bips/blob/master/bip-0044.mediawiki). For more information on Cosmos derivation paths [see the hd package](https://github.com/cosmos/cosmos-sdk/blob/develop/crypto/keys/hd/hdpath.go#L30).

Once you have the Cosmos app installed on your Ledger, and the Ledger is accessible from the machine you are using `gaiacli` from you can create a new account key using the Ledger:

```bash
$ gaiacli keys add {{ .Key.Name }} --ledger
NAME:	          TYPE:	  ADDRESS:						                                  PUBKEY:
{{ .Key.Name }}	ledger	cosmosaccaddr1aw64xxr80lwqqdk8u2xhlrkxqaxamkr3e2g943	cosmosaccpub1addwnpepqvhs678gh9aqrjc2tg2vezw86csnvgzqq530ujkunt5tkuc7lhjkz5mj629
```

This key will only be accessible while the Ledger is plugged in and unlocked. To send some coins with this key, run the following:

```bash
$ gaiacli send --from {{ .Key.Name }} --to {{ .Destination.AccAddr }} --chain-id=gaia-7000
```

You will be asked to review and confirm the transaction on the Ledger. Once you do this you should see the result in the console! Now you can use your Ledger to manage your Atoms and Stake!
