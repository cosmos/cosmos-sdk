package poa

// TODO make proposal for creating validator and increasing weight
import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/tendermint/tendermint/libs/common"
	tmtypes "github.com/tendermint/tendermint/types"
)

// NewParamChangeProposalHandler creates a new governance Handler for a ParamChangeProposal
func NewPOAProposalHandler(k Keeper) govtypes.Handler {
	return func(ctx sdk.Context, content govtypes.Content) sdk.Error {
		switch c := content.(type) {
		case MsgProposeCreateValidator:
			return handleMsgProposeCreateValidatorl(ctx, k, c)

		default:
			errMsg := fmt.Sprintf("unrecognized poa proposal content type: %T", c)
			return sdk.ErrUnknownRequest(errMsg)
		}
	}
}

func handleMsgProposeCreateValidatorl(ctx sdk.Context, k Keeper, c MsgProposeCreateValidator) sdk.Error {
	val := c.Validator
	// check to see if the pubkey or sender has been registered before
	if _, found := k.GetValidator(ctx, val.ValidatorAddress); found {
		return stakingtypes.ErrValidatorOwnerExists(k.Codespace())
	}

	if _, found := k.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(val.PubKey)); found {
		return stakingtypes.ErrValidatorPubKeyExists(k.Codespace())
	}

	if _, err := val.Description.EnsureLength(); err != nil {
		return err
	}

	if ctx.ConsensusParams() != nil {
		tmPubKey := tmtypes.TM2PB.PubKey(val.PubKey)
		if !common.StringInSlice(tmPubKey.Type, ctx.ConsensusParams().Validator.PubKeyTypes) {
			return stakingtypes.ErrValidatorPubKeyTypeNotSupported(k.Codespace(),
				tmPubKey.Type,
				ctx.ConsensusParams().Validator.PubKeyTypes)
		}
	}

	validator := NewValidator(val.ValidatorAddress, val.PubKey, val.Description)

	k.SetValidator(ctx, validator)
	k.SetValidatorByConsAddr(ctx, validator)
	k.SetNewValidatorByPowerIndex(ctx, validator)

	// call the after-creation hook
	k.AfterValidatorCreated(ctx, validator.OperatorAddress)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			EventTypeCreateValidator,
			sdk.NewAttribute(AttributeKeyValidator, val.ValidatorAddress.String()),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, AttributeValueCategory),
		),
	})

	// sdk.Result{Events: ctx.EventManager().Events()}
	return nil
}
