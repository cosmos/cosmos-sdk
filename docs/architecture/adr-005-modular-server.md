# ADR 005: Modular Server

## Changelog

- 2020/12/22: Initial Draft

## Author

- Federico Kunze ([@fedekunze](https://github.com/fedekunze))

## Status

Draft, Not Implemented

## Abstract

This document outlines a modular and standard approach for using servers on the SDK, so that
applications can also run custom services besides the default ones provided by the SDK.

## Context

The current `Server` implementation allows for applications to register routes, API and gRPC
services to the SDK. The problem with the current approach is that the SDK uses a concrete type,
which prevents other SDK-based blockchains to extend the services provided by the application
without forking the application `start` command (eg: `<app>d start`).

This modularity and extensibility relies on the following limitations:

    - Consistent services
    - Configuration extensibility

### Consistent services

The `start` command executes 3 steps for each of the services provided by an SDK
application server:

    1. Creation of the service
    2. Registration of the services directly with the server
    3. Start/Stop the service process
  
In the current implementation, [some services](https://github.com/cosmos/cosmos-sdk/blob/f9dc082059d63423f96315ff913a8745c3178a7d/server/grpc/server.go#L14-L42) wrap these steps together in a single function, while
[others](https://github.com/cosmos/cosmos-sdk/blob/f9dc082059d63423f96315ff913a8745c3178a7d/server/api/server.go#L81-L119) implement them as separate functions as the ones above. The lack of a standard approach for
these steps results in difficulty for extensibility as each of these functions are individually
called by the `start` command after checking if the service is enabled or not by the configuration.

### Configuration extensibility

Each service relies on the configuration options defined on `config.tml`. Thus, in order to extend
the services provided by the server, the SDK must handle custom TOML files provided by each
application.

## Decision

The proposed approach standardizes the server and its services, so that the app start clearly states
the 3 steps outlined above. To accomplish this a new `Service` and `Server` interface will be
introduced:

```go
type Service interface {
    RegisterRoutes() bool
    Start(config.ServerConfig) error
    Stop() error
}

type Server interface {
    GetServices() []Service
    RegisterServices() error
    Start(config.ServerConfig) error
    Stop() error
}
```

### Configuration

Since the enablement of a service depends on the application configuration, an additional
`ServerConfig` interface is required so that the current configuration utility functions are still
applicable for the SDK and extensible concrete configurations:

```go
// ServerConfig extends the SDK default configuration TOML
type ServerConfig interface {
    // The SDK config defines the 5 configurations used by default: Base, Telemetry, API, gRPC and State Sync. 
    GetSDKConfig() *Config
}
```

### Start Command

```go
func startInProcess(ctx *Context, clientCtx client.Context, appCreator types.AppCreator) error {
  // ...
  app := appCreator(ctx.Logger, db, traceWriter, ctx.Viper)
  cfg := app.GetConfig(ctx.Viper)
  // ...

  // 1. Create application-specific server
  server := app.NewServer(clientCtx, ctx.Logger, cfg)

  // 2. Register routes from each service 
  // NOTE: routes are registered regardless if the service is enabled or not
  if err := server.RegisterServices(); err != nil {
    return err
  }

  // 3. Start each of the services.
  // NOTE: each service must check if it's enabled via the configuration
  if err := server.Start(cfg); err != nil {
    return err
  }

  defer func() {
    // ...
    // stop all the services
    if err := server.Stop(); err != nil {
      ctx.Logger.Error("failed to stop server", "error", err)
    }
    // ...
  }
  // ...
}
```



## Consequences

> This section describes the resulting context, after applying the decision. All consequences should be listed here, not just the "positive" ones. A particular decision may have positive, negative, and neutral consequences, but all of them affect the team and project in the future.

### Positive

- Standardize all the services provided by the SDK
- Modularize the server so that it allows for custom application-specific services.

### Negative

- Breaking changes to the server and configuration

### Neutral

- Refactor of existing and work-in-progress services (gRPC, Rosetta, etc).

### Backwards Compatibility

> All ADRs that introduce backwards incompatibilities must include a section describing these incompatibilities and their severity. The ADR must explain how the author proposes to deal with these incompatibilities. ADR submissions without a sufficient backwards compatibility treatise may be rejected outright.

This ADR introduces significant changes to the `Server`, and `Config` concrete types due to the
refactor. It also updates as well as to the `servertypes.Application` interface. In particular it
removes all the registration functions which will be handled directly by each service during the app
`start`. Each application `Register...` function will be migrated to the `RegisterRoutes()` function
of an independent `Service`.

```go
type Application interface {
    abci.Application

    NewServer(client.Context,log.Logger, config.ServerConfig)
}
```

## Further Discussions

## Test Cases

## References
