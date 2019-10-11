# Interfaces

## Prerequisites

* [Anatomy of an SDK Application](../basics/app-anatomy.md)
* [Lifecycle of a Transaction](../basics/tx-lifecycle.md)


## Synopsis

Typically, SDK applications include some type of interface that users interact with to utilize the application's functionalities. This document introduces user command-line and REST interfaces.

- [Types of Application Interfaces](#types-of-application-interfaces)
- [Module vs Application Interfaces](#module-vs-application-interfaces)
  + [Module Developer Responsibilities](#module-developer-responsibilities)
  + [Application Developer Responsibilities](#application-developer-responsibilities)


## Types of Application Interfaces

SDK applications generally have a Command-Line Interface (CLI) and REST Interface to support interactions with a [full-node](../core/node.md). The SDK is opinionated about how to create these two interfaces; all modules specify [Cobra commands](https://github.com/spf13/cobra) and register routes using [Gorilla Mux routers](https://github.com/gorilla/mux). The CLI and REST Interface are conventionally defined in the application `app/cmd/cli` folder.


## Module vs Application Interfaces

The process of creating an application interface is distinct from creating a [module interface](../building-modules/interfaces.md), though the components are closely intertwined. As expected, the module interface handles the bulk of the underlying logic, defining ways for end-users to create [messages](../building-modules/messages-and-queries.md#messages) handled by the module and [queries](../building-modules/messages-and-queries.md#queries) to the subset of application state within the scope of the module. On the other hand, the application interfaces aggregate module-level interfaces in order to route messages and queries to the appropriate modules. Application interfaces also handle root-level responsibilities such as signing and broadcasting [transactions](../core/transactions.md) that wrap messages.

### Module Developer Responsibilities

In regards to interfaces, module developers include the following definitions:

* **CLI commands:** Specifically, [Transaction commands](../building-modules/interfaces.md#transaction-commands) and [Query commands](../building-modules/interfaces.md#query-commands). These are commands that users will invoke when interacting with the application to create transactions and queries. For example, if an application enables sending coins through the [`auth`](https://github.com/cosmos/cosmos-sdk/tree/67f6b021180c7ef0bcf25b6597a629aca27766b8/docs/spec/auth) module, users will create `tx auth send` transactions.
* **Request Handlers:** Also categorized into Transaction and Query requests. Transactions will require HTTP [Request Types](../building-modules/interfaces.md#request-types) in addition to [Request Handlers](../building-modules/interfaces.md#request-handlers) in order to encapsulate all of the user's options (e.g. gas prices).
* **REST Routes:** Given a router, the module interface registers paths with the aforementioned [Request Handlers](../building-modules/interfaces.md#request-handlers) for each type of request.

Module interfaces are designed to be generic. Both commands and request types  include required user input (through flags or request body) which are different for each application. This section of documents will only detail application interfaces; to read about how to build module interfaces, click [here](../building-modules/interfaces.md).

### Application Developer Responsibilities

In regards to interfaces, application developers include:

* **CLI Root Command:** The [root command](./cli.md#root-command) adds subcommands to include all of the functionality for the application, mainly module [transaction](./cli.md#transaction-commands) and [query](./cli.md#query-commands) commands from the application's module(s).
* **App Configurations:** All application-specific values are the responsibility of the application developer, including the [`codec`](../core/encoding.md) used to marshal requests before relaying them to a node.
* **User Configurations:** Some values are specific to the user, such as the user's address and which node they are connected to. The CLI has a [configurations](./cli.md#configurations) function to set these values.
* **RegisterRoutes Function:** [Routes](./rest.md#registerroutes) must be registered and passed to an instantiated [REST server](./rest.md#rest-server) so that it knows how to route requests for this particular application.


## Next

Read about the [Lifecycle of a Query](./query-lifecycle.md).
