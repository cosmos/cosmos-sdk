package keeper

import (
	"time"

	"github.com/tendermint/tendermint/crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/slashing/types"
)

func (k Keeper) AfterValidatorBonded(ctx sdk.Context, address sdk.ConsAddress, _ sdk.ValAddress) error {
	// Update the signing info start height or create a new signing info
	signingInfo, found := k.GetValidatorSigningInfo(ctx, address)
	if found {
		signingInfo.StartHeight = ctx.BlockHeight()
	} else {
		signingInfo = types.NewValidatorSigningInfo(
			address,
			ctx.BlockHeight(),
			0,
			time.Unix(0, 0),
			false,
			0,
		)
	}

	k.SetValidatorSigningInfo(ctx, address, signingInfo)

	return nil
}

// AfterValidatorCreated adds the address-pubkey relation when a validator is created.
func (k Keeper) AfterValidatorCreated(ctx sdk.Context, valAddr sdk.ValAddress) error {
	validator := k.sk.Validator(ctx, valAddr)
	consPk, err := validator.ConsPubKey()
	if err != nil {
		return err
	}

	return k.AddPubkey(ctx, consPk)
}

// AfterValidatorRemoved deletes the address-pubkey relation when a validator is removed,
func (k Keeper) AfterValidatorRemoved(ctx sdk.Context, address sdk.ConsAddress) error {
	k.deleteAddrPubkeyRelation(ctx, crypto.Address(address))
	return nil
}
