# Upgrading Cosmos SDK v2 [DRAFT, see #22531](https://github.com/cosmos/cosmos-sdk/issues/22531)

This guide provides instructions for upgrading to specific versions of Cosmos SDK.
Note, always read the **SimApp** section for more information on application wiring updates.

## Upgrading from v0.52.x to v2.

First and foremost, v2 uses [depinject](./depinject/README.md) to wire the module and application dependencies.
This guide assumes that you have already made your modules depinject compatible, and that you made use of depinject in your application.

### Modules

### Server


### SimApp

With the migration to server/v2 and runtime/v2 some changes are required in the `root.go` and `app.go` of your application.

#### `app.go`


#### `root.go`

Thanks to the more modular design of v2, the application developer is free to pick which server to use in their application:

```go
// wire server commands
return serverv2.AddCommands[T](
	rootCmd,
	logger,
	simApp,
	deps.GlobalConfig,
	initServerConfig(),
	deps.ConsensusServer,
	grpcServer,
	storeComponent,
	telemetryServer,
	restServer,
	grpcgatewayServer,
)
```

## Upgrading from v0.50.x to v2

Upgrading directly from v0.50.x to v2 is supported.
Modules should be updated to support all the latest changes in the SDK.

Read the module section from the v0.52 [UPGRADING.md](UPGRADING.md) file for more information.
Then simply follow the instructions from the [v0.52 section](#upgrading-from-v052x-to-v200) from this file.
