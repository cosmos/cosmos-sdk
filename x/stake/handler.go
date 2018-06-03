package stake

import (
	"bytes"

	abci "github.com/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/stake/keeper"
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
		return ErrValidatorExistsAddr(k.Codespace()).Result()
	}
	if msg.Bond.Denom != k.GetParams(ctx).BondDenom {
		return ErrBadBondingDenom(k.Codespace()).Result()
	}

	validator := NewValidator(msg.ValidatorAddr, msg.PubKey, msg.Description)
	k.SetValidator(ctx, validator)
	k.SetValidatorByPubKeyIndex(ctx, validator)
	tags := sdk.NewTags(
		"action", []byte("createValidator"),
		"validator", msg.ValidatorAddr.Bytes(),
		"moniker", []byte(msg.Description.Moniker),
		"identity", []byte(msg.Description.Identity),
	)

	// move coins from the msg.Address account to a (self-bond) delegator account
	// the validator account and global shares are updated within here
	delegateTags, err := k.Delegate(ctx, msg.ValidatorAddr, msg.Bond, validator)
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
		return ErrBadValidatorAddr(k.Codespace()).Result()
	}

	// XXX move to types
	// replace all editable fields (clients should autofill existing values)
	validator.Description.Moniker = msg.Description.Moniker
	validator.Description.Identity = msg.Description.Identity
	validator.Description.Website = msg.Description.Website
	validator.Description.Details = msg.Description.Details

	k.UpdateValidator(ctx, validator)
	tags := sdk.NewTags(
		"action", []byte("editValidator"),
		"validator", msg.ValidatorAddr.Bytes(),
		"moniker", []byte(msg.Description.Moniker),
		"identity", []byte(msg.Description.Identity),
	)
	return sdk.Result{
		Tags: tags,
	}
}

func handleMsgDelegate(ctx sdk.Context, msg types.MsgDelegate, k keeper.PrivlegedKeeper) sdk.Result {

	validator, found := k.GetValidator(ctx, msg.ValidatorAddr)
	if !found {
		return ErrBadValidatorAddr(k.Codespace()).Result()
	}
	if msg.Bond.Denom != k.GetParams(ctx).BondDenom {
		return ErrBadBondingDenom(k.Codespace()).Result()
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

	// check if bond has any shares in it unbond
	bond, found := k.GetDelegation(ctx, msg.DelegatorAddr, msg.ValidatorAddr)
	if !found {
		return ErrNoDelegatorForAddress(k.Codespace()).Result()
	}

	var delShares sdk.Rat

	// test that there are enough shares to unbond
	if !msg.SharesPercent.Equal(sdk.ZeroRat()) {
		if !bond.Shares.GT(sdk.ZeroRat()) {
			return ErrNotEnoughBondShares(k.Codespace(), bond.Shares.String()).Result()
		}
	} else {
		if bond.Shares.LT(msg.SharesAmount) {
			return ErrNotEnoughBondShares(k.Codespace(), bond.Shares.String()).Result()
		}
	}

	// get validator
	validator, found := k.GetValidator(ctx, msg.ValidatorAddr)
	if !found {
		return ErrNoValidatorForAddress(k.Codespace()).Result()
	}

	// retrieve the amount of bonds to remove
	if !msg.SharesPercent.Equal(sdk.ZeroRat()) {
		delShares = bond.Shares.Mul(msg.SharesPercent)
	}

	// subtract bond tokens from delegator bond
	bond.Shares = bond.Shares.Sub(delShares)

	// remove the bond
	revokeValidator := false
	if bond.Shares.IsZero() {

		// if the bond is the owner of the validator then
		// trigger a revoke validator
		if bytes.Equal(bond.DelegatorAddr, validator.Owner) &&
			validator.Revoked == false {
			revokeValidator = true
		}

		k.RemoveDelegation(ctx, bond)
	} else {
		// Update bond height
		bond.Height = ctx.BlockHeight()
		k.SetDelegation(ctx, bond)
	}

	// Add the coins
	pool := k.GetPool(ctx)
	validator, pool, returnAmount := validator.RemoveDelShares(pool, delShares)
	k.SetPool(ctx, pool)
	returnCoins := sdk.Coins{{k.GetParams(ctx).BondDenom, returnAmount}}
	k.coinKeeper.AddCoins(ctx, bond.DelegatorAddr, returnCoins)

	/////////////////////////////////////
	// revoke validator if necessary
	if revokeValidator {
		validator.Revoked = true
	}

	validator = k.UpdateValidator(ctx, validator)

	if validator.DelegatorShares.IsZero() {
		k.RemoveValidator(ctx, validator.Owner)
	}

	tags := sdk.NewTags("action", []byte("unbond"), "delegator", msg.DelegatorAddr.Bytes(), "validator", msg.ValidatorAddr.Bytes())
	return sdk.Result{
		Tags: tags,
	}
}

func handleMsgCompleteUnbonding(ctx sdk.Context, msg types.MsgCompleteUnbonding, k keeper.PrivlegedKeeper) sdk.Result {
	// XXX
	return sdk.Result{}
}

func handleMsgBeginRedelegate(ctx sdk.Context, msg types.MsgBeginRedelegate, k keeper.PrivlegedKeeper) sdk.Result {
	// XXX
	return sdk.Result{}
}

func handleMsgCompleteRedelegate(ctx sdk.Context, msg types.MsgCompleteRedelegate, k keeper.PrivlegedKeeper) sdk.Result {
	// XXX
	return sdk.Result{}
}
