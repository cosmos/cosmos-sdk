# Tendermint

Tendermint is software for securely and consistently replicating an application on many machines. By securely, we mean that Tendermint works even if up to 1/3 of machines fail in arbitrary ways. By consistently, we mean that every non-faulty machine sees the same transaction log and computes the same state. Secure and consistent replication is a fundamental problem in distributed systems; it plays a critical role in the fault tolerance of a broad range of applications, from currencies, to elections, to infrastructure orchestration, and beyond.

Tendermint is designed to be easy-to-use, simple-to-understand, highly performant, and useful for a wide variety of distributed applications.

## Byzantine Fault Tolerance

The ability to tolerate machines failing in arbitrary ways, including becoming malicious, is known as Byzantine fault tolerance (BFT). The theory of BFT is decades old, but software implementations have only became popular recently, due largely to the success of “blockchain technology” like Bitcoin and Ethereum. Blockchain technology is just a re-formalization of BFT in a more modern setting, with emphasis on peer-to-peer networking and cryptographic authentication. The name derives from the way transactions are batched in blocks, where each block contains a cryptographic hash of the previous one, forming a chain. In practice, the blockchain data structure actually optimizes BFT design.

## Application Blockchain Interface

Tendermint consists of two chief technical components: a blockchain consensus engine and a generic application interface. The consensus engine, called Tendermint Core, ensures that the same transactions are recorded on every machine in the same order. The application interface, called the Application Blockchain Interface (ABCI), enables the transactions to be processed in any programming language. Unlike other blockchain and consensus solutions developers can use Tendermint for BFT state machine replication in any programming language or development environment. Visit the [Tendermint docs](https://tendermint.readthedocs.io/projects/tools/en/master/introduction.html#abci-overview) for a deep dive into the ABCI.

## Understanding the roles of the different layers

It is important to have a good understanding of the respective responsibilities of both the *Application* and the *Consensus Engine*.

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

## Application frameworks

Even if Tendermint makes it easy for developers to build their own blockchain by enabling them to focus on the *Application* layer of their blockchain, building an *Application* can be a challenging task in itself. This is why *Application Frameworks* exist. They provide developers with a secure and features-heavy environment to develop Tendermint-based applications. Here are some examples of *Application Frameworks* :

- The [Cosmos SDK](/sdk/overview.md) is an ABCI framework written in Go. 
- [Lotion JS](/lotion/overview.md) is an ABCI framework written in JavaScript.

