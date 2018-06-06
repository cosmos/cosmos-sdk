package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	crypto "github.com/tendermint/go-crypto"
)

// name to idetify transaction types
const MsgType = "stake"

// XXX remove: think it makes more sense belonging with the Params so we can
// initialize at genesis - to allow for the same tests we should should make
// the ValidateBasic() function a return from an initializable function
// ValidateBasic(bondDenom string) function
const StakingToken = "steak"

//Verify interface at compile time
var _, _, _ sdk.Msg = &MsgCreateValidator{}, &MsgEditValidator{}, &MsgDelegate{}
var _, _ sdk.Msg = &MsgBeginUnbonding{}, &MsgCompleteUnbonding{}
var _, _ sdk.Msg = &MsgBeginRedelegate{}, &MsgCompleteRedelegate{}

//______________________________________________________________________

// MsgCreateValidator - struct for unbonding transactions
type MsgCreateValidator struct {
	Description
	ValidatorAddr  sdk.Address   `json:"address"`
	PubKey         crypto.PubKey `json:"pubkey"`
	SelfDelegation sdk.Coin      `json:"self_delegation"`
}

func NewMsgCreateValidator(validatorAddr sdk.Address, pubkey crypto.PubKey,
	selfDelegation sdk.Coin, description Description) MsgCreateValidator {
	return MsgCreateValidator{
		Description:    description,
		ValidatorAddr:  validatorAddr,
		PubKey:         pubkey,
		SelfDelegation: selfDelegation,
	}
}

//nolint
func (msg MsgCreateValidator) Type() string { return MsgType }
func (msg MsgCreateValidator) GetSigners() []sdk.Address {
	return []sdk.Address{msg.ValidatorAddr}
}

// get the bytes for the message signer to sign on
func (msg MsgCreateValidator) GetSignBytes() []byte {
	return MsgCdc.MustMarshalBinary(msg)
}

// quick validity check
func (msg MsgCreateValidator) ValidateBasic() sdk.Error {
	if msg.ValidatorAddr == nil {
		return ErrNilValidatorAddr(DefaultCodespace)
	}
	if msg.SelfDelegation.Denom != StakingToken {
		return ErrBadDenom(DefaultCodespace)
	}
	if msg.SelfDelegation.Amount <= 0 {
		return ErrBadDelegationAmount(DefaultCodespace)
	}
	empty := Description{}
	if msg.Description == empty {
		return sdk.NewError(DefaultCodespace, CodeInvalidInput, "description must be included")
	}
	return nil
}

//______________________________________________________________________

// MsgEditValidator - struct for editing a validator
type MsgEditValidator struct {
	Description
	ValidatorAddr sdk.Address `json:"address"`
}

func NewMsgEditValidator(validatorAddr sdk.Address, description Description) MsgEditValidator {
	return MsgEditValidator{
		Description:   description,
		ValidatorAddr: validatorAddr,
	}
}

//nolint
func (msg MsgEditValidator) Type() string { return MsgType }
func (msg MsgEditValidator) GetSigners() []sdk.Address {
	return []sdk.Address{msg.ValidatorAddr}
}

// get the bytes for the message signer to sign on
func (msg MsgEditValidator) GetSignBytes() []byte {
	b, err := MsgCdc.MarshalJSON(msg)
	if err != nil {
		panic(err)
	}
	return b
}

// quick validity check
func (msg MsgEditValidator) ValidateBasic() sdk.Error {
	if msg.ValidatorAddr == nil {
		return sdk.NewError(DefaultCodespace, CodeInvalidInput, "nil validator address")
	}
	empty := Description{}
	if msg.Description == empty {
		return sdk.NewError(DefaultCodespace, CodeInvalidInput, "transaction must include some information to modify")
	}
	return nil
}

//______________________________________________________________________

// MsgDelegate - struct for bonding transactions
type MsgDelegate struct {
	DelegatorAddr sdk.Address `json:"delegator_addr"`
	ValidatorAddr sdk.Address `json:"validator_addr"`
	Bond          sdk.Coin    `json:"bond"`
}

func NewMsgDelegate(delegatorAddr, validatorAddr sdk.Address, bond sdk.Coin) MsgDelegate {
	return MsgDelegate{
		DelegatorAddr: delegatorAddr,
		ValidatorAddr: validatorAddr,
		Bond:          bond,
	}
}

//nolint
func (msg MsgDelegate) Type() string { return MsgType }
func (msg MsgDelegate) GetSigners() []sdk.Address {
	return []sdk.Address{msg.DelegatorAddr}
}

// get the bytes for the message signer to sign on
func (msg MsgDelegate) GetSignBytes() []byte {
	b, err := MsgCdc.MarshalJSON(msg)
	if err != nil {
		panic(err)
	}
	return b
}

// quick validity check
func (msg MsgDelegate) ValidateBasic() sdk.Error {
	if msg.DelegatorAddr == nil {
		return ErrNilDelegatorAddr(DefaultCodespace)
	}
	if msg.ValidatorAddr == nil {
		return ErrNilValidatorAddr(DefaultCodespace)
	}
	if msg.Bond.Denom != StakingToken {
		return ErrBadDenom(DefaultCodespace)
	}
	if msg.Bond.Amount <= 0 {
		return ErrBadDelegationAmount(DefaultCodespace)
	}
	return nil
}

//______________________________________________________________________

// MsgDelegate - struct for bonding transactions
type MsgBeginRedelegate struct {
	DelegatorAddr    sdk.Address `json:"delegator_addr"`
	ValidatorSrcAddr sdk.Address `json:"validator_source_addr"`
	ValidatorDstAddr sdk.Address `json:"validator_destination_addr"`
	SharesAmount     sdk.Rat     `json:"shares_amount"`
	SharesPercent    sdk.Rat     `json:"shares_percent"`
}

func NewMsgBeginRedelegate(delegatorAddr, validatorSrcAddr,
	validatorDstAddr sdk.Address, sharesAmount, sharesPercent sdk.Rat) MsgBeginRedelegate {

	return MsgBeginRedelegate{
		DelegatorAddr:    delegatorAddr,
		ValidatorSrcAddr: validatorSrcAddr,
		ValidatorDstAddr: validatorDstAddr,
		SharesAmount:     sharesAmount,
		SharesPercent:    sharesPercent,
	}
}

//nolint
func (msg MsgBeginRedelegate) Type() string { return MsgType }
func (msg MsgBeginRedelegate) GetSigners() []sdk.Address {
	return []sdk.Address{msg.DelegatorAddr}
}

// get the bytes for the message signer to sign on
func (msg MsgBeginRedelegate) GetSignBytes() []byte {
	b, err := MsgCdc.MarshalJSON(msg)
	if err != nil {
		panic(err)
	}
	return b
}

// quick validity check
func (msg MsgBeginRedelegate) ValidateBasic() sdk.Error {
	if msg.DelegatorAddr == nil {
		return ErrNilDelegatorAddr(DefaultCodespace)
	}
	if msg.ValidatorSrcAddr == nil {
		return ErrNilValidatorAddr(DefaultCodespace)
	}
	if msg.ValidatorDstAddr == nil {
		return ErrNilValidatorAddr(DefaultCodespace)
	}
	err := testShares(msg.SharesAmount, msg.SharesPercent)
	if err != nil {
		return err
	}
	return nil
}

func testShares(sharesAmount, sharesPercent sdk.Rat) sdk.Error {
	if !sharesAmount.IsZero() && !sharesPercent.IsZero() {
		return ErrBothShareMsgsGiven(DefaultCodespace)
	}
	if sharesAmount.IsZero() && sharesPercent.IsZero() {
		return ErrNeitherShareMsgsGiven(DefaultCodespace)
	}
	if sharesPercent.IsZero() && !sharesAmount.IsZero() && sharesAmount.LTE(sdk.ZeroRat()) {
		return ErrBadSharesAmount(DefaultCodespace)
	}
	if sharesAmount.IsZero() && (sharesPercent.LTE(sdk.ZeroRat()) || sharesPercent.GT(sdk.OneRat())) {
		return ErrBadSharesPercent(DefaultCodespace)
	}
	return nil
}

// MsgDelegate - struct for bonding transactions
type MsgCompleteRedelegate struct {
	DelegatorAddr    sdk.Address `json:"delegator_addr"`
	ValidatorSrcAddr sdk.Address `json:"validator_source_addr"`
	ValidatorDstAddr sdk.Address `json:"validator_destination_addr"`
}

func NewMsgCompleteRedelegate(delegatorAddr, validatorSrcAddr,
	validatorDstAddr sdk.Address) MsgCompleteRedelegate {

	return MsgCompleteRedelegate{
		DelegatorAddr:    delegatorAddr,
		ValidatorSrcAddr: validatorSrcAddr,
		ValidatorDstAddr: validatorDstAddr,
	}
}

//nolint
func (msg MsgCompleteRedelegate) Type() string { return MsgType }
func (msg MsgCompleteRedelegate) GetSigners() []sdk.Address {
	return []sdk.Address{msg.DelegatorAddr}
}

// get the bytes for the message signer to sign on
func (msg MsgCompleteRedelegate) GetSignBytes() []byte {
	b, err := MsgCdc.MarshalJSON(msg)
	if err != nil {
		panic(err)
	}
	return b
}

// quick validity check
func (msg MsgCompleteRedelegate) ValidateBasic() sdk.Error {
	if msg.DelegatorAddr == nil {
		return ErrNilDelegatorAddr(DefaultCodespace)
	}
	if msg.ValidatorSrcAddr == nil {
		return ErrNilValidatorAddr(DefaultCodespace)
	}
	if msg.ValidatorDstAddr == nil {
		return ErrNilValidatorAddr(DefaultCodespace)
	}
	return nil
}

//______________________________________________________________________

// MsgBeginUnbonding - struct for unbonding transactions
type MsgBeginUnbonding struct {
	DelegatorAddr sdk.Address `json:"delegator_addr"`
	ValidatorAddr sdk.Address `json:"validator_addr"`
	SharesAmount  sdk.Rat     `json:"shares_amount"`
	SharesPercent sdk.Rat     `json:"shares_percent"`
}

func NewMsgBeginUnbonding(delegatorAddr, validatorAddr sdk.Address, sharesAmount, sharesPercent sdk.Rat) MsgBeginUnbonding {
	return MsgBeginUnbonding{
		DelegatorAddr: delegatorAddr,
		ValidatorAddr: validatorAddr,
		SharesAmount:  sharesAmount,
		SharesPercent: sharesPercent,
	}
}

//nolint
func (msg MsgBeginUnbonding) Type() string              { return MsgType }
func (msg MsgBeginUnbonding) GetSigners() []sdk.Address { return []sdk.Address{msg.DelegatorAddr} }

// get the bytes for the message signer to sign on
func (msg MsgBeginUnbonding) GetSignBytes() []byte {
	b, err := MsgCdc.MarshalJSON(msg)
	if err != nil {
		panic(err)
	}
	return b
}

// quick validity check
func (msg MsgBeginUnbonding) ValidateBasic() sdk.Error {
	if msg.DelegatorAddr == nil {
		return ErrNilDelegatorAddr(DefaultCodespace)
	}
	if msg.ValidatorAddr == nil {
		return ErrNilValidatorAddr(DefaultCodespace)
	}
	err := testShares(msg.SharesAmount, msg.SharesPercent)
	if err != nil {
		return err
	}
	return nil
}

// MsgCompleteUnbonding - struct for unbonding transactions
type MsgCompleteUnbonding struct {
	DelegatorAddr sdk.Address `json:"delegator_addr"`
	ValidatorAddr sdk.Address `json:"validator_addr"`
}

func NewMsgCompleteUnbonding(delegatorAddr, validatorAddr sdk.Address) MsgCompleteUnbonding {
	return MsgCompleteUnbonding{
		DelegatorAddr: delegatorAddr,
		ValidatorAddr: validatorAddr,
	}
}

//nolint
func (msg MsgCompleteUnbonding) Type() string              { return MsgType }
func (msg MsgCompleteUnbonding) GetSigners() []sdk.Address { return []sdk.Address{msg.DelegatorAddr} }

// get the bytes for the message signer to sign on
func (msg MsgCompleteUnbonding) GetSignBytes() []byte {
	b, err := MsgCdc.MarshalJSON(msg)
	if err != nil {
		panic(err)
	}
	return b
}

// quick validity check
func (msg MsgCompleteUnbonding) ValidateBasic() sdk.Error {
	if msg.DelegatorAddr == nil {
		return ErrNilDelegatorAddr(DefaultCodespace)
	}
	if msg.ValidatorAddr == nil {
		return ErrNilValidatorAddr(DefaultCodespace)
	}
	return nil
}
