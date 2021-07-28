# Cosmos SDK v0.43.0-RC3 Release Notes

This is a fourth Release Candidate for v0.43. If no issue will be reported this will be the final release.

This release contains one bug fix:

- [\#9793](https://github.com/cosmos/cosmos-sdk/pull/9793) Fixed ECDSA/secp256r1 transaction malleability.

It also includes a couple of minor improvements:

- [\#9750](https://github.com/cosmos/cosmos-sdk/pull/9750) Emit events for tx signature and sequence, so clients can now query txs by signature (`tx.signature='<base64_sig>'`) or by address and sequence combo (`tx.acc_seq='<addr>/<seq>'`).
- (cli) [\#9717](https://github.com/cosmos/cosmos-sdk/pull/9717) Added CLI flag `--output json/text` to `tx` cli commands.
