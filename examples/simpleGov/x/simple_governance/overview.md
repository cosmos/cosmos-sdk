Developing a Tendermint-based blockchain means that you only have to code the application (i.e. the business logic). But that in itself can prove to be rather difficult. This is why the Cosmos-SDK exists.

The Cosmos-SDK is a template framework to build secure blockchain applications on top of Tendermint. It is based on two major principles:

- **Composability:**  The goal of the Cosmos-SDK is to create an ecosystem of modules that allow developers to easily spin up sidechains without having to code every single functionality of their application. Anyone can create a module for the Cosmos-SDK, and using already-built modules in your blockchain is as simple as importing them into your application. For example, the Tendermint team is building a set of basic modules that are needed for the Cosmos Hub, like accounts, staking, IBC, governance. Now if you want to develop a public Tendermint blockchain compatible with Cosmos that has the aforementioned functionalities, you just have to import these already-built modules. As a developer, you only have to create the modules required by your application that do not already exist. As the Cosmos ecosystem develops, we expect the modules ecosystem to gracefully develop, making it easier and easier to develop complex blockchain applications.
- **Capabilities:** Most developers will need to access other modules when building their own modules. The Cosmos-SDK being an open framework, it is likely that some of these modules will be malicious. To address these threats, the Cosmos-SDK is designed to be the foundation of a capabilities-based system. In practice, this means that instead of having each module keep an access control list to give access to other modules, each module implement `mappers` that can be passed to other modules to grant a pre-defined set of capabilities. For example, if an instance of module A's `mapper` is passed to module B, module B will be able to call a restricted set of module A's functions. The *capabilities* of each mapper are defined by the module's developer, and it is the job of the application developer to instanciate and pass mappers from module to module properly. For a deeper look at capabilities, you can read this cool [article](http://habitatchronicles.com/2017/05/what-are-capabilities/)

Now that we have a better understanding of the high level principles of the SDK, let us take a deeper look at how a Cosmos-SDK application is constructed.

*Note: For now the Cosmos-SDK only exists in Golang, which means that module developers can only develop SDK modules in Golang. In the future, we expect that Cosmos-SDK in other programming languages will pop up*

### Application architecture

The Cosmos-SDK gives the basic template for your application architecture. You can find this template [here](https://github.com/cosmos/cosmos-sdk).

Before we start analyzing the different directories, let us remind some basic concepts about blockchain applications. In essence, a blockchain application is simply a replicated state machine. There is a state (e.g. for a cryptocurrency, how many coins each account holds), and transactions that trigger state transitions. As the application developer, what you do is just define the state, the transactions types and how different transactions modify the state. And that's exactly what modules do. They define stores, functions to interact with stores (`keeper`), transaction types and functions to handle these transactions (`handlers`). The Cosmos-SDK exists so that it is easy to:

- Code modules
- Integrate modules into a coherent blockchain application
- Run the blockchain application

With this in mind, let us go through the important directories of the SDK:

- `baseapp`: This defines the template for a basic application. Basically it implements the ABCI protocol so that your Cosmos-SDK application can communicate with the underlying Tendermint node.
- `client`: Command-Line to interface with the application
- `server`: Rest server to communicate with the node
- `examples`: Contains example on how to build a working application based on `baseapp` and modules
- `store`: Contains code for the multistore. The multistore is basically your state. Each module can create any number of KVStores from the multistore. Be careful to properly handle access rights to each store with appropriate `keepers`.
- `types`: Common types required in any SDK-based application.
- `x`: This is where modules live. You will find all the already-built modules in this directory. To use any of these modules, you just need to properly import them in your application. We will see how in the [App - Bridging it all together] section.

So by now you should have realized how easy it is to build a Tendermint blockchain on top of the Cosmos-SDK. You just have to follow these simple steps:

1. Clone the Cosmos-SDK repo
2. Code the modules needed by your application that do not already exist
3. Create your app directory. In the app main file, import the module you need and instantiate the different stores.
4. Launch your blockchain.

Easy as pie! With the introduction over, let us delve into practice and learn how to code a SDK module.