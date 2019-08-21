# Introduction to SDK Modules

## Pre-requisite Reading

- [Anatomy of an SDK application](../basics/app-anatomy.md)
- [Lifecycle of an SDK transaction](../basics/tx-lifecycle.md)

## Synopsis

Modules define most of the logic of any SDK application. Developers compose module together to build their custom application-specific blockchains. This document outlines the basic concepts behind SDK modules and how to approach module management. 

- [Role of Modules in an SDK application](#role-of-modules-in-an-sdk-application)
- [How to Approach Building Modules as a Developer](#how-to-approach-building-modules-as-a-developer)
- [Application Module Interface](#application-module-interface)
- [Module Manager](#module-manager)

## Role of Modules in an SDK Application

The Cosmos SDK can be thought as the Ruby-on-Rails of blockchain development. It comes with a core that provides the basic functionalities every blockchain application need, like a boilerplate implementation of the ABCI to communicate with the underlying consensus engine, a multistore to persist state, a server to form a full-node and interfaces to handle queries.  

On top of this core, the SDK enables developers to build modules that implement the business logic of their application. In other words, SDK modules implement the bulk of the logic of applications, while the core does the wiring and enables modules to be composed together. The end goal is to build a robust ecosystem of open-source SDK modules, making it increasingly easier to build complex blockchain applications. 

SDK Modules can be seen as little state-machines within the state-machine. They generally define a subset of the state using one ore multiple `KVStore` in the [main multistore](../core/store.md), as well as a subset of [`message` types](./message.md). These `message`s are routed by one of the main component of SDK core, [`baseapp`](../core/baseapp.md), to the [`handler`](./handler.md) of the module that define them. 

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

While there is no definitive guidelines for writing modules, here are some important design principles developers should keep in mind when building them:

- **Composability**: SDK applications are almost always composed of multiple modules. This means developers need to carefully consider the integration of their module not only with the core of the Cosmos SDK, but also with other modules. The former is achieved by following standard design patterns outlined [here](#main-components-of-sdk-modules), while the latter is achieved by properly exposing the store(s) of the module via the [`keeper`](./keeper.md). 
- **Specialization**: A direct consequence of the **composability** feature is that modules should be **specialized**. Developers should carefully establish the scope of their module and not batch multiple functionalities into the same module. This separation of concern enables modules to be re-used in other projects and improves the upgradability of the application. **Specialization** also plays an important role in the [object-capabilities model](../core/ocap.md) of the Cosmos SDK. 
- **Capabilities**: Most modules need to read and/or write to the store(s) of other modules. However, in an open-source environment, it is possible for some module to be malicious. That is why module developers need to carefully think not only about how their module interracts with other modules, and how to give access to the module's store(s). The Cosmos SDK takes a capabilities-oriented approach to inter-module security. This means that each store defined by a module is accessed by a `key`, which is held by the module's [`keeper`](./keeper.md). This `keeper` defines how to access the store(s) and under what conditions. Access to the module's store(s) is done by passing a reference to the module's `keeper`. 

## Main Components of SDK Module

Modules are by convention defined in the `.x/` subfolder (e.g. the `bank` module will be defined in the `./x/bank` folder). They generally share the same core components:

- Custom [`message` types](./message.md) to trigger state-transitions. 
- A [`handler`](./handler.md) used to process messages when they are routed to the module by [`baseapp`](../core/baseapp.md#message-routing). 
- A  [`keeper`](./keeper.md), used to access the module's store(s) and update the state. 
- A [`querier`](./querier.md), used to process user queries when they are routed to the module by [`baseapp`](../core/baseapp.md#query-routing).
- Interfaces, for end users to query the subset of the state defined by the module and create `message`s of the custom types defined in the module.

In addition to these components, modules implement the `AppModule` interface in order to be managed by the [`module manager`](./module-manager.md). 

## Next

Read more on the [`AppModule` interface and the `module manager`](./module-manager.md)

