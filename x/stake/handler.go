package stake

import (
	"bytes"

	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/abci/types"
)

func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		// NOTE msg already has validate basic run
		switch msg := msg.(type) {
		case MsgCreateValidator:
			return handleMsgCreateValidator(ctx, msg, k)
		case MsgEditValidator:
			return handleMsgEditValidator(ctx, msg, k)
		case MsgDelegate:
			return handleMsgDelegate(ctx, msg, k)
		case MsgUnbond:
			return handleMsgUnbond(ctx, msg, k)
		default:
			return sdk.ErrTxDecode("invalid message parse in staking module").Result()
		}
	}
}

// Called every block, process inflation, update validator set
func EndBlocker(ctx sdk.Context, k Keeper) (ValidatorUpdates []abci.Validator) {
	pool := k.GetPool(ctx)

	// Process Validator Provisions
	blockTime := ctx.BlockHeader().Time // XXX assuming in seconds, confirm
	if pool.InflationLastTime+blockTime >= 3600 {
		pool.InflationLastTime = blockTime
		pool = k.processProvisions(ctx)
	}

	// save the params
	k.setPool(ctx, pool)

	// reset the intra-transaction counter
	k.setIntraTxCounter(ctx, 0)

	// calculate validator set changes
	ValidatorUpdates = k.getTendermintUpdates(ctx)
	k.clearTendermintUpdates(ctx)
	return
}

//_____________________________________________________________________

// These functions assume everything has been authenticated,
// now we just perform action and save

func handleMsgCreateValidator(ctx sdk.Context, msg MsgCreateValidator, k Keeper) sdk.Result {

	// check to see if the pubkey or sender has been registered before
	_, found := k.GetValidator(ctx, msg.ValidatorAddr)
	if found {
		return ErrValidatorExistsAddr(k.codespace).Result()
	}
	if msg.Bond.Denom != k.GetParams(ctx).BondDenom {
		return ErrBadBondingDenom(k.codespace).Result()
	}
	if ctx.IsCheckTx() {
		return sdk.Result{}
	}

	validator := NewValidator(msg.ValidatorAddr, msg.PubKey, msg.Description)
	k.setValidator(ctx, validator)
	k.setValidatorByPubKeyIndex(ctx, validator)
	tags := sdk.NewTags(
		"action", []byte("createValidator"),
		"validator", msg.ValidatorAddr.Bytes(),
		"moniker", []byte(msg.Description.Moniker),
		"identity", []byte(msg.Description.Identity),
	)

	// move coins from the msg.Address account to a (self-bond) delegator account
	// the validator account and global shares are updated within here
	delegateTags, err := delegate(ctx, k, msg.ValidatorAddr, msg.Bond, validator)
	if err != nil {
		return err.Result()
	}
	tags = tags.AppendTags(delegateTags)
	return sdk.Result{
		Tags: tags,
	}
}

func handleMsgEditValidator(ctx sdk.Context, msg MsgEditValidator, k Keeper) sdk.Result {

	// validator must already be registered
	validator, found := k.GetValidator(ctx, msg.ValidatorAddr)
	if !found {
		return ErrBadValidatorAddr(k.codespace).Result()
	}
	if ctx.IsCheckTx() {
		return sdk.Result{}
	}

	// XXX move to types
	// replace all editable fields (clients should autofill existing values)
	validator.Description.Moniker = msg.Description.Moniker
	validator.Description.Identity = msg.Description.Identity
	validator.Description.Website = msg.Description.Website
	validator.Description.Details = msg.Description.Details

	k.updateValidator(ctx, validator)
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

func handleMsgDelegate(ctx sdk.Context, msg MsgDelegate, k Keeper) sdk.Result {

	validator, found := k.GetValidator(ctx, msg.ValidatorAddr)
	if !found {
		return ErrBadValidatorAddr(k.codespace).Result()
	}
	if msg.Bond.Denom != k.GetParams(ctx).BondDenom {
		return ErrBadBondingDenom(k.codespace).Result()
	}
	if validator.Revoked == true {
		return ErrValidatorRevoked(k.codespace).Result()
	}
	if ctx.IsCheckTx() {
		return sdk.Result{}
	}
	tags, err := delegate(ctx, k, msg.DelegatorAddr, msg.Bond, validator)
	if err != nil {
		return err.Result()
	}
	return sdk.Result{
		Tags: tags,
	}
}

// common functionality between handlers
func delegate(ctx sdk.Context, k Keeper, delegatorAddr sdk.Address,
	bondAmt sdk.Coin, validator Validator) (sdk.Tags, sdk.Error) {

	// Get or create the delegator bond
	bond, found := k.GetDelegation(ctx, delegatorAddr, validator.Owner)
	if !found {
		bond = Delegation{
			DelegatorAddr: delegatorAddr,
			ValidatorAddr: validator.Owner,
			Shares:        sdk.ZeroRat(),
		}
	}

	// Account new shares, save
	pool := k.GetPool(ctx)
	_, _, err := k.coinKeeper.SubtractCoins(ctx, bond.DelegatorAddr, sdk.Coins{bondAmt})
	if err != nil {
		return nil, err
	}
	validator, pool, newShares := validator.addTokensFromDel(pool, bondAmt.Amount)
	bond.Shares = bond.Shares.Add(newShares)

	// Update bond height
	bond.Height = ctx.BlockHeight()

	k.setPool(ctx, pool)
	k.setDelegation(ctx, bond)
	k.updateValidator(ctx, validator)
	tags := sdk.NewTags("action", []byte("delegate"), "delegator", delegatorAddr.Bytes(), "validator", validator.Owner.Bytes())
	return tags, nil
}

func handleMsgUnbond(ctx sdk.Context, msg MsgUnbond, k Keeper) sdk.Result {

	// check if bond has any shares in it unbond
	bond, found := k.GetDelegation(ctx, msg.DelegatorAddr, msg.ValidatorAddr)
	if !found {
		return ErrNoDelegatorForAddress(k.codespace).Result()
	}

	var delShares sdk.Rat

	// test that there are enough shares to unbond
	if msg.Shares == "MAX" {
		if !bond.Shares.GT(sdk.ZeroRat()) {
			return ErrNotEnoughBondShares(k.codespace, msg.Shares).Result()
		}
	} else {
		var err sdk.Error
		delShares, err = sdk.NewRatFromDecimal(msg.Shares)
		if err != nil {
			return err.Result()
		}
		if bond.Shares.LT(delShares) {
			return ErrNotEnoughBondShares(k.codespace, msg.Shares).Result()
		}
	}

	// get validator
	validator, found := k.GetValidator(ctx, msg.ValidatorAddr)
	if !found {
		return ErrNoValidatorForAddress(k.codespace).Result()
	}

	if ctx.IsCheckTx() {
		return sdk.Result{}
	}

	// retrieve the amount of bonds to remove (TODO remove redundancy already serialized)
	if msg.Shares == "MAX" {
		delShares = bond.Shares
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

		k.removeDelegation(ctx, bond)
	} else {
		// Update bond height
		bond.Height = ctx.BlockHeight()
		k.setDelegation(ctx, bond)
	}

	// Add the coins
	pool := k.GetPool(ctx)
	validator, pool, returnAmount := validator.removeDelShares(pool, delShares)
	k.setPool(ctx, pool)
	returnCoins := sdk.Coins{{k.GetParams(ctx).BondDenom, returnAmount}}
	k.coinKeeper.AddCoins(ctx, bond.DelegatorAddr, returnCoins)

	/////////////////////////////////////
	// revoke validator if necessary
	if revokeValidator {
		validator.Revoked = true
	}

	validator = k.updateValidator(ctx, validator)

	if validator.DelegatorShares.IsZero() {
		k.removeValidator(ctx, validator.Owner)
	}

	tags := sdk.NewTags("action", []byte("unbond"), "delegator", msg.DelegatorAddr.Bytes(), "validator", msg.ValidatorAddr.Bytes())
	return sdk.Result{
		Tags: tags,
	}
}
