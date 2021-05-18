# Cosmos SDK v0.42.5 "Stargate" Release Notes

This release includes various minor bugfixes and improvments, including:

- Fix support for building the Cosmos SDK on ARM architectures,
- Fix the `[appd] keys show/list` CLI subcommands for multisigs,
- Internal code performance improvment.

It also introduces one new feature: adding the `[appd] config` subcommand back to the SDK.

See the [Cosmos SDK v0.42.5 milestone](https://github.com/cosmos/cosmos-sdk/milestone/44?closed=1) on our issue tracker for the exhaustive list of all changes.

### The `config` Subcommand

One new feature introduced in the Stargate series was the merging of the two CLI binaries `[appd]` and `[appcli]` into one single application binary. In this process, the `[appcli] config` subcommand, which was used to save client-side configuration into a TOML file, was removed.

Due to [popular demand](https://github.com/cosmos/cosmos-sdk/issues/8529), we have introduced this feature back to the SDK, under the `[appd] config` subcommand. The functionality is as follows:

- `[appd] config`: Output all client-side configuration.
- `[appd] config [config-name]`: Get the given configuration (e.g. `keyring-backend` or `node-id`).
- `[appd] config [config-name] [config-value]`: Set and persist the given configuration with the new value.

All configurations are persisted to the filesystem, under the path `$APP_HOME/config/client.toml`. For the list of all possible client-side configurations, please have a look at that file, as it is heavily commented.

Environment variables binding to client-side configuration also works. For example, the command `KEYRING_BACKEND=os [appd] tx bank send ...` will bind ENV variable to the `keyring-backend` config. The order or precedence for config is: `flags > env vars > client.toml file`.
