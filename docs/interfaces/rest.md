<!--
order: 4
synopsis: "This document describes how to create a REST interface for an SDK **application**. A separate document for creating a [**module**](../building-modules/intro.md) REST interface can be found [here](#../module-interfaces.md#rest)."
-->

# REST Interface

## Prerequisites {hide}

- [Query Lifecycle](./query-lifecycle.md) {prereq}
- [Application CLI](./cli.md) {prereq}

## Application REST Interface

Building the REST Interface for an application is done by [aggregating REST Routes](#registering-routes) defined in the application's modules. This interface is served by a REST Server [REST server](#rest-server), which route requests and output responses in the application itself. The SDK comes with its own REST Server by default. To enable it, the `rest.ServeCommand` command needs to be added as a subcommand of the `rootCmd` in the `main()` function of the [CLI interface](./cli.md):

```go
rootCmd.AddCommand(rest.ServeCommand(cdc, registerRoutes))
```

Users will then be able to use the application CLI to start a new REST server, a local server through which they can securely interact with the application without downloading the entire state. The command entered by users would look something like this:

```bash
appcli rest-server --chain-id <chainID> --trust-node
```

Note that if `trust-node` is set to `false`, the REST server will verify the query proof against the merkle root (contained in the block header).

## REST Server

A REST Server is used to receive and route HTTP Requests, obtain the results from the application, and return a response to the user. The REST Server defined by the SDK `rest` package contains the following:

- **Router:** A router for HTTP requests. A new router can be instantiated for an application and used to match routes based on path, request method, headers, etc. The SDK uses the [Gorilla Mux Router](https://github.com/gorilla/mux).
- **CLIContext:** A [`CLIContext`](./query-lifecycle.md#clicontext) created for a user interaction.
- **Keybase:** A [Keybase](../basics/accounts.md#keybase) is a key manager.
- **Logger:** A logger from Tendermint `Log`, a log package structured around key-value pairs that allows logging level to be set differently for different keys. The logger takes `Debug()`, `Info()`, and `Error()`s.
- **Listener:** A [listener](https://golang.org/pkg/net/#Listener) from the net package.

Of the five, the only attribute that application developers need interact with is the `router`: they need to add routes to it so that the REST server can properly handle queries. See the next section for more information on registering routes.

In order to enable the REST Server in an SDK application, the `rest.ServeCommand` needs to be added to the application's command-line interface. See the [above section](#application-rest-interface) for more information.

## Registering Routes

To include routes for each module in an application, the CLI must have some kind of function to register routes in its REST Server. This function is called `RegisterRoutes()`, and is utilized by the `ServeCommand` and must include routes for each of the application's modules. Since each module used by an SDK application implements a [`RegisterRESTRoutes`](../building-modules/module-interfaces.md#rest) function, application developers simply use the [Module Manager](../building-modules/module-manager.md) to call this function for each module (this is done in the [application's constructor](../basics/app-anatomy.md#constructor-function)).

At the bare minimum, a `RegisterRoutes()` function should use the SDK client package `RegisterRoutes()` function to be able to route RPC calls, and instruct the application Module Manager to call `RegisterRESTRoutes()` for all of its modules. This is done in the `main.go` file of the CLI (typically located in `./cmd/appcli/main.go`).

```go
func registerRoutes(rs *rest.RestServer) {
	client.RegisterRoutes(rs.CliCtx, rs.Mux)
	app.ModuleBasics.RegisterRESTRoutes(rs.CliCtx, rs.Mux)
}
```

This function is specific to the application and passed in to the `ServeCommand`, which should be added to the `rootCmd` as such:

```go
rootCmd.AddCommand(rest.ServeCommand(cdc, registerRoutes))
```

## Cross-Origin Resource Sharing (CORS)

[CORS policies](https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS) are not enabled by default to help with security. If you would like to use the rest-server in a public environment we recommend you provide a reverse proxy, this can be done with [nginx](https://www.nginx.com/). For testing and development purposes there is an `unsafe_cors` flag that can be passed to the cmd to enable accepting cors from everyone.

```sh
gaiacli rest-server --chain-id=test \
    --laddr=tcp://localhost:1317 \
    --node tcp://localhost:26657 \
    --trust-node=true --unsafe-cors
```
