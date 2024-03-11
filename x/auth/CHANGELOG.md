<!--
Guiding Principles:
Changelogs are for humans, not machines.
There should be an entry for every single version.
The same types of changes should be grouped.
Versions and sections should be linkable.
The latest version comes first.
The release date of each version is displayed.
Mention whether you follow Semantic Versioning.
Usage:
Change log entries are to be added to the Unreleased section under the
appropriate stanza (see below). Each entry should ideally include a tag and
the Github issue reference in the following format:
* (<tag>) [#<issue-number>] Changelog message.
Types of changes (Stanzas):
"Features" for new features.
"Improvements" for changes in existing functionality.
"Deprecated" for soon-to-be removed features.
"Bug Fixes" for any bug fixes.
"API Breaking" for breaking exported APIs used by developers building on SDK.
Ref: https://keepachangelog.com/en/1.0.0/
-->

# Changelog

## [Unreleased]

### Features

* [#18641](https://github.com/cosmos/cosmos-sdk/pull/18641) Support the ability to broadcast unordered transactions per ADR-070. See UPGRADING.md for more details on integration.
* [#18281](https://github.com/cosmos/cosmos-sdk/pull/18281) Support broadcasting multiple transactions.
* (vesting) [#17810](https://github.com/cosmos/cosmos-sdk/pull/17810) Add the ability to specify a start time for continuous vesting accounts.

### Improvements

* [#18780](https://github.com/cosmos/cosmos-sdk/pull/18780) Move sig verification out of the for loop, into the authenticate method.
* [#19188](https://github.com/cosmos/cosmos-sdk/pull/19188) Remove creation of `BaseAccount` when sending a message to an account that does not exist. 
    * When signing a transaction with an account that has not been created accountnumber 0 must be used

### CLI Breaking Changes

* (vesting) [#18100](https://github.com/cosmos/cosmos-sdk/pull/18100) `appd tx vesting create-vesting-account` takes an amount of coin as last argument instead of second. Coins are space separated.

### API Breaking Changes

* [#19447](https://github.com/cosmos/cosmos-sdk/pull/19447) Address and validator address codecs are now arguments of `NewTxConfig`. `NewDefaultSigningOptions` has been replaced with `NewSigningOptions` which takes address and validator address codecs as arguments.
* [#17985](https://github.com/cosmos/cosmos-sdk/pull/17985) Remove `StdTxConfig`
* [#19161](https://github.com/cosmos/cosmos-sdk/pull/19161) Remove `simulate` from `SetGasMeter`
* [#19363](https://github.com/cosmos/cosmos-sdk/pull/19363) Remove `IterateAccounts` and `GetAllAccounts` methods from the AccountKeeper interface and Keeper.
* [#19290](https://github.com/cosmos/cosmos-sdk/issues/19290) Pass `appmodule.Environment` to NewKeeper instead of passing individual services. 
* [#19535](https://github.com/cosmos/cosmos-sdk/pull/19535) Remove vesting account creation when the chain is running. The accounts module is required for creating vesting accounts on a running chain. 
<!-- TODO add a link to lockup accounts docs -->

### Consensus Breaking Changes

* [#18817](https://github.com/cosmos/cosmos-sdk/pull/18817) SigVerification, GasConsumption, IncreaseSequence ante decorators have all been joined into one SigVerification decorator. Gas consumption during TX validation flow has reduced.
* [#19093](https://github.com/cosmos/cosmos-sdk/pull/19093) SetPubKeyDecorator was merged into SigVerification, gas consumption is almost halved for a simple tx.

### Bug Fixes

* [#19148](https://github.com/cosmos/cosmos-sdk/pull/19148) Checks the consumed gas for verifying a multisig pubKey signature during simulation.
* [#19239](https://github.com/cosmos/cosmos-sdk/pull/19239) Sets from flag in multi-sign command to avoid no key name provided error.
* [#19099](https://github.com/cosmos/cosmos-sdk/pull/19099) `verifyIsOnCurve` now checks if we are simulating to avoid malformed public key error.
