package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/crypto"
)

// name to idetify transaction types
const MsgType = "stake"

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
	b, err := MsgCdc.MarshalJSON(struct {
		Description
		ValidatorAddr string   `json:"address"`
		PubKey        string   `json:"pubkey"`
		Bond          sdk.Coin `json:"bond"`
	}{
		Description:   msg.Description,
		ValidatorAddr: sdk.MustBech32ifyVal(msg.ValidatorAddr),
		PubKey:        sdk.MustBech32ifyValPub(msg.PubKey),
	})
	if err != nil {
		panic(err)
	}
	return b
}

// quick validity check
func (msg MsgCreateValidator) ValidateBasic() sdk.Error {
	if msg.ValidatorAddr == nil {
		return ErrNilValidatorAddr(DefaultCodespace)
	}
	if !(msg.SelfDelegation.Amount.GT(sdk.ZeroInt())) {
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
	b, err := MsgCdc.MarshalJSON(struct {
		Description
		ValidatorAddr string `json:"address"`
	}{
		Description:   msg.Description,
		ValidatorAddr: sdk.MustBech32ifyVal(msg.ValidatorAddr),
	})
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
	b, err := MsgCdc.MarshalJSON(struct {
		DelegatorAddr string   `json:"delegator_addr"`
		ValidatorAddr string   `json:"validator_addr"`
		Bond          sdk.Coin `json:"bond"`
	}{
		DelegatorAddr: sdk.MustBech32ifyAcc(msg.DelegatorAddr),
		ValidatorAddr: sdk.MustBech32ifyVal(msg.ValidatorAddr),
		Bond:          msg.Bond,
	})
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
	if !(msg.Bond.Amount.GT(sdk.ZeroInt())) {
		return ErrBadDelegationAmount(DefaultCodespace)
	}
	return nil
}

//______________________________________________________________________

// MsgDelegate - struct for bonding transactions
type MsgBeginRedelegate struct {
	DelegatorAddr    sdk.Address `json:"delegator_addr"`
	ValidatorSrcAddr sdk.Address `json:"validator_src_addr"`
	ValidatorDstAddr sdk.Address `json:"validator_dst_addr"`
	SharesAmount     sdk.Rat     `json:"shares_amount"`
}

func NewMsgBeginRedelegate(delegatorAddr, validatorSrcAddr,
	validatorDstAddr sdk.Address, sharesAmount sdk.Rat) MsgBeginRedelegate {

	return MsgBeginRedelegate{
		DelegatorAddr:    delegatorAddr,
		ValidatorSrcAddr: validatorSrcAddr,
		ValidatorDstAddr: validatorDstAddr,
		SharesAmount:     sharesAmount,
	}
}

//nolint
func (msg MsgBeginRedelegate) Type() string { return MsgType }
func (msg MsgBeginRedelegate) GetSigners() []sdk.Address {
	return []sdk.Address{msg.DelegatorAddr}
}

// get the bytes for the message signer to sign on
func (msg MsgBeginRedelegate) GetSignBytes() []byte {
	b, err := MsgCdc.MarshalJSON(struct {
		DelegatorAddr    string `json:"delegator_addr"`
		ValidatorSrcAddr string `json:"validator_src_addr"`
		ValidatorDstAddr string `json:"validator_dst_addr"`
		SharesAmount     string `json:"shares"`
	}{
		DelegatorAddr:    sdk.MustBech32ifyAcc(msg.DelegatorAddr),
		ValidatorSrcAddr: sdk.MustBech32ifyVal(msg.ValidatorSrcAddr),
		ValidatorDstAddr: sdk.MustBech32ifyVal(msg.ValidatorDstAddr),
		SharesAmount:     msg.SharesAmount.String(),
	})
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
	if msg.SharesAmount.LTE(sdk.ZeroRat()) {
		return ErrBadSharesAmount(DefaultCodespace)
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
	b, err := MsgCdc.MarshalJSON(struct {
		DelegatorAddr    string `json:"delegator_addr"`
		ValidatorSrcAddr string `json:"validator_src_addr"`
		ValidatorDstAddr string `json:"validator_dst_addr"`
	}{
		DelegatorAddr:    sdk.MustBech32ifyAcc(msg.DelegatorAddr),
		ValidatorSrcAddr: sdk.MustBech32ifyVal(msg.ValidatorSrcAddr),
		ValidatorDstAddr: sdk.MustBech32ifyVal(msg.ValidatorDstAddr),
	})
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
}

func NewMsgBeginUnbonding(delegatorAddr, validatorAddr sdk.Address, sharesAmount sdk.Rat) MsgBeginUnbonding {
	return MsgBeginUnbonding{
		DelegatorAddr: delegatorAddr,
		ValidatorAddr: validatorAddr,
		SharesAmount:  sharesAmount,
	}
}

//nolint
func (msg MsgBeginUnbonding) Type() string              { return MsgType }
func (msg MsgBeginUnbonding) GetSigners() []sdk.Address { return []sdk.Address{msg.DelegatorAddr} }

// get the bytes for the message signer to sign on
func (msg MsgBeginUnbonding) GetSignBytes() []byte {
	b, err := MsgCdc.MarshalJSON(struct {
		DelegatorAddr string `json:"delegator_addr"`
		ValidatorAddr string `json:"validator_addr"`
		SharesAmount  string `json:"shares_amount"`
	}{
		DelegatorAddr: sdk.MustBech32ifyAcc(msg.DelegatorAddr),
		ValidatorAddr: sdk.MustBech32ifyVal(msg.ValidatorAddr),
		SharesAmount:  msg.SharesAmount.String(),
	})
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
	if msg.SharesAmount.LTE(sdk.ZeroRat()) {
		return ErrBadSharesAmount(DefaultCodespace)
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
	b, err := MsgCdc.MarshalJSON(struct {
		DelegatorAddr string `json:"delegator_addr"`
		ValidatorAddr string `json:"validator_src_addr"`
	}{
		DelegatorAddr: sdk.MustBech32ifyAcc(msg.DelegatorAddr),
		ValidatorAddr: sdk.MustBech32ifyVal(msg.ValidatorAddr),
	})
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
