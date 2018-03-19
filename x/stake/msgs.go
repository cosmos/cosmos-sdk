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

// DeclareCandidacyMsg - struct for unbonding transactions
type DeclareCandidacyMsg struct {
	Address     sdk.Address   `json:"address"`
	Bond        sdk.Coin      `json:"bond"`
	PubKey      crypto.PubKey `json:"pubkey"`
	Description Description   `json:"description"`
}

func NewDeclareCandidacyMsg(address sdk.Address, pubkey crypto.PubKey, bond sdk.Coin, description Description) DeclareCandidacyMsg {
	return DeclareCandidacyMsg{
		Address:     address,
		Bond:        bond,
		Description: description,
		PubKey:      pubkey,
	}
}

func (msg DeclareCandidacyMsg) Type() string                            { return MsgType }
func (msg DeclareCandidacyMsg) Get(key interface{}) (value interface{}) { return nil }
func (msg DeclareCandidacyMsg) GetSigners() []sdk.Address               { return []sdk.Address{msg.Address} }
func (msg DeclareCandidacyMsg) String() string {
	return fmt.Sprintf("DeclareCandidacyMsg{Address: %v, Bond: %v, PubKey: %v, Description: %v}", msg.Address, msg.Bond, msg.PubKey, msg.Description)
}

// ValidateBasic - Check for non-empty candidate, and valid coins
func (msg DeclareCandidacyMsg) ValidateBasic() sdk.Error {
	if len(msg.Address) == 0 {
		return sdk.ErrInvalidAddress(msg.Address.String())
	}
	if !msg.Bond.Amount <= 0 {
		return sdk.ErrInvalidCoins(msg.Bond.String())
	}
	empty := Description{}
	if msg.Description == empty {
		return ErrInvalidDescription(msg.Description) // TODO: Create Error
	}
	// TODO: Verify pubkey somehow?
	return nil
}

// get the bytes for the message signer to sign on
func (msg DeclareCandidacyMsg) GetSignBytes() []byte {
	b, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return b
}

//______________________________________________________________________

// EditCandidacyMsg - struct for editing a candidate
type EditCandidacyMsg struct {
	Address     sdk.Address `json:"address"`
	Description Description `json:"description"`
}

func NewEditCandidacyMsg(address sdk.Address, description Description) EditCandidacyMsg {
	return EditCandidacyMsg{
		Address:     address,
		Description: description,
	}
}

func (msg EditCandidacyMsg) Type() string                            { return MsgType }
func (msg EditCandidacyMsg) Get(key interface{}) (value interface{}) { return nil }
func (msg EditCandidacyMsg) GetSigners() []sdk.Address               { return []sdk.Address{msg.Address} }
func (msg EditCandidacyMsg) String() string {
	return fmt.Sprintf("EditCandidacyMsg{Address: %v, Description: %v}", msg.Address, msg.Description)
}

// quick validity check
func (msg EditCandidacyMsg) ValidateBasic() sdk.Error {
	if len(msg.Address) == 0 {
		return sdk.ErrInvalidAddress(msg.Address.String())
	}
	empty := Description{}
	if msg.Description == empty {
		return ErrInvalidDescription(msg.Description)
	}
	// TODO: Verify pubkey somehow?
	return nil
}

// get the bytes for the message signer to sign on
func (msg EditCandidacyMsg) GetSignBytes() []byte {
	b, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return b
}

//______________________________________________________________________

// DelegateMsg - struct for bonding transactions
type DelegateMsg struct {
	Delegator sdk.Address `json:"delegator"`
	Delegatee sdk.Address `json:"delegatee"`
	Bond      sdk.Coin    `json:"bond"`
}

func NewDelegateMsg(delegator sdk.Address, delegatee sdk.Address, bond sdk.Coin) DelegateMsg {
	return DelegateMsg{
		Delegator: delegator,
		Delegatee: delegatee,
		Bond:      bond,
	}
}

func (msg DelegateMsg) Type() string                            { return MsgType }
func (msg DelegateMsg) Get(key interface{}) (value interface{}) { return nil }
func (msg DelegateMsg) GetSigners() []sdk.Address               { return []sdk.Address{msg.Delegator} }
func (msg DelegateMsg) String() string {
	return fmt.Sprintf("DelegateMsg{Delegator: %v, Delegatee: %v, Bond: %v}", msg.Delegator, msg.Delegatee, msg.Bond)
}

// quick validity check
func (msg DelegateMsg) ValidateBasic() sdk.Error {
	if len(msg.Delegator) == 0 {
		return sdk.ErrInvalidAddress(msg.Delegator.String())
	}
	if len(msg.Delegatee) == 0 {
		return sdk.ErrInvalidAddress(msg.Delegatee.String())
	}
	if !msg.Bond.Amount <= 0 {
		return sdk.ErrInvalidCoins(msg.Bond.String())
	}
	return nil
}

// get the bytes for the message signer to sign on
func (msg DelegateMsg) GetSignBytes() []byte {
	b, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return b
}

//______________________________________________________________________

// UnbondMsg - struct for unbonding transactions
type UnbondMsg struct {
	Address sdk.Address `json:"address"`
	Shares  string      `json:"shares"`
}

func NewUnbondMsg(address sdk.Address, shares string) UnbondMsg {
	return UnbondMsg{
		Address: address,
		Shares:  shares,
	}
}

func (msg UnbondMsg) Type() string                            { return MsgType }
func (msg UnbondMsg) Get(key interface{}) (value interface{}) { return nil }
func (msg UnbondMsg) GetSigners() []sdk.Address               { return []sdk.Address{msg.Delegator} }
func (msg UnbondMsg) String() string {
	return fmt.Sprintf("UnbondMsg{Address: %v, Shares: %v}", msg.Address, msg.Shares)
}

// quick validity check
func (msg UnbondMsg) ValidateBasic() sdk.Error {
	err := msg.MsgAddr.ValidateBasic()
	if err != nil {
		return err
	}

	if msg.Shares == "MAX" {
		return ErrCandidateEmpty() // TODO: Understand?
	}
	return nil
}

// get the bytes for the message signer to sign on
func (msg UnbondMsg) GetSignBytes() []byte {
	b, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return b
}
