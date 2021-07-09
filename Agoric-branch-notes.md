We have:

- @zmanian is in charge of branches used by proto-mainnet `gaiad`
- `Agoric-proto` - should be identical to `cosmos/cosmos-sdk` + changes needed
  by proto-mainnet
- `Agoric` - the head of this branch should be eventually used by the
  `agoric-sdk`.  Needs to be kept up-to-date with all agoric-specific patches.

For new features:

- Create a new development branch off of `Agoric-proto` or `Agoric` in `agoric-labs/cosmos-sdk`
- Test, review, merge PR
- Consider creating a rebased PR against upstream (`cosmos/cosmos-sdk`) to merge the feature

Merging `Agoric-proto` back into `Agoric`:

- `git checkout Agoric`
- `git merge Agoric-proto`
- `git push origin Agoric`
- test as if a new cosmos-sdk release

Upon new cosmos-sdk releases:

- Say we now have `cosmos/cosmos-sdk@v0.43.0-rc0`
- Merge `upstream/v0.43.0-rc0` into `Agoric`, create tag `v0.43.0-rc0.agoric`
- Push tag to origin
- Use that tag in the `agoric-sdk/go.mod` in the `replace` directive for `cosmos-sdk`
- build `agoric-sdk` and do any manual tests locally
- submit a PR against `agoric-sdk` for CI tests to run, then merge
