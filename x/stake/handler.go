package stake

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/spf13/viper"
	crypto "github.com/tendermint/go-crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
)

// separated for testing
func InitState(ctx sdk.Context, key, value string) error {

	params := loadParams(store)
	switch key {
	case "allowed_bond_denom":
		params.AllowedBondDenom = value
	case "max_vals", "gas_bond", "gas_unbond":

		i, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("input must be integer, Error: %v", err.Error())
		}

		switch key {
		case "max_vals":
			if i < 0 {
				return errors.New("cannot designate negative max validators")
			}
			params.MaxVals = uint16(i)
		case "gas_bond":
			params.GasDelegate = int64(i)
		case "gas_unbound":
			params.GasUnbond = int64(i)
		}
	default:
		return sdk.ErrUnknownKey(key)
	}

	saveParams(store, params)
	return nil
}

//_______________________________________________________________________

func NewHandler(stakeKey sdk.StoreKey, ck bank.CoinKeeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {

		res := sdk.Result{}

		err := msg.ValidateBasic()
		if err != nil {
			return res, err
		}

		// return the fee for each tx type
		if ctx.IsCheckTx() {
			// XXX: add some tags so we can search it!
			switch txInner := tx.Unwrap().(type) {
			case TxDeclareCandidacy:
				return sdk.NewCheck(params.GasDeclareCandidacy, "")
			case TxEditCandidacy:
				return sdk.NewCheck(params.GasEditCandidacy, "")
			case TxDelegate:
				return sdk.NewCheck(params.GasDelegate, "")
			case TxUnbond:
				return sdk.NewCheck(params.GasUnbond, "")
			default:
				return sdk.ErrUnknownTxType(tx)
			}
		}

		// TODO: remove redundancy
		// also we don't need to check the res - gas is already deducted in sdk
		_, err = h.CheckTx(ctx, store, tx, nil)
		if err != nil {
			return
		}

		sender, err := getTxSender(ctx)
		if err != nil {
			return
		}

		keeper := NewKeeper(ctx, stakeKey)
		transact := NewTransact(ctx, ck)

		// Run the transaction
		switch _tx := tx.Unwrap().(type) {
		case TxDeclareCandidacy:
			res.GasUsed = params.GasDeclareCandidacy
			return res, transact.declareCandidacy(_tx)
		case TxEditCandidacy:
			res.GasUsed = params.GasEditCandidacy
			return res, transact.editCandidacy(_tx)
		case TxDelegate:
			res.GasUsed = params.GasDelegate
			return res, transact.delegate(_tx)
		case TxUnbond:
			//context with hold account permissions
			params := loadParams(store)
			res.GasUsed = params.GasUnbond
			//ctx2 := ctx.WithPermissions(params.HoldBonded) //TODO remove this line if non-permissioned ctx works
			return res, transact.unbond(_tx)
		}
		return
	}
}

// get the sender from the ctx and ensure it matches the tx pubkey
func getTxSender(ctx sdk.Context) (sender crypto.Address, err error) {
	senders := ctx.GetPermissions("", auth.NameSigs)
	if len(senders) != 1 {
		return sender, ErrMissingSignature()
	}
	return senders[0], nil
}

//_____________________________________________________________________

// common fields to all transactions
type transact struct {
	sender     crypto.Address
	keeper     Keeper
	coinKeeper bank.CoinKeeper
	params     Params
	gs         *GlobalState
}

//  XXX move keeper creation to application?
func newTransact(ctx sdk.Context, keeper Keeper, ck bank.CoinKeeper) transact {
	return transact{
		sender:     sender,
		keeper:     keeper,
		coinKeeper: ck,
		params:     keeper.loadParams(),
		gs:         keeper.loadGlobalState(),
	}
}

//_____________________________________________________________________
// helper functions
// TODO move from deliver with new SDK should only be dependant on store to send coins in NEW SDK

// move a candidates asset pool from bonded to unbonded pool
func (tr transact) bondedToUnbondedPool(candidate *Candidate) error {

	// replace bonded shares with unbonded shares
	tokens := tr.gs.removeSharesBonded(candidate.Assets)
	candidate.Assets = tr.gs.addTokensUnbonded(tokens)
	candidate.Status = Unbonded

	return tr.transfer(tr.params.HoldBonded, tr.params.HoldUnbonded,
		sdk.Coins{{tr.params.AllowedBondDenom, tokens}})
}

// move a candidates asset pool from unbonded to bonded pool
func (tr transact) unbondedToBondedPool(candidate *Candidate) error {

	// replace bonded shares with unbonded shares
	tokens := tr.gs.removeSharesUnbonded(candidate.Assets)
	candidate.Assets = tr.gs.addTokensBonded(tokens)
	candidate.Status = Bonded

	return tr.transfer(tr.params.HoldUnbonded, tr.params.HoldBonded,
		sdk.Coins{{tr.params.AllowedBondDenom, tokens}})
}

// return an error if the bonds coins are incorrect
func checkDenom(keeper Keeper, tx BondUpdate) error {
	if tx.Bond.Denom != keeper.loadParams().AllowedBondDenom {
		return fmt.Errorf("Invalid coin denomination")
	}
	return nil
}

//_____________________________________________________________________

// These functions assume everything has been authenticated,
// now we just perform action and save
func (tr transact) declareCandidacy(tx TxDeclareCandidacy) error {

	// check to see if the pubkey or sender has been registered before
	candidate := tr.keeper.loadCandidate(tx.PubKey)
	if candidate != nil {
		return fmt.Errorf("cannot bond to pubkey which is already declared candidacy"+
			" PubKey %v already registered with %v candidate address",
			candidate.PubKey, candidate.Owner)
	}
	err := checkDenom(tx.BondUpdate, tr.keeper)
	if err != nil {
		return err
	}
	// XXX end of old check tx

	// create and save the empty candidate
	bond := tr.keeper.loadCandidate(tx.PubKey)
	if bond != nil {
		return ErrCandidateExistsAddr()
	}
	candidate := NewCandidate(tx.PubKey, tr.sender, tx.Description)
	tr.keeper.saveCandidate(candidate)

	// move coins from the tr.sender account to a (self-bond) delegator account
	// the candidate account and global shares are updated within here
	txDelegate := TxDelegate{tx.BondUpdate}
	return tr.delegateWithCandidate(txDelegate, candidate)
}

func (tr transact) editCandidacy(tx TxEditCandidacy) error {

	// candidate must already be registered
	candidate := tr.keeper.loadCandidate(tx.PubKey)
	if candidate == nil { // does PubKey exist
		return fmt.Errorf("cannot delegate to non-existant PubKey %v", tx.PubKey)
	}
	// XXX end of old check tx

	// Get the pubKey bond account
	candidate := tr.keeper.loadCandidate(tx.PubKey)
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

	tr.keeper.saveCandidate(candidate)
	return nil
}

func (tr transact) delegate(tx TxDelegate) error {

	candidate := tr.keeper.loadCandidate(tx.PubKey)
	if candidate == nil { // does PubKey exist
		return fmt.Errorf("cannot delegate to non-existant PubKey %v", tx.PubKey)
	}
	err := checkDenom(tx.BondUpdate, tr.keeper)
	if err != nil {
		return err
	}
	// end of old check tx

	// Get the pubKey bond account
	candidate := tr.keeper.loadCandidate(tx.PubKey)
	if candidate == nil {
		return ErrBondNotNominated()
	}
	return tr.delegateWithCandidate(tx, candidate)
}

func (tr transact) delegateWithCandidate(tx TxDelegate, candidate *Candidate) error {

	if candidate.Status == Revoked { //candidate has been withdrawn
		return ErrBondNotNominated()
	}

	var poolAccount crypto.Address
	if candidate.Status == Bonded {
		poolAccount = tr.params.HoldBonded
	} else {
		poolAccount = tr.params.HoldUnbonded
	}

	// XXX refactor all steps like this into GlobalState.addBondedTokens()
	// Move coins from the delegator account to the bonded pool account
	err := tr.transfer(tr.sender, poolAccount, sdk.Coins{tx.Bond})
	if err != nil {
		return err
	}

	// Get or create the delegator bond
	bond := tr.keeper.loadDelegatorBond(tr.sender, tx.PubKey)
	if bond == nil {
		bond = &DelegatorBond{
			PubKey: tx.PubKey,
			Shares: sdk.ZeroRat,
		}
	}

	// Account new shares, save
	bond.Shares = bond.Shares.Add(candidate.addTokens(tx.Bond.Amount, tr.gs))
	tr.keeper.saveCandidate(candidate)
	tr.keeper.saveDelegatorBond(tr.sender, bond)
	tr.keeper.saveGlobalState(tr.gs)
	return nil
}

func (tr transact) unbond(tx TxUnbond) error {

	// check if bond has any shares in it unbond
	bond := tr.keeper.loadDelegatorBond(tr.sender, tx.PubKey)
	sharesStr := viper.GetString(tx.Shares)
	if bond.Shares.LT(sdk.ZeroRat) { // bond shares < tx shares
		return fmt.Errorf("no shares in account to unbond")
	}

	// if shares set to special case Max then we're good
	if sharesStr != "MAX" {
		// test getting rational number from decimal provided
		shares, err := sdk.NewRatFromDecimal(sharesStr)
		if err != nil {
			return err
		}

		// test that there are enough shares to unbond
		if bond.Shares.LT(shares) {
			return fmt.Errorf("not enough bond shares to unbond, have %v, trying to unbond %v",
				bond.Shares, tx.Shares)
		}
	}
	// XXX end of old checkTx

	// get delegator bond
	bond := tr.keeper.loadDelegatorBond(tr.sender, tx.PubKey)
	if bond == nil {
		return ErrNoDelegatorForAddress()
	}

	// retrieve the amount of bonds to remove (TODO remove redundancy already serialized)
	var shares sdk.Rat
	if tx.Shares == "MAX" {
		shares = bond.Shares
	} else {
		var err error
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
	candidate := tr.keeper.loadCandidate(tx.PubKey)
	if candidate == nil {
		return ErrNoCandidateForAddress()
	}

	revokeCandidacy := false
	if bond.Shares.IsZero() {

		// if the bond is the owner of the candidate then
		// trigger a revoke candidacy
		if tr.sender.Equals(candidate.Owner) &&
			candidate.Status != Revoked {
			revokeCandidacy = true
		}

		// remove the bond
		tr.keeper.removeDelegatorBond(tr.sender, tx.PubKey)
	} else {
		tr.keeper.saveDelegatorBond(tr.sender, bond)
	}

	// transfer coins back to account
	var poolAccount crypto.Address
	if candidate.Status == Bonded {
		poolAccount = tr.params.HoldBonded
	} else {
		poolAccount = tr.params.HoldUnbonded
	}

	returnCoins := candidate.removeShares(shares, tr.gs)
	err := tr.transfer(poolAccount, tr.sender,
		sdk.Coins{{tr.params.AllowedBondDenom, returnCoins}})
	if err != nil {
		return err
	}

	// lastly if an revoke candidate if necessary
	if revokeCandidacy {

		// change the share types to unbonded if they were not already
		if candidate.Status == Bonded {
			err = tr.bondedToUnbondedPool(candidate)
			if err != nil {
				return err
			}
		}

		// lastly update the status
		candidate.Status = Revoked
	}

	// deduct shares from the candidate and save
	if candidate.Liabilities.IsZero() {
		tr.keeper.removeCandidate(tx.PubKey)
	} else {
		tr.keeper.saveCandidate(candidate)
	}

	tr.keeper.saveGlobalState(tr.gs)
	return nil
}
