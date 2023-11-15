package keeper

import (
	"context"
	"errors"

	"cosmossdk.io/collections"
	"github.com/cosmos/cosmos-sdk/x/auth/types"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewAccountWithAddress implements AccountKeeperI.
func (ak AccountKeeper) NewAccountWithAddress(ctx context.Context, addr sdk.AccAddress) sdk.AccountI {
	acc := ak.proto()
	err := acc.SetAddress(addr)
	if err != nil {
		panic(err)
	}

	return ak.NewAccount(ctx, acc)
}

// NewAccount sets the next account number to a given account interface
func (ak AccountKeeper) NewAccount(ctx context.Context, acc sdk.AccountI) sdk.AccountI {
	if err := acc.SetAccountNumber(ak.NextAccountNumber(ctx)); err != nil {
		panic(err)
	}

	return acc
}

// HasAccount implements AccountKeeperI.
func (ak AccountKeeper) HasAccount(ctx context.Context, addr sdk.AccAddress) bool {
	has, _ := ak.Accounts.Has(ctx, addr)

	if !has {
		cosmosAddr := ak.GetCorrespondingCosmosAddressIfExists(ctx, addr)
		if cosmosAddr == nil {
			return false
		}
		has, _ = ak.Accounts.Has(ctx, cosmosAddr)
	}
	return has
}

// HasExactAccount implements AccountKeeperI.
// Checks if account exists based on address directly, doesn't check for mapping.
// Original cosmos implementation of HasAccount
func (ak AccountKeeper) HasExactAccount(ctx context.Context, addr sdk.AccAddress) bool {
	has, _ := ak.Accounts.Has(ctx, addr)
	return has
}

// IsModuleAccount implements AccountKeeperI.
func (ak AccountKeeper) IsModuleAccount(ctx context.Context, addr sdk.AccAddress) bool {
	acc := ak.GetAccount(ctx, addr)
	if acc != nil {
		_, isModuleAccount := acc.(types.ModuleAccountI)
		return isModuleAccount
	}
	return false
}

// GetAccount implements AccountKeeperI.
func (ak AccountKeeper) GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI {
	acc, err := ak.Accounts.Get(ctx, addr)
	if err != nil && !errors.Is(err, collections.ErrNotFound) {
		cosmosAddr := ak.GetCorrespondingCosmosAddressIfExists(ctx, addr)
		if cosmosAddr == nil {
			return nil
		}
		acc, err = ak.Accounts.Get(ctx, cosmosAddr)
		if err != nil && !errors.Is(err, collections.ErrNotFound) {
			panic(err)
		}
	}
	return acc
}

// GetAllAccounts returns all accounts in the accountKeeper.
func (ak AccountKeeper) GetAllAccounts(ctx context.Context) (accounts []sdk.AccountI) {
	ak.IterateAccounts(ctx, func(acc sdk.AccountI) (stop bool) {
		accounts = append(accounts, acc)
		return false
	})

	return accounts
}

// SetAccount implements AccountKeeperI.
func (ak AccountKeeper) SetAccount(ctx context.Context, acc sdk.AccountI) {
	err := ak.Accounts.Set(ctx, acc.GetAddress(), acc)
	if err != nil {
		panic(err)
	}
}

// RemoveAccount removes an account for the account mapper store.
// NOTE: this will cause supply invariant violation if called
func (ak AccountKeeper) RemoveAccount(ctx context.Context, acc sdk.AccountI) {
	err := ak.Accounts.Remove(ctx, acc.GetAddress())
	if err != nil {
		panic(err)
	}
}

// IterateAccounts iterates over all the stored accounts and performs a callback function.
// Stops iteration when callback returns true.
func (ak AccountKeeper) IterateAccounts(ctx context.Context, cb func(account sdk.AccountI) (stop bool)) {
	err := ak.Accounts.Walk(ctx, nil, func(_ sdk.AccAddress, value sdk.AccountI) (bool, error) {
		return cb(value), nil
	})
	if err != nil {
		panic(err)
	}
}

func (ak AccountKeeper) GetCorrespondingEthAddressIfExists(ctx context.Context, cosmosAddr sdk.AccAddress) (correspondingEthAddr sdk.AccAddress) {
	if cosmosAddr == nil {
		return nil
	}
	mapping := ak.Store(ctx, types.CosmosAddressToEthAddressKey)
	return mapping.Get(cosmosAddr)
}

func (ak AccountKeeper) GetCorrespondingCosmosAddressIfExists(ctx context.Context, ethAddr sdk.AccAddress) (correspondingCosmosAddr sdk.AccAddress) {
	if ethAddr == nil {
		return nil
	}
	mapping := ak.Store(ctx, types.EthAddressToCosmosAddressKey)
	return mapping.Get(ethAddr)

}

func (ak AccountKeeper) SetCorrespondingAddresses(ctx context.Context, cosmosAddr sdk.AccAddress, ethAddr sdk.AccAddress) {
	ak.AddToEthToCosmosAddressMap(ctx, ethAddr, cosmosAddr)
	ak.AddToCosmosToEthAddressMap(ctx, cosmosAddr, ethAddr)

}

func (ak AccountKeeper) AddToCosmosToEthAddressMap(ctx context.Context, cosmosAddr sdk.AccAddress, ethAddr sdk.AccAddress) {
	cosmosAddrToEthAddrMapping := ak.Store(ctx, types.CosmosAddressToEthAddressKey)
	cosmosAddrToEthAddrMapping.Set(cosmosAddr, ethAddr)
}

func (ak AccountKeeper) AddToEthToCosmosAddressMap(ctx context.Context, ethAddr sdk.AccAddress, cosmosAddr sdk.AccAddress) {
	ethAddrToCosmosAddrMapping := ak.Store(ctx, types.EthAddressToCosmosAddressKey)
	ethAddrToCosmosAddrMapping.Set(ethAddr, cosmosAddr)
}

func (ak AccountKeeper) IterateEthToCosmosAddressMapping(ctx context.Context, cb func(ethAddress, cosmosAddress sdk.AccAddress) bool) {
	store := ak.storeService.OpenKVStore(ctx)

	iterator := storetypes.KVStorePrefixIterator(runtime.KVStoreAdapter(store), types.KeyPrefix(types.EthAddressToCosmosAddressKey))

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		//Guard to prevent out of index panic
		if len(iterator.Key()) > len(types.KeyPrefix(types.EthAddressToCosmosAddressKey)) {
			addressKey := iterator.Key()[len(types.KeyPrefix(types.EthAddressToCosmosAddressKey)):]
			if cb(addressKey, iterator.Value()) {
				break
			}
		}
	}

}
func (ak AccountKeeper) IterateCosmosToEthAddressMapping(ctx context.Context, cb func(cosmosAddress, ethAddress sdk.AccAddress) bool) {
	store := ak.storeService.OpenKVStore(ctx)
	iterator := storetypes.KVStorePrefixIterator(runtime.KVStoreAdapter(store), types.KeyPrefix(types.CosmosAddressToEthAddressKey))

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		//Guard to prevent out of index panic
		if len(iterator.Key()) > len(types.KeyPrefix(types.CosmosAddressToEthAddressKey)) {
			addressKey := iterator.Key()[len(types.KeyPrefix(types.CosmosAddressToEthAddressKey)):]
			if cb(addressKey, iterator.Value()) {
				break
			}
		}

	}
}

// GetMergedAccountAddressIfExists gets merged cosmos account address if exists , else returns address passed in
func (ak AccountKeeper) GetMergedAccountAddressIfExists(ctx context.Context, addr sdk.AccAddress) sdk.AccAddress {
	acct := ak.GetAccount(ctx, addr)
	if acct == nil {
		return addr
	}
	return acct.GetAddress()
}

// GetMappedAddress gets corresponding eth address if exists, else tries to get corresponding cosmos address. If both don't exist, it returns nil
func (ak AccountKeeper) GetMappedAddress(ctx context.Context, addr sdk.AccAddress) sdk.AccAddress {
	if address := ak.GetCorrespondingEthAddressIfExists(ctx, addr); address != nil {
		return address
	}
	if address := ak.GetCorrespondingCosmosAddressIfExists(ctx, addr); address != nil {
		return address
	}
	return nil
}
