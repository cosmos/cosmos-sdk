## Tendermint Core and ABCI

Cosmos-SDK is a framework to develop the *application* layer of the blockchain. This application can be plugged on any consensus engine (*consensus* + *networking* layers) that supports a simple protocol called the ABCI, short for [Application-Blockchain Interface](https://github.com/tendermint/abci).

Tendermint Core is the default consensus engine on which the Cosmos-SDK is built. It is important to have a good understanding of the respective responsibilities of both the *Application* and the *Consensus Engine*.

Responsibilities of the *Consensus Engine*:
- Propagate transactions
- Agree on the order of valid transactions

Reponsibilities of the *Application*:
- Generate Transactions
- Check if transactions are valid
- Process Transactions (includes state changes)

It is worth underlining that the *Consensus Engine* has knowledge of a given validator set for each block, but that it is the responsiblity of the *Application* to trigger validator set changes. This is the reason why it is possible to build both **public and private chains** with the Cosmos-SDK and Tendermint. A chain will be public or private depending on the rules, defined at application level, that governs a validator's set changes.

The ABCI establishes the connection between the *Consensus Engine* and the *Application*. Essentially, it boils down to two messages:

- `CheckTx`: Ask the application if the transaction is valid. When a validator's node receives a transaction, it will run `CheckTx` on it. If the transaction is valid, it is added to the mempool.
- `DeliverTx`: Ask the application to process the transaction and update the state.

Let us give a high-level overview of  how the *Consensus Engine* and the *Application* interract with each other.

- At all times, when the consensus engine (Tendermint Core) of a validator node receives a transaction, it passes it to the application via `CheckTx` to check its validity. If it is valid, the transaction is added to the mempool.
- Let us say we are at block N. There is a validator set V. A proposer of the next block is selected from V by the *Consensus Engine*. The proposer gathers valid transaction from its mempool to form a new block. Then, the block is gossiped to other validators to be signed/commited. The block becomes block N+1 once 2/3+ of V have signed a *precommit* on it (For a more detailed explanation of the consensus algorithm, click [here](https://github.com/tendermint/tendermint/wiki/Byzantine-Consensus-Algorithm)).
- When block N+1 is signed by 2/3+ of V, it is gossipped to full-nodes. When full-nodes receive the block, they confirm its validity. A block is valid if it it holds valid signatures from more than 2/3 of V and if all the transactions in the block are valid. To check the validity of transactions, the *Consensus Engine* transfers them to the application via `DeliverTx`. After each transaction, `DeliverTx` returns a new state if the transaction was valid. At the end of the block, a final state is committed. Of course, this means that the order of transaction within a block matters.