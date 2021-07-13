<!--
order: 7
-->

# Running a Testnet

The `simd testnet` subcommand makes it easy to initialize and start a simulated test network for testing purposes. {synopsis}

In addition to the commands for [running a node](/run-node/run-node.html), the `simd` binary also includes a `testnet` command that allows you to start a simulated test network in-process or to initialize files for a simulated test network that runs in a separate process. 

## Initialize Files

First, let's take a look at the `init-files` subcommand.

This is similar to the `init` command when initializing a single node, but in this case we are initializing multiple nodes, generating the genesis transactions for each node, and then collecting those transactions.

The `init-files` subcommand initializes a test network that will run in a separate process. It is not a prerequisite for the `start` subcommand ([see below](#start-testnet)).

In order to initialize the files for a test network, run the following command:

```bash
simd testnet init-files
```

You should see the following output in your terminal:

```bash
Successfully initialized 4 node directories
```

The default output directory is a relative `.testnets` directory. Let's take a look at the files created within the `.testnets` directory.

### gentxs

Within the `.testnets` directory, there is a `gentxs` directory that includes a genesis transaction for each validator node. The default number of nodes is `4`, so you should see four files that include JSON encoded genesis transactions.

### nodes

Within the `.testnets` directory, there is a `node` directory for each node. Within each `node` directory, there is a `simd` directory. The `simd` directory is the home directory for each node (i.e. each `simd` directory includes the same configuration and data files as the default `~/.simapp` directory when running a single node).

## Start Testnet

Now, let's take a look at the `start` subcommand.

The `start` subcommand both initializes and starts an in-process test network. This is the fastest way to spin up a local test network for testing purposes.

You can start the local test network by running the following command:

```bash
simd testnet start
```

You should see something similar to the following:

```bash
acquiring test network lock
preparing test network with chain-id "chain-mtoD9v"


+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
++       THIS MNEMONIC IS FOR TESTING PURPOSES ONLY        ++
++                DO NOT USE IN PRODUCTION                 ++
++                                                         ++
++  sustain know debris minute gate hybrid stereo custom   ++
++  divorce cross spoon machine latin vibrant term oblige  ++
++   moment beauty laundry repeat grab game bronze truly   ++
+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++


starting test network...
started test network
press the Enter Key to terminate
```

The test network is running in-process, which means the test network will terminate once you either close the terminal window or you press the Enter key.

In order to test the test network, use the `simd` binary. The first node uses the same default addresses when starting a single node network.

Check the status of the first node:

```
simd status
```

Import the key from the provided mnemonic:

```
simd keys add test --recover --keyring-backend test
```

Check the balance of the account address:

```
simd q bank balances [address]
```

Use this test account to execute transactions.

## Testnet Options

You can customize the configuration of the testnet with flags. In order to see all flag options, append the `--help` flag to each command.

One of the flag options is the `--v` flag. This flag allows you to set the number of validators that the test network will have (i.e. the number of nodes to initialize and genesis transactions to generate). The default value for the `--v` flag is `4`, which is why four node directories were initialized in the example above.