---
sidebar_position: 1
---

# Guide to Multisig transactions

## Overview

Multisignature accounts are accounts that are generated from multiple public keys. A multisig necessitates that any transaction made on its behalf must be signed by a specified threshold of its members.

A common use case for multisigs is to increase security of a signing account, and/or enable multiple parties to agree on and authorize a transaction.

The first step is to create a multisig signing key by using the public keys of all possible signers and the minimum threshold of addresses that are needed to sign any transaction from the account. The threshold can be the same amount as the total number of addresses comprising the multisig.

Whatever machine is generating the multisig, it should at least have all of the public keys imported into the same keyring.

When you want to create a multisig transaction, you would create the transaction as normal, but instead of signing it with a single account's private key, you would need to sign it with the private keys of the accounts that make up the multisig key.

This is done by signing the transaction multiple times, once with each private key. The order of the signatures matters and must match the order of the public keys in the multisig key.

Once you have a transaction with the necessary signatures, it can be broadcasted to the network. The network will verify that the transaction has the necessary signatures from the accounts in the multisig key before it is executed.

## Step by step guide to multisig transactions

This tutorial will use the test keyring which will store the keys in the default home directory `~/.simapp` unless otherwise specified.
Verify which keys are available in the test keyring by running `--keyring-backend test`.

Prior to this tutorial set the keyring backend to "test" in `~/.simapp/client.toml` to always the test keyring which will specify a consistent keyring for the entirety of the tutorial. Additionally, set the default keyring by running `simd config set client keyring-backend test`.

```shell
simd keys list
```

If you don't already have accounts listed create the accounts using the below.

```shell
simd keys add alice
simd keys add bob
simd keys add recipient
```

Alternatively the public keys comprising the multisig can be imported into the keyring.

```shell
simd keys add alice --pubkey <public key> --keyring backend test
```
    
Create the multisig account between bob and alice.

```shell
simd keys add alice-bob-multisig --multisig alice,bob --multisig-threshold 2
```
    
Before generating any transaction, verify the balance of each account and note the amount. This step is crucial to confirm that the transaction can be processed successfully.

```shell
simd query bank balances my_validator
simd query bank balances alice-bob-multisig
```

Ensure that the alice-bob-multisig account is funded with a sufficient balance to complete the transaction (gas included). In our case, the genesis account, my_validator, holds our funds. Therefore, we will transfer funds from the `my_validator` account to the `alice-bob-multisig` account.
Fund the multisig by sending it `stake` from the genesis account.

```shell
 simd tx bank send my_validator alice-bob-multisig "10000stake"
```
    
Check both accounts again to see if the funds have transferred.

```shell
simd query bank balances alice-bob-multisig
```

Initiate the transaction. This command will create a transaction from the multisignature account `alice-bob-multisig` to send 1000stake to the recipient account. The transaction will be generated but not broadcasted yet.

```shell
simd tx bank send alice-bob-multisig recipient 1000stake --generate-only --chain-id my-test-chain  > tx.json
```

Alice signs the transaction using their key and refers to the multisig address. Execute the command below to accomplish this:

```shell
simd tx sign --from alice --multisig=cosmos1re6mg24kvzjzmwmly3dqrqzdkruxwvctw8wwds tx.json --chain-id my-test-chain > tx-signed-alice.json
```
    
Let's repeat for Bob.

```shell
simd tx sign --from bob --multisig=cosmos1re6mg24kvzjzmwmly3dqrqzdkruxwvctw8wwds tx.json --chain-id my-test-chain > tx-signed-bob.json
```

Execute a multisign transaction by using the `simd tx multisign` command. This command requires the names and signed transactions of all the participants in the multisig account. Here, Alice and Bob are the participants:

```shell
simd tx multisign tx.json alice-bob-multisig tx-signed-alice.json tx-signed-bob.json --chain-id my-test-chain > tx-signed.json
```

Once the multisigned transaction is generated, it needs to be broadcasted to the network. This is done using the simd tx broadcast command:
    
```shell
simd tx broadcast tx-signed.json --chain-id my-test-chain --gas auto --fees 250stake
```

Once the transaction is broadcasted, it's a good practice to verify if the transaction was successful. You can query the recipient's account balance again to confirm if the funds were indeed transferred:

```shell
simd query bank balances alice-bob-multisig
```
