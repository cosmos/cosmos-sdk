package types

import "cosmossdk.io/math"

// import (
// 	"fmt"
// 	"strings"
// 	"time"

// 	"cosmossdk.io/core/address"
// 	"cosmossdk.io/errors"
// 	"cosmossdk.io/math"

// 	"github.com/cosmos/cosmos-sdk/codec"
// 	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
// 	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
// 	sdk "github.com/cosmos/cosmos-sdk/types"
// 	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
// )

// const (
// 	// TODO: Why can't we just have one string description which can be JSON by convention
// 	MaxMonikerLength         = 70
// 	MaxIdentityLength        = 3000
// 	MaxWebsiteLength         = 140
// 	MaxSecurityContactLength = 140
// 	MaxDetailsLength         = 280
// )

// var (
// 	BondStatusUnspecified = BondStatus_name[int32(Unspecified)]
// 	BondStatusUnbonded    = BondStatus_name[int32(Unbonded)]
// 	BondStatusUnbonding   = BondStatus_name[int32(Unbonding)]
// 	BondStatusBonded      = BondStatus_name[int32(Bonded)]
// )

// var _ sdk.ValidatorI = Validator{}

// // NewValidator constructs a new Validator
// func NewValidator(operator string, pubKey cryptotypes.PubKey, description Description) (Validator, error) {
// 	pkAny, err := codectypes.NewAnyWithValue(pubKey)
// 	if err != nil {
// 		return Validator{}, err
// 	}

// 	return Validator{
// 		OperatorAddress:         operator,
// 		ConsensusPubkey:         pkAny,
// 		Jailed:                  false,
// 		Status:                  Unbonded,
// 		Tokens:                  math.ZeroInt(),
// 		DelegatorShares:         math.LegacyZeroDec(),
// 		Description:             description,
// 		UnbondingHeight:         int64(0),
// 		UnbondingTime:           time.Unix(0, 0).UTC(),
// 		Commission:              NewCommission(math.LegacyZeroDec(), math.LegacyZeroDec(), math.LegacyZeroDec()),
// 		MinSelfDelegation:       math.OneInt(),
// 		UnbondingOnHoldRefCount: 0,
// 	}, nil
// }

// // Validators is a collection of Validator
// type Validators struct {
// 	Validators     []Validator
// 	ValidatorCodec address.Codec
// }

// func (v Validators) String() (out string) {
// 	for _, val := range v.Validators {
// 		out += val.String() + "\n"
// 	}

// 	return strings.TrimSpace(out)
// }

// // UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
// func (v Validators) UnpackInterfaces(c codectypes.AnyUnpacker) error {
// 	for i := range v.Validators {
// 		if err := v.Validators[i].UnpackInterfaces(c); err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }

// // return the redelegation
// func MustMarshalValidator(cdc codec.BinaryCodec, validator *Validator) []byte {
// 	return cdc.MustMarshal(validator)
// }

// // unmarshal a redelegation from a store value
// func MustUnmarshalValidator(cdc codec.BinaryCodec, value []byte) Validator {
// 	validator, err := UnmarshalValidator(cdc, value)
// 	if err != nil {
// 		panic(err)
// 	}

// 	return validator
// }

// // unmarshal a redelegation from a store value
// func UnmarshalValidator(cdc codec.BinaryCodec, value []byte) (v Validator, err error) {
// 	err = cdc.Unmarshal(value, &v)
// 	return v, err
// }

// // IsBonded checks if the validator status equals Bonded
// func (v Validator) IsBonded() bool {
// 	return v.GetStatus() == sdk.Bonded
// }

// // IsUnbonded checks if the validator status equals Unbonded
// func (v Validator) IsUnbonded() bool {
// 	return v.GetStatus() == sdk.Unbonded
// }

// // IsUnbonding checks if the validator status equals Unbonding
// func (v Validator) IsUnbonding() bool {
// 	return v.GetStatus() == sdk.Unbonding
// }

// func NewDescription(moniker, identity, website, securityContact, details string) Description {
// 	return Description{
// 		Moniker:         moniker,
// 		Identity:        identity,
// 		Website:         website,
// 		SecurityContact: securityContact,
// 		Details:         details,
// 	}
// }

// // EnsureLength ensures the length of a validator's description.
// func (d Description) EnsureLength() (Description, error) {
// 	if len(d.Moniker) > MaxMonikerLength {
// 		return d, errors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid moniker length; got: %d, max: %d", len(d.Moniker), MaxMonikerLength)
// 	}

// 	if len(d.Identity) > MaxIdentityLength {
// 		return d, errors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid identity length; got: %d, max: %d", len(d.Identity), MaxIdentityLength)
// 	}

// 	if len(d.Website) > MaxWebsiteLength {
// 		return d, errors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid website length; got: %d, max: %d", len(d.Website), MaxWebsiteLength)
// 	}

// 	if len(d.SecurityContact) > MaxSecurityContactLength {
// 		return d, errors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid security contact length; got: %d, max: %d", len(d.SecurityContact), MaxSecurityContactLength)
// 	}

// 	if len(d.Details) > MaxDetailsLength {
// 		return d, errors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid details length; got: %d, max: %d", len(d.Details), MaxDetailsLength)
// 	}

// 	return d, nil
// }

// // SetInitialCommission attempts to set a validator's initial commission. An
// // error is returned if the commission is invalid.
// func (v Validator) SetInitialCommission(commission Commission) (Validator, error) {
// 	if err := commission.Validate(); err != nil {
// 		return v, err
// 	}

// 	v.Commission = commission

// 	return v, nil
// }

// // In some situations, the exchange rate becomes invalid, e.g. if
// // Validator loses all tokens due to slashing. In this case,
// // make all future delegations invalid.
// func (v Validator) InvalidExRate() bool {
// 	return v.Tokens.IsZero() && v.DelegatorShares.IsPositive()
// }

// // calculate the token worth of provided shares
// func (v Validator) TokensFromShares(shares math.LegacyDec) math.LegacyDec {
// 	return (shares.MulInt(v.Tokens)).Quo(v.DelegatorShares)
// }

// // calculate the token worth of provided shares, truncated
// func (v Validator) TokensFromSharesTruncated(shares math.LegacyDec) math.LegacyDec {
// 	return (shares.MulInt(v.Tokens)).QuoTruncate(v.DelegatorShares)
// }

// // TokensFromSharesRoundUp returns the token worth of provided shares, rounded
// // up.
// func (v Validator) TokensFromSharesRoundUp(shares math.LegacyDec) math.LegacyDec {
// 	return (shares.MulInt(v.Tokens)).QuoRoundUp(v.DelegatorShares)
// }

// // SharesFromTokens returns the shares of a delegation given a bond amount. It
// // returns an error if the validator has no tokens.
// func (v Validator) SharesFromTokens(amt math.Int) (math.LegacyDec, error) {
// 	if v.Tokens.IsZero() {
// 		return math.LegacyZeroDec(), ErrInsufficientShares
// 	}

// 	return v.GetDelegatorShares().MulInt(amt).QuoInt(v.GetTokens()), nil
// }

// // SharesFromTokensTruncated returns the truncated shares of a delegation given
// // a bond amount. It returns an error if the validator has no tokens.
// func (v Validator) SharesFromTokensTruncated(amt math.Int) (math.LegacyDec, error) {
// 	if v.Tokens.IsZero() {
// 		return math.LegacyZeroDec(), ErrInsufficientShares
// 	}

// 	return v.GetDelegatorShares().MulInt(amt).QuoTruncate(math.LegacyNewDecFromInt(v.GetTokens())), nil
// }

// // get the bonded tokens which the validator holds
// func (v Validator) BondedTokens() math.Int {
// 	if v.IsBonded() {
// 		return v.Tokens
// 	}

// 	return math.ZeroInt()
// }

// // ConsensusPower gets the consensus-engine power. Aa reduction of 10^6 from
// // validator tokens is applied
// func (v Validator) ConsensusPower(r math.Int) int64 {
// 	if v.IsBonded() {
// 		return v.PotentialConsensusPower(r)
// 	}

// 	return 0
// }

// // PotentialConsensusPower returns the potential consensus-engine power.
// func (v Validator) PotentialConsensusPower(r math.Int) int64 {
// 	return sdk.TokensToConsensusPower(v.Tokens, r)
// }

// // UpdateStatus updates the location of the shares within a validator
// // to reflect the new status
// func (v Validator) UpdateStatus(newStatus BondStatus) Validator {
// 	v.Status = newStatus
// 	return v
// }

// // AddTokensFromDel adds tokens to a validator
// func (v Validator) AddTokensFromDel(amount math.Int) (Validator, math.LegacyDec) {
// 	// calculate the shares to issue
// 	var issuedShares math.LegacyDec
// 	if v.DelegatorShares.IsZero() {
// 		// the first delegation to a validator sets the exchange rate to one
// 		issuedShares = math.LegacyNewDecFromInt(amount)
// 	} else {
// 		shares, err := v.SharesFromTokens(amount)
// 		if err != nil {
// 			panic(err)
// 		}

// 		issuedShares = shares
// 	}

// 	v.Tokens = v.Tokens.Add(amount)
// 	v.DelegatorShares = v.DelegatorShares.Add(issuedShares)

// 	return v, issuedShares
// }

// // RemoveTokens removes tokens from a validator
// func (v Validator) RemoveTokens(tokens math.Int) Validator {
// 	if tokens.IsNegative() {
// 		panic(fmt.Sprintf("should not happen: trying to remove negative tokens %v", tokens))
// 	}

// 	if v.Tokens.LT(tokens) {
// 		panic(fmt.Sprintf("should not happen: only have %v tokens, trying to remove %v", v.Tokens, tokens))
// 	}

// 	v.Tokens = v.Tokens.Sub(tokens)

// 	return v
// }

// func (v Validator) IsJailed() bool            { return v.Jailed }
// func (v Validator) GetMoniker() string        { return v.Description.Moniker }
// func (v Validator) GetStatus() sdk.BondStatus { return sdk.BondStatus(v.Status) }
// func (v Validator) GetOperator() string {
// 	return v.OperatorAddress
// }

// // ConsPubKey returns the validator PubKey as a cryptotypes.PubKey.
// func (v Validator) ConsPubKey() (cryptotypes.PubKey, error) {
// 	pk, ok := v.ConsensusPubkey.GetCachedValue().(cryptotypes.PubKey)
// 	if !ok {
// 		return nil, errors.Wrapf(sdkerrors.ErrInvalidType, "expecting cryptotypes.PubKey, got %T", pk)
// 	}

// 	return pk, nil
// }

// // GetConsAddr extracts Consensus key address
// func (v Validator) GetConsAddr() ([]byte, error) {
// 	pk, ok := v.ConsensusPubkey.GetCachedValue().(cryptotypes.PubKey)
// 	if !ok {
// 		return nil, errors.Wrapf(sdkerrors.ErrInvalidType, "expecting cryptotypes.PubKey, got %T", pk)
// 	}

// 	return pk.Address().Bytes(), nil
// }

func (v Validator) GetTokens() math.Int { return v.Tokens }

// func (v Validator) GetBondedTokens() math.Int { return v.BondedTokens() }
// func (v Validator) GetConsensusPower(r math.Int) int64 {
// 	return v.ConsensusPower(r)
// }
// func (v Validator) GetCommission() math.LegacyDec      { return v.Commission.Rate }
// func (v Validator) GetMinSelfDelegation() math.Int     { return v.MinSelfDelegation }
// func (v Validator) GetDelegatorShares() math.LegacyDec { return v.DelegatorShares }

// // UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
// func (v Validator) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
// 	var pk cryptotypes.PubKey
// 	return unpacker.UnpackAny(v.ConsensusPubkey, &pk)
// }
