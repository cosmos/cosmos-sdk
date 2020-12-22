# ADR 005: Modular Server

## Changelog

- 2020/12/22: Initial Draft

## Author

- Federico Kunze ([@fedekunze](https://github.com/fedekunze))

## Status

Draft, Not Implemented

## Abstract

This document outlines a modular and standard approach for using servers on the SDK, so that applications can also run custom services besides the default ones provided by the SDK.

## Context

The current `Server` implementation allows for applications to register routes, API and gRPC services to the SDK. The problem with the current approach is that the SDK uses a concrete type, which prevents other SDK-based blockchains to extend the services provided by the application without forking the application `start` command (eg: `<app>d start`).

<!-- TODO: extensibility of services -->

### Inconsistent service implementation

The `start` command can be summarized in 3 steps for each of the services provided by an SDK application server:

    1. Creation of the service
    2. Registration of the services directly with the server
    3. Start/Stop the service process

### Configuration extensibility

Each service relies on the configuration options defined on `config.tml`. Thus, in order to extend the services provided by the server, the SDK must handle custom TOML files provided by each application. This proposal also extends the config utility function to return an interface that extends the 5 configurations for the SDK: Base, telemetry, API, gRPC and State Sync.

```go
// ServerConfig extends the SDK default configuration TOML
type ServerConfig interface {
    GetSDKConfig() *Config
}
```

## Decision

A new interface for service will be defined:

```go
type Service interface {
    RegisterRoutes() bool
    Start(config.ServerConfig) error
    Stop() error
}
```

The `Server` will be refactored to include a slice of services:

```go
type Server interface {
    GetServices() []Service
    RegisterRoutes() bool
    Start(config.ServerConfig) error
    Stop() error
}
```

### Configuration

```go
type Generator func(cfg *Config) ServerConfig
```


### Start Command

```go
func startInProcess(ctx *Context, clientCtx client.Context, appCreator types.AppCreator) error {
  // ...
  app := appCreator(ctx.Logger, db, traceWriter, ctx.Viper)
  cfg := config.GetConfig(ctx.Viper, app.ConfigGenerator())
  // ...

  server := app.NewServer(clientCtx, ctx.Logger, cfg)

  // NOTE: routes are registered regardless if the service is enabled or not
  server.RegisterRoutes()

  // start each of the services.
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


### Backwards Compatibility

> All ADRs that introduce backwards incompatibilities must include a section describing these incompatibilities and their severity. The ADR must explain how the author proposes to deal with these incompatibilities. ADR submissions without a sufficient backwards compatibility treatise may be rejected outright.


### Positive

- Standardize all the services provided by the SDK
- Modularize the server so that it allows for custom application-specific services.

### Negative

- Breaking changes to the server and configuration

### Neutral

{neutral consequences}


## Further Discussions

While an ADR is in the DRAFT or PROPOSED stage, this section should contain a summary of issues to be solved in future iterations (usually referencing comments from a pull-request discussion).
Later, this section can optionally list ideas or improvements the author or reviewers found during the analysis of this ADR.


## Test Cases [optional]

Test cases for an implementation are mandatory for ADRs that are affecting consensus changes. Other ADRs can choose to include links to test cases if applicable.


## References

- {reference link}
