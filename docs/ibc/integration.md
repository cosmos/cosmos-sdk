<!--
order: 2
-->

# Integration

Learn how to integrate IBC to your application and send data packets to other chains. {synopsis}

This document outlines the required steps to integrate and configure the [IBC
module](https://github.com/cosmos/cosmos-sdk/tree/release/v0.42.x/x/ibc) to your Cosmos SDK application and
send fungible token transfers to other chains.

## Integrating the IBC module

Integrating the IBC module to your SDK-based application is straighforward. The general changes can be summarized in the following steps:

- Add required modules to the `module.BasicManager`
- Define additional `Keeper` fields for the new modules on the `App` type
- Add the module's `StoreKeys` and initialize their `Keepers`
- Set up corresponding routers and routes for the `ibc` and `evidence` modules
- Add the modules to the module `Manager`
- Add modules to `Begin/EndBlockers` and `InitGenesis`
- Update the module `SimulationManager` to enable simulations

### Module `BasicManager` and `ModuleAccount` permissions

The first step is to add the following modules to the `BasicManager`: `x/capability`, `x/ibc`,
`x/evidence` and `x/ibc/applications/transfer`. After that, we need to grant `Minter` and `Burner` permissions to
the `ibc-transfer` `ModuleAccount` to mint and burn relayed tokens.

```go
// app.go
var (

  ModuleBasics = module.NewBasicManager(
    // ...
    capability.AppModuleBasic{},
    ibc.AppModuleBasic{},
    evidence.AppModuleBasic{},
    transfer.AppModuleBasic{}, // i.e ibc-transfer module
  )

  // module account permissions
  maccPerms = map[string][]string{
    // other module accounts permissions
    // ...
    ibctransfertypes.ModuleName:    {authtypes.Minter, authtypes.Burner},
)
```

### Application fields

Then, we need to register the `Keepers` as follows:

```go
// app.go
type App struct {
  // baseapp, keys and subspaces definitions

  // other keepers
  // ...
  IBCKeeper        *ibckeeper.Keeper // IBC Keeper must be a pointer in the app, so we can SetRouter on it correctly
  EvidenceKeeper   evidencekeeper.Keeper // required to set up the client misbehaviour route
  TransferKeeper   ibctransferkeeper.Keeper // for cross-chain fungible token transfers

  // make scoped keepers public for test purposes
  ScopedIBCKeeper      capabilitykeeper.ScopedKeeper
  ScopedTransferKeeper capabilitykeeper.ScopedKeeper

  /// ...
  /// module and simulation manager definitions
}
```

### Configure the `Keepers`

During initialization, besides initializing the IBC `Keepers` (for the  `x/ibc`, and
`x/ibc/applications/transfer` modules), we need to grant specific capabilities through the capability module
`ScopedKeepers` so that we can authenticate the object-capability permissions for each of the IBC
channels.

```go
func NewApp(...args) *App {
  // define codecs and baseapp

  // add capability keeper and ScopeToModule for ibc module
  app.CapabilityKeeper = capabilitykeeper.NewKeeper(appCodec, keys[capabilitytypes.StoreKey], memKeys[capabilitytypes.MemStoreKey])

  // grant capabilities for the ibc and ibc-transfer modules
  scopedIBCKeeper := app.CapabilityKeeper.ScopeToModule(ibchost.ModuleName)
  scopedTransferKeeper := app.CapabilityKeeper.ScopeToModule(ibctransfertypes.ModuleName)

  // ... other modules keepers

  // Create IBC Keeper
  app.IBCKeeper = ibckeeper.NewKeeper(
  appCodec, keys[ibchost.StoreKey], app.StakingKeeper, scopedIBCKeeper,
  )

  // Create Transfer Keepers
  app.TransferKeeper = ibctransferkeeper.NewKeeper(
    appCodec, keys[ibctransfertypes.StoreKey],
    app.IBCKeeper.ChannelKeeper, &app.IBCKeeper.PortKeeper,
    app.AccountKeeper, app.BankKeeper, scopedTransferKeeper,
  )
  transferModule := transfer.NewAppModule(app.TransferKeeper)

  // Create evidence Keeper for to register the IBC light client misbehaviour evidence route
  evidenceKeeper := evidencekeeper.NewKeeper(
    appCodec, keys[evidencetypes.StoreKey], &app.StakingKeeper, app.SlashingKeeper,
  )

  // .. continues
}
```

### Register `Routers`

IBC needs to know which module is bound to which port so that it can route packets to the
appropriate module and call the appropriate callbacks. The port to module name mapping is handled by
IBC's port `Keeper`. However, the mapping from module name to the relevant callbacks is accomplished
by the port
[`Router`](https://github.com/cosmos/cosmos-sdk/tree/release/v0.42.x/x/ibc//core/05-port/types/router.go) on the
IBC module.

Adding the module routes allows the IBC handler to call the appropriate callback when processing a
channel handshake or a packet.

The second `Router` that is required is the evidence module router. This router handles general
evidence submission and routes the business logic to each registered evidence handler. In the case
of IBC, it is required to submit evidence for [light client
misbehaviour](https://github.com/cosmos/ics/tree/master/spec/ics-002-client-semantics#misbehaviour)
in order to freeze a client and prevent further data packets from being sent/received.

Currently, a `Router` is static so it must be initialized and set correctly on app initialization.
Once the `Router` has been set, no new routes can be added.

```go
// app.go
func NewApp(...args) *App {
  // .. continuation from above

  // Create static IBC router, add ibc-tranfer module route, then set and seal it
  ibcRouter := port.NewRouter()
  ibcRouter.AddRoute(ibctransfertypes.ModuleName, transferModule)
  // Setting Router will finalize all routes by sealing router
  // No more routes can be added
  app.IBCKeeper.SetRouter(ibcRouter)

  // create static Evidence routers

  evidenceRouter := evidencetypes.NewRouter().
    // add IBC ClientMisbehaviour evidence handler
    AddRoute(ibcclient.RouterKey, ibcclient.HandlerClientMisbehaviour(app.IBCKeeper.ClientKeeper))

  // Setting Router will finalize all routes by sealing router
  // No more routes can be added
  evidenceKeeper.SetRouter(evidenceRouter)

  // set the evidence keeper from the section above
  app.EvidenceKeeper = *evidenceKeeper

  // .. continues
```

### Module Managers

In order to use IBC, we need to add the new modules to the module `Manager` and to the `SimulationManager` in case your application supports [simulations](./../building-modules/simulator.md).

```go
// app.go
func NewApp(...args) *App {
  // .. continuation from above

  app.mm = module.NewManager(
    // other modules
    // ...
    capability.NewAppModule(appCodec, *app.CapabilityKeeper),
    evidence.NewAppModule(app.EvidenceKeeper),
    ibc.NewAppModule(app.IBCKeeper),
    transferModule,
  )

  // ...

  app.sm = module.NewSimulationManager(
    // other modules
    // ...
    capability.NewAppModule(appCodec, *app.CapabilityKeeper),
    evidence.NewAppModule(app.EvidenceKeeper),
    ibc.NewAppModule(app.IBCKeeper),
    transferModule,
  )

  // .. continues
```

### Application ABCI Ordering

One addition from IBC is the concept of `HistoricalEntries` which are stored on the staking module.
Each entry contains the historical information for the `Header` and `ValidatorSet` of this chain which is stored
at each height during the `BeginBlock` call. The historical info is required to introspect the
past historical info at any given height in order to verify the light client `ConsensusState` during the
connection handhake.

The IBC module also has
[`BeginBlock`](https://github.com/cosmos/cosmos-sdk/tree/release/v0.42.x/x/ibc/core/02-client/abci.go) logic as
well. This is optional as it is only required if your application uses the [localhost
client](https://github.com/cosmos/ics/blob/master/spec/ics-009-loopback-client) to connect two
different modules from the same chain.

::: tip
Only register the ibc module to the `SetOrderBeginBlockers` if your application will use the
localhost (_aka_ loopback) client.
:::

```go
// app.go
func NewApp(...args) *App {
  // .. continuation from above

  // add evidence, staking and ibc modules to BeginBlockers
  app.mm.SetOrderBeginBlockers(
    // other modules ...
    evidencetypes.ModuleName, stakingtypes.ModuleName, ibchost.ModuleName,
  )

  // ...

  // NOTE: Capability module must occur first so that it can initialize any capabilities
  // so that other modules that want to create or claim capabilities afterwards in InitChain
  // can do so safely.
  app.mm.SetOrderInitGenesis(
    capabilitytypes.ModuleName,
    // other modules ...
    ibchost.ModuleName, evidencetypes.ModuleName, ibctransfertypes.ModuleName,
  )

  // .. continues
```

::: warning
**IMPORTANT**: The capability module **must** be declared first in `SetOrderInitGenesis`
:::

That's it! You have now wired up the IBC module and are now able to send fungible tokens across
different chains. If you want to have a broader view of the changes take a look into the SDK's
[`SimApp`](https://github.com/cosmos/cosmos-sdk/tree/release/v0.42.x/simapp/app.go).

## Next {hide}

Learn about how to create [custom IBC modules](./custom.md) for your application {hide}
