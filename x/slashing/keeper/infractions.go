package keeper

import (
	"fmt"

	"github.com/tendermint/tendermint/crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/slashing/types"
)

// HandleValidatorSignature handles a validator signature, must be called once per validator per block.
func (k Keeper) HandleValidatorSignature(ctx sdk.Context, addr crypto.Address, power int64, signed bool) {
	logger := k.Logger(ctx)
	height := ctx.BlockHeight()

	// fetch the validator public key
	consAddr := sdk.ConsAddress(addr)
	if _, err := k.GetPubkey(ctx, addr); err != nil {
		panic(fmt.Sprintf("Validator consensus-address %s not found", consAddr))
	}

	// fetch signing info
	signInfo, found := k.GetValidatorSigningInfo(ctx, consAddr)
	if !found {
		panic(fmt.Sprintf("Expected signing info for validator %s but not found", consAddr))
	}

	// this is a relative index, so it counts blocks the validator *should* have signed
	// will use the 0-value default signing info if not present, except for start height
	index := signInfo.IndexOffset % k.SignedBlocksWindow(ctx)
	signInfo.IndexOffset++

	// Update signed block bit array & counter
	// This counter just tracks the sum of the bit array
	// That way we avoid needing to read/write the whole array each time
	previous := k.GetValidatorMissedBlockBitArray(ctx, consAddr, index)
	missed := !signed
	switch {
	case !previous && missed:
		// Array value has changed from not missed to missed, increment counter
		k.SetValidatorMissedBlockBitArray(ctx, consAddr, index, true)
		signInfo.MissedBlocksCounter++
	case previous && !missed:
		// Array value has changed from missed to not missed, decrement counter
		k.SetValidatorMissedBlockBitArray(ctx, consAddr, index, false)
		signInfo.MissedBlocksCounter--
	default:
		// Array value at this index has not changed, no need to update counter
	}

	if missed {
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeLiveness,
				sdk.NewAttribute(types.AttributeKeyAddress, consAddr.String()),
				sdk.NewAttribute(types.AttributeKeyMissedBlocks, fmt.Sprintf("%d", signInfo.MissedBlocksCounter)),
				sdk.NewAttribute(types.AttributeKeyHeight, fmt.Sprintf("%d", height)),
			),
		)

		logger.Info(
			fmt.Sprintf("Absent validator %s at height %d, %d missed, threshold %d", consAddr, height, signInfo.MissedBlocksCounter, k.MinSignedPerWindow(ctx)))
	}

	minHeight := signInfo.StartHeight + k.SignedBlocksWindow(ctx)
	maxMissed := k.SignedBlocksWindow(ctx) - k.MinSignedPerWindow(ctx)

	// if we are past the minimum height and the validator has missed too many blocks, punish them
	if height > minHeight && signInfo.MissedBlocksCounter > maxMissed {
		validator := k.sk.ValidatorByConsAddr(ctx, consAddr)
		if validator != nil && !validator.IsJailed() {

			// Downtime confirmed: slash and jail the validator
			logger.Info(fmt.Sprintf("Validator %s past min height of %d and below signed blocks threshold of %d",
				consAddr, minHeight, k.MinSignedPerWindow(ctx)))

			// We need to retrieve the stake distribution which signed the block, so we subtract ValidatorUpdateDelay from the evidence height,
			// and subtract an additional 1 since this is the LastCommit.
			// Note that this *can* result in a negative "distributionHeight" up to -ValidatorUpdateDelay-1,
			// i.e. at the end of the pre-genesis block (none) = at the beginning of the genesis block.
			// That's fine since this is just used to filter unbonding delegations & redelegations.
			distributionHeight := height - sdk.ValidatorUpdateDelay - 1

			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					types.EventTypeSlash,
					sdk.NewAttribute(types.AttributeKeyAddress, consAddr.String()),
					sdk.NewAttribute(types.AttributeKeyPower, fmt.Sprintf("%d", power)),
					sdk.NewAttribute(types.AttributeKeyReason, types.AttributeValueMissingSignature),
					sdk.NewAttribute(types.AttributeKeyJailed, consAddr.String()),
				),
			)
			k.sk.Slash(ctx, consAddr, distributionHeight, power, k.SlashFractionDowntime(ctx))
			k.sk.Jail(ctx, consAddr)

			signInfo.JailedUntil = ctx.BlockHeader().Time.Add(k.DowntimeJailDuration(ctx))

			// We need to reset the counter & array so that the validator won't be immediately slashed for downtime upon rebonding.
			signInfo.MissedBlocksCounter = 0
			signInfo.IndexOffset = 0
			k.clearValidatorMissedBlockBitArray(ctx, consAddr)
		} else {
			// Validator was (a) not found or (b) already jailed, don't slash
			logger.Info(
				fmt.Sprintf("Validator %s would have been slashed for downtime, but was either not found in store or already jailed", consAddr),
			)
		}
	}

	// Set the updated signing info
	k.SetValidatorSigningInfo(ctx, consAddr, signInfo)
}
