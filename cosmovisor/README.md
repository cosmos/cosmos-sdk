# Cosmosvisor

`cosmovisor` is a small process manager for Cosmos SDK application binaries that monitors the governance module for incoming chain upgrade proposals. If it sees a proposal that gets approved, `cosmovisor` can automatically download the new binary, stop the current binary, switch from the old binary to the new one, and finally restart the node with the new binary.

#### Design

Cosmovisor is designed to be used as a wrapper for a `Cosmos SDK` app:

* it will pass arguments to the associated app (configured by `DAEMON_NAME` env variable).
  Running `cosmovisor run arg1 arg2 ....` will run `app arg1 arg2 ...`;
* it will manage an app by restarting and upgrading if needed;
* it is configured using environment variables, not positional arguments.

*Note: If new versions of the application are not set up to run in-place store migrations, migrations will need to be run manually before restarting `cosmovisor` with the new binary. For this reason, we recommend applications adopt in-place store migrations.*

*Note: If validators would like to enable the auto-download option (which [we don't recommend](#auto-download)), and they are currently running an application using Cosmos SDK `v0.42`, they will need to use Cosmovisor [`v0.1`](https://github.com/cosmos/cosmos-sdk/releases/tag/cosmovisor%2Fv0.1.0). Later versions of Cosmovisor do not support Cosmos SDK `v0.44.3` or earlier if the auto-download option is enabled.*

## Contributing

Cosmovisor is part of the Cosmos SDK monorepo, but it's a separate module with it's own release schedule.

Release branches have the following format `release/cosmovisor/vA.B.x`, where A and B are a number (e.g. `release/cosmovisor/v0.1.x`). Releases are tagged using the following format: `cosmovisor/vA.B.C`.

## Setup

### Installation

To install the latest version of `cosmovisor`, run the following command:

```
go install github.com/cosmos/cosmos-sdk/cosmovisor/cmd/cosmovisor@latest
```

To install a previous version, you can specify the version. IMPORTANT: Chains that use Cosmos-SDK v0.44.3 or earlier (eg v0.44.2) and want to use auto-download feature MUST use Cosmovisor v0.1.0

```
go install github.com/cosmos/cosmos-sdk/cosmovisor/cmd/cosmovisor@v0.1.0
```

It is possible to confirm the version of cosmovisor when using Cosmovisor v1.0.0, but it is not possible to do so with `v0.1.0`.

You can also install from source by pulling the cosmos-sdk repository and switching to the correct version and building as follows:

```
git clone git@github.com:cosmos/cosmos-sdk
cd cosmos-sdk
git checkout cosmovisor/vx.x.x
cd cosmovisor
make
```

This will build cosmovisor in your current directory. Afterwards you may want to put it into your machine's PATH like as follows:

```
cp cosmovisor ~/go/bin/cosmovisor
```

*Note: If you are using go `v1.15` or earlier, you will need to use `go get`, and you may want to run the command outside a project directory.*

### Command Line Arguments And Environment Variables

The first argument passed to `cosmovisor` is the action for `cosmovisor` to take. Options are:

* `help`, `--help`, or `-h` - Output `cosmovisor` help information and check your `cosmovisor` configuration.
* `run` - Run the configured binary using the rest of the provided arguments.
* `version`, or `--version` - Output the `cosmovisor` version and also run the binary with the `version` argument.

All arguments passed to `cosmovisor run` will be passed to the application binary (as a subprocess). `cosmovisor` will return `/dev/stdout` and `/dev/stderr` of the subprocess as its own. For this reason, `cosmovisor run` cannot accept any command-line arguments other than those available to the application binary.

*Note: Use of `cosmovisor` without one of the action arguments is deprecated. For backwards compatability, if the first argument is not an action argument, `run` is assumed. However, this fallback might be removed in future versions, so it is recommended that you always provide `run`.

`cosmovisor` reads its configuration from environment variables:

* `DAEMON_HOME` is the location where the `cosmovisor/` directory is kept that contains the genesis binary, the upgrade binaries, and any additional auxiliary files associated with each binary (e.g. `$HOME/.gaiad`, `$HOME/.regend`, `$HOME/.simd`, etc.).
* `DAEMON_NAME` is the name of the binary itself (e.g. `gaiad`, `regend`, `simd`, etc.).
* `DAEMON_ALLOW_DOWNLOAD_BINARIES` (*optional*), if set to `true`, will enable auto-downloading of new binaries (for security reasons, this is intended for full nodes rather than validators). By default, `cosmovisor` will not auto-download new binaries.
* `DAEMON_RESTART_AFTER_UPGRADE` (*optional*, default = `true`), if `true`, restarts the subprocess with the same command-line arguments and flags (but with the new binary) after a successful upgrade. Otherwise (`false`), `cosmovisor` stops running after an upgrade and requires the system administrator to manually restart it. Note restart is only after the upgrade and does not auto-restart the subprocess after an error occurs.
* `DAEMON_POLL_INTERVAL` is the interval length for polling the upgrade plan file. The value can either be a number (in milliseconds) or a duration (e.g. `1s`). Default: 300 milliseconds.
* `UNSAFE_SKIP_BACKUP` (defaults to `false`), if set to `true`, upgrades directly without performing a backup. Otherwise (`false`, default) backs up the data before trying the upgrade. The default value of false is useful and recommended in case of failures and when a backup needed to rollback. We recommend using the default backup option `UNSAFE_SKIP_BACKUP=false`.
* `DAEMON_PREUPGRADE_MAX_RETRIES` (defaults to `0`). The maximum number of times to call `pre-upgrade` in the application after exit status of `31`. After the maximum number of retries, cosmovisor fails the upgrade.

### Folder Layout

`$DAEMON_HOME/cosmovisor` is expected to belong completely to `cosmovisor` and the subprocesses that are controlled by it. The folder content is organized as follows:

```
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

The `cosmovisor/` directory incudes a subdirectory for each version of the application (i.e. `genesis` or `upgrades/<name>`). Within each subdirectory is the application binary (i.e. `bin/$DAEMON_NAME`) and any additional auxiliary files associated with each binary. `current` is a symbolic link to the currently active directory (i.e. `genesis` or `upgrades/<name>`). The `name` variable in `upgrades/<name>` is the URI-encoded name of the upgrade as specified in the upgrade module plan.

Please note that `$DAEMON_HOME/cosmovisor` only stores the *application binaries*. The `cosmovisor` binary itself can be stored in any typical location (e.g. `/usr/local/bin`). The application will continue to store its data in the default data directory (e.g. `$HOME/.gaiad`) or the data directory specified with the `--home` flag. `$DAEMON_HOME` is independent of the data directory and can be set to any location. If you set `$DAEMON_HOME` to the same directory as the data directory, you will end up with a configuation like the following:

```
.gaiad
├── config
├── data
└── cosmovisor
```

## Usage

The system administrator is responsible for:

- installing the `cosmovisor` binary
- configuring the host's init system (e.g. `systemd`, `launchd`, etc.)
- appropriately setting the environmental variables
- manually installing the `genesis` folder
- manually installing the `upgrades/<name>` folders

`cosmovisor` will set the `current` link to point to `genesis` at first start (i.e. when no `current` link exists) and then handle switching binaries at the correct points in time so that the system administrator can prepare days in advance and relax at upgrade time.

In order to support downloadable binaries, a tarball for each upgrade binary will need to be packaged up and made available through a canonical URL. Additionally, a tarball that includes the genesis binary and all available upgrade binaries can be packaged up and made available so that all the necessary binaries required to sync a fullnode from start can be easily downloaded.

The `DAEMON` specific code and operations (e.g. tendermint config, the application db, syncing blocks, etc.) all work as expected. The application binaries' directives such as command-line flags and environment variables also work as expected.

### Detecting Upgrades

`cosmovisor` is polling the `$DAEMON_HOME/data/upgrade-info.json` file for new upgrade instructions. The file is created by the x/upgrade module in `BeginBlocker` when an upgrade is detected and the blockchain reaches the upgrade height.
The following heuristic is applied to detect the upgrade:

+ When starting, `cosmovisor` doesn't know much about currently running upgrade, except the binary which is `current/bin/`. It tries to read the `current/update-info.json` file to get information about the current upgrade name.
+ If neither `cosmovisor/current/upgrade-info.json` nor `data/upgrade-info.json` exist, then `cosmovisor` will wait for `data/upgrade-info.json` file to trigger an upgrade.
+ If `cosmovisor/current/upgrade-info.json` doesn't exist but `data/upgrade-info.json` exists, then `cosmovisor` assumes that whatever is in `data/upgrade-info.json` is a valid upgrade request. In this case `cosmovisor` tries immediately to make an upgrade according to the `name` attribute in `data/upgrade-info.json`.
+ Otherwise, `cosmovisor` waits for changes in `upgrade-info.json`. As soon as a new upgrade name is recorded in the file, `cosmovisor` will trigger an upgrade mechanism.

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

```
> gaiad tx gov submit-proposal software-upgrade Vega \
--title Vega \
--deposit 100uatom \
--upgrade-height 7368420 \
--upgrade-info '{"binaries":{"linux/amd64":"https://github.com/cosmos/gaia/releases/download/v6.0.0-rc1/gaiad-v6.0.0-rc1-linux-amd64","linux/arm64":"https://github.com/cosmos/gaia/releases/download/v6.0.0-rc1/gaiad-v6.0.0-rc1-linux-arm64","darwin/amd64":"https://github.com/cosmos/gaia/releases/download/v6.0.0-rc1/gaiad-v6.0.0-rc1-darwin-amd64"}}' \
--description "upgrade to Vega" \
--gas 400000 \
--from user \
--chain-id test \
--home test/val2 \
--node tcp://localhost:36657 \
--yes
```

2. Store a link to a file that contains all information in the above format (e.g. if you want to specify lots of binaries, changelog info, etc. without filling up the blockchain). For example:

```
https://example.com/testnet-1001-info.json?checksum=sha256:deaaa99fda9407c4dbe1d04bd49bab0cc3c1dd76fa392cd55a9425be074af01e
```

When `cosmovisor` is triggered to download the new binary, `cosmovisor` will parse the `"binaries"` field, download the new binary with [go-getter](https://github.com/hashicorp/go-getter), and unpack the new binary in the `upgrades/<name>` folder so that it can be run as if it was installed manually.

Note that for this mechanism to provide strong security guarantees, all URLs should include a SHA 256/512 checksum. This ensures that no false binary is run, even if someone hacks the server or hijacks the DNS. `go-getter` will always ensure the downloaded file matches the checksum if it is provided. `go-getter` will also handle unpacking archives into directories (in this case the download link should point to a `zip` file of all data in the `bin` directory).

To properly create a sha256 checksum on linux, you can use the `sha256sum` utility. For example:

```
sha256sum ./testdata/repo/zip_directory/autod.zip
```

The result will look something like the following: `29139e1381b8177aec909fab9a75d11381cab5adf7d3af0c05ff1c9c117743a7`.

You can also use `sha512sum` if you would prefer to use longer hashes, or `md5sum` if you would prefer to use broken hashes. Whichever you choose, make sure to set the hash algorithm properly in the checksum argument to the URL.

## Example: SimApp Upgrade

The following instructions provide a demonstration of `cosmovisor` using the simulation application (`simapp`) shipped with the Cosmos SDK's source code. The following commands are to be run from within the `cosmos-sdk` repository.

First, check out the latest `v0.42` release:

```
git checkout v0.42.7
```

Compile the `simd` binary:

```
make build
```

Reset `~/.simapp` (never do this in a production environment):

```
./build/simd unsafe-reset-all
```

Configure the `simd` binary for testing:

```
./build/simd config chain-id test
./build/simd config keyring-backend test
./build/simd config broadcast-mode block
```

Initialize the node and overwrite any previous genesis file (never do this in a production environment):

<!-- TODO: init does not read chain-id from config -->

```
./build/simd init test --chain-id test --overwrite
```

Set the minimum gas price to `0stake` in `~/.simapp/config/app.toml`:

```
minimum-gas-prices = "0stake"
```

Create a new key for the validator, then add a genesis account and transaction:

<!-- TODO: add-genesis-account does not read keyring-backend from config -->
<!-- TODO: gentx does not read chain-id from config -->

```
./build/simd keys add validator
./build/simd add-genesis-account validator 1000000000stake --keyring-backend test
./build/simd gentx validator 1000000stake --chain-id test
./build/simd collect-gentxs
```

Set the required environment variables:

```
export DAEMON_NAME=simd
export DAEMON_HOME=$HOME/.simapp
```

Set the optional environment variable to trigger an automatic restart:

```
export DAEMON_RESTART_AFTER_UPGRADE=true
```

Create the folder for the genesis binary and copy the `simd` binary:

```
mkdir -p $DAEMON_HOME/cosmovisor/genesis/bin
cp ./build/simd $DAEMON_HOME/cosmovisor/genesis/bin
```

For the sake of this demonstration, amend `voting_period` in `genesis.json` to a reduced time of 20 seconds (`20s`):

```
cat <<< $(jq '.app_state.gov.voting_params.voting_period = "20s"' $HOME/.simapp/config/genesis.json) > $HOME/.simapp/config/genesis.json
```

Next, we will hardcode a modification in `simapp` to simulate a code change. In `simapp/app.go`, find the line containing the `UpgradeKeeper` initialization. It should look like the following:

```go
app.UpgradeKeeper = upgradekeeper.NewKeeper(skipUpgradeHeights, keys[upgradetypes.StoreKey], appCodec, homePath)
```

After that line, add the following:

```go
app.UpgradeKeeper.SetUpgradeHandler("test1", func(ctx sdk.Context, plan upgradetypes.Plan) {
	// Add some coins to a random account
	addr, err := sdk.AccAddressFromBech32("cosmos18cgkqduwuh253twzmhedesw3l7v3fm37sppt58")
	if err != nil {
		panic(err)
	}
	err = app.BankKeeper.AddCoins(ctx, addr, sdk.Coins{sdk.Coin{Denom: "stake", Amount: sdk.NewInt(345600000)}})
	if err != nil {
		panic(err)
	}
})
```

Now recompile the `simd` binary with the added upgrade handler:

```
make build
```

Create the folder for the upgrade binary and copy the `simd` binary:

```
mkdir -p $DAEMON_HOME/cosmovisor/upgrades/test1/bin
cp ./build/simd $DAEMON_HOME/cosmovisor/upgrades/test1/bin
```

Start `cosmosvisor`:

```
cosmovisor run start
```

Open a new terminal window and submit an upgrade proposal along with a deposit and a vote (these commands must be run within 20 seconds of each other):

```
./build/simd tx gov submit-proposal software-upgrade test1 --title upgrade --description upgrade --upgrade-height 20 --from validator --yes
./build/simd tx gov deposit 1 10000000stake --from validator --yes
./build/simd tx gov vote 1 yes --from validator --yes
```

The upgrade will occur automatically at height 20.
