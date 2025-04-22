# Cosmos SDK Fork

Fork of Cosmos SDK v0.50.x for Celestia App.
The fork include the following changes compared to upstream:

* Store app version in consensus param store
* Modify continuous vesting account to add start time
* Re-add query router for custom abci queries
* Add v0.52 helpers to facilitate testing
* Disable heavy bank migrations
* Backport improvements for DOS protection for x/authz
* Support historical account number queries
* Support [CIP-30](https://github.com/celestiaorg/CIPs/blob/main/cips/cip-030.md)
* The `prove` flag for is set to `false` for tx queries, similarly to celestia/cosmos-sdk v0.46
* The x/staking migration for delegation keys has been made a lazy migration
* The x/staking migration for historical info keys has been made a lazy migration
* The x/slashing migration for validator missed block bitmap has been made a lazy migration
* The x/slashing key prefix for validator missed block bitmap has been changed from `0x02` to `0x12`
* The celestia-core celestia-core `BlockAPI` is exposed through the app grpc server. When running in standalone mode the app uses a proxy service to maintain support through same the app grpc port.

Read the [CHANGELOG.md](CHANGELOG.md) for more details.

## Audits

| Date       | Auditor                                       | Version                                                                                              | Report                              |
|------------|-----------------------------------------------|------------------------------------------------------------------------------------------------------|-------------------------------------|
| 2025/04/22 | [Informal Systems](https://informal.systems/) | [dba8171](https://github.com/celestiaorg/cosmos-sdk/commit/dba8171d9f829d90134b1669468831625ee89b0e) | [cip-31.pdf](docs/audit/cip-31.pdf) |
