package keeper

import (
	"context"
	"fmt"

	"github.com/cockroachdb/errors"

	st "cosmossdk.io/api/cosmos/staking/v1beta1"
	"cosmossdk.io/core/comet"
	"cosmossdk.io/core/event"
	"cosmossdk.io/x/slashing/types"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// HandleValidatorSignature handles a validator signature, must be called once per validator per block.
func (k Keeper) HandleValidatorSignature(ctx context.Context, addr cryptotypes.Address, power int64, signed comet.BlockIDFlag) error {
	params, err := k.Params.Get(ctx)
	if err != nil {
		return err
	}
	return k.HandleValidatorSignatureWithParams(ctx, params, addr, power, signed)
}

func (k Keeper) HandleValidatorSignatureWithParams(ctx context.Context, params types.Params, addr cryptotypes.Address, power int64, signed comet.BlockIDFlag) error {
	logger := k.Logger(ctx)
	height := k.environment.HeaderService.GetHeaderInfo(ctx).Height

	// fetch the validator public key
	consAddr := sdk.ConsAddress(addr)

	// don't update missed blocks when validator's jailed
	val, err := k.sk.ValidatorByConsAddr(ctx, consAddr)
	if err != nil {
		return err
	}

	if val.IsJailed() {
		return nil
	}

	// read the cons address again because validator may've rotated it's key
	valConsAddr, err := val.GetConsAddr()
	if err != nil {
		return err
	}

	consAddr = sdk.ConsAddress(valConsAddr)

	// fetch signing info
	signInfo, err := k.ValidatorSigningInfo.Get(ctx, consAddr)
	if err != nil {
		return err
	}

	signedBlocksWindow := params.SignedBlocksWindow

	// Compute the relative index, so we count the blocks the validator *should*
	// have signed. We will use the 0-value default signing info if not present,
	// except for start height. The index is in the range [0, SignedBlocksWindow)
	// and is used to see if a validator signed a block at the given height, which
	// is represented by a bit in the bitmap.
	index := signInfo.IndexOffset % signedBlocksWindow
	signInfo.IndexOffset++

	// determine if the validator signed the previous block
	previous, err := k.GetMissedBlockBitmapValue(ctx, consAddr, index)
	if err != nil {
		return errors.Wrap(err, "failed to get the validator's bitmap value")
	}

	missed := signed == comet.BlockIDFlagAbsent
	switch {
	case !previous && missed:
		// Bitmap value has changed from not missed to missed, so we flip the bit
		// and increment the counter.
		if err := k.SetMissedBlockBitmapValue(ctx, consAddr, index, true); err != nil {
			return err
		}

		signInfo.MissedBlocksCounter++

	case previous && !missed:
		// Bitmap value has changed from missed to not missed, so we flip the bit
		// and decrement the counter.
		if err := k.SetMissedBlockBitmapValue(ctx, consAddr, index, false); err != nil {
			return err
		}

		signInfo.MissedBlocksCounter--

	default:
		// bitmap value at this index has not changed, no need to update counter
	}

	minSignedPerWindow := params.MinSignedPerWindowInt()

	consStr, err := k.sk.ConsensusAddressCodec().BytesToString(consAddr)
	if err != nil {
		return err
	}

	if missed {
		if err := k.environment.EventService.EventManager(ctx).EmitKV(
			types.EventTypeLiveness,
			event.NewAttribute(types.AttributeKeyAddress, consStr),
			event.NewAttribute(types.AttributeKeyMissedBlocks, fmt.Sprintf("%d", signInfo.MissedBlocksCounter)),
			event.NewAttribute(types.AttributeKeyHeight, fmt.Sprintf("%d", height)),
		); err != nil {
			return err
		}

		logger.Debug(
			"absent validator",
			"height", height,
			"validator", consStr,
			"missed", signInfo.MissedBlocksCounter,
			"threshold", minSignedPerWindow,
		)
	}

	minHeight := signInfo.StartHeight + signedBlocksWindow
	maxMissed := signedBlocksWindow - minSignedPerWindow

	// if we are past the minimum height and the validator has missed too many blocks, punish them
	if height > minHeight && signInfo.MissedBlocksCounter > maxMissed {
		validator, err := k.sk.ValidatorByConsAddr(ctx, consAddr)
		if err != nil {
			return err
		}
		if validator != nil && !validator.IsJailed() {
			// Downtime confirmed: slash and jail the validator
			// We need to retrieve the stake distribution which signed the block, so we subtract ValidatorUpdateDelay from the evidence height,
			// and subtract an additional 1 since this is the LastCommit.
			// Note that this *can* result in a negative "distributionHeight" up to -ValidatorUpdateDelay-1,
			// i.e. at the end of the pre-genesis block (none) = at the beginning of the genesis block.
			// That's fine since this is just used to filter unbonding delegations & redelegations.
			distributionHeight := height - sdk.ValidatorUpdateDelay - 1

			slashFractionDowntime, err := k.SlashFractionDowntime(ctx)
			if err != nil {
				return err
			}

			coinsBurned, err := k.sk.SlashWithInfractionReason(ctx, consAddr, distributionHeight, power, slashFractionDowntime, st.Infraction_INFRACTION_DOWNTIME)
			if err != nil {
				return err
			}

			if err := k.environment.EventService.EventManager(ctx).EmitKV(
				types.EventTypeSlash,
				event.NewAttribute(types.AttributeKeyAddress, consStr),
				event.NewAttribute(types.AttributeKeyPower, fmt.Sprintf("%d", power)),
				event.NewAttribute(types.AttributeKeyReason, types.AttributeValueMissingSignature),
				event.NewAttribute(types.AttributeKeyJailed, consStr),
				event.NewAttribute(types.AttributeKeyBurnedCoins, coinsBurned.String()),
			); err != nil {
				return err
			}

			err = k.sk.Jail(ctx, consAddr)
			if err != nil {
				return err
			}
			downtimeJailDur, err := k.DowntimeJailDuration(ctx)
			if err != nil {
				return err
			}
			signInfo.JailedUntil = k.environment.HeaderService.GetHeaderInfo(ctx).Time.Add(downtimeJailDur)

			// We need to reset the counter & bitmap so that the validator won't be
			// immediately slashed for downtime upon re-bonding.
			signInfo.MissedBlocksCounter = 0
			signInfo.IndexOffset = 0
			err = k.DeleteMissedBlockBitmap(ctx, consAddr)
			if err != nil {
				return err
			}

			logger.Info(
				"slashing and jailing validator due to liveness fault",
				"height", height,
				"validator", consStr,
				"min_height", minHeight,
				"threshold", minSignedPerWindow,
				"slashed", slashFractionDowntime.String(),
				"jailed_until", signInfo.JailedUntil,
			)
		} else {
			// validator was (a) not found or (b) already jailed so we do not slash
			logger.Info(
				"validator would have been slashed for downtime, but was either not found in store or already jailed",
				"validator", consStr,
			)
		}
	}

	// Set the updated signing info
	return k.ValidatorSigningInfo.Set(ctx, consAddr, signInfo)
}
