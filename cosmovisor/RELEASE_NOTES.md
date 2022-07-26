# Cosmovisor v1.2.0 Release Notes

### New Features

With the `cosmovisor init` command, all the necessary folders for using cosmovisor are automatically created. You do not need to manually symlink the chain binary anymore.

We've added a new configuration option: `DAEMON_RESTART_DELAY` (as env variable). When set, Cosmovisor will wait that delay between the node halt and backup. See the [README](https://github.com/cosmos/cosmos-sdk/blob/main/cosmovisor/README.md#command-line-arguments-and-environment-variables) file for more details.

### Bug Fixes

* Fix Cosmovisor binary usage for pre-upgrade. Cosmovisor was using the wrong binary when running a `pre-upgrade` command.
 
### Changelog

For more details, please see the [CHANGELOG](https://github.com/cosmos/cosmos-sdk/blob/cosmovisor/v1.1.0/cosmovisor/CHANGELOG.md).
