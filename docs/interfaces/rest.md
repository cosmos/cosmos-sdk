# REST Interface

## Prerequisites

* [Query Lifecycle](./query-lifecycle.md)
* [Application CLI](./cli.md)

## Synopsis

This document describes how to create a REST interface for an SDK application. A separate document for creating module REST Routes can be found [here](#../module-interfaces.md#rest).

- [Application REST Interface](#application-rest-interface)
- [Request Types](#request-types)
- [REST Server](#rest-server)
- [Registering Routes](#registering-routes)

## Application REST Interface

Building the REST Interface for an application involves creating a [REST server](./rest.md#rest-server) to route requests and output responses. The SDK has its own REST Server type used for [LCDs](../core/node.md) (light-client daemons). It has a `ServeCommand` that takes in an application's `codec` and `RegisterRoutes()` function, starts up a new REST Server, and registers routes using function provided from the application. To enable this command, it should be added as a subcommand of the root command `RootCmd` in the `main()` function of the CLI interface.

Users can use the application CLI to start a new LCD, a local server through which they can securely interact with the application without downloading the entire state. The command entered by users would look something like this:

```bash
appcli rest-server --chain-id <chainID> --trust-node
```

## Request Types

HTTP Request types are defined by the module interfaces for every type of transaction. The structs all include a base request `baseReq`, the name of the request, and any arguments for the transaction.

### BaseReq

`BaseReq` is a type defined in the SDK that encapsulates much of the transaction configurations similar to CLI command flags. Users must provide the information in the body of their requests.

* `From` indicates which [account](../core/accounts-fees.md) the transaction originates from. This account is used to sign the transaction.
*	`Memo` sends a memo along with the transaction.
*	`ChainID` specifies the unique identifier of the blockchain the transaction pertains to.
*	`AccountNumber` is an identifier for the account.
*	`Sequence`is the value of a counter measuring how many transactions have been sent from the account. It is used to prevent replay attacks.
*	`Gas` refers to how much [gas](../core/gas.md), which represents computational resources, Tx consumes. Gas is dependent on the transaction and is not precisely calculated until execution, but can be estimated by providing auto as the value for `Gas`.
*	`GasAdjustment` can be used to scale gas up in order to avoid underestimating. For example, users can specify their gas adjustment as 1.5 to use 1.5 times the estimated gas.
*	`GasPrices` specifies how much the user is willing pay per unit of gas, which can be one or multiple denominations of tokens. For example, --gas-prices=0.025uatom, 0.025upho means the user is willing to pay 0.025uatom AND 0.025upho per unit of gas.
*	`Fees` specifies how much in [fees](../core/accounts-fees.md) the user is willing to pay in total. Note that the user only needs to provide either `gas-prices` or `fees`, but not both, because they can be derived from each other.
*	`Simulate` instructs the application to ignore gas and simulate the transaction running without broadcasting.

Additionally, each request may contain arguments such as a specific address pertaining to the request.

## REST Server

A REST Server is used to receive and route HTTP Requests, obtain the results from the application, and return the response to the user. The REST Server defined by the SDK LCD package contains the following:

* **Router:** A router for HTTP requests. A new router can be instantiated for an application and used to match routes based on path, request method, headers, etc. The SDK uses the [Gorilla Mux Router](https://github.com/gorilla/mux).
* **CLIContext:** A [`CLIContext`](./query-lifecycle.md#clicontext) created for a user interaction.
* **Keybase:** A [Keybase](../core/keys-accounts.md) is a key manager.
* **Logger:** A logger from Tendermint `Log`, a log package structured around key-value pairs that allows logging level to be set differently for different keys. The logger takes `Debug()`, `Info()`, and `Error()`s.
* **Listener:** A [listener](https://golang.org/pkg/net/#Listener) from the net package.

Of the five, the only attribute that developers will need to configure is the router.

## Registering Routes

To include routes for each module in an application, the CLI must have some kind of function to Register Routes in its REST Server. This `RegisterRoutes()` function is utilized by the `ServeCommand` and must include routes for each of the application's modules. Since each module used by an SDK application implements a [`RegisterRESTRoutes`](../building-modules.md#rest) function, application developers simply use the Module Manager to call this function for each module.

At the bare minimum, a `RegisterRoutes()` function should use the SDK client package `RegisterRoutes()` function to be able to route RPC calls, and instruct the application Module Manager to call `RegisterRESTRoutes()` for all of its modules:

```go
func registerRoutes(rs *lcd.RestServer) {
	client.RegisterRoutes(rs.CliCtx, rs.Mux)
	app.ModuleBasics.RegisterRESTRoutes(rs.CliCtx, rs.Mux)
}
```

This function is specific to the application and passed in to the `ServeCommand`, which should be added to the `rootCmd` as such:

```go
rootCmd.AddCommand(lcd.ServeCommand(cdc, registerRoutes))
```
