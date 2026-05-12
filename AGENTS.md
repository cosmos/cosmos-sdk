# Repository guidelines for coding agents

This file provides guidance for automated agents contributing to this repository.

## Repository Overview

- This is the **Cosmos SDK**, a modular blockchain SDK for building application-specific blockchains.
- The project is written in Go. Key areas:
  - **`x/`** — Core SDK modules (auth, bank, staking, gov, distribution, consensus, etc.) and supplementary modules (authz, feegrant, epochs, etc.)
  - **`baseapp/`** — Base application and ABCI handling
  - **`client/`** — CLI and client utilities
  - **`store/`** — State storage (IAVL, multistore)
  - **`crypto/`** — Key types and signing
  - **`simapp/`** — Demo application (simd) used for testing and as a reference
  - **`contrib/x/`** — Deprecated modules (circuit, crisis, nft); not actively maintained and not covered by the Bug Bounty program
  - **`enterprise/`** — Enterprise modules (poa, group) with different licensing; see `enterprise/README.md`. (`x/group` moved here from `contrib/` in v0.54.0.)
- Most unit tests live next to the code they test. The `tests/` directory holds cross-cutting tests: `tests/integration/` (integration), `tests/e2e/` (e2e), `tests/systemtests/` (system), `tests/fuzz/` (fuzz).
- Protobuf definitions live under `proto/`.
- Agents should **not** run `test-e2e`, `test-sim-*`, or `test-system` by default — they are slow, often need Docker or pre-built binaries, and depend on environment specifics. Run them only when a change explicitly requires that level of coverage.

## Submodules

Each of the following directories has its own `go.mod` and must be tidied / tested independently:

- Library submodules: `api`, `client/v2`, `collections`, `core`, `depinject`, `errors`, `log`, `math`, `store`
- App / test / tools: `simapp`, `tests`, `tests/systemtests`, `tools/cosmovisor` (built via `make cosmovisor`), `tools/confix` (built via `make confix`), `tools/systemtests`
- Enterprise: `enterprise/poa`, `enterprise/poa/simapp`, `enterprise/group`, `enterprise/group/simapp`

Running `go test ./...` from the repo root only tests the root module. Use `make test` so the `run-tests` target walks every submodule.

## Documentation

- **API & tutorials:** https://docs.cosmos.network/
- **Architecture:** `docs/architecture/` — follow existing patterns when adding features
- **Module overview:** `x/README.md` — lists all modules and their purposes
- **Upgrade notes:** `UPGRADING.md` — required reading before introducing a breaking change; add a corresponding entry alongside the changelog note
- **Security:** `SECURITY.md` — report security-sensitive findings here, not in a normal PR

## Scope

- **IBC** is maintained in a separate repo ([ibc-go](https://github.com/cosmos/ibc-go)); do not add IBC logic to the Cosmos SDK.
- **Deprecated modules** in `contrib/x/` (circuit, crisis, nft) — avoid adding new features; prefer core `x/` modules.

## Agent guardrails

- **State-machine and consensus code requires extra care.** Changes that affect deterministic state must use the `State Machine Breaking` changelog stanza and usually warrant an `UPGRADING.md` entry.
- **Do not hand-edit generated files.** `*.pb.go`, `*.pb.gw.go`, mocks under `testutil/`, and similar generated output should be regenerated via `make proto-all` / `make mocks`.
- **Do not edit `go.sum`, `go.work`, or `go.work.sum` manually** — use `make tidy-all` to update dependencies across all modules.
- **Do not run `golangci-lint` directly with custom paths** — use `make lint`, which invokes `./scripts/go-lint-all.bash --timeout=15m` to match CI.
- **Enterprise licensing** — modules under `enterprise/` use different licenses; do not copy code into other parts of the SDK without checking.

## Development Workflow

1. **Pre-push verification.** From the repo root, run in order:
   - `make tidy-all` (only if dependencies changed)
   - `make build`
   - `make lint` (depends on `make lint-install`; the first run will install the linter)
   - `make test` (or `make test-unit`)

   Fix any failures locally before pushing.
2. **Protobuf changes.** Run `make proto-all` (format, lint, generate). Requires Docker and the proto-builder image. After proto changes, run `make build` and `make test` to ensure nothing is broken.
3. **Mocks.** If you add or change interfaces that require mocks, run `make mocks` before pushing.
4. **Lint fixes.** Run `make lint-fix` to auto-fix lint issues.
5. **Enterprise tests.** When changing enterprise modules, also run `make enterprise-all-test` (or `make enterprise-poa-test` / `make enterprise-group-test`).

`make all` (`tools build lint test vulncheck`) is the full local check matching CI; use it before requesting review on a large change.

## Changelog

- Add entries to `CHANGELOG.md` under the `## UNRELEASED` section.
- Format: `* (tag) [#PR-number](https://github.com/cosmos/cosmos-sdk/pull/PR-number) message`
  - Example: `* (x/staking) [#12345](https://github.com/cosmos/cosmos-sdk/pull/12345) Fix validator power calculation`
- Tags indicate the affected area: `(x/bank)`, `(store)`, `(baseapp)`, `(enterprise/poa)`, etc.
- Stanzas (use the appropriate one): `Features`, `Improvements`, `Deprecated`, `Bug Fixes`, `Breaking Changes` (top-level umbrella), `Client Breaking`, `CLI Breaking`, `API Breaking`, `State Machine Breaking`.
- For breaking changes, also add an entry to `UPGRADING.md`.
- See the header of `CHANGELOG.md` for full formatting rules.

## Commit Messages

- Follow the [Conventional Commits](https://www.conventionalcommits.org/) specification. Common types: `feat`, `fix`, `docs`, `test`, `deps`, `chore`.
- Include the proposed commit message in the pull request description when opening a PR.

## Pull Request Workflow

Before requesting review:

- [ ] Pre-push verification passes (see **Development Workflow**)
- [ ] Changelog entry added under `## UNRELEASED` for user-facing or notable changes
- [ ] `UPGRADING.md` entry added for breaking changes
- [ ] CI is green (build, lint, tests)
- [ ] PR is focused — split unrelated changes into separate PRs

## Enterprise Modules

- Located in `enterprise/` (e.g. `enterprise/poa`, `enterprise/group`).
- Each has its own `go.mod` and Makefile. Use `make enterprise-<module>-<target>` or `make enterprise-all-<target>`.
- **Different licensing** — review the LICENSE file in each enterprise module before use or copying code.
- Tests: `make enterprise-poa-test` or `make enterprise-group-test`.
