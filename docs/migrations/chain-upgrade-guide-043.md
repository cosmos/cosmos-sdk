<!--
order: 1
-->

# Chain Upgrade Guide to v0.43

This document explains how to perform a chain upgrade from v0.42 to v0.43. {synopsis}

::: warning
You must upgrade to Stargate v0.42 before upgrading to v0.43. If you have not done so, please see [Chain Upgrade Guide to v0.42](/v0.42/migrations/chain-upgrade-guide-040.html).
:::

## Prerequisite Readings

- [Cosmovisor](../run-node/cosmovisor.html) {prereq}
- [In-Place Store Migrations](../core/upgrade.html) {prereq}
- [Upgrading Modules](../building-modules/upgrade.html) {prereq}

## In-Place Store Migrations

We recommend using [In-Place Store Migrations](../core/upgrade.html) to upgrade your chain from v0.42 to v0.43. The first step is to make sure all your modules follow the [Module Upgrade Guide](../building-modules/upgrade.html). The second step is to add an [upgrade handler](../core/upgrade.html#running-migrations) to `app.go`.

In this document, we'll provide an example of what the upgrade handler should look like for a chain upgrading module versions for the first time. It is critical to note that the initial version of each module must be set to `1` (not `0`) or else the upgrade will run [`InitGenesis`](../building-modules/genesis.html#initgenesis) for each module.

## Preparing Upgrade Binaries

We recommend validators use [Cosmovisor](../run-node/cosmovisor.html), which is a process manager for running application binaries. For security reasons, we recommend validators build their own upgrade binaries rather than enabling the auto-download option. Using Cosmovisor with the auto-restart option will ensure the upgrade is run with zero downtime.

We also recommend application developers prepare and maintain a tarball with the genesis binary and all available upgrade binaries. This tarball can be used to sync a full node from start without having to manually run each upgrade. See [Cosmovisor](../run-node/cosmovisor.html) for more information about setting up the auto-download option.

## Example: Simapp Upgrade

The following example will walk through the upgrade process using `simapp` as our blockchain application. We will be upgrading `simapp` from v0.42 to v0.43.
 
### Initial Setup

From within the `cosmos-sdk` repository, check out `v0.42.6`:

```
git checkout v0.42.6
```

Build the `simd` binary for `v0.42.6` (the genesis binary):

```
make build
```

Configure `simd` and initialize the node:

<!-- TODO: init does not read chain-id from config -->

```
rm -rf $HOME/.simapp
./build/simd config chain-id test
./build/simd config keyring-backend test
./build/simd config broadcast-mode block
./build/simd init test --chain-id test
```

For the purpose of this demonstration, change `voting_period` in `genesis.json` to a reduced time of 20 seconds (`20s`):

```
cat <<< $(jq '.app_state.gov.voting_params.voting_period = "20s"' $HOME/.simapp/config/genesis.json) > $HOME/.simapp/config/genesis.json
```

Create a genesis account and transaction, and then load the genesis transaction:

<!-- TODO: add-genesis-account does not read keyring-backend from config -->
<!-- TODO: gentx does not read chain-id from config -->

```
./build/simd keys add validator
./build/simd add-genesis-account validator 5000000000stake --keyring-backend test
./build/simd gentx validator 1000000stake --chain-id test
./build/simd collect-gentxs
```

Now that our node is initialized and we are ready to start a new `simapp` chain, let's set up `cosmovisor` and build the genesis binary.

### Cosmovisor Setup

First, install or update `cosmovisor`:

```
go get github.com/cosmos/cosmos-sdk/cosmovisor/cmd/cosmovisor
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

Create the cosmovisor genesis folder and copy the `v0.42.6` binary:

```
mkdir -p $DAEMON_HOME/cosmovisor/genesis/bin
cp ./build/simd $DAEMON_HOME/cosmovisor/genesis/bin
```

Now that `cosmovisor` is installed and we have set up the genesis binary, let's add the upgrade handler and then set up the upgrade binary.

### Chain Upgrade

<!-- TODO: update master to v0.43.0 -->

Check out `master`:

```
git checkout master
```

Add the following to `simapp/app.go` starting on line 260:

```go
	app.registerUpgradeHandlers()
```

Add the following to `simapp/app.go` starting on line 420:

```go
func (app *SimApp) registerUpgradeHandlers() {
	app.UpgradeKeeper.SetUpgradeHandler("v0.43", func(ctx sdk.Context, plan upgradetypes.Plan, _ module.VersionMap) (module.VersionMap, error) {
		// 1st-time running in-store migrations, using 1 as fromVersion to
		// avoid running InitGenesis.
		fromVM := map[string]uint64{
			"auth":         1,
			"bank":         1,
			"capability":   1,
			"crisis":       1,
			"distribution": 1,
			"evidence":     1,
			"gov":          1,
			"mint":         1,
			"params":       1,
			"slashing":     1,
			"staking":      1,
			"upgrade":      1,
			"vesting":      1,
			"ibc":          1,
			"genutil":      1,
			"transfer":     1,
		}

		return app.mm.RunMigrations(ctx, app.configurator, fromVM)
	})

	upgradeInfo, err := app.UpgradeKeeper.ReadUpgradeInfoFromDisk()
	if err != nil {
		panic(err)
	}

	if upgradeInfo.Name == "v0.43" && !app.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		storeUpgrades := storetypes.StoreUpgrades{
			Added: []string{"authz", "feegrant"},
		}

		// configure store loader that checks if version == upgradeHeight and applies store upgrades
		app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &storeUpgrades))
	}
}
```

Add `storetypes` to imports:

```go
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
```

Build the `simd` binary for `v0.43.0` (the upgrade binary):

```
make build
```

Create the cosmovisor genesis folder and copy the `v0.43.0` binary:

```
mkdir -p $DAEMON_HOME/cosmovisor/upgrades/v0.43/bin
cp ./build/simd $DAEMON_HOME/cosmovisor/upgrades/v0.43/bin
```

### Upgrade Proposal

Start the node using cosmovisor:

```
cosmovisor start
```

Submit an upgrade proposal along with a deposit and a vote (these commands must be run within 20 seconds of each other):

```
./build/simd tx gov submit-proposal software-upgrade v0.43 --title upgrade --description upgrade --upgrade-height 20 --from validator --yes
./build/simd tx gov deposit 1 10000000stake --from validator --yes
./build/simd tx gov vote 1 yes --from validator --yes
```

Confirm that the upgrades at height 20.