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

We recommend using [In-Place Store Migrations](../core/upgrade.html) to upgrade your chain from v0.42 to v0.43. The first step is to make sure all your modules follow the [Module Upgrade Guide](../building-modules/upgrade.html). The second step is to add an [upgrade handler](../core/upgrade.html#running-migrations) to `app.go`. In this document, we'll provide an example of what the upgrade handler should look like for a chain upgrading module versions for the first time.

## Preparing Upgrade Binaries

We recommend validators use [Cosmovisor](../run-node/cosmovisor.html) and blockchain application developers prepare a tarball that includes the genesis binary and all available upgrade binaries so that validators can easily download the necessary binaries to run the upgrade. This tarball can also be used to sync a fullnode from start. In this example, we will build the genesis binary and the upgrade binary.

## Example: Simapp Upgrade

The following example will walk through the upgrade process using `simapp`. We will be upgrading `simapp` from v0.42 to v0.43.
 
### Initialize Chain

From within the `cosmos-sdk` repository, check out `v0.42.6`:

```
git checkout v0.42.6
```

Build the `simd` binary:

```
make build
```

Configure and initialize the node:

<!-- TODO: init does not read chain-id from config -->

```
rm -rf $HOME/.simapp
./build/simd config chain-id test
./build/simd config keyring-backend test
./build/simd config broadcast-mode block
./build/simd init test --chain-id test
```

For the purpose of this demontration, amend `voting_period` in `genesis.json` to a reduced time of 10 seconds (`10s`):

```
cat <<< $(jq '.app_state.gov.voting_params.voting_period = "10s"' $HOME/.simapp/config/genesis.json) > $HOME/.simapp/config/genesis.json
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

Now that the chain is ready to be started, let's set up `cosmovisor` and the genesis binary.

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

Create the cosmovisor genesis folder and copy the `v0.42` binary:

```
mkdir -p $DAEMON_HOME/cosmovisor/genesis/bin
cp ./build/simd $DAEMON_HOME/cosmovisor/genesis/bin
```

Now that `cosmovisor` and the genesis binary are ready, let's add the upgrade handler and set up the upgrade binary.

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

Add the following to `simapp/app.go` starting on line 420 (or anywhere in file):

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

Now recompile a new binary and make a copy of it in `$DAEMON_HOME/cosmosvisor/upgrades/v0.43/bin`:

```
make build
mkdir -p $DAEMON_HOME/cosmovisor/upgrades/v0.43/bin
cp ./build/simd $DAEMON_HOME/cosmovisor/upgrades/v0.43/bin
```

### Upgrade Proposal

Start chain using cosmovisor:

```
cosmovisor start
```

Submit a governance proposal:

```
./build/simd tx gov submit-proposal software-upgrade v0.43 --title upgrade --description upgrade --upgrade-height 20 --from validator --yes
./build/simd tx gov deposit 1 10000000stake --from validator --yes
./build/simd tx gov vote 1 yes --from validator --yes
```

Confirm upgrades at height 20.
