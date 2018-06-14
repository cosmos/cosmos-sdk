package stake

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
var _, _, _, _ sdk.Msg = &MsgCreateValidator{}, &MsgEditValidator{}, &MsgDelegate{}, &MsgUnbond{}

//______________________________________________________________________

// MsgCreateValidator - struct for unbonding transactions
type MsgCreateValidator struct {
	Description
	ValidatorAddr sdk.Address   `json:"address"`
	PubKey        crypto.PubKey `json:"pubkey"`
	Bond          sdk.Coin      `json:"bond"`
}

func NewMsgCreateValidator(validatorAddr sdk.Address, pubkey crypto.PubKey,
	bond sdk.Coin, description Description) MsgCreateValidator {
	return MsgCreateValidator{
		Description:   description,
		ValidatorAddr: validatorAddr,
		PubKey:        pubkey,
		Bond:          bond,
	}
}

//nolint
func (msg MsgCreateValidator) Type() string { return MsgType }
func (msg MsgCreateValidator) GetSigners() []sdk.Address {
	return []sdk.Address{msg.ValidatorAddr}
}

// get the bytes for the message signer to sign on
func (msg MsgCreateValidator) GetSignBytes() []byte {
	b, err := msgCdc.MarshalJSON(struct {
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
		return ErrValidatorEmpty(DefaultCodespace)
	}
	if msg.Bond.Denom != StakingToken {
		return ErrBadBondingDenom(DefaultCodespace)
	}
	if msg.Bond.Amount <= 0 {
		return ErrBadBondingAmount(DefaultCodespace)
	}
	empty := Description{}
	if msg.Description == empty {
		return newError(DefaultCodespace, CodeInvalidInput, "description must be included")
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
	b, err := msgCdc.MarshalJSON(struct {
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
		return ErrValidatorEmpty(DefaultCodespace)
	}
	empty := Description{}
	if msg.Description == empty {
		return newError(DefaultCodespace, CodeInvalidInput, "transaction must include some information to modify")
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
	b, err := msgCdc.MarshalJSON(struct {
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
		return ErrBadDelegatorAddr(DefaultCodespace)
	}
	if msg.ValidatorAddr == nil {
		return ErrBadValidatorAddr(DefaultCodespace)
	}
	if msg.Bond.Denom != StakingToken {
		return ErrBadBondingDenom(DefaultCodespace)
	}
	if msg.Bond.Amount <= 0 {
		return ErrBadBondingAmount(DefaultCodespace)
	}
	return nil
}

//______________________________________________________________________

// MsgUnbond - struct for unbonding transactions
type MsgUnbond struct {
	DelegatorAddr sdk.Address `json:"delegator_addr"`
	ValidatorAddr sdk.Address `json:"validator_addr"`
	Shares        string      `json:"shares"`
}

func NewMsgUnbond(delegatorAddr, validatorAddr sdk.Address, shares string) MsgUnbond {
	return MsgUnbond{
		DelegatorAddr: delegatorAddr,
		ValidatorAddr: validatorAddr,
		Shares:        shares,
	}
}

//nolint
func (msg MsgUnbond) Type() string              { return MsgType }
func (msg MsgUnbond) GetSigners() []sdk.Address { return []sdk.Address{msg.DelegatorAddr} }

// get the bytes for the message signer to sign on
func (msg MsgUnbond) GetSignBytes() []byte {
	b, err := msgCdc.MarshalJSON(struct {
		DelegatorAddr string `json:"delegator_addr"`
		ValidatorAddr string `json:"validator_addr"`
		Shares        string `json:"shares"`
	}{
		DelegatorAddr: sdk.MustBech32ifyAcc(msg.DelegatorAddr),
		ValidatorAddr: sdk.MustBech32ifyVal(msg.ValidatorAddr),
		Shares:        msg.Shares,
	})
	if err != nil {
		panic(err)
	}
	return b
}

// quick validity check
func (msg MsgUnbond) ValidateBasic() sdk.Error {
	if msg.DelegatorAddr == nil {
		return ErrBadDelegatorAddr(DefaultCodespace)
	}
	if msg.ValidatorAddr == nil {
		return ErrBadValidatorAddr(DefaultCodespace)
	}
	if msg.Shares != "MAX" {
		rat, err := sdk.NewRatFromDecimal(msg.Shares)
		if err != nil {
			return ErrBadShares(DefaultCodespace)
		}
		if rat.IsZero() || rat.LT(sdk.ZeroRat()) {
			return ErrBadShares(DefaultCodespace)
		}
	}
	return nil
}
