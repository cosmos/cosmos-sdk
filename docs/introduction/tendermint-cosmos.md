## Tendermint and Cosmos

Blockchains can be divided into three conceptual layers:

- **Networking:** Responsible for propagating transactions.
- **Consensus:** Enables validator nodes to agree on the next set of transactions to process (i.e. add blocks of transactions to the blockchain).
- **Application:** Responsible for updating the state given a set of transactions, i.e. processing transactions.

The *networking* layer makes sure that each node receives transactions. The *consensus* layer makes sure that each node agrees on the same transactions to modify their local state. As for the *application* layer, it processes transactions. Given a transaction and a state, the application will return a new state. In Bitcoin for example, the application state is a ledger or list of balances for each account (in reality, it's a list of UTXO, short for Unspent Transaction Output, but let's call them balances for the sake of simplicity), and the transactions modify the application's state by changing these list of balances. In the case of Ethereum, the application is a virtual machine. Each transaction goes through this virtual machine and modifies the application state according to the the smart contract that is called within it. 

Before Tendermint, building a blockchain required building all three layers from the ground up. It was such a tedious task that most developers preferred to fork or replicate the Bitcoin codebase, but were constrainted by the limitations of Bitcoin's protocol. The Ethereum Virtual Machine (EVM) was designed to solve this problem and simplify decentralized application development by allowing customizable logic to be executed through smart contracts. But it did not resolve the limitations (interoperability, scalability and customization) of blockchains themselves. Go-Ethereum remains a very monolithic tech stack that is difficult to hard-fork much like Bitcoin's codebase. 

Tendermint was designed to address these issues and provide developers with an laternative. The goal of Tendermint is to provide the *networking* and *consensus* layers of a blockchain as a generic engine to power any application developers want to build. With Tendermint, developers only have to focus on the *application* layer, thereby saving hundreds of hours of work and costly development set-ups. For reference, Tendermint also designates the name of the byzantine fault tolerant consensus algorithm used within the Tendermint Core engine.

Tendermint connects the blockchain engine, Tendermint Core (*networking* and *consensus* layers), to the *application* layer via a socket protocol called the  [ABCI](https://github.com/tendermint/abci), short for Application-BlockChain Interface. Developers only have to implement a few messages to build an ABCI-enabled application that runs on top of the Tendermint Core engine. ABCI is language agnostic, meaning that developers can build the *application* part of their blockchain in any programming language. Building on top of the Tendermint Core engine also provides the following benefits:

- **Public or private blockchain capable.** Developers can deploy any blockchain application, permissioned (private) and permissionless (public), on top of Tendermint Core. 
- **Performance.** Tendermint Core is a state-of-the-art blockchain consensus engine able to handle large number of transactions in short timespan. A block time on Tendermint Core can be as low as one second and can process thousands of transactions in that time period.
- **Instant finality.** A property of the Tendermint consensus algorithm is instant finality, meaning that forks are never created, as long as less than a third of the validators are malicious (byzantine). Users can be sure their transactions are finalized as soon as a block is created.
- **Security.** Tendermint Core's consensus is not only fault tolerant, it’s optimally Byzantine fault-tolerant (BFT), with accountability. If the blockchain forks, there is a way to determine liability.
- **Light-client support**. Tendermint provides built-in light-clients.

But most importantly, Tendermint is natively compatible with the [Inter-Blockchain Communication Protocol](https://github.com/cosmos/cosmos-sdk/tree/develop/docs/spec/ibc) (IBC). This means that any Tendermint-based blockchain, whether public or private, can be natively connected to the Cosmos ecosystem and securely exchange tokens with other blockchains in the ecosystem. Note that benefiting from interoperability via IBC and Cosmos preserves the sovereignty of your Tendermint chain. Non-Tendermint chains can also be connected to Cosmos via IBC adapters or Peg-Zones, but this is out of scope for this document.

For a more detailed overview of the Cosmos ecosystem, you can read [this article](https://blog.cosmos.network/understanding-the-value-proposition-of-cosmos-ecaef63350d).

For more on Tendermint, go [here](tendermint.md)