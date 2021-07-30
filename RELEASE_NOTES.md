# Cosmos SDK v0.42.8 "Stargate" Release Notes

This release includes various minor bugfixes and improvments, including:

- emit events for tx signature and sequence, so clients can now query txs by signature (`tx.signature='<base64_sig>'`) or by address and sequence combo (`tx.acc_seq='<addr>/<seq>'`),
- support other signing algorithms than `secp256k1` with `--ledger` in with the CLI `keys add` command,
- add missing documentation for CLI flag `--output json/text` to all `tx` cli commands.

See the [Cosmos SDK v0.42.8 milestone](https://github.com/cosmos/cosmos-sdk/milestone/50?closed=1) on our issue tracker for the exhaustive list of all changes.
