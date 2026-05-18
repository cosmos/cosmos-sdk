# Remove `depinject` Plan

This document is the execution checklist to remove `depinject` from the repository in small, reviewable PRs.

Detailed Slice 1 symbol map: `docs/internal/remove-depinject-slice1-map.md`.

## Goal

* Remove the `depinject` runtime and package from the SDK codebase.
* Replace module wiring with explicit constructors and manual app wiring.
* Remove `cosmossdk.io/api` usage from module wiring and app-assembly paths, then remove the module dependency from all SDK modules.
* Preserve behavior and test coverage while minimizing blast radius per PR.

## Inventory Snapshot

Current `depinject` usage appears in all major surfaces:

* Core wiring and runtime integration (`runtime/`, `core/appmodule/`, `core/appconfig/`).
* Module implementations in `x/` and `enterprise/` (many `module.go` files and direct `depinject` imports).
* App and test wiring (`simapp/`, `testutil/`, integration and e2e tests).
* Public schemas and docs (`proto/cosmos/app/v1alpha1/*`, docs ADRs and guides).
* Module dependencies (`go.mod` files in root and multiple submodules include `cosmossdk.io/depinject`).

Concrete high-impact code references (non-test first):

* Core/runtime:
    * `runtime/module.go`
    * `core/appmodule/module.go`
    * `core/appmodule/option.go`
    * `core/appmodule/register.go`
    * `core/appconfig/config.go`
* Core modules in `x/`:
    * `x/auth/module.go`
    * `x/auth/vesting/module.go`
    * `x/authz/module/module.go`
    * `x/bank/module.go`
    * `x/consensus/module.go`
    * `x/distribution/module.go`
    * `x/evidence/module.go`
    * `x/feegrant/module/module.go`
    * `x/genutil/module.go`
    * `x/mint/module.go`
    * `x/params/module.go`
    * `x/slashing/module.go`
    * `x/staking/module.go`
    * `x/upgrade/module.go`
* Enterprise/contrib:
    * `enterprise/group/x/group/module/module.go`
    * `contrib/x/circuit/module.go`
    * `contrib/x/crisis/module.go`
    * `contrib/x/nft/module/module.go`
* Test and app wiring helpers:
    * `testutil/network/network.go`
    * `testutil/configurator/configurator.go`
    * `testutil/sims/app_helpers.go`

Notable direct depinject adapter files to retire early:

* `x/gov/depinject.go`
* `x/protocolpool/depinject.go`
* `x/epochs/depinject.go`
* `testutil/x/counter/depinject.go`

## Migration Strategy

* Keep each PR behavior-preserving and independently testable.
* Prefer additive wiring first (manual path), then delete old depinject path.
* Land module-by-module or surface-by-surface to avoid a single giant diff.
* Keep temporary compatibility shims only when they reduce risk for adjacent PRs.

## PR Slices and Owners

### Slice 0: Baseline and Tracking

* Owner: SDK core maintainers (wiring/runtime).
* Create this plan file and keep status/checkpoints updated.
* Add a lightweight CI guard that reports new `cosmossdk.io/depinject` imports (warn-only first).
* Exit criteria: clear migration backlog and no ambiguity about ownership.

### Slice 1: Core Wiring Interface

* Owner: SDK core maintainers (runtime/core).
* Define manual wiring primitives to replace dependency graph resolution.
* Remove `appmodule.Register` and the package `init()` registration flow; move to explicit module descriptor assembly in app wiring.
* Refactor `runtime/module.go`, `core/appmodule/`, and `core/appconfig/` call paths to no longer require depinject container APIs.
* Preserve module registration and ordering semantics.
* Exit criteria: core app wiring works without invoking depinject APIs.

### Slice 2: Module Adapter Conversion (`x/`)

* Owner: module maintainers (staking/gov/auth/bank/etc.).
* Remove per-module depinject adapters and provide explicit constructor/wiring functions in each module package.
* Prioritize modules with dedicated adapter files (`x/gov`, `x/protocolpool`, `x/epochs`), then remaining `module.go` imports.
* Exit criteria: all core `x/` modules build and test without importing `cosmossdk.io/depinject`.

### Slice 3: Enterprise and Contrib Modules

* Owner: enterprise module maintainers.
* Convert `enterprise/group`, `enterprise/poa`, and `contrib/x/*` module wiring away from depinject.
* Keep licensing and module-boundary constraints unchanged.
* Exit criteria: enterprise and contrib module builds/tests pass without depinject imports.

### Slice 4: App and Test Harness Migration

* Owner: simapp and test-infra maintainers.
* Update `simapp`, `testutil/network`, `testutil/sims`, integration/system/e2e helpers to manual wiring.
* Remove depinject-specific scaffolding from test helpers and sample app config usage.
* Exit criteria: `make build`, `make lint`, `make test` pass with no functional regressions.

### Slice 5: Schema, Proto, and Docs Cleanup

* Owner: API/docs maintainers.
* Remove or replace depinject-oriented protobuf and docs references where still user-facing:
    * `proto/cosmos/app/v1alpha1/config.proto`
    * `proto/cosmos/app/v1alpha1/module.proto`
    * `proto/cosmos/app/v1alpha1/query.proto`
* Update ADRs and docs to reflect the new explicit wiring model.
* Exit criteria: docs/proto narrative no longer instructs or implies depinject usage; no new code paths require `cosmossdk.io/api` module config wrappers for wiring.

### Slice 6: Depinject Dependency Removal and Package Deletion

* Owner: SDK core maintainers.
* Remove `cosmossdk.io/depinject` from all `go.mod` files.
* Delete `depinject/` package after all callsites are gone.
* Run full validation in repo order:
    * `make tidy-all`
    * `make build`
    * `make lint`
    * `make test`
* Exit criteria: zero depinject imports and no depinject module dependency in any submodule.

### Slice 7: `cosmossdk.io/api` Dependency Removal

* Owner: SDK core maintainers + API owners.
* Replace any remaining `cosmossdk.io/api/.../module/v1` config structs used only for wiring with native SDK configuration types.
* Migrate helper packages (including test configurators) away from `appconfig.WrapAny` and api module config wrappers.
* Remove `cosmossdk.io/api` from all `go.mod` files where it is no longer required.
* Delete or relocate internal compatibility shims that exist solely to consume api module config types.
* Run full validation in repo order:
    * `make tidy-all`
    * `make build`
    * `make lint`
    * `make test`
* Exit criteria: no runtime or test wiring path depends on `cosmossdk.io/api`; dependency is removed from SDK modules.

## Execution Checklist

* [ ] Slice 0 landed
* [ ] Slice 1 landed
* [ ] Slice 2 landed
* [ ] Slice 3 landed
* [ ] Slice 4 landed
* [ ] Slice 5 landed
* [ ] Slice 6 landed
* [ ] Slice 7 landed

## Risk Notes

* Wiring changes can alter module init order; treat ordering as a compatibility contract.
* App/test harness migration may expose hidden assumptions previously handled by depinject graph resolution.
* Proto/doc changes may need coordinated communication for downstream integrators.
* `cosmossdk.io/api` removal may require coordinated migration guidance for downstream apps still using app config protobuf wrappers.
