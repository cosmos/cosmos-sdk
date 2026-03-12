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
  - **`contrib/`** — Additional modules (circuit, crisis, nft) — deprecated, not actively maintained
  - **`enterprise/`** — Enterprise modules (poa, group) with different licensing; see `enterprise/README.md`
- **Submodules** (each with its own `go.mod`): `client/v2`, `core`, `depinject`, `errors`, `math`, `collections`, `store`, `log`, `api`, `tools/cosmovisor`, `tools/confix`, `simapp`, `enterprise/*`
- Protobuf definitions live under `proto/`.
- Tests live throughout the repo. Unit and integration tests in `tests/`; system tests in `tests/systemtests/`; e2e tests in `tests/e2e/`.
- Agents are **not** expected to run `test-e2e`, `test-sim-*`, or `test-system` by default — these are slow and environment-sensitive.

## Documentation

- **API & tutorials:** https://docs.cosmos.network/
- **Architecture:** `docs/architecture/` — follow existing patterns when adding features
- **Module overview:** `x/README.md` — lists all modules and their purposes

## Scope

- **IBC** is maintained in a separate repo ([ibc-go](https://github.com/cosmos/ibc-go)); do not add IBC logic to the Cosmos SDK
- **Deprecated modules** in `contrib/` (circuit, crisis, nft) — avoid adding new features; prefer core `x/` modules
- **Enterprise modules** in `enterprise/` have different licensing — see `enterprise/README.md` before use

## Common Mistakes

- **Do not edit `go.sum` manually** — use `make tidy-all` to manage dependencies
- **Do not run `golangci-lint` directly** with custom paths — use `make lint` to match CI
- **Enterprise licensing** — modules under `enterprise/` use different licenses; do not copy code without checking

## Development Workflow

1. **Pre-push verification**
   - Before committing or pushing, run in order: `make tidy-all`, `make build`, `make lint`, `make test` (or `make test-unit`)
   - Fix any failures locally before pushing
   - Use `make lint` (or repo equivalent) — never run `golangci-lint` directly with custom paths; use `./scripts/go-lint-all.bash` via the Makefile to match CI
2. **Protobuf changes**
   - Run `make proto-all` (format, lint, generate). Requires Docker and the proto-builder image.
   - After proto changes, run `make build` and `make test` to ensure nothing is broken.
3. **Formatting and linting**
   - Run `make lint` to lint all modules
   - Run `make lint-fix` to automatically fix lint issues
4. **Testing**
   - `make test` or `make test-unit` — runs unit tests across all submodules
   - Tests run per-submodule (each directory with `go.mod`); `run-tests` iterates over them
   - For enterprise modules: `make enterprise-all-test` or `make enterprise-poa-test` / `make enterprise-group-test`
   - Do **not** run `make test-e2e`, `make test-sim-*`, or `make test-system` unless explicitly needed
5. **After making changes to dependencies**
   - Run `make tidy-all` to tidy dependencies across all modules.
6. **Mocks**
   - If you add or change interfaces that require mocks, run `make mocks` before pushing.

## Changelog

- Add entries to `CHANGELOG.md` under the `## UNRELEASED` section.
- Format: `* (tag) #issue-number message` — e.g. `* (x/staking) #12345 Fix validator power calculation`
- Tags indicate the affected area: `(x/bank)`, `(store)`, `(baseapp)`, `(enterprise/poa)`, etc.
- Use the appropriate stanza: Features, Improvements, Bug Fixes, Breaking Changes, Client Breaking, API Breaking, State Machine Breaking, etc.
- See the header of `CHANGELOG.md` for full formatting rules.

## Commit Messages

- Follow the [Conventional Commits](https://www.conventionalcommits.org/) specification. Common types: `feat`, `fix`, `docs`, `test`, `deps`, `chore`.
- Include the proposed commit message in the pull request description when opening a PR.

## Pull Request Workflow

- Add a changelog entry to `CHANGELOG.md` under `## UNRELEASED` for user-facing or notable changes
- Ensure CI passes (build, lint, tests) before requesting review
- Keep PRs focused — smaller changes are easier to review

## Enterprise Modules

- Located in `enterprise/` (e.g. `enterprise/poa`, `enterprise/group`).
- Each has its own `go.mod` and Makefile. Use `make enterprise-<module>-<target>` or `make enterprise-all-<target>`.
- **Different licensing** — review the LICENSE file in each enterprise module before use.
- When changing enterprise modules, run their specific tests: `make enterprise-poa-test` or `make enterprise-group-test`.

## Tools

- **Cosmovisor** — `make cosmovisor` (in `tools/cosmovisor`)
- **Confix** — `make confix` (in `tools/confix`)

## Full verification

For a full local check (matches `make all`): `make tools build lint test vulncheck`