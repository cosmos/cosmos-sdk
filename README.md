# Cosmos SDK Fork

Fork of Cosmos SDK **v0.50.x** for **AtomOne**.

---

## Overview

This fork is based on the official [Cosmos SDK](https://github.com/cosmos/cosmos-sdk) and includes specific changes
required by AtomOne.  
We now depend directly on the upstream SDK for most modules — **local copies of SDK packages have been removed** to
reduce maintenance overhead and ensure alignment with upstream development.

---

## Key Differences from Upstream

Compared to the original Cosmos SDK, this fork includes the following changes:

- Store app version in consensus param store
- Re-add query router for custom ABCI queries
- Add v0.52 helpers to facilitate testing
- Backport improvements for DOS protection in `x/authz`
- Support historical account number queries

See the [CHANGELOG.md](CHANGELOG.md) for a complete list of modifications.

---

## Repository Cleanup

As of [PR #22](https://github.com/atomone-hub/cosmos-sdk/pull/22), all local workspaces that duplicated upstream SDK
packages were removed.  
This change simplifies the project structure and ensures the fork remains lightweight and maintainable.

### Future Modifications

If you need to modify any of the removed modules in the future:

1. **Fork** the specific module from `cosmos/cosmos-sdk`.
2. **Apply** your local changes in this repository.
3. **Update** the import paths accordingly.

---
