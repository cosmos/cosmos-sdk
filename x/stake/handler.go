package stake

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/stake/keeper"
	"github.com/cosmos/cosmos-sdk/x/stake/tags"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

func NewHandler(k keeper.Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		// NOTE msg already has validate basic run
		switch msg := msg.(type) {
		case types.MsgCreateValidator:
			return handleMsgCreateValidator(ctx, msg, k)
		case types.MsgEditValidator:
			return handleMsgEditValidator(ctx, msg, k)
		case types.MsgDelegate:
			return handleMsgDelegate(ctx, msg, k)
		case types.MsgBeginRedelegate:
			return handleMsgBeginRedelegate(ctx, msg, k)
		case types.MsgCompleteRedelegate:
			return handleMsgCompleteRedelegate(ctx, msg, k)
		case types.MsgBeginUnbonding:
			return handleMsgBeginUnbonding(ctx, msg, k)
		case types.MsgCompleteUnbonding:
			return handleMsgCompleteUnbonding(ctx, msg, k)
		default:
			return sdk.ErrTxDecode("invalid message parse in staking module").Result()
		}
	}
}

// Called every block, process inflation, update validator set
func EndBlocker(ctx sdk.Context, k keeper.Keeper) (ValidatorUpdates []abci.Validator) {
	pool := k.GetPool(ctx)
	params := k.GetParams(ctx)

	// Process types.Validator Provisions
	blockTime := ctx.BlockHeader().Time
	if blockTime-pool.InflationLastTime >= 3600 {
		pool.InflationLastTime = blockTime
		pool = pool.ProcessProvisions(params)
	}

	// save the params
	k.SetPool(ctx, pool)

	// reset the intra-transaction counter
	k.SetIntraTxCounter(ctx, 0)

	// calculate validator set changes
	ValidatorUpdates = k.GetTendermintUpdates(ctx)
	k.ClearTendermintUpdates(ctx)
	return
}

//_____________________________________________________________________

// These functions assume everything has been authenticated,
// now we just perform action and save

func handleMsgCreateValidator(ctx sdk.Context, msg types.MsgCreateValidator, k keeper.Keeper) sdk.Result {

	// check to see if the pubkey or sender has been registered before
	_, found := k.GetValidator(ctx, msg.ValidatorAddr)
	if found {
		return ErrValidatorOwnerExists(k.Codespace()).Result()
	}
	_, found = k.GetValidatorByPubKey(ctx, msg.PubKey)
	if found {
		return ErrValidatorPubKeyExists(k.Codespace()).Result()
	}
	if msg.Delegation.Denom != k.GetParams(ctx).BondDenom {
		return ErrBadDenom(k.Codespace()).Result()
	}

	validator := NewValidator(msg.ValidatorAddr, msg.PubKey, msg.Description)
	k.SetValidator(ctx, validator)
	k.SetValidatorByPubKeyIndex(ctx, validator)

	// move coins from the msg.Address account to a (self-delegation) delegator account
	// the validator account and global shares are updated within here
	_, err := k.Delegate(ctx, msg.DelegatorAddr, msg.Delegation, validator, true)
	if err != nil {
		return err.Result()
	}

	tags := sdk.NewTags(
		tags.Action, tags.ActionCreateValidator,
		tags.DstValidator, []byte(msg.ValidatorAddr.String()),
		tags.Moniker, []byte(msg.Description.Moniker),
		tags.Identity, []byte(msg.Description.Identity),
	)
	return sdk.Result{
		Tags: tags,
	}
}

func handleMsgEditValidator(ctx sdk.Context, msg types.MsgEditValidator, k keeper.Keeper) sdk.Result {

	// validator must already be registered
	validator, found := k.GetValidator(ctx, msg.ValidatorAddr)
	if !found {
		return ErrNoValidatorFound(k.Codespace()).Result()
	}

	// replace all editable fields (clients should autofill existing values)
	description, err := validator.Description.UpdateDescription(msg.Description)
	if err != nil {
		return err.Result()
	}
	validator.Description = description

	k.UpdateValidator(ctx, validator)
	tags := sdk.NewTags(
		tags.Action, tags.ActionEditValidator,
		tags.DstValidator, []byte(msg.ValidatorAddr.String()),
		tags.Moniker, []byte(description.Moniker),
		tags.Identity, []byte(description.Identity),
	)
	return sdk.Result{
		Tags: tags,
	}
}

func handleMsgDelegate(ctx sdk.Context, msg types.MsgDelegate, k keeper.Keeper) sdk.Result {

	validator, found := k.GetValidator(ctx, msg.ValidatorAddr)
	if !found {
		return ErrNoValidatorFound(k.Codespace()).Result()
	}
	if msg.Delegation.Denom != k.GetParams(ctx).BondDenom {
		return ErrBadDenom(k.Codespace()).Result()
	}
	if validator.Revoked == true {
		return ErrValidatorRevoked(k.Codespace()).Result()
	}
	_, err := k.Delegate(ctx, msg.DelegatorAddr, msg.Delegation, validator, true)
	if err != nil {
		return err.Result()
	}

	tags := sdk.NewTags(
		tags.Action, tags.ActionDelegate,
		tags.Delegator, []byte(msg.DelegatorAddr.String()),
		tags.DstValidator, []byte(msg.ValidatorAddr.String()),
	)
	return sdk.Result{
		Tags: tags,
	}
}

func handleMsgBeginUnbonding(ctx sdk.Context, msg types.MsgBeginUnbonding, k keeper.Keeper) sdk.Result {
	err := k.BeginUnbonding(ctx, msg.DelegatorAddr, msg.ValidatorAddr, msg.SharesAmount)
	if err != nil {
		return err.Result()
	}

	tags := sdk.NewTags(
		tags.Action, tags.ActionBeginUnbonding,
		tags.Delegator, []byte(msg.DelegatorAddr.String()),
		tags.SrcValidator, []byte(msg.ValidatorAddr.String()),
	)
	return sdk.Result{Tags: tags}
}

func handleMsgCompleteUnbonding(ctx sdk.Context, msg types.MsgCompleteUnbonding, k keeper.Keeper) sdk.Result {

	err := k.CompleteUnbonding(ctx, msg.DelegatorAddr, msg.ValidatorAddr)
	if err != nil {
		return err.Result()
	}

	tags := sdk.NewTags(
		tags.Action, ActionCompleteUnbonding,
		tags.Delegator, []byte(msg.DelegatorAddr.String()),
		tags.SrcValidator, []byte(msg.ValidatorAddr.String()),
	)

	return sdk.Result{Tags: tags}
}

func handleMsgBeginRedelegate(ctx sdk.Context, msg types.MsgBeginRedelegate, k keeper.Keeper) sdk.Result {
	err := k.BeginRedelegation(ctx, msg.DelegatorAddr, msg.ValidatorSrcAddr,
		msg.ValidatorDstAddr, msg.SharesAmount)
	if err != nil {
		return err.Result()
	}

	tags := sdk.NewTags(
		tags.Action, tags.ActionBeginRedelegation,
		tags.Delegator, []byte(msg.DelegatorAddr.String()),
		tags.SrcValidator, []byte(msg.ValidatorSrcAddr.String()),
		tags.DstValidator, []byte(msg.ValidatorDstAddr.String()),
	)
	return sdk.Result{Tags: tags}
}

func handleMsgCompleteRedelegate(ctx sdk.Context, msg types.MsgCompleteRedelegate, k keeper.Keeper) sdk.Result {
	err := k.CompleteRedelegation(ctx, msg.DelegatorAddr, msg.ValidatorSrcAddr, msg.ValidatorDstAddr)
	if err != nil {
		return err.Result()
	}

	tags := sdk.NewTags(
		tags.Action, tags.ActionCompleteRedelegation,
		tags.Delegator, []byte(msg.DelegatorAddr.String()),
		tags.SrcValidator, []byte(msg.ValidatorSrcAddr.String()),
		tags.DstValidator, []byte(msg.ValidatorDstAddr.String()),
	)
	return sdk.Result{Tags: tags}
}
