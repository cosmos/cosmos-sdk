# BaseApp

## Pre-requisite Reading

- [Anatomy of an SDK application](../basics/app-anatomy.md)
- [Lifecycle of an SDK transaction](../basics/tx-lifecycle.md)

## Synopsis

This document describes `baseapp`, the abstraction that implements most of the common functionalities of an SDK application. 

- [Introduction](#introduction)
- [Type Definition](#type-definition)
- [Constructor](#constructor)
- [States](#states)
- [Routing](#routing)
- [Main ABCI Messages](#abci)
    + [CheckTx](#checktx)
    + [DeliverTx](#delivertx)
- [RunTx(), AnteHandler and RunMsgs](#runtx()-,antehandler-and-runmsgs())
    + [RunTx()](#runtx())
    + [AnteHandler](#antehandler)
    + [RunMsgs()](#runmsgs())
- [Other ABCI Message](#other-abci-message)
    + [InitChain](#initchain)
    + [BeginBlock](#beginblock)
    + [EndBlock](#endblock)
    + [Commit](#commit)
    + [Info](#info)
    + [Query](#query)
    
## Introduction

`baseapp` is a base class that implements the core of an SDK application, namely:

- The [Application-Blockchain Interface](#abci), for the state-machine to communicate with the underlying consensus engine (e.g. Tendermint). 
- A [Router](#routing), to route [messages](./tx-msgs.md) and [queries](./querier.md) to the appropriate [module](../building-modules/modules.md).
- Different [states](#states), as the state-machine can have different parallel states updated based on the ABCI message received. 

The goal of `baseapp` is to provide the fundamental layer of an SDK application that developers can easily extend to build their own custom application. Usually, developers will create a custom type for their application, like so:

```go 
type app struct {
    *bam.BaseApp // reference to baseapp
    cdc *codec.Codec

    // list of application store keys

    // list of application keepers

    // module manager
}
```

Extending the application with `baseapp` gives the former access to all of `baseapp`'s methods. This allows developers to compose their custom application with the modules they want, while not having to concern themselves with the hard work of implementing the ABCI, the routing and state management logic. 

## Type Definition

The [`baseapp` type](https://github.com/cosmos/cosmos-sdk/blob/master/baseapp/baseapp.go#L45-L91) holds many important parameters for any Cosmos SDK based application. Let us go through the most important components.

*Note: Not all parameters are described, only the most important ones. Refer to the [type definition](https://github.com/cosmos/cosmos-sdk/blob/master/baseapp/baseapp.go#L45-L91) for the full list* 

First, the important parameters that are initialized during the initialization of the application:

- [`CommitMultiStore`](./store.md#commit-multi-store): This is the main store of the application, which holds the canonical state that is committed at the [end of each block](#commit). This store is **not** cached, meaning it is not used to update the application's intermediate (un-committed) states. The `CommitMultiStore` is a multi-store, meaning a store of stores. Each module of the application uses one or multiple `KVStores` in the multi-store to persist their subset of the state. 
- [Database](./store.md#database): The `db` is used by the `CommitMultiStore` to handle data storage.
- [Router](#message-routing): The `router` facilitates the routing of [`messages`](./tx-msgs.md) to the appropriate module for it to be processed. Here `message` refers to the transaction components that need to be processed by the application in order to update the state, and not to ABCI messages which implement the interface between the application and the underlying consensus engine. 
- [Query Router](#query-routing): The `query router` facilitates the routing of [queries](./querier.md) to the appropriate module for it to be processed. These `queries` are not ABCI messages themselves, but they are relayed to the application from the underlying consensus engine via the ABCI message [`Query`](#query).
- [`TxDecoder`](https://godoc.org/github.com/cosmos/cosmos-sdk/types#TxDecoder): It is used to decode transaction `[]byte` relayed by the underlying Tendermint engine.
- [`BaseKey`]: This key is used to access the [main store](./store.md#main-store) in the `CommitMultiStore`. The main store is used to persist data related to the core of the application, like consensus parameters.  
- [`AnteHandler`](#antehandler): This handler is used to handle signature verification and fee payment when a transaction is received.
- [`initChainer`](../basics/app-anatomy.md#initchainer), [`beginBlocker` and `endBlocker`](../basics/app-anatomy.md#beginblocker-and-endblocker): These  are the functions executed when the application receives the [InitChain], [BeginBlock] and [EndBlock] ABCI messages from the underlying Tendermint engine. 

Then, parameters used to define [volatile states](#volatile-states) (i.e. cached states):

- `checkState`: This state is updated during [`CheckTx`](#checktx), and reset on [`Commit`](#commit).
- `deliverState`: This state is updated during [`DeliverTx`](#delivertx), and reset on [`Commit`](#commit).

Finally, a few more important parameters:

- `voteInfos`: This parameter carries the list of validators whose precommit is missing, either because they did not vote or because the proposer did not include their vote. This information is carried by the [context](#context) and can be used by the application for various things like punishing absent validators.
- `minGasPrices`: This parameter defines the minimum [gas prices](./accounts-fees-gas.md#gas) accepted by the node. This is a local parameter, meaning each full-node can set a different `minGasPrices`. It is run by the [`anteHandler`](./accounts-fees-gas.md#antehandler) during `CheckTx`, mainly as a spam protection mechanism. The transaction enters the [mempool](https://tendermint.com/docs/tendermint-core/mempool.html#transaction-ordering) only if the gas prices of the transaction is superior to one of the minimum gas price in `minGasPrices` (i.e. if `minGasPrices == 1uatom, 1upho`, the `gas-price` of the transaction must be superior to `1uatom` OR `1upho`).
- `appVersion`: Version of the application. It is set in the [application's constructor function](../basics/app-anatomy.md#constructor-function).

## Constructor

`NewBaseApp(name string, logger log.Logger, db dbm.DB, txDecoder sdk.TxDecoder, options ...func(*BaseApp),)` is the constructor function for `baseapp`. It is called from the [application's constructor function](../basics/app-anatomy.md#constructor-function) each time the full-node is started. 

`baseapp`'s constructor function is pretty straightforward. The only thing worth noting is the possibility to add additional [`options`](https://github.com/cosmos/cosmos-sdk/blob/master/baseapp/options.go) to `baseapp` by passing `options functions` to the constructor function, which will execute them in order. `options` are generally `setter` functions for important parameters, like `SetPruning()` to active pruning or `SetMinGasPrices()` to set the node's `min-gas-prices`. 

A list of `options` examples can be found [here](https://github.com/cosmos/cosmos-sdk/blob/master/baseapp/options.go). Naturally, developers can add additional `options` based on their application's needs. 

## States 

`baseapp` handles various parallel states for different purposes. There is the [main state](#main-state), which is the canonical state of the application, and volatile states like [`checkState`](#checkState) and [`deliverState`](#deliverstate), which are used to handle temporary states inbetween updates of the main state made during [`Commit`](#commit).

```
            Updated whenever an unconfirmed   Updated whenever a transaction         To serve user queries relayed
            transaction is received from the  is received from the underlying        from the underlying consensus
            underlying consensus engine via   consensus engine (as part of a block)  engine via the Query ABCI message
            CheckTx                           proposal via DeliverTx                   
                +----------------------+      +----------------------+       +----------------------+
                |   CheckState(t)(0)   |      |  DeliverState(t)(0)  |       |    QueryState(t)     |
                +----------------------+      |                      |       |                      |
CheckTx(tx1)               |                  |                      |       |                      |
                           v                  |                      |       |                      |
                +----------------------+      |                      |       |                      |
                |   CheckState(t)(1)   |      |                      |       |                      |
                +----------------------+      |                      |       |                      |
CheckTx(tx2)               |                  |                      |       |                      |
                           v                  |                      |       |                      |
                +----------------------+      |                      |       |                      |
                |   CheckState(t)(2)   |      |                      |       |                      |
                +----------------------+      |                      |       |                      |
CheckTx(tx3)               |                  |                      |       |                      |
                           v                  |                      |       |                      |
                +----------------------+      |                      |       |                      |
                |   CheckState(t)(3)   |      |                      |       |                      |
                +----------------------+      +----------------------+       |                      |
DeliverTx(tx1)             |                             |                   |                      |
                           v                             v                   |                      |
                +----------------------+      +----------------------+       |                      |
                |                      |      |  DeliverState(t)(1)  |       |                      |
                |                      |      +----------------------+       |                      |
DeliverTx(tx2)  |                      |                 |                   |                      | 
                |                      |                 v                   |                      |
                |                      |      +----------------------+       |                      |
                |                      |      |  DeliverState(t)(2)  |       |                      |
                |                      |      +----------------------+       |                      |
DeliverTx(tx3)  |                      |                 |                   |                      |
                |                      |                 v                   |                      |
                |                      |      +----------------------+       |                      |
                |                      |      |  DeliverState(t)(3)  |       |                      |
                +----------------------+      +----------------------+       +----------------------+
Commit()                  |                              |                               |
                          v                              v                               v
                +----------------------+      +----------------------+       +----------------------+
                |  CheckState(t+1)(0)  |      | DeliverState(t+1)(0) |       |   QueryState(t+1)    |
                +----------------------+      |                      |       |                      |
                          .                              .                               .
                          .                              .                               .
                          .                              .                               .
                
```

### Main State

The main state is the canonical state of the application. It is initialized on [`InitChain`](#initchain and updated on [`Commit`](#abci-commit) at the end of each block. 

```
+--------+                              +--------+
|        |                              |        |
|   S    +----------------------------> |   S'   |
|        |   For each T in B: apply(T)  |        |
+--------+                              +--------+
```

The main state is held by `baseapp` in a structure called the [`CommitMultiStore`](./store.md#commit-multi-store). This multi-store is used by developers to instantiate all the stores they need for each of their application's modules.   

### Volatile States

Volatile - or cached - states are used in between [`Commit`s](#commit) to manage temporary states. They are reset to the latest version of the main state after it is committed. There are two main volatile states:

- `checkState`: This cached state is initialized during [`InitChain`](#initchain), updated during [`CheckTx`](#abci-checktx) when an unconfirmed transaction is received, and reset to the [main state](#main-state) on [`Commit`](#abci-commit). 
- `deliverState`: This cached state is initialized during [`BeginBlock`](#beginblock), updated during [`DeliverTx`](#abci-delivertx) when a transaction included in a block is processed, and reset to the [main state](#main-state) on [`Commit`](#abci-commit).

Both `checkState` and `deliverState` are of type [`state`](https://github.com/cosmos/cosmos-sdk/blob/master/baseapp/baseapp.go#L973-L976), which includes:

- A [`CacheMultiStore`](https://github.com/cosmos/cosmos-sdk/blob/master/store/cachemulti/store.go), which is a cached version of the main `CommitMultiStore`. A new version of this store is committed at the end of each successful `CheckTx()`/`DeliverTx()` execution. 
- A [`Context`](./context.md), which carries general information (like raw transaction size, block height, ...) that might be needed in order to process the transaction during `CheckTx()` and `DeliverTx()`. The `context` also holds a cache-wrapped version of the `CacheMultiStore`, so that the `CacheMultiStore` can maintain the correct version even if an internal step of `CheckTx()` or `DeliverTx()` fails. 

## Routing

When messages and queries are received by the application, they must be routed to the appropriate module in order to be processed. Routing is done via `baseapp`, which holds a `router` for messages, and a `query router` for queries.

### Message Routing

[`Message`s](../building-modules/messages.md) need to be routed after they are extracted from transactions, which are sent from the underlying Tendermint engine via the [`CheckTx`](#checktx) and [`DeliverTx`](#delivertx) ABCI messages. To do so, `baseapp` holds a [`router`](https://github.com/cosmos/cosmos-sdk/blob/master/baseapp/router.go) which maps `paths` (`string`) to the appropriate module [`handler`](./handler.md). Usually, the `path` is the name of the module.

The application's `router` is initilalized with all the routes using the application's [module manager](./modules.md#module-manager), which itself is initialized with all the application's modules in the application's [constructor](../basics/app-anatomy.md#app-constructor). 

### Query Routing

Similar to messages, queries need to be routed to the appropriate module's [querier](./querier.md). To do so, `baseapp` holds a [`query router`](https://github.com/cosmos/cosmos-sdk/blob/master/baseapp/queryrouter.go), which maps `paths` (`string`) to the appropriate module [`querier`](./querier.md). Usually, the `path` is the name of the module. 

Just like the `router`, the `query router` is initilalized with all the query routes using the application's [module manager](./modules.md#module-manager), which itself is initialized with all the application's modules in the application's [constructor](../basics/app-anatomy.md#app-constructor).

## Main ABCI Messages

The [Application-Blockchain Interface](https://tendermint.com/docs/spec/abci/) (ABCI) is a generic interface that connects a state-machine with a consensus engine to form a functional full-node. It can be wrapped in any language, and needs to be implemented by each application-specific blockchain built on top of an ABCI-compatible consensus engine like Tendermint. 

The consensus engine handles two main tasks:

- The networking logic, which mainly consists in gossiping block parts, transactions and consensus votes.
- The consensus logic, which results in the deterministic ordering of transactions in the form of blocks. 

It is **not** the role of the consensus engine to define the state or the validity of transactions. Generally, transactions are handled by the consensus engine in the form of `[]bytes`, and relayed to the application via the ABCI to be decoded and processed. At keys moments in the networking and consensus processes (e.g. beginning of a block, commit of a block, reception of an unconfirmed transaction, ...), the consensus engine emits ABCI messages for the state-machine to act on. 

Developers building on top of the Cosmos SDK need not implement the ABCI themselves, as all the ABCI messages are implemented as a set of `baseapp`'s methods in the Cosmos SDK. Let us go through the main ABCI messages that `baseapp` handles: [`CheckTx`](#checktx) and [`DeliverTx`](#delivertx). Note that these ABCI messages are different from the [`message`s](../building-modules/messages.md)  contained in [`transactions`](../basics/transactions.md), the purpose of which is to trigger state-transitions. 

### CheckTx

The `CheckTx` ABCI message is sent by the underlying consensus engine when a new unconfirmed (i.e. not yet included in a valid block) transaction is received by a full-node, and it is handled by the `CheckTx(req abci.RequestCheckTx)` method of `baseapp` (abbreviated to `CheckTx()` thereafter). The role of `CheckTx()` is to guard the full-node's mempool (where unconfirmed transactions are stored until they are included in a block) from spam transactions. Unconfirmed transactions are relayed to peers only if they pass `CheckTx()`.

`CheckTx()` can perform both *stateful* and *stateless* checks, but developers should strive to make them lightweight. In the Cosmos SDK, after [decoding transactions](./encoding.md), `CheckTx()` is implemented to do the following checks:

1. Extract the `message`s from the transaction.
2. Perform *stateless* checks by calling `ValidateBasic()` on each of the `messages`. This is done first, as *stateless* checks are less computationally expensive than *stateful* checks. If `ValidateBasic()` fail, `CheckTx` returns before running *stateful* checks, which saves resources.
3. Perform non-module related *stateful* checks on the account. This step is mainly about checking that the `message` signatures are valid, that enough fees are provided and that the sending account has enough funds to pay for said fees. Note that no precise [`gas`](./accounts-fees-gas.md#gas) counting occurs here, as `message`s are not processed. Usually, the [`anteHandler`](./accounts-fees-gas.md#antehandler) will check that the `gas` provided with the transaction is superior to a minimum reference gas amount based on the raw transaction size, in order to avoid spam with transactions that provide 0 gas. 
4. Ensure that a [`Route`](#message-routing) exists for each `message`, but do **not** actually process `message`s. `Message`s only need to be processed when the canonical state need to be updated, which happens during `DeliverTx`. 

Steps 2. and 3. are  performed by the [`anteHandler`](./accounts-fees-gas.md#antehandler) in the [`RunTx()`](#runtx()-,antehandler-and-runmsgs()) function, which `CheckTx()` calls with the `runTxModeCheck` mode. During each step of `CheckTx()`, a special [volatile state](#volatile-states) called `checkState` is updated. This state is used to keep track of the temporary changes triggered by the `CheckTx()` calls of each transaction without modifying the [main canonical state](#main-state) . For example, when a transaction goes through `CheckTx()`, the transaction's fees are deducted from the sender's account in `checkState`. If a second transaction is received from the same account before the first is processed, and the account has consumed all its funds in `checkState` during the first transaction, the second transaction will fail `CheckTx`() and be rejected. In any case, the sender's account will not actually pay the fees until the transaction is actually included in a block, because `checkState` never gets committed to the main state. `checkState` is reset to the latest state of the main state each time a blocks gets [committed](#commit).

`CheckTx()` returns a response to the underlying consensus engine of type [`abci.ResponseCheckTx`](https://tendermint.com/docs/spec/abci/abci.html#messages). The response contains:

- `Code (uint32)`: Response Code. `0` if successful. 
- `Data ([]byte)`: Result bytes, if any.
- `Log (string):` The output of the application's logger. May be non-deterministic.
- `Info (string):` Additional information. May be non-deterministic.
- `GasWanted (int64)`: Amount of gas requested for transaction. It is provided by users when they generate the transaction. 
- `GasUsed (int64)`: Amount of gas consumed by transaction. During `CheckTx`, this value is computed by multiplying the standard cost of a transaction byte by the size of the raw transaction (click [here](https://github.com/cosmos/cosmos-sdk/blob/master/x/auth/ante.go#L101) for an example). 
- `Tags ([]cmn.KVPair)`: Key-Value tags for filtering and indexing transactions (eg. by account).
- `Codespace (string)`: Namespace for the Code.

### DeliverTx

When the underlying consensus engine receives a block proposal, each transaction in the block needs to be processed by the application. To that end, the underlying consensus engine sends a `DeliverTx` message to the application for each transaction in a sequential order. This ABCI message is handled by the `DeliverTx()` method of `baseapp` (abbreviated to `DeliverTx()` thereafter).

Before the first transaction of a given block is processed, a [volatile state](#volatile-states) called `deliverState` is intialized during [`BeginBlock`](#beginblock). This state is updated each time a transaction is processed via `DeliverTx()`, and committed to the [main state](#main-state) when the block is [committed](#commit), after what is is set to `nil`. 

`DeliverTx()` performs the **exact same steps as `CheckTx()`**, with a little caveat at step 3 and the addition of a fifth step:

3. The `anteHandler` does **not** check that the transaction's `gas-prices` is sufficient. That is because the `min-gas-prices` value `gas-prices` is checked against is local to the node, and therefore what is enough for one full-node might not be for another. This means that the proposer can potentially include transactions for free, although they are not incentivised to do so, as they earn a bonus on the total fee of the block they propose. 
5. For each `message` in the transaction, route to the appropriate module's [`handler`](../building-modules/handler.md). Additional *stateful* checks are performed, and the cache-wrapped multistore held in `deliverState`'s `context` is updated by the module's `keeper`. If the `handler` returns successfully, the cache-wrapped multistore held in `context` is written to `deliverState` `CacheMultiStore`. 

During step 5., each read/write to the store increases the value of `GasConsumed`. You can find the default cost of each operation [here](https://github.com/cosmos/cosmos-sdk/blob/master/store/types/gas.go#L142-L150). At any point, if `GasConsumed > GasWanted`, the function returns with `Code != 0` and `DeliverTx()` fails. 

`DeliverTx()` returns a response to the underlying consensus engine of type [`abci.ResponseCheckTx`](https://tendermint.com/docs/spec/abci/abci.html#messages). The response contains:

- `Code (uint32)`: Response Code. `0` if successful. 
- `Data ([]byte)`: Result bytes, if any.
- `Log (string):` The output of the application's logger. May be non-deterministic.
- `Info (string):` Additional information. May be non-deterministic.
- `GasWanted (int64)`: Amount of gas requested for transaction. It is provided by users when they generate the transaction. 
- `GasUsed (int64)`: Amount of gas consumed by transaction. During `DeliverTx`, this value is computed by multiplying the standard cost of a transaction byte by the size of the raw transaction (click [here](https://github.com/cosmos/cosmos-sdk/blob/master/x/auth/ante.go#L101) for an example), and by adding gas each time a read/write to the store occurs. 
- `Tags ([]cmn.KVPair)`: Key-Value tags for filtering and indexing transactions (eg. by account).
- `Codespace (string)`: Namespace for the Code.

## RunTx(), AnteHandler and RunMsgs()

### RunTx()

`RunTx()` is called from `CheckTx()`/`DeliverTx()` to handle the transaction, with `runTxModeCheck` or `runTxModeDeliver` as parameter to differentiate between the two modes of execution. Note that when `RunTx()` receives a transaction, it has already been decoded. 

The first thing `RunTx()` does upon being called is to retrieve the `context`'s `CacheMultiStore` by calling the `getContextForTx()` function with the appropriate mode (either `runTxModeCheck` or `runTxModeDeliver`). This `CacheMultiStore` is a cached version of the main store instantiated during `BeginBlock` for `DeliverTx` and during the `Commit` of the previous block for `CheckTx`.  After that, two `defer func()` are called for [`gas`](./accounts-fees-gas.md#gas) management. They are executed when `RunTx()` returns and make sure `gas` is actually consumed, and will throw errors, if any.

After that, `RunTx()` calls `ValidateBasic()` on each `message`in the `Tx`, which runs prelimary *stateless* validity checks. If any `message` fails to pass `ValidateBasic()`, `RunTx()` returns with an error. 

Then, the [`anteHandler`](#antehandler) of the application is run (if it exists). In preparation of this step, both the `checkState`/`deliverState`'s `context` and `context`'s `CacheMultiStore` are cached-wrapped using the [`cacheTxContext()`](https://github.com/cosmos/cosmos-sdk/blob/master/baseapp/baseapp.go#L781-L798) function. This allows `RunTx()` not to commit the changes made to the state during the execution of `anteHandler` if it ends up failing. It also prevents the module implementing the `anteHandler` from writing to state, which is an important part of the [object-capabilities](./ocap.md) of the Cosmos SDK. 

Finally, the [`RunMsgs()`](#runmsgs()) function is called to process the `messages`s in the `Tx`. In preparation of this step, just like with the `anteHandler`, both the `checkState`/`deliverState`'s `context` and `context`'s `CacheMultiStore` are cached-wrapped using the `cacheTxContext()` function.

### AnteHandler

The `AnteHandler` is a special handler that implements the [`anteHandler` interface](https://github.com/cosmos/cosmos-sdk/blob/master/types/handler.go#L8) and is used to authenticate the transaction before the transaction's internal messages are processed.

The `AnteHandler` is theoretically optional, but still a very important component of public blockchain networks. It serves 3 primary purposes:

- Be a primary line of defense against spam and second line of defense (the first one being the mempool) against transaction replay with fees deduction and [`sequence`](./tx-msgs.md#sequence) checking. 
- Perform preliminary *stateful* validity checks like ensuring signatures are valid or that the sender has enough funds to pay for fees. 
- Play a role in the incentivisation of stakeholders via the collection of transaction fees. 

`baseapp` holds an `anteHandler` as parameter, which is initialized in the [application's constructor](../basics/app-anatomy.md#application-constructor). The most widely used `anteHandler` today is that of the [`auth` module](https://github.com/cosmos/cosmos-sdk/blob/master/x/auth/ante.go).

### RunMsgs()

`RunMsgs()` is called from `RunTx()` with `runTxModeCheck` as parameter to check the existence of a route for each message contained in the transaction, and with `runTxModeDeliver` to actually process the `message`s. 

First, it retreives the `message`'s `route` using the `Msg.Route()` method. Then, using the application's [`router`](#routing) and the `route`, it checks for the existence of a `handler`. At this point, if `mode == runTxModeCheck`, `RunMsgs()` returns. If instead `mode == runTxModeDeliver`, the [`handler`](../building-modules.md#handler) function for the message is executed, before `RunMsgs()` returns. 

## Other ABCI Messages

### InitChain

The [`InitChain` ABCI message](https://tendermint.com/docs/app-dev/abci-spec.html#initchain) is sent from the underlying Tendermint engine when the chain is first started, and is handled by the `InitChain(req abci.RequestInitChain)` method of `baseapp`. It is mainly used to **initialize** parameters and state like:

- [Consensus Parameters](https://tendermint.com/docs/spec/abci/apps.html#consensus-parameters) via `setConsensusParams`. 
- [`checkState` and `deliverState`](#volatile-states) via `setCheckState` and `setDeliverState`.
- The [block gas meter](../basics/accounts-fees-gas.md#block-gas-meter), with infinite gas to process genesis transactions. 

Finally, the `InitChain(req abci.RequestInitChain)` method of `baseapp` calls the [`initChainer()`](../basics/app-anatomy.md#initchainer) of the application in order to initialize the main state of the application from the [`genesis file`](./genesis.md) and, if defined, call the `InitGenesis` function of each of the application's modules.

### BeginBlock

The [`BeginBlock` ABCI message](#https://tendermint.com/docs/app-dev/abci-spec.html#beginblock) is sent from the underlying Tendermint engine when a block proposal created by the correct proposer is received, before [`DeliverTx()`](#delivertx) is run for each transaction in the block. It allows developers to have logic be executed at the beginning of each block. In the Cosmos SDK, the `BeginBlock(req abci.RequestBeginBlock)` method does the following:

- Initialize [`deliverState`](#volatile-states) with the latest header using the `req abci.RequestBeginBlock` passed as parameter via the [`setDeliverState`](https://github.com/cosmos/cosmos-sdk/blob/master/baseapp/baseapp.go#L283-L289) function. 
- Initialize the [block gas meter](../basics/accounts-fees-gas.md#block-gas-meter) with the `maxGas` limit. The `gas` consumed within the block cannot go above `maxGas`. This parameter is defined in the application's consensus parameters. 
- Run the application's [`begingBlocker()`](../basics/app-anatomy.md#beginblocker-and-endblock), which mainly runs the `BeginBlocker()` method of each of the application's modules.
- Set the [`VoteInfos`](https://tendermint.com/docs/app-dev/abci-spec.html#voteinfo) of the application, i.e. the list of validators whose *precommit* for the previous block was included by the proposer of the current block. This information is carried into the [`Context`](./context.md) so that it can be used during `DeliverTx` and `EndBlock`.

### EndBlock

The [`EndBlock` ABCI message](#https://tendermint.com/docs/app-dev/abci-spec.html#endblock) is sent from the underlying Tendermint engine after [`DeliverTx`](#delivertx) as been run for each transactioni n the block. It allows developers to have logic be executed at the end of each block. In the Cosmos SDK, the bulk `EndBlock(req abci.RequestEndBlock)` method is to run the application's [`endBlocker()`](../basics/app-anatomy.md#beginblocker-and-endblock), which mainly runs the `EndBlocker()` method of each of the application's modules. 

### Commit

The [`Commit` ABCI message](https://tendermint.com/docs/app-dev/abci-spec.html#commit) is sent from the underlying Tendermint engine after the full-node has received *precommits* from 2/3+ of validators (weighted by voting power). On the `baseapp` end, the `Commit(res abci.ResponseCommit)` function is implemented to commit all the valid state transitions that occured during `BeginBlock()`, `DeliverTx()` and `EndBlock()` and to reset state for the next block. 

To commit state-transitions, the `Commit` function calls the `Write()` function on `deliverState.ms`, where `deliverState.ms` is a cached multistore of the main store `app.cms`. Then, the `Commit` function sets `checkState` to the latest header (obtbained from `deliverState.ctx.BlockHeader`) and `deliverState` to `nil`.

Finally, `Commit` returns the hash of the commitment of `app.cms` back to the underlying consensus engine. This hash is used as a reference in the header of the next block. 

### Info

The [`Info` ABCI message](https://tendermint.com/docs/app-dev/abci-spec.html#info) is a simple query from the underlying consensus engine, notably used to sync the latter with the application during a handshake that happens on startup. When called, the `Info(res abci.ResponseInfo)` function from `baseapp` will return the application's name, version and the hash of the last commit of `app.cms`. 

### Query 

The [`Query` ABCI message](https://tendermint.com/docs/app-dev/abci-spec.html#query) is used to serve queries received from the underlying consensus engine, including queries received via RPC like Tendermint RPC. It is the main entrypoint to build interfaces with the application. The application must respect a few rules when implementing the `Query` method, which are outlined [here](https://tendermint.com/docs/app-dev/abci-spec.html#query). 

The `baseapp` implementation of the `Query(req abci.RequestQuery)` method is a simple dispatcher serving 4 main categories of queries:

- Application-related queries like querying the application's version, which are served via the `handleQueryApp` method.
- Direct queries to the multistore, which are served by the `handlerQueryStore` method. These direct queryeis are different from custom queries which go through `app.queryRouter`, and are mainly used by third-party service provider like block explorers. 
- P2P queries, which are served via the `handleQueryP2P` method. These queries return either `app.addrPeerFilter` or `app.ipPeerFilter` that contain the list of peers filtered by address or IP respectively. These lists are first initialized via `options` in `baseapp`'s [constructor](#constructor).
- Custom queries, which encompass most queries, are served via the `handleQueryCustom` method. The `handleQueryCustom` cache-wraps the multistore before using the `queryRoute` obtained from [`app.queryRouter`](#query-routing) to map the query to the appropriate module's [`querier`](../building-modules/querier.md). 

## Next

Learn more about [stores](./store.md).