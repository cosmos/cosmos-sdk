# Cosmosvisor Quick Start

`cosmovisor` is a small process manager around Cosmos SDK binaries that monitors the governance module via stdout to see if there's a chain upgrade proposal coming in. If it see a proposal that gets approved it can be run manually or automatically to download the new code, stop the node, run the migration script, replace the node binary, and start with the new genesis file.

## Installation

Run:

`go get github.com/cosmos/cosmos-sdk/cosmovisor/cmd/cosmovisor`

## Command Line Arguments And Environment Variables

All arguments passed to the `cosmovisor` program will be passed to the current daemon binary (as a subprocess).
It will return `/dev/stdout` and `/dev/stderr` of the subprocess as its own. Because of that, it cannot accept
any command line arguments, nor print anything to output (unless it terminates unexpectedly before executing a
binary).

`cosmovisor` reads its configuration from environment variables:

* `DAEMON_HOME` is the location where upgrade binaries should be kept (e.g. `$HOME/.gaiad` or `$HOME/.xrnd`).
* `DAEMON_NAME` is the name of the binary itself (eg. `xrnd`, `gaiad`, `simd`, etc).
* `DAEMON_ALLOW_DOWNLOAD_BINARIES` (*optional*) if set to `true` will enable auto-downloading of new binaries
(for security reasons, this is intended for full nodes rather than validators).
* `DAEMON_RESTART_AFTER_UPGRADE` (*optional*) if set to `true` it will restart the sub-process with the same
command line arguments and flags (but new binary) after a successful upgrade. By default, `cosmovisor` dies
afterwards and allows the supervisor to restart it if needed. Note that this will not auto-restart the child
if there was an error.

## Data Folder Layout

`$DAEMON_HOME/cosmovisor` is expected to belong completely to `cosmovisor` and 
subprocesses that are controlled by it. The folder content is organised as follows:

```
.
├── current -> genesis or upgrades/<name>
├── genesis
│   └── bin
│       └── $DAEMON_NAME
└── upgrades
    └── <name>
        └── bin
            └── $DAEMON_NAME
```

Each version of the Cosmos SDK application is stored under either `genesis` or `upgrades/<name>`, which holds `bin/$DAEMON_NAME`
along with any other needed files such as auxiliary client programs or libraries. `current` is a symbolic link to the currently
active folder (so `current/bin/$DAEMON_NAME` is the currently active binary).

*Note: the `name` variable in `upgrades/<name>` holds the URI-encoded name of the upgrade as specified in the upgrade module plan.*

Please note that `$DAEMON_HOME/cosmovisor` just stores the *binaries* and associated *program code*.
The `cosmovisor` binary can be stored in any typical location (eg `/usr/local/bin`). The actual blockchain
program will store it's data under their default data directory (e.g. `$HOME/.gaiad`) which is independent of
the `$DAEMON_HOME`. You can choose to set `$DAEMON_HOME` to the actual binary's home directory and then end up
with a configuation like the following, but this is left as a choice to the system admininstrator for best
directory layout:

```
.gaiad
├── config
├── data
└── cosmovisor
```

## Usage

The system administrator admin is responsible for:
* installing the `cosmovisor` binary and configure the host's init system (e.g. `systemd`, `launchd`, etc) along with the environmental variables appropriately;
* installing the `genesis` folder manually;
* installing the `upgrades/<name>` folders manually.

`cosmovisor` will set the `current` link to point to `genesis` at first start (when no `current` link exists) and handles
binaries switch overs at the correct points in time, so that the system administrator can prepare days in advance and relax at upgrade time.

Note that blockchain applications that wish to support upgrades may package up a genesis `cosmovisor` tarball with this information,
just as they prepare the genesis binary tarball. In fact, they may offer a tarball will all upgrades up to current point for easy download
for those who wish to sync a fullnode from start.

The `DAEMON` specific code and operations (e.g. tendermint config, the application db, syncing blocks, etc) are performed as normal.
Application binaries' directives such as command-line flags and environment variables work normally.

## Example: simd

The following instructions provide a demonstration of `cosmovisor`'s integration with the `simd` application
shipped along the Cosmos SDK's source code.

First compile `simd`:

```
cd cosmos-sdk/
make build
```

Create a new key and setup the `simd` node:

```
rm -rf $HOME/.simapp
./build/simd keys --keyring-backend=test add validator
./build/simd init testing --chain-id test
./build/simd add-genesis-account --keyring-backend=test $(./build/simd keys --keyring-backend=test show validator -a) 1000000000stake,1000000000validatortoken
./build/simd gentx --keyring-backend test --chain-id test validator 100000stake
./build/simd collect-gentxs
```

Set the required environment variables:

```
export DAEMON_NAME=simd         # binary name
export DAEMON_HOME=$HOME/.simapp  # daemon's home directory
```

Create the `cosmovisor`’s genesis folders and deploy the binary:

```
mkdir -p $DAEMON_HOME/cosmovisor/genesis/bin
cp ./build/simd $DAEMON_HOME/cosmovisor/genesis/bin
```

For the sake of this demonstration, we would amend `voting_params.voting_period` in `.simapp/config/genesis.json` to a reduced time ~5 minutes (300s) and eventually launch `cosmosvisor`:

```
cosmovisor start
```

Submit a software upgrade proposal:

```
./build/simd tx gov submit-proposal software-upgrade test1 --title "upgrade-demo" --description "upgrade"  --from validator --upgrade-height 100 --deposit 10000000stake --chain-id test --keyring-backend test -y
```
 
Query the proposal to ensure it was correctly broadcast and added to a block:

```
./build/simd query gov proposal 1
```
 
Submit a `Yes` vote for the upgrade proposal:

```
./build/simd tx gov vote 1 yes --from validator --keyring-backend test --chain-id test -y
```

For the sake of this demonstration, we will hardcode a modification in `simapp` to simulate a code change.
In `simapp/app.go`, find the line containing the upgrade Keeper initialisation, it should look like
`app.UpgradeKeeper = upgradekeeper.NewKeeper(skipUpgradeHeights, keys[upgradetypes.StoreKey], appCodec, homePath)`.
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

Now recompile a new binary and place it in `$DAEMON_HOME/cosmosvisor/upgrades/test1/bin`:

```
make build
cp ./build/simd $DAEMON_HOME/cosmovisor/upgrades/test1/bin
```

The upgrade will occur automatically at height 100.
