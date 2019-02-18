package types

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// nolint
const (
	// TODO: Why can't we just have one string description which can be JSON by convention
	MaxMonikerLength  = 70
	MaxIdentityLength = 3000
	MaxWebsiteLength  = 140
	MaxDetailsLength  = 280
)

// Validator defines the total amount of bond shares and their exchange rate to
// coins. Slashing results in a decrease in the exchange rate, allowing correct
// calculation of future undelegations without iterating over delegators.
// When coins are delegated to this validator, the validator is credited with a
// delegation whose number of bond shares is based on the amount of coins delegated
// divided by the current exchange rate. Voting power can be calculated as total
// bonded shares multiplied by exchange rate.
type Validator struct {
	OperatorAddr            sdk.ValAddress `json:"operator_address"`    // address of the validator's operator; bech encoded in JSON
	ConsPubKey              crypto.PubKey  `json:"consensus_pubkey"`    // the consensus public key of the validator; bech encoded in JSON
	Jailed                  bool           `json:"jailed"`              // has the validator been jailed from bonded status?
	Status                  sdk.BondStatus `json:"status"`              // validator status (bonded/unbonding/unbonded)
	Tokens                  sdk.Int        `json:"tokens"`              // delegated tokens (incl. self-delegation)
	DelegatorShares         sdk.Dec        `json:"delegator_shares"`    // total shares issued to a validator's delegators
	Description             Description    `json:"description"`         // description terms for the validator
	UnbondingHeight         int64          `json:"unbonding_height"`    // if unbonding, height at which this validator has begun unbonding
	UnbondingCompletionTime time.Time      `json:"unbonding_time"`      // if unbonding, min time for the validator to complete unbonding
	Commission              Commission     `json:"commission"`          // commission parameters
	MinSelfDelegation       sdk.Int        `json:"min_self_delegation"` // validator's self declared minimum self delegation
}

// Validators is a collection of Validator
type Validators []Validator

func (v Validators) String() (out string) {
	for _, val := range v {
		out += val.String() + "\n"
	}
	return strings.TrimSpace(out)
}

// ToSDKValidators -  convenience function convert []Validators to []sdk.Validators
func (v Validators) ToSDKValidators() (validators []sdk.Validator) {
	for _, val := range v {
		validators = append(validators, val)
	}
	return validators
}

// NewValidator - initialize a new validator
func NewValidator(operator sdk.ValAddress, pubKey crypto.PubKey, description Description) Validator {
	return Validator{
		OperatorAddr:            operator,
		ConsPubKey:              pubKey,
		Jailed:                  false,
		Status:                  sdk.Unbonded,
		Tokens:                  sdk.ZeroInt(),
		DelegatorShares:         sdk.ZeroDec(),
		Description:             description,
		UnbondingHeight:         int64(0),
		UnbondingCompletionTime: time.Unix(0, 0).UTC(),
		Commission:              NewCommission(sdk.ZeroDec(), sdk.ZeroDec(), sdk.ZeroDec()),
		MinSelfDelegation:       sdk.OneInt(),
	}
}

// return the redelegation
func MustMarshalValidator(cdc *codec.Codec, validator Validator) []byte {
	return cdc.MustMarshalBinaryLengthPrefixed(validator)
}

// unmarshal a redelegation from a store value
func MustUnmarshalValidator(cdc *codec.Codec, value []byte) Validator {
	validator, err := UnmarshalValidator(cdc, value)
	if err != nil {
		panic(err)
	}
	return validator
}

// unmarshal a redelegation from a store value
func UnmarshalValidator(cdc *codec.Codec, value []byte) (validator Validator, err error) {
	err = cdc.UnmarshalBinaryLengthPrefixed(value, &validator)
	return validator, err
}

// String returns a human readable string representation of a validator.
func (v Validator) String() string {
	bechConsPubKey, err := sdk.Bech32ifyConsPub(v.ConsPubKey)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf(`Validator
  Operator Address:           %s
  Validator Consensus Pubkey: %s
  Jailed:                     %v
  Status:                     %s
  Tokens:                     %s
  Delegator Shares:           %s
  Description:                %s
  Unbonding Height:           %d
  Unbonding Completion Time:  %v
  Minimum Self Delegation:    %v
  Commission:                 %s`, v.OperatorAddr, bechConsPubKey,
		v.Jailed, sdk.BondStatusToString(v.Status), v.Tokens,
		v.DelegatorShares, v.Description,
		v.UnbondingHeight, v.UnbondingCompletionTime, v.MinSelfDelegation, v.Commission)
}

// this is a helper struct used for JSON de- and encoding only
type bechValidator struct {
	OperatorAddr            sdk.ValAddress `json:"operator_address"`    // the bech32 address of the validator's operator
	ConsPubKey              string         `json:"consensus_pubkey"`    // the bech32 consensus public key of the validator
	Jailed                  bool           `json:"jailed"`              // has the validator been jailed from bonded status?
	Status                  sdk.BondStatus `json:"status"`              // validator status (bonded/unbonding/unbonded)
	Tokens                  sdk.Int        `json:"tokens"`              // delegated tokens (incl. self-delegation)
	DelegatorShares         sdk.Dec        `json:"delegator_shares"`    // total shares issued to a validator's delegators
	Description             Description    `json:"description"`         // description terms for the validator
	UnbondingHeight         int64          `json:"unbonding_height"`    // if unbonding, height at which this validator has begun unbonding
	UnbondingCompletionTime time.Time      `json:"unbonding_time"`      // if unbonding, min time for the validator to complete unbonding
	Commission              Commission     `json:"commission"`          // commission parameters
	MinSelfDelegation       sdk.Int        `json:"min_self_delegation"` // minimum self delegation
}

// MarshalJSON marshals the validator to JSON using Bech32
func (v Validator) MarshalJSON() ([]byte, error) {
	bechConsPubKey, err := sdk.Bech32ifyConsPub(v.ConsPubKey)
	if err != nil {
		return nil, err
	}

	return codec.Cdc.MarshalJSON(bechValidator{
		OperatorAddr:            v.OperatorAddr,
		ConsPubKey:              bechConsPubKey,
		Jailed:                  v.Jailed,
		Status:                  v.Status,
		Tokens:                  v.Tokens,
		DelegatorShares:         v.DelegatorShares,
		Description:             v.Description,
		UnbondingHeight:         v.UnbondingHeight,
		UnbondingCompletionTime: v.UnbondingCompletionTime,
		MinSelfDelegation:       v.MinSelfDelegation,
		Commission:              v.Commission,
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
		OperatorAddr:            bv.OperatorAddr,
		ConsPubKey:              consPubKey,
		Jailed:                  bv.Jailed,
		Tokens:                  bv.Tokens,
		Status:                  bv.Status,
		DelegatorShares:         bv.DelegatorShares,
		Description:             bv.Description,
		UnbondingHeight:         bv.UnbondingHeight,
		UnbondingCompletionTime: bv.UnbondingCompletionTime,
		Commission:              bv.Commission,
		MinSelfDelegation:       bv.MinSelfDelegation,
	}
	return nil
}

// only the vitals
func (v Validator) TestEquivalent(v2 Validator) bool {
	return v.ConsPubKey.Equals(v2.ConsPubKey) &&
		bytes.Equal(v.OperatorAddr, v2.OperatorAddr) &&
		v.Status.Equal(v2.Status) &&
		v.Tokens.Equal(v2.Tokens) &&
		v.DelegatorShares.Equal(v2.DelegatorShares) &&
		v.Description == v2.Description &&
		v.Commission.Equal(v2.Commission)
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
	if len(d.Moniker) > MaxMonikerLength {
		return d, ErrDescriptionLength(DefaultCodespace, "moniker", len(d.Moniker), MaxMonikerLength)
	}
	if len(d.Identity) > MaxIdentityLength {
		return d, ErrDescriptionLength(DefaultCodespace, "identity", len(d.Identity), MaxIdentityLength)
	}
	if len(d.Website) > MaxWebsiteLength {
		return d, ErrDescriptionLength(DefaultCodespace, "website", len(d.Website), MaxWebsiteLength)
	}
	if len(d.Details) > MaxDetailsLength {
		return d, ErrDescriptionLength(DefaultCodespace, "details", len(d.Details), MaxDetailsLength)
	}

	return d, nil
}

// ABCIValidatorUpdate returns an abci.ValidatorUpdate from a staking validator type
// with the full validator power
func (v Validator) ABCIValidatorUpdate() abci.ValidatorUpdate {
	return abci.ValidatorUpdate{
		PubKey: tmtypes.TM2PB.PubKey(v.ConsPubKey),
		Power:  v.TendermintPower(),
	}
}

// ABCIValidatorUpdateZero returns an abci.ValidatorUpdate from a staking validator type
// with zero power used for validator updates.
func (v Validator) ABCIValidatorUpdateZero() abci.ValidatorUpdate {
	return abci.ValidatorUpdate{
		PubKey: tmtypes.TM2PB.PubKey(v.ConsPubKey),
		Power:  0,
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
			pool = pool.notBondedTokensToBonded(v.Tokens)
		}
	case sdk.Unbonding:

		switch NewStatus {
		case sdk.Unbonding:
			return v, pool
		case sdk.Bonded:
			pool = pool.notBondedTokensToBonded(v.Tokens)
		}
	case sdk.Bonded:

		switch NewStatus {
		case sdk.Bonded:
			return v, pool
		default:
			pool = pool.bondedTokensToNotBonded(v.Tokens)
		}
	}

	v.Status = NewStatus
	return v, pool
}

// removes tokens from a validator
func (v Validator) RemoveTokens(pool Pool, tokens sdk.Int) (Validator, Pool) {
	if tokens.IsNegative() {
		panic(fmt.Sprintf("should not happen: trying to remove negative tokens %v", tokens))
	}
	if v.Tokens.LT(tokens) {
		panic(fmt.Sprintf("should not happen: only have %v tokens, trying to remove %v", v.Tokens, tokens))
	}
	v.Tokens = v.Tokens.Sub(tokens)
	// TODO: It is not obvious from the name of the function that this will happen. Either justify or move outside.
	if v.Status == sdk.Bonded {
		pool = pool.bondedTokensToNotBonded(tokens)
	}
	return v, pool
}

// SetInitialCommission attempts to set a validator's initial commission. An
// error is returned if the commission is invalid.
func (v Validator) SetInitialCommission(commission Commission) (Validator, sdk.Error) {
	if err := commission.Validate(); err != nil {
		return v, err
	}

	v.Commission = commission
	return v, nil
}

// AddTokensFromDel adds tokens to a validator
// CONTRACT: Tokens are assumed to have come from not-bonded pool.
func (v Validator) AddTokensFromDel(pool Pool, amount sdk.Int) (Validator, Pool, sdk.Dec) {

	// bondedShare/delegatedShare
	exRate := v.DelegatorShareExRate()
	if exRate.IsZero() {
		panic("zero exRate should not happen")
	}

	if v.Status == sdk.Bonded {
		pool = pool.notBondedTokensToBonded(amount)
	}

	v.Tokens = v.Tokens.Add(amount)
	issuedShares := sdk.NewDecFromInt(amount).Quo(exRate)
	v.DelegatorShares = v.DelegatorShares.Add(issuedShares)

	return v, pool, issuedShares
}

// RemoveDelShares removes delegator shares from a validator.
// NOTE: because token fractions are left in the valiadator,
//       the exchange rate of future shares of this validator can increase.
// CONTRACT: Tokens are assumed to move to the not-bonded pool.
func (v Validator) RemoveDelShares(pool Pool, delShares sdk.Dec) (Validator, Pool, sdk.Int) {

	remainingShares := v.DelegatorShares.Sub(delShares)
	var issuedTokens sdk.Int
	if remainingShares.IsZero() {

		// last delegation share gets any trimmings
		issuedTokens = v.Tokens
		v.Tokens = sdk.ZeroInt()
	} else {

		// leave excess tokens in the validator
		// however fully use all the delegator shares
		issuedTokens = v.DelegatorShareExRate().Mul(delShares).TruncateInt()
		v.Tokens = v.Tokens.Sub(issuedTokens)
		if v.Tokens.IsNegative() {
			panic("attempting to remove more tokens than available in validator")
		}
	}

	v.DelegatorShares = remainingShares
	if v.Status == sdk.Bonded {
		pool = pool.bondedTokensToNotBonded(issuedTokens)
	}

	return v, pool, issuedTokens
}

// DelegatorShareExRate gets the exchange rate of tokens over delegator shares.
// UNITS: tokens/delegator-shares
func (v Validator) DelegatorShareExRate() sdk.Dec {
	if v.DelegatorShares.IsZero() {
		// the first delegation to a validator sets the exchange rate to one
		return sdk.OneDec()
	}
	return sdk.NewDecFromInt(v.Tokens).Quo(v.DelegatorShares)
}

// get the bonded tokens which the validator holds
func (v Validator) BondedTokens() sdk.Int {
	if v.Status == sdk.Bonded {
		return v.Tokens
	}
	return sdk.ZeroInt()
}

// get the Tendermint Power
// a reduction of 10^6 from validator tokens is applied
func (v Validator) TendermintPower() int64 {
	if v.Status == sdk.Bonded {
		return v.PotentialTendermintPower()
	}
	return 0
}

// potential Tendermint power
func (v Validator) PotentialTendermintPower() int64 {
	return sdk.TokensToTendermintPower(v.Tokens)
}

// ensure fulfills the sdk validator types
var _ sdk.Validator = Validator{}

// nolint - for sdk.Validator
func (v Validator) GetJailed() bool                  { return v.Jailed }
func (v Validator) GetMoniker() string               { return v.Description.Moniker }
func (v Validator) GetStatus() sdk.BondStatus        { return v.Status }
func (v Validator) GetOperator() sdk.ValAddress      { return v.OperatorAddr }
func (v Validator) GetConsPubKey() crypto.PubKey     { return v.ConsPubKey }
func (v Validator) GetConsAddr() sdk.ConsAddress     { return sdk.ConsAddress(v.ConsPubKey.Address()) }
func (v Validator) GetTokens() sdk.Int               { return v.Tokens }
func (v Validator) GetBondedTokens() sdk.Int         { return v.BondedTokens() }
func (v Validator) GetTendermintPower() int64        { return v.TendermintPower() }
func (v Validator) GetCommission() sdk.Dec           { return v.Commission.Rate }
func (v Validator) GetMinSelfDelegation() sdk.Int    { return v.MinSelfDelegation }
func (v Validator) GetDelegatorShares() sdk.Dec      { return v.DelegatorShares }
func (v Validator) GetDelegatorShareExRate() sdk.Dec { return v.DelegatorShareExRate() }
