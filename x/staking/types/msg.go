package types

import (
	"bytes"
	"encoding/json"

	"github.com/tendermint/tendermint/crypto"
	yaml "gopkg.in/yaml.v2"

	"github.com/cosmos/cosmos-sdk/types"
)

// ensure Msg interface compliance at compile time
var (
	_ types.Msg = &MsgCreateValidator{}
	_ types.Msg = &MsgEditValidator{}
	_ types.Msg = &MsgDelegate{}
	_ types.Msg = &MsgUndelegate{}
	_ types.Msg = &MsgBeginRedelegate{}
)

//______________________________________________________________________

// MsgCreateValidator - struct for bonding transactions
type MsgCreateValidator struct {
	Description       Description      `json:"description" yaml:"description"`
	Commission        CommissionRates  `json:"commission" yaml:"commission"`
	MinSelfDelegation types.Int        `json:"min_self_delegation" yaml:"min_self_delegation"`
	DelegatorAddress  types.AccAddress `json:"delegator_address" yaml:"delegator_address"`
	ValidatorAddress  types.ValAddress `json:"validator_address" yaml:"validator_address"`
	PubKey            crypto.PubKey    `json:"pubkey" yaml:"pubkey"`
	Value             types.Coin       `json:"value" yaml:"value"`
}

type msgCreateValidatorJSON struct {
	Description       Description      `json:"description" yaml:"description"`
	Commission        CommissionRates  `json:"commission" yaml:"commission"`
	MinSelfDelegation types.Int        `json:"min_self_delegation" yaml:"min_self_delegation"`
	DelegatorAddress  types.AccAddress `json:"delegator_address" yaml:"delegator_address"`
	ValidatorAddress  types.ValAddress `json:"validator_address" yaml:"validator_address"`
	PubKey            string           `json:"pubkey" yaml:"pubkey"`
	Value             types.Coin       `json:"value" yaml:"value"`
}

// Default way to create validator. Delegator address and validator address are the same
func NewMsgCreateValidator(
	valAddr types.ValAddress, pubKey crypto.PubKey, selfDelegation types.Coin,
	description Description, commission CommissionRates, minSelfDelegation types.Int,
) MsgCreateValidator {

	return MsgCreateValidator{
		Description:       description,
		DelegatorAddress:  types.AccAddress(valAddr),
		ValidatorAddress:  valAddr,
		PubKey:            pubKey,
		Value:             selfDelegation,
		Commission:        commission,
		MinSelfDelegation: minSelfDelegation,
	}
}

//nolint
func (msg MsgCreateValidator) Route() string { return RouterKey }
func (msg MsgCreateValidator) Type() string  { return "create_validator" }

// Return address(es) that must sign over msg.GetSignBytes()
func (msg MsgCreateValidator) GetSigners() []types.AccAddress {
	// delegator is first signer so delegator pays fees
	addrs := []types.AccAddress{msg.DelegatorAddress}

	if !bytes.Equal(msg.DelegatorAddress.Bytes(), msg.ValidatorAddress.Bytes()) {
		// if validator addr is not same as delegator addr, validator must sign
		// msg as well
		addrs = append(addrs, types.AccAddress(msg.ValidatorAddress))
	}
	return addrs
}

// MarshalJSON implements the json.Marshaler interface to provide custom JSON
// serialization of the MsgCreateValidator type.
func (msg MsgCreateValidator) MarshalJSON() ([]byte, error) {
	return json.Marshal(msgCreateValidatorJSON{
		Description:       msg.Description,
		Commission:        msg.Commission,
		DelegatorAddress:  msg.DelegatorAddress,
		ValidatorAddress:  msg.ValidatorAddress,
		PubKey:            types.MustBech32ifyConsPub(msg.PubKey),
		Value:             msg.Value,
		MinSelfDelegation: msg.MinSelfDelegation,
	})
}

// UnmarshalJSON implements the json.Unmarshaler interface to provide custom
// JSON deserialization of the MsgCreateValidator type.
func (msg *MsgCreateValidator) UnmarshalJSON(bz []byte) error {
	var msgCreateValJSON msgCreateValidatorJSON
	if err := json.Unmarshal(bz, &msgCreateValJSON); err != nil {
		return err
	}

	msg.Description = msgCreateValJSON.Description
	msg.Commission = msgCreateValJSON.Commission
	msg.DelegatorAddress = msgCreateValJSON.DelegatorAddress
	msg.ValidatorAddress = msgCreateValJSON.ValidatorAddress
	var err error
	msg.PubKey, err = types.GetConsPubKeyBech32(msgCreateValJSON.PubKey)
	if err != nil {
		return err
	}
	msg.Value = msgCreateValJSON.Value
	msg.MinSelfDelegation = msgCreateValJSON.MinSelfDelegation

	return nil
}

// custom marshal yaml function due to consensus pubkey
func (msg MsgCreateValidator) MarshalYAML() (interface{}, error) {
	bs, err := yaml.Marshal(struct {
		Description       Description
		Commission        CommissionRates
		MinSelfDelegation types.Int
		DelegatorAddress  types.AccAddress
		ValidatorAddress  types.ValAddress
		PubKey            string
		Value             types.Coin
	}{
		Description:       msg.Description,
		Commission:        msg.Commission,
		MinSelfDelegation: msg.MinSelfDelegation,
		DelegatorAddress:  msg.DelegatorAddress,
		ValidatorAddress:  msg.ValidatorAddress,
		PubKey:            types.MustBech32ifyConsPub(msg.PubKey),
		Value:             msg.Value,
	})

	if err != nil {
		return nil, err
	}

	return string(bs), nil
}

// GetSignBytes returns the message bytes to sign over.
func (msg MsgCreateValidator) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return types.MustSortJSON(bz)
}

// quick validity check
func (msg MsgCreateValidator) ValidateBasic() types.Error {
	// note that unmarshaling from bech32 ensures either empty or valid
	if msg.DelegatorAddress.Empty() {
		return ErrNilDelegatorAddr(DefaultCodespace)
	}
	if msg.ValidatorAddress.Empty() {
		return ErrNilValidatorAddr(DefaultCodespace)
	}
	if !types.AccAddress(msg.ValidatorAddress).Equals(msg.DelegatorAddress) {
		return ErrBadValidatorAddr(DefaultCodespace)
	}
	if msg.Value.Amount.LTE(types.ZeroInt()) {
		return ErrBadDelegationAmount(DefaultCodespace)
	}
	if msg.Description == (Description{}) {
		return types.NewError(DefaultCodespace, CodeInvalidInput, "description must be included")
	}
	if msg.Commission == (CommissionRates{}) {
		return types.NewError(DefaultCodespace, CodeInvalidInput, "commission must be included")
	}
	if err := msg.Commission.Validate(); err != nil {
		return err
	}
	if !msg.MinSelfDelegation.IsPositive() {
		return ErrMinSelfDelegationInvalid(DefaultCodespace)
	}
	if msg.Value.Amount.LT(msg.MinSelfDelegation) {
		return ErrSelfDelegationBelowMinimum(DefaultCodespace)
	}

	return nil
}

// MsgEditValidator - struct for editing a validator
type MsgEditValidator struct {
	Description
	ValidatorAddress types.ValAddress `json:"address" yaml:"address"`

	// We pass a reference to the new commission rate and min self delegation as it's not mandatory to
	// update. If not updated, the deserialized rate will be zero with no way to
	// distinguish if an update was intended.
	//
	// REF: #2373
	CommissionRate    *types.Dec `json:"commission_rate" yaml:"commission_rate"`
	MinSelfDelegation *types.Int `json:"min_self_delegation" yaml:"min_self_delegation"`
}

func NewMsgEditValidator(valAddr types.ValAddress, description Description, newRate *types.Dec, newMinSelfDelegation *types.Int) MsgEditValidator {
	return MsgEditValidator{
		Description:       description,
		CommissionRate:    newRate,
		ValidatorAddress:  valAddr,
		MinSelfDelegation: newMinSelfDelegation,
	}
}

//nolint
func (msg MsgEditValidator) Route() string { return RouterKey }
func (msg MsgEditValidator) Type() string  { return "edit_validator" }
func (msg MsgEditValidator) GetSigners() []types.AccAddress {
	return []types.AccAddress{types.AccAddress(msg.ValidatorAddress)}
}

// get the bytes for the message signer to sign on
func (msg MsgEditValidator) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return types.MustSortJSON(bz)
}

// quick validity check
func (msg MsgEditValidator) ValidateBasic() types.Error {
	if msg.ValidatorAddress.Empty() {
		return types.NewError(DefaultCodespace, CodeInvalidInput, "nil validator address")
	}

	if msg.Description == (Description{}) {
		return types.NewError(DefaultCodespace, CodeInvalidInput, "transaction must include some information to modify")
	}

	if msg.MinSelfDelegation != nil && !msg.MinSelfDelegation.IsPositive() {
		return ErrMinSelfDelegationInvalid(DefaultCodespace)
	}

	if msg.CommissionRate != nil {
		if msg.CommissionRate.GT(types.OneDec()) || msg.CommissionRate.IsNegative() {
			return types.NewError(DefaultCodespace, CodeInvalidInput, "commission rate must be between 0 and 1, inclusive")
		}
	}

	return nil
}

// MsgDelegate - struct for bonding transactions
type MsgDelegate struct {
	DelegatorAddress types.AccAddress `json:"delegator_address" yaml:"delegator_address"`
	ValidatorAddress types.ValAddress `json:"validator_address" yaml:"validator_address"`
	Amount           types.Coin       `json:"amount" yaml:"amount"`
}

func NewMsgDelegate(delAddr types.AccAddress, valAddr types.ValAddress, amount types.Coin) MsgDelegate {
	return MsgDelegate{
		DelegatorAddress: delAddr,
		ValidatorAddress: valAddr,
		Amount:           amount,
	}
}

//nolint
func (msg MsgDelegate) Route() string { return RouterKey }
func (msg MsgDelegate) Type() string  { return "delegate" }
func (msg MsgDelegate) GetSigners() []types.AccAddress {
	return []types.AccAddress{msg.DelegatorAddress}
}

// get the bytes for the message signer to sign on
func (msg MsgDelegate) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return types.MustSortJSON(bz)
}

// quick validity check
func (msg MsgDelegate) ValidateBasic() types.Error {
	if msg.DelegatorAddress.Empty() {
		return ErrNilDelegatorAddr(DefaultCodespace)
	}
	if msg.ValidatorAddress.Empty() {
		return ErrNilValidatorAddr(DefaultCodespace)
	}
	if msg.Amount.Amount.LTE(types.ZeroInt()) {
		return ErrBadDelegationAmount(DefaultCodespace)
	}
	return nil
}

//______________________________________________________________________

// MsgDelegate - struct for bonding transactions
type MsgBeginRedelegate struct {
	DelegatorAddress    types.AccAddress `json:"delegator_address" yaml:"delegator_address"`
	ValidatorSrcAddress types.ValAddress `json:"validator_src_address" yaml:"validator_src_address"`
	ValidatorDstAddress types.ValAddress `json:"validator_dst_address" yaml:"validator_dst_address"`
	Amount              types.Coin       `json:"amount" yaml:"amount"`
}

func NewMsgBeginRedelegate(delAddr types.AccAddress, valSrcAddr,
	valDstAddr types.ValAddress, amount types.Coin) MsgBeginRedelegate {

	return MsgBeginRedelegate{
		DelegatorAddress:    delAddr,
		ValidatorSrcAddress: valSrcAddr,
		ValidatorDstAddress: valDstAddr,
		Amount:              amount,
	}
}

//nolint
func (msg MsgBeginRedelegate) Route() string { return RouterKey }
func (msg MsgBeginRedelegate) Type() string  { return "begin_redelegate" }
func (msg MsgBeginRedelegate) GetSigners() []types.AccAddress {
	return []types.AccAddress{msg.DelegatorAddress}
}

// get the bytes for the message signer to sign on
func (msg MsgBeginRedelegate) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return types.MustSortJSON(bz)
}

// quick validity check
func (msg MsgBeginRedelegate) ValidateBasic() types.Error {
	if msg.DelegatorAddress.Empty() {
		return ErrNilDelegatorAddr(DefaultCodespace)
	}
	if msg.ValidatorSrcAddress.Empty() {
		return ErrNilValidatorAddr(DefaultCodespace)
	}
	if msg.ValidatorDstAddress.Empty() {
		return ErrNilValidatorAddr(DefaultCodespace)
	}
	if msg.Amount.Amount.LTE(types.ZeroInt()) {
		return ErrBadSharesAmount(DefaultCodespace)
	}
	return nil
}

// MsgUndelegate - struct for unbonding transactions
type MsgUndelegate struct {
	DelegatorAddress types.AccAddress `json:"delegator_address" yaml:"delegator_address"`
	ValidatorAddress types.ValAddress `json:"validator_address" yaml:"validator_address"`
	Amount           types.Coin       `json:"amount" yaml:"amount"`
}

func NewMsgUndelegate(delAddr types.AccAddress, valAddr types.ValAddress, amount types.Coin) MsgUndelegate {
	return MsgUndelegate{
		DelegatorAddress: delAddr,
		ValidatorAddress: valAddr,
		Amount:           amount,
	}
}

//nolint
func (msg MsgUndelegate) Route() string { return RouterKey }
func (msg MsgUndelegate) Type() string  { return "begin_unbonding" }
func (msg MsgUndelegate) GetSigners() []types.AccAddress {
	return []types.AccAddress{msg.DelegatorAddress}
}

// get the bytes for the message signer to sign on
func (msg MsgUndelegate) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return types.MustSortJSON(bz)
}

// quick validity check
func (msg MsgUndelegate) ValidateBasic() types.Error {
	if msg.DelegatorAddress.Empty() {
		return ErrNilDelegatorAddr(DefaultCodespace)
	}
	if msg.ValidatorAddress.Empty() {
		return ErrNilValidatorAddr(DefaultCodespace)
	}
	if msg.Amount.Amount.LTE(types.ZeroInt()) {
		return ErrBadSharesAmount(DefaultCodespace)
	}
	return nil
}
