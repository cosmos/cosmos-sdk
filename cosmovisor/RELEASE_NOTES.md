# Cosmovisor v0.1.0 Release Notes

This is the first tracked release of Cosmovisor. It contains the original behavior of scanning app stdin and stdout.
Since the original design, this release contains one important feature: state backup. Since v0.1, by default, cosmovisor will make a state backup (`<app_directory>/data` directory). Backup will be skipped if `UNSAFE_SKIP_BACKUP=true` is set.

Updates to this release will be pushed to `release/cosmovisor/v0.1.x` branch.

Please see the [CHANGELOG](https://github.com/cosmos/cosmos-sdk/blob/release/cosmovisor/v0.1.x/cosmovisor/CHANGELOG.md) for more details.
