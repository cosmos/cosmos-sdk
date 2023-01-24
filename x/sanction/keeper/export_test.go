package keeper

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/sanction"
)

// This file is available only to unit tests and houses functions for doing
// things with private keeper package stuff.

var (
	// OnlyTestsConcatBzPlusCap, for unit tests, exposes the concatBzPlusCap function.
	OnlyTestsConcatBzPlusCap = concatBzPlusCap

	// OnlyTestsToCoinsOrDefault, for unit tests, exposes the toCoinsOrDefault function.
	OnlyTestsToCoinsOrDefault = toCoinsOrDefault

	// OnlyTestsToAccAddrs, for unit tests, exposes the toAccAddrs function.
	OnlyTestsToAccAddrs = toAccAddrs
)

// OnlyTestsWithGovKeeper, for unit tests, creates a copy of this, setting the govKeeper to the provided one.
func (k Keeper) OnlyTestsWithGovKeeper(govKeeper sanction.GovKeeper) Keeper {
	k.govKeeper = govKeeper
	return k
}

// OnlyTestsWithAuthority, for unit tests, creates a copy of this, setting the authority to the provided one.
func (k Keeper) OnlyTestsWithAuthority(authority string) Keeper {
	k.authority = authority
	return k
}

// OnlyTestsWithUnsanctionableAddrs, for unit tests, creates a copy of this, setting the unsanctionableAddrs to the provided one.
// This does not add the provided ones to the unsanctionableAddrs, it overwrites the
// existing ones with the ones provided.
func (k Keeper) OnlyTestsWithUnsanctionableAddrs(unsanctionableAddrs map[string]bool) Keeper {
	k.unsanctionableAddrs = unsanctionableAddrs
	return k
}

// OnlyTestsGetStoreKey, for unit tests, exposes this keeper's storekey.
func (k Keeper) OnlyTestsGetStoreKey() storetypes.StoreKey {
	return k.storeKey
}

// OnlyTestsGetMsgSanctionTypeURL, for unit tests, exposes this keeper's msgSanctionTypeURL.
func (k Keeper) OnlyTestsGetMsgSanctionTypeURL() string {
	return k.msgSanctionTypeURL
}

// OnlyTestsGetMsgUnsanctionTypeURL, for unit tests, exposes this keeper's msgUnsanctionTypeURL.
func (k Keeper) OnlyTestsGetMsgUnsanctionTypeURL() string {
	return k.msgUnsanctionTypeURL
}

// OnlyTestsGetMsgExecLegacyContentTypeURL, for unit tests, exposes this keeper's msgExecLegacyContentTypeURL.
func (k Keeper) OnlyTestsGetMsgExecLegacyContentTypeURL() string {
	return k.msgExecLegacyContentTypeURL
}

// OnlyTestsGetParamAsCoinsOrDefault, for unit tests, exposes this keeper's getParamAsCoinsOrDefault function.
func (k Keeper) OnlyTestsGetParamAsCoinsOrDefault(ctx sdk.Context, name string, dflt sdk.Coins) sdk.Coins {
	return k.getParamAsCoinsOrDefault(ctx, name, dflt)
}

// OnlyTestsGetLatestTempEntry, for unit tests, exposes this keeper's getLatestTempEntry function.
func (k Keeper) OnlyTestsGetLatestTempEntry(store sdk.KVStore, addr sdk.AccAddress) []byte {
	return k.getLatestTempEntry(store, addr)
}

// OnlyTestsGetParam, for unit tests, exposes this keeper's getParam function.
func (k Keeper) OnlyTestsGetParam(store sdk.KVStore, name string) (string, bool) {
	return k.getParam(store, name)
}

// OnlyTestsSetParam, for unit tests, exposes this keeper's setParam function.
func (k Keeper) OnlyTestsSetParam(store sdk.KVStore, name, value string) {
	k.setParam(store, name, value)
}

// OnlyTestsDeleteParam, for unit tests, exposes this keeper's deleteParam function.
func (k Keeper) OnlyTestsDeleteParam(store sdk.KVStore, name string) {
	k.deleteParam(store, name)
}

// OnlyTestsProposalGovHook, for unit tests, exposes this keeper's proposalGovHook function.
func (k Keeper) OnlyTestsProposalGovHook(ctx sdk.Context, proposalID uint64) {
	k.proposalGovHook(ctx, proposalID)
}

// OnlyTestsIsModuleGovHooksMsgURL, for unit tests, exposes this keeper's isModuleGovHooksMsgURL function.
func (k Keeper) OnlyTestsIsModuleGovHooksMsgURL(url string) bool {
	return k.isModuleGovHooksMsgURL(url)
}

// OnlyTestsGetMsgAddresses, for unit tests, exposes this keeper's getMsgAddresses function.
func (k Keeper) OnlyTestsGetMsgAddresses(msg *codectypes.Any) []sdk.AccAddress {
	return k.getMsgAddresses(msg)
}

// OnlyTestsGetImmediateMinDeposit, for unit tests, exposes this keeper's getImmediateMinDeposit function.
func (k Keeper) OnlyTestsGetImmediateMinDeposit(ctx sdk.Context, msg *codectypes.Any) sdk.Coins {
	return k.getImmediateMinDeposit(ctx, msg)
}
