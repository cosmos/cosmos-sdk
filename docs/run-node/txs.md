<!--
order: 4
-->

# Generating, Signing and Broadcasting Transactions

This document describes how to generate an (unsigned) transaction, signing it (with one or multiple keys), and broadcasting it to the network. {synopsis}

## Using the CLI

The easiest way to send transactions is using the CLI, as we have seen in the previous page when [interacting with a node](./interact-node.md#using-the-cli). For example, running the following command

```bash
simd tx send $MY_VALIDATOR_ADDRESS $RECIPIENT 1000stake --chain-id my-test-chain
```

will run the following steps:

- generate a transaction with one `Msg` (`x/bank`'s `MsgSend`), and print the generated transaction to the console.
- ask the user for confirmation to send the transaction from the `$MY_VALIDATOR_ADDRESS` account.
- fetch `$MY_VALIDATOR_ADDRESS` in the keyring. This is possible because we have [set up the CLI's keyring](./keyring.md) in a previous step.
- sign the generated transaction with the keyring's account.
- broadcast the signed transaction to the network. This is possible because the CLI connects to the node's Tendermint RPC endpoint.

The CLI bundles all the necessary steps into a simple-to-use user experience. However, it's possible to run all the steps individually too.

### Generating a Transaction

Generating a transaction can simply be done by appending the `--generate-only` flag on any `tx` command, e.g.:

```bash
simd tx send $MY_VALIDATOR_ADDRESS $RECIPIENT 1000stake --chain-id my-test-chain --generate-only
```

This will output the unsigned transaction as JSON in the console.

### Signing a Transaction
