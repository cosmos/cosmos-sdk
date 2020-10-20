# Cosmosvisor Quick Start

`cosmovisor` is a small process manager around Cosmos SDK binaries that use the upgrade module to allow
for smooth and configurable management of upgrading binaries as a live chain is upgraded, and can be
used to simplify validator operations while doing upgrades or to make syncing a full node for genesis
simple. The `cosmovisor` program monitors the stdout of Cosmos SDK application's executable to look for
messages from the upgrade module indicating a pending or required upgrade and act appropriately.

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
subprocesses the controlled by it. The folder content is organised as follows:

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

## Example

1) Install cosmovisor 
`go get github.com/cosmos/cosmos-sdk/cosmovisor/cmd/cosmovisor`

2) Compile simd from sdk
```
cd cosmos-sdk/
make build
```

3) Setup simd node using following instructions
```
rm -rf $HOME/.simapp

./build/simd init testing --chain-id test

./build/simd add-genesis-account $(./build/simd keys show validator -a) 1000000000stake,1000000000validatortoken

./build/simd gentx validator --chain-id test

./build/simd collect-gentxs
```

4) Set the required environment variables:
```
export DAEMON_NAME=simd         # binary name
export DAEMON_HOME=$HOME/.simapp  # daemon's home directory
```

5) Create the cosmovisor’s genesis sub-directories and deploy the binary:
```
mkdir -p $DAEMON_HOME/cosmovisor/genesis/bin
cp ./build/simd $DAEMON_HOME/cosmovisor/genesis/bin
```

6) Change `voting_params.voting_period` in `.gaiad/config/genesis.json` to a reduced time ~5 minutes(300s)

7) Start the daemon using cosmovisor
`cosmovisor start`

8) Submit a software upgrade proposal
`./build/simd tx gov submit-proposal software-upgrade test1 --title "test1" --description "upgrade"  --from validator --upgrade-height 100 --deposit 10000000stake -y --chain-id test`
 
9) Query the proposal to ensure proposal has been created
`./build/simd query gov proposal 1`
 
10) Vote for the proposal with yes

 `./build/simd tx gov vote 1 yes --from validator -y --chain-id test`
 
11) Add the following code snippet to sdk/simapp/app.go

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

12) Comple a new binary using the updated app.go and move it to cosmosvisor/upgrades/<upgrade name>/bin

```
make build
mkdir -p $DAEMON_HOME/cosmovisor/genesis/bin
cp ./build/simd $DAEMON_HOME/cosmovisor/upgrades/test1/bin
```

13)The upgrade should occur automatically at height 100
