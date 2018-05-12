package stake

import (
	"bytes"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	abci "github.com/tendermint/abci/types"
	crypto "github.com/tendermint/go-crypto"
)

// Validator defines the total amount of bond shares and their exchange rate to
// coins. Accumulation of interest is modelled as an in increase in the
// exchange rate, and slashing as a decrease.  When coins are delegated to this
// validator, the validator is credited with a Delegation whose number of
// bond shares is based on the amount of coins delegated divided by the current
// exchange rate. Voting power can be calculated as total bonds multiplied by
// exchange rate.
type Validator struct {
	Status  sdk.BondStatus `json:"status"`  // bonded status
	Address sdk.Address    `json:"address"` // sender of BondTx - UnbondTx returns here
	PubKey  crypto.PubKey  `json:"pub_key"` // pubkey of validator

	// note: There should only be one of the following 3 shares ever active in a delegator
	//       multiple terms are only added here for clarity.
	BondedShares    sdk.Rat `json:"bonded_shares"`    // total shares of bonded global hold pool
	UnbondingShares sdk.Rat `json:"unbonding_shares"` // total shares of unbonding global hold pool
	UnbondedShares  sdk.Rat `json:"unbonded_shares"`  // total shares of unbonded global hold pool

	DelegatorShares sdk.Rat `json:"liabilities"` // total shares issued to a validator's delegators

	Description        Description `json:"description"`            // description terms for the validator
	BondHeight         int64       `json:"validator_bond_height"`  // earliest height as a bonded validator
	BondIntraTxCounter int16       `json:"validator_bond_counter"` // block-local tx index of validator change
	ProposerRewardPool sdk.Coins   `json:"proposer_reward_pool"`   // XXX reward pool collected from being the proposer

	Commission            sdk.Rat `json:"commission"`              // XXX the commission rate of fees charged to any delegators
	CommissionMax         sdk.Rat `json:"commission_max"`          // XXX maximum commission rate which this validator can ever charge
	CommissionChangeRate  sdk.Rat `json:"commission_change_rate"`  // XXX maximum daily increase of the validator commission
	CommissionChangeToday sdk.Rat `json:"commission_change_today"` // XXX commission rate change today, reset each day (UTC time)

	// fee related
	PrevBondedShares sdk.Rat `json:"prev_bonded_shares"` // total shares of a global hold pools
}

// Validators - list of Validators
type Validators []Validator

// NewValidator - initialize a new validator
func NewValidator(address sdk.Address, pubKey crypto.PubKey, description Description) Validator {
	return Validator{
		Status:                sdk.Unbonded,
		Address:               address,
		PubKey:                pubKey,
		BondedShares:          sdk.ZeroRat(),
		UnbondingShares:       sdk.ZeroRat(),
		UnbondedShares:        sdk.ZeroRat(),
		DelegatorShares:       sdk.ZeroRat(),
		Description:           description,
		BondHeight:            int64(0),
		BondIntraTxCounter:    int16(0),
		ProposerRewardPool:    sdk.Coins{},
		Commission:            sdk.ZeroRat(),
		CommissionMax:         sdk.ZeroRat(),
		CommissionChangeRate:  sdk.ZeroRat(),
		CommissionChangeToday: sdk.ZeroRat(),
		PrevBondedShares:      sdk.ZeroRat(),
	}
}

// only the vitals - does not check bond height of IntraTxCounter
func (v Validator) equal(c2 Validator) bool {
	return v.Status == c2.Status &&
		v.PubKey.Equals(c2.PubKey) &&
		bytes.Equal(v.Address, c2.Address) &&
		v.BondedShares.Equal(c2.BondedShares) &&
		v.DelegatorShares.Equal(c2.DelegatorShares) &&
		v.Description == c2.Description &&
		//v.BondHeight == c2.BondHeight &&
		//v.BondIntraTxCounter == c2.BondIntraTxCounter && // counter is always changing
		v.ProposerRewardPool.IsEqual(c2.ProposerRewardPool) &&
		v.Commission.Equal(c2.Commission) &&
		v.CommissionMax.Equal(c2.CommissionMax) &&
		v.CommissionChangeRate.Equal(c2.CommissionChangeRate) &&
		v.CommissionChangeToday.Equal(c2.CommissionChangeToday) &&
		v.PrevBondedShares.Equal(c2.PrevBondedShares)
}

// intended to be used with require/assert:  require.True(ValEq(...))
func ValEq(t *testing.T, exp, got Validator) (*testing.T, bool, string, Validator, Validator) {
	return t, exp.equal(got), "expected:\t%v\ngot:\t\t%v", exp, got
}

// Description - description fields for a validator
type Description struct {
	Moniker  string `json:"moniker"`
	Identity string `json:"identity"`
	Website  string `json:"website"`
	Details  string `json:"details"`
}

func NewDescription(moniker, identity, website, details string) Description {
	return Description{
		Moniker:  moniker,
		Identity: identity,
		Website:  website,
		Details:  details,
	}
}

// get the exchange rate of tokens over delegator shares
// UNITS: eq-val-bonded-shares/delegator-shares
func (v Validator) DelegatorShareExRate(p Pool) sdk.Rat {
	if v.DelegatorShares.IsZero() {
		return sdk.OneRat()
	}
	tokens := p.EquivalentBondedShares(v)
	return tokens.Quo(v.DelegatorShares)
}

// abci validator from stake validator type
func (v Validator) abciValidator(cdc *wire.Codec) abci.Validator {
	return abci.Validator{
		PubKey: v.PubKey.Bytes(),
		Power:  v.BondedShares.Evaluate(),
	}
}

// abci validator from stake validator type
// with zero power used for validator updates
func (v Validator) abciValidatorZero(cdc *wire.Codec) abci.Validator {
	return abci.Validator{
		PubKey: v.PubKey.Bytes(),
		Power:  0,
	}
}

//XXX updateDescription function
//XXX enforce limit to number of description characters

//______________________________________________________________________

// ensure fulfills the sdk validator types
var _ sdk.Validator = Validator{}

// nolint - for sdk.Validator
func (v Validator) GetStatus() sdk.BondStatus { return v.Status }
func (v Validator) GetAddress() sdk.Address   { return v.Address }
func (v Validator) GetPubKey() crypto.PubKey  { return v.PubKey }
func (v Validator) GetPower() sdk.Rat         { return v.BondedShares }
func (v Validator) GetBondHeight() int64      { return v.BondHeight }
