package stake

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	crypto "github.com/tendermint/go-crypto"
)

// name to idetify transaction types
var MsgType = "stake"

//Verify interface at compile time
var _, _, _, _ sdk.Msg = &DeclareCandidacyMsg{}, &EditCandidacyMsg{}, &DelegateMsg{}, &UnbondMsg{}

//______________________________________________________________________

// MsgAddr - struct for bonding or unbonding transactions
type MsgAddr struct {
	Address sdk.Address `json:"address"`
}

func NewMsgAddr(address sdk.Address) MsgAddr {
	return MsgAddr{
		Address: address,
	}
}

// nolint
func (msg MsgAddr) Type() string                            { return Name }
func (msg MsgAddr) Get(key interface{}) (value interface{}) { return nil }
func (msg MsgAddr) GetSigners() []sdk.Address               { return []sdk.Address{msg.Address} }
func (msg MsgAddr) String() string {
	return fmt.Sprintf("MsgAddr{Address: %v}", msg.Address)
}

// ValidateBasic - Check for non-empty candidate, and valid coins
func (msg MsgAddr) ValidateBasic() sdk.Error {
	if msg.Address == nil {
		return ErrCandidateEmpty()
	}
	return nil
}

//______________________________________________________________________

// DeclareCandidacyMsg - struct for unbonding transactions
type DeclareCandidacyMsg struct {
	MsgAddr
	Description
	Bond    sdk.Coin      `json:"bond"`
	Address sdk.Address   `json:"address"`
	PubKey  crypto.PubKey `json:"pubkey"`
}

func NewDeclareCandidacyMsg(address sdk.Address, pubkey crypto.PubKey, bond sdk.Coin, description Description) DeclareCandidacyMsg {
	return DeclareCandidacyMsg{
		MsgAddr:     NewMsgAddr(address),
		Description: description,
		Bond:        bond,
		Address:     address,
		PubKey:      pubkey,
	}
}

// get the bytes for the message signer to sign on
func (msg DeclareCandidacyMsg) GetSignBytes() []byte {
	b, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return b
}

// quick validity check
func (msg DeclareCandidacyMsg) ValidateBasic() sdk.Error {
	err := msg.MsgAddr.ValidateBasic()
	if err != nil {
		return err
	}
	err = validateCoin(msg.Bond)
	if err != nil {
		return err
	}
	empty := Description{}
	if msg.Description == empty {
		return newError(CodeInvalidInput, "description must be included")
	}
	return nil
}

//______________________________________________________________________

// EditCandidacyMsg - struct for editing a candidate
type EditCandidacyMsg struct {
	MsgAddr
	Description
}

func NewEditCandidacyMsg(address sdk.Address, description Description) EditCandidacyMsg {
	return EditCandidacyMsg{
		MsgAddr:     NewMsgAddr(address),
		Description: description,
	}
}

// get the bytes for the message signer to sign on
func (msg EditCandidacyMsg) GetSignBytes() []byte {
	b, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return b
}

// quick validity check
func (msg EditCandidacyMsg) ValidateBasic() sdk.Error {
	err := msg.MsgAddr.ValidateBasic()
	if err != nil {
		return err
	}
	empty := Description{}
	if msg.Description == empty {
		return newError(CodeInvalidInput, "Transaction must include some information to modify")
	}
	return nil
}

//______________________________________________________________________

// DelegateMsg - struct for bonding transactions
type DelegateMsg struct {
	MsgAddr
	Bond sdk.Coin `json:"bond"`
}

func NewDelegateMsg(address sdk.Address, bond sdk.Coin) DelegateMsg {
	return DelegateMsg{
		MsgAddr: NewMsgAddr(address),
		Bond:    bond,
	}
}

// get the bytes for the message signer to sign on
func (msg DelegateMsg) GetSignBytes() []byte {
	b, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return b
}

// quick validity check
func (msg DelegateMsg) ValidateBasic() sdk.Error {
	err := msg.MsgAddr.ValidateBasic()
	if err != nil {
		return err
	}
	err = validateCoin(msg.Bond)
	if err != nil {
		return err
	}
	return nil
}

//______________________________________________________________________

// UnbondMsg - struct for unbonding transactions
type UnbondMsg struct {
	MsgAddr
	Shares string `json:"shares"`
}

func NewUnbondMsg(address sdk.Address, shares string) UnbondMsg {
	return UnbondMsg{
		MsgAddr: NewMsgAddr(address),
		Shares:  shares,
	}
}

// get the bytes for the message signer to sign on
func (msg UnbondMsg) GetSignBytes() []byte {
	b, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return b
}

// quick validity check
func (msg UnbondMsg) ValidateBasic() sdk.Error {
	err := msg.MsgAddr.ValidateBasic()
	if err != nil {
		return err
	}

	if msg.Shares == "MAX" {
		return ErrCandidateEmpty()
	}
	return nil
}

//______________________________________________________________________
// helper

func validateCoin(coin sdk.Coin) sdk.Error {
	coins := sdk.Coins{coin}
	if !coins.IsValid() {
		return sdk.ErrInvalidCoins(coins)
	}
	if !coins.IsPositive() {
		return sdk.ErrInvalidCoins(coins) // XXX: add "Amount must be > 0" ?
	}
	return nil
}
