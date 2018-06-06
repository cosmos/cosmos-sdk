package stake

import (
	"bytes"
	"encoding/json"

	abci "github.com/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/stake/keeper"
	"github.com/cosmos/cosmos-sdk/x/stake/tags"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
)

func NewHandler(k keeper.PrivlegedKeeper) sdk.Handler {
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
func EndBlocker(ctx sdk.Context, k keeper.PrivlegedKeeper) (ValidatorUpdates []abci.Validator) {
	pool := k.GetPool(ctx)

	// Process types.Validator Provisions
	blockTime := ctx.BlockHeader().Time // XXX assuming in seconds, confirm
	if pool.InflationLastTime+blockTime >= 3600 {
		pool.InflationLastTime = blockTime
		pool = k.ProcessProvisions(ctx)
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

func handleMsgCreateValidator(ctx sdk.Context, msg types.MsgCreateValidator, k keeper.PrivlegedKeeper) sdk.Result {

	// check to see if the pubkey or sender has been registered before
	_, found := k.GetValidator(ctx, msg.ValidatorAddr)
	if found {
		return ErrValidatorAlreadyExists(k.Codespace()).Result()
	}
	if msg.SelfDelegation.Denom != k.GetParams(ctx).BondDenom {
		return ErrBadDenom(k.Codespace()).Result()
	}

	validator := NewValidator(msg.ValidatorAddr, msg.PubKey, msg.Description)
	k.SetValidator(ctx, validator)
	k.SetValidatorByPubKeyIndex(ctx, validator)
	tags := sdk.NewTags(
		tags.Action, tags.ActionCreateValidator,
		tags.DstValidator, msg.ValidatorAddr.Bytes(),
		tags.Moniker, []byte(msg.Description.Moniker),
		tags.Identity, []byte(msg.Description.Identity),
	)

	// move coins from the msg.Address account to a (self-delegation) delegator account
	// the validator account and global shares are updated within here
	delegateTags, err := k.Delegate(ctx, msg.ValidatorAddr, msg.SelfDelegation, validator)
	if err != nil {
		return err.Result()
	}
	tags = tags.AppendTags(delegateTags)
	return sdk.Result{
		Tags: tags,
	}
}

func handleMsgEditValidator(ctx sdk.Context, msg types.MsgEditValidator, k keeper.PrivlegedKeeper) sdk.Result {

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
		tags.DstValidator, msg.ValidatorAddr.Bytes(),
		tags.Moniker, []byte(description.Moniker),
		tags.Identity, []byte(description.Identity),
	)
	return sdk.Result{
		Tags: tags,
	}
}

func handleMsgDelegate(ctx sdk.Context, msg types.MsgDelegate, k keeper.PrivlegedKeeper) sdk.Result {

	validator, found := k.GetValidator(ctx, msg.ValidatorAddr)
	if !found {
		return ErrNoValidatorFound(k.Codespace()).Result()
	}
	if msg.Bond.Denom != k.GetParams(ctx).BondDenom {
		return ErrBadDenom(k.Codespace()).Result()
	}
	if validator.Revoked == true {
		return ErrValidatorRevoked(k.Codespace()).Result()
	}
	tags, err := k.Delegate(ctx, msg.DelegatorAddr, msg.Bond, validator)
	if err != nil {
		return err.Result()
	}
	return sdk.Result{
		Tags: tags,
	}
}

func handleMsgBeginUnbonding(ctx sdk.Context, msg types.MsgBeginUnbonding, k keeper.PrivlegedKeeper) sdk.Result {

	// check if delegation has any shares in it unbond
	delegation, found := k.GetDelegation(ctx, msg.DelegatorAddr, msg.ValidatorAddr)
	if !found {
		return ErrNoDelegatorForAddress(k.Codespace()).Result()
	}

	var delShares sdk.Rat

	// retrieve the amount to remove
	if !msg.SharesPercent.IsZero() {
		delShares = delegation.Shares.Mul(msg.SharesPercent)
		if !delegation.Shares.GT(sdk.ZeroRat()) {
			return ErrNotEnoughDelegationShares(k.Codespace(), delegation.Shares.String()).Result()
		}
	} else {
		delShares = msg.SharesAmount
		if delegation.Shares.LT(msg.SharesAmount) {
			return ErrNotEnoughDelegationShares(k.Codespace(), delegation.Shares.String()).Result()
		}
	}

	// get validator
	validator, found := k.GetValidator(ctx, msg.ValidatorAddr)
	if !found {
		return ErrNoValidatorFound(k.Codespace()).Result()
	}

	// subtract shares from delegator
	delegation.Shares = delegation.Shares.Sub(delShares)

	// remove the delegation
	if delegation.Shares.IsZero() {

		// if the delegation is the owner of the validator then
		// trigger a revoke validator
		if bytes.Equal(delegation.DelegatorAddr, validator.Owner) && validator.Revoked == false {
			validator.Revoked = true
		}
		k.RemoveDelegation(ctx, delegation)
	} else {
		// Update height
		delegation.Height = ctx.BlockHeight()
		k.SetDelegation(ctx, delegation)
	}

	// remove the coins from the validator
	pool := k.GetPool(ctx)
	validator, pool, returnAmount := validator.RemoveDelShares(pool, delShares)
	k.SetPool(ctx, pool)

	// create the unbonding delegation
	params := k.GetParams(ctx)
	minTime := ctx.BlockHeader().Time + params.UnbondingTime
	minHeight := ctx.BlockHeight() + params.MinUnbondingBlocks

	ubd := UnbondingDelegation{
		DelegatorAddr: delegation.DelegatorAddr,
		ValidatorAddr: delegation.ValidatorAddr,
		MinTime:       minTime,
		MinHeight:     minHeight,
		Balance:       sdk.Coin{params.BondDenom, returnAmount},
		Slashed:       sdk.Coin{},
	}
	k.SetUnbondingDelegation(ctx, ubd)

	/////////////////////////////////////
	// revoke validator if necessary

	validator = k.UpdateValidator(ctx, validator)
	if validator.DelegatorShares.IsZero() {
		k.RemoveValidator(ctx, validator.Owner)
	}

	tags := sdk.NewTags(
		tags.Action, tags.ActionBeginUnbonding,
		tags.Delegator, msg.DelegatorAddr.Bytes(),
		tags.SrcValidator, msg.ValidatorAddr.Bytes(),
	)
	return sdk.Result{Tags: tags}
}

func handleMsgCompleteUnbonding(ctx sdk.Context, msg types.MsgCompleteUnbonding, k keeper.PrivlegedKeeper) sdk.Result {

	ubd, found := k.GetUnbondingDelegation(ctx, msg.DelegatorAddr, msg.ValidatorAddr)
	if !found {
		return ErrNoRedelegation(k.Codespace()).Result()
	}

	// ensure that enough time has passed
	ctxTime := ctx.BlockHeader().Time
	ctxHeight := ctx.BlockHeight()
	if ubd.MinTime > ctxTime {
		return ErrNotMature(k.Codespace(), "unbonding", "unit-time", ubd.MinTime, ctxTime).Result()
	}
	if ubd.MinHeight > ctxHeight {
		return ErrNotMature(k.Codespace(), "unbonding", "block-height", ubd.MinHeight, ctxHeight).Result()
	}

	k.CoinKeeper().AddCoins(ctx, ubd.DelegatorAddr, sdk.Coins{ubd.Balance})
	k.RemoveUnbondingDelegation(ctx, ubd)

	tags := sdk.NewTags(
		TagAction, ActionCompleteUnbonding,
		TagDelegator, msg.DelegatorAddr.Bytes(),
		TagSrcValidator, msg.ValidatorAddr.Bytes(),
	)

	// add slashed tag only if there has been some slashing
	if !ubd.Slashed.IsZero() {
		bz, err := json.Marshal(ubd.Slashed)
		if err != nil {
			panic(err)
		}
		tags = tags.AppendTag(string(TagSlashed), bz)
	}
	return sdk.Result{Tags: tags}
}

func handleMsgBeginRedelegate(ctx sdk.Context, msg types.MsgBeginRedelegate, k keeper.PrivlegedKeeper) sdk.Result {
	// XXX
	return sdk.Result{}
}

func handleMsgCompleteRedelegate(ctx sdk.Context, msg types.MsgCompleteRedelegate, k keeper.PrivlegedKeeper) sdk.Result {
	// XXX
	return sdk.Result{}
}
