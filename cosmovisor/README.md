# Cosmosvisor Quick Start

`cosmovisor` is a small process manager for Cosmos SDK binaries that monitors the governance module via stdout for incoming chain upgrade proposals. If it sees a proposal that gets approved, it can be run manually or automatically to download the new binary, stop the current binary, run the migration script, replace the old node binary with the new one, and finally restart the node with the new genesis file.

## Installation

To install `cosmovisor`, run the following command:

```
go get github.com/cosmos/cosmos-sdk/cosmovisor/cmd/cosmovisor
```

## Command Line Arguments And Environment Variables

All arguments passed to the `cosmovisor` program will be passed to the current daemon binary (as a subprocess). `cosmovisor` will return `/dev/stdout` and `/dev/stderr` of the subprocess as its own. For this reason, `cosmovisor` cannot accept any command line arguments, nor print anything to output (unless it terminates unexpectedly before executing a binary).

`cosmovisor` reads its configuration from environment variables:

* `DAEMON_HOME` is the location where the `cosmovisor/` directory is kept that contains the upgrade binaries (e.g. `$HOME/.gaiad`, `$HOME/.regend`, `$HOME/.simd`, etc.).
* `DAEMON_NAME` is the name of the binary itself (e.g. `gaiad`, `regend`, `simd`, etc.).
* `DAEMON_ALLOW_DOWNLOAD_BINARIES` (*optional*), if set to `true`, will enable auto-downloading of new binaries (for security reasons, this is intended for full nodes rather than validators). By default, `cosmovisor` will not auto-download new binaries.
* `DAEMON_RESTART_AFTER_UPGRADE` (*optional*), if set to `true`, will restart the subprocess with the same command line arguments and flags (but with the new binary) after a successful upgrade. By default, `cosmovisor` stops running after an upgrade and requires the system administrator to manually restart it. Note that `cosmovisor` will not auto-restart the subprocess if there was an error.
* `DAEMON_POLL_INTERVAL` is the interval length in milliseconds for polling the upgrade plan file. Default: 300.
* `DAEMON_UPGRADE_INFO_FILE` is a full path to the upgrade plan file created by the upgrade module in `BeginBlocker` when a new upgrade plan is detected. On start, `cosmovisor` checks if the path is valid, if the base directory exists, and if a filename is provided. If the file name is wrong, the upgrade request will never be handled. Default: `<DAEMON_HOME>/data/upgrade-info.json`.

## Data Folder Layout

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

Each version of the Cosmos SDK application is stored under either `genesis` or `upgrades/<name>`, which holds `bin/$DAEMON_NAME` along with any other needed files such as auxiliary client programs or libraries. `current` is a symbolic link to the currently active folder and `current/bin/$DAEMON_NAME` is the currently active binary.

*Note: The `name` variable in `upgrades/<name>` holds the URI-encoded name of the upgrade as specified in the upgrade module plan.*

Please note that `$DAEMON_HOME/cosmovisor` just stores the *binaries* and associated *program code*. The `cosmovisor` binary can be stored in any typical location (e.g. `/usr/local/bin`). The actual blockchain program will store its data under the default data directory (e.g. `$HOME/.gaiad`) which is independent of `$DAEMON_HOME`. `$DAEMON_HOME` can be set to any location. If you set `$DAEMON_HOME` to the default data directory, you will end up with a configuation like the following:

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
- installing the `genesis` folder manually
- installing the `upgrades/<name>` folders manually

`cosmovisor` will set the `current` link to point to `genesis` at first start (when no `current` link exists) and will handle switching binaries at the correct points in time so that the system administrator can prepare days in advance and relax at upgrade time.

Note that blockchain applications that wish to support upgrades may package up a genesis `cosmovisor` tarball with this information, just as they prepare the genesis binary tarball. In fact, they may package up a tarball with all upgrades up to a current point so that the upgrades can be easily downloaded for others who wish to sync a fullnode from start.

The `DAEMON` specific code and operations (e.g. tendermint config, the application db, syncing blocks, etc.) all work as expected. The application binaries' directives such as command-line flags and environment variables also work as expected.


### Detecting Upgrades

`cosmovisor` is polling the `$DAEMON_UPGRADE_INFO_FILE` file for new upgrade instructions (defaults to `$DAEMON_HOME/data/upgrade-info.json`). The following heuristic is applied to detect the upgrade:
+ When starting `cosmovisor`, `cosmovisor` doesn't know much about currently running upgrade, except the binary (which is either in `current/bin/` or `genesis/bin` if the former doesn't exists). It tries to read the `current/update-info.json` file to get information about the current upgrade name.
+ If `cosmovisor/current/upgrade-info.json` doesn't exist then `cosmovisor` assumes that `data/upgrade-info.json` is an upgrade request. If `data/upgrade-info.json` when starting a cosmovisor but `cosmovisor/current/upgrade-info.json` doesn't exist, then `cosmovisor` tries to make an upgrade according to the `name` attribute in `data/upgrade-info.json`.
+ Otherwise, we wait for the changes in `upgrade-info.json` - as soon as a new upgrade name will be recorded in that file, we trigger an upgrade mechanism.

During the upgrade, we auto-download a new binary (if auto-download is enabled), and link a new directory to the `current` symbolic link based on the `upgrade-info.json:name`. At the end we save `data/upgrade-info.json` to `cosmovisor/current/upgrade-info.json`.


## Auto-Download

Generally, the system requires that the system administrator place all relevant binaries on the disk before the upgrade happens. However, for people who don't need such control and want an easier setup (maybe they are syncing a non-validating fullnode and want to do little maintenance), there is another option.

If you set `DAEMON_ALLOW_DOWNLOAD_BINARIES=true`, and no local binary can be found when an upgrade is triggered, `cosmovisor` will attempt to download and install the binary itself. The plan stored in the upgrade module has an info field for arbitrary JSON. This info is expected to be outputed on the halt log message. There are two valid formats to specify a download in such a message:

1. Store an os/architecture -> binary URI map in the upgrade plan info field as JSON under the `"binaries"` key. For example:

```json
{
  "binaries": {
    "linux/amd64":"https://example.com/gaia.zip?checksum=sha256:aec070645fe53ee3b3763059376134f058cc337247c978add178b6ccdfb0019f"
  }
}
```

2. Store a link to a file that contains all information in the above format (e.g. if you want to specify lots of binaries, changelog info, etc. without filling up the blockchain). For example:

```
https://example.com/testnet-1001-info.json?checksum=sha256:deaaa99fda9407c4dbe1d04bd49bab0cc3c1dd76fa392cd55a9425be074af01e
```

This file contained in the link will be retrieved by [go-getter](https://github.com/hashicorp/go-getter) and the `"binaries"` field will be parsed as above.

If there is no local binary, `DAEMON_ALLOW_DOWNLOAD_BINARIES=true`, and we can access a canonical url for the new binary, then the `cosmovisor` will download it with [go-getter](https://github.com/hashicorp/go-getter) and unpack it into the `upgrades/<name>` folder to be run as if we installed it manually.

Note that for this mechanism to provide strong security guarantees, all URLs should include a SHA 256/512 checksum. This ensures that no false binary is run, even if someone hacks the server or hijacks the DNS. `go-getter` will always ensure the downloaded file matches the checksum if it is provided. `go-getter` will also handle unpacking archives into directories (in this case the download link should point to a `zip` file of all data in the `bin` directory).

To properly create a sha256 checksum on linux, you can use the `sha256sum` utility. For example:

```
sha256sum ./testdata/repo/zip_directory/autod.zip
```

The result will look something like the following: `29139e1381b8177aec909fab9a75d11381cab5adf7d3af0c05ff1c9c117743a7`.

You can also use `sha512sum` if you would prefer to use longer hashes, or `md5sum` if you would prefer to use broken hashes. Whichever you choose, make sure to set the hash algorithm properly in the checksum argument to the URL.

## Example: simd

The following instructions provide a demonstration of `cosmovisor`'s integration with the `simd` application shipped along the Cosmos SDK's source code. The following commands are to be run from within the `cosmos-sdk` repository.

First compile `simd`:

```
make build
```

Create a new key and set up the `simd` node:

```
rm -rf $HOME/.simapp
./build/simd keys --keyring-backend=test add validator
./build/simd init testing --chain-id test
./build/simd add-genesis-account --keyring-backend=test validator 1000000000stake
./build/simd gentx --keyring-backend test --chain-id test validator 1000000stake
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

Create the `cosmovisor`’s genesis folders and copy the `simd` binary:

```
mkdir -p $DAEMON_HOME/cosmovisor/genesis/bin
cp ./build/simd $DAEMON_HOME/cosmovisor/genesis/bin
```

For the sake of this demonstration, amend `voting_params.voting_period` in `$HOME/.simapp/config/genesis.json` to a reduced time of ~5 minutes (`300s`) and then start `cosmosvisor`:

```
cosmovisor start
```

Open a new terminal window and submit a software upgrade proposal:

```
./build/simd tx gov submit-proposal software-upgrade test1 --title "upgrade-demo" --description "upgrade"  --from validator --upgrade-height 100 --deposit 10000000stake --chain-id test --keyring-backend test -y
```

Query the proposal to ensure it was correctly broadcast and added to a block:

```
./build/simd query gov proposal 1
```

Submit a `yes` vote for the upgrade proposal:

```
./build/simd tx gov vote 1 yes --from validator --keyring-backend test --chain-id test -y
```

For the sake of this demonstration, we will hardcode a modification in `simapp` to simulate a code change. In `simapp/app.go`, find the line containing the `UpgradeKeeper` initialization. It should look like the following:

```
app.UpgradeKeeper = upgradekeeper.NewKeeper(skipUpgradeHeights, keys[upgradetypes.StoreKey], appCodec, homePath)
```

After that line, add the following snippet:

 ```
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

Now recompile a new binary and make a copy of it in `$DAEMON_HOME/cosmosvisor/upgrades/test1/bin` (you may need to run `export DAEMON_HOME=$HOME/.simapp` again if you are using a new window):

```
make build
mkdir -p $DAEMON_HOME/cosmovisor/upgrades/test1/bin
cp ./build/simd $DAEMON_HOME/cosmovisor/upgrades/test1/bin
```

The upgrade will occur automatically at height 100.
