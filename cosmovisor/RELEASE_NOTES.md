# Cosmovisor v1.1.0 Release Notes

### New execution model

With this release we are shifting to a new CLI design: 

* in the past, Cosmovisor was designed to act as a wrapper for a Cosmos App. An admin could link it and use it instead of the Cosmos App. When running it will pass all options and configuration  parameters to the app. Hence the only way to configure the Cosmovisor was through environment variables.
* now, we are moving to a more traditional model, where Cosmovisor has it's own command set and is a true supervisor.

New commands have been added:

* `run` will start the Cosmos App and pass remaining arguments to the app (similar to `npm run`)
* `help` will display Cosmovisor help
* `version` will display both Cosmovisor and the associated app version.

The existing way of starting an app with Cosmovisor has been deprecated (`cosmovisor [app params]`) and will be removed in the future version. Please use `cosmovisor run [app pararms]`  instead.

### New Features

We added a new configuration option: `DAEMON_BACKUP_DIR` (as env variable). When set, Cosmovisor will create backup the app data backup in that directory (instead of using the app home directory) before running the update. See the [README](https://github.com/cosmos/cosmos-sdk/blob/main/cosmovisor/README.md#command-line-arguments-and-environment-variables) file for more details.

### Bug Fixes

* Fixed `cosmovisor version` output when installed using 'go install github.com/cosmos/cosmos-sdk/cosmovisor/cmd/cosmovisor@v1.0.0'.

### Changelog

For more details, please see the [CHANGELOG](https://github.com/cosmos/cosmos-sdk/blob/cosmovisor/v1.1.0/cosmovisor/CHANGELOG.md).
