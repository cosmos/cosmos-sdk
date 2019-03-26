/*
Package upgrade provides a Cosmos SDK module that can be used for smoothly upgrading a live Cosmos chain to a
new software version. It accomplishes this by providing a BeginBlocker hook that prevents the blockchain state
machine from proceeding once a pre-defined upgrade block time or height has been reached. The module does not prescribe
anything regarding how governance decides to do an upgrade, but just the mechanism for coordinating the upgrade safely.
Without software support for upgrades, upgrading a live chain is risky because all of the validators need to pause
their state machines at exactly the same point in the process. If this is not done correctly, there can be state
inconsistencies which are hard to recover from.

Integrating With An App

Setup an upgrade Keeper for the app and then define a BeginBlocker that calls the upgrade
keeper's BeginBlocker method:
    func (app *myApp) BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock) abci.ResponseBeginBlock {
    	app.upgradeKeeper.BeginBlocker(ctx, req)
    	return abci.ResponseBeginBlock{}
    }

The app must then integrate the upgrade keeper with its governance module as appropriate. The governance module
should call ScheduleUpgrade to schedule an upgrade and ClearUpgradePlan to cancel a pending upgrade.

Performing Upgrades

Upgrades can be scheduled at either a predefined block height or time. Once this block height or time is reached, the
existing software will cease to process ABCI messages and a new version with code that handles the upgrade must be deployed.
All upgrades are coordinated by a unique upgrade name that cannot be reused on the same blockchain. In order for the upgrade
module to know that the upgrade has been safely applied, a handler with the name of the upgrade must be installed.
Here is an example handler for an upgrade named "my-fancy-upgrade":
	app.upgradeKeeper.SetUpgradeHandler("my-fancy-upgrade", func(ctx sdk.Context, plan upgrade.Plan) {
		// Perform any migrations of the state store needed for this upgrade
	})

This upgrade handler performs the dual function of alerting the upgrade module that the named upgrade has been applied,
as well as providing the opportunity for the upgraded software to perform any necessary state migrations.

Default and Custom Shutdown Behavior

Before "crashing" the ABCI state machine in the BeginBlocker method, the upgrade module will log an error
that looks like:
	UPGRADE "<Name>" NEEDED at height <NNNN>: <Info>
where Name are Info are the values of the respective fields on the upgrade Plan.

The default method for "crashing" the state machine is to panic with an error which will stop the state machine
but usually not exit the process. This behavior can be overridden using the keeper's SetDoShutdowner method. A custom
shutdown method could perform some other steps such as alerting another running process that an upgrade is needed,
as well as crashing the program by another method such as os.Exit().
*/
package upgrade
