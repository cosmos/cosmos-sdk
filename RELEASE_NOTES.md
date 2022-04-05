# Cosmos SDK v0.45.2 Release Notes

This release introduces bug fixes and improvements on the Cosmos SDK v0.45 series:

Highlights:

- Add hooks to allow modules to add things to state-sync. Please see [PR #10961](https://github.com/cosmos/cosmos-sdk/pull/10961) for more information.
- Register [`EIP191`](https://eips.ethereum.org/EIPS/eip-191) as an available `SignMode` for chains to use. Please note that in v0.45.2, the Cosmos SDK does **not** support EIP-191 out of the box. But if your chain wants to integrate EIP-191, it's possible to do so by passing a `TxConfig` with your own sign mode handler which implements EIP-191, using the new provided `authtx.NewTxConfigWithHandler` function.
- Add a new `rollback` CLI command to perform a state rollback by one block. Read more in [PR #11179](https://github.com/cosmos/cosmos-sdk/pull/11179).
- Some new queries were added:
  - x/authz: `GrantsByGrantee` to query grants by grantee,
  - x/bank: `SpendableBalances` to query an account's total (paginated) spendable balances,
  - TxService: `GetBlockWithTxs` to fetch a block along with all its transactions, decoded.
- Some bug fixes, such as:
  - Update the prune `everything` strategy to store the last two heights.
  - Fix data race in store trace component.
  - Fix cgo secp signature verification and update libscep256k1 library.

See the [Cosmos SDK v0.45.2 Changelog](https://github.com/cosmos/cosmos-sdk/blob/v0.45.2/CHANGELOG.md) for the exhaustive list of all changes and check other fixes in the 0.45.x release series.

**Full Commit History**: https://github.com/cosmos/cosmos-sdk/compare/v0.45.1...v0.45.2
