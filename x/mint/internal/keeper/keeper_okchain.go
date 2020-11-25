package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/mint/internal/types"
)

func (k Keeper) AddYieldFarming(ctx sdk.Context, yieldAmt sdk.Coins) error {
	// todo: verify farmModuleName
	if len(k.farmModuleName) == 0 {
		return nil
	}
	return k.supplyKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, k.farmModuleName, yieldAmt)
}

// get the minter custom
func (k Keeper) GetMinterCustom(ctx sdk.Context) (minter types.MinterCustom) {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(types.MinterKey)
	if b != nil {
		k.cdc.MustUnmarshalBinaryLengthPrefixed(b, &minter)
	}
	return
}

// set the minter custom
func (k Keeper) SetMinterCustom(ctx sdk.Context, minter types.MinterCustom) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshalBinaryLengthPrefixed(minter)
	store.Set(types.MinterKey, b)
}

func (k Keeper) UpdateMinterCustom(ctx sdk.Context, minter *types.MinterCustom, params types.Params) {
	var provisionAmtPerBlock sdk.Dec
	if ctx.BlockHeight() == 0 || minter.NextBlockToUpdate == 0 {
		provisionAmtPerBlock = k.GetOriginalMintedPerBlock()
	} else {
		provisionAmtPerBlock = minter.MintedPerBlock.AmountOf(params.MintDenom).Mul(params.DeflationRate)
	}

	// update new MinterCustom
	minter.MintedPerBlock = sdk.NewDecCoinsFromDec(params.MintDenom, provisionAmtPerBlock)
	minter.NextBlockToUpdate += params.DeflationEpoch * params.BlocksPerYear

	k.SetMinterCustom(ctx, *minter)
}


//______________________________________________________________________

// GetOriginalMintedPerBlock returns the init tokens per block.
func (k Keeper) GetOriginalMintedPerBlock() sdk.Dec {
	return k.originalMintedPerBlock
}

// SetOriginalMintedPerBlock sets the init tokens per block.
func (k Keeper) SetOriginalMintedPerBlock(originalMintedPerBlock sdk.Dec) {
	k.originalMintedPerBlock = originalMintedPerBlock
}


// ValidateMinterCustom validate minter
func ValidateOriginalMintedPerBlock(originalMintedPerBlock sdk.Dec) error {
	if originalMintedPerBlock.IsNegative() {
		panic("init tokens per block must be non-negative")
	}

	return nil
}