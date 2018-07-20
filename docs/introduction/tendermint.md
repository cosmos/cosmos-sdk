# Tendermint

Tendermint is software for securely and consistently replicating an application on many machines. By securely, we mean that Tendermint works even if up to 1/3 of machines fail in arbitrary ways. By consistently, we mean that every non-faulty machine sees the same transaction log and computes the same state. Secure and consistent replication is a fundamental problem in distributed systems; it plays a critical role in the fault tolerance of a broad range of applications, from currencies, to elections, to infrastructure orchestration, and beyond.

Tendermint is designed to be easy-to-use, simple-to-understand, highly performant, and useful for a wide variety of distributed applications.

## Byzantine Fault Tolerance
The ability to tolerate machines failing in arbitrary ways, including becoming malicious, is known as Byzantine fault tolerance (BFT). The theory of BFT is decades old, but software implementations have only became popular recently, due largely to the success of “blockchain technology” like Bitcoin and Ethereum. Blockchain technology is just a re-formalization of BFT in a more modern setting, with emphasis on peer-to-peer networking and cryptographic authentication. The name derives from the way transactions are batched in blocks, where each block contains a cryptographic hash of the previous one, forming a chain. In practice, the blockchain data structure actually optimizes BFT design.

## Application Blockchain Interface
Tendermint consists of two chief technical components: a blockchain consensus engine and a generic application interface. The consensus engine, called Tendermint Core, ensures that the same transactions are recorded on every machine in the same order. The application interface, called the Application Blockchain Interface (ABCI), enables the transactions to be processed in any programming language. Unlike other blockchain and consensus solutions developers can use Tendermint for BFT state machine replication in any programming language or development environment. Visit the [Tendermint docs](https://tendermint.readthedocs.io/projects/tools/en/master/introduction.html#abci-overview) for a deep dive into the ABCI.

The [Cosmos SDK](/sdk/overview.md) is an ABCI framework written in Go. [Lotion JS](/lotion/overview.md) is an ABCI framework written in JavaScript.
