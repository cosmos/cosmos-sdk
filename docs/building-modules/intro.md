<!--
order: 1 
-->

# Introduction to SDK Modules

Modules define most of the logic of SDK applications. Developers compose module together using the Cosmos SDK to build their custom application-specific blockchains. This document outlines the basic concepts behind SDK modules and how to approach module management. {synopsis}

## Pre-requisite Readings

- [Anatomy of an SDK application](../basics/app-anatomy.md) {prereq}
- [Lifecycle of an SDK transaction](../basics/tx-lifecycle.md) {prereq}

## Role of Modules in an SDK Application

The Cosmos SDK can be thought as the Ruby-on-Rails of blockchain development. It comes with a core that provides the basic functionalities every blockchain application needs, like a [boilerplate implementation of the ABCI](../core/baseapp.md) to communicate with the underlying consensus engine, a [`multistore`](../core/store.md#multistore) to persist state, a [server](../core/node.md) to form a full-node and [interfaces](./module-interfaces.md) to handle queries.  

On top of this core, the Cosmos SDK enables developers to build modules that implement the business logic of their application. In other words, SDK modules implement the bulk of the logic of applications, while the core does the wiring and enables modules to be composed together. The end goal is to build a robust ecosystem of open-source SDK modules, making it increasingly easier to build complex blockchain applications. 

SDK Modules can be seen as little state-machines within the state-machine. They generally define a subset of the state using one or more `KVStore`s in the [main multistore](../core/store.md), as well as a subset of [message types](./messages-and-queries.md#messages). These messages are routed by one of the main component of SDK core, [`BaseApp`](../core/baseapp.md), to the [`Msg` service](./msg-services.md) of the module that define them. 

```
                                      +
                                      |
                                      |  Transaction relayed from the full-node's consensus engine 
                                      |  to the node's application via DeliverTx
                                      |  
                                      |
                                      |
                +---------------------v--------------------------+
                |                 APPLICATION                    |
                |                                                |
                |     Using baseapp's methods: Decode the Tx,    |
                |     extract and route the message(s)           |
                |                                                |
                +---------------------+--------------------------+
                                      |
                                      |
                                      |
                                      +---------------------------+
                                                                  |
                                                                  |
                                                                  |
                                                                  |  Message routed to the correct
                                                                  |  module to be processed
                                                                  |
                                                                  |
+----------------+  +---------------+  +----------------+  +------v----------+
|                |  |               |  |                |  |                 |
|  AUTH MODULE   |  |  BANK MODULE  |  | STAKING MODULE |  |   GOV MODULE    |
|                |  |               |  |                |  |                 |
|                |  |               |  |                |  | Handles message,|
|                |  |               |  |                |  | Updates state   |
|                |  |               |  |                |  |                 |
+----------------+  +---------------+  +----------------+  +------+----------+
                                                                  |
                                                                  |
                                                                  |
                                                                  |
                                       +--------------------------+
                                       |
                                       | Return result to the underlying consensus engine (e.g. Tendermint)
                                       | (0=Ok, 1=Err)
                                       v
```

As a result of this architecture, building an SDK application usually revolves around writing modules to implement the specialized  logic of the application, and composing them with existing modules to complete the application. Developers will generally work on modules that implement logic needed for their specific use case that do not exist yet, and will use existing modules for more generic functionalities like staking, accounts or token management. 

## How to Approach Building Modules as a Developer

While there are no definitive guidelines for writing modules, here are some important design principles developers should keep in mind when building them:

- **Composability**: SDK applications are almost always composed of multiple modules. This means developers need to carefully consider the integration of their module not only with the core of the Cosmos SDK, but also with other modules. The former is achieved by following standard design patterns outlined [here](#main-components-of-sdk-modules), while the latter is achieved by properly exposing the store(s) of the module via the [`keeper`](./keeper.md). 
- **Specialization**: A direct consequence of the **composability** feature is that modules should be **specialized**. Developers should carefully establish the scope of their module and not batch multiple functionalities into the same module. This separation of concern enables modules to be re-used in other projects and improves the upgradability of the application. **Specialization** also plays an important role in the [object-capabilities model](../core/ocap.md) of the Cosmos SDK. 
- **Capabilities**: Most modules need to read and/or write to the store(s) of other modules. However, in an open-source environment, it is possible for some module to be malicious. That is why module developers need to carefully think not only about how their module interacts with other modules, but also about how to give access to the module's store(s). The Cosmos SDK takes a capabilities-oriented approach to inter-module security. This means that each store defined by a module is accessed by a `key`, which is held by the module's [`keeper`](./keeper.md). This `keeper` defines how to access the store(s) and under what conditions. Access to the module's store(s) is done by passing a reference to the module's `keeper`. 

## Main Components of SDK Modules

Modules are by convention defined in the `./x/` subfolder (e.g. the `bank` module will be defined in the `./x/bank` folder). They generally share the same core components:

- A  [`keeper`](./keeper.md), used to access the module's store(s) and update the state.
- A [`Msg` service](./messages-and-queries.md#messages) used to process messages when they are routed to the module by [`BaseApp`](../core/baseapp.md#message-routing) and trigger state-transitions.
- A [query service](./query-services.md), used to process user queries when they are routed to the module by [`BaseApp`](../core/baseapp.md#query-routing).
- Interfaces, for end users to query the subset of the state defined by the module and create `message`s of the custom types defined in the module.

In addition to these components, modules implement the `AppModule` interface in order to be managed by the [`module manager`](./module-manager.md). 

Please refer to the [structure document](./structure.md) to learn about the recommended structure of a module's directory. 

## Next {hide}

Read more on the [`AppModule` interface and the `module manager`](./module-manager.md) {hide}
