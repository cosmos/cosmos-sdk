package slashing

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/stake"
	crypto "github.com/tendermint/go-crypto"
)

// Keeper of the slashing store
type Keeper struct {
	storeKey    sdk.StoreKey
	cdc         *wire.Codec
	stakeKeeper stake.Keeper

	// codespace
	codespace sdk.CodespaceType
}

// NewKeeper creates a slashing keeper
func NewKeeper(cdc *wire.Codec, key sdk.StoreKey, sk stake.Keeper, codespace sdk.CodespaceType) Keeper {
	keeper := Keeper{
		storeKey:    key,
		cdc:         cdc,
		stakeKeeper: sk,
		codespace:   codespace,
	}
	return keeper
}

// handle a validator signing two blocks at the same height
func (k Keeper) handleDoubleSign(ctx sdk.Context, height int64, timestamp int64, pubkey crypto.PubKey) {
	logger := ctx.Logger().With("module", "x/slashing")
	age := ctx.BlockHeader().Time - timestamp

	// Double sign too old
	if age > MaxEvidenceAge {
		logger.Info(fmt.Sprintf("Ignored double sign from %s at height %d, age of %d past max age of %d", pubkey.Address(), height, age, MaxEvidenceAge))
		return
	}

	// Double sign confirmed
	logger.Info(fmt.Sprintf("Confirmed double sign from %s at height %d, age of %d less than max age of %d", pubkey.Address(), height, age, MaxEvidenceAge))
	k.stakeKeeper.Slash(ctx, pubkey, height, SlashFractionDoubleSign)
}

// handle a validator signature, must be called once per validator per block
func (k Keeper) handleValidatorSignature(ctx sdk.Context, pubkey crypto.PubKey, signed bool) {
	logger := ctx.Logger().With("module", "x/slashing")
	height := ctx.BlockHeight()
	if !signed {
		logger.Info(fmt.Sprintf("Absent validator %s at height %d", pubkey.Address(), height))
	}
	address := pubkey.Address()

	// Local index, so counts blocks validator *should* have signed
	// Will use the 0-value default signing info if not present
	signInfo, _ := k.getValidatorSigningInfo(ctx, address)
	index := signInfo.IndexOffset % SignedBlocksWindow
	signInfo.IndexOffset++

	// Update signed block bit array & counter
	previous := k.getValidatorSigningBitArray(ctx, address, index)
	if previous == signed {
		// No need to update counter
	} else if previous && !signed {
		// Signed => unsigned, decrement counter
		k.setValidatorSigningBitArray(ctx, address, index, false)
		signInfo.SignedBlocksCounter--
	} else if !previous && signed {
		// Unsigned => signed, increment counter
		k.setValidatorSigningBitArray(ctx, address, index, true)
		signInfo.SignedBlocksCounter++
	}

	minHeight := signInfo.StartHeight + SignedBlocksWindow
	if height > minHeight && signInfo.SignedBlocksCounter < MinSignedPerWindow {
		// Downtime confirmed, slash, revoke, and jail the validator
		logger.Info(fmt.Sprintf("Validator %s past min height of %d and below signed blocks threshold of %d", pubkey.Address(), minHeight, MinSignedPerWindow))
		k.stakeKeeper.Slash(ctx, pubkey, height, SlashFractionDowntime)
		k.stakeKeeper.Revoke(ctx, pubkey)
		signInfo.JailedUntil = ctx.BlockHeader().Time + DowntimeUnbondDuration
	}

	// Set the updated signing info
	k.setValidatorSigningInfo(ctx, address, signInfo)
}
