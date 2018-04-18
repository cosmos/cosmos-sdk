package stake

import (
	"bytes"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
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

func NewHandler(k Keeper, ck bank.CoinKeeper) sdk.Handler {
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

//_______________________________________________

// NewEndBlocker generates sdk.EndBlocker
// Performs tick functionality
func NewEndBlocker(k Keeper) sdk.EndBlocker {
	return func(ctx sdk.Context, req abci.RequestEndBlock) (res abci.ResponseEndBlock) {
		res.ValidatorUpdates = k.Tick(ctx)
		return
	}
}

//_____________________________________________________________________

// These functions assume everything has been authenticated,
// now we just perform action and save

func handleMsgDeclareCandidacy(ctx sdk.Context, msg MsgDeclareCandidacy, k Keeper) sdk.Result {

	// check to see if the pubkey or sender has been registered before
	_, found := k.GetCandidate(ctx, msg.CandidateAddr)
	if found {
		return ErrCandidateExistsAddr(k.codespace).Result()
	}
	if msg.Bond.Denom != k.GetParams(ctx).BondDenom {
		return ErrBadBondingDenom(k.codespace).Result()
	}
	if ctx.IsCheckTx() {
		return sdk.Result{
			GasUsed: GasDeclareCandidacy,
		}
	}

	candidate := NewCandidate(msg.CandidateAddr, msg.PubKey, msg.Description)
	k.setCandidate(ctx, candidate)

	// move coins from the msg.Address account to a (self-bond) delegator account
	// the candidate account and global shares are updated within here
	err := delegate(ctx, k, msg.CandidateAddr, msg.Bond, candidate)
	if err != nil {
		return err.Result()
	}
	return sdk.Result{}
}

func handleMsgEditCandidacy(ctx sdk.Context, msg MsgEditCandidacy, k Keeper) sdk.Result {

	// candidate must already be registered
	candidate, found := k.GetCandidate(ctx, msg.CandidateAddr)
	if !found {
		return ErrBadCandidateAddr(k.codespace).Result()
	}
	if ctx.IsCheckTx() {
		return sdk.Result{
			GasUsed: GasEditCandidacy,
		}
	}
	if candidate.Status == Unbonded { //candidate has been withdrawn
		return ErrBondNotNominated(k.codespace).Result()
	}

	// XXX move to types
	// replace all editable fields (clients should autofill existing values)
	candidate.Description.Moniker = msg.Description.Moniker
	candidate.Description.Identity = msg.Description.Identity
	candidate.Description.Website = msg.Description.Website
	candidate.Description.Details = msg.Description.Details

	k.setCandidate(ctx, candidate)
	return sdk.Result{}
}

func handleMsgDelegate(ctx sdk.Context, msg MsgDelegate, k Keeper) sdk.Result {

	candidate, found := k.GetCandidate(ctx, msg.CandidateAddr)
	if !found {
		return ErrBadCandidateAddr(k.codespace).Result()
	}
	if msg.Bond.Denom != k.GetParams(ctx).BondDenom {
		return ErrBadBondingDenom(k.codespace).Result()
	}
	if candidate.Status == Revoked {
		return ErrCandidateRevoked(k.codespace).Result()
	}
	if ctx.IsCheckTx() {
		return sdk.Result{
			GasUsed: GasDelegate,
		}
	}
	err := delegate(ctx, k, msg.DelegatorAddr, msg.Bond, candidate)
	if err != nil {
		return err.Result()
	}
	return sdk.Result{}
}

// common functionality between handlers
func delegate(ctx sdk.Context, k Keeper, delegatorAddr sdk.Address,
	bondAmt sdk.Coin, candidate Candidate) sdk.Error {

	// Get or create the delegator bond
	bond, found := k.getDelegatorBond(ctx, delegatorAddr, candidate.Address)
	if !found {
		bond = DelegatorBond{
			DelegatorAddr: delegatorAddr,
			CandidateAddr: candidate.Address,
			Shares:        sdk.ZeroRat,
		}
	}

	// Account new shares, save
	pool := k.GetPool(ctx)
	_, err := k.coinKeeper.SubtractCoins(ctx, bond.DelegatorAddr, sdk.Coins{bondAmt})
	if err != nil {
		return err
	}
	pool, candidate, newShares := pool.candidateAddTokens(candidate, bondAmt.Amount)
	bond.Shares = bond.Shares.Add(newShares)

	k.setDelegatorBond(ctx, bond)
	k.setCandidate(ctx, candidate)
	k.setPool(ctx, pool)
	return nil
}

func handleMsgUnbond(ctx sdk.Context, msg MsgUnbond, k Keeper) sdk.Result {

	// check if bond has any shares in it unbond
	bond, found := k.getDelegatorBond(ctx, msg.DelegatorAddr, msg.CandidateAddr)
	if !found {
		return ErrNoDelegatorForAddress(k.codespace).Result()
	}
	if !bond.Shares.GT(sdk.ZeroRat) { // bond shares < msg shares
		return ErrInsufficientFunds(k.codespace).Result()
	}

	// test getting rational number from decimal provided
	shares, err := sdk.NewRatFromDecimal(msg.Shares)
	if err != nil {
		return err.Result()
	}

	// test that there are enough shares to unbond
	if msg.Shares == "MAX" {
		if !bond.Shares.GT(sdk.ZeroRat) {
			return ErrNotEnoughBondShares(k.codespace, msg.Shares).Result()
		}
	} else {
		if bond.Shares.LT(shares) {
			return ErrNotEnoughBondShares(k.codespace, msg.Shares).Result()
		}
	}

	// get candidate
	candidate, found := k.GetCandidate(ctx, msg.CandidateAddr)
	if !found {
		return ErrNoCandidateForAddress(k.codespace).Result()
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

		// if the bond is the owner of the candidate then
		// trigger a revoke candidacy
		if bytes.Equal(bond.DelegatorAddr, candidate.Address) &&
			candidate.Status != Revoked {
			revokeCandidacy = true
		}

		k.removeDelegatorBond(ctx, bond)
	} else {
		k.setDelegatorBond(ctx, bond)
	}

	// Add the coins
	p := k.GetPool(ctx)
	p, candidate, returnAmount := p.candidateRemoveShares(candidate, shares)
	returnCoins := sdk.Coins{{k.GetParams(ctx).BondDenom, returnAmount}}
	k.coinKeeper.AddCoins(ctx, bond.DelegatorAddr, returnCoins)

	/////////////////////////////////////

	// revoke candidate if necessary
	if revokeCandidacy {

		// change the share types to unbonded if they were not already
		if candidate.Status == Bonded {
			p, candidate = p.bondedToUnbondedPool(candidate)
		}

		// lastly update the status
		candidate.Status = Revoked
	}

	// deduct shares from the candidate
	if candidate.Liabilities.IsZero() {
		k.removeCandidate(ctx, candidate.Address)
	} else {
		k.setCandidate(ctx, candidate)
	}
	k.setPool(ctx, p)
	return sdk.Result{}
}

// TODO use or remove
//// Perform all the actions required to bond tokens to a delegator bond from their account
//func BondCoins(ctx sdk.Context, k Keeper, bond DelegatorBond,
//candidate Candidate, amount sdk.Coin) (DelegatorBond, Candidate, Pool, sdk.Error) {

//pool := k.GetPool(ctx)
//_, err := k.coinKeeper.SubtractCoins(ctx, bond.DelegatorAddr, sdk.Coins{amount})
//if err != nil {
//return bond, candidate, pool, err
//}
//pool, candidate, newShares := pool.candidateAddTokens(candidate, amount.Amount)
//bond.Shares = bond.Shares.Add(newShares)
//return bond, candidate, pool, nil
//}
//// Perform all the actions required to bond tokens to a delegator bond from their account
//func UnbondCoins(ctx sdk.Context, k Keeper, bond DelegatorBond,
//candidate Candidate, shares sdk.Rat) (DelegatorBond, Candidate, Pool, sdk.Error) {

//pool := k.GetPool(ctx)

//// subtract bond tokens from delegator bond
//if bond.Shares.LT(shares) {
//errMsg := fmt.Sprintf("cannot unbond %v shares, only have %v shares available", shares, bond.Shares)
//return bond, candidate, pool, sdk.ErrInsufficientFunds(errMsg)
//}
//bond.Shares = bond.Shares.Sub(shares)

//pool, candidate, returnAmount := p.candidateRemoveShares(candidate, shares)
//returnCoins := sdk.Coins{{k.GetParams(ctx).BondDenom, returnAmount}}

//_, err := k.coinKeeper.AddCoins(ctx, candidate.Address, returnCoins)
//if err != nil {
//return err
//}
//return bond, candidate, pool, nil
//}
