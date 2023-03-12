---
sidebar_position: 1
---

# Cosmovisor

`cosmovisor` is a small process manager for Cosmos SDK application binaries that monitors the governance module for incoming chain upgrade proposals. If it sees a proposal that gets approved, `cosmovisor` can automatically download the new binary, stop the current binary, switch from the old binary to the new one, and finally restart the node with the new binary.

* [Cosmovisor](#cosmovisor)
    * [Design](#design)
    * [Contributing](#contributing)
    * [Setup](#setup)
        * [Installation](#installation)
        * [Command Line Arguments And Environment Variables](#command-line-arguments-and-environment-variables)
        * [Folder Layout](#folder-layout)
    * [Usage](#usage)
        * [Initialization](#initialization)
        * [Detecting Upgrades](#detecting-upgrades)
        * [Auto-Download](#auto-download)
    * [Example: SimApp Upgrade](#example-simapp-upgrade)
        * [Chain Setup](#chain-setup)
            * [Prepare Cosmovisor and Start the Chain](#prepare-cosmovisor-and-start-the-chain)
            * [Update App](#update-app)

## Design

Cosmovisor is designed to be used as a wrapper for a `Cosmos SDK` app:

* it will pass arguments to the associated app (configured by `DAEMON_NAME` env variable).
  Running `cosmovisor run arg1 arg2 ....` will run `app arg1 arg2 ...`;
* it will manage an app by restarting and upgrading if needed;
* it is configured using environment variables, not positional arguments.

*Note: If new versions of the application are not set up to run in-place store migrations, migrations will need to be run manually before restarting `cosmovisor` with the new binary. For this reason, we recommend applications adopt in-place store migrations.*

*Note: If validators would like to enable the auto-download option (which [we don't recommend](#auto-download)), and they are currently running an application using Cosmos SDK `v0.42`, they will need to use Cosmovisor [`v0.1`](https://github.com/cosmos/cosmos-sdk/releases/tag/cosmovisor%2Fv0.1.0). Later versions of Cosmovisor do not support Cosmos SDK `v0.44.3` or earlier if the auto-download option is enabled.*

## Contributing

Cosmovisor is part of the Cosmos SDK monorepo, but it's a separate module with it's own release schedule.

Release branches have the following format `release/cosmovisor/vA.B.x`, where A and B are a number (e.g. `release/cosmovisor/v1.3.x`). Releases are tagged using the following format: `cosmovisor/vA.B.C`.

## Setup

### Installation

You can download Cosmovisor from the [GitHub releases](https://github.com/cosmos/cosmos-sdk/releases/tag/cosmovisor%2Fv1.3.0).

To install the latest version of `cosmovisor`, run the following command:

```shell
go install cosmossdk.io/tools/cosmovisor/cmd/cosmovisor@latest
```

To install a previous version, you can specify the version. IMPORTANT: Chains that use Cosmos SDK v0.44.3 or earlier (eg v0.44.2) and want to use auto-download feature MUST use `cosmovisor v0.1.0`

```shell
go install github.com/cosmos/cosmos-sdk/cosmovisor/cmd/cosmovisor@v0.1.0
```

Run `cosmovisor version` to check the cosmovisor version.

Alternatively, for building from source, simply run `make cosmovisor`. The binary will be located in `tools/cosmovisor`.

:::warning
Building from source using `make cosmovisor` won't display the correct `cosmovisor` version.
:::

### Command Line Arguments And Environment Variables

The first argument passed to `cosmovisor` is the action for `cosmovisor` to take. Options are:

* `help`, `--help`, or `-h` - Output `cosmovisor` help information and check your `cosmovisor` configuration.
* `run` - Run the configured binary using the rest of the provided arguments.
* `version` - Output the `cosmovisor` version and also run the binary with the `version` argument.

All arguments passed to `cosmovisor run` will be passed to the application binary (as a subprocess). `cosmovisor` will return `/dev/stdout` and `/dev/stderr` of the subprocess as its own. For this reason, `cosmovisor run` cannot accept any command-line arguments other than those available to the application binary.

*Note: Use of `cosmovisor` without one of the action arguments is deprecated. For backwards compatibility, if the first argument is not an action argument, `run` is assumed. However, this fallback might be removed in future versions, so it is recommended that you always provide `run`.

`cosmovisor` reads its configuration from environment variables:

* `DAEMON_HOME` is the location where the `cosmovisor/` directory is kept that contains the genesis binary, the upgrade binaries, and any additional auxiliary files associated with each binary (e.g. `$HOME/.gaiad`, `$HOME/.regend`, `$HOME/.simd`, etc.).
* `DAEMON_NAME` is the name of the binary itself (e.g. `gaiad`, `regend`, `simd`, etc.).
* `DAEMON_ALLOW_DOWNLOAD_BINARIES` (*optional*), if set to `true`, will enable auto-downloading of new binaries (for security reasons, this is intended for full nodes rather than validators). By default, `cosmovisor` will not auto-download new binaries.
* `DAEMON_RESTART_AFTER_UPGRADE` (*optional*, default = `true`), if `true`, restarts the subprocess with the same command-line arguments and flags (but with the new binary) after a successful upgrade. Otherwise (`false`), `cosmovisor` stops running after an upgrade and requires the system administrator to manually restart it. Note restart is only after the upgrade and does not auto-restart the subprocess after an error occurs.
* `DAEMON_RESTART_DELAY` (*optional*, default none), allow a node operator to define a delay between the node halt (for upgrade) and backup by the specified time. The value must be a duration (e.g. `1s`).
* `DAEMON_POLL_INTERVAL` (*optional*, default 300 milliseconds), is the interval length for polling the upgrade plan file. The value must be a duration (e.g. `1s`).
* `DAEMON_DATA_BACKUP_DIR` option to set a custom backup directory. If not set, `DAEMON_HOME` is used.
* `UNSAFE_SKIP_BACKUP` (defaults to `false`), if set to `true`, upgrades directly without performing a backup. Otherwise (`false`, default) backs up the data before trying the upgrade. The default value of false is useful and recommended in case of failures and when a backup needed to rollback. We recommend using the default backup option `UNSAFE_SKIP_BACKUP=false`.
* `DAEMON_PREUPGRADE_MAX_RETRIES` (defaults to `0`). The maximum number of times to call `pre-upgrade` in the application after exit status of `31`. After the maximum number of retries, Cosmovisor fails the upgrade.
* `COSMOVISOR_DISABLE_LOGS` (defaults to `false`). If set to true, this will disable Cosmovisor logs (but not the underlying process) completely. This may be useful, for example, when a Cosmovisor subcommand you are executing returns a valid JSON you are then parsing, as logs added by Cosmovisor make this output not a valid JSON.

### Folder Layout

`$DAEMON_HOME/cosmovisor` is expected to belong completely to `cosmovisor` and the subprocesses that are controlled by it. The folder content is organized as follows:

```text
.
├── current -> genesis or upgrades/<name>
├── genesis
│   └── bin
│       └── $DAEMON_NAME
└── upgrades
    └── <name>
        ├── bin
        │   └── $DAEMON_NAME
        └── upgrade-info.json
```

The `cosmovisor/` directory incudes a subdirectory for each version of the application (i.e. `genesis` or `upgrades/<name>`). Within each subdirectory is the application binary (i.e. `bin/$DAEMON_NAME`) and any additional auxiliary files associated with each binary. `current` is a symbolic link to the currently active directory (i.e. `genesis` or `upgrades/<name>`). The `name` variable in `upgrades/<name>` is the lowercased URI-encoded name of the upgrade as specified in the upgrade module plan. Note that the upgrade name path are normalized to be lowercased: for instance, `MyUpgrade` is normalized to `myupgrade`, and its path is `upgrades/myupgrade`.

Please note that `$DAEMON_HOME/cosmovisor` only stores the *application binaries*. The `cosmovisor` binary itself can be stored in any typical location (e.g. `/usr/local/bin`). The application will continue to store its data in the default data directory (e.g. `$HOME/.gaiad`) or the data directory specified with the `--home` flag. `$DAEMON_HOME` is independent of the data directory and can be set to any location. If you set `$DAEMON_HOME` to the same directory as the data directory, you will end up with a configuation like the following:

```text
.gaiad
├── config
├── data
└── cosmovisor
```

## Usage

The system administrator is responsible for:

* installing the `cosmovisor` binary
* configuring the host's init system (e.g. `systemd`, `launchd`, etc.)
* appropriately setting the environmental variables
* creating the `<DAEMON_HOME>/cosmovisor` directory
* creating the `<DAEMON_HOME>/cosmovisor/genesis/bin` folder
* creating the `<DAEMON_HOME>/cosmovisor/upgrades/<name>/bin` folders
* placing the different versions of the `<DAEMON_NAME>` executable in the appropriate `bin` folders.

`cosmovisor` will set the `current` link to point to `genesis` at first start (i.e. when no `current` link exists) and then handle switching binaries at the correct points in time so that the system administrator can prepare days in advance and relax at upgrade time.

In order to support downloadable binaries, a tarball for each upgrade binary will need to be packaged up and made available through a canonical URL. Additionally, a tarball that includes the genesis binary and all available upgrade binaries can be packaged up and made available so that all the necessary binaries required to sync a fullnode from start can be easily downloaded.

The `DAEMON` specific code and operations (e.g. tendermint config, the application db, syncing blocks, etc.) all work as expected. The application binaries' directives such as command-line flags and environment variables also work as expected.

### Initialization

The `cosmovisor init <path to executable>` command creates the folder structure required for using cosmovisor.

It does the following:

* creates the `<DAEMON_HOME>/cosmovisor` folder if it doesn't yet exist
* creates the `<DAEMON_HOME>/cosmovisor/genesis/bin` folder if it doesn't yet exist
* copies the provided executable file to `<DAEMON_HOME>/cosmovisor/genesis/bin/<DAEMON_NAME>`
* creates the `current` link, pointing to the `genesis` folder

It uses the `DAEMON_HOME` and `DAEMON_NAME` environment variables for folder location and executable name.

The `cosmovisor init` command is specifically for initializing cosmovisor, and should not be confused with a chain's `init` command (e.g. `cosmovisor run init`).

### Detecting Upgrades

`cosmovisor` is polling the `$DAEMON_HOME/data/upgrade-info.json` file for new upgrade instructions. The file is created by the x/upgrade module in `BeginBlocker` when an upgrade is detected and the blockchain reaches the upgrade height.
The following heuristic is applied to detect the upgrade:

* When starting, `cosmovisor` doesn't know much about currently running upgrade, except the binary which is `current/bin/`. It tries to read the `current/update-info.json` file to get information about the current upgrade name.
* If neither `cosmovisor/current/upgrade-info.json` nor `data/upgrade-info.json` exist, then `cosmovisor` will wait for `data/upgrade-info.json` file to trigger an upgrade.
* If `cosmovisor/current/upgrade-info.json` doesn't exist but `data/upgrade-info.json` exists, then `cosmovisor` assumes that whatever is in `data/upgrade-info.json` is a valid upgrade request. In this case `cosmovisor` tries immediately to make an upgrade according to the `name` attribute in `data/upgrade-info.json`.
* Otherwise, `cosmovisor` waits for changes in `upgrade-info.json`. As soon as a new upgrade name is recorded in the file, `cosmovisor` will trigger an upgrade mechanism.

When the upgrade mechanism is triggered, `cosmovisor` will:

1. if `DAEMON_ALLOW_DOWNLOAD_BINARIES` is enabled, start by auto-downloading a new binary into `cosmovisor/<name>/bin` (where `<name>` is the `upgrade-info.json:name` attribute);
2. update the `current` symbolic link to point to the new directory and save `data/upgrade-info.json` to `cosmovisor/current/upgrade-info.json`.

### Auto-Download

Generally, `cosmovisor` requires that the system administrator place all relevant binaries on disk before the upgrade happens. However, for people who don't need such control and want an automated setup (maybe they are syncing a non-validating fullnode and want to do little maintenance), there is another option.

**NOTE: we don't recommend using auto-download** because it doesn't verify in advance if a binary is available. If there will be any issue with downloading a binary, the cosmovisor will stop and won't restart an App (which could lead to a chain halt).

If `DAEMON_ALLOW_DOWNLOAD_BINARIES` is set to `true`, and no local binary can be found when an upgrade is triggered, `cosmovisor` will attempt to download and install the binary itself based on the instructions in the `info` attribute in the `data/upgrade-info.json` file. The files is constructed by the x/upgrade module and contains data from the upgrade `Plan` object. The `Plan` has an info field that is expected to have one of the following two valid formats to specify a download:

1. Store an os/architecture -> binary URI map in the upgrade plan info field as JSON under the `"binaries"` key. For example:

    ```json
    {
      "binaries": {
        "linux/amd64":"https://example.com/gaia.zip?checksum=sha256:aec070645fe53ee3b3763059376134f058cc337247c978add178b6ccdfb0019f"
      }
    }
    ```

    You can include multiple binaries at once to ensure more than one environment will receive the correct binaries:

    ```json
    {
      "binaries": {
        "linux/amd64":"https://example.com/gaia.zip?checksum=sha256:aec070645fe53ee3b3763059376134f058cc337247c978add178b6ccdfb0019f",
        "linux/arm64":"https://example.com/gaia.zip?checksum=sha256:aec070645fe53ee3b3763059376134f058cc337247c978add178b6ccdfb0019f",
        "darwin/amd64":"https://example.com/gaia.zip?checksum=sha256:aec070645fe53ee3b3763059376134f058cc337247c978add178b6ccdfb0019f"
        }
    }
    ```

    When submitting this as a proposal ensure there are no spaces. An example command using `gaiad` could look like:

    ```shell
    > gaiad tx upgrade software-upgrade Vega \
    --title Vega \
    --deposit 100uatom \
    --upgrade-height 7368420 \
    --upgrade-info '{"binaries":{"linux/amd64":"https://github.com/cosmos/gaia/releases/download/v6.0.0-rc1/gaiad-v6.0.0-rc1-linux-amd64","linux/arm64":"https://github.com/cosmos/gaia/releases/download/v6.0.0-rc1/gaiad-v6.0.0-rc1-linux-arm64","darwin/amd64":"https://github.com/cosmos/gaia/releases/download/v6.0.0-rc1/gaiad-v6.0.0-rc1-darwin-amd64"}}' \
    --summary "upgrade to Vega" \
    --gas 400000 \
    --from user \
    --chain-id test \
    --home test/val2 \
    --node tcp://localhost:36657 \
    --yes
    ```

2. Store a link to a file that contains all information in the above format (e.g. if you want to specify lots of binaries, changelog info, etc. without filling up the blockchain). For example:

    ```text
    https://example.com/testnet-1001-info.json?checksum=sha256:deaaa99fda9407c4dbe1d04bd49bab0cc3c1dd76fa392cd55a9425be074af01e
    ```

When `cosmovisor` is triggered to download the new binary, `cosmovisor` will parse the `"binaries"` field, download the new binary with [go-getter](https://github.com/hashicorp/go-getter), and unpack the new binary in the `upgrades/<name>` folder so that it can be run as if it was installed manually.

Note that for this mechanism to provide strong security guarantees, all URLs should include a SHA 256/512 checksum. This ensures that no false binary is run, even if someone hacks the server or hijacks the DNS. `go-getter` will always ensure the downloaded file matches the checksum if it is provided. `go-getter` will also handle unpacking archives into directories (in this case the download link should point to a `zip` file of all data in the `bin` directory).

To properly create a sha256 checksum on linux, you can use the `sha256sum` utility. For example:

```shell
sha256sum ./testdata/repo/zip_directory/autod.zip
```

The result will look something like the following: `29139e1381b8177aec909fab9a75d11381cab5adf7d3af0c05ff1c9c117743a7`.

You can also use `sha512sum` if you would prefer to use longer hashes, or `md5sum` if you would prefer to use broken hashes. Whichever you choose, make sure to set the hash algorithm properly in the checksum argument to the URL.

## Example: SimApp Upgrade

The following instructions provide a demonstration of `cosmovisor` using the simulation application (`simapp`) shipped with the Cosmos SDK's source code. The following commands are to be run from within the `cosmos-sdk` repository.

### Chain Setup

Let's create a new chain using the `v0.44` version of simapp (the Cosmos SDK demo app):

```shell
git checkout v0.44.6
make build
```

Clean `~/.simapp` (never do this in a production environment):

```shell
./build/simd unsafe-reset-all
```

Set up app config:

```shell
./build/simd config set client chain-id test
./build/simd config set client keyring-backend test
./build/simd config set client broadcast-mode sync
```

Initialize the node and overwrite any previous genesis file (never do this in a production environment):

<!-- TODO: init does not read chain-id from config -->

```shell
./build/simd init test --chain-id test --overwrite
```

Set the minimum gas price to `0stake` in `~/.simapp/config/app.toml`:

```shell
minimum-gas-prices = "0stake"
```

For the sake of this demonstration, amend `voting_period` in `genesis.json` to a reduced time of 20 seconds (`20s`):

```shell
cat <<< $(jq '.app_state.gov.voting_params.voting_period = "20s"' $HOME/.simapp/config/genesis.json) > $HOME/.simapp/config/genesis.json
```

Create a validator, and setup genesis transaction:

```shell
./build/simd keys add validator
./build/simd genesis add-genesis-account validator 1000000000stake --keyring-backend test
./build/simd genesis gentx validator 1000000stake --chain-id test
./build/simd genesis collect-gentxs
```

#### Prepare Cosmovisor and Start the Chain

Set the required environment variables:

```shell
export DAEMON_NAME=simd
export DAEMON_HOME=$HOME/.simapp
```

Set the optional environment variable to trigger an automatic app restart:

```shell
export DAEMON_RESTART_AFTER_UPGRADE=true
```

Create the folder for the genesis binary and copy the `simd` binary:

```shell
mkdir -p $DAEMON_HOME/cosmovisor/genesis/bin
cp ./build/simd $DAEMON_HOME/cosmovisor/genesis/bin
```

Now you can run cosmovisor with simapp v0.44:

```shell
cosmovisor run start
```

#### Update App

Update app to the latest version (e.g. v0.45).

Next, we can add a migration - which is defined using `x/upgrade` [upgrade plan](https://github.com/cosmos/cosmos-sdk/blob/main/docs/core/upgrade.md) (you may refer to a past version if you are using an older Cosmos SDK release). In a migration we can do any deterministic state change.

Build the new version `simd` binary:

```shell
make build
```

Create the folder for the upgrade binary and copy the `simd` binary:

```shell
mkdir -p $DAEMON_HOME/cosmovisor/upgrades/test1/bin
cp ./build/simd $DAEMON_HOME/cosmovisor/upgrades/test1/bin
```

Open a new terminal window and submit an upgrade proposal along with a deposit and a vote (these commands must be run within 20 seconds of each other):

**<= v0.45**:

```shell
./build/simd tx gov submit-proposal software-upgrade test1 --title upgrade --description upgrade --upgrade-height 200 --from validator --yes
./build/simd tx gov deposit 1 10000000stake --from validator --yes
./build/simd tx gov vote 1 yes --from validator --yes
```

**v0.46, v0.47**:

```shell
./build/simd tx gov submit-legacy-proposal software-upgrade test1 --title upgrade --description upgrade --upgrade-height 200 --from validator --yes
./build/simd tx gov deposit 1 10000000stake --from validator --yes
./build/simd tx gov vote 1 yes --from validator --yes
```

**>= v0.48+**:

```shell
./build/simd tx upgrade software-upgrade test1 --title upgrade --summary upgrade --upgrade-height 200 --from validator --yes
./build/simd tx gov deposit 1 10000000stake --from validator --yes
./build/simd tx gov vote 1 yes --from validator --yes
```

The upgrade will occur automatically at height 200. Note: you may need to change the upgrade height in the snippet above if your test play takes more time.
