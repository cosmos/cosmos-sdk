package types

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"time"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/encoding"
	tmtypes "github.com/tendermint/tendermint/types"
	"gopkg.in/yaml.v2"

	"github.com/cosmos/cosmos-sdk/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	// TODO: Why can't we just have one string description which can be JSON by convention
	MaxMonikerLength         = 70
	MaxIdentityLength        = 3000
	MaxWebsiteLength         = 140
	MaxSecurityContactLength = 140
	MaxDetailsLength         = 280
)

var (
	BondStatusUnspecified = BondStatus_name[int32(Unspecified)]
	BondStatusUnbonded    = BondStatus_name[int32(Unbonded)]
	BondStatusUnbonding   = BondStatus_name[int32(Unbonding)]
	BondStatusBonded      = BondStatus_name[int32(Bonded)]
)

var _ ValidatorI = Validator{}

// NewValidator constructs a new Validator
//nolint:interfacer
func NewValidator(operator sdk.ValAddress, pubKey crypto.PubKey, description Description) Validator {
	var pkStr string
	if pubKey != nil {
		pkStr = sdk.MustBech32ifyPubKey(sdk.Bech32PubKeyTypeConsPub, pubKey)
	}

	return Validator{
		OperatorAddress:   operator.String(),
		ConsensusPubkey:   pkStr,
		Jailed:            false,
		Status:            Unbonded,
		Tokens:            sdk.ZeroInt(),
		DelegatorShares:   sdk.ZeroDec(),
		Description:       description,
		UnbondingHeight:   int64(0),
		UnbondingTime:     time.Unix(0, 0).UTC(),
		Commission:        NewCommission(sdk.ZeroDec(), sdk.ZeroDec(), sdk.ZeroDec()),
		MinSelfDelegation: sdk.OneInt(),
	}
}

// String implements the Stringer interface for a Validator object.
func (v Validator) String() string {
	out, _ := yaml.Marshal(v)
	return string(out)
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
func (v Validators) ToSDKValidators() (validators []ValidatorI) {
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
	return bytes.Compare(v[i].GetOperator().Bytes(), v[j].GetOperator().Bytes()) == -1
}

// Implements sort interface
func (v Validators) Swap(i, j int) {
	it := v[i]
	v[i] = v[j]
	v[j] = it
}

// return the redelegation
func MustMarshalValidator(cdc codec.BinaryMarshaler, validator Validator) []byte {
	return cdc.MustMarshalBinaryBare(&validator)
}

// unmarshal a redelegation from a store value
func MustUnmarshalValidator(cdc codec.BinaryMarshaler, value []byte) Validator {
	validator, err := UnmarshalValidator(cdc, value)
	if err != nil {
		panic(err)
	}

	return validator
}

// unmarshal a redelegation from a store value
func UnmarshalValidator(cdc codec.BinaryMarshaler, value []byte) (v Validator, err error) {
	err = cdc.UnmarshalBinaryBare(value, &v)
	return v, err
}

// IsBonded checks if the validator status equals Bonded
func (v Validator) IsBonded() bool {
	return v.GetStatus() == Bonded
}

// IsUnbonded checks if the validator status equals Unbonded
func (v Validator) IsUnbonded() bool {
	return v.GetStatus() == Unbonded
}

// IsUnbonding checks if the validator status equals Unbonding
func (v Validator) IsUnbonding() bool {
	return v.GetStatus() == Unbonding
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
	pk, err := encoding.PubKeyToProto(v.GetConsPubKey())
	if err != nil {
		panic(err)
	}

	return abci.ValidatorUpdate{
		PubKey: pk,
		Power:  v.ConsensusPower(),
	}
}

// ABCIValidatorUpdateZero returns an abci.ValidatorUpdate from a staking validator type
// with zero power used for validator updates.
func (v Validator) ABCIValidatorUpdateZero() abci.ValidatorUpdate {
	pk, err := encoding.PubKeyToProto(v.GetConsPubKey())
	if err != nil {
		panic(err)
	}

	return abci.ValidatorUpdate{
		PubKey: pk,
		Power:  0,
	}
}

// ToTmValidator casts an SDK validator to a tendermint type Validator.
func (v Validator) ToTmValidator() *tmtypes.Validator {
	return tmtypes.NewValidator(v.GetConsPubKey(), v.ConsensusPower())
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
func (v Validator) UpdateStatus(newStatus BondStatus) Validator {
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

// MinEqual defines a more minimum set of equality conditions when comparing two
// validators.
func (v Validator) MinEqual(other Validator) bool {
	return v.ConsensusPubkey == other.ConsensusPubkey &&
		(v.OperatorAddress == other.OperatorAddress) &&
		v.Status == other.Status &&
		v.Tokens.Equal(other.Tokens) &&
		v.DelegatorShares.Equal(other.DelegatorShares) &&
		v.Description == other.Description &&
		v.Commission.Equal(other.Commission)
}

func (v Validator) IsJailed() bool        { return v.Jailed }
func (v Validator) GetMoniker() string    { return v.Description.Moniker }
func (v Validator) GetStatus() BondStatus { return v.Status }
func (v Validator) GetOperator() sdk.ValAddress {
	if v.OperatorAddress == "" {
		return nil
	}
	addr, err := sdk.ValAddressFromBech32(v.OperatorAddress)
	if err != nil {
		panic(err)
	}
	return addr
}
func (v Validator) GetConsPubKey() crypto.PubKey {
	// The way things are refactored now, v.ConsensusPubkey is sometimes a TM
	// ed25519 pubkey, sometimes our own ed25519 pubkey. This is very ugly and
	// inconsistent.
	// Luckily, here we coerce it into a TM ed25519 pubkey always, as this
	// pubkey will be passed into TM (eg calling encoding.PubKeyToProto).
	pk := sdk.MustGetPubKeyFromBech32(sdk.Bech32PubKeyTypeConsPub, v.ConsensusPubkey)

	if intoTmPk, ok := pk.(cryptotypes.IntoTmPubKey); ok {
		return intoTmPk.AsTmPubKey()
	}

	return pk
}
func (v Validator) GetConsAddr() sdk.ConsAddress  { return sdk.ConsAddress(v.GetConsPubKey().Address()) }
func (v Validator) GetTokens() sdk.Int            { return v.Tokens }
func (v Validator) GetBondedTokens() sdk.Int      { return v.BondedTokens() }
func (v Validator) GetConsensusPower() int64      { return v.ConsensusPower() }
func (v Validator) GetCommission() sdk.Dec        { return v.Commission.Rate }
func (v Validator) GetMinSelfDelegation() sdk.Int { return v.MinSelfDelegation }
func (v Validator) GetDelegatorShares() sdk.Dec   { return v.DelegatorShares }
