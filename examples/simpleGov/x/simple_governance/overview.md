Developing a Tendermint-based blockchain means that you only have to code the application (i.e. the business logic). But that in itself can prove to be rather difficult. This is why the Cosmos-SDK exists.

The Cosmos-SDK is a template framework to build secure blockchain applications on top of Tendermint. It is based on two major principles:

- **Composability:**  The goal of the Cosmos-SDK is to create an ecosystem of modules that allow developers to easily spin up sidechains without having to code every single functionality of their application. Anyone can create a module for the Cosmos-SDK, and using already-built modules in your blockchain is as simple as importing them into your application. For example, the Tendermint team is building a set of basic modules that are needed for the Cosmos Hub, like accounts, staking, IBC, governance. Now if you want to develop a public Tendermint blockchain compatible with Cosmos that has the aforementioned functionalities, you just have to import these already-built modules. As a developer, you only have to create the modules required by your application that do not already exist. As the Cosmos ecosystem develops, we expect the modules ecosystem to gracefully develop, making it easier and easier to develop complex blockchain applications.
- **Capabilities:** Most developers will need to access other modules when building their own modules. The Cosmos-SDK being an open framework, it is likely that some of these modules will be malicious. To address these threats, the Cosmos-SDK is designed to be the foundation of a capabilities-based system. In practice, this means that instead of having each module keep an access control list to give access to other modules, each module implement `mappers` that can be passed to other modules to grant a pre-defined set of capabilities. For example, if an instance of module A's `mapper` is passed to module B, module B will be able to call a restricted set of module A's functions. The *capabilities* of each mapper are defined by the module's developer, and it is the job of the application developer to instanciate and pass mappers from module to module properly. For a deeper look at capabilities, you can read this cool [article](http://habitatchronicles.com/2017/05/what-are-capabilities/)

Now that we have a better understanding of the high level principles of the SDK, let us take a deeper look at how a Cosmos-SDK application is constructed.

*Note: For now the Cosmos-SDK only exists in Golang, which means that module developers can only develop SDK modules in Golang. In the future, we expect that Cosmos-SDK in other programming languages will pop up*

### Reminder on Tendermint and ABCI

Todo

### Application architecture

The Cosmos-SDK gives the basic template for your application architecture. You can find this template [here](https://github.com/cosmos/cosmos-sdk).

In essence, a blockchain application is simply a replicated state machine. There is a state (e.g. for a cryptocurrency, how many coins each account holds), and transactions that trigger state transitions. As the application developer, your job is just define the state, the transactions types and how different transactions modify the state. 

#### Modularity

The Cosmos-SDK is a module-based framework. Each module is in itself a little state-machine that can be gracefully combined with other modules to produce a coherent application. In other words, modules define a sub-section of the global state and of the transaction types. Then, it is the job of the root application to route messages to the correct modules depending on their respective types. To understand this process, let us take a look at a simplified standard cycle of the state-machine.

Upon receiving a transaction from the Tendermint Core engine, here is whatthe application does:

1. Decode the transaction and get the message
2. Route the message to the appropriate module using the `Msg.Type()` method
3. Run the transaction in the module. Modify the state if the transaction is valid.
4. Return new state or error message

Steps 1, 2 and 4 are handled by the root application. Step 3 is handled by the appropriate module. 

#### SDK Components 

With this in mind, let us go through the important directories of the SDK:

- `baseapp`: This defines the template for a basic application. Basically it implements the ABCI protocol so that your Cosmos-SDK application can communicate with the underlying Tendermint node.
- `client`: Command-Line to interface with the application
- `server`: Rest server to communicate with the node
- `examples`: Contains example on how to build a working application based on `baseapp` and modules
- `store`: Contains code for the multistore. The multistore is basically your state. Each module can create any number of KVStores from the multistore. Be careful to properly handle access rights to each store with appropriate `keepers`.
- `types`: Common types required in any SDK-based application.
- `x`: This is where modules live. You will find all the already-built modules in this directory. To use any of these modules, you just need to properly import them in your application. We will see how in the [App - Bridging it all together] section.

#### Introductory Coderun

##### KVStore

The KVStore provides the basic persistence layer for your SDK application.

https://github.com/cosmos/cosmos-sdk/blob/3fc7200f1d1045a19efc30395e5916f9ef1b42b7/types/store.go#L91-L121

You can mount multiple KVStores onto your application, e.g. one for staking, one for accounts, one for IBC, and so on.

https://github.com/cosmos/cosmos-sdk/blob/3fc7200f1d1045a19efc30395e5916f9ef1b42b7/examples/basecoin/app/app.go#L90

The implementation of a KVStore is responsible for providing any Merkle proofs for each query, if requested.

https://github.com/cosmos/cosmos-sdk/blob/3fc7200f1d1045a19efc30395e5916f9ef1b42b7/store/iavlstore.go#L135

Stores can be cache-wrapped to provide transactions at the persistence level
(and this is well supported for iterators as well). This feature is used to
provide a layer of transactional isolation for transaction processing after the
"AnteHandler" deducts any associated fees for the transaction.  Cache-wrapping
can also be useful when implementing a virtual-machine or scripting environment
for the blockchain.

##### go-amino

The Cosmos-SDK uses
[go-amino](https://github.com/cosmos/cosmos-sdk/blob/96451b55fff107511a65bf930b81fb12bed133a1/examples/basecoin/app/app.go#L97-L111)
extensively to serialize and deserialize Go types into Protobuf3 compatible
bytes.

Go-amino (e.g. over https://github.com/golang/protobuf) uses reflection to
encode/decode any Go object.  This lets the SDK developer focus on defining
data structures in Go without the need to maintain a separate schema for
Proto3.  In addition, Amino extends Proto3 with native support for interfaces
and concrete types.

For example, the Cosmos SDK's `x/auth` package imports the PubKey interface
from `tendermint/go-crypto` , where PubKey implementations include those for
Ed25519 and Secp256k1.  Each auth.BaseAccount has a PubKey.

https://github.com/cosmos/cosmos-sdk/blob/d309abb4b9bdb01272e54e048063502110c801fa/x/auth/account.go#L35-L43

Amino knows what concrete type to decode for each interface value
based on what concretes are registered for the interface.

For example, the "Basecoin" example app knows about Ed25519 and Secp256k1 keys
because they are registered by the app's codec below:

https://github.com/cosmos/cosmos-sdk/blob/d309abb4b9bdb01272e54e048063502110c801fa/examples/basecoin/app/app.go#L101

For more information on Go-Amino, see https://github.com/tendermint/go-amino.

##### Keys, Keepers, and Mappers

The Cosmos SDK is designed to enable an ecosystem of libraries that can be
imported together to form a whole application.  To make this ecosystem
more secure, we've developed a design pattern that follows the principle of 
least-authority.

Mappers and Keepers provide access to KV stores via the context.  The only
difference between the two is that a Mapper provides a lower-level API, so
generally a Keeper might hold references to other Keepers and Mappers but not
vice versa.

Mappers and Keepers don't hold any references to any stores directly.  They only
hold a "key" (the `sdk.StoreKey` below):

https://github.com/cosmos/cosmos-sdk/blob/fc0e4013278d41fab4f3ac73f28a42bc45889106/x/auth/mapper.go#L14-L24

This way, you can hook everything up in your main app.go file and see what
components have access to what stores and other components.

https://github.com/cosmos/cosmos-sdk/blob/d309abb4b9bdb01272e54e048063502110c801fa/examples/basecoin/app/app.go#L65-L70

Later during the execution of a transaction (e.g. via ABCI DeliverTx after a
block commit) the context is passed in as the first argument.  The context
includes references to any relevant KV stores, but you can only access them if
you hold the associated key.

https://github.com/cosmos/cosmos-sdk/blob/fc0e4013278d41fab4f3ac73f28a42bc45889106/x/auth/mapper.go#L44-L53

Mappers and Keepers cannot hold direct references to stores because the store
is not known at app initialization time.  The store is dynamically created (and
wrapped via CacheKVStore as needed to provide a transactional context) for
every committed transaction (via ABCI DeliverTx) and mempool check transaction
(via ABCI CheckTx). 

##### Tx, Msg, Handler, and AnteHandler

A transaction (Tx interface) is a signed/authenticated message (Msg interface).

Transactions that are discovered by the Tendermint mempool are processed by the
AnteHandler ("ante" just means "before") where the validity of the transaction
is checked and any fees are collected.

Transactions that get committed in a block first get processed through the
AnteHandler, and if the transaction is valid after fees are deducted, they are
processed through the appropriate Handler.

In either case, the transaction bytes must first be parsed.  The default
transaction parser uses Amino.  Most SDK developers will want to use the
standard transaction structure defined in the `x/auth` package (and the
corresponding AnteHandler implementation also provided in `x/auth`):

https://github.com/cosmos/cosmos-sdk/blob/fc0e4013278d41fab4f3ac73f28a42bc45889106/x/auth/stdtx.go#L12-L18

Various packages generally define their own message types.  The Basecoin
example app includes multiple message types that are registered in app.go:

https://github.com/cosmos/cosmos-sdk/blob/d309abb4b9bdb01272e54e048063502110c801fa/examples/basecoin/app/app.go#L102-L106

Finally, handlers are added to the router in your app.go file to map messages
to their corresponding handlers. (In the future we will provide more routing
features to enable pattern matching for more flexibility).

https://github.com/cosmos/cosmos-sdk/blob/d309abb4b9bdb01272e54e048063502110c801fa/examples/basecoin/app/app.go#L78-L83

##### EndBlocker

The EndBlocker hook allows us to register callback logic to be performed at the
end of each block.  This lets us process background events, such as processing
validator inflationary atom provisions:

https://github.com/cosmos/cosmos-sdk/blob/3fc7200f1d1045a19efc30395e5916f9ef1b42b7/x/stake/handler.go#L32-L37

By the way, the SDK provides a staking module, which provides all the
bonding/unbonding funcionality for the Cosmos Hub:
https://github.com/cosmos/cosmos-sdk/tree/develop/x/stake (staking module)

#### Start working

So by now you should have realized how easy it is to build a Tendermint blockchain on top of the Cosmos-SDK. You just have to follow these simple steps:

1. Clone the Cosmos-SDK repo
2. Code the modules needed by your application that do not already exist
3. Create your app directory. In the app main file, import the module you need and instantiate the different stores.
4. Launch your blockchain.

Easy as pie! With the introduction over, let us delve into practice and learn how to code a SDK module.