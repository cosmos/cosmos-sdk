# REST Interface

## Prerequisites

* [Query Lifecycle](./query-lifecycle.md)
* [Application CLI](./cli.md)

## Synopsis

This document describes how to create a REST interface for an SDK application. A separate document for creating a module REST interface can be found [here](#../module-interfaces.md#rest).

- [Application REST Interface](#application-rest-interface)
- [REST Server](#rest-server)
- [Registering Routes](#registering-routes)

## Application REST Interface

Building the REST Interface for an application involves creating a [REST server](./rest.md#rest-server) to route requests and output responses. The SDK has its own REST Server type used for [LCDs](../core/node.md) (light-client daemons). It has a `ServeCommand` that takes in an application's `codec` and `RegisterRoutes()` function, starts up a new REST Server, and registers routes using function provided from the application. To enable this command, it should be added as a subcommand of the root command in the `main()` function of the CLI interface:

```go
rootCmd.AddCommand(lcd.ServeCommand(cdc, registerRoutes))
```

Users can use the application CLI to start a new LCD, a local server through which they can securely interact with the application without downloading the entire state. The command entered by users would look something like this:

```bash
appcli rest-server --chain-id <chainID> --trust-node
```


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
