package slashing

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/tendermint/tendermint/crypto"
)

// Keeper of the slashing store
type Keeper struct {
	storeKey     sdk.StoreKey
	cdc          *wire.Codec
	validatorSet sdk.ValidatorSet

	// codespace
	codespace sdk.CodespaceType
}

// NewKeeper creates a slashing keeper
func NewKeeper(cdc *wire.Codec, key sdk.StoreKey, vs sdk.ValidatorSet, codespace sdk.CodespaceType) Keeper {
	keeper := Keeper{
		storeKey:     key,
		cdc:          cdc,
		validatorSet: vs,
		codespace:    codespace,
	}
	return keeper
}

// handle a validator signing two blocks at the same height
func (k Keeper) handleDoubleSign(ctx sdk.Context, pubkey crypto.PubKey, infractionHeight int64, timestamp int64, power int64) {
	logger := ctx.Logger().With("module", "x/slashing")
	time := ctx.BlockHeader().Time
	age := time - timestamp
	address := sdk.ValAddress(pubkey.Address())

	// Double sign too old
	if age > MaxEvidenceAge {
		logger.Info(fmt.Sprintf("Ignored double sign from %s at height %d, age of %d past max age of %d", pubkey.Address(), infractionHeight, age, MaxEvidenceAge))
		return
	}

	// Double sign confirmed
	logger.Info(fmt.Sprintf("Confirmed double sign from %s at height %d, age of %d less than max age of %d", pubkey.Address(), infractionHeight, age, MaxEvidenceAge))

	// Slash validator
	k.validatorSet.Slash(ctx, pubkey, infractionHeight, power, SlashFractionDoubleSign)

	// Revoke validator
	k.validatorSet.Revoke(ctx, pubkey)

	// Jail validator
	signInfo, found := k.getValidatorSigningInfo(ctx, address)
	if !found {
		panic(fmt.Sprintf("Expected signing info for validator %s but not found", address))
	}
	signInfo.JailedUntil = time + DoubleSignUnbondDuration
	k.setValidatorSigningInfo(ctx, address, signInfo)
}

// handle a validator signature, must be called once per validator per block
func (k Keeper) handleValidatorSignature(ctx sdk.Context, pubkey crypto.PubKey, power int64, signed bool) {
	logger := ctx.Logger().With("module", "x/slashing")
	height := ctx.BlockHeight()
	address := sdk.ValAddress(pubkey.Address())

	// Local index, so counts blocks validator *should* have signed
	// Will use the 0-value default signing info if not present, except for start height
	signInfo, found := k.getValidatorSigningInfo(ctx, address)
	if !found {
		// If this validator has never been seen before, construct a new SigningInfo with the correct start height
		signInfo = NewValidatorSigningInfo(height, 0, 0, 0)
	}
	index := signInfo.IndexOffset % SignedBlocksWindow
	signInfo.IndexOffset++

	// Update signed block bit array & counter
	// This counter just tracks the sum of the bit array
	// That way we avoid needing to read/write the whole array each time
	previous := k.getValidatorSigningBitArray(ctx, address, index)
	if previous == signed {
		// Array value at this index has not changed, no need to update counter
	} else if previous && !signed {
		// Array value has changed from signed to unsigned, decrement counter
		k.setValidatorSigningBitArray(ctx, address, index, false)
		signInfo.SignedBlocksCounter--
	} else if !previous && signed {
		// Array value has changed from unsigned to signed, increment counter
		k.setValidatorSigningBitArray(ctx, address, index, true)
		signInfo.SignedBlocksCounter++
	}

	if !signed {
		logger.Info(fmt.Sprintf("Absent validator %s at height %d, %d signed, threshold %d", pubkey.Address(), height, signInfo.SignedBlocksCounter, MinSignedPerWindow))
	}
	minHeight := signInfo.StartHeight + SignedBlocksWindow
	if height > minHeight && signInfo.SignedBlocksCounter < MinSignedPerWindow {
		// Downtime confirmed, slash, revoke, and jail the validator
		logger.Info(fmt.Sprintf("Validator %s past min height of %d and below signed blocks threshold of %d", pubkey.Address(), minHeight, MinSignedPerWindow))
		k.validatorSet.Slash(ctx, pubkey, height, power, SlashFractionDowntime)
		k.validatorSet.Revoke(ctx, pubkey)
		signInfo.JailedUntil = ctx.BlockHeader().Time + DowntimeUnbondDuration
	}

	// Set the updated signing info
	k.setValidatorSigningInfo(ctx, address, signInfo)
}
