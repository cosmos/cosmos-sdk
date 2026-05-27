# Slice 1 Symbol-Level Migration Map (`runtime` + `core/appmodule`)

This document maps each `depinject` callsite in Slice 1 to a concrete manual-wiring replacement.

Long-term target note: this migration also targets eventual removal of `cosmossdk.io/api` from wiring paths, but that work is tracked as a later slice after depinject/runtime deletion.

## Current status

* Completed in this branch:
    * `runtime/module.go` no longer imports `cosmossdk.io/depinject`.
    * `AppInputs` is now an alias of explicit `AppBuilderInputs` (no embedded marker type).
    * Store-key and store-service assembly now has explicit module-name APIs:
        * `KVStoreKeyForModule`
        * `TransientStoreKeyForModule`
        * `MemoryStoreKeyForModule`
        * `ObjectStoreKeyForModule`
        * corresponding `*ServiceForModule` helpers.
    * Compatibility wrappers (`Provide*`) now rely on a local `ModuleNameProvider` interface instead of `depinject.ModuleKey`.
    * `AddressCodecInputs` no longer embeds `depinject.In`.
    * `testutil/network` now keeps the explicit fixture path in `network.go` plus one remaining root-module legacy helper constructor (no separate legacy shim file).
    * First depinject-network caller migrated to explicit fixture path:
        * `tests/integration/server/grpc/out_of_gas_test.go`.
    * `network.MinimumAppConfig()` callsites and definition removed from the repository.
    * `tests/e2e/distribution` now uses explicit fixtures for `externalPoolEnabled=true`; the `false` branch still uses transitional depinject app config.
    * Transitional depinject bridge/shim files have been removed and inlined into module/runtime wiring:
        * `runtime/depinject_bridge.go` deleted
        * all `x/*/depinject_legacy.go` files deleted
        * `testutil/network/depinject_legacy.go` deleted
    * Common root-module integration tests call a single centralized legacy helper:
        * `network.DefaultConfigWithLegacyRootModules()`
      so depinject app-config assembly is isolated to one `testutil/network` callsite for those suites.
    * `x/auth/types`, `x/distribution/client/cli`, and `tests/e2e/distribution` now use explicit fixture configs (no dedicated auth/distribution legacy helper constructors remain).
    * `DefaultConfigWithAppConfig(...)` and `DefaultConfigWithAppConfigWithQueryGasLimit(...)` have been removed from `testutil/network`; depinject assembly is now inlined only inside `DefaultConfigWithLegacyRootModules()`.
    * Direct depinject callsite removed from `tests/e2e/auth/suite.go` (`TestGetBroadcastCommandWithoutOfflineFlag`) by reusing explicit `network.DefaultConfig(simapp.NewTestNetworkFixture).TxConfig`.
    * Direct depinject callsite removed from `simapp/app_test.go` (`TestAddressCodecFactory`) by exercising `runtime.ProvideAddressCodec` directly.
    * Added `simtestutil.SetupWithNopLogger(...)` and `simtestutil.SetupAtGenesisWithNopLogger(...)` to centralize common `depinject.Supply(log.NewNopLogger())` wiring.
    * Migrated direct depinject boilerplate callsites to the new simtestutil helpers:
        * `x/auth/module_test.go`
        * `x/mint/module_test.go`
        * `x/staking/module_test.go`
        * `x/slashing/abci_test.go`
        * `tests/integration/distribution/module_test.go`
        * `tests/integration/gov/module_test.go`
        * `tests/integration/runtime/query_test.go`
        * `x/bank/app_test.go`
        * `x/staking/app_test.go`
        * `x/slashing/app_test.go`
        * `client/grpc_query_test.go` (via `simtestutil.InjectWithNopLogger`)
        * `tests/integration/gov/genesis_test.go`
        * `types/query/pagination_test.go`
        * `x/auth/keeper/keeper_bench_test.go`
        * `tests/integration/tx/context_test.go` (retains depinject only for `ProvideCustomGetSigners`, but no direct nop-logger supply callsite)
        * `baseapp/block_gas_test.go` (via `simtestutil.InjectWithNopLogger`)
        * `baseapp/msg_service_router_test.go` (`TestMsgService`, via `simtestutil.InjectWithNopLogger`)
        * `x/slashing/keeper/slash_redelegation_test.go`
        * `x/gov/simulation/operations_test.go`
        * `x/distribution/simulation/operations_test.go`
        * `x/slashing/simulation/operations_test.go`
    * Targeted validation passes again after bridge updates:
        * `go test ./server/grpc -run TestIntegrationTestSuite`
        * `go test ./server/api -run TestGRPCWebTestSuite`
        * `go test ./client/rpc -run TestIntegrationTestSuite`
        * `go test ./client/grpc/cmtservice`
        * `go test ./x/auth/types -run TestAccountRetriever`
        * `go test ./x/distribution/client/cli -run TestCLITestSuite`

## Scope

* In scope:
    * `runtime/module.go`
    * `core/appmodule/module.go`
    * `core/appmodule/option.go`
    * `core/appmodule/register.go`
    * `core/appconfig/config.go`
* Out of scope for this slice:
    * Rewriting all `x/*` module provider functions (`ModuleInputs`, `ModuleOutputs`, `ProvideModule`) at once.
    * Full deletion of `depinject/` package (handled in later slices).

## Runtime Callsite Map

### `runtime/module.go`

* `appmodule.Register(..., appmodule.Provide(...), appmodule.Invoke(...))`
    * Current role: registers runtime providers/invokers into depinject appconfig registry.
    * Replacement: register a runtime module descriptor in a local registry that exposes explicit hooks:
        * `BuildRuntime(ctx, cfg, deps) -> RuntimeArtifacts`
        * `FinalizeRuntime(app, artifacts) error`
* `type BaseAppOption func(*baseapp.BaseApp)` + `IsManyPerContainerType()`
    * Current role: marks `BaseAppOption` as depinject multi-bind.
    * Replacement: plain `[]BaseAppOption` collected through explicit app wiring options (no marker method).
* `type AppInputs struct { depinject.In ... }`
    * Current role: depinject struct-based input aggregation.
    * Replacement: explicit constructor arguments object with no embedded marker:
        * `type SetupAppInputs struct { ... }`
        * built by runtime wiring code, not by dependency graph reflection.
* `SetupAppBuilder(inputs AppInputs)`
    * Current role: invoked by depinject as an invoker after providers run.
    * Replacement: direct call from runtime bootstrap flow:
        * `func (b *Bootstrap) SetupAppBuilder(inputs SetupAppInputs) error`
* `ProvideKVStoreKey/ProvideTransientStoreKey/ProvideMemoryStoreKey/ProvideObjectStoreKey` with `depinject.ModuleKey`
    * Current role: module-scoped store-key provider keyed by calling module.
    * Replacement: explicit per-module allocation loop:
        * iterate configured modules deterministically (`[]ModuleDescriptor` order)
        * call `register*StoreKey(moduleName string, cfg *runtimev1alpha1.Module, app *AppBuilder)`
* `ProvideKVStoreService/ProvideMemoryStoreService/ProvideTransientStoreService`
    * Current role: wrap module-scoped keys as store services.
    * Replacement: create services directly at module-instantiation time from allocated keys.
* `type AddressCodecInputs struct { depinject.In ... }` and `ProvideAddressCodec(in AddressCodecInputs)`
    * Current role: optional dependency resolution via depinject.
    * Replacement: explicit runtime config resolver:
        * `ResolveAddressCodecs(authCfg, stakingCfg, factories) (address.Codec, ValidatorAddressCodec, ConsensusAddressCodec)`
* `ProvideApp(...)`
    * Current role: root provider returning codec/router/app artifacts into container.
    * Replacement: direct constructor:
        * `NewRuntimeArtifacts(interfaceRegistry) (RuntimeArtifacts, error)`
        * artifacts passed explicitly to module builders.

## Core Callsite Map

### `core/appmodule/module.go`

* `AppModule interface { depinject.OnePerModuleType; IsAppModule() }`
    * Current role: marks app modules as one-per-module graph outputs.
    * Replacement: remove depinject marker; keep explicit module identity in registry map keys:
        * `AppModule interface { IsAppModule() }`
        * module uniqueness enforced by `map[string]appmodule.AppModule` insertion checks.

### `core/appmodule/option.go`

* `type Option = depinjectappconfig.Option`
* `var Provide = depinjectappconfig.Provide`
* `var Invoke = depinjectappconfig.Invoke`
    * Current role: thin aliases to depinject module registration primitives.
    * Replacement: define native appmodule options:
        * `type Option interface{ apply(*ModuleRegistration) error }`
        * `Provide(...)` replaced by explicit builder registration entries.
        * `Invoke(...)` replaced by explicit post-build hooks on module registration.

### `core/appmodule/register.go`

* `var Register = depinjectappconfig.RegisterModule`
    * Current role: writes module initializer into depinject global registry via package `init()` side effects.
    * Replacement: remove `appmodule.Register` entirely.
        * delete global module registry and `init()` registration pattern.
        * each module package exports an explicit descriptor/builder value.
        * app wiring/bootstrap code owns the final ordered module list.

### `core/appconfig/config.go`

* `LoadJSON`, `LoadYAML`, `WrapAny`, `Compose` aliases to depinject appconfig.
    * Current role: produce `depinject.Config` for later `depinject.Inject`.
    * Replacement:
        * keep `LoadJSON`, `LoadYAML`, `WrapAny` but return native app config representation (`*appv1alpha1.Config` or typed wrapper).
        * replace `Compose` with deterministic build-plan assembly:
            * `Compose(appCfg) (BuildPlan, error)`
            * `BuildPlan` contains ordered module configs, bindings, and resolved runtime config.

## Transitional API Targets (Slice 1)

* `core/appmodule`:
    * remove global registration API:
        * delete `Register` and `register.go`.
        * remove dependency on package-level side effects.
    * add explicit descriptor contract used by app wiring:
        * `type ModuleDescriptor struct { Name; ConfigType; Build; PostBuildHooks }`
* `core/appconfig`:
    * replace depinject-backed `Compose` internals with build-plan generation.
* `runtime`:
    * add bootstrap entrypoint that consumes `BuildPlan` and returns `*runtime.AppBuilder`.
    * move store key provisioning from `depinject.ModuleKey` to explicit module-name loops.

## First PR Implementation Sequence (Slice 1A)

* Step 1: Remove `appmodule.Register` and introduce explicit module descriptor exports in `core/appmodule`.
* Step 2: Port `runtime/module.go` from registration side effects to explicit descriptor consumption.
* Step 3: Add runtime bootstrap scaffolding that performs:
    * runtime artifact creation (`codec`, routers, app shell)
    * module map assembly
    * `SetupAppBuilder` invocation without depinject.
* Step 4: Add temporary compatibility bridge at app wiring layer (not in `core/appmodule`) so existing modules can migrate incrementally.

## Validation for Slice 1A PR

* `go test ./core/... ./runtime/...`
* `make build`
* `make lint`
* spot-check one config-driven app wiring path in tests (e.g. `testutil/configurator`).
