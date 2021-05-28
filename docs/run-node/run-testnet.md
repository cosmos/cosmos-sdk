<!--
order: 7
-->

# Running a Testnet

The `simd testing` subcommand makes it easy to initialize and start test networks. {synopsis}

In addition to the commands for [running a node](/run-node/run-node.html), the `simd` binary also includes a `testnet` command that allows you to start an in-process simulated test network or to initialize files for a simulated test network that runs in a separate process. 

The `testnet` command is used within the Cosmos SDK for testing purposes. The command can also be implemented in the CLI for other blockchain applications built with the Cosmos SDK to achieve the same functionality.

## Initialize Files

First, let's take a look at the `init-files` subcommand.

This is similar to the `init` command when initializing a single node, but in this case we are initializing multiple nodes, generating the genesis transactions for each node, and then collecting those transactions.

The `init-files` subcommand initializes a test network that will run in a separate process. It is not a prerequisite for the `start` subcommand ([see below](#start-testnet)).

In order to initialize the files, run the following:

```bash
simd testnet init-files
```

You should see the following output in your terminal:

```bash
Successfully initialized 4 node directories
```

The default output directory is a relative `.testnets` directory. Let's take a look at the files created within the `.testnets` directory.

### gentxs

Within the `.testnets` directory, you will find a `gentxs` directory that includes a genesis transaction for each node. The default number of nodes is `4`, so you should see four `.json` files that include JSON encoded genesis transactions.

### nodes

Within the `.testnets` directory, you will also find a `node` directory for each node. This includes the configuration and data for each node.

Within each `node` directory, you will find a `simd` directory. The `simd` directory for each node includes the same configuration and data files that you would find in the `~/.simapp` directory when running the `simd` command for a single node.

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

The test network is running in-process, which means the test network will terminate once you either close the terminal window or press the Enter key.

## Custom Testnets

You can customize the testnet by adding flags. In order to see the flag options, you can append the `--help` flag to each command. For example, to see the flag options for the `init-files` subcommand, you can run the following:

```bash
simd testnet init-files --help
```

One of the flag options is the `--v` flag. This flag allows you to set the number of validators that the test network will have (.i.e the number of nodes to initialize and genesis transactions to generate). The default value for the `--v` flag is `4`, which is why four node directories were initialized in the example above.

You can also add the `--help` flag to the `start` command:

```bash
simd testnet start --help
```

A few of the flag options are `--api.address`, `--grpc.address`, and `--rpc.address`, which allow you to set the addresses for the API, gRPC, and RPC respectively.
