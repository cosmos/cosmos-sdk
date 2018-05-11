package stake

import (
	"bytes"

	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/abci/types"
)

//nolint
const (
	GasDeclareCandidacy int64 = 20
	GasEditCandidacy    int64 = 20
	GasDelegate         int64 = 20
	GasUnbond           int64 = 20
)

//_______________________________________________________________________

func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		// NOTE msg already has validate basic run
		switch msg := msg.(type) {
		case MsgDeclareCandidacy:
			return handleMsgDeclareCandidacy(ctx, msg, k)
		case MsgEditCandidacy:
			return handleMsgEditCandidacy(ctx, msg, k)
		case MsgDelegate:
			return handleMsgDelegate(ctx, msg, k)
		case MsgUnbond:
			return handleMsgUnbond(ctx, msg, k)
		default:
			return sdk.ErrTxDecode("invalid message parse in staking module").Result()
		}
	}
}

//_____________________________________________________________________

// NewEndBlocker generates sdk.EndBlocker
// Performs tick functionality
func NewEndBlocker(k Keeper) sdk.EndBlocker {
	return func(ctx sdk.Context, req abci.RequestEndBlock) (res abci.ResponseEndBlock) {
		res.ValidatorUpdates = k.Tick(ctx)
		return
	}
}

//_____________________________________________________________________

// InitGenesis - store genesis parameters
func InitGenesis(ctx sdk.Context, k Keeper, data GenesisState) {
	k.setPool(ctx, data.Pool)
	k.setParams(ctx, data.Params)
	for _, validator := range data.Validators {
		k.setValidator(ctx, validator)
	}
	for _, bond := range data.Bonds {
		k.setDelegation(ctx, bond)
	}
}

// WriteGenesis - output genesis parameters
func WriteGenesis(ctx sdk.Context, k Keeper) GenesisState {
	pool := k.GetPool(ctx)
	params := k.GetParams(ctx)
	validators := k.GetValidators(ctx, 32767)
	bonds := k.getBonds(ctx, 32767)
	return GenesisState{
		pool,
		params,
		validators,
		bonds,
	}
}

//_____________________________________________________________________

// These functions assume everything has been authenticated,
// now we just perform action and save

func handleMsgDeclareCandidacy(ctx sdk.Context, msg MsgDeclareCandidacy, k Keeper) sdk.Result {

	// check to see if the pubkey or sender has been registered before
	_, found := k.GetValidator(ctx, msg.ValidatorAddr)
	if found {
		return ErrValidatorExistsAddr(k.codespace).Result()
	}
	if msg.Bond.Denom != k.GetParams(ctx).BondDenom {
		return ErrBadBondingDenom(k.codespace).Result()
	}
	if ctx.IsCheckTx() {
		return sdk.Result{
			GasUsed: GasDeclareCandidacy,
		}
	}

	validator := NewValidator(msg.ValidatorAddr, msg.PubKey, msg.Description)
	k.setValidator(ctx, validator)
	tags := sdk.NewTags(
		"action", []byte("declareCandidacy"),
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

func handleMsgEditCandidacy(ctx sdk.Context, msg MsgEditCandidacy, k Keeper) sdk.Result {

	// validator must already be registered
	validator, found := k.GetValidator(ctx, msg.ValidatorAddr)
	if !found {
		return ErrBadValidatorAddr(k.codespace).Result()
	}
	if ctx.IsCheckTx() {
		return sdk.Result{
			GasUsed: GasEditCandidacy,
		}
	}

	// XXX move to types
	// replace all editable fields (clients should autofill existing values)
	validator.Description.Moniker = msg.Description.Moniker
	validator.Description.Identity = msg.Description.Identity
	validator.Description.Website = msg.Description.Website
	validator.Description.Details = msg.Description.Details

	k.setValidator(ctx, validator)
	tags := sdk.NewTags(
		"action", []byte("editCandidacy"),
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
	if validator.Status == sdk.Revoked {
		return ErrValidatorRevoked(k.codespace).Result()
	}
	if ctx.IsCheckTx() {
		return sdk.Result{
			GasUsed: GasDelegate,
		}
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
	bond, found := k.GetDelegation(ctx, delegatorAddr, validator.Address)
	if !found {
		bond = Delegation{
			DelegatorAddr: delegatorAddr,
			ValidatorAddr: validator.Address,
			Shares:        sdk.ZeroRat(),
		}
	}

	// Account new shares, save
	pool := k.GetPool(ctx)
	_, _, err := k.coinKeeper.SubtractCoins(ctx, bond.DelegatorAddr, sdk.Coins{bondAmt})
	if err != nil {
		return nil, err
	}
	pool, validator, newShares := pool.validatorAddTokens(validator, bondAmt.Amount)
	bond.Shares = bond.Shares.Add(newShares)

	// Update bond height
	bond.Height = ctx.BlockHeight()

	k.setDelegation(ctx, bond)
	k.setValidator(ctx, validator)
	k.setPool(ctx, pool)
	tags := sdk.NewTags("action", []byte("delegate"), "delegator", delegatorAddr.Bytes(), "validator", validator.Address.Bytes())
	return tags, nil
}

func handleMsgUnbond(ctx sdk.Context, msg MsgUnbond, k Keeper) sdk.Result {

	// check if bond has any shares in it unbond
	bond, found := k.GetDelegation(ctx, msg.DelegatorAddr, msg.ValidatorAddr)
	if !found {
		return ErrNoDelegatorForAddress(k.codespace).Result()
	}
	if !bond.Shares.GT(sdk.ZeroRat()) { // bond shares < msg shares
		return ErrInsufficientFunds(k.codespace).Result()
	}

	var shares sdk.Rat

	// test that there are enough shares to unbond
	if msg.Shares == "MAX" {
		if !bond.Shares.GT(sdk.ZeroRat()) {
			return ErrNotEnoughBondShares(k.codespace, msg.Shares).Result()
		}
	} else {
		var err sdk.Error
		shares, err = sdk.NewRatFromDecimal(msg.Shares)
		if err != nil {
			return err.Result()
		}
		if bond.Shares.LT(shares) {
			return ErrNotEnoughBondShares(k.codespace, msg.Shares).Result()
		}
	}

	// get validator
	validator, found := k.GetValidator(ctx, msg.ValidatorAddr)
	if !found {
		return ErrNoValidatorForAddress(k.codespace).Result()
	}

	if ctx.IsCheckTx() {
		return sdk.Result{
			GasUsed: GasUnbond,
		}
	}

	// retrieve the amount of bonds to remove (TODO remove redundancy already serialized)
	if msg.Shares == "MAX" {
		shares = bond.Shares
	}

	// subtract bond tokens from delegator bond
	bond.Shares = bond.Shares.Sub(shares)

	// remove the bond
	revokeCandidacy := false
	if bond.Shares.IsZero() {

		// if the bond is the owner of the validator then
		// trigger a revoke candidacy
		if bytes.Equal(bond.DelegatorAddr, validator.Address) &&
			validator.Status != sdk.Revoked {
			revokeCandidacy = true
		}

		k.removeDelegation(ctx, bond)
	} else {
		// Update bond height
		bond.Height = ctx.BlockHeight()
		k.setDelegation(ctx, bond)
	}

	// Add the coins
	p := k.GetPool(ctx)
	p, validator, returnAmount := p.validatorRemoveShares(validator, shares)
	returnCoins := sdk.Coins{{k.GetParams(ctx).BondDenom, returnAmount}}
	k.coinKeeper.AddCoins(ctx, bond.DelegatorAddr, returnCoins)

	/////////////////////////////////////

	// revoke validator if necessary
	if revokeCandidacy {

		// change the share types to unbonded if they were not already
		if validator.Status == sdk.Bonded {
			validator.Status = sdk.Unbonded
			p, validator = p.UpdateSharesLocation(validator)
		}

		// lastly update the status
		validator.Status = sdk.Revoked
	}

	// deduct shares from the validator
	if validator.DelegatorShares.IsZero() {
		k.removeValidator(ctx, validator.Address)
	} else {
		k.setValidator(ctx, validator)
	}
	k.setPool(ctx, p)
	tags := sdk.NewTags("action", []byte("unbond"), "delegator", msg.DelegatorAddr.Bytes(), "validator", msg.ValidatorAddr.Bytes())
	return sdk.Result{
		Tags: tags,
	}
}
