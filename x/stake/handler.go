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
	for _, candidate := range data.Candidates {
		k.setCandidate(ctx, candidate)
	}
	for _, bond := range data.Bonds {
		k.setDelegatorBond(ctx, bond)
	}
}

// WriteGenesis - output genesis parameters
func WriteGenesis(ctx sdk.Context, k Keeper) GenesisState {
	pool := k.GetPool(ctx)
	params := k.GetParams(ctx)
	candidates := k.GetCandidates(ctx, 32767)
	bonds := k.getBonds(ctx, 32767)
	return GenesisState{
		pool,
		params,
		candidates,
		bonds,
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
	tags := sdk.NewTags("action", []byte("declareCandidacy"), "candidate", msg.CandidateAddr.Bytes(), "moniker", []byte(msg.Description.Moniker), "identity", []byte(msg.Description.Identity))

	// move coins from the msg.Address account to a (self-bond) delegator account
	// the candidate account and global shares are updated within here
	delegateTags, err := delegate(ctx, k, msg.CandidateAddr, msg.Bond, candidate)
	if err != nil {
		return err.Result()
	}
	tags = tags.AppendTags(delegateTags)
	return sdk.Result{
		Tags: tags,
	}
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

	// XXX move to types
	// replace all editable fields (clients should autofill existing values)
	candidate.Description.Moniker = msg.Description.Moniker
	candidate.Description.Identity = msg.Description.Identity
	candidate.Description.Website = msg.Description.Website
	candidate.Description.Details = msg.Description.Details

	k.setCandidate(ctx, candidate)
	tags := sdk.NewTags("action", []byte("editCandidacy"), "candidate", msg.CandidateAddr.Bytes(), "moniker", []byte(msg.Description.Moniker), "identity", []byte(msg.Description.Identity))
	return sdk.Result{
		Tags: tags,
	}
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
	tags, err := delegate(ctx, k, msg.DelegatorAddr, msg.Bond, candidate)
	if err != nil {
		return err.Result()
	}
	return sdk.Result{
		Tags: tags,
	}
}

// common functionality between handlers
func delegate(ctx sdk.Context, k Keeper, delegatorAddr sdk.Address,
	bondAmt sdk.Coin, candidate Candidate) (sdk.Tags, sdk.Error) {

	// Get or create the delegator bond
	bond, found := k.GetDelegatorBond(ctx, delegatorAddr, candidate.Address)
	if !found {
		bond = DelegatorBond{
			DelegatorAddr: delegatorAddr,
			CandidateAddr: candidate.Address,
			Shares:        sdk.ZeroRat(),
		}
	}

	// Account new shares, save
	pool := k.GetPool(ctx)
	_, _, err := k.coinKeeper.SubtractCoins(ctx, bond.DelegatorAddr, sdk.Coins{bondAmt})
	if err != nil {
		return nil, err
	}
	pool, candidate, newShares := pool.candidateAddTokens(candidate, bondAmt.Amount)
	bond.Shares = bond.Shares.Add(newShares)

	// Update bond height
	bond.Height = ctx.BlockHeight()

	k.setDelegatorBond(ctx, bond)
	k.setCandidate(ctx, candidate)
	k.setPool(ctx, pool)
	tags := sdk.NewTags("action", []byte("delegate"), "delegator", delegatorAddr.Bytes(), "candidate", candidate.Address.Bytes())
	return tags, nil
}

func handleMsgUnbond(ctx sdk.Context, msg MsgUnbond, k Keeper) sdk.Result {

	// check if bond has any shares in it unbond
	bond, found := k.GetDelegatorBond(ctx, msg.DelegatorAddr, msg.CandidateAddr)
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
		// Update bond height
		bond.Height = ctx.BlockHeight()
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
	tags := sdk.NewTags("action", []byte("unbond"), "delegator", msg.DelegatorAddr.Bytes(), "candidate", msg.CandidateAddr.Bytes())
	return sdk.Result{
		Tags: tags,
	}
}
