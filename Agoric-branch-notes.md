We have:

- `Agoric` - the head of this branch should be eventually used by the
  [agoric-sdk](https://github.com/Agoric/agoric-sdk).
- [ag0](https://github.com/Agoric/ag0) [Agoric mainnet](https://agoric.com)
  phase 0's [gaiad-equivalent](https://github.com/cosmos/gaia)
- `Agoric-ag0` - should be identical to
  [cosmos/cosmos-sdk](https://github.com/cosmos/cosmos-sdk) + changes needed by
  `ag0`

For new features:

- Create a new development branch off of `Agoric` or `Agoric-ag0`
- Test, review, merge PR with `automerge` label for mergify.io
- Add `backport/ag0` label or `backport/Agoric` to have mergify port the PR to
  the other branch

Upon new cosmos-sdk releases:

- Say we now have `cosmos/cosmos-sdk@v0.43.0-rc0`
- Merge `upstream/v0.43.0-rc0` into `Agoric`, create tag `v0.43.0-rc0.agoric`
- Push tag to origin
- Use that tag in the `agoric-sdk/go.mod` in the `replace` directive for `cosmos-sdk`
- build `agoric-sdk` and do any manual tests locally
- submit a PR against `agoric-sdk` for CI tests to run, then merge
