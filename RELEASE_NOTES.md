# Cosmos SDK v0.44.1 Release Notes

This release introduces bug fixes and improvements on the Cosmos SDK v0.44 series.

The main bug fix concerns all users performing in-place store migrations from v0.42 to v0.44. A source of non-determinism in the upgrade process has been [detected and fixed](https://github.com/cosmos/cosmos-sdk/pull/10189) in this release, causing consensus errors. As such, **v0.44.0 is not safe to use when performing v0.42->v0.44 in-place store upgrades**, please use this release v0.44.1 instead. This does not impact genesis JSON dump upgrades nor fresh chains starting with v0.44.

Another bug fix concerns calling the ABCI `Query` method using `client.Context`. We modified ABCI queries to use `abci.QueryRequest`'s `Height` field if it is non-zero, otherwise continue using `client.Context`'s height. This is a minor client-breaking change for users of the `client.Context`.

Some CLI fixes are also included, such as:

- using pre-configured data for the CLI `add-genesis-account` command ([\#9969](https://github.com/cosmos/cosmos-sdk/pull/9969)),
- ensuring the `init` command reads the `--home` flag value correctly ([#10104](https://github.com/cosmos/cosmos-sdk/pull/10104)),
- fixing the error message when `period` or `period-limit` flag is not set on a feegrant grant transaction [\#10049](https://github.com/cosmos/cosmos-sdk/issues/10049).

v0.44.1 also includes performance improvements, namely:

- IAVL update to v0.17.1 which includes performance improvements on a batch load [\#10040](https://github.com/cosmos/cosmos-sdk/pull/10040),
- Speedup coins.AmountOf(), by removing many intermittent regex calls [\#10021](https://github.com/cosmos/cosmos-sdk/pull/10021),
- Improve CacheKVStore datastructures / algorithms, to no longer take O(N^2) time when interleaving iterators and insertions [\#10026](https://github.com/cosmos/cosmos-sdk/pull/10026).

See the [Cosmos SDK v0.44.1 milestone](https://github.com/cosmos/cosmos-sdk/milestone/56?closed=1) on our issue tracker for the exhaustive list of all changes.
