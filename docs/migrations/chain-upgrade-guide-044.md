<!--
order: 1
-->

# Chain Upgrade Guide to v0.44

This document provides guidelines for a chain upgrade from v0.42 to v0.44 and an example of the upgrade process using `simapp`. {synopsis}

::: tip
You must upgrade to Stargate v0.42 before upgrading to v0.44. If you have not done so, please see [Chain Upgrade Guide to v0.42](/v0.42/migrations/chain-upgrade-guide-040.html). Please note, v0.43 was discontinued shortly after being released and all chains should upgrade directly to v0.44 from v0.42.
:::

## Prerequisite Readings

* [Upgrading Modules](../building-modules/upgrade.html) {prereq}
* [In-Place Store Migrations](../core/upgrade.html) {prereq}
* [Cosmovisor](../run-node/cosmovisor.html) {prereq}

Cosmos SDK v0.44 introduces a new way of handling chain upgrades that no longer requires exporting state to JSON, making the necesssary changes, and then creating a new chain with the modified JSON as the new genesis file.

The IBC module for the Cosmos SDK has moved to its [own repository](https://github.com/cosmos/ibc-go) for v0.42 and later versions. If you are using IBC, make sure to also go through the [IBC migration docs](https://github.com/cosmos/ibc-go/blob/main/docs/migrations/ibc-migration-043.md).

Instead of starting a new chain, the upgrade binary will read the existing database and perform in-place store migrations. This new way of handling chain upgrades can be used alongside [Cosmovisor](../run-node/cosmovisor.html) to make the upgrade process seamless.

## In-Place Store Migrations

We recommend using [In-Place Store Migrations](../core/upgrade.html) to upgrade your chain from v0.42 to v0.44. The first step is to make sure all your modules follow the [Module Upgrade Guide](../building-modules/upgrade.html). The second step is to add an [upgrade handler](../core/upgrade.html#running-migrations) to `app.go`.

In this document, we'll provide an example of what the upgrade handler looks like for a chain upgrading module versions for the first time. It's critical to note that the initial consensus version of each module must be set to `1` rather than `0` or else the upgrade handler will re-initialize each module.

In addition to migrating existing modules, the upgrade handler also performs store upgrades for new modules. In the example below, we'll be adding store migrations for two new modules made available in v0.44: `x/authz` and `x/feegrant`.

## Using Cosmovisor

We recommend validators use [Cosmovisor](../run-node/cosmovisor.html), which is a process manager for running application binaries. For security reasons, we recommend validators build their own upgrade binaries rather than enabling the auto-download option. Validators may still choose to use the auto-download option if the necessary security guarantees are in place (i.e. the URL provided in the upgrade proposal for the downloadable upgrade binary includes a proper checksum).

::: tip
If validators would like to enable the auto-download option, and they are currently running an application using Cosmos SDK `v0.42`, they will need to use Cosmovisor [`v0.1`](https://github.com/cosmos/cosmos-sdk/releases/tag/cosmovisor%2Fv0.1.0). Later versions of Cosmovisor do not support Cosmos SDK `v0.42` or earlier if the auto-download option is enabled.
:::

Validators can use the auto-restart option to prevent unnecessary downtime during the upgrade process. The auto-restart option will automatically restart the chain with the upgrade binary once the chain has halted at the proposed upgrade height. With the auto-restart option, validators can prepare the upgrade binary in advance and then relax at the time of the upgrade.

## Migrating app.toml

With the update to `v0.44`, new server configuration options have been added to `app.toml`. The updates include new configuration sections for Rosetta and gRPC Web as well as a new configuration option for State Sync. Check out the default [`app.toml`](https://github.com/cosmos/cosmos-sdk/blob/release/v0.44.x/server/config/toml.go) file in the latest version of `v0.44` for more information.

## Example: Simapp Upgrade

The following example will walk through the upgrade process using `simapp` as our blockchain application. We will be upgrading `simapp` from v0.42 to v0.44. We will be building the upgrade binary ourselves and enabling the auto-restart option.

::: tip
In the following example, we start a new chain from `v0.42`. The binary for this version will be the genesis binary. For validators using Cosmovisor for the first time on an existing chain, either the binary for the current version of the chain should be used as the genesis binary (i.e. the starting binary) or validators should update the `current` symbolic link to point to the upgrade directory. For more information, see [Cosmovisor](../run-node/cosmovisor.html).
:::

### Initial Setup

From within the `cosmos-sdk` repository, check out the latest `v0.42.x` release:

```sh
git checkout release/v0.42.x
```

Build the `simd` binary for the latest `v0.42.x` release (the genesis binary):

```sh
make build
```

Reset `~/.simapp` (never do this in a production environment):

```sh
./build/simd unsafe-reset-all
```

Configure the `simd` binary for testing:

```sh
./build/simd config chain-id test
./build/simd config keyring-backend test
./build/simd config broadcast-mode block
```

Initialize the node and overwrite any previous genesis file (never do this in a production environment):

<!-- TODO: init does not read chain-id from config -->

```sh
./build/simd init test --chain-id test --overwrite
```

Set the minimum gas price to `0stake` in `~/.simapp/config/app.toml`:

```sh
minimum-gas-prices = "0stake"
```

For the purpose of this demonstration, change `voting_period` in `genesis.json` to a reduced time of 20 seconds (`20s`):

```sh
cat <<< $(jq '.app_state.gov.voting_params.voting_period = "20s"' $HOME/.simapp/config/genesis.json) > $HOME/.simapp/config/genesis.json
```

Create a new key for the validator, then add a genesis account and transaction:

<!-- TODO: add-genesis-account does not read keyring-backend from config -->
<!-- TODO: gentx does not read chain-id from config -->

```sh
./build/simd keys add validator
./build/simd add-genesis-account validator 5000000000stake --keyring-backend test
./build/simd gentx validator 1000000stake --chain-id test
./build/simd collect-gentxs
```

Now that our node is initialized and we are ready to start a new `simapp` chain, let's set up `cosmovisor` and the genesis binary.

### Cosmovisor Setup

Install the `cosmovisor` binary:

```sh
go install github.com/cosmos/cosmos-sdk/cosmovisor/cmd/cosmovisor@v0.1.0
```

::: tip
If you are using go `v1.15` or earlier, you will need to change out of the `cosmos-sdk` directory, run `go get` instead of `go install`, and then change back into the `cosmos-sdk` repository.
:::

Set the required environment variables:

```sh
export DAEMON_NAME=simd
export DAEMON_HOME=$HOME/.simapp
```

Set the optional environment variable to trigger an automatic restart:

```sh
export DAEMON_RESTART_AFTER_UPGRADE=true
```

Create the folder for the genesis binary and copy the `v0.42.x` binary:

```sh
mkdir -p $DAEMON_HOME/cosmovisor/genesis/bin
cp ./build/simd $DAEMON_HOME/cosmovisor/genesis/bin
```

Now that `cosmovisor` is installed and the genesis binary has been added, let's add the upgrade handler to `simapp/app.go` and prepare the upgrade binary.

### Chain Upgrade

<!-- TODO: update example to use v0.44.x -->

Check out `release/v0.44.x`:

```sh
git checkout release/v0.44.x
```

Add the following to `simapp/app.go` inside `NewSimApp` and after `app.UpgradeKeeper`:

```go
	app.registerUpgradeHandlers()
```

Add the following to `simapp/app.go` after `NewSimApp` (to learn more about the upgrade handler, see the [In-Place Store Migrations](../core/upgrade.html)):

```go
func (app *SimApp) registerUpgradeHandlers() {
	app.UpgradeKeeper.SetUpgradeHandler("v0.44", func(ctx sdk.Context, plan upgradetypes.Plan, _ module.VersionMap) (module.VersionMap, error) {
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

	if upgradeInfo.Name == "v0.44" && !app.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
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

Build the `simd` binary for `v0.44.x` (the upgrade binary):

```sh
make build
```

Create the folder for the upgrade binary and copy the `v0.44.x` binary:

```sh
mkdir -p $DAEMON_HOME/cosmovisor/upgrades/v0.44/bin
cp ./build/simd $DAEMON_HOME/cosmovisor/upgrades/v0.44/bin
```

Now that we have added the upgrade handler and prepared the upgrade binary, we are ready to start `cosmovisor` and simulate the upgrade proposal process.

### Upgrade Proposal

Start the node using `cosmovisor`:

```sh
cosmovisor start
```

Open a new terminal window and submit an upgrade proposal along with a deposit and a vote (these commands must be run within 20 seconds of each other):

```sh
./build/simd tx gov submit-proposal software-upgrade v0.44 --title upgrade --description upgrade --upgrade-height 20 --from validator --yes
./build/simd tx gov deposit 1 10000000stake --from validator --yes
./build/simd tx gov vote 1 yes --from validator --yes
```

Confirm the chain automatically upgrades at height 20.
