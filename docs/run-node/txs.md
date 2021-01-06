<!--
order: 4
-->

# Generating, Signing and Broadcasting Transactions

This document describes how to generate an (unsigned) transaction, signing it (with one or multiple keys), and broadcasting it to the network. {synopsis}

## Using the CLI

The easiest way to send transactions is using the CLI, as we have seen in the previous page when [interacting with a node](./interact-node.md#using-the-cli). For example, running the following command

```bash
simd tx bank send $MY_VALIDATOR_ADDRESS $RECIPIENT 1000stake --chain-id my-test-chain
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
simd tx bank send $MY_VALIDATOR_ADDRESS $RECIPIENT 1000stake --chain-id my-test-chain --generate-only
```

This will output the unsigned transaction as JSON in the console. We can also save the unsigned transaction to a file (to be passed around between signers more easily) by appending `> unsigned_tx.json` to the above command.

### Signing a Transaction

Signing a transaction using the CLI requires the unsigned transaction to be saved in a file. Let's assume the unsigned transaction is in a file called `unsigned_tx.json` in the current directory (see previous paragraph on how to save a transaction into a file). Then, simply run the following command:

```bash
simd tx sign unsigned_tx.json --chain-id my-test-chain --keyring-backend test --from $MY_VALIDATOR_ADDRESS
```

This command will decode the unsigned transaction and sign it with `SIGN_MODE_DIRECT` with `$MY_VALIDATOR_ADDRESS`'s key, which we already set up in the keyring. The signed transaction will be output as JSON to the console, and, as above, we can save it to a file by appending `> signed_tx.json`.

Some useful flags to consider in the `sign` command:

- `--sign-mode`: you may use `amino-json` to sign the transaction using `SIGN_MODE_LEGACY_AMINO_JSON`,
- `--offline`: sign in offline mode. This means that the `sign` command doesn't connect to the node to retrieve the signer's account number and sequence, both needed for signing. In this case, you must manually supply the `--account-number` and `--sequence` flags. This is useful for offline signing, e.g. signing in a secure environment which doesn't have access to the internet.

#### Signing with Multiple Signers

::: warning
Please note that signing a transaction with multiple signers or with a multisig account, where at least one signer uses `SIGN_MODE_DIRECT`, is not possible as of yet. You may follow [this Github issue](https://github.com/cosmos/cosmos-sdk/issues/8141) for more info.
:::

Signing with multiple signers is done with the `multisign` command. This command assumes that all signers use `SIGN_MODE_LEGACY_AMINO_JSON`. The flow is similar to the `sign` command flow, but instead of signing an unsigned transaction file, each signer signs the file signed by previous signers. The `multisign` command will append signatures to the existing transactions. It is important that signers sign the transaction **in the same order** as given by the transaction, which is retrievable using the `GetSigners()` method.

For example, starting with the `unsigned_tx.json`, and assuming the transaction has 4 signers, we would run:

```bash
# Let signer 1 sign the unsigned tx.
simd tx multisignsign unsigned_tx.json signer_key_1 --chain-id my-test-chain --keyring-backend test > partial_tx_1.json
# Signer 2 appends their signature.
simd tx multisignsign partial_tx_1.json signer_key_2 --chain-id my-test-chain --keyring-backend test > partial_tx_2.json
# Signer 3 appends their signature.
simd tx multisignsign partial_tx_2.json signer_key_3 --chain-id my-test-chain --keyring-backend test > partial_tx_3.json
# Signer 4 appends their signature. The final output is the fully signed tx.
simd tx multisignsign partial_tx_3.json signer_key_4 --chain-id my-test-chain --keyring-backend test > signed_tx.json
```

### Broadcasting a Transaction

Broadcasting a transaction is done simply using the following command:

```bash
simd tx broadcast tx_signed.json
```

You may optionally pass the `--broadcast-mode` flag to specify which response to receive from the node:

- `block`: the CLI waits for the tx to be committed in a block.
- `sync`: the CLI waits for a CheckTx execution response only.
- `async`: the CLI returns immediately.

## Programmatically with Go

It is possible to manipulate transactions programmatically via Go using the Cosmos SDK's `TxBuilder` interface.

### Generating a Transaction

Before generating a transaction, a new instance of a `TxBuilder` needs to be created. Since the SDK supports both Amino and Protobuf transactions, the first step would be to decide which encoding scheme to use. All the subsequent steps remain unchanged, whether you're using Amino or Protobuf, as `TxBuilder` abstracts the encoding mechanisms. In the following snippet, we will use Protobuf.

```go
import (
	"github.com/cosmos/cosmos-sdk/simapp"
)

// Choose your codec: Amino or Protobuf. Here, we use Protobuf, given by the
// following function.
encCfg := simapp.MakeTestEncodingConfig()

// Create a new TxBuilder.
txBuilder := encCfg.TxConfig.NewTxBuilder()
```

We can also set up some keys and addresses that will send and receive the transactions. Here, for the purpose of the tutorial, we will be using some dummy data to create keys.

```go
import (
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
)

priv1, _, addr1 := testdata.KeyTestPubAddr()
priv2, _, addr2 := testdata.KeyTestPubAddr()
priv3, _, addr3 := testdata.KeyTestPubAddr()
```

Populating the `TxBuilder` can be done via its [methods](https://github.com/cosmos/cosmos-sdk/blob/v0.40.0-rc6/client/tx_config.go#L32-L45):

```go
import (
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// Define two x/bank MsgSend messages:
// - from addr1 to addr3,
// - from addr2 to addr3.
// This means that the transactions needs two signers: addr1 and addr2.
msg1 := banktypes.NewMsgSend(addr1, addr3, types.NewCoins(types.NewInt64Coin("atom", 12)))
msg2 := banktypes.NewMsgSend(addr2, addr3, types.NewCoins(types.NewInt64Coin("atom", 34)))

err := txBuilder.SetMsgs(msg1, msg2)
if err != nil {
    return err
}

txBuilder.SetGasLimit(...)
txBuilder.SetFeeAmount(...)
txBuilder.SetMemo(...)
```

At this point, the `TxBuilder` is correctly populated, and the underlying transaction is ready to be signed.

## Using CosmJS (JavaScript & TypeScript)

CosmJS aims to build client libraries in JavaScript that can be embedded in web applications. Please see [https://cosmos.github.io/cosmjs](https://cosmos.github.io/cosmjs) for more information.
