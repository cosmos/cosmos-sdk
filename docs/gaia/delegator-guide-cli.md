# Delegator Guide (CLI)

This document contains all the necessary information for delegators to interract with the Cosmos Hub through the Command-Line Interface (CLI).

It also contains instructions on how to manage accounts, restore accounts from the fundraiser and use a ledger nano device.

## Table of contents

- [Installing `gaiacli`](#installing-gaiacli)
- [Cosmos Accounts](#cosmos-accounts)
    + [Restoring an account from the fundrasier](#restoring-an-account-from-the-fundraiser)
    + [Creating an account](#creating-an-account)
- [Accessing the Cosmos Hub network](#accessing-the-cosmos-hub-network)
    + [Running your own full-node](#running-your-own-full-node)
    + [Running a light-node](#running-a-light-node)
    + [Connecting to a remote full-node](#connecting-to-a-remote-full-node)
- [Setting up `gaiacli`](#setting-up-gaiacli)
- [Querying the state](querying-the-state)
- [Bonding Atoms and Withdrawing rewards](bonding-atoms-and-withdrawing-rewards)
- [Participating in Governance](participating-in-governance)
- [Signing transactions from an offline computer](#signing-transactions-from-an-offline-computer)

## Installing `gaiacli` 

`gaiacli`: This is the command-line interface to interract with a `gaiad` full-node. 

[**Download the binaries**]

[**Install from source**](https://cosmos.network/docs/gaia/installation.html)

## Cosmos Accounts

At the core of every Cosmos Account, there is a seed, which takes the form of a 12 or 24-words mnemonic. From this mnemonic, it is possible to create any number of Cosmos accounts, i.e. pairs of private key/public key. 

```
     Account 0                         Account 1                         Account 2

+------------------+              +------------------+               +------------------+
|                  |              |                  |               |                  |
|    Address 0     |              |    Address 1     |               |    Address 2     |
|        ^         |              |        ^         |               |        ^         |
|        |         |              |        |         |               |        |         |
|        |         |              |        |         |               |        |         |
|        |         |              |        |         |               |        |         |
|        +         |              |        +         |               |        +         |
|  Public key 0    |              |  Public key 1    |               |  Public key 2    |
|        ^         |              |        ^         |               |        ^         |
|        |         |              |        |         |               |        |         |
|        |         |              |        |         |               |        |         |
|        |         |              |        |         |               |        |         |
|        +         |              |        +         |               |        +         |
|  Private key 0   |              |  Private key 1   |               |  Private key 2   |
|        ^         |              |        ^         |               |        ^         |
+------------------+              +------------------+               +------------------+
         |                                 |                                  |
         |                                 |                                  |
         |                                 |                                  |
         +--------------------------------------------------------------------+
                                           |
                                           |
                                 +---------+---------+
                                 |                   |
                                 |  Mnemonic (Seed)  |
                                 |                   |
                                 +-------------------+
```

The funds stored in an account are controlled by the private key. This private key is generated using a one-way function from the mnemonic. If you loose the private key, you can retrieve it using the mnemonic. However, if you loose the mnemonic, you loose access to all the derived private keys. Likewise, if someone gains access to your mnemonic, they gain access to all the associated accounts. 

::: danger
**NEVER LOOSE YOUR MNEMONIC. WRITE IT DOWN ON A PIECE OF PAPER AND STORE IT SOMEWHERE SAFE. IF YOU LOOSE IT, THERE IS NO WAY TO RETRIEVE IT. IF SOMEONE GAINS ACCESS TO IT, THEY GAIN ACCESS TO ALL THE ASSOCIATED ACCOUNTS.**
:::

The address is a human-readble public information (e.g. `cosmos10snjt8dmpr5my0h76xj48ty80uzwhraqalu4eg`) that identifies your account. When someone wants to send you fund, they send it to your address. It is not possible to find the private key associated with a given address. 

### Restoring an account from the fundraiser

::: tip
*NOTE: This section only concerns fundraiser participants*
:::

If you participated in the fundraiser, you should be in possession of a 12-words mnemonic. Newly generated mnemonic use 24 words, but 12-words mnemonic are also compatible with all the Cosmos tools. 

#### On a ledger device

At the core of a ledger device, there is a mnemonic used to generate accounts on multiple blockchains (including the Cosmos Hub). Usually, you will create a new mnemonic when you initialize your ledger device. However, it is possible to tell the ledger device to use a mnemonic provided by the user instead. Let us go ahead and see how you can input the mnemonic you obtained during the fundraiser as the seed of your ledger device. 

::: warning
*NOTE: To do this, **it is preferrable to use a brand new ledger device.**. Indeed, there can be only one mnemonic per ledger device. If however you want to use a ledger  that is already initalized with a seed, you can reset it by going in `Settings`>`Device`>`Reset All`. **Please note that this will wipe out the seed currently stored on the device. If you have not properly secured the associated mnemonic, you could loose your funds!!!***
:::

The following steps need to be performed on an un-initialized ledger device:

1. Connect your ledger device to the computer via USB
2. Press both buttons
3. Do **NOT** choose the "Config as a new device" option. Instead, choose "Restore Configuration"
4. Choose a PIN
5. Choose the 12 words option
6. Input each of the words you got during the fundraiser, in the correct order. 

Your ledger is now correctly set up with your fundraiser mnemonic! Do not loose this mnemonic! If your ledger is compromised, you can always restore a new device again using the same mnemonic.

Next, click [here](#using-a-ledger-device) to generate an account. 

#### On a computer

::: warning
**NOTE: It is more secure to perform this action on an offline computer**
::: 

To restore an account using a fundraiser mnemonic and store the associated encrypted private key on a computer, use the following command:

```bash
gaiacli keys add <yourKeyName> --recover
```

You will be prompted to input a passphrase that is used to encrypt the private key of account `0` on disk. Each time you want to send a transaction, this password will be required. If you loose the password, you can always recover the private key with the mnemonic. 

- `<yourKeyName>` is the name of the account. It is a reference to the account number used to derive the key pair from the mnemonic. You will use this name to identify your account when you want to send a transaction.
- You can add the optional `--account` flag to specify the path (`0`, `1`, `2`, ...) you want to use to generate your account. By default, account `0` is generated. 

### Creating an account

To create an account, you just need to have `gaiacli` installed. Before creating it, you need to know where you intend to store and interract with your private keys. The best options are to store them in an offline dedicated computer or a ledger device. Storing them on your regular online computer involves more risk.

#### Using a ledger device

When you initialize your ledger, a 24-words mnemonic is generated and stored in the device. This mnemonic is compatible with Cosmos and Cosmos Accounts can be derived from it. Therefore, all you have to do is make your ledger compatible with `gaiacli`. To do so, you need to go through the following steps:

1. Download the Ledger Live app [here](https://www.ledger.com/pages/ledger-live)
2. Connect your ledger via USB and update to the latest firmware
3. Go to the ledger live app store, and download the "Cosmos" application (this can take a while)
4. Navigate to the Cosmos app on your ledger device

Then, to create an account, use the following command:

```bash
gaiacli keys add <yourAccountName> --ledger 
```

- `<yourKeyName>` is the name of the account. It is a reference to the account number used to derive the key pair from the mnemonic. You will use this name to identify your account when you want to send a transaction.
- You can add the optional `--account` flag to specify the path (`0`, `1`, `2`, ...) you want to use to generate your account. By default, account `0` is generated. 

#### Using a computer 

::: warning
**NOTE: It is more secure to perform this action on an offline computer**
:::

To generate an account, just use the following command:

```bash
gaiacli keys add <yourKeyName>
```

The command will generate a 24-words mnemonic and account `0` at the same time. You will be prompted to input a passphrase that is used to encrypt the private key of account `0` on disk. Each time you want to send a transaction, this password will be required. If you loose the password, you can always recover the private key with the mnemonic. 

::: danger
**NEVER LOOSE YOUR MNEMONIC. WRITE IT DOWN ON A PIECE OF PAPER AND STORE IT SOMEWHERE SAFE. IF YOU LOOSE IT, THERE IS NO WAY TO RETRIEVE IT. IF SOMEONE GAINS ACCESS TO IT, THEY GAIN ACCESS TO ALL THE ASSOCIATED ACCOUNTS**
::: 

- `<yourKeyName>` is the name of the account. It is a reference to the account number used to derive the key pair from the mnemonic. You will use this name to identify your account when you want to send a transaction.
- You can add the optional `--account` flag to specify the path (`0`, `1`, `2`, ...) you want to use to generate your account. By default, account `0` is generated. 


You can generate more accounts from the same mnemonic using the following command:

```bash
gaiacli keys add <yourKeyName> --recover --account 1
```

This command will prompt you to input a passphrase as well as your mnemonic. Change the account number to generate a different account. 


## Accessing the Cosmos Hub network

In order to query the state and send transactions, you need a way to access the network. To do so, you can either run your own full-node, or connect to someone else's.

::: warning
**NOTE: Never give your mnemonic, private key or passphrase to anyone. If a full-node operator requests it, they are trying to abuse you**.
::: 

### Running your own full-node

This is the most secure option, but comes with relatively high resources requirement. In order to run your own full-node, you need a good bandwidth and at least 1TB of disk space. 

You will find the tutorial to install `gaiad` [here](https://cosmos.network/docs/gaia/installation.html), and the guide to run a full-node [here](https://cosmos.network/docs/gaia/join-testnet.html).

// TODO: Add root-of-trust and note about long-range-attacks when implemented

### Running a light-node

// TODO when functionality is implemented

### Connecting to a remote full-node

If you do not want or cannot run your own node, you can connect to someone else's full-node. You should pick an operator you trust, because a malicious operator could return  incorrect query results or censor your transactions. However, they will never be able to steal your funds, as your private keys are stored locally on your computer or ledger device. Possible options of full-node operators include validators, wallet providers or exchanges. 

In order to connect to the full-node, you will need an address of the following form: `https://77.87.106.33:26657` (*Note: This is a placeholder*). This address has to be communicated by the full-node operator you choose to trust. You will use this address in the [following section](#setting-up-gaiacli).

## Setting up `gaiacli`

`gaiacli` is the tool that enables you to interract with the node that runs on the Cosmos Hub network, wether you run it yourself or not (see [accessing the cosmos hub network](#accession-the-cosmos-hub-network)). Let us set it up properly.

In order to set up `gaiacli`, use the following command:

```bash
gaiacli config <flag> <value>
```

It allows you to set a default value for each given flag. 

First, set up the address of the full-node you want to connect to:

```bash
gaiacli config node <host>:<port

// example: gaiacli config node https://77.87.106.33:26657
```

If you run your own full-node, just use `tcp://localhost:26657` as the address. 

Then, let us set the default value of the `--trust-node` flag:

```bash
gaiacli config trust-node false

// Set to true if you run a light-client node, false otherwise
```

Finally, let us set the `chain-id` of the blockchain we want to interract with:

```bash
gaiacli config chain-id gos-3
```

## Querying the state

`gaiacli` lets you query all relevant information from the blockchain, like account balances, amount of bonded tokens, outsanding rewards, governance proposals and more. Next is a list of the most useful commands for delegator. Please make sure you [set up gaiacli](#setting-up-gaiacli) before trying them.

```bash
// query account balances and other account-related information
gaiacli query account

// query the list of validators
gaiacli query validators

// query the information of a validator given their address (e.g. cosmos10snjt8dmpr5my0h76xj48ty80uzwhraqalu4eg)
gaiacli query validator <validatorAddress>

// query all delegations made from a delegator given their address (e.g. cosmos10snjt8dmpr5my0h76xj48ty80uzwhraqalu4eg)
gaiacli query delegations <delegatorAddress>

// query a specific delegation made from a delegator to a validator given their addresses
gaiacli query delegations <delegatorAddress> <validatorAddress>

// query the rewards of a delegator given a delegator address (e.g. cosmos10snjt8dmpr5my0h76xj48ty80uzwhraqalu4eg)
gaiacli query distr rewards <delegatorAddress> 

// query all proposals currently open for depositing
gaiacli query proposals --status deposit_period

// query all proposals currently open for voting
gaiacli query proposals --status voting_period

// query a proposal given its proposalID
gaiacli query proposal <proposalID>
```

For more commands, just type:

```bash
gaiacli query
```

For each command, you can use the `--h` flag to get more information.

## Bonding Atoms and Withdrawing rewards

::: warning
**Before bonding Atoms, please read the [delegator faq](https://cosmos.network/resources/delegators) to understand the risk and responsabilities involved with delegating**
:::

::: warning
**Note: These commands need to be run on an online computer. It is more secure to perform them commands using a ledger device. For the offline procedure, click [here](#signing-transactions-from-an-offline-computer).**
::: 

```bash
// Bond Atoms 
// ex value for flags: <amountToBound>=10000stake, <bech32AddressOfValidator>=cosmosvaloper18thamkhnj9wz8pa4nhnp9rldprgant57pk2m8s, <gasPrice>=0.001stake

gaiacli tx staking --amount <amountToBond> --validator <bech32AddressOfValidator> --from <delegatorKeyName> --gas auto --gas-prices <gasPrice>

// Withdraw rewards

gaiacli tx distr withdraw-rewards --from <delegatorKeyName>
```

To confirm that your transaction went through, you can use the following queries:

```bash
// your balance should change after you bond Atoms or withdraw rewards
gaiacli query account

// you should have delegations after you bond Atom
gaiacli query delegations <delegatorAddress>
```

Double check with a block explorer if you interract with the network through a trusted full-node. 

### A note on gas and fees

Transactions on the Cosmos Hub network need to include a transaction fee in order to be processed. This fee pays for the gas required to run the transaction. The formula is the following:

```
fees = gas * gasPrices
```

The `gas` is dependent on the transaction. Different transaction require different amount of `gas`. The `gas` amount for a transaction is calculated as it is being processed, but there is a way to estimate it beforehand by using the `auto` value for the `gas` flag. Of course, this only gives an estimate. You can adjust this estimate with the flag `--gas-adjustment` (default `1.0`) if you want to be sure you provide enough `gas` for the transaction. 

The `gasPrice` is the price of each unit of `gas`. Each validator sets a `min-gas-price` value, and will only include transaction that have a `gasPrice` superior to their `min-gas-price`. 

The transaction `fees` is the product of `gas` and `gasPrice`. As a user, you have to input 2 out of 3. The higher the `gasPrice`, the higher the chance that your transaction will get included in a block. 

## Participating in governance

### Primer on governance

The Cosmos Hub has a built-in governance system that lets bonded Atom holder vote on proposals. There are three types of proposal:

- `Text Proposals`: These are the most basic type of proposals. They can be used to get the opinion of the network on a given topic. 
- `Parameter Proposals`: These are used to update the value of an existing parameter.
- `Software Upgrade Proposal`: These are used to propose an upgrade of the Hub's software.

Any Atom holder can submit a proposal. In order for the proposal to be open for voting, it needs to come with a `deposit` that is superior to a parameter called `minDeposit`. The `deposit` need not be provided in its entirety by the submitter. If the initial proposer's `deposit` is not sufficient, the proposal enters the `deposit_period` status. Then, any Atom holder can increase the deposit by sending a `depositTx`. 

Once the `deposit` reaches `minDeposit`, the proposal enters the `voting_period`, which lasts 2 weeks. Any **bonded** Atom hodler can then cast a vote on this proposal. The options are `Yes`, `No`, `NoWithVeto` and `Abstain`. The weight of the vote is based on the amount of bonded Atoms of the sender. If they don't vote, delegator inherit the vote of their validator. However, delegators can override their validator's vote by sending a vote themselves. 

At the end of the voting period, the proposal is accepted if there is more than 50% `Yes` votes (excluding `Abstain ` votes) and less than 33.33% of `NoWithVeto` votes (excluding `Abstain` votes).

### In practice

::: warning
**Note: These commands need to be run on an online computer. It is more secure to perform them commands using a ledger device. For the offline procedure, click [here](#signing-transactions-from-an-offline-computer).**
::: 

```bash
// Submit a Proposal
// <type>=text/parameter_change/software_upgrade
// ex value for flag: <gasPrice>=0.0001stake

gaiacli tx gov submit-proposal --title "Test Proposal" --description "My awesome proposal" --type <type> --deposit=10stake --gas auto --gas-prices <gasPrice> --from <delegatorKeyName>

// Increase deposit of a proposal
// Retrieve proposalID from $gaiacli query gov proposals --status deposit_period
// ex value for parameter: <deposit>=1stake

gaiacli tx gov deposit <proposalID> <deposit> --gas auto --gas-prices <gasPrice> --from <delegatorKeyName>

// Vote on a proposal
// Retrieve proposalID from $gaiacli query gov proposals --status voting_period 
// <option>=yes/no/no_with_veto/abstain

gaiacli tx gov vote <proposalID> <option> --gas auto --gas-prices <gasPrice> --from <delegatorKeyName>
```

## Signing transactions from an offline computer

If you do not have a ledger device and want to interract with your private key on an offline computer, you can use the following procedure. First, generate an unsigned transaction on an **online computer** with the following command (example with a bonding transaction):

```bash
// Bond Atoms 
// ex value for flags: <amountToBound>=10000stake, <bech32AddressOfValidator>=cosmosvaloper18thamkhnj9wz8pa4nhnp9rldprgant57pk2m8s, <gasPrice>=0.001stake

gaiacli tx staking --amount <amountToBond> --validator <bech32AddressOfValidator> --gas auto --gas-prices <gasPrice> --generate-only > unsignedTX.json
```

Then, copy `unsignedTx.json` and transfer it (e.g. via USB) to the offline computer. If it is not done already, [create an account on the offline computer](#using-a-computer). Now, sign the transaction using the following command:

```bash
gaiacli tx sign unsignedTx.json --from <delegatorKeyName> > signedTx.json
```

Copy `signedTx.json` and transfer it back to the online computer. Finally, use the following command to broadcast the transaction:

```bash
gaiacli tx broadcast signedTx.json
```
