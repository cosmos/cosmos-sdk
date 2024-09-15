package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/core/address"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/x/nft"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Keeper of the nft store
type Keeper struct {
	appmodule.Environment
	cdc codec.BinaryCodec
	bk  nft.BankKeeper
	ac  address.Codec
}

// NewKeeper creates a new nft Keeper instance
func NewKeeper(env appmodule.Environment,
	cdc codec.BinaryCodec, ak nft.AccountKeeper, bk nft.BankKeeper,
) Keeper {
	// ensure nft module account is set
	if addr := ak.GetModuleAddress(nft.ModuleName); addr == nil {
		panic("the nft module account has not been set")
	}
	return Keeper{
		Environment: env,
		cdc:         cdc,
		bk:          bk,
		ac:          ak.AddressCodec(),
	}
}

// Stake locks an NFT for a specified duration
func (k Keeper) Stake(ctx context.Context, classId string, nftId string, owner sdk.AccAddress, stakeDuration uint64) error {
	// Implementation of staking logic
	// This is a placeholder and needs to be implemented based on your requirements
	return nil
}

// New method to set the creator
func (k Keeper) setCreator(ctx context.Context, classID string, nftID string, creator string) {
	store := k.KVStoreService.OpenKVStore(ctx)
	key := creatorStoreKey(classID, nftID)
	store.Set(key, []byte(creator))
}

// Helper function to generate the creator store key
func creatorStoreKey(classID, nftID string) []byte {
	return []byte(fmt.Sprintf("%s/creator/%s/%s", nft.ModuleName, classID, nftID))
}

// You can add any additional methods or logic here that are not already defined in nft.go or msg_server.go
