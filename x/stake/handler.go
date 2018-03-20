package stake

import (
	"bytes"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
)

//nolint
const (
	GasDeclareCandidacy int64 = 20
	GasEditCandidacy    int64 = 20
	GasDelegate         int64 = 20
	GasUnbond           int64 = 20
)

//XXX fix initstater
// separated for testing
//func InitState(ctx sdk.Context, k Keeper, key, value string) sdk.Error {

//params := k.getParams(ctx)
//switch key {
//case "allowed_bond_denom":
//params.BondDenom = value
//case "max_vals", "gas_bond", "gas_unbond":

//i, err := strconv.Atoi(value)
//if err != nil {
//return sdk.ErrUnknownRequest(fmt.Sprintf("input must be integer, Error: %v", err.Error()))
//}

//switch key {
//case "max_vals":
//if i < 0 {
//return sdk.ErrUnknownRequest("cannot designate negative max validators")
//}
//params.MaxValidators = uint16(i)
//case "gas_bond":
//GasDelegate = int64(i)
//case "gas_unbound":
//GasUnbond = int64(i)
//}
//default:
//return sdk.ErrUnknownRequest(key)
//}

//k.setParams(params)
//return nil
//}

//_______________________________________________________________________

func NewHandler(k Keeper, ck bank.CoinKeeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		// NOTE msg already has validate basic run
		switch msg := msg.(type) {
		case MsgDeclareCandidacy:
			return k.handleMsgDeclareCandidacy(ctx, msg)
		case MsgEditCandidacy:
			return k.handleMsgEditCandidacy(ctx, msg)
		case MsgDelegate:
			return k.handleMsgDelegate(ctx, msg)
		case MsgUnbond:
			return k.handleMsgUnbond(ctx, msg)
		default:
			return sdk.ErrTxParse("invalid message parse in staking module").Result()
		}
	}
}

//_____________________________________________________________________

// XXX should be send in the msg (init in CLI)
//func getSender() sdk.Address {
//signers := msg.GetSigners()
//if len(signers) != 1 {
//return sdk.ErrUnauthorized("there can only be one signer for staking transaction").Result()
//}
//sender := signers[0]
//}

//_____________________________________________________________________
// helper functions

// move a candidates asset pool from bonded to unbonded pool
func (k Keeper) bondedToUnbondedPool(ctx sdk.Context, candidate *Candidate) {

	// replace bonded shares with unbonded shares
	tokens := k.getGlobalState(ctx).removeSharesBonded(candidate.Assets)
	candidate.Assets = k.getGlobalState(ctx).addTokensUnbonded(tokens)
	candidate.Status = Unbonded
}

// move a candidates asset pool from unbonded to bonded pool
func (k Keeper) unbondedToBondedPool(ctx sdk.Context, candidate *Candidate) {

	// replace unbonded shares with bonded shares
	tokens := k.getGlobalState(ctx).removeSharesUnbonded(candidate.Assets)
	candidate.Assets = k.getGlobalState(ctx).addTokensBonded(tokens)
	candidate.Status = Bonded
}

//_____________________________________________________________________

// These functions assume everything has been authenticated,
// now we just perform action and save

func (k Keeper) handleMsgDeclareCandidacy(ctx sdk.Context, msg MsgDeclareCandidacy) sdk.Result {

	// check to see if the pubkey or sender has been registered before
	if k.getCandidate(msg.Address) != nil {
		return ErrCandidateExistsAddr()
	}
	if msg.bond.Denom != k.getParams().BondDenom {
		return ErrBadBondingDenom()
	}
	if ctx.IsCheckTx() {
		return sdk.Result{
			GasUsed: GasDeclareCandidacy,
		}
	}

	candidate := NewCandidate(msg.PubKey, msg.Address, msg.Description)
	k.setCandidate(candidate)

	// move coins from the msg.Address account to a (self-bond) delegator account
	// the candidate account and global shares are updated within here
	txDelegate := NewMsgDelegate(msg.Address, msg.Bond)
	return delegateWithCandidate(txDelegate, candidate)
}

func (k Keeper) handleMsgEditCandidacy(ctx sdk.Context, msg MsgEditCandidacy) sdk.Result {

	// candidate must already be registered
	if k.getCandidate(msg.Address) == nil {
		return ErrBadCandidateAddr().Result()
	}
	if ctx.IsCheckTx() {
		return sdk.Result{
			GasUsed: GasEditCandidacy,
		}
	}

	// Get the pubKey bond account
	candidate := k.getCandidate(msg.Address)
	if candidate == nil {
		return ErrBondNotNominated().Result()
	}
	if candidate.Status == Unbonded { //candidate has been withdrawn
		return ErrBondNotNominated().Result()
	}

	//check and edit any of the editable terms
	if msg.Description.Moniker != "" {
		candidate.Description.Moniker = msg.Description.Moniker
	}
	if msg.Description.Identity != "" {
		candidate.Description.Identity = msg.Description.Identity
	}
	if msg.Description.Website != "" {
		candidate.Description.Website = msg.Description.Website
	}
	if msg.Description.Details != "" {
		candidate.Description.Details = msg.Description.Details
	}

	k.setCandidate(candidate)
	return nil
}

func (k Keeper) handleMsgDelegate(ctx sdk.Context, msg MsgDelegate) sdk.Result {

	if k.getCandidate(msg.Address) == nil {
		return ErrBadCandidateAddr().Result()
	}
	if msg.bond.Denom != k.getParams().BondDenom {
		return ErrBadBondingDenom().Result()
	}
	if ctx.IsCheckTx() {
		return sdk.Result{
			GasUsed: GasDelegate,
		}
	}

	// Get the pubKey bond account
	candidate := k.getCandidate(msg.Address)
	if candidate == nil {
		return ErrBondNotNominated().Result()
	}

	return tr.delegateWithCandidate(msg, candidate).Result()
}

func (k Keeper) delegateWithCandidate(ctx sdk.Context, candidateAddr, delegatorAddr sdk.Address,
	bondAmt sdk.Coin, candidate Candidate) sdk.Error {

	if candidate.Status == Revoked { //candidate has been withdrawn
		return ErrBondNotNominated()
	}

	// Get or create the delegator bond
	bond := k.getDelegatorBond(tr.sender, canad)
	if bond == nil {
		bond = &DelegatorBond{
			CandidateAddr: delegatorAddr,
			DelegatorAddr: candidateAddr,
			Shares:        sdk.ZeroRat,
		}
	}

	// Account new shares, save
	err := BondCoins(bond, candidate, msg.Bond)
	if err != nil {
		return err.Result()
	}
	k.setDelegatorBond(tr.sender, bond)
	k.setCandidate(candidate)
	k.setGlobalState(tr.gs)
	return nil
}

// Perform all the actions required to bond tokens to a delegator bond from their account
func (k Keeper) BondCoins(ctx sdk.Context, bond DelegatorBond, amount sdk.Coin) sdk.Error {

	_, err := k.coinKeeper.SubtractCoins(ctx, bond.DelegatorAddr, sdk.Coins{amount})
	if err != nil {
		return err
	}
	newShares := candidate.addTokens(tokens.Amount, tr.gs)
	bond.Shares = bond.Shares.Add(newShares)
	k.SetDelegatorBond()
	return nil
}

// Perform all the actions required to bond tokens to a delegator bond from their account
func (k Keeper) UnbondCoins(ctx sdk.Context, bond *DelegatorBond, candidate *Candidate, shares sdk.Rat) sdk.Error {

	// subtract bond tokens from delegator bond
	if bond.Shares.LT(shares) {
		return sdk.ErrInsufficientFunds("") //XXX variables inside
	}
	bond.Shares = bond.Shares.Sub(shares)

	returnAmount := candidate.removeShares(shares, tr.gs)
	returnCoins := sdk.Coins{{tr.params.BondDenom, returnAmount}}

	_, err := tr.coinKeeper.AddCoins(ctx, candidate.Address, returnCoins)
	if err != nil {
		return err
	}
	return nil
}

func (k Keeper) handleMsgUnbond(ctx sdk.Context, msg MsgUnbond) sdk.Result {

	// check if bond has any shares in it unbond
	bond := k.getDelegatorBond(sender, msg.Address)
	if bond == nil {
		return ErrNoDelegatorForAddress().Result()
	}
	if !bond.Shares.GT(sdk.ZeroRat) { // bond shares < msg shares
		return ErrInsufficientFunds().Result()
	}

	// if shares set to special case Max then we're good
	if msg.Shares != "MAX" {
		// test getting rational number from decimal provided
		shares, err := sdk.NewRatFromDecimal(msg.Shares)
		if err != nil {
			return err.Result()
		}

		// test that there are enough shares to unbond
		if !bond.Shares.GT(shares) {
			return ErrNotEnoughBondShares(msg.Shares).Result()
		}
	}
	if ctx.IsCheckTx() {
		return sdk.Result{
			GasUsed: GasUnbond,
		}
	}

	// retrieve the amount of bonds to remove (TODO remove redundancy already serialized)
	var shares sdk.Rat
	var err sdk.Error
	if msg.Shares == "MAX" {
		shares = bond.Shares
	} else {
		shares, err = sdk.NewRatFromDecimal(msg.Shares)
		if err != nil {
			return err.Result()
		}
	}

	// subtract bond tokens from delegator bond
	if bond.Shares.LT(shares) { // bond shares < msg shares
		return ErrInsufficientFunds().Result()
	}
	bond.Shares = bond.Shares.Sub(shares)

	// get pubKey candidate
	candidate := k.getCandidate(msg.Address)
	if candidate == nil {
		return ErrNoCandidateForAddress().Result()
	}

	revokeCandidacy := false
	if bond.Shares.IsZero() {

		// if the bond is the owner of the candidate then
		// trigger a revoke candidacy
		if bytes.Equal(tr.sender, candidate.Address) &&
			candidate.Status != Revoked {
			revokeCandidacy = true
		}

		// remove the bond
		k.removeDelegatorBond(ctx, msg.Address)
	} else {
		k.setDelegatorBond(tr.sender, bond)
	}

	// Add the coins
	returnAmount := candidate.removeShares(shares, tr.gs)
	returnCoins := sdk.Coins{{tr.params.BondDenom, returnAmount}}
	tr.coinKeeper.AddCoins(ctx, tr.sender, returnCoins)

	// lastly if an revoke candidate if necessary
	if revokeCandidacy {

		// change the share types to unbonded if they were not already
		if candidate.Status == Bonded {
			tr.bondedToUnbondedPool(candidate)
		}

		// lastly update the status
		candidate.Status = Revoked
	}

	// deduct shares from the candidate and save
	if candidate.Liabilities.IsZero() {
		k.removeCandidate(msg.Address)
	} else {
		k.setCandidate(candidate)
	}

	k.setGlobalState(tr.gs)
	return sdk.Result{}
}
