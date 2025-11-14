# Cosmos SDK for **AtomOne**

Fork of Cosmos SDK **v0.50.x** for **AtomOne**.

Compared to the original Cosmos SDK, this fork includes the following changes:

- Store app version in consensus param store
- Re-add query router for custom ABCI queries
- Add v0.52 helpers to facilitate testing
- Backport improvements for DOS protection in `x/authz`
- Support historical account number queries
- Removal of go modules to keep mono go.mod in monorepo.

See the [CHANGELOG.md](CHANGELOG.md) for a complete list of modifications.
