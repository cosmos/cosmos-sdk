<!--
order: 3
-->

# Interacting with the Node

## Pre-requisite Readings

- [Running a Node](./run-node.md) {prereq}

## Via CLI

Now that your chain is running, it is time to try sending tokens from the first account you created to a second account. In a new terminal window, start by running the following query command:

```bash
simd query account $(simd keys show my_validator -a) --chain-id my-test-chain
```

You should see the current balance of the account you created, equal to the original balance of `stake` you granted it minus the amount you delegated via the `gentx`. Now, create a second account:

```bash
simd keys add receiver
```

The command above creates a local key-pair that is not yet registered on the chain. An account is registered the first time it receives tokens from another account. Now, run the following command to send tokens to the second account:

```bash
simd tx send $(simd keys show my_validator -a) $(simd keys show receiver -a) 1000stake --chain-id my-test-chain
```

Check that the second account did receive the tokens:

```bash
simd query account $(simd keys show receiver -a) --chain-id my-test-chain
```

Finally, delegate some of the stake tokens sent to the `receiver` account to the validator:

```bash
simd tx staking delegate $(simd keys show my_validator --bech val -a) 500stake --from receiver --chain-id my-test-chain
```

Try to query the total delegations to `validator`:

```bash
simd query staking delegations-to $(simd keys show my_validator --bech val -a) --chain-id my-test-chain
```

You should see two delegations, the first one made from the `gentx`, and the second one you just performed from the `receiver` account.
