# Interfaces

## Prerequisites

* [Anatomy of an SDK Application](../basics/app-anatomy.md)
* [Lifecycle of a Transaction](../basics/tx-lifecycle.md)


## Synopsis

Every application must include some interface users can use to interact with the defined functionalities. This document introduces user interfaces for SDK applications.

- [Types of Application Interfaces](#types-of-application-interfaces)
- [Module vs Application Interfaces](#module-vs-application-interfaces)
  + [Module Developer Responsibilities](#module-developer-responsibilities)
  + [Application Developer Responsibilities](#application-developer-responsibilities)


## Types of Application Interfaces

SDK applications should have a Command-Line Interface (CLI) and REST Interface to support HTTP requests. The SDK is opinionated about how to create these two interfaces; all modules specify [Cobra commands](https://github.com/spf13/cobra) and register routes using [Gorilla Mux routers](https://github.com/gorilla/mux). 


## Module vs Application Interfaces

The CLI and REST Interface are conventionally defined in the application `/cmd/cli` folder. The process of creating an application interface is vastly different from a module interface, though the components are closely intertwined. As expected, the module interface handles the bulk of the underlying logic, unpacking user requests into arguments and routes, and neatly marshaling everything into ABCI requests to be handled by Baseapp. On the other hand, the application interface handles the user configurations and customizations, instantiates the application-specific values and objects, and passes everything to module interface functions.

### Module Developer Responsibilities

In regards to interfaces, the module developers' responsibilities include:

* **CLI commands:** Specifically, [Transaction commands](../building-modules/interfaces.md#transaction-commands) and [Query commands](../building-modules/interfaces.md#query-commands). These are commands that users will invoke when interacting with the application. For example, if an application enables sending coins through the [`auth`](https://github.com/cosmos/cosmos-sdk/tree/67f6b021180c7ef0bcf25b6597a629aca27766b8/docs/spec/auth) module, users will create `auth send` transactions.
* **Request Handlers:** Also categorized into Transaction and Query requests. Transactions will require HTTP [Request Types](../building-modules/interfaces.md#request-types) in addition to [Request Handlers](../building-modules/interfaces.md#request-handlers) in order to encapsulate all of the user's options (e.g. gas prices).
* **REST Routes:** Given a router, the module interface registers paths with the aforementioned [Request Handlers](../building-modules/interfaces.md#request-handlers) for each type of request.

Module interfaces are designed to be generic. Both commands and request types will include required user input (through flags or request body) which will be different for each application. This section of documents will only detail application interfaces; to read about how to build module interfaces, click [here](../building-modules/interfaces.md).

### Application Developer Responsibilities

In regards to interfaces, the application developers' responsibilities include:

* **CLI Root Command:** The root command adds subcommands to include all of the functionality for the application, mainly module transaction and query commands.
* **App Configurations:** All application-specific values are the responsibility of the application developer, including the `codec` used to marshal requests before relaying them to a node.
* **User Configurations:** Some values are specific to the user, such as the user's address and which node they are connected to.
* **RegisterRoutes Function:** To be passed to an instantiated REST Server so that it knows how to route requests for this particular application.
