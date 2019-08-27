package poa

import (
	"fmt"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/common"
	tmtypes "github.com/tendermint/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		switch msg := msg.(type) {
		case MsgCreateValidator:
			return handleMsgCreateValidator(ctx, msg, k)

		case MsgEditValidator:
			return handleMsgEditValidator(ctx, msg, k)

		default:
			errMsg := fmt.Sprintf("unrecognized staking message type: %T", msg)
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

// Called every block, update validator set
func EndBlocker(ctx sdk.Context, k Keeper) []abci.ValidatorUpdate {
	// Calculate validator set changes.
	//
	// NOTE: ApplyAndReturnValidatorSetUpdates has to come before
	// UnbondAllMatureValidatorQueue.
	// This fixes a bug when the unbonding period is instant (is the case in
	// some of the tests). The test expected the validator to be completely
	// unbonded after the Endblocker (go from Bonded -> Unbonding during
	// ApplyAndReturnValidatorSetUpdates and then Unbonding -> Unbonded during
	// UnbondAllMatureValidatorQueue).
	validatorUpdates := k.ApplyAndReturnValidatorSetUpdates(ctx)

	// Unbond all mature validators from the unbonding queue.
	k.UnbondAllMatureValidatorQueue(ctx)

	return validatorUpdates
}

// These functions assume everything has been authenticated,
// now we just perform action and save

func handleMsgCreateValidator(ctx sdk.Context, msg MsgCreateValidator, k Keeper) sdk.Result {
	// check to see if the pubkey or sender has been registered before
	if _, found := k.GetValidator(ctx, msg.ValidatorAddress); found {
		return stakingtypes.ErrValidatorOwnerExists(k.Codespace()).Result()
	}

	if _, found := k.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(msg.PubKey)); found {
		return stakingtypes.ErrValidatorPubKeyExists(k.Codespace()).Result()
	}

	if _, err := msg.Description.EnsureLength(); err != nil {
		return err.Result()
	}

	if ctx.ConsensusParams() != nil {
		tmPubKey := tmtypes.TM2PB.PubKey(msg.PubKey)
		if !common.StringInSlice(tmPubKey.Type, ctx.ConsensusParams().Validator.PubKeyTypes) {
			return stakingtypes.ErrValidatorPubKeyTypeNotSupported(k.Codespace(),
				tmPubKey.Type,
				ctx.ConsensusParams().Validator.PubKeyTypes).Result()
		}
	}

	validator := NewValidator(msg.ValidatorAddress, msg.PubKey, msg.Description)

	k.SetValidator(ctx, validator)
	k.SetValidatorByConsAddr(ctx, validator)
	k.SetNewValidatorByPowerIndex(ctx, validator)

	// call the after-creation hook
	k.AfterValidatorCreated(ctx, validator.OperatorAddress)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			EventTypeCreateValidator,
			sdk.NewAttribute(AttributeKeyValidator, msg.ValidatorAddress.String()),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, AttributeValueCategory),
		),
	})

	return sdk.Result{Events: ctx.EventManager().Events()}
}

func handleMsgEditValidator(ctx sdk.Context, msg MsgEditValidator, k Keeper) sdk.Result {
	// validator must already be registered
	validator, found := k.GetValidator(ctx, msg.ValidatorAddress)
	if !found {
		return stakingtypes.ErrNoValidatorFound(k.Codespace()).Result()
	}

	// replace all editable fields (clients should autofill existing values)
	description, err := validator.Description.UpdateDescription(msg.Description)
	if err != nil {
		return err.Result()
	}

	validator.Description = description

	k.SetValidator(ctx, validator)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.ValidatorAddress.String()),
		),
	})

	return sdk.Result{Events: ctx.EventManager().Events()}
}
