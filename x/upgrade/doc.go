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

Halt Behavior

Before halting the ABCI state machine in the BeginBlocker method, the upgrade module will log an error
that looks like:
	UPGRADE "<Name>" NEEDED at height <NNNN>: <Info>
where Name are Info are the values of the respective fields on the upgrade Plan.

To perform the actual halt of the blockchain, the upgrade keeper simply panic's which prevents the ABCI state machine
from proceeding but doesn't actually exit the process. Exiting the process can cause issues for other nodes that start
to lose connectivity with the exiting nodes, thus this module prefers to just halt but not exit.

Will Upgrade and On Upgrade Callbacks

The upgrade keeper has two methods for setting callbacks - SetWillUpgrader and SetOnUpgrader. The will upgrade
callback will be called in the BeginBlocker of the first block after an upgrade is scheduled. The on upgrade callback
will be called in the BeginBlocker at the block where an upgrade is needed right before the state machine is halted.
The will upgrade callback can be used to notify some external process that an upgrade is needed so that it can
prepare binaries, etc. The on upgrade callback can notify some external process to actually begin the upgrade process.

*/
package upgrade
