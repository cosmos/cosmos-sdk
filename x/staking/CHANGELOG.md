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

### Improvements

* [#19277](https://github.com/cosmos/cosmos-sdk/pull/19277) Hooks calls on `SetUnbondingDelegationEntry`, `SetRedelegationEntry`, `Slash` and `RemoveValidator` returns errors instead of logging just like other hooks calls.
* [#18636](https://github.com/cosmos/cosmos-sdk/pull/18636) `IterateBondedValidatorsByPower`, `GetDelegatorBonded`, `Delegate`, `Unbond`, `Slash`, `Jail`, `SlashRedelegation`, `ApplyAndReturnValidatorSetUpdates` methods no longer panics on any kind of errors but instead returns appropriate errors.
* [#18506](https://github.com/cosmos/cosmos-sdk/pull/18506) Detect the length of the ed25519 pubkey in CreateValidator to prevent panic.

### API Breaking Changes

* [#18198](https://github.com/cosmos/cosmos-sdk/pull/18198): `Validator` and `Delegator` interfaces were moved to `github.com/cosmos/cosmos-sdk/types` to avoid interface dependency on staking in other modules. 
* [#17778](https://github.com/cosmos/cosmos-sdk/pull/17778) Use collections for `Params`
    * remove from `Keeper`: `GetParams`, `SetParams`
* [#17486](https://github.com/cosmos/cosmos-sdk/pull/17486) Use collections for `RedelegationQueueKey`:
    * remove from `types`: `GetRedelegationTimeKey`
    * remove from `Keeper`: `RedelegationQueueIterator`
* [#17562](https://github.com/cosmos/cosmos-sdk/pull/17562) Use collections for `ValidatorQueue`
    * remove from `types`: `GetValidatorQueueKey`, `ParseValidatorQueueKey`
    * remove from `Keeper`: `ValidatorQueueIterator`
* [#17498](https://github.com/cosmos/cosmos-sdk/pull/17498) Use collections for `LastValidatorPower`:
    * remove from `types`: `GetLastValidatorPowerKey`
    * remove from `Keeper`: `LastValidatorsIterator`, `IterateLastValidators`
* [#17291](https://github.com/cosmos/cosmos-sdk/pull/17291) Use collections for `UnbondingDelegationByValIndex`:
    * remove from `types`: `GetUBDKeyFromValIndexKey`, `GetUBDsByValIndexKey`, `GetUBDByValIndexKey`
* (x/slashing) [#17568](https://github.com/cosmos/cosmos-sdk/pull/17568) Use collections for `ValidatorMissedBlockBitmap`:
    * remove from `types`: `ValidatorMissedBlockBitmapPrefixKey`, `ValidatorMissedBlockBitmapKey`
* [#17481](https://github.com/cosmos/cosmos-sdk/pull/17481) Use collections for `UnbondingQueue`:
    * remove from `Keeper`: `UBDQueueIterator`
    * remove from `types`: `GetUnbondingDelegationTimeKey`
* [#17123](https://github.com/cosmos/cosmos-sdk/pull/17123) Use collections for `Validators`
* [#17270](https://github.com/cosmos/cosmos-sdk/pull/17270) Use collections for `UnbondingDelegation`:
    * remove from `types`: `GetUBDsKey`
    * remove from `Keeper`: `IterateUnbondingDelegations`, `IterateDelegatorUnbondingDelegations`
* [#17336](https://github.com/cosmos/cosmos-sdk/pull/17336) Use collections for `RedelegationByValDstIndexKey`:
    * remove from `types`: `GetREDByValDstIndexKey`, `GetREDsToValDstIndexKey`
* [#17332](https://github.com/cosmos/cosmos-sdk/pull/17332) Use collections for `RedelegationByValSrcIndexKey`:
    * remove from `types`: `GetREDKeyFromValSrcIndexKey`, `GetREDsFromValSrcIndexKey`
* [#17315](https://github.com/cosmos/cosmos-sdk/pull/17315) Use collections for `RedelegationKey`:
    * remove from `keeper`: `GetRedelegation`
* [#17260](https://github.com/cosmos/cosmos-sdk/pull/17260) Use collections for `DelegationKey`:
    * remove from `types`: `GetDelegationKey`, `GetDelegationsKey`
* [#17288](https://github.com/cosmos/cosmos-sdk/pull/17288) Use collections for `UnbondingIndex`:
    * remove from `types`: `GetUnbondingIndexKey`.
* [#17256](https://github.com/cosmos/cosmos-sdk/pull/17256) Use collections for `UnbondingID`.
* [#17260](https://github.com/cosmos/cosmos-sdk/pull/17260) Use collections for `ValidatorByConsAddr`:
    * remove from `types`: `GetValidatorByConsAddrKey`
* [#17248](https://github.com/cosmos/cosmos-sdk/pull/17248) Use collections for `UnbondingType`.
    * remove from `types`: `GetUnbondingTypeKey`.
* [#17063](https://github.com/cosmos/cosmos-sdk/pull/17063) Use collections for `HistoricalInfo`:
    * remove `Keeper`: `GetHistoricalInfo`, `SetHistoricalInfo`
* [#17062](https://github.com/cosmos/cosmos-sdk/pull/17062) Use collections for `ValidatorUpdates`:
    * remove `Keeper`: `SetValidatorUpdates`, `GetValidatorUpdates`
* [#17026](https://github.com/cosmos/cosmos-sdk/pull/17026) Use collections for `LastTotalPower`:
    * remove `Keeper`: `SetLastTotalPower`, `GetLastTotalPower`
* [#17335](https://github.com/cosmos/cosmos-sdk/pull/17335) Remove usage of `"cosmossdk.io/x/staking/types".Infraction_*` in favour of `"cosmossdk.io/api/cosmos/staking/v1beta1".Infraction_` in order to remove dependency between modules on staking
* [#17655](https://github.com/cosmos/cosmos-sdk/pull/17655) `QueryHistoricalInfo` was adjusted to return `HistoricalRecord` and marked `Hist` as deprecated.
* [#19414](https://github.com/cosmos/cosmos-sdk/pull/19414) Staking module takes an environment variable in `NewStakingKeeper` instead of individual services.

### State Breaking changes

* [#18142](https://github.com/cosmos/cosmos-sdk/pull/18142) Introduce `key_rotation_fee` param to calculate fees while rotating the keys
* [#17655](https://github.com/cosmos/cosmos-sdk/pull/17655) `HistoricalInfo` was replaced with `HistoricalRecord`, it removes the validator set and comet header and only keep what is needed for IBC.

### Bug Fixes
