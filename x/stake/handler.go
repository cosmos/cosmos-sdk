package stake

import (
	"encoding/json"

	abci "github.com/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/stake/keeper"
	"github.com/cosmos/cosmos-sdk/x/stake/tags"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
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

	// Process types.Validator Provisions
	blockTime := ctx.BlockHeader().Time
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

func handleMsgCreateValidator(ctx sdk.Context, msg types.MsgCreateValidator, k keeper.Keeper) sdk.Result {

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

	// move coins from the msg.Address account to a (self-delegation) delegator account
	// the validator account and global shares are updated within here
	_, delegation, validator, pool, err := k.Delegate(ctx, msg.ValidatorAddr, msg.SelfDelegation, validator)
	k.SetPool(ctx, pool)
	k.SetDelegation(ctx, delegation)
	k.UpdateValidator(ctx, validator)
	if err != nil {
		return err.Result()
	}

	tags := sdk.NewTags(
		tags.Action, tags.ActionCreateValidator,
		tags.DstValidator, msg.ValidatorAddr.Bytes(),
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
		tags.DstValidator, msg.ValidatorAddr.Bytes(),
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
	if msg.Bond.Denom != k.GetParams(ctx).BondDenom {
		return ErrBadDenom(k.Codespace()).Result()
	}
	if validator.Revoked == true {
		return ErrValidatorRevoked(k.Codespace()).Result()
	}
	_, delegation, validator, pool, err := k.Delegate(ctx, msg.DelegatorAddr, msg.Bond, validator)
	if err != nil {
		return err.Result()
	}
	k.SetPool(ctx, pool)
	k.SetDelegation(ctx, delegation)
	k.UpdateValidator(ctx, validator)
	tags := sdk.NewTags(
		tags.Action, tags.ActionDelegate,
		tags.Delegator, msg.DelegatorAddr.Bytes(),
		tags.DstValidator, msg.ValidatorAddr.Bytes(),
	)
	return sdk.Result{
		Tags: tags,
	}
}

func handleMsgBeginUnbonding(ctx sdk.Context, msg types.MsgBeginUnbonding, k keeper.Keeper) sdk.Result {

	delegation, validator, pool, returnAmount, err := k.UnbondDelegation(ctx, msg.DelegatorAddr, msg.ValidatorAddr, msg.SharesAmount)
	if err != nil {
		return err.Result()
	}

	k.SetPool(ctx, pool)

	// create the unbonding delegation
	params := k.GetParams(ctx)
	minTime := ctx.BlockHeader().Time + params.UnbondingTime

	ubd := UnbondingDelegation{
		DelegatorAddr: delegation.DelegatorAddr,
		ValidatorAddr: delegation.ValidatorAddr,
		MinTime:       minTime,
		Balance:       sdk.Coin{params.BondDenom, sdk.NewInt(returnAmount)},
		Slashed:       sdk.Coin{},
	}
	k.SetUnbondingDelegation(ctx, ubd)

	// update then remove validator if necessary
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

func handleMsgCompleteUnbonding(ctx sdk.Context, msg types.MsgCompleteUnbonding, k keeper.Keeper) sdk.Result {

	ubd, found := k.GetUnbondingDelegation(ctx, msg.DelegatorAddr, msg.ValidatorAddr)
	if !found {
		return ErrNoUnbondingDelegation(k.Codespace()).Result()
	}

	// ensure that enough time has passed
	ctxTime := ctx.BlockHeader().Time
	if ubd.MinTime > ctxTime {
		return ErrNotMature(k.Codespace(), "unbonding", "unit-time", ubd.MinTime, ctxTime).Result()
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

func handleMsgBeginRedelegate(ctx sdk.Context, msg types.MsgBeginRedelegate, k keeper.Keeper) sdk.Result {

	delegation, srcValidator, pool, returnAmount, err := k.UnbondDelegation(ctx, msg.DelegatorAddr, msg.ValidatorSrcAddr, msg.SharesAmount)
	if err != nil {
		return err.Result()
	}

	// update then remove the source validator if necessary
	srcValidator = k.UpdateValidator(ctx, srcValidator)
	if srcValidator.DelegatorShares.IsZero() {
		k.RemoveValidator(ctx, srcValidator.Owner)
	}

	params := k.GetParams(ctx)
	returnCoin := sdk.Coin{params.BondDenom, sdk.NewInt(returnAmount)}
	dstValidator, found := k.GetValidator(ctx, msg.ValidatorDstAddr)
	if !found {
		return ErrBadRedelegationDst(k.Codespace()).Result()
	}
	sharesCreated, delegation, dstValidator, pool, err := k.Delegate(ctx, msg.DelegatorAddr, returnCoin, dstValidator)
	k.SetPool(ctx, pool)
	k.SetDelegation(ctx, delegation)
	k.UpdateValidator(ctx, dstValidator)

	// create the unbonding delegation
	minTime := ctx.BlockHeader().Time + params.UnbondingTime

	red := Redelegation{
		DelegatorAddr:    msg.DelegatorAddr,
		ValidatorSrcAddr: msg.ValidatorSrcAddr,
		ValidatorDstAddr: msg.ValidatorDstAddr,
		MinTime:          minTime,
		SharesDst:        sharesCreated,
		SharesSrc:        msg.SharesAmount,
	}
	k.SetRedelegation(ctx, red)

	tags := sdk.NewTags(
		tags.Action, tags.ActionBeginRedelegation,
		tags.Delegator, msg.DelegatorAddr.Bytes(),
		tags.SrcValidator, msg.ValidatorSrcAddr.Bytes(),
		tags.DstValidator, msg.ValidatorDstAddr.Bytes(),
	)
	return sdk.Result{Tags: tags}
}

func handleMsgCompleteRedelegate(ctx sdk.Context, msg types.MsgCompleteRedelegate, k keeper.Keeper) sdk.Result {

	red, found := k.GetRedelegation(ctx, msg.DelegatorAddr, msg.ValidatorSrcAddr, msg.ValidatorDstAddr)
	if !found {
		return ErrNoRedelegation(k.Codespace()).Result()
	}

	// ensure that enough time has passed
	ctxTime := ctx.BlockHeader().Time
	if red.MinTime > ctxTime {
		return ErrNotMature(k.Codespace(), "redelegation", "unit-time", red.MinTime, ctxTime).Result()
	}

	k.RemoveRedelegation(ctx, red)

	tags := sdk.NewTags(
		tags.Action, tags.ActionCompleteRedelegation,
		tags.Delegator, msg.DelegatorAddr.Bytes(),
		tags.SrcValidator, msg.ValidatorSrcAddr.Bytes(),
		tags.DstValidator, msg.ValidatorDstAddr.Bytes(),
	)
	return sdk.Result{Tags: tags}
}
