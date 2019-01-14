# What is Gaia?

`gaia` is the name of the Cosmos SDK application for the Cosmos Hub. It comes with 2 main entrypoints:

- `gaiad`: The Gaia Daemon, runs a full-node of the `gaia` application.
- `gaiacli`: The Gaia command-line interface, which enables interraction with a Gaia full-node.

`gaia` is built on the Cosmos SDK using the following modules:

- `x/auth`: Accounts and signatures.
- `x/bank`: Token transfers.
- `x/staking`: Staking logic.
- `x/mint`: Inflation logic.
- `x/distribution`: Fee distribution logic.
- `x/slashing`: Slashing logic.
- `x/gov`: Governance logic.
- `x/ibc`: Inter-blockchain transfers.
- `x/params`: Handles app-level parameters.

>About the Cosmos Hub: The Cosmos Hub is the first Hub to be launched in the Cosmos Network. The role of a Hub is to facilitate transfers between blockchains. If a blockchain connects to a Hub via IBC, it automatically gains access to all the other blockchains that are connected to it. The Cosmos Hub is a public Proof-of-Stake chain. Its staking token is called the Atom.

Next, learn how to [install Gaia](./installation.md).
