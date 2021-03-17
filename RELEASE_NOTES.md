# Cosmos SDK v0.42.2 "Stargate" Release Notes

This maintenance release includes various bugfixes and performance improvements, and it does not introduce any breaking changes.

See the [Cosmos SDK v0.42.2 milestone](https://github.com/cosmos/cosmos-sdk/milestone/41?closed=1) on our issue tracker for further details.

### Keyring UX improvement

A number of macOS [users have reported](https://github.com/cosmos/cosmos-sdk/issues/8809) that their operating system's `keychain` prompt them for password to unlock the
keyring when using the `os` backend before executing any action. This release includes a small fix that automatically
adjusts applications keyring trust so that users are prompted for password only once when the keyring is unlocked.

### Tx search results support for order-by

Although the Tendermint Core's RPC `tx_search` endpoint has been supporting an order-by parameter for quite some time now,
the Cosmos SDK was in fact preventing the application to customise such parameter by automatically setting requests' order-by to "".
This releases introduces [the relevant order-by parameter support](https://github.com/cosmos/cosmos-sdk/issues/8686) when searching through Txs.

### Multisig accounts and v0.40 genesis files migration

This release includes a bug fix for [a v0.40 migration issue](https://github.com/cosmos/cosmos-sdk/issues/8776) affecting genesis files that contain
multisig accounts was reported.
