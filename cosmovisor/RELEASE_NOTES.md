# Cosmovisor v1.0.0 Release Notes

This is the first major release of Cosmovisor.
It changes the way Cosmovisor is searching for an upgrade event from an app.
Instead of scanning standard input and standard output logs, the Cosmovisor
observes the `$DAEMON_HOME/upgrade-info.json` file, that is produced by the
`x/upgrade` module. The `upgrade-info.json` files is created by the `x/upgrade`
module and contains information from the on-chain upgrade Plan record.
Using the file based approach solved many outstanding problems: freezing when
logs are too long, race condition with the `x/upgrade` handler, and potential
exploit (if a chain would allow to log an arbitrary message, then an attacker
could produce a fake upgrade signal and halt a chain or instrument a download
of modified, hacked binary when the auto download option is enabled).

## Auto downloads

Cosmovisor v1.0 supports auto downloads based on the information in the
`data/upgrade-info.json`. In the Cosmos SDK `< v0.44`, that file doesn't contain
`upgrade.Plan.Info`, that is needed for doing auto download. Hence Cosmovisor `v1.0`
auto download won't work with Apps updating from `v0.43` and earlier.

NOTE: we **don't recommend using auto download** functionality. It can lead to potential
chain halt when the upgrade Plan contains a bad link or the resource with the
binary will be temporarily unavailable. We are planning on adding a upgrade
verification command which can potentially solve this issue.

## Other updates

+ Changed default value of `DAEMON_RESTART_AFTER_UPGRADE` to `true`.
+ Added `version` command, which prints both the Cosmovisor and the associated app version.
+ Added `help`  command, which prints the Cosmovisor help without passing it to the associated version. This is an exception, because normally, Cosmovisor passes all arguments to the associated app.

For more details, please see the [CHANGELOG](https://github.com/cosmos/cosmos-sdk/blob/cosmovisor/v1.0.0/cosmovisor/CHANGELOG.md).
