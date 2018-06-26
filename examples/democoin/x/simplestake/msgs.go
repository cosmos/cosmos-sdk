package simplestake

import (
	"encoding/json"

	crypto "github.com/tendermint/go-crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

//_________________________________________________________----

// simple bond message
type MsgBond struct {
	Address sdk.Address   `json:"address"`
	Stake   sdk.Coin      `json:"coins"`
	PubKey  crypto.PubKey `json:"pub_key"`
}

func NewMsgBond(addr sdk.Address, stake sdk.Coin, pubKey crypto.PubKey) MsgBond {
	return MsgBond{
		Address: addr,
		Stake:   stake,
		PubKey:  pubKey,
	}
}

//nolint
func (msg MsgBond) Type() string              { return moduleName } //TODO update "stake/createvalidator"
func (msg MsgBond) GetSigners() []sdk.Address { return []sdk.Address{msg.Address} }

// basic validation of the bond message
func (msg MsgBond) ValidateBasic() sdk.Error {
	if msg.Stake.IsZero() {
		return ErrEmptyStake(DefaultCodespace)
	}

	if msg.PubKey == nil {
		return sdk.ErrInvalidPubKey("MsgBond.PubKey must not be empty")
	}

	return nil
}

// get bond message sign bytes
func (msg MsgBond) GetSignBytes() []byte {
	bz, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return bz
}

//_______________________________________________________________

// simple unbond message
type MsgUnbond struct {
	Address sdk.Address `json:"address"`
}

func NewMsgUnbond(addr sdk.Address) MsgUnbond {
	return MsgUnbond{
		Address: addr,
	}
}

//nolint
func (msg MsgUnbond) Type() string              { return moduleName } //TODO update "stake/createvalidator"
func (msg MsgUnbond) GetSigners() []sdk.Address { return []sdk.Address{msg.Address} }
func (msg MsgUnbond) ValidateBasic() sdk.Error  { return nil }

// get unbond message sign bytes
func (msg MsgUnbond) GetSignBytes() []byte {
	bz, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return bz
}
