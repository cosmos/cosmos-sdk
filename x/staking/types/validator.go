package types

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"time"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	"github.com/tendermint/tendermint/crypto/sr25519"
	tmtypes "github.com/tendermint/tendermint/types"
	yaml "gopkg.in/yaml.v2"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/staking/exported"
)

// nolint
const (
	// TODO: Why can't we just have one string description which can be JSON by convention
	MaxMonikerLength         = 70
	MaxIdentityLength        = 3000
	MaxWebsiteLength         = 140
	MaxSecurityContactLength = 140
	MaxDetailsLength         = 280
)

// Implements Validator interface
var _ exported.ValidatorI = Validator{}

// Validator defines the total amount of bond shares and their exchange rate to
// coins. Slashing results in a decrease in the exchange rate, allowing correct
// calculation of future undelegations without iterating over delegators.
// When coins are delegated to this validator, the validator is credited with a
// delegation whose number of bond shares is based on the amount of coins delegated
// divided by the current exchange rate. Voting power can be calculated as total
// bonded shares multiplied by exchange rate.
type Validator struct {
	OperatorAddress         sdk.ValAddress `json:"operator_address" yaml:"operator_address"`       // address of the validator's operator; bech encoded in JSON
	ConsPubKey              crypto.PubKey  `json:"consensus_pubkey" yaml:"consensus_pubkey"`       // the consensus public key of the validator; bech encoded in JSON
	Jailed                  bool           `json:"jailed" yaml:"jailed"`                           // has the validator been jailed from bonded status?
	Status                  sdk.BondStatus `json:"status" yaml:"status"`                           // validator status (bonded/unbonding/unbonded)
	Tokens                  sdk.Int        `json:"tokens" yaml:"tokens"`                           // delegated tokens (incl. self-delegation)
	DelegatorShares         sdk.Dec        `json:"delegator_shares" yaml:"delegator_shares"`       // total shares issued to a validator's delegators
	Description             Description    `json:"description" yaml:"description"`                 // description terms for the validator
	UnbondingHeight         int64          `json:"unbonding_height" yaml:"unbonding_height"`       // if unbonding, height at which this validator has begun unbonding
	UnbondingCompletionTime time.Time      `json:"unbonding_time" yaml:"unbonding_time"`           // if unbonding, min time for the validator to complete unbonding
	Commission              Commission     `json:"commission" yaml:"commission"`                   // commission parameters
	MinSelfDelegation       sdk.Int        `json:"min_self_delegation" yaml:"min_self_delegation"` // validator's self declared minimum self delegation
}

func NewValidator(operator sdk.ValAddress, pubKey crypto.PubKey, description Description) Validator {
	return Validator{
		OperatorAddress:         operator,
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

// ToProto converts a Validator into a ValidatorProto type.
func (v Validator) ToProto() ValidatorProto {
	var pk sdk.PublicKey

	switch t := v.ConsPubKey.(type) {
	case secp256k1.PubKeySecp256k1:
		pk = sdk.PublicKey{
			Pub: &sdk.PublicKey_Secp256K1{t.Bytes()},
		}

	case ed25519.PubKeyEd25519:
		pk = sdk.PublicKey{
			Pub: &sdk.PublicKey_Ed25519{t.Bytes()},
		}

	case sr25519.PubKeySr25519:
		pk = sdk.PublicKey{
			Pub: &sdk.PublicKey_Sr25519{t.Bytes()},
		}
	}

	return ValidatorProto{
		OperatorAddress:   v.OperatorAddress,
		ConsensusPubkey:   pk,
		Jailed:            v.Jailed,
		Status:            []byte{byte(v.Status)},
		Tokens:            v.Tokens,
		DelegatorShares:   v.DelegatorShares,
		Description:       v.Description,
		UnbondingHeight:   v.UnbondingHeight,
		UnbondingTime:     v.UnbondingCompletionTime,
		Commission:        v.Commission,
		MinSelfDelegation: v.MinSelfDelegation,
	}
}

// ToValidator converts a ValidatorProto to a Validator type.
func (vp ValidatorProto) ToValidator() Validator {
	var (
		pk  crypto.PubKey
		err error
	)

	switch t := vp.ConsensusPubkey.GetPub().(type) {
	case *sdk.PublicKey_Secp256K1:
		err = ModuleCdc.amino.UnmarshalBinaryBare(t.Secp256K1, &pk)

	case *sdk.PublicKey_Ed25519:
		err = ModuleCdc.amino.UnmarshalBinaryBare(t.Ed25519, &pk)

	case *sdk.PublicKey_Sr25519:
		err = ModuleCdc.amino.UnmarshalBinaryBare(t.Sr25519, &pk)

	default:
		err = fmt.Errorf("unsupported public key type: %T", t)
	}

	if err != nil {
		panic(fmt.Errorf("failed to decode public key: %w", err))
	}

	return Validator{
		OperatorAddress:         vp.OperatorAddress,
		ConsPubKey:              pk,
		Jailed:                  vp.Jailed,
		Status:                  sdk.BondStatus(vp.Status[0]),
		Tokens:                  vp.Tokens,
		DelegatorShares:         vp.DelegatorShares,
		Description:             vp.Description,
		UnbondingHeight:         vp.UnbondingHeight,
		UnbondingCompletionTime: vp.UnbondingTime,
		Commission:              vp.Commission,
		MinSelfDelegation:       vp.MinSelfDelegation,
	}
}

// String implements the Stringer interface for a Validator object.
func (v Validator) String() string {
	bechConsPubKey, _ := sdk.Bech32ifyPubKey(sdk.Bech32PubKeyTypeConsPub, v.ConsPubKey)

	bz, _ := yaml.Marshal(bechValidator{
		OperatorAddress:         v.OperatorAddress,
		ConsPubKey:              bechConsPubKey,
		Jailed:                  v.Jailed,
		Status:                  v.Status,
		Tokens:                  v.Tokens,
		DelegatorShares:         v.DelegatorShares,
		Description:             v.Description,
		UnbondingHeight:         v.UnbondingHeight,
		UnbondingCompletionTime: v.UnbondingCompletionTime,
		Commission:              v.Commission,
		MinSelfDelegation:       v.MinSelfDelegation,
	})

	return string(bz)
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
func (v Validators) ToSDKValidators() (validators []exported.ValidatorI) {
	for _, val := range v {
		validators = append(validators, val)
	}
	return validators
}

// ToTmValidators casts all validators to the corresponding tendermint type.
func (v Validators) ToTmValidators() []*tmtypes.Validator {
	validators := make([]*tmtypes.Validator, len(v))
	for i, val := range v {
		validators[i] = val.ToTmValidator()
	}
	return validators
}

// Sort Validators sorts validator array in ascending operator address order
func (v Validators) Sort() {
	sort.Sort(v)
}

// Implements sort interface
func (v Validators) Len() int {
	return len(v)
}

// Implements sort interface
func (v Validators) Less(i, j int) bool {
	return bytes.Compare(v[i].OperatorAddress, v[j].OperatorAddress) == -1
}

// Implements sort interface
func (v Validators) Swap(i, j int) {
	it := v[i]
	v[i] = v[j]
	v[j] = it
}

// return the redelegation
func MustMarshalValidator(cdc codec.Marshaler, validator Validator) []byte {
	valProto := validator.ToProto()
	return cdc.MustMarshalBinaryLengthPrefixed(&valProto)
}

// unmarshal a redelegation from a store value
func MustUnmarshalValidator(cdc codec.Marshaler, value []byte) Validator {
	validator, err := UnmarshalValidator(cdc, value)
	if err != nil {
		panic(err)
	}
	return validator
}

// unmarshal a redelegation from a store value
func UnmarshalValidator(cdc codec.Marshaler, value []byte) (Validator, error) {
	vp := ValidatorProto{}
	if err := cdc.UnmarshalBinaryLengthPrefixed(value, &vp); err != nil {
		return Validator{}, err
	}

	return vp.ToValidator(), nil
}

// this is a helper struct used for JSON de- and encoding only
type bechValidator struct {
	OperatorAddress         sdk.ValAddress `json:"operator_address" yaml:"operator_address"`       // the bech32 address of the validator's operator
	ConsPubKey              string         `json:"consensus_pubkey" yaml:"consensus_pubkey"`       // the bech32 consensus public key of the validator
	Jailed                  bool           `json:"jailed" yaml:"jailed"`                           // has the validator been jailed from bonded status?
	Status                  sdk.BondStatus `json:"status" yaml:"status"`                           // validator status (bonded/unbonding/unbonded)
	Tokens                  sdk.Int        `json:"tokens" yaml:"tokens"`                           // delegated tokens (incl. self-delegation)
	DelegatorShares         sdk.Dec        `json:"delegator_shares" yaml:"delegator_shares"`       // total shares issued to a validator's delegators
	Description             Description    `json:"description" yaml:"description"`                 // description terms for the validator
	UnbondingHeight         int64          `json:"unbonding_height" yaml:"unbonding_height"`       // if unbonding, height at which this validator has begun unbonding
	UnbondingCompletionTime time.Time      `json:"unbonding_time" yaml:"unbonding_time"`           // if unbonding, min time for the validator to complete unbonding
	Commission              Commission     `json:"commission" yaml:"commission"`                   // commission parameters
	MinSelfDelegation       sdk.Int        `json:"min_self_delegation" yaml:"min_self_delegation"` // minimum self delegation
}

// MarshalJSON marshals the validator to JSON using Bech32
func (v Validator) MarshalJSON() ([]byte, error) {
	bechConsPubKey, err := sdk.Bech32ifyPubKey(sdk.Bech32PubKeyTypeConsPub, v.ConsPubKey)
	if err != nil {
		return nil, err
	}

	return codec.Cdc.MarshalJSON(bechValidator{
		OperatorAddress:         v.OperatorAddress,
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
	consPubKey, err := sdk.GetPubKeyFromBech32(sdk.Bech32PubKeyTypeConsPub, bv.ConsPubKey)
	if err != nil {
		return err
	}
	*v = Validator{
		OperatorAddress:         bv.OperatorAddress,
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
		bytes.Equal(v.OperatorAddress, v2.OperatorAddress) &&
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

// IsBonded checks if the validator status equals Bonded
func (v Validator) IsBonded() bool {
	return v.GetStatus().Equal(sdk.Bonded)
}

// IsUnbonded checks if the validator status equals Unbonded
func (v Validator) IsUnbonded() bool {
	return v.GetStatus().Equal(sdk.Unbonded)
}

// IsUnbonding checks if the validator status equals Unbonding
func (v Validator) IsUnbonding() bool {
	return v.GetStatus().Equal(sdk.Unbonding)
}

// constant used in flags to indicate that description field should not be updated
const DoNotModifyDesc = "[do-not-modify]"

func NewDescription(moniker, identity, website, securityContact, details string) Description {
	return Description{
		Moniker:         moniker,
		Identity:        identity,
		Website:         website,
		SecurityContact: securityContact,
		Details:         details,
	}
}

// String implements the Stringer interface for a Description object.
func (d Description) String() string {
	out, _ := yaml.Marshal(d)
	return string(out)
}

// UpdateDescription updates the fields of a given description. An error is
// returned if the resulting description contains an invalid length.
func (d Description) UpdateDescription(d2 Description) (Description, error) {
	if d2.Moniker == DoNotModifyDesc {
		d2.Moniker = d.Moniker
	}
	if d2.Identity == DoNotModifyDesc {
		d2.Identity = d.Identity
	}
	if d2.Website == DoNotModifyDesc {
		d2.Website = d.Website
	}
	if d2.SecurityContact == DoNotModifyDesc {
		d2.SecurityContact = d.SecurityContact
	}
	if d2.Details == DoNotModifyDesc {
		d2.Details = d.Details
	}

	return NewDescription(
		d2.Moniker,
		d2.Identity,
		d2.Website,
		d2.SecurityContact,
		d2.Details,
	).EnsureLength()
}

// EnsureLength ensures the length of a validator's description.
func (d Description) EnsureLength() (Description, error) {
	if len(d.Moniker) > MaxMonikerLength {
		return d, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid moniker length; got: %d, max: %d", len(d.Moniker), MaxMonikerLength)
	}
	if len(d.Identity) > MaxIdentityLength {
		return d, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid identity length; got: %d, max: %d", len(d.Identity), MaxIdentityLength)
	}
	if len(d.Website) > MaxWebsiteLength {
		return d, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid website length; got: %d, max: %d", len(d.Website), MaxWebsiteLength)
	}
	if len(d.SecurityContact) > MaxSecurityContactLength {
		return d, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid security contact length; got: %d, max: %d", len(d.SecurityContact), MaxSecurityContactLength)
	}
	if len(d.Details) > MaxDetailsLength {
		return d, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid details length; got: %d, max: %d", len(d.Details), MaxDetailsLength)
	}

	return d, nil
}

// ABCIValidatorUpdate returns an abci.ValidatorUpdate from a staking validator type
// with the full validator power
func (v Validator) ABCIValidatorUpdate() abci.ValidatorUpdate {
	return abci.ValidatorUpdate{
		PubKey: tmtypes.TM2PB.PubKey(v.ConsPubKey),
		Power:  v.ConsensusPower(),
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

// ToTmValidator casts an SDK validator to a tendermint type Validator.
func (v Validator) ToTmValidator() *tmtypes.Validator {
	return tmtypes.NewValidator(v.ConsPubKey, v.ConsensusPower())
}

// SetInitialCommission attempts to set a validator's initial commission. An
// error is returned if the commission is invalid.
func (v Validator) SetInitialCommission(commission Commission) (Validator, error) {
	if err := commission.Validate(); err != nil {
		return v, err
	}

	v.Commission = commission
	return v, nil
}

// In some situations, the exchange rate becomes invalid, e.g. if
// Validator loses all tokens due to slashing. In this case,
// make all future delegations invalid.
func (v Validator) InvalidExRate() bool {
	return v.Tokens.IsZero() && v.DelegatorShares.IsPositive()
}

// calculate the token worth of provided shares
func (v Validator) TokensFromShares(shares sdk.Dec) sdk.Dec {
	return (shares.MulInt(v.Tokens)).Quo(v.DelegatorShares)
}

// calculate the token worth of provided shares, truncated
func (v Validator) TokensFromSharesTruncated(shares sdk.Dec) sdk.Dec {
	return (shares.MulInt(v.Tokens)).QuoTruncate(v.DelegatorShares)
}

// TokensFromSharesRoundUp returns the token worth of provided shares, rounded
// up.
func (v Validator) TokensFromSharesRoundUp(shares sdk.Dec) sdk.Dec {
	return (shares.MulInt(v.Tokens)).QuoRoundUp(v.DelegatorShares)
}

// SharesFromTokens returns the shares of a delegation given a bond amount. It
// returns an error if the validator has no tokens.
func (v Validator) SharesFromTokens(amt sdk.Int) (sdk.Dec, error) {
	if v.Tokens.IsZero() {
		return sdk.ZeroDec(), ErrInsufficientShares
	}

	return v.GetDelegatorShares().MulInt(amt).QuoInt(v.GetTokens()), nil
}

// SharesFromTokensTruncated returns the truncated shares of a delegation given
// a bond amount. It returns an error if the validator has no tokens.
func (v Validator) SharesFromTokensTruncated(amt sdk.Int) (sdk.Dec, error) {
	if v.Tokens.IsZero() {
		return sdk.ZeroDec(), ErrInsufficientShares
	}

	return v.GetDelegatorShares().MulInt(amt).QuoTruncate(v.GetTokens().ToDec()), nil
}

// get the bonded tokens which the validator holds
func (v Validator) BondedTokens() sdk.Int {
	if v.IsBonded() {
		return v.Tokens
	}
	return sdk.ZeroInt()
}

// get the consensus-engine power
// a reduction of 10^6 from validator tokens is applied
func (v Validator) ConsensusPower() int64 {
	if v.IsBonded() {
		return v.PotentialConsensusPower()
	}
	return 0
}

// potential consensus-engine power
func (v Validator) PotentialConsensusPower() int64 {
	return sdk.TokensToConsensusPower(v.Tokens)
}

// UpdateStatus updates the location of the shares within a validator
// to reflect the new status
func (v Validator) UpdateStatus(newStatus sdk.BondStatus) Validator {
	v.Status = newStatus
	return v
}

// AddTokensFromDel adds tokens to a validator
func (v Validator) AddTokensFromDel(amount sdk.Int) (Validator, sdk.Dec) {

	// calculate the shares to issue
	var issuedShares sdk.Dec
	if v.DelegatorShares.IsZero() {
		// the first delegation to a validator sets the exchange rate to one
		issuedShares = amount.ToDec()
	} else {
		shares, err := v.SharesFromTokens(amount)
		if err != nil {
			panic(err)
		}

		issuedShares = shares
	}

	v.Tokens = v.Tokens.Add(amount)
	v.DelegatorShares = v.DelegatorShares.Add(issuedShares)

	return v, issuedShares
}

// RemoveTokens removes tokens from a validator
func (v Validator) RemoveTokens(tokens sdk.Int) Validator {
	if tokens.IsNegative() {
		panic(fmt.Sprintf("should not happen: trying to remove negative tokens %v", tokens))
	}
	if v.Tokens.LT(tokens) {
		panic(fmt.Sprintf("should not happen: only have %v tokens, trying to remove %v", v.Tokens, tokens))
	}
	v.Tokens = v.Tokens.Sub(tokens)
	return v
}

// RemoveDelShares removes delegator shares from a validator.
// NOTE: because token fractions are left in the valiadator,
//       the exchange rate of future shares of this validator can increase.
func (v Validator) RemoveDelShares(delShares sdk.Dec) (Validator, sdk.Int) {

	remainingShares := v.DelegatorShares.Sub(delShares)
	var issuedTokens sdk.Int
	if remainingShares.IsZero() {

		// last delegation share gets any trimmings
		issuedTokens = v.Tokens
		v.Tokens = sdk.ZeroInt()
	} else {

		// leave excess tokens in the validator
		// however fully use all the delegator shares
		issuedTokens = v.TokensFromShares(delShares).TruncateInt()
		v.Tokens = v.Tokens.Sub(issuedTokens)
		if v.Tokens.IsNegative() {
			panic("attempting to remove more tokens than available in validator")
		}
	}

	v.DelegatorShares = remainingShares
	return v, issuedTokens
}

// nolint - for ValidatorI
func (v Validator) IsJailed() bool                { return v.Jailed }
func (v Validator) GetMoniker() string            { return v.Description.Moniker }
func (v Validator) GetStatus() sdk.BondStatus     { return v.Status }
func (v Validator) GetOperator() sdk.ValAddress   { return v.OperatorAddress }
func (v Validator) GetConsPubKey() crypto.PubKey  { return v.ConsPubKey }
func (v Validator) GetConsAddr() sdk.ConsAddress  { return sdk.ConsAddress(v.ConsPubKey.Address()) }
func (v Validator) GetTokens() sdk.Int            { return v.Tokens }
func (v Validator) GetBondedTokens() sdk.Int      { return v.BondedTokens() }
func (v Validator) GetConsensusPower() int64      { return v.ConsensusPower() }
func (v Validator) GetCommission() sdk.Dec        { return v.Commission.Rate }
func (v Validator) GetMinSelfDelegation() sdk.Int { return v.MinSelfDelegation }
func (v Validator) GetDelegatorShares() sdk.Dec   { return v.DelegatorShares }
