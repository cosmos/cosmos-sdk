<!--
order: 3
-->

# Interacting with the Node

## Pre-requisite Readings

- [Running a Node](./run-node.md) {prereq}

## Via CLI

Now that your chain is running, it is time to try sending tokens from the first account you created to a second account. In a new terminal window, start by running the following query command:

```bash
simd query account $MY_VALIDATOR_ADDRESS --chain-id my-test-chain
```

You should see the current balance of the account you created, equal to the original balance of `stake` you granted it minus the amount you delegated via the `gentx`. Now, create a second account:

```bash
simd keys add recipient --keyring-backend test

# Put the generated address in a variable for later use.
RECIPIENT=$(simd keys show recipient -a --keyring-backend test)
```

The command above creates a local key-pair that is not yet registered on the chain. An account is created the first time it receives tokens from another account. Now, run the following command to send tokens to the `recipient` account:

```bash
simd tx send $MY_VALIDATOR_ADDRESS $RECIPIENT 1000stake --chain-id my-test-chain

# Check that the recipient account did receive the tokens.
simd query account $RECIPIENT --chain-id my-test-chain
```

Finally, delegate some of the stake tokens sent to the `recipient` account to the validator:

```bash
simd tx staking delegate $(simd keys show my_validator --bech val -a --keyring-backend test) 500stake --from recipient --chain-id my-test-chain

# Query the total delegations to `validator`.
simd query staking delegations-to $(simd keys show my_validator --bech val -a --keyring-backend test) --chain-id my-test-chain
```

You should see two delegations, the first one made from the `gentx`, and the second one you just performed from the `recipient` account.
