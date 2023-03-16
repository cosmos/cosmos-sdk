package keeper

import (
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
)

// Keeper manages history of public keys per account
type Keeper struct {
	key storetypes.StoreKey
	cdc codec.BinaryCodec
	ak  authkeeper.AccountKeeper
}

func NewKeeper(storeKey storetypes.StoreKey, cdc codec.BinaryCodec, ak authkeeper.AccountKeeper) Keeper {
	return Keeper{
		key: storeKey,
		cdc: cdc,
		ak:  ak,
	}
}

// // GetPubKeyHistory Returns the PubKey history of the account at address by time: involves current pubkey
// func (pk Keeper) GetPubKeyHistory(ctx sdk.Context, addr sdk.AccAddress) ([]types.PubKeyHistory, error) {
// 	entries := []types.PubKeyHistory{}
// 	if pk.ak.GetAccount(ctx, addr) == nil {
// 		return entries, fmt.Errorf("account %s not found", addr.String())
// 	}
// 	iterator := pk.PubKeyHistoryIterator(ctx, addr)
// 	defer iterator.Close()
// 	for ; iterator.Valid(); iterator.Next() {
// 		entry := types.DecodeHistoricalEntry(pk.cdc, iterator.Value())
// 		entries = append(entries, entry)
// 	}
// 	currentEntry := pk.GetCurrentPubKeyEntry(ctx, addr)
// 	entries = append(entries, currentEntry)
// 	return entries, nil
// }
