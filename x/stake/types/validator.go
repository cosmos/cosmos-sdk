package types

import (
	"bytes"
	"fmt"
	"time"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Validator defines the total amount of bond shares and their exchange rate to
// coins. Accumulation of interest is modelled as an in increase in the
// exchange rate, and slashing as a decrease.  When coins are delegated to this
// validator, the validator is credited with a Delegation whose number of
// bond shares is based on the amount of coins delegated divided by the current
// exchange rate. Voting power can be calculated as total bonds multiplied by
// exchange rate.
type Validator struct {
	OperatorAddr sdk.ValAddress `json:"operator_address"` // address of the validator's operator; bech encoded in JSON
	ConsPubKey   crypto.PubKey  `json:"consensus_pubkey"` // the consensus public key of the validator; bech encoded in JSON
	Jailed       bool           `json:"jailed"`           // has the validator been jailed from bonded status?

	Status          sdk.BondStatus `json:"status"`           // validator status (bonded/unbonding/unbonded)
	Tokens          sdk.Dec        `json:"tokens"`           // delegated tokens (incl. self-delegation)
	DelegatorShares sdk.Dec        `json:"delegator_shares"` // total shares issued to a validator's delegators

	Description        Description `json:"description"`           // description terms for the validator
	BondHeight         int64       `json:"bond_height"`           // earliest height as a bonded validator
	BondIntraTxCounter int16       `json:"bond_intra_tx_counter"` // block-local tx index of validator change

	UnbondingHeight  int64     `json:"unbonding_height"` // if unbonding, height at which this validator has begun unbonding
	UnbondingMinTime time.Time `json:"unbonding_time"`   // if unbonding, min time for the validator to complete unbonding

	Commission            sdk.Dec `json:"commission"`              // XXX the commission rate of fees charged to any delegators
	CommissionMax         sdk.Dec `json:"commission_max"`          // XXX maximum commission rate which this validator can ever charge
	CommissionChangeRate  sdk.Dec `json:"commission_change_rate"`  // XXX maximum daily increase of the validator commission
	CommissionChangeToday sdk.Dec `json:"commission_change_today"` // XXX commission rate change today, reset each day (UTC time)
}

// NewValidator - initialize a new validator
func NewValidator(operator sdk.ValAddress, pubKey crypto.PubKey, description Description) Validator {
	return Validator{
		OperatorAddr:          operator,
		ConsPubKey:            pubKey,
		Jailed:                false,
		Status:                sdk.Unbonded,
		Tokens:                sdk.ZeroDec(),
		DelegatorShares:       sdk.ZeroDec(),
		Description:           description,
		BondHeight:            int64(0),
		BondIntraTxCounter:    int16(0),
		UnbondingHeight:       int64(0),
		UnbondingMinTime:      time.Unix(0, 0).UTC(),
		Commission:            sdk.ZeroDec(),
		CommissionMax:         sdk.ZeroDec(),
		CommissionChangeRate:  sdk.ZeroDec(),
		CommissionChangeToday: sdk.ZeroDec(),
	}
}

// what's kept in the store value
type validatorValue struct {
	ConsPubKey            crypto.PubKey
	Jailed                bool
	Status                sdk.BondStatus
	Tokens                sdk.Dec
	DelegatorShares       sdk.Dec
	Description           Description
	BondHeight            int64
	BondIntraTxCounter    int16
	UnbondingHeight       int64
	UnbondingMinTime      time.Time
	Commission            sdk.Dec
	CommissionMax         sdk.Dec
	CommissionChangeRate  sdk.Dec
	CommissionChangeToday sdk.Dec
}

// return the redelegation without fields contained within the key for the store
func MustMarshalValidator(cdc *codec.Codec, validator Validator) []byte {
	val := validatorValue{
		ConsPubKey:            validator.ConsPubKey,
		Jailed:                validator.Jailed,
		Status:                validator.Status,
		Tokens:                validator.Tokens,
		DelegatorShares:       validator.DelegatorShares,
		Description:           validator.Description,
		BondHeight:            validator.BondHeight,
		BondIntraTxCounter:    validator.BondIntraTxCounter,
		UnbondingHeight:       validator.UnbondingHeight,
		UnbondingMinTime:      validator.UnbondingMinTime,
		Commission:            validator.Commission,
		CommissionMax:         validator.CommissionMax,
		CommissionChangeRate:  validator.CommissionChangeRate,
		CommissionChangeToday: validator.CommissionChangeToday,
	}
	return cdc.MustMarshalBinary(val)
}

// unmarshal a redelegation from a store key and value
func MustUnmarshalValidator(cdc *codec.Codec, operatorAddr, value []byte) Validator {
	validator, err := UnmarshalValidator(cdc, operatorAddr, value)
	if err != nil {
		panic(err)
	}
	return validator
}

// unmarshal a redelegation from a store key and value
func UnmarshalValidator(cdc *codec.Codec, operatorAddr, value []byte) (validator Validator, err error) {
	if len(operatorAddr) != sdk.AddrLen {
		err = fmt.Errorf("%v", ErrBadValidatorAddr(DefaultCodespace).Data())
		return
	}
	var storeValue validatorValue
	err = cdc.UnmarshalBinary(value, &storeValue)
	if err != nil {
		return
	}

	return Validator{
		OperatorAddr:          operatorAddr,
		ConsPubKey:            storeValue.ConsPubKey,
		Jailed:                storeValue.Jailed,
		Tokens:                storeValue.Tokens,
		Status:                storeValue.Status,
		DelegatorShares:       storeValue.DelegatorShares,
		Description:           storeValue.Description,
		BondHeight:            storeValue.BondHeight,
		BondIntraTxCounter:    storeValue.BondIntraTxCounter,
		UnbondingHeight:       storeValue.UnbondingHeight,
		UnbondingMinTime:      storeValue.UnbondingMinTime,
		Commission:            storeValue.Commission,
		CommissionMax:         storeValue.CommissionMax,
		CommissionChangeRate:  storeValue.CommissionChangeRate,
		CommissionChangeToday: storeValue.CommissionChangeToday,
	}, nil
}

// HumanReadableString returns a human readable string representation of a
// validator. An error is returned if the operator or the operator's public key
// cannot be converted to Bech32 format.
func (v Validator) HumanReadableString() (string, error) {
	bechConsPubKey, err := sdk.Bech32ifyConsPub(v.ConsPubKey)
	if err != nil {
		return "", err
	}

	resp := "Validator \n"
	resp += fmt.Sprintf("Operator Address: %s\n", v.OperatorAddr)
	resp += fmt.Sprintf("Validator Consensus Pubkey: %s\n", bechConsPubKey)
	resp += fmt.Sprintf("Jailed: %v\n", v.Jailed)
	resp += fmt.Sprintf("Status: %s\n", sdk.BondStatusToString(v.Status))
	resp += fmt.Sprintf("Tokens: %s\n", v.Tokens.String())
	resp += fmt.Sprintf("Delegator Shares: %s\n", v.DelegatorShares.String())
	resp += fmt.Sprintf("Description: %s\n", v.Description)
	resp += fmt.Sprintf("Bond Height: %d\n", v.BondHeight)
	resp += fmt.Sprintf("Unbonding Height: %d\n", v.UnbondingHeight)
	resp += fmt.Sprintf("Minimum Unbonding Time: %v\n", v.UnbondingMinTime)
	resp += fmt.Sprintf("Commission: %s\n", v.Commission.String())
	resp += fmt.Sprintf("Max Commission Rate: %s\n", v.CommissionMax.String())
	resp += fmt.Sprintf("Commission Change Rate: %s\n", v.CommissionChangeRate.String())
	resp += fmt.Sprintf("Commission Change Today: %s\n", v.CommissionChangeToday.String())

	return resp, nil
}

//___________________________________________________________________

// this is a helper struct used for JSON de- and encoding only
type bechValidator struct {
	OperatorAddr sdk.ValAddress `json:"operator_address"` // the bech32 address of the validator's operator
	ConsPubKey   string         `json:"consensus_pubkey"` // the bech32 consensus public key of the validator
	Jailed       bool           `json:"jailed"`           // has the validator been jailed from bonded status?

	Status          sdk.BondStatus `json:"status"`           // validator status (bonded/unbonding/unbonded)
	Tokens          sdk.Dec        `json:"tokens"`           // delegated tokens (incl. self-delegation)
	DelegatorShares sdk.Dec        `json:"delegator_shares"` // total shares issued to a validator's delegators

	Description        Description `json:"description"`           // description terms for the validator
	BondHeight         int64       `json:"bond_height"`           // earliest height as a bonded validator
	BondIntraTxCounter int16       `json:"bond_intra_tx_counter"` // block-local tx index of validator change

	UnbondingHeight  int64     `json:"unbonding_height"` // if unbonding, height at which this validator has begun unbonding
	UnbondingMinTime time.Time `json:"unbonding_time"`   // if unbonding, min time for the validator to complete unbonding

	Commission            sdk.Dec `json:"commission"`              // XXX the commission rate of fees charged to any delegators
	CommissionMax         sdk.Dec `json:"commission_max"`          // XXX maximum commission rate which this validator can ever charge
	CommissionChangeRate  sdk.Dec `json:"commission_change_rate"`  // XXX maximum daily increase of the validator commission
	CommissionChangeToday sdk.Dec `json:"commission_change_today"` // XXX commission rate change today, reset each day (UTC time)
}

// MarshalJSON marshals the validator to JSON using Bech32
func (v Validator) MarshalJSON() ([]byte, error) {
	bechConsPubKey, err := sdk.Bech32ifyConsPub(v.ConsPubKey)
	if err != nil {
		return nil, err
	}

	return codec.Cdc.MarshalJSON(bechValidator{
		OperatorAddr:          v.OperatorAddr,
		ConsPubKey:            bechConsPubKey,
		Jailed:                v.Jailed,
		Status:                v.Status,
		Tokens:                v.Tokens,
		DelegatorShares:       v.DelegatorShares,
		Description:           v.Description,
		BondHeight:            v.BondHeight,
		BondIntraTxCounter:    v.BondIntraTxCounter,
		UnbondingHeight:       v.UnbondingHeight,
		UnbondingMinTime:      v.UnbondingMinTime,
		Commission:            v.Commission,
		CommissionMax:         v.CommissionMax,
		CommissionChangeRate:  v.CommissionChangeRate,
		CommissionChangeToday: v.CommissionChangeToday,
	})
}

// UnmarshalJSON unmarshals the validator from JSON using Bech32
func (v *Validator) UnmarshalJSON(data []byte) error {
	bv := &bechValidator{}
	if err := codec.Cdc.UnmarshalJSON(data, bv); err != nil {
		return err
	}
	consPubKey, err := sdk.GetConsPubKeyBech32(bv.ConsPubKey)
	if err != nil {
		return err
	}
	*v = Validator{
		OperatorAddr:          bv.OperatorAddr,
		ConsPubKey:            consPubKey,
		Jailed:                bv.Jailed,
		Tokens:                bv.Tokens,
		Status:                bv.Status,
		DelegatorShares:       bv.DelegatorShares,
		Description:           bv.Description,
		BondHeight:            bv.BondHeight,
		BondIntraTxCounter:    bv.BondIntraTxCounter,
		UnbondingHeight:       bv.UnbondingHeight,
		UnbondingMinTime:      bv.UnbondingMinTime,
		Commission:            bv.Commission,
		CommissionMax:         bv.CommissionMax,
		CommissionChangeRate:  bv.CommissionChangeRate,
		CommissionChangeToday: bv.CommissionChangeToday,
	}
	return nil
}

//___________________________________________________________________

// only the vitals - does not check bond height of IntraTxCounter
// nolint gocyclo - why dis fail?
func (v Validator) Equal(c2 Validator) bool {
	return v.ConsPubKey.Equals(c2.ConsPubKey) &&
		bytes.Equal(v.OperatorAddr, c2.OperatorAddr) &&
		v.Status.Equal(c2.Status) &&
		v.Tokens.Equal(c2.Tokens) &&
		v.DelegatorShares.Equal(c2.DelegatorShares) &&
		v.Description == c2.Description &&
		v.Commission.Equal(c2.Commission) &&
		v.CommissionMax.Equal(c2.CommissionMax) &&
		v.CommissionChangeRate.Equal(c2.CommissionChangeRate) &&
		v.CommissionChangeToday.Equal(c2.CommissionChangeToday)
}

// return the TM validator address
func (v Validator) ConsAddress() sdk.ConsAddress {
	return sdk.ConsAddress(v.ConsPubKey.Address())
}

// constant used in flags to indicate that description field should not be updated
const DoNotModifyDesc = "[do-not-modify]"

// Description - description fields for a validator
type Description struct {
	Moniker  string `json:"moniker"`  // name
	Identity string `json:"identity"` // optional identity signature (ex. UPort or Keybase)
	Website  string `json:"website"`  // optional website link
	Details  string `json:"details"`  // optional details
}

// NewDescription returns a new Description with the provided values.
func NewDescription(moniker, identity, website, details string) Description {
	return Description{
		Moniker:  moniker,
		Identity: identity,
		Website:  website,
		Details:  details,
	}
}

// UpdateDescription updates the fields of a given description. An error is
// returned if the resulting description contains an invalid length.
func (d Description) UpdateDescription(d2 Description) (Description, sdk.Error) {
	if d2.Moniker == DoNotModifyDesc {
		d2.Moniker = d.Moniker
	}
	if d2.Identity == DoNotModifyDesc {
		d2.Identity = d.Identity
	}
	if d2.Website == DoNotModifyDesc {
		d2.Website = d.Website
	}
	if d2.Details == DoNotModifyDesc {
		d2.Details = d.Details
	}

	return Description{
		Moniker:  d2.Moniker,
		Identity: d2.Identity,
		Website:  d2.Website,
		Details:  d2.Details,
	}.EnsureLength()
}

// EnsureLength ensures the length of a validator's description.
func (d Description) EnsureLength() (Description, sdk.Error) {
	if len(d.Moniker) > 70 {
		return d, ErrDescriptionLength(DefaultCodespace, "moniker", len(d.Moniker), 70)
	}
	if len(d.Identity) > 3000 {
		return d, ErrDescriptionLength(DefaultCodespace, "identity", len(d.Identity), 3000)
	}
	if len(d.Website) > 140 {
		return d, ErrDescriptionLength(DefaultCodespace, "website", len(d.Website), 140)
	}
	if len(d.Details) > 280 {
		return d, ErrDescriptionLength(DefaultCodespace, "details", len(d.Details), 280)
	}

	return d, nil
}

// ABCIValidator returns an abci.Validator from a staked validator type.
func (v Validator) ABCIValidator() abci.Validator {
	return abci.Validator{
		PubKey:  tmtypes.TM2PB.PubKey(v.ConsPubKey),
		Address: v.ConsPubKey.Address(),
		Power:   v.BondedTokens().RoundInt64(),
	}
}

// ABCIValidatorZero returns an abci.Validator from a staked validator type
// with with zero power used for validator updates.
func (v Validator) ABCIValidatorZero() abci.Validator {
	return abci.Validator{
		PubKey:  tmtypes.TM2PB.PubKey(v.ConsPubKey),
		Address: v.ConsPubKey.Address(),
		Power:   0,
	}
}

// UpdateStatus updates the location of the shares within a validator
// to reflect the new status
func (v Validator) UpdateStatus(pool Pool, NewStatus sdk.BondStatus) (Validator, Pool) {

	switch v.Status {
	case sdk.Unbonded:

		switch NewStatus {
		case sdk.Unbonded:
			return v, pool
		case sdk.Bonded:
			pool = pool.looseTokensToBonded(v.Tokens)
		}
	case sdk.Unbonding:

		switch NewStatus {
		case sdk.Unbonding:
			return v, pool
		case sdk.Bonded:
			pool = pool.looseTokensToBonded(v.Tokens)
		}
	case sdk.Bonded:

		switch NewStatus {
		case sdk.Bonded:
			return v, pool
		default:
			pool = pool.bondedTokensToLoose(v.Tokens)
		}
	}

	v.Status = NewStatus
	return v, pool
}

// removes tokens from a validator
func (v Validator) RemoveTokens(pool Pool, tokens sdk.Dec) (Validator, Pool) {
	if v.Status == sdk.Bonded {
		pool = pool.bondedTokensToLoose(tokens)
	}

	v.Tokens = v.Tokens.Sub(tokens)
	return v, pool
}

//_________________________________________________________________________________________________________

// AddTokensFromDel adds tokens to a validator
func (v Validator) AddTokensFromDel(pool Pool, amount sdk.Int) (Validator, Pool, sdk.Dec) {

	// bondedShare/delegatedShare
	exRate := v.DelegatorShareExRate()
	amountDec := sdk.NewDecFromInt(amount)

	if v.Status == sdk.Bonded {
		pool = pool.looseTokensToBonded(amountDec)
	}

	v.Tokens = v.Tokens.Add(amountDec)
	issuedShares := amountDec.Quo(exRate)
	v.DelegatorShares = v.DelegatorShares.Add(issuedShares)

	return v, pool, issuedShares
}

// RemoveDelShares removes delegator shares from a validator.
func (v Validator) RemoveDelShares(pool Pool, delShares sdk.Dec) (Validator, Pool, sdk.Dec) {
	issuedTokens := v.DelegatorShareExRate().Mul(delShares)
	v.Tokens = v.Tokens.Sub(issuedTokens)
	v.DelegatorShares = v.DelegatorShares.Sub(delShares)

	if v.Status == sdk.Bonded {
		pool = pool.bondedTokensToLoose(issuedTokens)
	}

	return v, pool, issuedTokens
}

// DelegatorShareExRate gets the exchange rate of tokens over delegator shares.
// UNITS: tokens/delegator-shares
func (v Validator) DelegatorShareExRate() sdk.Dec {
	if v.DelegatorShares.IsZero() {
		return sdk.OneDec()
	}
	return v.Tokens.Quo(v.DelegatorShares)
}

// Get the bonded tokens which the validator holds
func (v Validator) BondedTokens() sdk.Dec {
	if v.Status == sdk.Bonded {
		return v.Tokens
	}
	return sdk.ZeroDec()
}

// Returns if the validator should be considered unbonded
func (v Validator) IsUnbonded(ctx sdk.Context) bool {
	switch v.Status {
	case sdk.Unbonded:
		return true
	case sdk.Unbonding:
		ctxTime := ctx.BlockHeader().Time
		if ctxTime.After(v.UnbondingMinTime) {
			return true
		}
	}
	return false
}

//______________________________________________________________________

// ensure fulfills the sdk validator types
var _ sdk.Validator = Validator{}

// nolint - for sdk.Validator
func (v Validator) GetJailed() bool              { return v.Jailed }
func (v Validator) GetMoniker() string           { return v.Description.Moniker }
func (v Validator) GetStatus() sdk.BondStatus    { return v.Status }
func (v Validator) GetOperator() sdk.ValAddress  { return v.OperatorAddr }
func (v Validator) GetPubKey() crypto.PubKey     { return v.ConsPubKey }
func (v Validator) GetConsAddr() sdk.ConsAddress { return sdk.ConsAddress(v.ConsPubKey.Address()) }
func (v Validator) GetPower() sdk.Dec            { return v.BondedTokens() }
func (v Validator) GetTokens() sdk.Dec           { return v.Tokens }
func (v Validator) GetCommission() sdk.Dec       { return v.Commission }
func (v Validator) GetDelegatorShares() sdk.Dec  { return v.DelegatorShares }
func (v Validator) GetBondHeight() int64         { return v.BondHeight }
