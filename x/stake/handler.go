package stake

import (
	"bytes"
	"fmt"
	"strconv"

	crypto "github.com/tendermint/go-crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
)

// separated for testing
func InitState(ctx sdk.Context, mapper Mapper, key, value string) sdk.Error {

	params := mapper.loadParams()
	switch key {
	case "allowed_bond_denom":
		params.BondDenom = value
	case "max_vals", "gas_bond", "gas_unbond":

		i, err := strconv.Atoi(value)
		if err != nil {
			return sdk.ErrUnknownRequest(fmt.Sprintf("input must be integer, Error: %v", err.Error()))
		}

		switch key {
		case "max_vals":
			if i < 0 {
				return sdk.ErrUnknownRequest("cannot designate negative max validators")
			}
			params.MaxVals = uint16(i)
		case "gas_bond":
			params.GasDelegate = int64(i)
		case "gas_unbound":
			params.GasUnbond = int64(i)
		}
	default:
		return sdk.ErrUnknownRequest(key)
	}

	mapper.saveParams(params)
	return nil
}

//_______________________________________________________________________

func NewHandler(mapper Mapper, ck bank.CoinKeeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {

		params := mapper.loadParams()

		err := msg.ValidateBasic()
		if err != nil {
			return err.Result() // TODO should also return gasUsed?
		}
		signers := msg.GetSigners()
		if len(signers) != 1 {
			return sdk.ErrUnauthorized("there can only be one signer for staking transaction").Result()
		}
		sender := signers[0]

		transact := newTransact(ctx, sender, mapper, ck)

		// Run the transaction
		switch msg := msg.(type) {
		case MsgDeclareCandidacy:
			res := transact.declareCandidacy(msg).Result()
			if !ctx.IsCheckTx() {
				res.GasUsed = params.GasDeclareCandidacy
			}
			return res
		case MsgEditCandidacy:
			res := transact.editCandidacy(msg).Result()
			if !ctx.IsCheckTx() {
				res.GasUsed = params.GasEditCandidacy
			}
			return res
		case MsgDelegate:
			res := transact.delegate(msg).Result()
			if !ctx.IsCheckTx() {
				res.GasUsed = params.GasDelegate
			}
			return res
		case MsgUnbond:
			res := transact.unbond(msg).Result()
			if !ctx.IsCheckTx() {
				res.GasUsed = params.GasUnbond
			}
			return res
		default:
			return sdk.ErrTxParse("invalid message parse in staking module").Result()
		}
	}
}

//_____________________________________________________________________

// common fields to all transactions
type transact struct {
	ctx        sdk.Context
	sender     crypto.Address
	mapper     Mapper
	coinKeeper bank.CoinKeeper
	params     Params
	gs         *GlobalState
}

func newTransact(ctx sdk.Context, sender sdk.Address, mapper Mapper, ck bank.CoinKeeper) transact {
	return transact{
		ctx:        ctx,
		sender:     sender,
		mapper:     mapper,
		coinKeeper: ck,
		params:     mapper.loadParams(),
		gs:         mapper.loadGlobalState(),
	}
}

//_____________________________________________________________________
// helper functions

// move a candidates asset pool from bonded to unbonded pool
func (tr transact) bondedToUnbondedPool(candidate *Candidate) {

	// replace bonded shares with unbonded shares
	tokens := tr.gs.removeSharesBonded(candidate.Assets)
	candidate.Assets = tr.gs.addTokensUnbonded(tokens)
	candidate.Status = Unbonded
}

// move a candidates asset pool from unbonded to bonded pool
func (tr transact) unbondedToBondedPool(candidate *Candidate) {

	// replace unbonded shares with bonded shares
	tokens := tr.gs.removeSharesUnbonded(candidate.Assets)
	candidate.Assets = tr.gs.addTokensBonded(tokens)
	candidate.Status = Bonded
}

// return an error if the bonds coins are incorrect
func checkDenom(mapper Mapper, bond sdk.Coin) sdk.Error {
	if bond.Denom != mapper.loadParams().BondDenom {
		return ErrBadBondingDenom()
	}
	return nil
}

//_____________________________________________________________________

// These functions assume everything has been authenticated,
// now we just perform action and save

func (tr transact) declareCandidacy(tx MsgDeclareCandidacy) sdk.Error {

	// check to see if the pubkey or sender has been registered before
	if tr.mapper.loadCandidate(tx.Address) != nil {
		return ErrCandidateExistsAddr()
	}
	err := checkDenom(tr.mapper, tx.Bond)
	if err != nil {
		return err
	}
	if tr.ctx.IsCheckTx() {
		return nil
	}

	candidate := NewCandidate(tx.PubKey, tr.sender, tx.Description)
	tr.mapper.saveCandidate(candidate)

	// move coins from the tr.sender account to a (self-bond) delegator account
	// the candidate account and global shares are updated within here
	txDelegate := NewMsgDelegate(tx.Address, tx.Bond)
	return tr.delegateWithCandidate(txDelegate, candidate)
}

func (tr transact) editCandidacy(tx MsgEditCandidacy) sdk.Error {

	// candidate must already be registered
	if tr.mapper.loadCandidate(tx.Address) == nil {
		return ErrBadCandidateAddr()
	}
	if tr.ctx.IsCheckTx() {
		return nil
	}

	// Get the pubKey bond account
	candidate := tr.mapper.loadCandidate(tx.Address)
	if candidate == nil {
		return ErrBondNotNominated()
	}
	if candidate.Status == Unbonded { //candidate has been withdrawn
		return ErrBondNotNominated()
	}

	//check and edit any of the editable terms
	if tx.Description.Moniker != "" {
		candidate.Description.Moniker = tx.Description.Moniker
	}
	if tx.Description.Identity != "" {
		candidate.Description.Identity = tx.Description.Identity
	}
	if tx.Description.Website != "" {
		candidate.Description.Website = tx.Description.Website
	}
	if tx.Description.Details != "" {
		candidate.Description.Details = tx.Description.Details
	}

	tr.mapper.saveCandidate(candidate)
	return nil
}

func (tr transact) delegate(tx MsgDelegate) sdk.Error {

	if tr.mapper.loadCandidate(tx.Address) == nil {
		return ErrBadCandidateAddr()
	}
	err := checkDenom(tr.mapper, tx.Bond)
	if err != nil {
		return err
	}
	if tr.ctx.IsCheckTx() {
		return nil
	}

	// Get the pubKey bond account
	candidate := tr.mapper.loadCandidate(tx.Address)
	if candidate == nil {
		return ErrBondNotNominated()
	}
	return tr.delegateWithCandidate(tx, candidate)
}

func (tr transact) delegateWithCandidate(tx MsgDelegate, candidate *Candidate) sdk.Error {

	if candidate.Status == Revoked { //candidate has been withdrawn
		return ErrBondNotNominated()
	}

	// Get or create the delegator bond
	bond := tr.mapper.loadDelegatorBond(tr.sender, tx.Address)
	if bond == nil {
		bond = &DelegatorBond{
			Address: tx.Address,
			Shares:  sdk.ZeroRat,
		}
	}

	// Account new shares, save
	err := tr.BondCoins(bond, candidate, tx.Bond)
	if err != nil {
		return err
	}
	tr.mapper.saveDelegatorBond(tr.sender, bond)
	tr.mapper.saveCandidate(candidate)
	tr.mapper.saveGlobalState(tr.gs)
	return nil
}

// Perform all the actions required to bond tokens to a delegator bond from their account
func (tr *transact) BondCoins(bond *DelegatorBond, candidate *Candidate, tokens sdk.Coin) sdk.Error {

	_, err := tr.coinKeeper.SubtractCoins(tr.ctx, candidate.Address, sdk.Coins{tokens})
	if err != nil {
		return err
	}
	newShares := candidate.addTokens(tokens.Amount, tr.gs)
	bond.Shares = bond.Shares.Add(newShares)
	return nil
}

// Perform all the actions required to bond tokens to a delegator bond from their account
func (tr *transact) UnbondCoins(bond *DelegatorBond, candidate *Candidate, shares sdk.Rat) sdk.Error {

	// subtract bond tokens from delegator bond
	if bond.Shares.LT(shares) {
		return sdk.ErrInsufficientFunds("") // TODO
	}
	bond.Shares = bond.Shares.Sub(shares)

	returnAmount := candidate.removeShares(shares, tr.gs)
	returnCoins := sdk.Coins{{tr.params.BondDenom, returnAmount}}

	_, err := tr.coinKeeper.AddCoins(tr.ctx, candidate.Address, returnCoins)
	if err != nil {
		return err
	}
	return nil
}

func (tr transact) unbond(tx MsgUnbond) sdk.Error {

	// check if bond has any shares in it unbond
	bond := tr.mapper.loadDelegatorBond(tr.sender, tx.Address)
	if bond == nil {
		return ErrNoDelegatorForAddress()
	}
	if !bond.Shares.GT(sdk.ZeroRat) { // bond shares < tx shares
		return ErrInsufficientFunds()
	}

	// if shares set to special case Max then we're good
	if tx.Shares != "MAX" {
		// test getting rational number from decimal provided
		shares, err := sdk.NewRatFromDecimal(tx.Shares)
		if err != nil {
			return err
		}

		// test that there are enough shares to unbond
		if !bond.Shares.GT(shares) {
			return ErrNotEnoughBondShares(tx.Shares)
		}
	}
	if tr.ctx.IsCheckTx() {
		return nil
	}

	// retrieve the amount of bonds to remove (TODO remove redundancy already serialized)
	var shares sdk.Rat
	var err sdk.Error
	if tx.Shares == "MAX" {
		shares = bond.Shares
	} else {
		shares, err = sdk.NewRatFromDecimal(tx.Shares)
		if err != nil {
			return err
		}
	}

	// subtract bond tokens from delegator bond
	if bond.Shares.LT(shares) { // bond shares < tx shares
		return ErrInsufficientFunds()
	}
	bond.Shares = bond.Shares.Sub(shares)

	// get pubKey candidate
	candidate := tr.mapper.loadCandidate(tx.Address)
	if candidate == nil {
		return ErrNoCandidateForAddress()
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
		tr.mapper.removeDelegatorBond(tr.sender, tx.Address)
	} else {
		tr.mapper.saveDelegatorBond(tr.sender, bond)
	}

	// Add the coins
	returnAmount := candidate.removeShares(shares, tr.gs)
	returnCoins := sdk.Coins{{tr.params.BondDenom, returnAmount}}
	tr.coinKeeper.AddCoins(tr.ctx, tr.sender, returnCoins)

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
		tr.mapper.removeCandidate(tx.Address)
	} else {
		tr.mapper.saveCandidate(candidate)
	}

	tr.mapper.saveGlobalState(tr.gs)
	return nil
}
