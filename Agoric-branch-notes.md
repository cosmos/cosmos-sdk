We use the following branches:

- `Agoric` - the trunk for our cosmos-sdk fork, used by [agoric-sdk](https://github.com/Agoric/agoric-sdk).
    - New features which we plan to use in a near-future release are added here.
    - Most releases are tagged here.
- `Agoric-vX.Y.Z-alpha.agoric.A.x` - a lazily-created release branch.
    - Branched off of release tag `vX.Y.Z-alpha.agoric.A`.
    - Created when we need a new release that omits some changes present on trunk.
    - Prefer to contain only cherry-picks from trunk.
- Pattern can be extended to sub-sub-branches should the need arise.

Tag convention:

- `vX.Y.Z-alpha.agoric.A` for tagged releases on the `Agoric` branch:
    - `vX.Y.Z` is the version of the latest integrated cosmos/cosmos-sdk release.
    - The `A` is an increasing sequence resetting to `1` after the integration of a new cosmos/cosmos-sdk release.
- `vX.Y.Z-alpha.agoric.A.B` for "patch" releases on branch `Agoric-vX.Y.Z-alpha.agoric.A.x`.
    - `B` starts at 1 and increases sequentially.
- Pattern can be extended to releases on sub-sub-branches should the need arise.

For new features:

- New features should be landed on `Agoric` first, then cherry-picked to a release branch as needed.
- Create a new development branch off of `Agoric`.
- Test, review, merge PR with `automerge` label for mergify.io
    - Don't forget to update `CHANGELOG-Agoric.md` with the change and a link to its PR.

Upon new cosmos-sdk releases, see [upgrading the Interchain Stack](https://github.com/Agoric/agoric-sdk/wiki/Upgrading-the-Interchain-Stack).

Historical work:
- [ag0](https://github.com/Agoric/ag0) [Agoric mainnet](https://agoric.com)
  phase 0's [gaiad-equivalent](https://github.com/cosmos/gaia)
- `Agoric-ag0` - should be identical to
  [cosmos/cosmos-sdk](https://github.com/cosmos/cosmos-sdk) + changes needed by
  `ag0`
